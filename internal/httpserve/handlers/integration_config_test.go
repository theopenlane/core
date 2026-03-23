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

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
)

const configTestProviderID = "def_test_gcpscc"

var configTestCredentialRef = types.NewCredentialSlotID("config_test")

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderSuccess() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderSuccess"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
		UserInput: json.RawMessage(
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

	credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, stored, types.NewCredentialSlotID("config_test"))
	require.NoError(t, err)
	require.True(t, ok)
	assert.Contains(t, string(credential.Data), "projectId")
	assert.Equal(t, `payload.severity == "HIGH"`, decodeClientConfigField(t, stored.Config.ClientConfig, "filterExpr"))
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderAcceptsDefinitionID() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderByID"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
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
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRejectsNonObjectPayload() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderRejectsNonObjectPayload"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(`["not","an","object"]`),
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
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
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
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	first := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "initial-project", "serviceAccountEmail": "initial@example.iam.gserviceaccount.com"})),
	})

	second := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		DefinitionID:   configTestProviderID,
		CredentialRef:  configTestCredentialRef,
		InstallationID: first.InstallationID,
		Body:           json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "updated-project", "serviceAccountEmail": "updated@example.iam.gserviceaccount.com"})),
	})

	assert.Equal(t, first.InstallationID, second.InstallationID)

	stored := suite.db.Integration.GetX(testUser.UserCtx, first.InstallationID)
	credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, stored, types.NewCredentialSlotID("config_test"))
	require.NoError(t, err)
	require.True(t, ok)

	providerData, err := jsonx.ToMap(credential.Data)
	require.NoError(t, err)
	assert.Equal(t, "updated-project", providerData["projectId"])
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUpdateExistingUserInputOnly() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUpdateUserInputOnly"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	first := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "initial-project", "serviceAccountEmail": "initial@example.iam.gserviceaccount.com"})),
	})

	second := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		DefinitionID:   configTestProviderID,
		InstallationID: first.InstallationID,
		UserInput:      json.RawMessage(mustMarshalJSON(t, map[string]any{"filterExpr": "payload.category == \"critical\""})),
	})

	assert.Equal(t, first.InstallationID, second.InstallationID)

	stored := suite.db.Integration.GetX(testUser.UserCtx, first.InstallationID)
	assert.Equal(t, enums.IntegrationStatusConnected, stored.Status)
	assert.Equal(t, `payload.category == "critical"`, decodeClientConfigField(t, stored.Config.ClientConfig, "filterExpr"))

	credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, stored, types.NewCredentialSlotID("config_test"))
	require.NoError(t, err)
	require.True(t, ok)

	providerData, err := jsonx.ToMap(credential.Data)
	require.NoError(t, err)
	assert.Equal(t, "initial-project", providerData["projectId"])
	assert.Equal(t, "initial@example.iam.gserviceaccount.com", providerData["serviceAccountEmail"])
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderAllowsUserInputOnlyUpdateWithoutCredentialSchema() {
	t := suite.T()

	const definitionID = "def_test_user_input_only"

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUserInputOnly"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{userInputOnlyTestDefinitionBuilder(definitionID, "oidc-generic")})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := suite.db.Integration.Create().
		SetOwnerID(testUser.OrganizationID).
		SetName("OIDC Generic").
		SetDefinitionID(definitionID).
		SetDefinitionSlug("oidc-generic").
		SetStatus(enums.IntegrationStatusConnected).
		SaveX(testUser.UserCtx)

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:   definitionID,
		InstallationID: rec.ID,
		UserInput:      json.RawMessage(mustMarshalJSON(t, map[string]any{"filterExpr": "payload.actor == \"service-account\""})),
	})

	httpRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+definitionID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(httpRec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusOK, httpRec.Code)

	stored := suite.db.Integration.GetX(testUser.UserCtx, rec.ID)
	assert.Equal(t, `payload.actor == "service-account"`, decodeClientConfigField(t, stored.Config.ClientConfig, "filterExpr"))
	assert.Equal(t, enums.IntegrationStatusConnected, stored.Status)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRejectsInstallationDefinitionMismatch() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderMismatch"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{
		configTestDefinitionBuilder(configTestProviderID, "gcpscc", false),
		configTestDefinitionBuilder("def_test_other", "other", false),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	other := performIntegrationConfigRequest(t, suite, testUser.UserCtx, "def_test_other", handlers.IntegrationConfigPayload{
		DefinitionID:  "def_test_other",
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "other-project", "serviceAccountEmail": "other@example.iam.gserviceaccount.com"})),
	})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:   configTestProviderID,
		CredentialRef:  configTestCredentialRef,
		InstallationID: other.InstallationID,
		Body:           json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
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
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, "gcpscc", true)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	require.Equal(t, http.StatusBadRequest, rec.Code)

	// A PENDING installation row must be created even when the health check fails.
	// The credential must not be stored and the status must not advance to CONNECTED.
	records, err := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(testUser.OrganizationID),
			integration.DefinitionIDEQ(configTestProviderID),
		).
		All(testUser.UserCtx)
	require.NoError(t, err)
	require.Len(t, records, 1, "expected one PENDING installation row after failed setup")
	assert.Equal(t, enums.IntegrationStatusPending, records[0].Status)

	_, credOk, credErr := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, records[0], types.NewCredentialSlotID("config_test"))
	require.NoError(t, credErr)
	assert.False(t, credOk, "credential must not be stored after a failed health check")
}

func configTestDefinitionBuilder(definitionID, slug string, failHealth bool) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
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
				DisplayName: "Config Test",
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: json.RawMessage(`{"type":"object","properties":{"filterExpr":{"type":"string"}}}`),
			},
			CredentialRegistrations: []types.CredentialRegistration{
				{
					Ref:         configTestCredentialRef,
					Name:        "Config Test Credential",
					Description: "Credential slot used by the config test definition.",
					Schema:      json.RawMessage(`{"type":"object","required":["projectId","serviceAccountEmail"],"properties":{"projectId":{"type":"string"},"serviceAccountEmail":{"type":"string"}}}`),
				},
			},
			Connections: []types.ConnectionRegistration{
				{
					CredentialRef:       configTestCredentialRef,
					Name:                "Config Test Connection",
					Description:         "Connect the config test definition using the configured credential payload.",
					CredentialRefs:      []types.CredentialSlotID{configTestCredentialRef},
					ValidationOperation: "health.default",
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: configTestCredentialRef,
						Name:          "Disconnect Config Test Connection",
						Description:   "Remove the persisted config test credential and disconnect this installation.",
					},
				},
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

func userInputOnlyTestDefinitionBuilder(definitionID, slug string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
				Slug:        slug,
				DisplayName: "User Input Test",
				Active:      true,
				Visible:     true,
			},
			UserInput: &types.UserInputRegistration{
				Schema: json.RawMessage(`{"type":"object","properties":{"filterExpr":{"type":"string"}}}`),
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
