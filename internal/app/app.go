package app

import (
	"context"

	"github.com/metal-toolbox/fleet-scheduler/internal/client"
	"github.com/metal-toolbox/fleet-scheduler/internal/util"
	"github.com/sirupsen/logrus"
)

type App struct {
	Ctx context.Context
	Cfg *util.Configuration

	Logger *logrus.Logger
}

func New(ctx context.Context, cfgFilePath string) (*App, error) {
	cfgFileBytes, err := util.LoadConfig(cfgFilePath)
	if err != nil {
		return nil, err
	}

	// TODO
	// err = config.ValidateClientParams(cfgFileBytes)
	// if err != nil {
	// 	return nil, err
	// }

	app := App {
		Cfg: cfgFileBytes,
		Logger: logrus.New(),
		Ctx: ctx,
	}
	app.Logger.Level = util.StringToLogrusLogLevel("info")

	return &app, nil
}

func (a *App) NewClient() (*client.Client, error) {
	loggerEntry := util.NewLogrusEntry(
		logrus.Fields{"component": "store.serverservice"},
		a.Logger,
	)

	new_client, err := client.New(a.Ctx, a.Cfg, loggerEntry)
	if err != nil {
		return new_client, err
	}

	return new_client, nil
}