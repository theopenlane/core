package handlers

import (
	"context"
	"errors"
	"strings"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/types"
)

var errIntegrationDefinitionMismatch = errors.New("integration installation does not match requested provider")

func (h *Handler) resolveIntegrationDefinition(reference string) (types.Definition, bool) {
	trimmed := strings.TrimSpace(reference)
	if trimmed == "" {
		return types.Definition{}, false
	}

	return h.IntegrationsRuntime.Registry().Definition(types.DefinitionID(trimmed))
}

func (h *Handler) loadOwnedIntegration(ctx context.Context, ownerID, installationID string) (*ent.Integration, error) {
	return h.DBClient.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.IDEQ(installationID),
		).
		Only(ctx)
}

func (h *Handler) loadOwnedIntegrationByDefinition(ctx context.Context, ownerID string, definitionID types.DefinitionID) (*ent.Integration, error) {
	return h.DBClient.Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.DefinitionIDEQ(string(definitionID)),
		).
		Only(ctx)
}

func validateInstallationDefinition(record *ent.Integration, definition types.Definition) error {
	if record == nil {
		return ErrIntegrationNotFound
	}

	if record.DefinitionID != string(definition.Spec.ID) {
		return errIntegrationDefinitionMismatch
	}

	return nil
}
