//go:build test

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func (suite *HandlerTestSuite) TestRunIntegrationOperationHealthCheckInline() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, operationTestDefinitionID)

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: integrationID,
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.RunIntegrationOperationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, opTestHealthCheckOperation.Name(), resp.Operation)
	assert.Contains(t, resp.Summary, "Integration operation completed")
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationInlinePolicy() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestInlineDefinitionID, true)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, operationTestInlineDefinitionID)

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: integrationID,
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestRepoSyncOperation.Name(),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestInlineDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.RunIntegrationOperationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, opTestRepoSyncOperation.Name(), resp.Operation)
	assert.Contains(t, resp.Summary, "Integration operation completed")
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationQueuedAsync() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, operationTestDefinitionID)

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: integrationID,
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestRepoSyncOperation.Name(),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.RunIntegrationOperationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "queued", resp.Status)
	assert.Equal(t, opTestRepoSyncOperation.Name(), resp.Operation)
	assert.Contains(t, resp.Summary, "queued")
	assert.NotEmpty(t, resp.Details)
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationUnauthorized() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationInvalidProvider() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/def_nonexistent/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationMissingOperationName() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		Body: handlers.RunIntegrationOperationBody{},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationUnknownOperation() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		Body: handlers.RunIntegrationOperationBody{
			Operation: "nonexistent.op",
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationInvalidConfig() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, operationTestDefinitionID)

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: integrationID,
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestValidatedOperation.Name(),
			Config:    json.RawMessage(`{"missing":"target_field"}`),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/%s/operations?integration_id=%s", operationTestDefinitionID, integrationID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationInstallationNotFound() {
	t := suite.T()

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) createOperationTestIntegration(t *testing.T, ctx context.Context, orgID, definitionID string) string {
	t.Helper()

	def, ok := suite.h.IntegrationsRuntime.Registry().Definition(definitionID)
	require.True(t, ok)

	rec, _, err := suite.h.IntegrationsRuntime.EnsureInstallation(ctx, orgID, "", def, nil, nil)
	require.NoError(t, err)

	credential := types.CredentialSet{
		Data: json.RawMessage(`{"token":"test-token"}`),
	}

	err = suite.h.IntegrationsRuntime.ReconcileCredential(ctx, rec, operationTestCredentialRef, credential)
	require.NoError(t, err)

	return rec.ID
}
