package client

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	// "github.com/metal-toolbox/fleet-scheduler/internal/model"

	// "github.com/pkg/errors"
	"github.com/coreos/go-oidc"
	"github.com/hashicorp/go-retryablehttp"
	conditionOrcapi "github.com/metal-toolbox/conditionorc/pkg/api/v1/client"
	"github.com/metal-toolbox/fleet-scheduler/internal/app"
	"github.com/metal-toolbox/fleet-scheduler/internal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	fleetDBapi "go.hollow.sh/serverservice/pkg/api/v1"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2/clientcredentials"
)

const timeout = 30 * time.Second

type Client struct {
	ssClient *fleetDBapi.Client
	coClient *conditionOrcapi.Client
	cfg      *app.Configuration
	ctx      context.Context
	logger   *logrus.Entry
}

func New(ctx context.Context, cfg *app.Configuration, logger *logrus.Entry) (*Client, error) {
	client := &Client{
		cfg:    cfg,
		ctx:    ctx,
		logger: logger,
	}

	err := client.newFleetDBClient()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize FleetDB Client (Serverservice)")
	}

	err = client.newConditionOrcClient()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize ConditionOrc Client")
	}

	return client, nil
}

func (c *Client) newFleetDBClient() error {
	var ssClient *fleetDBapi.Client
	var err error

	if c.cfg == nil {
		return ErrNilConfig
	}

	if c.cfg.FdbCfg.DisableOAuth {
		ssClient, err = newFleetDBClientWithoutOAuth(c.cfg.FdbCfg, c.logger)
	} else {
		ssClient, err = newFleetDBClientWithOAuth(c.ctx, c.cfg.FdbCfg, c.logger)
	}

	if err != nil {
		return err
	}

	c.ssClient = ssClient
	return err
}

func (c *Client) newConditionOrcClient() error {
	if c.cfg == nil {
		return ErrNilConfig
	}

	var coClient *conditionOrcapi.Client
	var err error

	if c.cfg.FdbCfg.DisableOAuth {
		coClient, err = conditionOrcapi.NewClient(c.cfg.FdbCfg.Endpoint)
	} else {
		var token string

		token, err = accessToken(c.ctx, model.ConditionsAPI, c.cfg.FdbCfg, true)
		if err != nil {
			return errors.Wrap(ErrAuth, string(model.ConditionsAPI)+err.Error())
		}

		coClient, err = conditionOrcapi.NewClient(c.cfg.FdbCfg.Endpoint, conditionOrcapi.WithAuthToken(token))
	}

	c.coClient = coClient
	return err
}

//// Client initialize helpers

func newFleetDBClientWithOAuth(ctx context.Context, cfg *app.ConfigOIDC, logger *logrus.Entry) (*fleetDBapi.Client, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerEndpoint)
	if err != nil {
		return nil, err
	}

	var clientID string
	if cfg.ClientID != "" {
		clientID = cfg.ClientID
	} else {
		clientID = "fleet-scheduler"
	}

	// setup oauth
	oauthConfig := clientcredentials.Config{
		ClientID:       clientID,
		ClientSecret:   cfg.ClientSecret,
		TokenURL:       provider.Endpoint().TokenURL,
		Scopes:         cfg.ClientScopes,
		EndpointParams: url.Values{"audience": []string{cfg.AudienceEndpoint}},
	}
	oAuthclient := oauthConfig.Client(ctx)

	// set up client
	retryableClient := retryablehttp.NewClient()
	retryableClient.HTTPClient = otelhttp.DefaultClient // use otel client so we can collect telemetry
	retryableClient.HTTPClient.Transport = oAuthclient.Transport
	retryableClient.HTTPClient.Jar = oAuthclient.Jar

	if logger.Level < logrus.DebugLevel {
		retryableClient.Logger = nil
	} else {
		retryableClient.Logger = logger
	}

	client := retryableClient.StandardClient()
	client.Timeout = timeout

	return fleetDBapi.NewClientWithToken(
		cfg.ClientSecret,
		cfg.Endpoint,
		client,
	)
}

func newFleetDBClientWithoutOAuth(cfg *app.ConfigOIDC, logger *logrus.Entry) (*fleetDBapi.Client, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	logHookFunc := func(l retryablehttp.Logger, r *http.Response) {
		// retryablehttp ignores 500 and all errors above 501. So we want to make sure those are logged.
		// https://github.com/hashicorp/go-retryablehttp/blob/4165cf8897205a879a06b20d1ed0a2a76fbb6a17/client.go#L521C80-L521C100
		if r.StatusCode == http.StatusInternalServerError || r.StatusCode > http.StatusNotImplemented {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Warn("fleetDBapi (serverservice) query returned 500 error, got error reading body: ", err.Error())
				return
			}

			logger.Warn("fleetDB (serverservice) query returned 500 error, body: ", string(b))
		}
	}

	// set up client
	retryableClient := retryablehttp.NewClient()
	retryableClient.HTTPClient = otelhttp.DefaultClient // use otel client so we can collect telemetry
	// log hook fo 500 errors since the the retryablehttp client masks them
	retryableClient.ResponseLogHook = logHookFunc

	if logger.Level < logrus.DebugLevel {
		retryableClient.Logger = nil
	} else {
		retryableClient.Logger = logger
	}

	client := retryableClient.StandardClient()
	client.Timeout = timeout

	return fleetDBapi.NewClientWithToken(
		"dummy",
		cfg.Endpoint,
		client,
	)
}
