package model

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestRelationsForService(t *testing.T) {
	rels, err := RelationsForService()
	assert.NilError(t, err)
	assert.Assert(t, rels != nil)

	// spot check
	assert.Check(t, is.Contains(rels, "can_view_control"))
	assert.Check(t, is.Contains(rels, "can_edit_control"))
	assert.Check(t, is.Contains(rels, "can_view_evidence"))
	assert.Check(t, is.Contains(rels, "can_edit_evidence"))
	assert.Check(t, is.Contains(rels, "can_view_api_token"))
	assert.Check(t, is.Contains(rels, "can_edit_api_token"))
	assert.Check(t, is.Contains(rels, "can_delete_api_token"))
}

func TestNormalizeScope(t *testing.T) {
	assert.Equal(t, "can_view", NormalizeScope("read"))
	assert.Equal(t, "can_edit", NormalizeScope("write"))
	assert.Equal(t, "can_delete", NormalizeScope("delete"))
	assert.Equal(t, "can_edit_control", NormalizeScope("write:control"))
	assert.Equal(t, "can_view_evidence", NormalizeScope("read:evidence"))
	assert.Equal(t, "can_view_api_token", NormalizeScope("read:api_token"))
	assert.Equal(t, "can_edit_api_token", NormalizeScope("write:api_token"))
}

func TestScopeOptions(t *testing.T) {
	opts, err := ScopeOptions()
	assert.NilError(t, err)
	assert.Assert(t, opts != nil)

	assert.Check(t, is.Contains(opts, "organization"))
	assert.Check(t, is.Contains(opts["organization"], "read"))
	assert.Check(t, is.Contains(opts["organization"], "write"))

	assert.Check(t, is.Contains(opts, "control"))
	assert.Check(t, is.Contains(opts["control"], "read"))
	assert.Check(t, is.Contains(opts["control"], "write"))
}
