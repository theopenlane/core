package webhooks

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/pkg/gala"
)

// Envelope is the durable payload emitted for one inbound integration webhook event
type Envelope struct {
	IntegrationID string `json:"integrationId"`
	DefinitionID  string `json:"definitionId"`
	Webhook       string `json:"webhook"`
	Event         string `json:"event"`
	DeliveryID    string `json:"deliveryId,omitempty"`
	Payload       json.RawMessage `json:"payload"`
	Headers       map[string]string `json:"headers,omitempty"`
}

// RegisterListeners attaches one Gala listener per registered webhook event topic
func RegisterListeners(runtime *gala.Gala, reg *registry.Registry, handle func(context.Context, Envelope) error) error {
	for _, event := range reg.WebhookListeners() {
		if _, err := gala.RegisterListeners(runtime.Registry(), gala.Definition[Envelope]{
			Topic: gala.Topic[Envelope]{
				Name: event.Topic,
			},
			Name: event.Name,
			Handle: func(ctx gala.HandlerContext, envelope Envelope) error {
				return handle(ctx.Context, envelope)
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
