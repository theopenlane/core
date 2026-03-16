package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/common/models"
	openmodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/integrations/definition"
)

const disconnectTestDefinitionID = "def_test_github"

func (suite *HandlerTestSuite) TestDisconnectIntegrationSuccess() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegration"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:provider", op, suite.h.DisconnectIntegration)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{githubTestDefinitionBuilder(disconnectTestDefinitionID)})
	defer restore()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	integrationID := suite.createTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, disconnectTestDefinitionID)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/integrations/%s?integration_id=%s", disconnectTestDefinitionID, integrationID), nil)
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
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:provider", op, suite.h.DisconnectIntegration)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{githubTestDefinitionBuilder(disconnectTestDefinitionID)})
	defer restore()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	req := httptest.NewRequest(http.MethodDelete, "/v1/integrations/"+disconnectTestDefinitionID, nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

// createTestIntegration creates an integration record in the DB and saves a test credential.
func (suite *HandlerTestSuite) createTestIntegration(t *testing.T, ctx context.Context, orgID, definitionID string) string {
	t.Helper()

	rec, err := suite.db.Integration.Create().
		SetOwnerID(orgID).
		SetName(definitionID).
		SetDefinitionID(definitionID).
		Save(ctx)
	require.NoError(t, err)

	credential := models.CredentialSet{
		ProviderData: json.RawMessage(`{"token":"secret"}`),
	}

	require.NoError(t, suite.h.IntegrationsRuntime.SaveCredential(ctx, rec, credential))

	return rec.ID
}

func (suite *HandlerTestSuite) TestDisconnectIntegrationInvalidProvider() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegrationInvalidProvider"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:provider", op, suite.h.DisconnectIntegration)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	req := httptest.NewRequest(http.MethodDelete, "/v1/integrations/invalid_provider", nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.True(t, rec.Code == http.StatusBadRequest || rec.Code == http.StatusNotFound)
}

func (suite *HandlerTestSuite) TestDisconnectIntegrationUnauthorized() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegrationUnauthorized"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:provider", op, suite.h.DisconnectIntegration)

	req := httptest.NewRequest(http.MethodDelete, "/v1/integrations/"+disconnectTestDefinitionID, nil)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) TestDisconnectIntegrationWithExplicitID() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegrationWithID"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:provider", op, suite.h.DisconnectIntegration)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{githubTestDefinitionBuilder(disconnectTestDefinitionID)})
	defer restore()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	integrationID := suite.createTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, disconnectTestDefinitionID)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/integrations/%s?integration_id=%s", disconnectTestDefinitionID, integrationID), nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp openmodels.DeleteIntegrationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, integrationID, resp.DeletedID)
}

func (suite *HandlerTestSuite) TestDisconnectIntegrationNotFoundWithID() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegrationNotFoundWithID"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:provider", op, suite.h.DisconnectIntegration)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{githubTestDefinitionBuilder(disconnectTestDefinitionID)})
	defer restore()

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	nonExistentID := "01HXYZ1234567890ABCDEFGHJ"

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/integrations/%s?integration_id=%s", disconnectTestDefinitionID, nonExistentID), nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
