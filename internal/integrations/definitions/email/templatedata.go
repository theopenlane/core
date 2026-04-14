package email

import (
	"maps"
	"time"

	"github.com/theopenlane/core/pkg/jsonx"
)

// buildTemplateData constructs the template variable map from RuntimeEmailConfig
// and user-defined vars. Config fields are serialized via jsonx.ToMap (using the
// camelCase json tags on RuntimeEmailConfig), then computed fields are added.
// All variables are flat top-level keys. Vars have the highest merge precedence
func buildTemplateData(config RuntimeEmailConfig, vars map[string]any) (map[string]any, error) {
	data, err := jsonx.ToMap(config)
	if err != nil {
		return nil, err
	}

	data["year"] = time.Now().Year()

	maps.Copy(data, vars)

	return data, nil
}
