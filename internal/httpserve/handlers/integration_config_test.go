//go:build test

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/integrations/definition"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

const configTestProviderID = "def_test_gcpscc"

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderSuccess() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderSuccess"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		Provider: configTestProviderID,
		Body:     handlers.IntegrationConfigBody(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
		UserInput: handlers.IntegrationConfigBody(
			mustMarshalJSON(t, map[string]any{"filterExpr": "payload.severity == \"HIGH\""}),
		),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationConfigResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "gcpscc", resp.Provider)

	stored := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(testUser.OrganizationID),
			integration.DefinitionIDEQ(configTestProviderID),
		).
		OnlyX(testUser.UserCtx)

	credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, stored)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Contains(t, string(credential.ProviderData), "projectId")
	assert.Equal(t, `payload.severity == "HIGH"`, decodeClientConfigField(t, stored.Config.ClientConfig, "filterExpr"))
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderAcceptsDefinitionID() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderByID"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		Provider: configTestProviderID,
		Body:     handlers.IntegrationConfigBody(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusOK, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderInvalidPayload() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderInvalid"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		Provider: configTestProviderID,
		Body:     handlers.IntegrationConfigBody(mustMarshalJSON(t, map[string]any{"serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUnauthorized() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUnauthorized"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		Provider: configTestProviderID,
		Body:     handlers.IntegrationConfigBody(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUpdateExisting() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUpdate"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	first := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		Provider: configTestProviderID,
		Body:     handlers.IntegrationConfigBody(mustMarshalJSON(t, map[string]any{"projectId": "initial-project", "serviceAccountEmail": "initial@example.iam.gserviceaccount.com"})),
	})

	second := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		Provider:       configTestProviderID,
		InstallationID: first.InstallationID,
		Body:           handlers.IntegrationConfigBody(mustMarshalJSON(t, map[string]any{"projectId": "updated-project", "serviceAccountEmail": "updated@example.iam.gserviceaccount.com"})),
	})

	assert.Equal(t, first.InstallationID, second.InstallationID)

	stored := suite.db.Integration.GetX(testUser.UserCtx, first.InstallationID)
	credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, stored)
	require.NoError(t, err)
	require.True(t, ok)

	providerData, err := jsonx.ToMap(credential.ProviderData)
	require.NoError(t, err)
	assert.Equal(t, "updated-project", providerData["projectId"])
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRejectsInstallationDefinitionMismatch() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderMismatch"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{
		configTestDefinitionBuilder(configTestProviderID, "gcpscc", false),
		configTestDefinitionBuilder("def_test_other", "other", false),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	other := performIntegrationConfigRequest(t, suite, testUser.UserCtx, "def_test_other", handlers.IntegrationConfigPayload{
		Provider: "def_test_other",
		Body:     handlers.IntegrationConfigBody(mustMarshalJSON(t, map[string]any{"projectId": "other-project", "serviceAccountEmail": "other@example.iam.gserviceaccount.com"})),
	})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		Provider:       configTestProviderID,
		InstallationID: other.InstallationID,
		Body:           handlers.IntegrationConfigBody(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderHealthFailureDoesNotPersistCredential() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderHealthFailure"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []definition.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", true)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		Provider: configTestProviderID,
		Body:     handlers.IntegrationConfigBody(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)

	count, err := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(testUser.OrganizationID),
			integration.DefinitionIDEQ(configTestProviderID),
		).
		Count(testUser.UserCtx)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func configTestDefinitionBuilder(definitionID, slug string, failHealth bool) definition.Builder {
	return definition.Builder(func() (types.Definition, error) {
		healthHandler := func(context.Context, types.OperationRequest) (json.RawMessage, error) {
			if failHealth {
				return nil, errors.New("health failed")
			}

			return json.RawMessage(`{"ok":true}`), nil
		}

		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				Slug:        slug,
				Version:     "v1",
				DisplayName: "Config Test",
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: json.RawMessage(`{"type":"object","properties":{"filterExpr":{"type":"string"}}}`),
			},
			Credentials: &types.CredentialRegistration{
				Schema: json.RawMessage(`{"type":"object","required":["projectId","serviceAccountEmail"],"properties":{"projectId":{"type":"string"},"serviceAccountEmail":{"type":"string"}}}`),
			},
			Operations: []types.OperationRegistration{
				{
					Name:        "health.default",
					Description: "Validate the config test installation",
					Topic:       gala.TopicName("integration." + slug + ".health.default"),
					Handle:      healthHandler,
				},
			},
		}, nil
	})
}

func performIntegrationConfigRequest(t *testing.T, suite *HandlerTestSuite, ctx context.Context, provider string, payload handlers.IntegrationConfigPayload) handlers.IntegrationConfigResponse {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+provider+"/config", bytes.NewReader(mustMarshalConfigPayload(t, payload)))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(ctx))
	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationConfigResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	return resp
}

func mustMarshalConfigPayload(t *testing.T, payload handlers.IntegrationConfigPayload) []byte {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	return body
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()

	body, err := json.Marshal(value)
	require.NoError(t, err)

	return body
}

func decodeClientConfigField(t *testing.T, raw json.RawMessage, key string) string {
	t.Helper()

	document, err := jsonx.ToMap(raw)
	require.NoError(t, err)

	value, _ := document[key].(string)

	return value
}
