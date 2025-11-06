//go:build cli

package mappedcontrol

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func mappedControlCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := createMappedControl(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records, recordErr := mappedControlRecords(result)
		if recordErr != nil {
			return true, speccli.OperationOutput{}, recordErr
		}

		return true, speccli.OperationOutput{Raw: result, Records: records}, nil
	}
}

func mappedControlUpdateHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := updateMappedControl(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records, recordErr := mappedControlRecords(result)
		if recordErr != nil {
			return true, speccli.OperationOutput{}, recordErr
		}

		return true, speccli.OperationOutput{Raw: result, Records: records}, nil
	}
}

func mappedControlDeleteHook(spec *speccli.DeleteSpec) speccli.DeletePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := deleteMappedControl(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records := []map[string]any{{spec.ResultField: result.DeleteMappedControl.DeletedID}}
		return true, speccli.OperationOutput{Raw: result, Records: records}, nil
	}
}

func mappedControlGetHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := fetchMappedControls(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records, recordErr := mappedControlRecords(result)
		if recordErr != nil {
			return true, speccli.OperationOutput{}, recordErr
		}

		return true, speccli.OperationOutput{Raw: result, Records: records}, nil
	}
}
