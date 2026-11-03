package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	openmodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

const disconnectTestDefinitionID = "def_01K0TESTDISC0000000000001"

func TestDisconnectIntegration(t *testing.T) {
	suite.setupTest(t)

	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:integrationID", suite.h.DisconnectIntegration)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{githubTestDefinitionBuilder(disconnectTestDefinitionID)})
	t.Cleanup(restore)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	owner := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})
	memberOrgOwner := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})
	member := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	err := suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: memberOrgOwner.OrganizationID,
		UserID:         member.ID,
		Role:           &enums.RoleMember,
	}).Exec(memberOrgOwner.UserCtx)
	assert.NilError(t, err)

	ownerIntegrationID := suite.createTestIntegration(t, owner.UserCtx, owner.OrganizationID, disconnectTestDefinitionID)
	memberOrgIntegrationID := suite.createTestIntegration(t, memberOrgOwner.UserCtx, memberOrgOwner.OrganizationID, disconnectTestDefinitionID)

	testCases := []struct {
		name               string
		ctx                context.Context
		integrationID      string
		expectedStatus     int
		expectDisconnected bool
		expectRetained     bool
	}{
		{
			name:               "happy path, owner disconnects",
			ctx:                owner.UserCtx,
			integrationID:      ownerIntegrationID,
			expectedStatus:     http.StatusOK,
			expectDisconnected: true,
		},
		{
			name:           "member cannot disconnect",
			ctx:            auth.NewTestContextWithOrgID(member.ID, memberOrgOwner.OrganizationID),
			integrationID:  memberOrgIntegrationID,
			expectRetained: true,
		},
		{
			name:           "integration not found",
			ctx:            owner.UserCtx,
			integrationID:  ulids.New().String(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthenticated request",
			ctx:            context.Background(),
			integrationID:  disconnectTestDefinitionID,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/v1/integrations/"+tc.integrationID, nil)

			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req.WithContext(tc.ctx))

			if tc.expectedStatus != 0 {
				assert.Check(t, is.Equal(tc.expectedStatus, rec.Code))
			} else {
				assert.Check(t, http.StatusOK != rec.Code)
			}

			if tc.expectDisconnected {
				var resp openmodels.DeleteIntegrationResponse
				assert.Check(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Check(t, resp.Success)
				assert.Check(t, is.DeepEqual(tc.integrationID, resp.DeletedID))
				assert.Check(t, is.Contains(resp.Message, "GitHub"))

				count, err := suite.db.Integration.Query().
					Where(
						integration.OwnerIDEQ(owner.OrganizationID),
						integration.DefinitionIDEQ(disconnectTestDefinitionID),
					).
					Count(owner.UserCtx)
				assert.Check(t, err)
				assert.Check(t, is.Equal(0, count))
			}

			if tc.expectRetained {
				installation, err := suite.db.Integration.Get(memberOrgOwner.UserCtx, tc.integrationID)
				assert.NilError(t, err)

				credential, ok, err := suite.keystore.LoadCredential(memberOrgOwner.UserCtx, installation, githubTestCredentialRef)
				assert.NilError(t, err)
				assert.Assert(t, ok)
				assert.Check(t, is.Contains(string(credential.Data), "secret"))
			}
		})
	}
}

// createTestIntegration creates an integration record in the DB and saves a test credential
func (suite *HandlerTestSuite) createTestIntegration(t *testing.T, ctx context.Context, orgID, definitionID string) string {
	t.Helper()

	def, ok := suite.h.IntegrationsRuntime.Registry().Definition(definitionID)
	assert.Check(t, ok)

	rec, _, err := suite.h.IntegrationsRuntime.EnsureInstallation(ctx, orgID, "", def, nil, nil)
	assert.Check(t, err)

	credential := types.CredentialSet{
		Data: json.RawMessage(`{"token":"secret"}`),
	}

	err = suite.h.IntegrationsRuntime.ReconcileCredential(ctx, rec, githubTestCredentialRef, credential)
	assert.Check(t, err)

	return rec.ID
}
