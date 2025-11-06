//go:build cli

package invite

import (
	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
)

func newAcceptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept",
		Short: "accept an invite to join an organization",
		RunE: func(c *cobra.Command, _ []string) error {
			payload, err := acceptInvite(c.Context())
			if err != nil {
				return err
			}

			return cmdpkg.JSONPrint(payload)
		},
	}

	cmd.Flags().StringP("token", "t", "", "invite token")

	return cmd
}
