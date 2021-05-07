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
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/awoodbeck/faang-stonks/api"
	"github.com/awoodbeck/faang-stonks/finance"
	"github.com/awoodbeck/faang-stonks/finance/iexcloud"
	"github.com/awoodbeck/faang-stonks/history/sqlite"
	"github.com/awoodbeck/faang-stonks/poll"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	rootCmd = &cobra.Command{
		Use:   "stonks",
		Short: "Stonks helps you track your financial decision, questionable or otherwise.",
		Long: `Stonks helps you track your financial decision, questionable or otherwise.

https://www.urbandictionary.com/define.php?term=Stonks`,
		Run: rootRun,
	}

	encoderCfg = zapcore.EncoderConfig{
		MessageKey: "msg",
		NameKey:    "name",

		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,

		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
)

func init() {
	cobra.OnInitialize(func() {
		// TODO: Add configuration file support. For the purposes of this MVP,
		// we'll stick to using the CLI and ENV for configuration.
		viper.SetEnvPrefix("stonks")
		viper.AutomaticEnv()
	})

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		log.Fatalf("binding flags to viper: %s", err)
	}
}

func rootRun(_ *cobra.Command, _ []string) {
	ret := 0
	defer os.Exit(ret)

	ctx, cancel := context.WithCancel(context.Background())

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

	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			zapcore.AddSync(
				&lumberjack.Logger{
					Filename:   "debug.log",
					Compress:   true,
					LocalTime:  true,
					MaxAge:     7,
					MaxBackups: 5,
					MaxSize:    100,
				},
			),
			zapcore.DebugLevel,
		),
	).Sugar()
	defer func() { _ = zl.Sync() }()

	storage, err := sqlite.New(
		sqlite.ConnMaxLifetime(-1),
		sqlite.DatabaseFile("database.sqlite"),
		sqlite.MaxIdleConnections(2),
		sqlite.Symbols(finance.DefaultSymbols),
	)
	if err != nil {
		zl.Error(err)
		ret = gracefulExit(cancel)
	}

	defer func() {
		if err := storage.Close(); err != nil {
			zl.Errorf("closing archiver: %v", err)
		}
	}()

	quotes, err := iexcloud.New(
		"token",
		iexcloud.BatchEndpoint(""),
		iexcloud.CallTimeout(0),
		iexcloud.InstrumentHTTPClient(),
	)
	if err != nil {
		zl.Error(err)
		ret = gracefulExit(cancel)
	}

	poller, err := poll.New(quotes, storage, zl)
	if err != nil {
		zl.Error(err)
		ret = gracefulExit(cancel)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		poller.Poll(ctx, time.Minute, finance.DefaultSymbols...)
		wg.Done()
	}()

	server, err := api.New(
		ctx, storage, zl,
		api.DisableInstrumentation(),
		api.IdleTimeout(0),
		api.ListenAddress(""),
		api.ReadHeaderTimeout(0),
	)
	if err != nil {
		zl.Error(err)
		ret = gracefulExit(cancel)
	}

	wg.Add(1)
	go func() {
		sErr := server.ListenAndServe()
		if sErr != nil && sErr != http.ErrServerClosed {
			zl.Errorf("API server: %v", sErr)
		}
		wg.Done()
	}()

	wg.Wait()
}

// gracefulExit cancels the context, signaling for the graceful shutdown of
// any goroutines that don't tolerate abrupt exits (e.g., SQLite), and calls
// runtime.Goexit() instead of os.Exit() to honor any pending deferred calls.
func gracefulExit(cancel context.CancelFunc) int {
	cancel()
	runtime.Goexit()
	return 1
}

// Execute on those stonks!
func Execute() error {
	return rootCmd.Execute()
}
