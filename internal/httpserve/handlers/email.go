package handlers

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/definitions/email"
)

// sendEmail dispatches a system email operation via the integrations runtime,
// following the same pattern as hooks.sendSystemEmail
func (h *Handler) sendEmail(ctx context.Context, operationName string, input any) error {
	if h.IntegrationsRuntime == nil {
		return nil
	}

	config, err := json.Marshal(input)
	if err != nil {
		return err
	}

	_, err = h.IntegrationsRuntime.ExecuteRuntimeOperation(ctx, email.DefinitionID.ID(), operationName, config)

	return err
}
