package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// finalizeDefinition derives credential form schemas, defaults credential slots, and composes operation input schemas
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

	operations, err := finalizeOperations(def)
	if err != nil {
		return types.Definition{}, err
	}

	def.Operations = operations

	return def, nil
}

// finalizeCredential fills stored/form schema from each other when unset
func finalizeCredential(connections []types.ConnectionRegistration, registration types.CredentialRegistration) types.CredentialRegistration {
	if len(registration.StoredSchema) == 0 {
		registration.StoredSchema = registration.Schema
	}

	authManaged := lo.ContainsBy(connections, func(connection types.ConnectionRegistration) bool {
		return connection.Auth != nil && connection.Auth.CredentialRef == registration.Ref
	})

	if !authManaged && len(registration.Schema) == 0 {
		registration.Schema = registration.StoredSchema
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

// finalizeOperations composes each operation's stored input schema from the uniform settings and its config schema
func finalizeOperations(def types.Definition) ([]types.OperationRegistration, error) {
	operations := slices.Clone(def.Operations)

	for i := range operations {
		operation := &operations[i]

		if operation.Input == nil {
			continue
		}

		schema, err := jsonx.MergeSchemas(types.OperationSettingsSchema(), operation.ConfigSchema)

		switch {
		case errors.Is(err, jsonx.ErrSchemaPropertyConflict):
			return nil, fmt.Errorf("%w: definition %s operation %s: %w", ErrOperationConfigReservedKey, def.ID, operation.Name, err)
		case err != nil:
			return nil, fmt.Errorf("definition %s operation %s input: %w", def.ID, operation.Name, err)
		}

		input := *operation.Input
		input.Schema = schema
		input.Validate = validateOperationInput(input.Validate)
		operation.Input = &input
	}

	return operations, nil
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
