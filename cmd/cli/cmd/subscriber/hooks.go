//go:build cli

package subscribers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func subscriberCreateHook(spec *speccli.CreateSpec) speccli.CreatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := createSubscriber(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records, recordErr := subscriberRecordsFromResult(result)
		if recordErr != nil {
			return true, speccli.OperationOutput{}, recordErr
		}

		return true, speccli.OperationOutput{Raw: result, Records: records}, nil
	}
}

func subscriberUpdateHook(spec *speccli.UpdateSpec) speccli.UpdatePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := updateSubscriber(ctx, client)
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

func subscriberDeleteHook(spec *speccli.DeleteSpec) speccli.DeletePreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := deleteSubscriber(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records := []map[string]any{{spec.ResultField: result.DeleteSubscriber.Email}}
		return true, speccli.OperationOutput{Raw: result, Records: records}, nil
	}
}

func subscriberGetHook(spec *speccli.GetSpec) speccli.GetPreHook {
	return func(ctx context.Context, _ *cobra.Command, client *openlaneclient.OpenlaneClient) (bool, speccli.OperationOutput, error) {
		result, err := fetchSubscribers(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		switch v := result.(type) {
		case *openlaneclient.GetSubscriberByEmail:
			out, wrapErr := speccli.WrapSingleResult(v, spec.ResultPath)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		case *openlaneclient.GetSubscribers:
			out, wrapErr := speccli.WrapListResult(v, spec.ListRoot)
			if wrapErr != nil {
				return true, speccli.OperationOutput{}, wrapErr
			}
			return true, out, nil
		default:
			return true, speccli.OperationOutput{}, fmt.Errorf("unexpected subscriber get result type %T", result)
		}
	}
}

func subscriberRecordsFromResult(result any) ([]map[string]any, error) {
	if result == nil {
		return []map[string]any{}, nil
	}

	payload, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var list []map[string]any
	if err := json.Unmarshal(payload, &list); err != nil {
		return nil, err
	}

	return list, nil
}
