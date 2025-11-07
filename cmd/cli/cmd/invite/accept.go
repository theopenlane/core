//go:build cli

package invite

import (
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
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

			return speccli.PrintJSON(payload)
		},
	}

	cmd.Flags().StringP("token", "t", "", "invite token")

	return cmd
}
