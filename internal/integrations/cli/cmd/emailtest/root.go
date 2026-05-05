//go:build examples

package emailtest

import (
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/internal/integrations/cli/cmd"
)

var command = &cobra.Command{
	Use:   "email-test",
	Short: "test email template rendering and delivery",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}
