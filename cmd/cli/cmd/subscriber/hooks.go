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
		result, err := createSubscribers(ctx, client)
		if err != nil {
			return true, speccli.OperationOutput{}, err
		}

		records, recordErr := buildSubscriberRecords(result)
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
		result, err := getSubscribers(ctx, client)
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

func buildSubscriberRecords(result *openlaneclient.CreateBulkSubscriber) ([]map[string]any, error) {
	if result == nil {
		return []map[string]any{}, nil
	}

	payload, err := json.Marshal(result.CreateBulkSubscriber.Subscribers)
	if err != nil {
		return nil, err
	}

	var subscribers []openlaneclient.Subscriber
	if err := json.Unmarshal(payload, &subscribers); err != nil {
		return nil, err
	}

	records := make([]map[string]any, len(subscribers))
	for i, sub := range subscribers {
		records[i] = map[string]any{
			"email":        sub.Email,
			"active":       sub.Active,
			"unsubscribed": sub.Unsubscribed,
			"tags":         sub.Tags,
		}
	}

	return records, nil
}
