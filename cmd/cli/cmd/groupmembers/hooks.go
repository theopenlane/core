//go:build cli

package groupmembers

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func groupMembershipCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := createGroupMembership(ctx, client)
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

func groupMembershipUpdateHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := updateGroupMembership(ctx, client)
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

func groupMembershipDeleteHook(spec *speccli.DeleteSpec) speccli.DeletePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := deleteGroupMembership(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records := []map[string]any{{spec.ResultField: result.DeleteGroupMembership.DeletedID}}
		return true, speccli.OperationOutput{Raw: result, Records: records}, nil
	}
}

func groupMembershipGetHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := fetchGroupMemberships(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		switch v := result.(type) {
		case *openlaneclient.GetGroupMembershipByID:
			out, wrapErr := speccli.WrapSingleResult(v, spec.ResultPath)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		case *openlaneclient.GetGroupMemberships:
			out, wrapErr := speccli.WrapListResult(v, spec.ListRoot)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		case *openlaneclient.GetAllGroupMemberships:
			out, wrapErr := speccli.WrapListResult(v, spec.ListRoot)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		default:
			return true, speccli.OperationOutput{}, nil
		}
	}
}
