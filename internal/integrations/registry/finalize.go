package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// finalizeDefinition derives credential form schemas, defaults credential slots, and wraps stored-input validation
func finalizeDefinition(def types.Definition) (types.Definition, error) {
	declared := lo.Map(def.CredentialRegistrations, func(registration types.CredentialRegistration, _ int) types.CredentialSlotID {
		return registration.Ref
	})

	def.CredentialRegistrations = lo.Map(def.CredentialRegistrations, func(registration types.CredentialRegistration, _ int) types.CredentialRegistration {
		return finalizeCredential(def.Connections, registration)
	})

	def.Clients = lo.Map(def.Clients, func(client types.ClientRegistration, _ int) types.ClientRegistration {
		return finalizeClient(client, declared)
	})

	def.Connections = lo.Map(def.Connections, func(connection types.ConnectionRegistration, _ int) types.ConnectionRegistration {
		return finalizeConnection(connection)
	})

	def.Operations = finalizeOperations(def)

	return def, nil
}

// finalizeCredential keys the stored layout by the slot name and fills stored/form schema from each other when unset
func finalizeCredential(connections []types.ConnectionRegistration, registration types.CredentialRegistration) types.CredentialRegistration {
	registration.Stored.Name = registration.Ref.String()

	if len(registration.Stored.Schema) == 0 {
		registration.Stored.Schema = registration.Schema
	}

	authManaged := lo.ContainsBy(connections, func(connection types.ConnectionRegistration) bool {
		return connection.Auth != nil && connection.Auth.CredentialRef == registration.Ref
	})

	if !authManaged && len(registration.Schema) == 0 {
		registration.Schema = registration.Stored.Schema
	}

	return registration
}

// finalizeClient defaults a client declaring no credential slots to every declared slot
func finalizeClient(client types.ClientRegistration, declared []types.CredentialSlotID) types.ClientRegistration {
	if len(client.CredentialRefs) == 0 {
		client.CredentialRefs = slices.Clone(declared)
	}

	return client
}

// finalizeConnection defaults a connection declaring no credential slots to its selecting slot
func finalizeConnection(connection types.ConnectionRegistration) types.ConnectionRegistration {
	if len(connection.CredentialRefs) == 0 {
		connection.CredentialRefs = []types.CredentialSlotID{connection.CredentialRef}
	}

	return connection
}

// finalizeOperations wraps each stored-input operation's validation with the uniform filter expression check
func finalizeOperations(def types.Definition) []types.OperationRegistration {
	operations := slices.Clone(def.Operations)

	for i := range operations {
		operation := &operations[i]

		if !operation.Stored {
			continue
		}

		operation.Input.Validate = validateOperationInput(operation.Input.Validate)
	}

	return operations
}

// validateOperationInput compiles the stored filter expression before running the definition's own validation
func validateOperationInput(next types.ValidateFunc) types.ValidateFunc {
	return func(ctx context.Context, req types.InstallationRequest, payload json.RawMessage) error {
		settings, err := types.OperationSettingsFrom(payload)
		if err != nil {
			return err
		}

		if settings.FilterExpr != "" {
			if err := providerkit.ValidateExpr(settings.FilterExpr); err != nil {
				return fmt.Errorf("%w: %w", ErrOperationFilterExprInvalid, err)
			}
		}

		if next == nil {
			return nil
		}

		return next(ctx, req, payload)
	}
}
