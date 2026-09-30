package registry

import (
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"slices"

	"github.com/invopop/jsonschema"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// finalizeDefinition derives the credential form schemas and the operation config sections, switches, and resolvers the builder leaves to the registry
func finalizeDefinition(def types.Definition) (types.Definition, error) {
	def.CredentialRegistrations = lo.Map(def.CredentialRegistrations, func(registration types.CredentialRegistration, _ int) types.CredentialRegistration {
		return finalizeCredential(def.Connections, registration)
	})

	operations, err := finalizeOperations(def)
	if err != nil {
		return types.Definition{}, err
	}

	def.Operations = operations

	return def, nil
}

// finalizeCredential fills the stored schema from a hand-built form schema and the form schema from the stored schema when no auth flow manages the slot
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

// finalizeOperations locates each operation's config section in the user input and derives its resolver and switch, enforcing one operation per section
func finalizeOperations(def types.Definition) ([]types.OperationRegistration, error) {
	inputRoot, inputDefs, err := jsonx.SchemaRoot(lo.FromPtr(def.UserInput).Schema)
	if err != nil {
		return nil, fmt.Errorf("definition %s user input: %w", def.ID, err)
	}

	operations := slices.Clone(def.Operations)
	claimed := map[string]string{}

	for i := range operations {
		operation := &operations[i]

		if len(operation.ConfigSchema) == 0 {
			continue
		}

		root, configDefs, err := jsonx.SchemaRoot(operation.ConfigSchema)
		if err != nil {
			return nil, fmt.Errorf("definition %s operation %s config: %w", def.ID, operation.Name, err)
		}

		key, err := matchSection(inputRoot, inputDefs, jsonx.SchemaID(operation.ConfigSchema), configDefs)
		if err != nil {
			return nil, fmt.Errorf("%w: definition %s operation %s", err, def.ID, operation.Name)
		}

		if key != "" {
			if holder, taken := claimed[key]; taken {
				return nil, fmt.Errorf("%w: definition %s operations %s and %s share section %s", ErrConfigSectionAmbiguous, def.ID, holder, operation.Name, key)
			}

			claimed[key] = operation.Name

			bindSection(operation, key)
		}

		if operation.Policy.Reconcile && operation.ConfigResolver == nil && root.Properties != nil && root.Properties.Len() > 0 {
			return nil, fmt.Errorf("%w: definition %s operation %s", ErrConfigSectionRequired, def.ID, operation.Name)
		}
	}

	return operations, nil
}

// bindSection fills the operation's resolver with the section lookup and its switch with the section's disable toggle, leaving authored values in place
func bindSection(operation *types.OperationRegistration, key string) {
	if operation.ConfigResolver == nil {
		operation.ConfigResolver = func(userInput json.RawMessage) json.RawMessage {
			raw, _ := jsonx.DecodeObjectKey[json.RawMessage](userInput, key)

			return raw
		}
	}

	if operation.Disabled == nil {
		resolver := operation.ConfigResolver

		operation.Disabled = func(userInput json.RawMessage) bool {
			toggle, err := jsonx.Decode[types.Switch](resolver(userInput))

			return err == nil && toggle.Disable
		}
	}
}

// matchSection returns the single user input property key whose referenced type is name and whose definition equals the operation config's, empty when no property references it
func matchSection(inputRoot *jsonschema.Schema, inputDefs jsonschema.Definitions, name string, configDefs jsonschema.Definitions) (string, error) {
	var candidates []string

	for pair := inputRoot.Properties.Oldest(); pair != nil; pair = pair.Next() {
		if pair.Value.Ref != "" && path.Base(pair.Value.Ref) == name {
			candidates = append(candidates, pair.Key)
		}
	}

	switch {
	case len(candidates) == 0:
		return "", nil
	case len(candidates) > 1:
		return "", fmt.Errorf("%w: properties %v reference %s", ErrConfigSectionAmbiguous, candidates, name)
	case !sameSchema(inputDefs[name], configDefs[name]):
		return "", fmt.Errorf("%w: property %s type %s", ErrConfigSectionMismatch, candidates[0], name)
	default:
		return candidates[0], nil
	}
}

// sameSchema reports whether two schema nodes encode to the same document, false when either is absent
func sameSchema(a, b *jsonschema.Schema) bool {
	if a == nil || b == nil {
		return false
	}

	left, err := json.Marshal(a)
	if err != nil {
		return false
	}

	right, err := json.Marshal(b)
	if err != nil {
		return false
	}

	return sameJSON(left, right)
}

// sameJSON reports whether two raw documents decode to the same value regardless of encoding
func sameJSON(a, b json.RawMessage) bool {
	var left, right any

	if err := json.Unmarshal(a, &left); err != nil {
		return false
	}

	if err := json.Unmarshal(b, &right); err != nil {
		return false
	}

	return reflect.DeepEqual(left, right)
}
