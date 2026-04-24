package handlers

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/operations"
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

	_, err = h.IntegrationsRuntime.Dispatch(ctx, operations.DispatchRequest{
		DefinitionID: email.DefinitionID.ID(),
		Operation:    operationName,
		Config:       config,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
	})

	return err
}
