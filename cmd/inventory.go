package cmd

import (
	"log"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/metal-toolbox/fleet-scheduler/internal/app"
	"github.com/metal-toolbox/fleet-scheduler/internal/client"
	"github.com/metal-toolbox/fleet-scheduler/internal/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var cmdInventory = &cobra.Command{
	Use:     "inventory",
	Short:   "gather all servers and create invetory for them",
	Version: version.Current().String(),
	Run: func(cmd *cobra.Command, _ []string) {
		err := inventory(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cmdInventory)
}

func inventory(ctx context.Context) error {
	otelCtx, otelShutdown := otelinit.InitOpenTelemetry(ctx, "fleet-scheduler")
	defer otelShutdown(ctx)

	otelCtxWithCancel, cancelFunc := context.WithCancel(otelCtx)
	defer cancelFunc()

	cfg, err := app.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	logger := logrus.New()
	logger.Level, err = logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return err
	}

	v := version.Current()
	logger.WithFields(logrus.Fields{
		"GitCommit":             v.GitCommit,
		"AppVersion":            v.AppVersion,
		"ServerServiceVersion:": v.ServerserviceVersion, // TODO; Swap out with fleetdb once migrated to fleetdb
		"ConditionOrcVersion:":  v.ConditionorcVersion,
	}).Info("running task: inventory")

	newClient, err := client.New(otelCtxWithCancel, cfg, logger)
	if err != nil {
		return err
	}

	err = newClient.CreateConditionInventoryForAllServers()
	if err != nil {
		return err
	}

	logger.Info("Task: 'CreateConditionInventoryForAllServers' complete")

	return nil
}
