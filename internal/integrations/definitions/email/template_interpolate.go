package email

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/pkg/jsonx"
)

// interpolatePayload executes Go template expressions in the raw JSON payload against a variable map
func interpolatePayload(client *Client, payload json.RawMessage) (json.RawMessage, error) {
	if len(payload) == 0 {
		return payload, nil
	}

	var document map[string]any
	if err := jsonx.RoundTrip(payload, &document); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	vars, err := buildTemplateVars(client.Config, document)
	if err != nil {
		return nil, fmt.Errorf("%w: building template vars: %w", ErrTemplateRenderFailed, err)
	}

	interpolated, err := interpolateTemplateValue(document, vars)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	result, err := jsonx.ToRawMessage(interpolated)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	return result, nil
}

// buildTemplateVars constructs the template variable map
func buildTemplateVars(cfg RuntimeEmailConfig, payloadVars map[string]any) (map[string]any, error) {
	var vars map[string]any
	if err := jsonx.RoundTrip(cfg, &vars); err != nil {
		return nil, err
	}

	for name := range nonTemplateConfigFields {
		delete(vars, name)
	}

	vars["year"] = strconv.Itoa(time.Now().Year())

	for _, variable := range payloadVariables {
		vars[variable.Name] = ""

		if value, ok := payloadVars[variable.Name]; ok {
			vars[variable.Name] = value
		}
	}

	return vars, nil
}

// interpolateTemplateValue recursively executes template expressions in strings within the value
func interpolateTemplateValue(value any, vars map[string]any) (any, error) {
	switch typed := value.(type) {
	case string:
		return render.ExecuteTextTemplate("payload-value", typed, vars)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			rendered, err := interpolateTemplateValue(item, vars)
			if err != nil {
				return nil, err
			}
			out = append(out, rendered)
		}

		return out, nil
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			rendered, err := interpolateTemplateValue(item, vars)
			if err != nil {
				return nil, err
			}
			out[key] = rendered
		}

		return out, nil
	default:
		return value, nil
	}
}
