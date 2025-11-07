//go:build cli

package trustcenternda

import (
	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
)

var rootCmd *cobra.Command

func findTrustCenterNdaCommand() *cobra.Command {
	for _, c := range cmdpkg.RootCmd.Commands() {
		if c.Use == "trust-center-nda" {
			return c
		}
	}
	return nil
}

func attachTrustCenterNdaExtras(cmd *cobra.Command) {
	rootCmd = cmd
	rootCmd.AddCommand(newSubmitCommand())
	rootCmd.AddCommand(newSendEmailCommand())
}
