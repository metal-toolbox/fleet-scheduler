package client

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/sync/semaphore"

	"github.com/metal-toolbox/fleet-scheduler/internal/util"
	serverserviceapi "go.hollow.sh/serverservice/pkg/api/v1"
)

const (
	// server service attribute to look up the BMC IP Address in
	bmcAttributeNamespace = "sh.hollow.bmc_info"
)

func newServerserviceClient(ctx context.Context, cfg *util.ServerserviceConfig, logger *logrus.Entry) (*serverserviceapi.Client, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	provider, err := oidc.NewProvider(ctx, cfg.OidcIssuerEndpoint)
	if err != nil {
		return nil, err
	}

	var clientID string
	if cfg.OidcClientID != "" {
		clientID = cfg.OidcClientID
	} else {
		clientID = "fleet-scheduler"
	}

	// setup oauth
	oauthConfig := clientcredentials.Config{
		ClientID:       clientID,
		ClientSecret:   cfg.OidcClientSecret,
		TokenURL:       provider.Endpoint().TokenURL,
		Scopes:         cfg.OidcClientScopes,
		EndpointParams: url.Values{"audience": []string{cfg.OidcAudienceEndpoint}},
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

	return serverserviceapi.NewClientWithToken(
		cfg.OidcClientSecret,
		cfg.Endpoint,
		client,
	)
}

func newServerserviceClientWithoutOAuth(cfg *util.ServerserviceConfig, logger *logrus.Entry) (*serverserviceapi.Client, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	// set up client
	retryableClient := retryablehttp.NewClient()
	retryableClient.HTTPClient = otelhttp.DefaultClient // use otel client so we can collect telemetry
	retryableClient.ResponseLogHook = func(l retryablehttp.Logger, r *http.Response) {
		if r.StatusCode == http.StatusInternalServerError {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Warn("serverservice query returned 500 error, got error reading body: ", err.Error())
				return
			}

			logger.Warn("serverservice query returned 500 error, body: ", string(b))
		}
	}

	if logger.Level < logrus.DebugLevel {
		retryableClient.Logger = nil
	} else {
		retryableClient.Logger = logger
	}

	client := retryableClient.StandardClient()
	client.Timeout = timeout

	return serverserviceapi.NewClientWithToken(
		"dummy",
		cfg.Endpoint,
		client,
	)
}

func (c *Client) AsyncGatherServers(page_size int) (chan serverserviceapi.Server, *semaphore.Weighted, error) {
	if c.ssClient == nil {
		return nil, nil, ErrSsClientIsNil
	}

	serverCh := make(chan serverserviceapi.Server)
	concLimiter := semaphore.NewWeighted(int64(page_size*page_size))

	go func() {
		c.gatherServers(page_size, serverCh, concLimiter)
	}()

	return serverCh, concLimiter, nil
}

func (c* Client) gatherServers(page_size int, serverCh chan serverserviceapi.Server, concLimiter *semaphore.Weighted) {
	// signal to reciever that we are done
	defer close(serverCh)

	// First page, use the response from it to figure out how many pages we have to loop through
	// Dont change page size
	servers, response, err := c.getServerPage(page_size, 1)
	if err != nil {
		c.logger.WithFields(logrus.Fields {
				"page_size": page_size,
				"page_index": 1,
		}).Logger.Error("Failed to get list of servers")
		return
	}
	total_pages := response.TotalPages

	if !concLimiter.TryAcquire(int64(response.PageSize)) {
		c.logger.Error("Failed to acquire semaphore! Going to attempt to continue.")
	}

	// send first page of servers to the channel
	for _, server := range(servers) {
		serverCh <- server
	}

	c.logger.WithFields(logrus.Fields{
		"index":      1,
		"iterations": total_pages,
		"got"       : len(servers),
	}).Trace("Got server page")

	// Start the second page, and loop through rest the pages
	for i := 2; i <= total_pages; i++ {
		servers, response, err = c.getServerPage(page_size, i)
		if err != nil {
			c.logger.WithFields(logrus.Fields {
				"page_size": page_size,
				"page_index": i,
			}).Logger.Error("Failed to get page of servers")

			continue
		}

		c.logger.WithFields(logrus.Fields{
			"index":      i,
			"iterations": total_pages,
			"got"       : len(servers),
		}).Trace("Got server page")

		// throttle this loop
		// Doing a spinlock to prevent a permanent lock if the ctx gets canceled
		for !concLimiter.TryAcquire(int64(response.PageSize)) && c.ctx.Err() == nil {
			time.Sleep(time.Second)
		}

		for _, server := range(servers) {
			serverCh <- server
		}
	}
}

func (c* Client) getServerPage(page_size int, page int) ([]serverserviceapi.Server, *serverserviceapi.ServerResponse, error) {
	params := &serverserviceapi.ServerListParams{
		FacilityCode: c.cfg.FacilityCode,
		AttributeListParams: []serverserviceapi.AttributeListParams{
			{
				Namespace: bmcAttributeNamespace,
			},
		},
		PaginationParams: &serverserviceapi.PaginationParams{
			Limit: page_size,
			Page:  page,
		},
	}

	return c.ssClient.List(c.ctx, params)
}