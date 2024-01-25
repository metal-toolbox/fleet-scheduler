package cmd

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/metal-toolbox/fleet-scheduler/internal/app"
	"github.com/metal-toolbox/fleet-scheduler/internal/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var cmdTest = &cobra.Command{
	Use:   "test",
	Short: "test",
	Run: func(cmd *cobra.Command, args []string) {
		err := test(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cmdTest)
}

// This test command will be purged once I verify everything is functional
func test(ctx context.Context) error {
	otelCtx, otelShutdown := otelinit.InitOpenTelemetry(ctx, "fleet-scheduler")
	defer otelShutdown(ctx)

	otelCtxWithCancel, cancelFunc := context.WithCancel(otelCtx)
	defer cancelFunc()

	newApp, err := app.New(otelCtxWithCancel, cfgFile)
	if err != nil {
		return err
	}

	loggerEntry := newApp.Logger.WithFields(logrus.Fields{"component": "store.serverservice"})
	loggerEntry.Level = newApp.Logger.Level

	// Just used to verify fleet-scheduler can authenticate
	_, err = client.New(newApp.Ctx, newApp.Cfg, loggerEntry)
	if err != nil {
		return err
	}

	// purge secrets from config before printing the config (for debug purposes)
	newApp.Cfg.FdbCfg.ClientSecret = ""
	newApp.Cfg.CoCfg.ClientSecret = ""

	var prettyJSON bytes.Buffer
	myJSON, err := json.Marshal(newApp.Cfg)
	if err != nil {
		return err
	}

	err = json.Indent(&prettyJSON, myJSON, "", "\t")
	if err != nil {
		return err
	}

	newApp.Logger.Info("Config: ", prettyJSON.String())

	return nil
}
