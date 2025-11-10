package jsonschemautil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/theopenlane/utils/rout"
	"github.com/xeipuuv/gojsonschema"
)

var errJSONSchemaValidation = errors.New("JSON schema validation failed")

// FieldErrorsFromResult converts gojsonschema validation errors into rout-compatible errors.
func FieldErrorsFromResult(res *gojsonschema.Result) []error {
	if res == nil || res.Valid() {
		return nil
	}

	errs := make([]error, 0, len(res.Errors()))
	for _, issue := range res.Errors() {
		field := issue.Field()
		if property, ok := issue.Details()["property"].(string); ok && property != "" {
			field = property
		}
		if field == "(root)" || strings.TrimSpace(field) == "" {
			field = "payload"
		}

		var err error
		switch issue.Type() {
		case "required":
			err = rout.MissingField(field)
		case "additional_property":
			err = rout.RestrictedField(field)
		default:
			err = &rout.FieldError{
				Field: field,
				Err:   fmt.Errorf("%w: %s", errJSONSchemaValidation, issue.Description()),
			}
		}

		errs = append(errs, err)
	}

	return errs
}
