package client

import (
	"context"
	"time"

	// "github.com/metal-toolbox/fleet-scheduler/internal/model"
	"github.com/metal-toolbox/fleet-scheduler/internal/util"

	// "github.com/pkg/errors"
	conditionorcapi "github.com/metal-toolbox/conditionorc/pkg/api/v1/client"
	"github.com/sirupsen/logrus"
	serverserviceapi "go.hollow.sh/serverservice/pkg/api/v1"
)

const timeout = 30 * time.Second

type Client struct {
	ssClient *serverserviceapi.Client
	coClient *conditionorcapi.Client
	cfg* util.Configuration
	ctx context.Context
	logger *logrus.Entry
}

func New(ctx context.Context, cfg* util.Configuration, logger *logrus.Entry) (*Client, error) {
	return &Client{
		ssClient: nil,
		coClient: nil,
		cfg: cfg,
		ctx: ctx,
		logger: logger,
	}, nil
}

func (c *Client) initServerServiceClient() error {
	if c.ssClient == nil {
		var ssClient *serverserviceapi.Client
		var err error

		if c.cfg.SsCfg.DisableOAuth {
			ssClient, err = newServerserviceClientWithoutOAuth(c.cfg.SsCfg, c.logger)
		} else {
			ssClient, err = newServerserviceClient(c.ctx, c.cfg.SsCfg, c.logger)
		}

		if err != nil {
			return err
		}

		c.ssClient = ssClient
		return err
	} else {
		return nil
	}
}

// TODO: Fix authentication for conditions
// func (c *Client) initConditionsClient() error {
// 	if c.coClient == nil {
// 		var coClient *conditionorcapi.Client = nil
// 		var err error = nil

// 		if c.cfg.SsCfg.DisableOAuth {
// 			coClient, err = conditionorcapi.NewClient(c.cfg.SsCfg.Endpoint)
// 		} else {
// 			token, err := util.AccessToken(c.ctx, model.ConditionsAPI, c.SsCfg.cfg, true)
// 			if err != nil {
// 				return errors.Wrap(ErrAuth, string(model.ConditionsAPI) + err.Error())
// 			}
// 			coClient, err = conditionorcapi.NewClient(c.SsCfg.cfg.Endpoint, conditionorcapi.WithAuthToken(token))
// 		}

// 		c.coClient = coClient
// 		return err
// 	} else {
// 		return nil
// 	}
// }