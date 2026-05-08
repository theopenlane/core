//go:build examples

package emailtemplate

import (
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/internal/integrations/cli/cmd"
)

// command is the parent `email-template` command
var command = &cobra.Command{
	Use:   "email-template",
	Short: "manage email templates",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}
