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

	"github.com/theopenlane/common/integrations/types"
	credentialmodels "github.com/theopenlane/common/models"
	models "github.com/theopenlane/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/integration"
)

func (suite *HandlerTestSuite) TestDisconnectIntegrationSuccess() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegration"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:provider", op, suite.h.DisconnectIntegration)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	integrationID := suite.createTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, types.ProviderType("github"))

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/integrations/github?integration_id=%s", integrationID), nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp models.DeleteIntegrationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, integrationID, resp.DeletedID)
	assert.Contains(t, resp.Message, "github")

	count, err := suite.db.Integration.Query().
		Where(
			integration.OwnerID(testUser.OrganizationID),
			integration.Kind("github"),
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

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	req := httptest.NewRequest(http.MethodDelete, "/v1/integrations/github", nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func (suite *HandlerTestSuite) createTestIntegration(t *testing.T, ctx context.Context, orgID string, provider types.ProviderType) string {
	t.Helper()

	payload, err := types.NewCredentialBuilder(provider).
		With(
			types.WithCredentialKind(types.CredentialKindMetadata),
			types.WithCredentialSet(credentialmodels.CredentialSet{
				ProviderData: map[string]any{"token": "secret"},
			}),
		).Build()
	require.NoError(t, err)

	_, err = suite.h.IntegrationStore.SaveCredential(ctx, orgID, payload)
	require.NoError(t, err)

	record := suite.db.Integration.Query().
		Where(
			integration.OwnerID(orgID),
			integration.Kind(string(provider)),
		).
		OnlyX(ctx)

	return record.ID
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

	req := httptest.NewRequest(http.MethodDelete, "/v1/integrations/github", nil)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) TestDisconnectIntegrationWithExplicitID() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegrationWithID"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:provider", op, suite.h.DisconnectIntegration)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	integrationID := suite.createTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, types.ProviderType("github"))

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/integrations/github?integration_id=%s", integrationID), nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp models.DeleteIntegrationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, integrationID, resp.DeletedID)
}

func (suite *HandlerTestSuite) TestDisconnectIntegrationNotFoundWithID() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "DisconnectIntegrationNotFoundWithID"
	suite.registerRouteOnce(http.MethodDelete, "/v1/integrations/:provider", op, suite.h.DisconnectIntegration)

	ctx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(ctx, &userInput{confirmedUser: true})

	nonExistentID := "01HXYZ1234567890ABCDEFGHJ"

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/integrations/github?integration_id=%s", nonExistentID), nil)
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
