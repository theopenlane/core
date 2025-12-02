package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/theopenlane/shared/logx"
)

const appName = "openlane"

var k *koanf.Koanf

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "A cli for interacting with the openlane core server",
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		err := initCmdFlags(cmd)
		cobra.CheckErr(err)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	go func() {
		<-ctx.Done()

		log.Info().Msg("shutting down gracefully...")
	}()

	cobra.CheckErr(rootCmd.ExecuteContext(ctx))
}

func init() {
	k = koanf.New(".") // Create a new koanf instance.

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().Bool("pretty", false, "enable pretty (human readable) logging output")
	rootCmd.PersistentFlags().Bool("debug", false, "debug logging output")
}

// initConfig reads in flags set for server startup
// all other configuration is done by the server with koanf
// refer to the README.md for more information
func initConfig() {
	if err := initCmdFlags(rootCmd); err != nil {
		log.Fatal().Err(err).Msg("error loading config")
	}

	setupLogging()
}

// initCmdFlags loads the flags from the command line into the koanf instance
func initCmdFlags(cmd *cobra.Command) error {
	return k.Load(posflag.Provider(cmd.Flags(), k.Delim(), k), nil)
}

func setupLogging() {
	level := zerolog.InfoLevel
	debug := k.Bool("debug")

	if debug {
		level = zerolog.DebugLevel
	}

	logx.Configure(logx.LoggerConfig{
		Level:         level,
		Pretty:        k.Bool("pretty"),
		Writer:        os.Stderr,
		IncludeCaller: debug,
		SetGlobal:     true,
	})
}
