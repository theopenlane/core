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

	integrationtypes "github.com/theopenlane/core/common/integrations/types"
	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/pkg/celx"
)

const (
	mappingSchemaVulnerability        = "Vulnerability"
	defaultCELInterruptCheckFrequency = 100
	defaultCELParserRecursionLimit    = 250
	defaultCELExpressionSizeLimit     = 100000
)

var defaultCELTimeout = 100 * time.Millisecond

// MappingEvaluator runs CEL expressions against integration payloads
// It is intentionally small to keep evaluation consistent across integrations
type MappingEvaluator struct {
	evaluator *celx.Evaluator
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
			decls.NewVariable("payload", types.DynType),
			decls.NewVariable("resource", types.StringType),
			decls.NewVariable("alert_type", types.StringType),
			decls.NewVariable("provider", types.StringType),
			decls.NewVariable("operation", types.StringType),
			decls.NewVariable("org_id", types.StringType),
			decls.NewVariable("integration_id", types.StringType),
			decls.NewVariable("config", types.DynType),
			decls.NewVariable("integration_config", types.DynType),
			decls.NewVariable("provider_state", types.DynType),
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

// mappingOverrideLookup normalizes override keys for fast lookup
func mappingOverrideLookup(config openapi.IntegrationConfig) map[string]openapi.IntegrationMappingOverride {
	out := make(map[string]openapi.IntegrationMappingOverride, len(config.MappingOverrides))
	for key, override := range config.MappingOverrides {
		normalized := normalizeMappingKey(key)
		if normalized == "" {
			continue
		}
		out[normalized] = override
	}

	return out
}

// resolveMappingOverride locates a mapping override using a provider/schema/variant key
func resolveMappingOverride(config openapi.IntegrationConfig, provider integrationtypes.ProviderType, schemaName string, variant string) (openapi.IntegrationMappingOverride, bool) {
	overrides := mappingOverrideLookup(config)
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
		if override, ok := overrides[key]; ok {
			return override, true
		}
	}

	return openapi.IntegrationMappingOverride{}, false
}

// resolveMappingSpec selects an override or default mapping for the given variant
func resolveMappingSpec(config openapi.IntegrationConfig, provider integrationtypes.ProviderType, schemaName string, variant string) (openapi.IntegrationMappingOverride, bool) {
	if override, ok := resolveMappingOverride(config, provider, schemaName, variant); ok {
		return override, true
	}

	return defaultMappingSpec(provider, schemaName, variant)
}

// allowedMappingKeys returns the set of allowed input keys for a schema
func allowedMappingKeys(schema integrationgenerated.IntegrationMappingSchema) map[string]struct{} {
	out := make(map[string]struct{}, len(schema.Fields))
	for _, field := range schema.Fields {
		key := strings.TrimSpace(field.InputKey)
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
	out := make(map[string]any, len(input))
	for key, value := range input {
		if _, ok := allowed[key]; !ok {
			continue
		}
		out[key] = value
	}

	return out
}

// validateMappingOutput checks required fields are present in mapped output
func validateMappingOutput(schema integrationgenerated.IntegrationMappingSchema, input map[string]any) error {
	for _, field := range schema.Fields {
		if !field.Required {
			continue
		}
		value, ok := input[field.InputKey]
		if !ok || value == nil {
			return fmt.Errorf("%w: %s", ErrMappingRequiredField, field.InputKey)
		}
		if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
			return fmt.Errorf("%w: %s", ErrMappingRequiredField, field.InputKey)
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
