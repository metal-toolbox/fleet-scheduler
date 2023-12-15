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

var cmdDummyServer = &cobra.Command{
	Use:   "dummyServer",
	Short: "gather all servers and create invetory for them",
	Run: func(cmd *cobra.Command, args []string) {
		dummyServer(cmd.Context())
	},
}

var serverCount uint

func init() {
	cmdDummyServer.PersistentFlags().UintVar(&serverCount, "serverCount", 1, "number of dummy servers you want to create")
	rootCmd.AddCommand(cmdDummyServer)
}

func dummyServer(ctx context.Context) {
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

	err = client.CreateNewDummyServers(serverCount)
	if err != nil {
		log.Fatal(err)
		return
	}

	app.Logger.Info("Task: 'CreateConditionInventoryForAllServers' complete")
}