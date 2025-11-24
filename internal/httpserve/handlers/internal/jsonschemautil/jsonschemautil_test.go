package jsonschemautil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/core/internal/httpserve/handlers/internal/jsonschemautil"
	"github.com/theopenlane/utils/rout"
)

func TestFieldErrorsFromResultRequired(t *testing.T) {
	schema := map[string]any{
		"type":     "object",
		"required": []string{"projectId"},
		"properties": map[string]any{
			"projectId": map[string]any{"type": "string"},
		},
	}

	result, err := gojsonschema.Validate(
		gojsonschema.NewGoLoader(schema),
		gojsonschema.NewGoLoader(map[string]any{}),
	)
	require.NoError(t, err)

	fieldErrs := jsonschemautil.FieldErrorsFromResult(result)
	require.Len(t, fieldErrs, 1)

	err = fieldErrs[0]
	assert.ErrorIs(t, err, rout.ErrMissingField)

	var fe *rout.FieldError
	require.ErrorAs(t, err, &fe)
	assert.Equal(t, "projectId", fe.Field)
}

func TestFieldErrorsFromResultAdditionalProperty(t *testing.T) {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
	}

	payload := map[string]any{"unexpected": "value"}

	result, err := gojsonschema.Validate(
		gojsonschema.NewGoLoader(schema),
		gojsonschema.NewGoLoader(payload),
	)
	require.NoError(t, err)

	fieldErrs := jsonschemautil.FieldErrorsFromResult(result)
	require.Len(t, fieldErrs, 1)

	err = fieldErrs[0]
	assert.ErrorIs(t, err, rout.ErrRestrictedField)

	var fe *rout.FieldError
	require.ErrorAs(t, err, &fe)
	assert.Equal(t, "unexpected", fe.Field)
}
