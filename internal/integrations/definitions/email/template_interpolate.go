package email

import (
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"time"

	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/pkg/jsonx"
)

// interpolatePayload executes Go template expressions in the raw JSON payload
// against a variable map built from the email client config and the payload's
// own recipient and campaign fields. Returns the interpolated JSON ready for
// unmarshal into the typed input struct
func interpolatePayload(client *EmailClient, payload json.RawMessage) (json.RawMessage, error) {
	if len(payload) == 0 {
		return payload, nil
	}

	vars, err := buildTemplateVars(client.Config, payload)
	if err != nil {
		return nil, fmt.Errorf("%w: building template vars: %w", ErrTemplateRenderFailed, err)
	}

	result, err := render.ExecuteTextTemplate("payload", string(payload), vars)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	return json.RawMessage(result), nil
}

// buildTemplateVars constructs the template variable map from the runtime config,
// computed variables, and payload-derived recipient/campaign fields. The keys match
// the variable names advertised by the template catalog
func buildTemplateVars(cfg RuntimeEmailConfig, payload json.RawMessage) (map[string]any, error) {
	var vars map[string]any
	if err := jsonx.RoundTrip(cfg, &vars); err != nil {
		return nil, err
	}

	for name := range nonTemplateConfigFields {
		delete(vars, name)
	}

	vars["year"] = strconv.Itoa(time.Now().Year())

	var payloadVars map[string]any
	if err := json.Unmarshal(payload, &payloadVars); err == nil {
		maps.Copy(vars, payloadVars)
	}

	return vars, nil
}
