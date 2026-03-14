package operations

import (
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

// resolvedTemplate captures a stored configuration and its permitted override keys
type resolvedTemplate struct {
	// Config is the base configuration for the operation
	Config json.RawMessage
	// AllowedOverrides contains keys that may be overridden at runtime
	AllowedOverrides map[string]struct{}
}

// operationTemplateFor loads an operation template from integration config
func operationTemplateFor(config types.IntegrationConfig, operation string) (resolvedTemplate, bool) {
	if len(config.OperationTemplates) == 0 || operation == "" {
		return resolvedTemplate{}, false
	}

	raw, ok := config.OperationTemplates[operation]
	if !ok {
		return resolvedTemplate{}, false
	}

	return templateFromConfig(raw)
}

// applyOperationTemplate merges a template with optional JSON overrides
func applyOperationTemplate(template resolvedTemplate, overrides json.RawMessage) (json.RawMessage, error) {
	if len(overrides) == 0 {
		return jsonx.CloneRawMessage(template.Config), nil
	}

	if len(template.AllowedOverrides) == 0 {
		return nil, ErrOperationTemplateOverridesNotAllowed
	}

	overrideMap, err := jsonx.ToRawMap(overrides)
	if err != nil {
		return nil, err
	}

	for key := range overrideMap {
		if _, ok := template.AllowedOverrides[key]; !ok {
			return nil, ErrOperationTemplateOverrideNotAllowed
		}
	}

	merged, _, err := jsonx.MergeObjectMap(template.Config, overrideMap)
	if err != nil {
		return nil, err
	}

	return merged, nil
}

// ResolveOperationConfig merges stored templates with optional JSON overrides
func ResolveOperationConfig(config types.IntegrationConfig, operation string, overrides json.RawMessage) (json.RawMessage, error) {
	if template, ok := operationTemplateFor(config, operation); ok {
		return applyOperationTemplate(template, overrides)
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

// templateFromConfig converts a stored OperationTemplate into a resolvedTemplate
func templateFromConfig(template types.OperationTemplate) (resolvedTemplate, bool) {
	overrides := parseOverrideKeys(template.AllowOverrides)
	if len(template.Config) == 0 && overrides == nil {
		return resolvedTemplate{}, false
	}

	config := jsonx.CloneRawMessage(template.Config)
	if config == nil {
		config = json.RawMessage("{}")
	}

	return resolvedTemplate{
		Config:           config,
		AllowedOverrides: overrides,
	}, true
}
