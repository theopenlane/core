package types //nolint:revive

import (
	"encoding/json"

	"github.com/theopenlane/core/v2/pkg/jsonx"
)

const (
	// ScopeVariablePayload identifies the payload variable in scope expressions
	ScopeVariablePayload = "payload"
	// ScopeVariableResource identifies the resource variable in scope expressions
	ScopeVariableResource = "resource"
	// ScopeVariableDefinition identifies the definition id variable (CEL name: provider)
	ScopeVariableDefinition = "provider"
	// ScopeVariableOperation identifies the operation name variable in scope expressions
	ScopeVariableOperation = "operation"
	// ScopeVariableConfig identifies operation config values in scope expressions
	ScopeVariableConfig = "config"
	// ScopeVariableInstallationConfig identifies installation config (CEL name: integration_config)
	ScopeVariableInstallationConfig = "integration_config"
	// ScopeVariableOrgID identifies the installation owner id in scope expressions
	ScopeVariableOrgID = "org_id"
	// ScopeVariableInstallationID identifies the installation id (CEL name: integration_id)
	ScopeVariableInstallationID = "integration_id"
)

// ScopeVars contains standard variables available to integration scope CEL expressions
type ScopeVars struct {
	// Payload contains payload data for filtering
	Payload json.RawMessage
	// Resource contains resource identity values
	Resource string
	// Definition identifies the definition by canonical id (CEL: provider)
	Definition string
	// Operation contains operation name values
	Operation string
	// Config contains operation config values
	Config json.RawMessage
	// InstallationConfig contains installation-level config values (CEL: integration_config)
	InstallationConfig json.RawMessage
	// OrgID contains installation owner id values
	OrgID string
	// InstallationID contains installed integration id values (CEL: integration_id)
	InstallationID string
}

// CELVars converts scope vars into CEL variable bindings
func (v ScopeVars) CELVars() map[string]any {
	return map[string]any{
		ScopeVariablePayload:            jsonx.DecodeAnyOrNil(v.Payload),
		ScopeVariableResource:           v.Resource,
		ScopeVariableDefinition:         v.Definition,
		ScopeVariableOperation:          v.Operation,
		ScopeVariableConfig:             jsonx.DecodeAnyOrNil(v.Config),
		ScopeVariableInstallationConfig: jsonx.DecodeAnyOrNil(v.InstallationConfig),
		ScopeVariableOrgID:              v.OrgID,
		ScopeVariableInstallationID:     v.InstallationID,
	}
}
