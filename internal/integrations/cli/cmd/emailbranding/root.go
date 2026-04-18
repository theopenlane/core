package emailbranding

import (
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/internal/integrations/cli/cmd"
)

// command is the parent `email-branding` command
var command = &cobra.Command{
	Use:   "email-branding",
	Short: "manage email branding configurations",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}
