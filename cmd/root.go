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
	"strings"
	"sync"
	"syscall"

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
		PreRun: rootPreRun,
		Run:    rootRun,
	}

	encoderCfg = zapcore.EncoderConfig{
		MessageKey: "msg",
		NameKey:    "name",

		LevelKey:    "level",
		EncodeLevel: zapcore.LowercaseLevelEncoder,

		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	logOutput zapcore.WriteSyncer = os.Stdout
)

func init() {
	cobra.OnInitialize(func() {
		// For the purposes of this MVP, we'll stick to configuring this app
		// using the CLI and ENV. But Viper allows us to use config files,
		// key-value data stores, etc., for configuration.
		replacer := strings.NewReplacer("-", "_")
		viper.SetEnvKeyReplacer(replacer)
		viper.SetEnvPrefix("stonks")
		viper.AutomaticEnv()
	})

	// API server settings
	rootCmd.Flags().Duration("api-idle-timeout", api.DefaultIdleTimeout, "duration clients are allowed to idle")
	rootCmd.Flags().StringP("a", "api-listen-addr", api.DefaultListenAddress, "API server host:port")
	rootCmd.Flags().Bool("api-metrics", true, "enable metrics for the API server")
	rootCmd.Flags().Duration("api-read-headers-timeout", api.DefaultReadHeaderTimeout, "duration clients have to send request headers")

	// IEX Cloud API client settings
	rootCmd.Flags().String("iex-batch-endpoint", iexcloud.DefaultBatchEndpoint, "IEX Cloud API batch endpoint URL")
	rootCmd.Flags().Duration("iex-call-timeout", iexcloud.DefaultTimeout, "API call timeout")
	rootCmd.Flags().Bool("iex-metrics", false, "collect metrics for IEX Cloud API calls")
	rootCmd.Flags().StringP("t", "iex-token", "", "IEX Cloud API token")

	// Logger settings
	rootCmd.Flags().StringP("l", "log", "stdout", "log file path")
	rootCmd.Flags().Bool("log-compress", false, "compress rotated log files")
	rootCmd.Flags().Bool("log-localtime", false, "log file names use local time, UTC otherwise")
	rootCmd.Flags().Int("log-max-age", 7, "max days to retain old log files")
	rootCmd.Flags().Int("log-max-backups", 5, "max number of old log files to retain")
	rootCmd.Flags().Int("log-max-size", 100, "max log file size in MB before rotation")

	// SQLite settings
	rootCmd.Flags().Duration("sqlite-conn-max-lifetime", sqlite.DefaultConnsMaxLifetime, "max client connection lifetime")
	rootCmd.Flags().StringP("d", "sqlite-database", sqlite.DefaultDatabaseFile, "database file path")
	rootCmd.Flags().Int("sqlite-max-idle-conn", sqlite.DefaultMaxIdleConns, "max idle client connections")

	// General settings
	rootCmd.Flags().DurationP("p", "poll", poll.DefaultPollDuration, "duration between stock quote updates")
	rootCmd.Flags().String("pprof-addr", "localhost:6060", "pprof host:port")
	rootCmd.Flags().StringSliceP("s", "symbols", finance.DefaultSymbols, "stock symbols")

	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		log.Fatalf("binding flags to viper: %s", err)
	}
}

func rootPreRun(_ *cobra.Command, _ []string) {
	if viper.GetString("iex-token") == "" {
		log.Fatal("IEX Cloud API token not set")
	}

	switch strings.ToLower(viper.GetString("log")) {
	case "stdout", "":
	default:
		logOutput = zapcore.AddSync(
			&lumberjack.Logger{
				Filename:   viper.GetString("log"),
				Compress:   viper.GetBool("log-compress"),
				LocalTime:  viper.GetBool("log-localtime"),
				MaxAge:     viper.GetInt("log-max-age"),
				MaxBackups: viper.GetInt("log-max-backups"),
				MaxSize:    viper.GetInt("log-max-size"),
			},
		)
	}
}

func rootRun(_ *cobra.Command, _ []string) {
	ret := 0
	defer os.Exit(ret)

	ctx, cancel := context.WithCancel(context.Background())
	zl := zap.New(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg),
			logOutput,
			zapcore.DebugLevel,
		),
	).Sugar()
	defer func() { _ = zl.Sync() }()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		zl.Info("shutting down ...")
		cancel()
	}()

	go func() { // pprof server
		_ = http.ListenAndServe(viper.GetString("pprof-addr"), nil)
	}()

	storage, err := sqlite.New(
		sqlite.ConnMaxLifetime(viper.GetDuration("sqlite-conn-max-lifetime")),
		sqlite.DatabaseFile(viper.GetString("sqlite-database")),
		sqlite.MaxIdleConnections(viper.GetInt("sqlite-max-idle-conn")),
		sqlite.Symbols(viper.GetStringSlice("symbols")),
	)
	if err != nil {
		zl.Error(err)
		gracefulExit(cancel, &ret)
	}

	defer func() {
		if err := storage.Close(); err != nil {
			zl.Errorf("closing archiver: %v", err)
		}
	}()

	var iexMetrics iexcloud.Option
	if viper.GetBool("iex-metrics") {
		iexMetrics = iexcloud.InstrumentHTTPClient()
	}

	quotes, err := iexcloud.New(
		viper.GetString("iex-token"),
		iexcloud.BatchEndpoint(viper.GetString("iex-batch-endpoint")),
		iexcloud.CallTimeout(viper.GetDuration("iex-call-timeout")),
		iexMetrics,
	)
	if err != nil {
		zl.Error(err)
		gracefulExit(cancel, &ret)
	}

	poller, err := poll.New(quotes, storage, zl)
	if err != nil {
		zl.Error(err)
		gracefulExit(cancel, &ret)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		poller.Poll(
			ctx,
			viper.GetDuration("poll"),
			viper.GetStringSlice("symbols")...,
		)
		wg.Done()
	}()

	var apiMetrics api.Option
	if !viper.GetBool("api-metrics") {
		apiMetrics = api.DisableInstrumentation()
	}
	server, err := api.New(
		ctx, storage, zl,
		apiMetrics,
		api.IdleTimeout(viper.GetDuration("api-idle-timeout")),
		api.ListenAddress(viper.GetString("api-listen-addr")),
		api.ReadHeaderTimeout(viper.GetDuration("api-read-headers-timeout")),
	)
	if err != nil {
		zl.Error(err)
		gracefulExit(cancel, &ret)
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

// gracefulExit sets the exit code and cancels the context, signaling for the
// graceful shutdown of any goroutines that don't tolerate abrupt exits (e.g.,
// SQLite), and calls runtime.Goexit() instead of os.Exit() to honor any
// pending deferred calls.
//
// Ultimately, the calling goroutine will need to run os.Exit() to terminate
// the application.
func gracefulExit(cancel context.CancelFunc, ret *int) {
	cancel()
	*ret = 1
	runtime.Goexit()
}

// Execute on those stonks!
func Execute() error {
	return rootCmd.Execute()
}
