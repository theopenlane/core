package emailruntime

import "github.com/xeipuuv/gojsonschema"

// ValidateJSONSchema validates payload against a JSON schema map
func ValidateJSONSchema(schema map[string]any, payload map[string]any) (bool, error) {
	if len(schema) == 0 {
		return true, nil
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewGoLoader(schema),
		gojsonschema.NewGoLoader(payload),
	)
	if err != nil {
		return false, err
	}

	return result.Valid(), nil
}
