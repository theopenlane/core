package jsonx

import (
	"encoding/json"

	"github.com/xeipuuv/gojsonschema"
)

// ValidateSchema returns the gojsonschema validation result
func ValidateSchema(schema any, document any) (*gojsonschema.Result, error) {
	return gojsonschema.Validate(toJSONLoader(schema), toJSONLoader(document))
}

// ValidationErrorStrings converts schema validation errors into string messages
func ValidationErrorStrings(result *gojsonschema.Result) []string {
	if result == nil || result.Valid() {
		return nil
	}

	errors := make([]string, 0, len(result.Errors()))
	for _, issue := range result.Errors() {
		errors = append(errors, issue.String())
	}

	return errors
}

func toJSONLoader(value any) gojsonschema.JSONLoader {
	switch typed := value.(type) {
	case gojsonschema.JSONLoader:
		return typed
	case []byte:
		return gojsonschema.NewBytesLoader(typed)
	case json.RawMessage:
		return gojsonschema.NewBytesLoader(typed)
	case string:
		return gojsonschema.NewStringLoader(typed)
	default:
		return gojsonschema.NewGoLoader(value)
	}
}
