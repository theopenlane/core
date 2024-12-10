package cmd

import (
	"os"

	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/theopenlane/core/pkg/logx/consolelog"
)

const appName = "openlane"

var k *koanf.Koanf

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "A cli for interacting with the openlane core server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := initCmdFlags(cmd)
		cobra.CheckErr(err)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
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
	output := consolelog.NewConsoleWriter()
	log.Logger = zerolog.New(os.Stderr).
		With().Timestamp().
		Logger()

	// set the log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// set the log level to debug if the debug flag is set and add additional information
	if k.Bool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		//		buildInfo, _ := debug.ReadBuildInfo()

		log.Logger = log.Logger.With().
			Caller().Logger()
	}

	// pretty logging for development
	if k.Bool("pretty") {
		log.Logger = log.Output(output)
		//		log.Logger = log.Output(zerolog.ConsoleWriter{
		//			Out:        os.Stderr,
		//			TimeFormat: time.RFC3339,
		//			FormatCaller: func(i interface{}) string {
		//				return filepath.Base(fmt.Sprintf("%s", i))
		//			},
		//		})
	}
}
