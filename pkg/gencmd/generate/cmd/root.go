//go:build gencmd

package cmd

import (
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

var (
	Config *koanf.Koanf
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "",
	Short: "generate the stub files for a given cli cmd",
	PersistentPreRun: func(cmd *cobra.Command, _ []string) {
		initConfiguration(cmd)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

func init() {
	Config = koanf.New(".") // Create a new koanf instance.
}

// initConfiguration loads the configuration from the command flags of the given cobra command
func initConfiguration(cmd *cobra.Command) {
	loadFlags(cmd)
}

func loadFlags(cmd *cobra.Command) {
	err := Config.Load(posflag.Provider(cmd.Flags(), Config.Delim(), Config), nil)

	cobra.CheckErr(err)
}
