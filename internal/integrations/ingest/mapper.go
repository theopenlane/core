package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/decls"
	"github.com/google/cel-go/common/types"
	"github.com/samber/lo"

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/pkg/celx"
)

const (
	mappingSchemaVulnerability        = integrationgenerated.IntegrationMappingSchemaVulnerability
	defaultCELInterruptCheckFrequency = 100
	defaultCELParserRecursionLimit    = 250
	defaultCELExpressionSizeLimit     = 100000

	// mappingKeySplitParts defines how many parts to split a mapping key into
	mappingKeySplitParts = 3
	// mappingKeyTwoParts indicates a mapping key with two parts
	mappingKeyTwoParts = 2
)

const (
	mappingVarPayload           = "payload"
	mappingVarResource          = "resource"
	mappingVarAlertType         = "alert_type"
	mappingVarProvider          = "provider"
	mappingVarOperation         = "operation"
	mappingVarOrgID             = "org_id"
	mappingVarIntegrationID     = "integration_id"
	mappingVarConfig            = "config"
	mappingVarIntegrationConfig = "integration_config"
	mappingVarProviderState     = "provider_state"
)

var defaultCELTimeout = 100 * time.Millisecond

var normalizedMappingSchemas = func() map[string]struct{} {
	out := map[string]struct{}{}
	for name := range integrationgenerated.IntegrationMappingSchemas {
		key := normalizeMappingKey(name)
		if key == "" {
			continue
		}
		out[key] = struct{}{}
	}

	return out
}()

// Evaluator defines the interface for mapping expression evaluation
type Evaluator interface {
	// EvaluateFilter evaluates a CEL filter expression and returns a boolean
	EvaluateFilter(ctx context.Context, expression string, vars map[string]any) (bool, error)
	// EvaluateMap evaluates a CEL expression and returns a JSON object map
	EvaluateMap(ctx context.Context, expression string, vars map[string]any) (map[string]any, error)
}

// MappingEvaluator runs CEL expressions against integration payloads
// It is intentionally small to keep evaluation consistent across integrations
type MappingEvaluator struct {
	evaluator *celx.Evaluator
}

// compile-time interface check
var _ Evaluator = (*MappingEvaluator)(nil)

// MappingVars holds CEL variables for integration mappings
type MappingVars struct {
	// Payload holds the raw provider payload for mapping
	Payload map[string]any
	// Resource identifies the upstream resource associated with the payload
	Resource string
	// AlertType identifies the alert type for the payload
	AlertType string
	// Provider identifies the integration provider
	Provider integrationtypes.ProviderType
	// Operation identifies the operation that produced the payload
	Operation integrationtypes.OperationName
	// OrgID identifies the organization that owns the integration
	OrgID string
	// IntegrationID identifies the integration record
	IntegrationID string
	// Config holds operation configuration values
	Config map[string]any
	// IntegrationConfig holds integration-level configuration values
	IntegrationConfig map[string]any
	// ProviderState holds provider state captured during activation
	ProviderState map[string]any
}

// Map converts MappingVars into the CEL variable map
func (m MappingVars) Map() map[string]any {
	return map[string]any{
		mappingVarPayload:           m.Payload,
		mappingVarResource:          m.Resource,
		mappingVarAlertType:         m.AlertType,
		mappingVarProvider:          string(m.Provider),
		mappingVarOperation:         string(m.Operation),
		mappingVarOrgID:             m.OrgID,
		mappingVarIntegrationID:     m.IntegrationID,
		mappingVarConfig:            m.Config,
		mappingVarIntegrationConfig: m.IntegrationConfig,
		mappingVarProviderState:     m.ProviderState,
	}
}

// NewMappingEvaluator creates a CEL evaluator for integration mappings
func NewMappingEvaluator() (*MappingEvaluator, error) {
	env, err := newMappingEnv()
	if err != nil {
		return nil, err
	}

	evalCfg := celx.EvalConfig{
		Timeout:                 defaultCELTimeout,
		InterruptCheckFrequency: defaultCELInterruptCheckFrequency,
		EvalOptimize:            true,
	}

	return &MappingEvaluator{
		evaluator: celx.NewEvaluator(env, evalCfg),
	}, nil
}

// newMappingEnv builds the CEL environment for mapping expressions
func newMappingEnv() (*cel.Env, error) {
	cfg := celx.EnvConfig{
		ParserRecursionLimit:        defaultCELParserRecursionLimit,
		ParserExpressionSizeLimit:   defaultCELExpressionSizeLimit,
		ExtendedValidations:         true,
		CrossTypeNumericComparisons: true,
	}

	return celx.NewEnv(cfg,
		cel.VariableDecls(
			decls.NewVariable(mappingVarPayload, types.DynType),
			decls.NewVariable(mappingVarResource, types.StringType),
			decls.NewVariable(mappingVarAlertType, types.StringType),
			decls.NewVariable(mappingVarProvider, types.StringType),
			decls.NewVariable(mappingVarOperation, types.StringType),
			decls.NewVariable(mappingVarOrgID, types.StringType),
			decls.NewVariable(mappingVarIntegrationID, types.StringType),
			decls.NewVariable(mappingVarConfig, types.DynType),
			decls.NewVariable(mappingVarIntegrationConfig, types.DynType),
			decls.NewVariable(mappingVarProviderState, types.DynType),
		),
	)
}

// EvaluateFilter evaluates a CEL filter expression and returns a boolean
func (m *MappingEvaluator) EvaluateFilter(ctx context.Context, expression string, vars map[string]any) (bool, error) {
	out, _, err := m.evaluator.Evaluate(ctx, expression, vars)
	if err != nil {
		return false, err
	}
	if out == nil {
		return false, ErrMappingOutputEmpty
	}
	if out.Type() != types.BoolType {
		return false, ErrMappingFilterType
	}

	value, ok := out.Value().(bool)
	if !ok {
		value = out.Equal(types.True) == types.True
	}

	return value, nil
}

// EvaluateMap evaluates a CEL expression and returns a JSON object map
func (m *MappingEvaluator) EvaluateMap(ctx context.Context, expression string, vars map[string]any) (map[string]any, error) {
	out, err := m.evaluator.EvaluateJSONMap(ctx, expression, vars)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, ErrMappingOutputEmpty
	}

	return out, nil
}

// normalizeMappingKey lowercases mapping override keys for comparison
func normalizeMappingKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

type mappingOverrideIndex struct {
	overrides         map[string]openapi.IntegrationMappingOverride
	hasSchema         map[string]struct{}
	hasProviderSchema map[string]struct{}
}

// newMappingOverrideIndex builds a lookup index for mapping overrides
func newMappingOverrideIndex(config openapi.IntegrationConfig) mappingOverrideIndex {
	init := mappingOverrideIndex{
		overrides:         make(map[string]openapi.IntegrationMappingOverride, len(config.MappingOverrides)),
		hasSchema:         map[string]struct{}{},
		hasProviderSchema: map[string]struct{}{},
	}

	return lo.Reduce(lo.Entries(config.MappingOverrides), func(acc mappingOverrideIndex, entry lo.Entry[string, openapi.IntegrationMappingOverride], _ int) mappingOverrideIndex {
		normalized := normalizeMappingKey(entry.Key)
		if normalized == "" {
			return acc
		}

		acc.overrides[normalized] = entry.Value
		schemaKey, providerKey := mappingKeyParts(normalized)
		if schemaKey != "" {
			acc.hasSchema[schemaKey] = struct{}{}
		}

		if providerKey != "" && schemaKey != "" {
			acc.hasProviderSchema[fmt.Sprintf("%s:%s", providerKey, schemaKey)] = struct{}{}
		}

		return acc
	}, init)
}

// HasAny reports whether any overrides exist for the provider and schema
func (m mappingOverrideIndex) HasAny(provider integrationtypes.ProviderType, schemaName string) bool {
	schemaKey := normalizeMappingKey(schemaName)
	if schemaKey == "" {
		return false
	}
	if _, ok := m.hasSchema[schemaKey]; ok {
		return true
	}

	providerKey := normalizeMappingKey(string(provider))
	if providerKey == "" {
		return false
	}

	_, ok := m.hasProviderSchema[fmt.Sprintf("%s:%s", providerKey, schemaKey)]

	return ok
}

// Resolve selects the most specific override for the provider, schema, and variant
func (m mappingOverrideIndex) Resolve(provider integrationtypes.ProviderType, schemaName string, variant string) (openapi.IntegrationMappingOverride, bool) {
	providerKey := normalizeMappingKey(string(provider))
	schemaKey := normalizeMappingKey(schemaName)
	variantKey := normalizeMappingKey(variant)

	candidates := []string{}
	if providerKey != "" && schemaKey != "" && variantKey != "" {
		candidates = append(candidates, fmt.Sprintf("%s:%s:%s", providerKey, schemaKey, variantKey))
	}

	if providerKey != "" && schemaKey != "" {
		candidates = append(candidates, fmt.Sprintf("%s:%s", providerKey, schemaKey))
	}

	if schemaKey != "" && variantKey != "" {
		candidates = append(candidates, fmt.Sprintf("%s:%s", schemaKey, variantKey))
	}

	if schemaKey != "" {
		candidates = append(candidates, schemaKey)
	}

	for _, key := range candidates {
		if override, ok := m.overrides[key]; ok {
			return override, true
		}
	}

	return openapi.IntegrationMappingOverride{}, false
}

// mappingKeyParts splits a normalized override key into schema and provider parts
func mappingKeyParts(key string) (schemaKey string, providerKey string) {
	parts := strings.SplitN(key, ":", mappingKeySplitParts)
	switch len(parts) {
	case 1:
		return parts[0], ""
	case mappingKeyTwoParts:
		if isMappingSchema(parts[0]) {
			return parts[0], ""
		}
		if isMappingSchema(parts[1]) {
			return parts[1], parts[0]
		}
		return parts[0], ""
	default:
		if isMappingSchema(parts[1]) {
			return parts[1], parts[0]
		}
		if isMappingSchema(parts[0]) {
			return parts[0], ""
		}
		return parts[1], parts[0]
	}
}

// isMappingSchema reports whether a name matches a known mapping schema
func isMappingSchema(value string) bool {
	_, ok := normalizedMappingSchemas[value]

	return ok
}

// resolveMappingSpecWithIndex resolves overrides using a precomputed index
func resolveMappingSpecWithIndex(index mappingOverrideIndex, provider integrationtypes.ProviderType, schemaName string, variant string) (openapi.IntegrationMappingOverride, bool) {
	if override, ok := index.Resolve(provider, schemaName, variant); ok {
		return override, true
	}

	return defaultMappingSpec(provider, schemaName, variant)
}

// allowedMappingKeys returns the set of allowed input keys for a schema
func allowedMappingKeys(schema integrationgenerated.IntegrationMappingSchema) map[string]struct{} {
	if len(schema.AllowedKeys) > 0 {
		return schema.AllowedKeys
	}
	out := make(map[string]struct{}, len(schema.Fields))
	for _, field := range schema.Fields {
		key := field.InputKey
		if key == "" {
			continue
		}
		out[key] = struct{}{}
	}

	return out
}

// filterMappingOutput strips fields that are not part of the schema mapping
func filterMappingOutput(schema integrationgenerated.IntegrationMappingSchema, input map[string]any) map[string]any {
	allowed := allowedMappingKeys(schema)
	return lo.PickBy(input, func(key string, _ any) bool {
		_, ok := allowed[key]

		return ok
	})
}

// validateMappingOutput checks required fields are present in mapped output
func validateMappingOutput(schema integrationgenerated.IntegrationMappingSchema, input map[string]any) error {
	requiredKeys := schema.RequiredKeys
	if len(requiredKeys) == 0 {
		for _, field := range schema.Fields {
			if field.Required {
				requiredKeys = append(requiredKeys, field.InputKey)
			}
		}
	}

	for _, key := range requiredKeys {
		value, ok := input[key]
		if !ok || value == nil {
			return ErrMappingRequiredField
		}
		if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
			return ErrMappingRequiredField
		}
	}

	return nil
}

// toMap converts an arbitrary value into a JSON object map
func toMap(value any) (map[string]any, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var out any
	if err := json.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}

	if out == nil {
		return map[string]any{}, nil
	}

	mapped, ok := out.(map[string]any)
	if !ok {
		return nil, ErrMappingOutputEmpty
	}

	return mapped, nil
}
