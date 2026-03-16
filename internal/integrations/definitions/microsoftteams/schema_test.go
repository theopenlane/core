package microsoftteams

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/providerkit"
)

func TestUserInputSchemaRequiresTenantID(t *testing.T) {
	t.Parallel()

	var document struct {
		Defs map[string]json.RawMessage `json:"$defs"`
	}
	var schema struct {
		Required []string `json:"required"`
	}

	require.NoError(t, json.Unmarshal(providerkit.SchemaFrom[UserInput](), &document))
	require.Contains(t, document.Defs, "UserInput")
	require.NoError(t, json.Unmarshal(document.Defs["UserInput"], &schema))
	assert.ElementsMatch(t, []string{"tenantId"}, schema.Required)
}
