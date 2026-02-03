package helpers

import (
	"strings"

	"github.com/samber/lo"

	commonhelpers "github.com/theopenlane/core/common/helpers"
)

// OperationTemplate captures a stored configuration and optional override keys.
type OperationTemplate struct {
	// Config is the base configuration for the operation.
	Config map[string]any
	// AllowedOverrides contains keys that may be overridden at runtime.
	AllowedOverrides map[string]struct{}
}

// OperationTemplateFor loads an operation template from integration metadata.
func OperationTemplateFor(metadata map[string]any, operation string) (OperationTemplate, bool) {
	if len(metadata) == 0 {
		return OperationTemplate{}, false
	}

	operation = strings.TrimSpace(operation)
	if operation == "" {
		return OperationTemplate{}, false
	}

	templates := operationTemplateCatalog(metadata)
	if len(templates) == 0 {
		return OperationTemplate{}, false
	}

	raw, ok := templates[operation]
	if !ok {
		return OperationTemplate{}, false
	}

	config, overrides := parseOperationTemplateEntry(raw)
	if config == nil && overrides == nil {
		return OperationTemplate{}, false
	}
	if config == nil {
		config = map[string]any{}
	}

	return OperationTemplate{
		Config:           commonhelpers.DeepCloneMap(config),
		AllowedOverrides: overrides,
	}, true
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

func operationTemplateCatalog(metadata map[string]any) map[string]any {
	if len(metadata) == 0 {
		return nil
	}

	if raw, ok := metadata["operation_templates"]; ok {
		if cast := toStringAnyMap(raw); len(cast) > 0 {
			return cast
		}
	}
	if raw, ok := metadata["operationTemplates"]; ok {
		if cast := toStringAnyMap(raw); len(cast) > 0 {
			return cast
		}
	}

	return nil
}

func parseOperationTemplateEntry(raw any) (map[string]any, map[string]struct{}) {
	entry := toStringAnyMap(raw)
	if len(entry) == 0 {
		return nil, nil
	}

	overrides := parseOverrideKeys(entry["allow_overrides"], entry["allowOverrides"])
	if configRaw, ok := entry["config"]; ok {
		return toStringAnyMap(configRaw), overrides
	}

	// Treat the entry itself as the config if no nested config is provided.
	config := commonhelpers.DeepCloneMap(entry)
	delete(config, "allow_overrides")
	delete(config, "allowOverrides")

	return config, overrides
}

func parseOverrideKeys(values ...any) map[string]struct{} {
	items := lo.FlatMap(values, func(value any, _ int) []string {
		return stringsFromAny(value)
	})
	items = lo.Filter(items, func(item string, _ int) bool {
		return item != ""
	})
	if len(items) == 0 {
		return nil
	}
	return lo.SliceToMap(items, func(item string) (string, struct{}) {
		return item, struct{}{}
	})
}
