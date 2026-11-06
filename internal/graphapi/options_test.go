package graphapi

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
)

func TestBulkAuthzObjectType(t *testing.T) {
	tests := map[string]struct {
		field      string
		objectType string
		batched    bool
	}{
		"bulk create":     {"createBulkControl", ent.TypeControl, true},
		"csv bulk create": {"createBulkCSVControl", ent.TypeControl, true},
		"acronym type":    {"createBulkCSVAPIToken", ent.TypeAPIToken, true},
		"single create":   {"createControl", "createControl", false},
		"bulk update":     {"updateBulkCSVControl", "updateBulkCSVControl", false},
		"bulk delete":     {"deleteBulkControl", "deleteBulkControl", false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			objectType, batched := bulkAuthzObjectType(tc.field)

			assert.Check(t, is.Equal(tc.objectType, objectType))
			assert.Check(t, is.Equal(tc.batched, batched))
		})
	}
}
