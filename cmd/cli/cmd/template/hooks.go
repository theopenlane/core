//go:build cli

package templates

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func templateCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := createTemplate(ctx, client)
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

func templateGetHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := getTemplates(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		switch v := result.(type) {
		case *openlaneclient.GetTemplateByID:
			out, wrapErr := speccli.WrapSingleResult(v, spec.ResultPath)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		case *openlaneclient.GetAllTemplates:
			out, wrapErr := speccli.WrapListResult(v, spec.ListRoot)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		default:
			return true, speccli.OperationOutput{}, fmt.Errorf("unexpected template get result type %T", result)
		}
	}
}
