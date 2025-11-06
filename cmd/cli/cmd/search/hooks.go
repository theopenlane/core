//go:build cli

package search

import (
	"context"

	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
)

func globalSearchHook(_ *speccli.PrimarySpec) speccli.PrimaryPreHook {
	return func(ctx context.Context, cmd *cobra.Command) error {
		client, err := cmdpkg.TokenAuth(ctx, cmdpkg.Config)
		if err != nil || client == nil {
			client, err = cmdpkg.SetupClientWithAuth(ctx)
			if err != nil {
				return err
			}
			defer cmdpkg.StoreSessionCookies(client)
		}

		results, err := executeSearch(ctx, client)
		if err != nil {
			return err
		}

		return renderSearchResults(results)
	}
}
