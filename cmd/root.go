// Package cmd adds command line support to this application. It uses Viper
// for configuration file and environment variable support and Cobra for CLI
// flag support. The end result is the application is able to use a number of
// methods to receive its configuration to support a variety of deployment
// scenarios.
package cmd

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

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
		// TODO: (adam) Add configuration file support. For the purposes of
		// this MVP, we'll stick to using the CLI and ENV for configuration.
		viper.SetEnvPrefix("stonks")
		viper.AutomaticEnv()
	})

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		log.Fatalf("binding flags to viper: %s", err)
	}
}

func rootRun(_ *cobra.Command, _ []string) {
	// TODO: spin up poller, metrics server, and API server.
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down ...")
		cancel()
	}()

	go func() { // pprof server
		// TODO: make this address configurable
		_ = http.ListenAndServe("localhost:6060", nil)
	}()
}

// Execute on those stonks!
func Execute() error {
	return rootCmd.Execute()
}
