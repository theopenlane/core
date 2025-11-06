//go:build cli

package control

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func controlCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := createControl(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		out, err := speccli.WrapSingleResult(result, spec.ResultPath)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		return true, out, nil
	}
}

func controlUpdateHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := updateControl(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		out, err := speccli.WrapSingleResult(result, spec.ResultPath)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		return true, out, nil
	}
}

func controlGetHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := getControl(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		switch v := result.(type) {
		case *openlaneclient.GetControlByID:
			out, wrapErr := speccli.WrapSingleResult(v, spec.ResultPath)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		case *openlaneclient.GetControls:
			out, wrapErr := speccli.WrapListResult(v, spec.ListRoot)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		case *openlaneclient.GetAllControls:
			out, wrapErr := speccli.WrapListResult(v, spec.ListRoot)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		default:
			return true, speccli.OperationOutput{}, fmt.Errorf("unexpected control get result type %T", result)
		}
	}
}
