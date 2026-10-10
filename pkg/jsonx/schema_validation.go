package jsonx

import (
	"encoding/json"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// emptyObject is the document validated in place of an absent payload
const emptyObject = "{}"

// invalidJSONIssue is the issue reported when a payload is not well-formed JSON
const invalidJSONIssue = "payload is not valid JSON"

// SchemaError reports the issues a document raised against its JSON schema
type SchemaError struct {
	// Issues lists each validation failure in schema order
	Issues []string
}

// Error joins the issues into one message
func (e *SchemaError) Error() string {
	return strings.Join(e.Issues, "; ")
}

// Unwrap exposes ErrSchemaInvalid so callers can match on the sentinel
func (e *SchemaError) Unwrap() error {
	return ErrSchemaInvalid
}

// Validate checks a payload against a JSON schema, treating an empty schema as unconstrained and an absent payload as an empty object
func Validate(schema, payload json.RawMessage) error {
	if len(schema) == 0 {
		return nil
	}

	if IsEmptyRawMessage(payload) {
		payload = json.RawMessage(emptyObject)
	}

	if !json.Valid(payload) {
		return &SchemaError{Issues: []string{invalidJSONIssue}}
	}

	result, err := ValidateSchema(schema, payload)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	}

	return &SchemaError{Issues: ValidationErrorStrings(result)}
}

// ValidateSchema validates a JSON document against a JSON schema and returns the raw gojsonschema result for caller-specific error handling.
func ValidateSchema(schema any, document any) (*gojsonschema.Result, error) {
	return gojsonschema.Validate(toJSONLoader(schema), toJSONLoader(document))
}

// ValidationErrorStrings converts schema validation errors into string messages.
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
