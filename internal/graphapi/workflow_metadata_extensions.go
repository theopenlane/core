package graphapi

import (
	"encoding/json"
	"slices"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
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
	// CredentialsSchema is the provider credentials schema
	CredentialsSchema json.RawMessage `json:"credentials_schema,omitempty"`
	// UserInputSchema is the installation-scoped provider schema
	UserInputSchema json.RawMessage `json:"user_input_schema,omitempty"`
	// Operations lists provider operation descriptors
	Operations []integrationOperationExtensionsDocument `json:"operations"`
}

// integrationOperationExtensionsDocument captures workflow metadata for one operation
type integrationOperationExtensionsDocument struct {
	// Name is the operation name
	Name string `json:"name"`
	// Kind is the operation kind
	Kind string `json:"kind"`
	// Description is the operation description
	Description string `json:"description,omitempty"`
	// Client is the operation client identifier
	Client string `json:"client,omitempty"`
	// ConfigSchema is the operation config schema
	ConfigSchema json.RawMessage `json:"config_schema,omitempty"`
}

// workflowMetadataExtensions builds extensible workflow metadata payloads for non-object schema surfaces
func workflowMetadataExtensions(source integrationMetadataSource) map[string]any {
	doc := workflowMetadataExtensionsDocument{
		Integrations: integrationWorkflowExtensionsDocument{
			ActionContract: integrationActionContractDocument{
				TargetSelector:    []string{"installation_id", "definition_id"},
				OperationSelector: []string{"operation_name", "operation_kind"},
				ScopeFields:       []string{"scope_expression", "scope_payload", "scope_resource"},
				ScopeVariables:    append([]string(nil), integrationScopeVariableNames...),
				RunTypes:          append([]string(nil), enums.IntegrationRunTypes...),
			},
			Providers: integrationWorkflowProviders(source),
		},
	}

	extensions, err := jsonx.ToMap(doc)
	if err != nil || extensions == nil {
		return map[string]any{}
	}

	return extensions
}

// integrationWorkflowProviders builds provider metadata for integration workflow extensions
func integrationWorkflowProviders(source integrationMetadataSource) []integrationProviderExtensionsDocument {
	if source == nil {
		return []integrationProviderExtensionsDocument{}
	}

	specs := source.Catalog()
	entries := make([]integrationProviderExtensionsDocument, 0, len(specs))

	for _, spec := range specs {
		def, ok := source.Definition(spec.ID)
		if !ok {
			continue
		}

		entry := integrationProviderExtensionsDocument{
			Provider:    string(spec.ID),
			DisplayName: spec.DisplayName,
			Category:    spec.Category,
		}

		if def.Credentials != nil {
			entry.CredentialsSchema = jsonx.CloneRawMessage(def.Credentials.Schema)
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
			Name:         string(op.Name),
			Kind:         string(op.Kind),
			Description:  op.Description,
			Client:       string(op.Client),
			ConfigSchema: jsonx.CloneRawMessage(op.ConfigSchema),
		}
	})
}
