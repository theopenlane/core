//go:build cli

package invite

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func inviteCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		if client == nil {
			return true, speccli.OperationOutput{}, fmt.Errorf("client is required")
		}

		result, err := createInvite(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		out, wrapErr := speccli.WrapSingleResult(result, spec.ResultPath)
		if wrapErr != nil {
			return true, speccli.OperationOutput{}, wrapErr
		}

		return true, out, nil
	}
}

func inviteDeleteHook(spec *speccli.DeleteSpec) speccli.DeletePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		if client == nil {
			return true, speccli.OperationOutput{}, fmt.Errorf("client is required")
		}

		result, err := deleteInvite(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records := []map[string]any{{spec.ResultField: result.DeleteInvite.DeletedID}}
		return true, speccli.OperationOutput{Raw: result, Records: records}, nil
	}
}

func inviteAcceptHook(_ *speccli.PrimarySpec) speccli.PrimaryPreHook {
	return func(ctx context.Context, _ *cobra.Command) error {
		payload, err := acceptInvite(ctx)
		if err != nil {
			return err
		}

		return speccli.PrintJSON(payload)
	}
}
