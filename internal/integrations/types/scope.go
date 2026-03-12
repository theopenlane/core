package types

import (
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// ScopeVariablePayload identifies the provider payload variable in scope expressions
	ScopeVariablePayload = "payload"
	// ScopeVariableResource identifies the provider resource variable in scope expressions
	ScopeVariableResource = "resource"
	// ScopeVariableProvider identifies the provider kind variable in scope expressions
	ScopeVariableProvider = "provider"
	// ScopeVariableOperation identifies the operation name variable in scope expressions
	ScopeVariableOperation = "operation"
	// ScopeVariableConfig identifies operation config values in scope expressions
	ScopeVariableConfig = "config"
	// ScopeVariableIntegrationConfig identifies integration-level config values in scope expressions
	ScopeVariableIntegrationConfig = "integration_config"
	// ScopeVariableProviderState identifies persisted provider state values in scope expressions
	ScopeVariableProviderState = "provider_state"
	// ScopeVariableOrgID identifies the integration owner id in scope expressions
	ScopeVariableOrgID = "org_id"
	// ScopeVariableIntegrationID identifies the installed integration id in scope expressions
	ScopeVariableIntegrationID = "integration_id"
)

// ScopeVars contains standard variables available to integration scope CEL expressions
type ScopeVars struct {
	// Payload contains provider payload data for filtering
	Payload json.RawMessage
	// Resource contains provider resource identity values
	Resource string
	// Provider contains provider kind values
	Provider ProviderType
	// Operation contains operation name values
	Operation OperationName
	// Config contains operation config values
	Config json.RawMessage
	// IntegrationConfig contains integration-level config values
	IntegrationConfig json.RawMessage
	// ProviderState contains integration provider state values
	ProviderState json.RawMessage
	// OrgID contains integration owner id values
	OrgID string
	// IntegrationID contains installed integration id values
	IntegrationID string
}

// CELVars converts scope vars into CEL variable bindings
func (v ScopeVars) CELVars() map[string]any {
	return map[string]any{
		ScopeVariablePayload:           jsonx.DecodeAnyOrNil(v.Payload),
		ScopeVariableResource:          v.Resource,
		ScopeVariableProvider:          string(v.Provider),
		ScopeVariableOperation:         string(v.Operation),
		ScopeVariableConfig:            jsonx.DecodeAnyOrNil(v.Config),
		ScopeVariableIntegrationConfig: jsonx.DecodeAnyOrNil(v.IntegrationConfig),
		ScopeVariableProviderState:     jsonx.DecodeAnyOrNil(v.ProviderState),
		ScopeVariableOrgID:             v.OrgID,
		ScopeVariableIntegrationID:     v.IntegrationID,
	}
}
