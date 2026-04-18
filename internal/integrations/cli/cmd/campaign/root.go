package campaign

import (
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/internal/integrations/cli/cmd"
)

// command is the parent `campaign` command
var command = &cobra.Command{
	Use:   "campaign",
	Short: "create campaigns (with targets) and launch campaign sends",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}
