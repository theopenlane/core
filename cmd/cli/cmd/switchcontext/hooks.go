//go:build cli

package switchcontext

import (
	"context"

	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
)

func switchContextHook(_ *speccli.PrimarySpec) speccli.PrimaryPreHook {
	return func(ctx context.Context, cmd *cobra.Command) error {
		client, err := cmdpkg.SetupClientWithAuth(ctx)
		if err != nil {
			return err
		}

		resp, err := switchOrganization(ctx, client)
		if err != nil {
			return err
		}

		cmd.Printf("Successfully switched to organization: %s!\n", cmdpkg.Config.String("target-org"))
		cmd.Println("auth tokens successfully stored in keychain")

		_ = resp // nothing to render beyond messages
		return nil
	}
}
