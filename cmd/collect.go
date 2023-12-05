package cmd

import (
	"log"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/metal-toolbox/fleet-scheduler/internal/app"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var cmdCollect = &cobra.Command{
	Use:   "collect",
	Short: "collect em",
	Run: func(cmd *cobra.Command, args []string) {
		collect(cmd.Context())
	},
}

func init() {
	RootCmd.AddCommand(cmdCollect)
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

	client, err := app.NewClient()
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