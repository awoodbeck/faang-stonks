// Package cmd adds command line support to this application. It uses Viper
// for configuration file and environment variable support and Cobra for CLI
// flag support. The end result is the application is able to use a number of
// methods to receive its configuration to support a variety of deployment
// scenarios.
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "stonks",
	Short: "Stonks helps you track your financial decision, questionable or otherwise.",
	Long: `Stonks helps you track your financial decision, questionable or otherwise.

https://www.urbandictionary.com/define.php?term=Stonks`,
	Run: rootRun,
}

func init() {
	cobra.OnInitialize(func() {
		// TODO: (adam) Add configuration file support. For the purposes of this
		// MVP, we'll stick to using the CLI and ENV for configuration.
		viper.SetEnvPrefix("stonks")
		viper.AutomaticEnv()
	})
}

func rootRun(_ *cobra.Command, _ []string) {}

// Execute Stonks
func Execute() error {
	return rootCmd.Execute()
}
