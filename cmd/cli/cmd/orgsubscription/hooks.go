//go:build cli

package orgsubscription

import (
	"context"

	"github.com/spf13/cobra"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func orgSubscriptionGetHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		if client == nil {
			var err error
			client, err = acquireClient(ctx)
			if err != nil {
				return true, speccli.OperationOutput{}, err
			}
		}

		result, err := fetchOrgSubscriptions(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		if err := renderOrgSubscriptions(result); err != nil {
			return true, speccli.OperationOutput{}, err
		}

		return true, speccli.OperationOutput{Raw: result}, nil
	}
}

func acquireClient(ctx context.Context) (*openlaneclient.OpenlaneClient, error) {
	client, err := cmdpkg.TokenAuth(ctx, cmdpkg.Config)
	if err != nil || client == nil {
		client, err = cmdpkg.SetupClientWithAuth(ctx)
		if err != nil {
			return nil, err
		}
		cmdpkg.StoreSessionCookies(client)
	}

	return client, nil
}
