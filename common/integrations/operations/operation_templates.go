package operations

import (
	"strings"

	commonhelpers "github.com/theopenlane/core/common/helpers"
	openapi "github.com/theopenlane/core/common/openapi"
)

// OperationTemplate captures a stored configuration and optional override keys.
type OperationTemplate struct {
	// Config is the base configuration for the operation.
	Config map[string]any
	// AllowedOverrides contains keys that may be overridden at runtime.
	AllowedOverrides map[string]struct{}
}

// OperationTemplateFor loads an operation template from integration config.
func OperationTemplateFor(config *openapi.IntegrationConfig, operation string) (OperationTemplate, bool) {
	if config == nil || len(config.OperationTemplates) == 0 {
		return OperationTemplate{}, false
	}

	operation = strings.TrimSpace(operation)
	if operation == "" {
		return OperationTemplate{}, false
	}

	raw, ok := config.OperationTemplates[operation]
	if !ok {
		return OperationTemplate{}, false
	}

	return operationTemplateFromConfig(raw)
}

// ApplyOperationTemplate merges a template with optional overrides.
func ApplyOperationTemplate(template OperationTemplate, overrides map[string]any) (map[string]any, error) {
	result := commonhelpers.DeepCloneMap(template.Config)

	if len(overrides) == 0 {
		return result, nil
	}

	if len(template.AllowedOverrides) == 0 {
		return nil, ErrOperationTemplateOverridesNotAllowed
	}

	for key, value := range overrides {
		if _, ok := template.AllowedOverrides[key]; !ok {
			return nil, ErrOperationTemplateOverrideNotAllowed
		}
		if result == nil {
			result = map[string]any{}
		}
		result[key] = value
	}

	return result, nil
}

// ResolveOperationConfig merges stored templates with optional overrides.
func ResolveOperationConfig(config *openapi.IntegrationConfig, operation string, overrides map[string]any) (map[string]any, error) {
	if template, ok := OperationTemplateFor(config, operation); ok {
		return ApplyOperationTemplate(template, overrides)
	}
	if len(overrides) == 0 {
		return nil, nil
	}
	return commonhelpers.DeepCloneMap(overrides), nil
}

func parseOverrideKeys(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}

	out := make(map[string]struct{}, len(values))
	for _, item := range values {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out[item] = struct{}{}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func operationTemplateFromConfig(template openapi.IntegrationOperationTemplate) (OperationTemplate, bool) {
	config := commonhelpers.DeepCloneMap(template.Config)
	overrides := parseOverrideKeys(template.AllowOverrides)
	if config == nil && overrides == nil {
		return OperationTemplate{}, false
	}
	if config == nil {
		config = map[string]any{}
	}

	return OperationTemplate{
		Config:           config,
		AllowedOverrides: overrides,
	}, true
}
