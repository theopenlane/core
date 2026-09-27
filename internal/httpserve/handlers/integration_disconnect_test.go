package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	openmodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const disconnectTestDefinitionID = "def_01K0TESTDISC0000000000001"

func (suite *HandlerTestSuite) TestDisconnectIntegration() {
	t := suite.T()

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
	require.NoError(t, err)

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
				assert.Equal(t, tc.expectedStatus, rec.Code)
			} else {
				assert.NotEqual(t, http.StatusOK, rec.Code)
			}

			if tc.expectDisconnected {
				var resp openmodels.DeleteIntegrationResponse
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.True(t, resp.Success)
				assert.Equal(t, tc.integrationID, resp.DeletedID)
				assert.Contains(t, resp.Message, "GitHub")

				count, err := suite.db.Integration.Query().
					Where(
						integration.OwnerIDEQ(owner.OrganizationID),
						integration.DefinitionIDEQ(disconnectTestDefinitionID),
					).
					Count(owner.UserCtx)
				assert.NoError(t, err)
				assert.Zero(t, count)
			}

			if tc.expectRetained {
				installation, err := suite.db.Integration.Get(memberOrgOwner.UserCtx, tc.integrationID)
				require.NoError(t, err)

				credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(memberOrgOwner.UserCtx, installation, githubTestCredentialRef)
				require.NoError(t, err)
				require.True(t, ok)
				assert.Contains(t, string(credential.Data), "secret")
			}
		})
	}
}

// createTestIntegration creates an integration record in the DB and saves a test credential
func (suite *HandlerTestSuite) createTestIntegration(t *testing.T, ctx context.Context, orgID, definitionID string) string {
	t.Helper()

	def, ok := suite.h.IntegrationsRuntime.Definition(definitionID)
	assert.True(t, ok)

	rec, _, err := suite.h.IntegrationsRuntime.EnsureInstallation(ctx, orgID, "", def)
	assert.NoError(t, err)

	credential := types.CredentialSet{
		Data: json.RawMessage(`{"token":"secret"}`),
	}

	err = suite.h.IntegrationsRuntime.Reconcile(ctx, rec, nil, githubTestCredentialRef.ID(), &credential, nil)
	assert.NoError(t, err)

	return rec.ID
}
