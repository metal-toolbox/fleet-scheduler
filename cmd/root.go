package cmd

import (
	"fmt"
	"os"

	"github.com/metal-toolbox/fleet-scheduler/internal/version"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "fleet-scheduler",
	Short:   "execute commands to manage the fleet",
	Version: version.Current().String(),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "set config file path. Default location is in the env variable FLEET_SCHEDULER_CONFIG")
}
