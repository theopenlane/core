package operations

import (
	"encoding/json"
	"maps"

	"github.com/samber/lo"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

// OperationTemplate captures a stored configuration and optional override keys
type OperationTemplate struct {
	// Config is the base configuration for the operation
	Config json.RawMessage
	// AllowedOverrides contains keys that may be overridden at runtime
	AllowedOverrides map[string]struct{}
}

// OperationTemplateFor loads an operation template from integration config
func OperationTemplateFor(config *openapi.IntegrationConfig, operation string) (OperationTemplate, bool) {
	if config == nil || len(config.OperationTemplates) == 0 {
		return OperationTemplate{}, false
	}

	if operation == "" {
		return OperationTemplate{}, false
	}

	raw, ok := config.OperationTemplates[operation]
	if !ok {
		return OperationTemplate{}, false
	}

	return operationTemplateFromConfig(raw)
}

// ApplyOperationTemplate merges a template with optional JSON overrides
func ApplyOperationTemplate(template OperationTemplate, overrides json.RawMessage) (json.RawMessage, error) {
	if len(overrides) == 0 {
		return jsonx.CloneRawMessage(template.Config), nil
	}

	if len(template.AllowedOverrides) == 0 {
		return nil, ErrOperationTemplateOverridesNotAllowed
	}

	var overrideMap map[string]json.RawMessage
	if err := json.Unmarshal(overrides, &overrideMap); err != nil {
		return nil, err
	}

	for key := range overrideMap {
		if _, ok := template.AllowedOverrides[key]; !ok {
			return nil, ErrOperationTemplateOverrideNotAllowed
		}
	}

	var baseMap map[string]json.RawMessage
	if len(template.Config) > 0 {
		if err := json.Unmarshal(template.Config, &baseMap); err != nil {
			return nil, err
		}
	}

	if baseMap == nil {
		baseMap = map[string]json.RawMessage{}
	}

	maps.Copy(baseMap, overrideMap)

	return json.Marshal(baseMap)
}

// ResolveOperationConfig merges stored templates with optional JSON overrides
func ResolveOperationConfig(config *openapi.IntegrationConfig, operation string, overrides json.RawMessage) (json.RawMessage, error) {
	if template, ok := OperationTemplateFor(config, operation); ok {
		return ApplyOperationTemplate(template, overrides)
	}

	if len(overrides) == 0 {
		return nil, nil
	}

	return jsonx.CloneRawMessage(overrides), nil
}

// parseOverrideKeys normalizes and deduplicates override keys
func parseOverrideKeys(values []string) map[string]struct{} {
	filtered := lo.Filter(values, func(value string, _ int) bool {
		return value != ""
	})
	if len(filtered) == 0 {
		return nil
	}

	return mapx.MapSetFromSlice(lo.Uniq(filtered))
}

// operationTemplateFromConfig converts stored template config into an OperationTemplate
func operationTemplateFromConfig(template openapi.IntegrationOperationTemplate) (OperationTemplate, bool) {
	overrides := parseOverrideKeys(template.AllowOverrides)
	if len(template.Config) == 0 && overrides == nil {
		return OperationTemplate{}, false
	}

	config := jsonx.CloneRawMessage(template.Config)
	if config == nil {
		config = json.RawMessage("{}")
	}

	return OperationTemplate{
		Config:           config,
		AllowedOverrides: overrides,
	}, true
}
