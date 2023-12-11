package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	reAuth  bool
)

var rootCmd = &cobra.Command{
	Use:   "fleet-scheduler",
	Short: "execute commands to manage the fleet",

	DisableAutoGenTag: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/mctl/config.yml)")
	rootCmd.PersistentFlags().BoolVar(&reAuth, "reauth", false, "re-authenticate with oauth services")
}
