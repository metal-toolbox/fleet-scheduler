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
		inventory(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(cmdInventory)
}

func inventory(ctx context.Context) {
	otelCtx, otelShutdown := otelinit.InitOpenTelemetry(ctx, "fleet-scheduler")
	defer otelShutdown(ctx)

	otelCtxWithCancel, cancelFunc := context.WithCancel(otelCtx)
	defer cancelFunc()

	newApp, err := app.New(otelCtxWithCancel, cfgFile)
	if err != nil {
		log.Print(err)
		return
	}

	loggerEntry := newApp.Logger.WithFields(logrus.Fields{"component": "store.serverservice"})
	loggerEntry.Level = newApp.Logger.Level
	newClient, err := client.New(newApp.Ctx, newApp.Cfg, loggerEntry)
	if err != nil {
		log.Print(err)
		return
	}

	err = newClient.CreateConditionInventoryForAllServers()
	if err != nil {
		log.Print(err)
		return
	}

	newApp.Logger.Info("Task: 'CreateConditionInventoryForAllServers' complete")
}
