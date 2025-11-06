//go:build cli

package orgmembers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func orgMembershipCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := createOrgMembership(ctx, client)
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

func orgMembershipUpdateHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := updateOrgMembership(ctx, client)
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

func orgMembershipDeleteHook(spec *speccli.DeleteSpec) speccli.DeletePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := deleteOrgMembership(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records := []map[string]any{{spec.ResultField: result.DeleteOrgMembership.DeletedID}}
		return true, speccli.OperationOutput{Raw: result, Records: records}, nil
	}
}

func orgMembershipGetHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := fetchOrgMemberships(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		switch v := result.(type) {
		case *openlaneclient.GetOrgMemberships:
			out, wrapErr := speccli.WrapListResult(v, spec.ListRoot)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		default:
			out, wrapErr := speccli.WrapListResult(result, spec.ListRoot)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		}
	}
}
