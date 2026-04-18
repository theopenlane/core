package integration

import (
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/internal/integrations/cli/cmd"
)

// command is the parent `integration` command; subcommands register in init().
var command = &cobra.Command{
	Use:   "integration",
	Short: "configure integration providers and run provider operations",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}
