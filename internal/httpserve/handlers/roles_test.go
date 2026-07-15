package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func (suite *HandlerTestSuite) TestOrganizationRolesHandler() {
	t := suite.T()

	suite.registerRouteOnce("GET", "account/organization-roles", suite.h.RolesHandler)

	req := httptest.NewRequest(http.MethodGet, "/account/organization-roles", nil)
	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req.WithContext(testUser1.UserCtx))

	response := recorder.Result()
	defer response.Body.Close()

	var out models.RolesReply
	require.NoError(t, json.NewDecoder(response.Body).Decode(&out))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.True(t, out.Success)

	assert.Contains(t, out.Roles, models.OrganizationRole{
		ID:          "policy_manager",
		Name:        "Policy Manager",
		Description: "Manage all policies and procedures",
	})
	assert.Contains(t, out.Roles, models.OrganizationRole{
		ID:          "risk_manager",
		Name:        "Risk Manager",
		Description: "Manage risks, vulnerabilities, findings, and remediation",
	})
}

func (suite *HandlerTestSuite) TestOrganizationRolesAssignmentHandler() {
	t := suite.T()

	suite.registerRouteOnce("POST", "account/organization-roles", suite.h.AssignOrganizationRolesHandler)
	suite.registerRouteOnce("DELETE", "account/organization-roles", suite.h.DeleteOrganizationRolesHandler)

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ownerCtx := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.OrganizationID, auth.WithOrganizationRole(auth.OwnerRole))
	ownerCtx = privacy.DecisionContext(ownerCtx, privacy.Allow)
	ownerCtx = generated.NewContext(ownerCtx, suite.db)

	group, err := suite.db.Group.Create().
		SetName("Role Test Group " + testUser2.ID).
		SetDescription("Group for organization role assignment tests").
		SetOwnerID(testUser1.OrganizationID).
		Save(ctx)
	require.NoError(t, err)

	member := suite.userBuilder(ctx)
	memberRole := enums.RoleMember
	err = suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		UserID:         member.ID,
		OrganizationID: testUser1.OrganizationID,
		Role:           &memberRole,
	}).Exec(ctx)
	require.NoError(t, err)

	memberCtx := auth.NewTestContextWithOrgID(member.ID, testUser1.OrganizationID, auth.WithOrganizationRole(auth.MemberRole))
	memberCtx = privacy.DecisionContext(memberCtx, privacy.Allow)
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
				OrganizationID: testUser1.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusOK,
			success:    true,
		},
		{
			name:   "assign existing role to a user",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: testUser1.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusOK,
			success:    true,
		},
		{
			name:   "assign new role to a group",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: testUser1.OrganizationID,
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
				OrganizationID: testUser1.OrganizationID,
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
				OrganizationID: testUser1.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusOK,
			success:    true,
		},
		{
			name:   "invalid role",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: testUser1.OrganizationID,
				Role:           "not_a_role",
				UserIDs:        []string{testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "missing subjects to apply roles to",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: testUser1.OrganizationID,
				Role:           "policy_manager",
			},
			ctx:        ownerCtx,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "assign role to user from unauthorized organization",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: testUser2.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{testUser2.ID},
			},
			ctx:        ownerCtx,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "member cannot assign role to a user",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: testUser1.OrganizationID,
				Role:           "policy_manager",
				UserIDs:        []string{testUser2.ID},
			},
			ctx:        memberCtx,
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "member cannot assign role to a group",
			method: http.MethodPost,
			request: models.OrganizationRolesRequest{
				OrganizationID: testUser1.OrganizationID,
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
			require.NoError(t, err)

			req := httptest.NewRequest(tc.method, "/account/organization-roles", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req.WithContext(tc.ctx))

			response := recorder.Result()
			defer response.Body.Close()

			var out models.OrganizationRolesReply
			require.NoError(t, json.NewDecoder(response.Body).Decode(&out))

			assert.Equal(t, tc.statusCode, recorder.Code)
			assert.Equal(t, tc.success, out.Success)

			if !tc.success {
				return
			}

			assert.Equal(t, tc.request.OrganizationID, out.OrganizationID)
			assert.Equal(t, tc.request.Role, out.Role)
		})
	}
}

func (suite *HandlerTestSuite) TestAccountRolesMeHandler() {
	t := suite.T()

	suite.registerRouteOnce("GET", "account/roles/me", suite.h.AccountRolesMeHandler)

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	group, err := suite.db.Group.Create().
		SetName("Role Me Test Group " + testUser1.ID).
		SetDescription("Group for account roles me tests").
		SetOwnerID(testUser1.OrganizationID).
		Save(ctx)
	require.NoError(t, err)

	_, err = suite.db.GroupMembership.Create().
		SetGroupID(group.ID).
		SetUserID(testUser1.ID).
		Save(ctx)
	require.NoError(t, err)

	tuple := fgax.TupleKey{
		Subject: fgax.Entity{
			Kind:       fgax.Kind(generated.TypeUser),
			Identifier: testUser1.ID,
		},
		Object: fgax.Entity{
			Kind:       fgax.Kind(generated.TypeOrganization),
			Identifier: testUser1.OrganizationID,
		},
		Relation: fgax.Relation("policy_manager"),
	}

	_, err = suite.db.Authz.WriteTupleKeys(testUser1.UserCtx, []fgax.TupleKey{tuple}, nil)
	require.NoError(t, err)

	groupTuple := fgax.TupleKey{
		Subject: fgax.Entity{
			Kind:       fgax.Kind(generated.TypeGroup),
			Identifier: group.ID,
			Relation:   fgax.Relation(fgax.MemberRelation),
		},
		Object: fgax.Entity{
			Kind:       fgax.Kind(generated.TypeOrganization),
			Identifier: testUser1.OrganizationID,
		},
		Relation: fgax.Relation("risk_manager"),
	}

	_, err = suite.db.Authz.WriteTupleKeys(testUser1.UserCtx, []fgax.TupleKey{groupTuple}, nil)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/account/roles/me", nil)
	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req.WithContext(testUser1.UserCtx))

	res := recorder.Result()
	defer res.Body.Close()

	var out models.AccountRolesMeReply
	require.NoError(t, json.NewDecoder(res.Body).Decode(&out))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.True(t, out.Success)
	assert.Equal(t, testUser1.OrganizationID, out.OrganizationID)
	assert.ElementsMatch(t, []models.OrganizationRole{
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
	}, out.Roles)
}
