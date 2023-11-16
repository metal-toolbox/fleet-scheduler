package app

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type App struct {
	logger *logrus.Logger
}

func New() (*App, error) {
	var logger = logrus.New()
	logger.Out = os.Stdout

	switch logLevel {
	case LogLevelDebug:
		logger.SetLevel(logrus.DebugLevel)
	case LogLevelTrace:
		logger.SetLevel(logrus.TraceLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Load up configs
	v := viper.New()
	v.SetConfigFile("config.yaml")
	v.AddConfigPath(".")
	err := v.ReadInConfig()

	if err != nil {
		logger.Error("Failed to find viper config file")
	}

	return &App {

	}
}