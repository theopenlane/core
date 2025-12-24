//go:build cli

// Package cmd is the cobra cli implementation for the core server
package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	appName         = "openlane"
	defaultRootHost = "http://localhost:17608"
	TableOutput     = "table"
	JSONOutput      = "json"
)

var (
	cfgFile      string
	OutputFormat string
	InputFile    string
	Config       *koanf.Koanf

	// Pagination options
	// First is the first number of items to return in a paginated response, max of 100
	First *int64
	first int64
	// Last is the last number of items to return in a paginated response, max of 100
	Last *int64
	last int64
	// After is the cursor to start after in a paginated response
	After *string
	after string
	// Before is the cursor to end before in a paginated response
	Before *string
	before string
	// OrderBy is the field to order by in a paginated response
	OrderBy *string
	orderBy string
	// OrderDirection is the direction to order by in a paginated response
	OrderDirection *string
	orderDirection string
)

var (
	// RootHost contains the root url for the API
	RootHost string
	// GraphAPIHost contains the url for the graph api
	GraphAPIHost string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   appName,
	Short: "the openlane cli",
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		initConfiguration(cmd)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

func init() {
	Config = koanf.New(".") // Create a new koanf instance.

	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/."+appName+".yaml)")
	RootCmd.PersistentFlags().String("host", defaultRootHost, "api host url")

	// CSRF flags
	RootCmd.PersistentFlags().Bool("disable-csrf", false, "disable CSRF client, can only be used if the server has disabled CSRF protection, not recommended for production use")

	// Token flags
	RootCmd.PersistentFlags().String("token", "", "api token used for authentication, takes precedence over other auth methods")
	RootCmd.PersistentFlags().String("pat", "", "personal access token used for authentication")
	RootCmd.PersistentFlags().String("jwt", "", "jwt used for authentication")

	// Logging flags
	RootCmd.PersistentFlags().Bool("debug", false, "enable debug logging")
	RootCmd.PersistentFlags().Bool("pretty", false, "enable pretty (human readable) logging output")

	// Output flags
	RootCmd.PersistentFlags().StringVarP(&OutputFormat, "format", "z", TableOutput, "output format (json, table)")
	RootCmd.PersistentFlags().StringVar(&InputFile, "csv", "", "csv input file instead of stdin")

	// Pagination flags
	RootCmd.PersistentFlags().Int64Var(&first, "first", 0, "first number of items to return in a paginated response, max of 100")
	RootCmd.PersistentFlags().Int64Var(&last, "last", 0, "last number of items to return in a paginated response, max of 100")
	RootCmd.PersistentFlags().StringVar(&after, "after", "", "cursor to start after in a paginated response")
	RootCmd.PersistentFlags().StringVar(&before, "before", "", "cursor to end before in a paginated response")
	RootCmd.PersistentFlags().StringVar(&orderBy, "order-by", "", "order by field (due, created_at, updated_at)")
	RootCmd.PersistentFlags().StringVar(&orderDirection, "order-direction", "DESC", "order direction (ASC, DESC)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// load the flags to ensure we know the correct config file path
	initConfiguration(RootCmd)

	// load the config file and env vars
	loadConfigFile()

	// reload because flags and env vars take precedence over file
	initConfiguration(RootCmd)

	// set the host url
	RootHost = Config.String("host")

	// set the pagination options based on the flags
	// if both first and last are set, we use first, otherwise we use last
	// we don't need to set defaults as the API will handle this
	if after != "" {
		After = &after
	} else if before != "" {
		Before = &before
	}

	if first != 0 {
		First = &first
	} else if last != 0 {
		Last = &last
	}

	if orderBy != "" {
		OrderBy = &orderBy
	}

	if orderDirection != "" {
		OrderDirection = &orderDirection
	}

	// setup logging configuration
	setupLogging()
}

// setupLogging configures the logger based on the command flags
func setupLogging() {
	log.Logger = zerolog.New(os.Stderr).
		With().Timestamp().
		Logger().
		With().Str("app", appName).
		Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if Config.Bool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if Config.Bool("pretty") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

// initConfiguration loads the configuration from the command flags of the given cobra command
func initConfiguration(cmd *cobra.Command) {
	loadEnvVars()

	loadFlags(cmd)
}

func loadConfigFile() {
	if cfgFile == "" {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		cfgFile = filepath.Join(home, "."+appName+".yaml")
	}

	// If the config file does not exist, do nothing
	if _, err := os.Stat(cfgFile); errors.Is(err, os.ErrNotExist) {
		return
	}

	err := Config.Load(file.Provider(cfgFile), yaml.Parser())

	cobra.CheckErr(err)
}

func loadEnvVars() {
	err := Config.Load(env.Provider(".", env.Opt{
		Prefix: "CORE_",
		TransformFunc: func(key, v string) (string, any) {
			key = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(key, "CORE_")), "_", ".")

			if strings.Contains(v, ",") {
				return key, strings.Split(v, ",")
			}

			return key, v
		},
	}), nil)

	cobra.CheckErr(err)
}

func loadFlags(cmd *cobra.Command) {
	err := Config.Load(posflag.Provider(cmd.Flags(), Config.Delim(), Config), nil)

	cobra.CheckErr(err)
}
