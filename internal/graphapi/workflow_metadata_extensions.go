package graphapi

import (
	"encoding/json"
	"slices"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// integrationScopeVariableNames lists CEL variables exposed to integration scope expressions
var integrationScopeVariableNames = []string{
	"payload",
	"resource",
	"provider",
	"operation",
	"config",
	"integration_config",
	"provider_state",
	"org_id",
	"integration_id",
}

// workflowMetadataExtensionsDocument is the serialized shape for workflow metadata extensions
type workflowMetadataExtensionsDocument struct {
	// Integrations stores integration action configuration metadata
	Integrations integrationWorkflowExtensionsDocument `json:"integrations"`
}

// integrationWorkflowExtensionsDocument describes integration action metadata for workflow configuration surfaces
type integrationWorkflowExtensionsDocument struct {
	// ActionContract describes selector and scope metadata for integration actions
	ActionContract integrationActionContractDocument `json:"action_contract"`
	// Providers describes provider and operation metadata published by the registry
	Providers []integrationProviderExtensionsDocument `json:"providers"`
}

// integrationActionContractDocument defines required selectors and scope metadata for integration actions
type integrationActionContractDocument struct {
	// TargetSelector lists integration target selector fields
	TargetSelector []string `json:"target_selector"`
	// OperationSelector lists integration operation selector fields
	OperationSelector []string `json:"operation_selector"`
	// ScopeFields lists integration scope fields
	ScopeFields []string `json:"scope_fields"`
	// ScopeVariables lists CEL variables exposed to scope expressions
	ScopeVariables []string `json:"scope_variables"`
	// RunTypes lists integration run type values
	RunTypes []string `json:"run_types"`
}

// integrationProviderExtensionsDocument captures workflow metadata for one provider
type integrationProviderExtensionsDocument struct {
	// Provider is the canonical definition identifier
	Provider string `json:"provider"`
	// DisplayName is the provider display name
	DisplayName string `json:"display_name"`
	// Category is the provider category
	Category string `json:"category"`
	// Auth describes the install or auth flow exposed by the definition
	Auth *types.AuthRegistration `json:"auth,omitempty"`
	// CredentialSchemas lists the provider credential slots
	CredentialSchemas []integrationCredentialExtensionsDocument `json:"credential_schemas,omitempty"`
	// UserInputSchema is the installation-scoped provider schema
	UserInputSchema json.RawMessage `json:"user_input_schema,omitempty"`
	// Operations lists provider operation descriptors
	Operations []integrationOperationExtensionsDocument `json:"operations"`
}

// integrationCredentialExtensionsDocument captures workflow metadata for one provider credential slot
type integrationCredentialExtensionsDocument struct {
	// Ref is the durable credential slot identifier
	Ref types.CredentialRef `json:"ref"`
	// Name is the user-facing credential slot name
	Name string `json:"name,omitempty"`
	// Description describes the credential slot
	Description string `json:"description,omitempty"`
	// Schema is the credential collection schema
	Schema json.RawMessage `json:"schema,omitempty"`
}

// integrationOperationExtensionsDocument captures workflow metadata for one operation
type integrationOperationExtensionsDocument struct {
	// Name is the operation name
	Name string `json:"name"`
	// Description is the operation description
	Description string `json:"description,omitempty"`
	// ConfigSchema is the operation config schema
	ConfigSchema json.RawMessage `json:"config_schema,omitempty"`
}

// workflowMetadataExtensions builds extensible workflow metadata payloads for non-object schema surfaces
func workflowMetadataExtensions(rt *integrationsruntime.Runtime) map[string]any {
	doc := workflowMetadataExtensionsDocument{
		Integrations: integrationWorkflowExtensionsDocument{
			ActionContract: integrationActionContractDocument{
				TargetSelector:    []string{"installation_id", "definition_id"},
				OperationSelector: []string{"operation_name"},
				ScopeFields:       []string{"scope_expression", "scope_payload", "scope_resource"},
				ScopeVariables:    append([]string(nil), integrationScopeVariableNames...),
				RunTypes:          append([]string(nil), enums.IntegrationRunTypes...),
			},
			Providers: integrationWorkflowProviders(rt),
		},
	}

	extensions, err := jsonx.ToMap(doc)
	if err != nil || extensions == nil {
		return map[string]any{}
	}

	return extensions
}

// integrationWorkflowProviders builds provider metadata for integration workflow extensions
func integrationWorkflowProviders(rt *integrationsruntime.Runtime) []integrationProviderExtensionsDocument {
	if rt == nil {
		return []integrationProviderExtensionsDocument{}
	}

	specs := rt.Catalog()
	entries := make([]integrationProviderExtensionsDocument, 0, len(specs))

	for _, spec := range specs {
		def, ok := rt.Definition(spec.ID)
		if !ok {
			continue
		}

		entry := integrationProviderExtensionsDocument{
			Provider:    spec.ID,
			DisplayName: spec.DisplayName,
			Category:    spec.Category,
		}

		if def.Auth != nil && (def.Auth.StartPath != "" || def.Auth.CallbackPath != "" || def.Auth.OAuth != nil) {
			entry.Auth = def.Auth
		}

		if len(def.CredentialRegistrations) > 0 {
			entry.CredentialSchemas = lo.Map(def.CredentialRegistrations, func(credential types.CredentialRegistration, _ int) integrationCredentialExtensionsDocument {
				return integrationCredentialExtensionsDocument{
					Ref:         credential.Ref,
					Name:        credential.Name,
					Description: credential.Description,
					Schema:      jsonx.CloneRawMessage(credential.Schema),
				}
			})
		}

		if def.UserInput != nil {
			entry.UserInputSchema = jsonx.CloneRawMessage(def.UserInput.Schema)
		}

		operations := buildOperationEntries(def.Operations)
		slices.SortFunc(operations, func(a, b integrationOperationExtensionsDocument) int {
			if a.Name < b.Name {
				return -1
			}
			if a.Name > b.Name {
				return 1
			}
			return 0
		})
		entry.Operations = operations
		entries = append(entries, entry)
	}

	return entries
}

// buildOperationEntries converts operation registrations to extension documents
func buildOperationEntries(ops []types.OperationRegistration) []integrationOperationExtensionsDocument {
	return lo.Map(ops, func(op types.OperationRegistration, _ int) integrationOperationExtensionsDocument {
		return integrationOperationExtensionsDocument{
			Name:         op.Name,
			Description:  op.Description,
			ConfigSchema: jsonx.CloneRawMessage(op.ConfigSchema),
		}
	})
}
