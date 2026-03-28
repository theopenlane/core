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

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	operationTestDefinitionID = "def_01K0TESTOPS00000000000001"
	operationTestPath         = "/v1/integrations/:definitionID/operations"
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

var (
	operationTestCredentialRef                      = types.NewCredentialSlotID("op_test")
	opTestHealthSchema, opTestHealthCheckOperation  = providerkit.OperationSchema[OperationTestHealthCheck]()
	opTestRepoSyncSchema, opTestRepoSyncOperation   = providerkit.OperationSchema[OperationTestRepoSync]()
	opTestValidatedSchema, opTestValidatedOperation = providerkit.OperationSchema[OperationTestValidated]()
)

func operationTestDefinitionBuilder(definitionID string, inlineNonHealth bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				DisplayName: "Operation Test",
				Active:      true,
				Visible:     true,
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:    operationTestCredentialRef,
					Name:   "Op Test Credential",
					Schema: json.RawMessage(`{"type":"object","properties":{"token":{"type":"string"}}}`),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:  operationTestCredentialRef,
					Name:           "Op Test Connection",
					CredentialRefs: []types.CredentialSlotID{operationTestCredentialRef},
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         opTestHealthCheckOperation.Name(),
					Description:  "Validate the test credential",
					Topic:        types.OperationTopic(definitionID, opTestHealthCheckOperation.Name()),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: opTestHealthSchema,
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"ok":true}`), nil
					},
				},
				{
					Name:         opTestRepoSyncOperation.Name(),
					Description:  "Sync repositories",
					Topic:        types.OperationTopic(definitionID, opTestRepoSyncOperation.Name()),
					Policy:       types.ExecutionPolicy{Inline: inlineNonHealth},
					ConfigSchema: opTestRepoSyncSchema,
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"synced":true}`), nil
					},
				},
				{
					Name:         opTestValidatedOperation.Name(),
					Description:  "Operation with config schema",
					Topic:        types.OperationTopic(definitionID, opTestValidatedOperation.Name()),
					ConfigSchema: opTestValidatedSchema,
					Policy:       types.ExecutionPolicy{Inline: true},
					Handle: func(context.Context, types.OperationRequest) (json.RawMessage, error) {
						return json.RawMessage(`{"validated":true}`), nil
					},
				},
			},
		}, nil
	})
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationHealthCheckInline() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "RunIntegrationOperationHealthCheckInline"
	suite.registerRouteOnce(http.MethodPost, operationTestPath, op, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, operationTestDefinitionID)

	body, err := json.Marshal(handlers.IntegrationOperationPayload{
		IntegrationID: integrationID,
		Body: handlers.IntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationOperationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, opTestHealthCheckOperation.Name(), resp.Operation)
	assert.Contains(t, resp.Summary, "Integration operation completed")
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationInlinePolicy() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "RunIntegrationOperationInlinePolicy"
	suite.registerRouteOnce(http.MethodPost, operationTestPath, op, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, true)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, operationTestDefinitionID)

	body, err := json.Marshal(handlers.IntegrationOperationPayload{
		IntegrationID: integrationID,
		Body: handlers.IntegrationOperationBody{
			Operation: opTestRepoSyncOperation.Name(),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationOperationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, opTestRepoSyncOperation.Name(), resp.Operation)
	assert.Contains(t, resp.Summary, "Integration operation completed")
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationQueuedAsync() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "RunIntegrationOperationQueuedAsync"
	suite.registerRouteOnce(http.MethodPost, operationTestPath, op, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, operationTestDefinitionID)

	body, err := json.Marshal(handlers.IntegrationOperationPayload{
		IntegrationID: integrationID,
		Body: handlers.IntegrationOperationBody{
			Operation: opTestRepoSyncOperation.Name(),
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationOperationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "queued", resp.Status)
	assert.Equal(t, opTestRepoSyncOperation.Name(), resp.Operation)
	assert.Contains(t, resp.Summary, "queued")
	assert.NotEmpty(t, resp.Details)
}

func (suite *HandlerTestSuite) TestRunIntegrationOperationUnauthorized() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "RunIntegrationOperationUnauthorized"
	suite.registerRouteOnce(http.MethodPost, operationTestPath, op, suite.h.RunIntegrationOperation)

	body, err := json.Marshal(handlers.IntegrationOperationPayload{
		Body: handlers.IntegrationOperationBody{
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

	op := openapi3.NewOperation()
	op.OperationID = "RunIntegrationOperationInvalidProvider"
	suite.registerRouteOnce(http.MethodPost, operationTestPath, op, suite.h.RunIntegrationOperation)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationOperationPayload{
		Body: handlers.IntegrationOperationBody{
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

	op := openapi3.NewOperation()
	op.OperationID = "RunIntegrationOperationMissingName"
	suite.registerRouteOnce(http.MethodPost, operationTestPath, op, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationOperationPayload{
		Body: handlers.IntegrationOperationBody{},
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

	op := openapi3.NewOperation()
	op.OperationID = "RunIntegrationOperationUnknown"
	suite.registerRouteOnce(http.MethodPost, operationTestPath, op, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationOperationPayload{
		Body: handlers.IntegrationOperationBody{
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

	op := openapi3.NewOperation()
	op.OperationID = "RunIntegrationOperationInvalidConfig"
	suite.registerRouteOnce(http.MethodPost, operationTestPath, op, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, testUser.UserCtx, testUser.OrganizationID, operationTestDefinitionID)

	body, err := json.Marshal(handlers.IntegrationOperationPayload{
		IntegrationID: integrationID,
		Body: handlers.IntegrationOperationBody{
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

	op := openapi3.NewOperation()
	op.OperationID = "RunIntegrationOperationInstallNotFound"
	suite.registerRouteOnce(http.MethodPost, operationTestPath, op, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationOperationPayload{
		Body: handlers.IntegrationOperationBody{
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

	rec, err := suite.db.Integration.Create().
		SetOwnerID(orgID).
		SetName(definitionID).
		SetDefinitionID(definitionID).
		Save(ctx)
	require.NoError(t, err)

	credential := types.CredentialSet{
		Data: json.RawMessage(`{"token":"test-token"}`),
	}

	err = suite.h.IntegrationsRuntime.Reconcile(ctx, rec, nil, operationTestCredentialRef, &credential, nil)
	require.NoError(t, err)

	return rec.ID
}
