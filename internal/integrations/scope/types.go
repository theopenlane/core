package scope

import (
	"time"

	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

const (
	// VariablePayload identifies the provider payload variable
	VariablePayload = "payload"
	// VariableResource identifies the provider resource variable
	VariableResource = "resource"
	// VariableProvider identifies the provider kind variable
	VariableProvider = "provider"
	// VariableOperation identifies the operation name variable
	VariableOperation = "operation"
	// VariableConfig identifies operation config values
	VariableConfig = "config"
	// VariableIntegrationConfig identifies integration-level config values
	VariableIntegrationConfig = "integration_config"
	// VariableProviderState identifies persisted provider state values
	VariableProviderState = "provider_state"
	// VariableOrgID identifies the integration owner id
	VariableOrgID = "org_id"
	// VariableIntegrationID identifies the installed integration id
	VariableIntegrationID = "integration_id"
	// VariableAlertType identifies the alert type variable used in ingest mapping expressions
	VariableAlertType = "alert_type"
)

const (
	defaultCELInterruptCheckFrequency = 100
	defaultCELParserRecursionLimit    = 250
	defaultCELExpressionSizeLimit     = 100000
	defaultCELTimeout                 = 100 * time.Millisecond
	defaultEmptyExpressionResult      = true
)

// EvaluatorConfig configures CEL scope evaluator behavior
type EvaluatorConfig struct {
	// Timeout is the maximum expression execution duration
	Timeout time.Duration
	// InterruptCheckFrequency controls CEL interruption check intervals
	InterruptCheckFrequency uint
	// ParserRecursionLimit controls CEL parser recursion depth
	ParserRecursionLimit int
	// ParserExpressionSizeLimit controls maximum expression size
	ParserExpressionSizeLimit int
	// EmptyExpressionResult controls the result returned when expression is empty
	EmptyExpressionResult bool
}

// ScopeVars contains standard variables available to integration scope expressions
type ScopeVars struct {
	// Payload contains provider payload data for filtering
	Payload map[string]any
	// Resource contains provider resource identity values
	Resource string
	// Provider contains provider kind values
	Provider integrationtypes.ProviderType
	// Operation contains operation name values
	Operation integrationtypes.OperationName
	// Config contains operation config values
	Config map[string]any
	// IntegrationConfig contains integration-level config values
	IntegrationConfig map[string]any
	// ProviderState contains integration provider state values
	ProviderState map[string]any
	// OrgID contains integration owner id values
	OrgID string
	// IntegrationID contains installed integration id values
	IntegrationID string
}

// Map converts scope vars into CEL variable bindings
func (v ScopeVars) Map() map[string]any {
	return map[string]any{
		VariablePayload:           v.Payload,
		VariableResource:          v.Resource,
		VariableProvider:          string(v.Provider),
		VariableOperation:         string(v.Operation),
		VariableConfig:            v.Config,
		VariableIntegrationConfig: v.IntegrationConfig,
		VariableProviderState:     v.ProviderState,
		VariableOrgID:             v.OrgID,
		VariableIntegrationID:     v.IntegrationID,
	}
}

// DefaultEvaluatorConfig returns the default scope evaluator configuration
func DefaultEvaluatorConfig() EvaluatorConfig {
	return EvaluatorConfig{
		Timeout:                   defaultCELTimeout,
		InterruptCheckFrequency:   defaultCELInterruptCheckFrequency,
		ParserRecursionLimit:      defaultCELParserRecursionLimit,
		ParserExpressionSizeLimit: defaultCELExpressionSizeLimit,
		EmptyExpressionResult:     defaultEmptyExpressionResult,
	}
}
