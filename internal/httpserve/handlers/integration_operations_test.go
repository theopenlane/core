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

const (
	operationTestDefinitionID       = "def_01K0TESTOPS00000000000001"
	operationTestInlineDefinitionID = "def_01K0TESTOPS00000000000002"
	operationTestPath               = "/v1/integrations/:definitionID/operations"
)

// OperationTestHealthCheck is the config type for the test health check operation
type OperationTestHealthCheck struct{}

// OperationTestRepoSync is the config type for the test repository sync operation
type OperationTestRepoSync struct{}

// OperationTestValidated is the config type for the test validated operation
type OperationTestValidated struct {
	// Target is the required target field
	Target string `json:"target" jsonschema:"required"`
}

// operationTestCredential is the credential type stored by the operation test definition
type operationTestCredential struct {
	// Token is the stored token
	Token string `json:"token"`
}

var (
	operationTestCredentialRef = types.CredentialRefOf[operationTestCredential]()
	opTestHealthCheckOperation = types.OperationRefOf[OperationTestHealthCheck]()
	opTestRepoSyncOperation    = types.OperationRefOf[OperationTestRepoSync]()
	opTestValidatedOperation   = types.OperationRefOf[OperationTestValidated]()
)

func operationTestDefinitionBuilder(definitionID string, inlineNonHealth bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		definition := types.NewDefinitionRef(definitionID)

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Operation Test",
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				operationTestCredentialRef.Registration(types.CredentialRegistration{
					Name: "Op Test Credential",
				}),
			},
			HealthCheck: types.CredentialHealthCheck(func(context.Context, types.OperationRequest) (json.RawMessage, error) {
				return json.RawMessage(`{"ok":true}`), nil
			}),
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  operationTestCredentialRef.ID(),
					Name:           "Op Test Connection",
					CredentialRefs: []types.CredentialSlotID{operationTestCredentialRef.ID()},
				},
			},
			Operations: []types.OperationRegistration{
				opTestHealthCheckOperation.Registration(definition, types.OperationRegistration{
					Description: "Validate the test credential",
					Policy:      types.ExecutionPolicy{Inline: true},
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"ok":true}`), nil
					},
				}),
				opTestRepoSyncOperation.Registration(definition, types.OperationRegistration{
					Description: "Sync repositories",
					Policy:      types.ExecutionPolicy{Inline: inlineNonHealth},
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"synced":true}`), nil
					},
				}),
				opTestValidatedOperation.Registration(definition, types.OperationRegistration{
					Description: "Operation with config schema",
					Policy:      types.ExecutionPolicy{Inline: true},
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"validated":true}`), nil
					},
				}),
			},
		}, nil
	})
}

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

	def, ok := suite.h.IntegrationsRuntime.Definition(definitionID)
	require.True(t, ok)

	rec, _, err := suite.h.IntegrationsRuntime.EnsureInstallation(ctx, orgID, "", def)
	require.NoError(t, err)

	credential := types.CredentialSet{
		Data: json.RawMessage(`{"token":"test-token"}`),
	}

	err = suite.h.IntegrationsRuntime.Reconcile(ctx, rec, nil, operationTestCredentialRef.ID(), &credential, nil)
	require.NoError(t, err)

	return rec.ID
}
