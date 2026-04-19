package cmd

import (
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/internal/integrations/cli/config"
)

// AppName is the binary / application name
const AppName = "integrations"

// Output format values
const (
	TableOutput = "table"
	JSONOutput  = "json"
)

var (
	cfgFile      string
	Config       *koanf.Koanf
	configLoader *config.Loader
)

// RootCmd is the integrations CLI root command. Subcommands handle
// integration-specific setup (provider configuration, email template seeding,
// branding, etc.) and obtain an authenticated client via openlane.Connect.
var RootCmd = &cobra.Command{
	Use:   AppName,
	Short: "integrations CLI for configuring Openlane integrations",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		return configLoader.Apply(cmd)
	},
}

// Execute runs the root command
func Execute() {
	cobra.CheckErr(RootCmd.Execute())
}

// ConfigPath returns the resolved config file path
func ConfigPath() string { return cfgFile }

// RenderOutput dispatches to the requested output format handler
func RenderOutput(format string, jsonFn, tableFn func() error) error {
	switch format {
	case JSONOutput:
		if jsonFn == nil {
			return ErrUnsupportedOutputFormat
		}

		return jsonFn()
	case TableOutput:
		if tableFn == nil {
			return ErrUnsupportedOutputFormat
		}

		return tableFn()
	default:
		return ErrUnsupportedOutputFormat
	}
}

// init sets up the root command configuration and flags
func init() {
	Config = koanf.New(".")
	configLoader = config.New(config.Options{
		AppName:    AppName,
		Config:     Config,
		ConfigFile: &cfgFile,
		EnvPrefix:  "OPENLANE_",
	})

	cobra.OnInitialize(func() {
		cobra.CheckErr(configLoader.InitSources())
	})

	fs := RootCmd.PersistentFlags()

	fs.StringVar(&cfgFile, "config", "", "config file (default is "+config.DefaultConfigDir+"/"+config.DefaultConfigFile+")")

	fs.String("host", config.DefaultHost, "openlane api url")
	config.SetConfigKey(fs, "host", "openlane.host")

	fs.StringP("token", "t", "", "openlane api token or personal access token")
	config.SetConfigKey(fs, "token", "openlane.auth.token")

	fs.String("email", config.DefaultEmail, "email address for credential login")
	config.SetConfigKey(fs, "email", "openlane.auth.email")

	fs.String("password", config.DefaultPassword, "password for credential login")
	config.SetConfigKey(fs, "password", "openlane.auth.password")

	fs.String("auth-mode", config.DefaultAuthMode, "authentication mode: auto, token, or credentials")
	config.SetConfigKey(fs, "auth-mode", "openlane.auth.mode")

	fs.StringP("format", "z", TableOutput, "output format (json, table)")
	config.SetConfigKey(fs, "format", "output.format")

	fs.Bool("debug", false, "enable debug logging")
	config.SetConfigKey(fs, "debug", "logging.debug")

	fs.Bool("pretty", false, "enable pretty (human readable) logging output")
	config.SetConfigKey(fs, "pretty", "logging.pretty")
}
