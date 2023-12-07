package app

import (
	"context"

	"github.com/sirupsen/logrus"
)

type App struct {
	Ctx context.Context
	Cfg *Configuration

	Logger *logrus.Logger
}

func New(ctx context.Context, cfgFilePath string) (*App, error) {
	cfgFileBytes, err := loadConfig(cfgFilePath)
	if err != nil {
		return nil, err
	}

	app := App {
		Cfg: cfgFileBytes,
		Logger: logrus.New(),
		Ctx: ctx,
	}

	switch app.Cfg.LogLevel {
	case LogLevelDebug:
		app.Logger.Level = logrus.DebugLevel
	case LogLevelTrace:
		app.Logger.Level = logrus.TraceLevel
	default:
		app.Logger.Level = logrus.InfoLevel
	}

	return &app, nil
}