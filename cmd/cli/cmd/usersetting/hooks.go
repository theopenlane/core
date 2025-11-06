//go:build cli

package usersetting

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func userSettingGetHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := getUserSettings(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		switch v := result.(type) {
		case *openlaneclient.GetUserSettingByID:
			out, wrapErr := speccli.WrapSingleResult(v, spec.ResultPath)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		case *openlaneclient.GetAllUserSettings:
			out, wrapErr := speccli.WrapListResult(v, spec.ListRoot)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		default:
			return true, speccli.OperationOutput{}, fmt.Errorf("unexpected usersetting get result type %T", result)
		}
	}
}

func userSettingUpdateHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := updateUserSetting(ctx, client)
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
