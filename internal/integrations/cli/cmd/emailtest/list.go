package emailtest

import (
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/internal/integrations/cli/cmd"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all registered email template dispatchers",
	RunE: func(c *cobra.Command, _ []string) error {
		return list()
	},
}

func init() {
	command.AddCommand(listCmd)
}

// list prints a table of all registered email dispatchers
func list() error {
	ops := email.AllEmailOperations()

	headers := []string{"Name", "Description", "CustomerSelectable"}
	rows := make([][]string, 0, len(ops))

	for _, op := range ops {
		rows = append(rows, []string{
			op.Name,
			op.Description,
			cmd.BoolStr(op.CustomerSelectable),
		})
	}

	return cmd.RenderTable(ops, headers, rows)
}
