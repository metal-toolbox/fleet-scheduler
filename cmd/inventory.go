package cmd

import (
	"log"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/metal-toolbox/fleet-scheduler/internal/app"
	"github.com/metal-toolbox/fleet-scheduler/internal/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var cmdInventory = &cobra.Command{
	Use:   "inventory",
	Short: "gather all servers and create invetory for them",
	Run: func(cmd *cobra.Command, args []string) {
		collect(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(cmdInventory)
}

func collect(ctx context.Context) {
	otelCtx, otelShutdown := otelinit.InitOpenTelemetry(ctx, "fleet-scheduler")
	defer otelShutdown(ctx)

	otelCtxWithCancel, cancelFunc := context.WithCancel(otelCtx)
	defer cancelFunc()

	app, err := app.New(otelCtxWithCancel, cfgFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	loggerEntry := app.Logger.WithFields(logrus.Fields{"component": "store.serverservice"})
	loggerEntry.Level = app.Logger.Level
	client, err := client.New(app.Ctx, app.Cfg, loggerEntry)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = client.CollectServers()
	if err != nil {
		log.Fatal(err)
		return
	}

	app.Logger.Info("collection completed")
}