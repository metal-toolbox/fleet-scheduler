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
	conditionOrcApi "github.com/metal-toolbox/conditionorc/pkg/api/v1/client"
	"github.com/metal-toolbox/fleet-scheduler/internal/app"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	fleetDBApi "go.hollow.sh/serverservice/pkg/api/v1"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2/clientcredentials"
)

const timeout = 30 * time.Second

type Client struct {
	fdbClient *fleetDBApi.Client
	coClient  *conditionOrcApi.Client
	cfg       *app.Configuration
	ctx       context.Context
	logger    *logrus.Entry
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
	if c.cfg == nil {
		return ErrNilConfig
	}

	var err error
	var client *http.Client
	var secret string
	if c.cfg.FdbCfg.DisableOAuth {
		secret = "dummy"

		logHookFunc := func(l retryablehttp.Logger, r *http.Response) {
			// retryablehttp ignores 500 and all errors above 501. So we want to make sure those are logged.
			// https://github.com/hashicorp/go-retryablehttp/blob/4165cf8897205a879a06b20d1ed0a2a76fbb6a17/client.go#L521C80-L521C100
			if r.StatusCode == http.StatusInternalServerError || r.StatusCode > http.StatusNotImplemented {
				// named newErr so the linter doesnt get mad
				b, newErr := io.ReadAll(r.Body)
				if newErr != nil {
					c.logger.Warn("fleetDB (serverservice) query returned 500 error, got error reading body: ", newErr.Error())
					return
				}

				c.logger.Warn("fleetDB (serverservice) query returned 500 error, body: ", string(b))
			}
		}
		client = c.setUpClientWithoutOAuth(logHookFunc)
	} else {
		secret = c.cfg.FdbCfg.ClientSecret

		client, err = c.setUpClientWithOAuth(c.cfg.FdbCfg)
		if err != nil {
			return err
		}
	}

	c.fdbClient, err = fleetDBApi.NewClientWithToken(
		secret,
		c.cfg.FdbCfg.Endpoint,
		client,
	)
	if err != nil {
		return err
	}

	return err
}

func (c *Client) newConditionOrcClient() error {
	if c.cfg == nil {
		return ErrNilConfig
	}

	var err error
	var client *http.Client
	var secret string
	if c.cfg.CoCfg.DisableOAuth {
		secret = "dummy"

		logHookFunc := func(l retryablehttp.Logger, r *http.Response) {
			// retryablehttp ignores 500 and all errors above 501. So we want to make sure those are logged.
			// https://github.com/hashicorp/go-retryablehttp/blob/4165cf8897205a879a06b20d1ed0a2a76fbb6a17/client.go#L521C80-L521C100
			if r.StatusCode == http.StatusInternalServerError || r.StatusCode > http.StatusNotImplemented {
				// named newErr so the linter doesnt get mad
				b, newErr := io.ReadAll(r.Body)
				if newErr != nil {
					c.logger.Warn("conditionOrc query returned 500 error, got error reading body: ", newErr.Error())
					return
				}

				c.logger.Warn("conditionOrc query returned 500 error, body: ", string(b))
			}
		}
		client = c.setUpClientWithoutOAuth(logHookFunc)
	} else {
		secret = c.cfg.CoCfg.ClientSecret

		client, err = c.setUpClientWithOAuth(c.cfg.CoCfg)
		if err != nil {
			return err
		}
	}

	c.coClient, err = conditionOrcApi.NewClient(
		c.cfg.CoCfg.Endpoint,
		conditionOrcApi.WithAuthToken(secret),
		conditionOrcApi.WithHTTPClient(client),
	)
	if err != nil {
		return err
	}

	return err
}

//// Client initialize helpers

func (c *Client) setUpClientWithoutOAuth(logHookFunc func(l retryablehttp.Logger, r *http.Response)) *http.Client {
	// set up client
	retryableClient := retryablehttp.NewClient()
	retryableClient.HTTPClient = otelhttp.DefaultClient // use otel client so we can collect telemetry
	// log hook fo 500 errors since the the retryablehttp client masks them
	retryableClient.ResponseLogHook = logHookFunc

	if c.logger.Level < logrus.DebugLevel {
		retryableClient.Logger = nil
	} else {
		retryableClient.Logger = c.logger
	}

	client := retryableClient.StandardClient()
	client.Timeout = timeout

	return client
}

func (c *Client) setUpClientWithOAuth(cfg *app.ConfigOIDC) (*http.Client, error) {
	provider, err := oidc.NewProvider(c.ctx, cfg.IssuerEndpoint)
	if err != nil {
		return nil, err
	}

	// setup oauth
	oauthConfig := clientcredentials.Config{
		ClientID:       cfg.ClientID,
		ClientSecret:   cfg.ClientSecret,
		TokenURL:       provider.Endpoint().TokenURL,
		Scopes:         cfg.ClientScopes,
		EndpointParams: url.Values{"audience": []string{cfg.AudienceEndpoint}},
	}
	oAuthclient := oauthConfig.Client(c.ctx)

	// set up client
	retryableClient := retryablehttp.NewClient()
	retryableClient.HTTPClient = otelhttp.DefaultClient // use otel client so we can collect telemetry
	retryableClient.HTTPClient.Transport = oAuthclient.Transport
	retryableClient.HTTPClient.Jar = oAuthclient.Jar

	if c.logger.Level < logrus.DebugLevel {
		retryableClient.Logger = nil
	} else {
		retryableClient.Logger = c.logger
	}

	client := retryableClient.StandardClient()
	client.Timeout = timeout

	return client, nil
}
