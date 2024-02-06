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

	cfg, err := app.LoadConfig(cfgFile)
	if err != nil {
		return err
	}

	logger := logrus.New()
	logger.Level, err = logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return err
	}

	// Just used to verify fleet-scheduler can authenticate
	_, err = client.New(otelCtxWithCancel, cfg, logger)
	if err != nil {
		return err
	}

	// purge secrets from config before printing the config (for debug purposes)
	cfg.FdbCfg.ClientSecret = ""
	cfg.CoCfg.ClientSecret = ""

	var prettyJSON bytes.Buffer
	myJSON, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = json.Indent(&prettyJSON, myJSON, "", "\t")
	if err != nil {
		return err
	}

	logger.Info("Config: ", prettyJSON.String())

	return nil
}