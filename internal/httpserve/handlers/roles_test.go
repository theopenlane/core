package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestOrganizationRolesHandler(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce("GET", "account/organization-roles", suite.h.RolesHandler)

	req := httptest.NewRequest(http.MethodGet, "/account/organization-roles", nil)
	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req.WithContext(suite.testUser1.UserCtx))

	response := recorder.Result()
	defer response.Body.Close()

	var out models.RolesResponse
	assert.NilError(t, json.NewDecoder(response.Body).Decode(&out))

	assert.Check(t, is.Equal(http.StatusOK, recorder.Code))
	assert.Check(t, out.Success)

	assert.Check(t, is.Contains(out.Roles, models.OrganizationRole{
		ID:          "policy_manager",
		Name:        "Policy Manager",
		Description: "Manage all policies and procedures",
	}))
	assert.Check(t, is.Contains(out.Roles, models.OrganizationRole{
		ID:          "risk_manager",
		Name:        "Risk Manager",
		Description: "Manage risks, vulnerabilities, findings, and remediation",
	}))
}

func TestOrganizationRolesAssignmentHandler(t *testing.T) {
	suite := suite.setupParallelTest(t).withTestUsers()

	suite.registerRouteOnce("POST", "account/organization-roles", suite.h.AssignOrganizationRolesHandler)
	suite.registerRouteOnce("DELETE", "account/organization-roles", suite.h.DeleteOrganizationRolesHandler)

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ownerCtx := auth.NewTestContextWithOrgID(suite.testUser1.ID, suite.testUser1.OrganizationID, auth.WithOrganizationRole(auth.OwnerRole))
	ownerCtx = auth.WithInternalOperationContext(ownerCtx)
	ownerCtx = generated.NewContext(ownerCtx, suite.db)

	group, err := suite.db.Group.Create().
		SetName("Role Test Group " + suite.testUser2.ID).
		SetDescription("Group for organization role assignment tests").
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(ctx)
	assert.NilError(t, err)

	member := suite.userBuilder(ctx)
	memberRole := enums.RoleMember
	err = suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		UserID:         member.ID,
		OrganizationID: suite.testUser1.OrganizationID,
		Role:           &memberRole,
	}).Exec(ctx)
	assert.NilError(t, err)

	memberCtx := auth.NewTestContextWithOrgID(member.ID, suite.testUser1.OrganizationID, auth.WithOrganizationRole(auth.MemberRole))
	memberCtx = auth.WithInternalOperationContext(memberCtx)
	memberCtx = generated.NewContext(memberCtx, suite.db)

	cases := []struct {
		name       string
		method     string
		request    models.OrganizationRolesRequest
		ctx        context.Context
		statusCode int
		success    bool
	}{
		{
			name:   "assign new role to a user",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser1.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{suite.testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusOK,
			success:    true,
		},
		{
			name:   "assign existing role to a user",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser1.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{suite.testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusOK,
			success:    true,
		},
		{
			name:   "assign new role to a group",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser1.OrganizationID,
				Role:           "risk_manager",
				GroupIDs:       []string{group.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusOK,
			success:    true,
		},
		{
			name:   "delete role from a group",
			method: http.MethodDelete,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser1.OrganizationID,
				Role:           "risk_manager",
				GroupIDs:       []string{group.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusOK,
			success:    true,
		},
		{
			name:   "delete role from a user",
			method: http.MethodDelete,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser1.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{suite.testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusOK,
			success:    true,
		},
		{
			name:   "invalid role",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser1.OrganizationID,
				Role:           "not_a_role",
				UserIDs:        []string{suite.testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "missing subjects to apply roles to",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser1.OrganizationID,
				Role:           "policy_manager",
			},
			ctx:        ownerCtx,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "assign role to user from unauthorized organization",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser2.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{suite.testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "member cannot assign role to a user",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser1.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{suite.testUser2.ID},
			},
			ctx:        memberCtx,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "member cannot assign role to a group",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: suite.testUser1.OrganizationID,
				Role:           "risk_manager",
				GroupIDs:       []string{group.ID},
			},
			ctx:        memberCtx,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(tc.request)
			assert.NilError(t, err)

			req := httptest.NewRequest(tc.method, "/account/organization-roles", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req.WithContext(tc.ctx))

			response := recorder.Result()
			defer response.Body.Close()

			var out models.OrganizationRolesResponse
			assert.NilError(t, json.NewDecoder(response.Body).Decode(&out))

			assert.Check(t, is.Equal(tc.statusCode, recorder.Code))
			assert.Check(t, is.DeepEqual(tc.success, out.Success))

			if !tc.success {
				return
			}

			assert.Check(t, is.DeepEqual(tc.request.OrganizationID, out.OrganizationID))
			assert.Check(t, is.DeepEqual(tc.request.Role, out.Role))
		})
	}
}

func TestAccountRolesMeHandler(t *testing.T) {
	suite := suite.setupParallelTest(t).withTestUsers()

	suite.registerRouteOnce("GET", "account/roles/me", suite.h.AccountRolesMeHandler)

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	group, err := suite.db.Group.Create().
		SetName("Role Me Test Group " + suite.testUser1.ID).
		SetDescription("Group for account roles me tests").
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(ctx)
	assert.NilError(t, err)

	_, err = suite.db.GroupMembership.Create().
		SetGroupID(group.ID).
		SetUserID(suite.testUser1.ID).
		Save(ctx)
	assert.NilError(t, err)

	tuple := fgax.TupleKey{
		Subject: fgax.Entity{
			Kind:       fgax.Kind(generated.TypeUser),
			Identifier: suite.testUser1.ID,
		},
		Object: fgax.Entity{
			Kind:       fgax.Kind(generated.TypeOrganization),
			Identifier: suite.testUser1.OrganizationID,
		},
		Relation: fgax.Relation("policy_manager"),
	}

	_, err = suite.db.Authz.WriteTupleKeys(suite.testUser1.UserCtx, []fgax.TupleKey{tuple}, nil)
	assert.NilError(t, err)

	groupTuple := fgax.TupleKey{
		Subject: fgax.Entity{
			Kind:       fgax.Kind(generated.TypeGroup),
			Identifier: group.ID,
			Relation:   fgax.Relation(fgax.MemberRelation),
		},
		Object: fgax.Entity{
			Kind:       fgax.Kind(generated.TypeOrganization),
			Identifier: suite.testUser1.OrganizationID,
		},
		Relation: fgax.Relation("risk_manager"),
	}

	_, err = suite.db.Authz.WriteTupleKeys(suite.testUser1.UserCtx, []fgax.TupleKey{groupTuple}, nil)
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/account/roles/me", nil)
	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req.WithContext(suite.testUser1.UserCtx))

	res := recorder.Result()
	defer res.Body.Close()

	var out models.AccountRolesMeResponse
	assert.NilError(t, json.NewDecoder(res.Body).Decode(&out))

	assert.Check(t, is.Equal(http.StatusOK, recorder.Code))
	assert.Check(t, out.Success)
	assert.Check(t, is.DeepEqual(suite.testUser1.OrganizationID, out.OrganizationID))
	expectedRoles := []models.OrganizationRole{
		{
			ID:          "policy_manager",
			Name:        "Policy Manager",
			Description: "Manage all policies and procedures",
		},
		{
			ID:          "risk_manager",
			Name:        "Risk Manager",
			Description: "Manage risks, vulnerabilities, findings, and remediation",
		},
	}

	assert.Check(t, is.Len(out.Roles, len(expectedRoles)))

	for _, role := range expectedRoles {
		assert.Check(t, is.Contains(out.Roles, role))
	}
}
