package cmd

import (
	"log"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/metal-toolbox/fleet-scheduler/internal/app"
	"github.com/metal-toolbox/fleet-scheduler/internal/client"
	"github.com/metal-toolbox/fleet-scheduler/internal/metrics"
	"github.com/metal-toolbox/fleet-scheduler/internal/version"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	pageSize      int
	inFlightPages int
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
	rootCmd.PersistentFlags().IntVar(&pageSize, "page-size", 4, "Define how many servers to query per request")
	rootCmd.PersistentFlags().IntVar(&inFlightPages, "inflight-pages", 1, "Define how many server pages to queue up before waiting for the previous to finish creating the condition")
	rootCmd.AddCommand(cmdInventory)
}

func inventory(ctx context.Context) error {
	cfg, err := app.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	logger := logrus.New()
	logger.Level, err = logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return err
	}

	metricsPusher := metrics.NewPusher(logger, "inventory")
	metricsPusher.AddCollector(collectors.NewGoCollector())
	err = metricsPusher.Start()
	if err != nil {
		return err
	}

	otelCtx, otelShutdown := otelinit.InitOpenTelemetry(ctx, "fleet-scheduler")
	defer otelShutdown(ctx)

	otelCtxWithCancel, cancelFunc := context.WithCancel(otelCtx)
	defer cancelFunc()

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

	err = newClient.CreateConditionInventoryForAllServers(pageSize, inFlightPages)
	if err != nil {
		return err
	}

	logger.Info("Task: 'CreateConditionInventoryForAllServers' complete")

	metricsPusher.KillAndWait()

	return nil
}
