//go:build test

package integrations

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// healthHandler validates the bound credential, failing on the marker values
func healthHandler(_ context.Context, req types.OperationRequest) (json.RawMessage, error) {
	if sa, ok, err := ServiceAccountCredential.Resolve(req.Credentials); err == nil && ok && sa.ProjectID == FailProjectID {
		return nil, ErrHealthFailed
	}

	if tok, ok, err := TokenCredential.Resolve(req.Credentials); err == nil && ok && tok.Token == FailToken {
		return nil, ErrHealthFailed
	}

	return json.RawMessage(`{"ok":true}`), nil
}

// repoSyncHandler runs with the built client
func repoSyncHandler(context.Context, types.OperationRequest, *Client, repoSync) (json.RawMessage, error) {
	return json.RawMessage(`{"synced":true}`), nil
}

// validatedHandler backs the inline operation with a required config field
func validatedHandler(context.Context, types.OperationRequest, validatedRun) (json.RawMessage, error) {
	return json.RawMessage(`{"validated":true}`), nil
}

// idleCycle reports zero drift
func idleCycle[Config any](context.Context, types.OperationRequest, Config) (json.RawMessage, error) {
	return nil, nil
}

// idleClientCycle reports zero drift once the test client resolves
func idleClientCycle(context.Context, types.OperationRequest, *Client, unresolvableCycle) (json.RawMessage, error) {
	return nil, nil
}

// failingCycle always fails
func failingCycle(context.Context, types.OperationRequest, exhaustingCycle) (json.RawMessage, error) {
	return nil, ErrCycleFailed
}

// webhookInboundEvent decodes the inbound webhook payload
func webhookInboundEvent(req types.WebhookInboundRequest) (types.WebhookReceivedEvent, error) {
	var envelope struct {
		Event      string `json:"event"`
		DeliveryID string `json:"delivery_id"`
	}
	if err := json.Unmarshal(req.Payload, &envelope); err != nil {
		return types.WebhookReceivedEvent{}, err
	}

	if envelope.Event == "" {
		return types.WebhookReceivedEvent{}, nil
	}

	return types.WebhookReceivedEvent{
		Name:       envelope.Event,
		DeliveryID: envelope.DeliveryID,
		Payload:    req.Payload,
	}, nil
}
