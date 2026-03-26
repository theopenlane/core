package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	openmodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

const disconnectTestDefinitionID = "def_01K0TESTDISC0000000000001"

func (suite *HandlerTestSuite) TestDisconnectIntegrationSuccess() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegration"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:integrationID", op, suite.h.DisconnectIntegration)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{githubTestDefinitionBuilder(disconnectTestDefinitionID)})
	defer restore()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	integrationID := suite.createTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, disconnectTestDefinitionID)

	req := httptest.NewRequest(http.MethodDelete, "/v1/integrations/"+integrationID, nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp openmodels.DeleteIntegrationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, integrationID, resp.DeletedID)
	assert.Contains(t, resp.Message, "GitHub")

	count, err := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(testUser.OrganizationID),
			integration.DefinitionIDEQ(disconnectTestDefinitionID),
		).
		Count(testUser.UserCtx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func (suite *HandlerTestSuite) TestDisconnectIntegrationNotFound() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegrationNotFound"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:integrationID", op, suite.h.DisconnectIntegration)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	req := httptest.NewRequest(http.MethodDelete, "/v1/integrations/01HXYZ1234567890ABCDEFGHJ", nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

// createTestIntegration creates an integration record in the DB and saves a test credential
func (suite *HandlerTestSuite) createTestIntegration(t *testing.T, ctx context.Context, orgID, definitionID string) string {
	t.Helper()

	rec, err := suite.db.Integration.Create().
		SetOwnerID(orgID).
		SetName(definitionID).
		SetDefinitionID(definitionID).
		Save(ctx)
	require.NoError(t, err)

	credential := types.CredentialSet{
		Data: json.RawMessage(`{"token":"secret"}`),
	}

	err = suite.h.IntegrationsRuntime.Reconcile(ctx, rec, nil, githubTestCredentialRef, &credential, nil)
	require.NoError(t, err)

	return rec.ID
}

func (suite *HandlerTestSuite) TestDisconnectIntegrationUnauthorized() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegrationUnauthorized"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:integrationID", op, suite.h.DisconnectIntegration)

	req := httptest.NewRequest(http.MethodDelete, "/v1/integrations/"+disconnectTestDefinitionID, nil)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
