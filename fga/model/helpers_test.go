package model

import (
	"slices"
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
	assert.Check(t, is.Contains(relations, "share_link_manager"))
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

func TestScopeList(t *testing.T) {
	scopes, err := ScopeList()
	assert.NilError(t, err)

	assert.Check(t, is.Contains(scopes, "control:read"))
	assert.Check(t, is.Contains(scopes, "control:write"))
	assert.Check(t, is.Contains(scopes, "api_token:delete"))
	assert.Check(t, slices.IsSorted(scopes))
}

func TestScopeRelation(t *testing.T) {
	tests := []struct {
		relation string
		object   string
		want     string
	}{
		{relation: CanView, object: "control", want: "can_view_control"},
		{relation: CanEdit, object: "vendor_risk_score", want: "can_edit_vendor_risk_score"},
		{relation: CanDelete, object: "api_token", want: "can_delete_api_token"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			assert.Check(t, is.Equal(ScopeRelation(tc.relation, tc.object), tc.want))
		})
	}
}

func TestParseScopeRelation(t *testing.T) {
	tests := []struct {
		scope        string
		wantRelation string
		wantObject   string
		wantOK       bool
	}{
		{scope: "control:read", wantRelation: CanView, wantObject: "control", wantOK: true},
		{scope: "vendor_risk_score:write", wantRelation: CanEdit, wantObject: "vendor_risk_score", wantOK: true},
		{scope: "Risk:Delete", wantRelation: CanDelete, wantObject: "risk", wantOK: true},
		{scope: "control:admin"},
		{scope: "read"},
		{scope: ""},
	}

	for _, tc := range tests {
		t.Run(tc.scope, func(t *testing.T) {
			relation, object, ok := ParseScopeRelation(tc.scope)
			assert.Check(t, is.Equal(ok, tc.wantOK))
			assert.Check(t, is.Equal(relation, tc.wantRelation))
			assert.Check(t, is.Equal(object, tc.wantObject))
		})
	}
}
