package email

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/utils/jsonx"
)

// interpolatePayload executes Go template expressions in the raw JSON payload against a variable map
func interpolatePayload(client *Client, payload json.RawMessage, operationVars []OperationVariable, systemKeys []string) (json.RawMessage, error) {
	if len(payload) == 0 {
		return payload, nil
	}

	var document map[string]any
	if err := jsonx.RoundTrip(payload, &document); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	vars, err := buildTemplateVars(client.Config, document, operationVars)
	if err != nil {
		return nil, fmt.Errorf("%w: building template vars: %w", ErrTemplateRenderFailed, err)
	}

	interpolated, err := interpolateAuthoredValues(document, vars, systemKeys)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	result, err := jsonx.ToRawMessage(interpolated)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	return result, nil
}

// buildTemplateVars constructs the template variable map. Only allowlisted config fields are
// included, so operational and secret config (apikey, resendsecret, provider, URLs, …) is never
// interpolatable into rendered output regardless of which client's config backs the render
func buildTemplateVars(cfg RuntimeEmailConfig, payloadVars map[string]any, operationVars []OperationVariable) (map[string]any, error) {
	var cfgVars map[string]any
	if err := jsonx.RoundTrip(cfg, &cfgVars); err != nil {
		return nil, err
	}

	vars := make(map[string]any, len(customerTemplateConfigFields)+len(payloadVariables)+1)
	for name := range customerTemplateConfigFields {
		if value, ok := cfgVars[name]; ok {
			vars[name] = value
		}
	}

	vars["year"] = strconv.Itoa(time.Now().Year())

	for _, variable := range payloadVariables {
		setPayloadVar(vars, payloadVars, variable.Name, variable.Name)
	}

	for _, variable := range operationVars {
		setPayloadVar(vars, payloadVars, variable.Name, variable.PayloadKey)
	}

	return vars, nil
}

// setPayloadVar sets the named variable from the payload key, defaulting to empty when the key is absent
func setPayloadVar(vars, payloadVars map[string]any, name, key string) {
	vars[name] = ""

	if value, ok := payloadVars[key]; ok {
		vars[name] = value
	}
}

// baseSystemKeys are the per-send fields every operation receives from the system
var baseSystemKeys = slices.Concat(payloadKeys[RecipientInfo](), payloadKeys[CampaignContext]())

// payloadKeys returns the top-level JSON keys of T
func payloadKeys[T any]() []string {
	return lo.Map(jsonx.PropertyDescriptors[T](), func(p jsonx.PropertyDescriptor, _ int) string {
		return p.Name
	})
}

// isSystemKey reports whether the payload key is supplied by the system, either for every operation or for this one
func isSystemKey(key string, systemKeys []string) bool {
	return slices.Contains(baseSystemKeys, key) || slices.Contains(systemKeys, key)
}

// interpolateAuthoredValues executes template expressions in the top-level payload values, skipping system-supplied fields
func interpolateAuthoredValues(document map[string]any, vars map[string]any, systemKeys []string) (map[string]any, error) {
	out := make(map[string]any, len(document))

	for key, value := range document {
		if isSystemKey(key, systemKeys) {
			out[key] = value

			continue
		}

		rendered, err := interpolateTemplateValue(value, vars)
		if err != nil {
			return nil, err
		}

		out[key] = rendered
	}

	return out, nil
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
