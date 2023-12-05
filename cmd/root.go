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

var RootCmd = &cobra.Command{
	Use:   "fleet-scheduler",
	Short: "execute commands to manage the fleet",

	DisableAutoGenTag: true,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/mctl/config.yml)")
	RootCmd.PersistentFlags().BoolVar(&reAuth, "reauth", false, "re-authenticate with oauth services")
}
