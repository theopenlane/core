package registry

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/invopop/jsonschema"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// userInputSections is the user input schema's top-level properties that reference a named type, with the schema definitions they resolve against
type userInputSections struct {
	// refs maps each top-level property key to the basename of the type it references, empty for properties without a reference
	refs map[string]string
	// defs holds the user input schema definitions
	defs jsonschema.Definitions
}

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
	sections, err := resolveUserInputSections(def.UserInput)
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

		key, err := sections.match(jsonx.SchemaID(operation.ConfigSchema), configDefs)
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

// bindSection fills the operation's resolver with the section lookup and its switch with the config switch applied to the resolved section, leaving authored values in place
func bindSection(operation *types.OperationRegistration, key string) {
	if operation.ConfigResolver == nil {
		operation.ConfigResolver = func(userInput json.RawMessage) json.RawMessage {
			raw, _ := jsonx.DecodeObjectKey[json.RawMessage](userInput, key)

			return raw
		}
	}

	if operation.Disabled == nil && operation.ConfigDisabled != nil {
		resolve, disabled := operation.ConfigResolver, operation.ConfigDisabled

		operation.Disabled = func(userInput json.RawMessage) bool {
			return disabled(resolve(userInput))
		}
	}
}

// resolveUserInputSections indexes the user input schema's top-level properties that reference a named type, empty when the definition has no user input
func resolveUserInputSections(userInput *types.UserInputRegistration) (userInputSections, error) {
	if userInput == nil {
		return userInputSections{}, nil
	}

	refs, err := jsonx.PropertyRefs(userInput.Schema)
	if err != nil {
		return userInputSections{}, err
	}

	_, defs, err := jsonx.SchemaRoot(userInput.Schema)
	if err != nil {
		return userInputSections{}, err
	}

	return userInputSections{refs: refs, defs: defs}, nil
}

// match returns the single property key whose referenced type is name and whose definition equals the operation config's, empty when no property references it
func (s userInputSections) match(name string, configDefs jsonschema.Definitions) (string, error) {
	candidates := lo.Keys(lo.PickBy(s.refs, func(_ string, ref string) bool { return ref != "" && ref == name }))
	slices.Sort(candidates)

	switch {
	case len(candidates) == 0:
		return "", nil
	case len(candidates) > 1:
		return "", fmt.Errorf("%w: properties %v reference %s", ErrConfigSectionAmbiguous, candidates, name)
	case !sameSchema(s.defs[name], configDefs[name]):
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
