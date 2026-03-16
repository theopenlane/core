package catalog

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/definition"
)

func TestBuiltInDefinitionsUserInputSchemaIncludesFilterExpr(t *testing.T) {
	t.Parallel()

	definitions, err := definition.BuildAll(Builders(Config{})...)
	require.NoError(t, err)

	for _, def := range definitions {
		if def.UserInput == nil || len(def.UserInput.Schema) == 0 {
			continue
		}

		properties := schemaProperties(t, def.UserInput.Schema)
		require.Containsf(t, properties, "filterExpr", "%s user input schema missing filterExpr", def.Slug)
	}
}

func schemaProperties(t *testing.T, raw json.RawMessage) map[string]json.RawMessage {
	t.Helper()

	var document struct {
		Ref        string                     `json:"$ref"`
		Defs       map[string]json.RawMessage `json:"$defs"`
		Properties map[string]json.RawMessage `json:"properties"`
	}

	require.NoError(t, json.Unmarshal(raw, &document))

	if document.Ref == "" {
		return document.Properties
	}

	defName := strings.TrimPrefix(document.Ref, "#/$defs/")
	require.Contains(t, document.Defs, defName)

	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
	}

	require.NoError(t, json.Unmarshal(document.Defs[defName], &schema))

	return schema.Properties
}
