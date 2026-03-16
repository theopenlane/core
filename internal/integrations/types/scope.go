package types

import (
	"encoding/json"

	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// ScopeVariablePayload identifies the payload variable in scope expressions
	ScopeVariablePayload = "payload"
	// ScopeVariableResource identifies the resource variable in scope expressions
	ScopeVariableResource = "resource"
	// ScopeVariableDefinition identifies the definition slug variable in scope expressions (named "provider" for CEL expression compatibility)
	ScopeVariableDefinition = "provider"
	// ScopeVariableOperation identifies the operation name variable in scope expressions
	ScopeVariableOperation = "operation"
	// ScopeVariableConfig identifies operation config values in scope expressions
	ScopeVariableConfig = "config"
	// ScopeVariableInstallationConfig identifies installation-level config values in scope expressions (named "integration_config" for CEL expression compatibility)
	ScopeVariableInstallationConfig = "integration_config"
	// ScopeVariableProviderState identifies persisted provider state values in scope expressions
	ScopeVariableProviderState = "provider_state"
	// ScopeVariableOrgID identifies the installation owner id in scope expressions
	ScopeVariableOrgID = "org_id"
	// ScopeVariableInstallationID identifies the installation id in scope expressions (named "integration_id" for CEL expression compatibility)
	ScopeVariableInstallationID = "integration_id"
)

// ScopeVars contains standard variables available to integration scope CEL expressions
type ScopeVars struct {
	// Payload contains payload data for filtering
	Payload json.RawMessage
	// Resource contains resource identity values
	Resource string
	// DefinitionID identifies the definition by canonical ID (exposed as "provider" in CEL for compatibility)
	DefinitionID string
	// Operation contains operation name values
	Operation string
	// Config contains operation config values
	Config json.RawMessage
	// InstallationConfig contains installation-level config values (exposed as "integration_config" in CEL for compatibility)
	InstallationConfig json.RawMessage
	// ProviderState contains installation provider state values
	ProviderState json.RawMessage
	// OrgID contains installation owner id values
	OrgID string
	// InstallationID contains installed integration id values (exposed as "integration_id" in CEL for compatibility)
	InstallationID string
}

// CELVars converts scope vars into CEL variable bindings
func (v ScopeVars) CELVars() map[string]any {
	return map[string]any{
		ScopeVariablePayload:            jsonx.DecodeAnyOrNil(v.Payload),
		ScopeVariableResource:           v.Resource,
		ScopeVariableDefinition:         v.DefinitionID,
		ScopeVariableOperation:          v.Operation,
		ScopeVariableConfig:             jsonx.DecodeAnyOrNil(v.Config),
		ScopeVariableInstallationConfig: jsonx.DecodeAnyOrNil(v.InstallationConfig),
		ScopeVariableProviderState:      jsonx.DecodeAnyOrNil(v.ProviderState),
		ScopeVariableOrgID:              v.OrgID,
		ScopeVariableInstallationID:     v.InstallationID,
	}
}
