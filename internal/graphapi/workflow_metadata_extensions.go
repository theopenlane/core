package graphapi

import (
	"sort"

	"github.com/theopenlane/core/common/enums"
	integrationscope "github.com/theopenlane/core/internal/integrations/scope"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

// integrationScopeVariableNames lists CEL variables exposed to integration scope expressions
var integrationScopeVariableNames = []string{
	integrationscope.VariablePayload,
	integrationscope.VariableResource,
	integrationscope.VariableProvider,
	integrationscope.VariableOperation,
	integrationscope.VariableConfig,
	integrationscope.VariableIntegrationConfig,
	integrationscope.VariableProviderState,
	integrationscope.VariableOrgID,
	integrationscope.VariableIntegrationID,
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
	// Provider is the provider identifier
	Provider string `json:"provider"`
	// DisplayName is the provider display name
	DisplayName string `json:"display_name"`
	// Category is the provider category
	Category string `json:"category"`
	// AuthKind is the provider auth kind
	AuthKind string `json:"auth_kind"`
	// CredentialsSchema is the provider credentials schema
	CredentialsSchema map[string]any `json:"credentials_schema,omitempty"`
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
	ConfigSchema map[string]any `json:"config_schema,omitempty"`
	// OutputSchema is the operation output schema
	OutputSchema map[string]any `json:"output_schema,omitempty"`
}

// workflowMetadataExtensions builds extensible workflow metadata payloads for non-object schema surfaces
func workflowMetadataExtensions(source integrationMetadataSource) map[string]any {
	doc := workflowMetadataExtensionsDocument{
		Integrations: integrationWorkflowExtensionsDocument{
			ActionContract: integrationActionContractDocument{
				TargetSelector:    []string{"integration_id", "provider"},
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

	catalog := source.ProviderMetadataCatalog()
	providers := make([]integrationtypes.ProviderType, 0, len(catalog))
	for provider := range catalog {
		providers = append(providers, provider)
	}
	sort.Slice(providers, func(i int, j int) bool {
		return providers[i] < providers[j]
	})

	providerEntries := make([]integrationProviderExtensionsDocument, 0, len(providers))
	for _, provider := range providers {
		meta := catalog[provider]
		entry := integrationProviderExtensionsDocument{
			Provider:    string(provider),
			DisplayName: meta.DisplayName,
			Category:    meta.Category,
			AuthKind:    string(meta.Auth),
		}
		if len(meta.Schema) > 0 {
			entry.CredentialsSchema = mapx.DeepCloneMapAny(meta.Schema)
		}

		descriptors := source.OperationDescriptors(provider)
		sort.Slice(descriptors, func(i int, j int) bool {
			return descriptors[i].Name < descriptors[j].Name
		})

		operations := make([]integrationOperationExtensionsDocument, 0, len(descriptors))
		for _, descriptor := range descriptors {
			op := integrationOperationExtensionsDocument{
				Name:        string(descriptor.Name),
				Kind:        string(descriptor.Kind),
				Description: descriptor.Description,
				Client:      string(descriptor.Client),
			}
			if len(descriptor.ConfigSchema) > 0 {
				op.ConfigSchema = mapx.DeepCloneMapAny(descriptor.ConfigSchema)
			}
			if len(descriptor.OutputSchema) > 0 {
				op.OutputSchema = mapx.DeepCloneMapAny(descriptor.OutputSchema)
			}

			operations = append(operations, op)
		}
		entry.Operations = operations

		providerEntries = append(providerEntries, entry)
	}

	return providerEntries
}
