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
	assert.Equal(t, "can_edit_control", NormalizeScope("control:write"))
	assert.Equal(t, "can_view_evidence", NormalizeScope("evidence:read"))
	assert.Equal(t, "can_view_api_token", NormalizeScope("api_token:read"))
	assert.Equal(t, "can_edit_api_token", NormalizeScope("api_token:write"))
}

func TestOrganizationRolesExistInModel(t *testing.T) {
	roles, err := OrganizationRoles()
	assert.NilError(t, err)
	assert.Assert(t, len(roles) > 0)

	model, err := GetRolesAuthorizationModel()
	assert.NilError(t, err)

	relations := map[string]struct{}{}

	for _, td := range model.GetTypeDefinitions() {
		if td.GetType() != "organization" || td.Metadata == nil || td.Metadata.Relations == nil {
			continue
		}

		for rel := range *td.Metadata.Relations {
			relations[rel] = struct{}{}
		}
	}

	for _, role := range roles {
		_, ok := relations[role.ID]
		assert.Check(t, ok, "@role %q is annotated on %s, which is not an organization relation", role.Name, role.ID)
	}

	assert.Check(t, is.Contains(relations, "campaign_manager"))
	assert.Check(t, is.Contains(relations, "integration_manager"))
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
