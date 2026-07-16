package operations

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestInjectLinks_NoRules(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"externalID":"f-1","category":"S3.8"}`)

	result, err := injectLinks(context.Background(), nil, "org-1", nil, "Finding", payload)
	assert.NilError(t, err)
	assert.Equal(t, string(result), string(payload))
}

func TestInjectLinks_UnknownSchema(t *testing.T) {
	t.Parallel()

	_, err := injectLinks(context.Background(), nil, "org-1", []types.LinkRule{{TargetSchema: "Control"}}, "NotASchema", json.RawMessage(`{}`))
	assert.ErrorIs(t, err, ErrLinkTargetSchemaNotFound)
}
