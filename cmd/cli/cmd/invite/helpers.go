//go:build cli

package invite

import (
	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
)

var rootCmd *cobra.Command

func attachInviteExtras(cmd *cobra.Command) {
	rootCmd = cmd
	rootCmd.AddCommand(newAcceptCommand())
}

func findInviteCommand() *cobra.Command {
	for _, c := range cmdpkg.RootCmd.Commands() {
		if c.Use == "invite" {
			return c
		}
	}
	return nil
}
