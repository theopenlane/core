//go:build test

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/integrationwebhook"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	definitionscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	configTestProviderID           = "def_01K0TESTCFG00000000000001"
	configTestFailHealthProviderID = "def_01K0TESTCFG00000000000002"
)

// ConfigTestHealthCheck is the config type for the config test health check operation
type ConfigTestHealthCheck struct{}

var (
	configTestCredentialRef                        = types.NewCredentialSlotID("config_test")
	configHealthSchema, configHealthCheckOperation = providerkit.OperationSchema[ConfigTestHealthCheck]()
)

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderSuccess() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderSuccess"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
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

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationConfigResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, configTestProviderID, resp.Provider)

	stored := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(testUser.OrganizationID),
			integration.DefinitionIDEQ(configTestProviderID),
		).
		OnlyX(testUser.UserCtx)

	credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, stored, types.NewCredentialSlotID("config_test"))
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Contains(t, string(credential.Data), "projectId")
	assert.Equal(t, `payload.severity == "HIGH"`, decodeClientConfigField(t, stored.Config.ClientConfig, "filterExpr"))
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderReturnsSCIMEndpointDetails() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderSCIM"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{definitionscim.Builder()})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID: definitionscim.DefinitionID.ID(),
		UserInput:    json.RawMessage(mustMarshalJSON(t, map[string]any{"name": "Okta Production"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+definitionscim.DefinitionID.ID()+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationConfigResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	assert.True(t, resp.Success)
	assert.Equal(t, definitionscim.DefinitionID.ID(), resp.Provider)
	assert.NotEmpty(t, resp.IntegrationID)
	assert.NotEmpty(t, resp.WebhookSecret)
	assert.True(t, strings.HasPrefix(resp.WebhookEndpointURL, "http://example.com/v1/integrations/scim/"))
	assert.True(t, strings.HasSuffix(resp.WebhookEndpointURL, "/v2"))

	stored := suite.db.Integration.GetX(testUser.UserCtx, resp.IntegrationID)
	assert.Equal(t, enums.IntegrationStatusConnected, stored.Status)
	assert.Equal(t, "Okta Production", stored.Name)

	webhook := suite.db.IntegrationWebhook.Query().
		Where(
			integrationwebhook.IntegrationIDEQ(stored.ID),
			integrationwebhook.NameEQ(definitionscim.SCIMAuthWebhook.Name()),
			integrationwebhook.ExternalEventIDIsNil(),
		).
		OnlyX(testUser.UserCtx)

	assert.NotNil(t, webhook.EndpointID)
	assert.NotNil(t, webhook.EndpointURL)
	assert.Equal(t, "/v1/integrations/scim/"+*webhook.EndpointID+"/v2", *webhook.EndpointURL)
	assert.Equal(t, "http://example.com"+*webhook.EndpointURL, resp.WebhookEndpointURL)
	assert.Equal(t, webhook.SecretToken, resp.WebhookSecret)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderAcceptsDefinitionID() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderByID"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
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

	assert.Equal(t, http.StatusOK, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderInvalidPayload() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderInvalid"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
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

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderRejectsNonObjectPayload() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderRejectsNonObjectPayload"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
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

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUnauthorized() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUnauthorized"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
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

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUpdateExisting() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUpdate"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	first := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "initial-project", "serviceAccountEmail": "initial@example.iam.gserviceaccount.com"})),
	})

	second := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		IntegrationID: first.IntegrationID,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "updated-project", "serviceAccountEmail": "updated@example.iam.gserviceaccount.com"})),
	})

	assert.Equal(t, first.IntegrationID, second.IntegrationID)

	stored := suite.db.Integration.GetX(testUser.UserCtx, first.IntegrationID)
	credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, stored, types.NewCredentialSlotID("config_test"))
	assert.NoError(t, err)
	assert.True(t, ok)

	providerData, err := jsonx.ToMap(credential.Data)
	assert.NoError(t, err)
	assert.Equal(t, "updated-project", providerData["projectId"])
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUpdateExistingUserInputOnly() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUpdateUserInputOnly"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestProviderID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	first := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "initial-project", "serviceAccountEmail": "initial@example.iam.gserviceaccount.com"})),
	})

	second := performIntegrationConfigRequest(t, suite, testUser.UserCtx, configTestProviderID, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		IntegrationID: first.IntegrationID,
		UserInput:     json.RawMessage(mustMarshalJSON(t, map[string]any{"filterExpr": "payload.category == \"critical\""})),
	})

	assert.Equal(t, first.IntegrationID, second.IntegrationID)

	stored := suite.db.Integration.GetX(testUser.UserCtx, first.IntegrationID)
	assert.Equal(t, enums.IntegrationStatusConnected, stored.Status)
	assert.Equal(t, `payload.category == "critical"`, decodeClientConfigField(t, stored.Config.ClientConfig, "filterExpr"))

	credential, ok, err := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, stored, types.NewCredentialSlotID("config_test"))
	assert.NoError(t, err)
	assert.True(t, ok)

	providerData, err := jsonx.ToMap(credential.Data)
	assert.NoError(t, err)
	assert.Equal(t, "initial-project", providerData["projectId"])
	assert.Equal(t, "initial@example.iam.gserviceaccount.com", providerData["serviceAccountEmail"])
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderAllowsUserInputOnlyUpdateWithoutCredentialSchema() {
	t := suite.T()

	const definitionID = "def_01K0TESTUIONLY000000000001"

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUserInputOnly"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{userInputOnlyTestDefinitionBuilder(definitionID)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	rec := suite.db.Integration.Create().
		SetOwnerID(testUser.OrganizationID).
		SetName("OIDC Generic").
		SetDefinitionID(definitionID).
		SetStatus(enums.IntegrationStatusConnected).
		SaveX(testUser.UserCtx)

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:  definitionID,
		IntegrationID: rec.ID,
		UserInput:     json.RawMessage(mustMarshalJSON(t, map[string]any{"filterExpr": "payload.actor == \"service-account\""})),
	})

	httpRec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+definitionID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(httpRec, req.WithContext(testUser.UserCtx))

	assert.Equal(t, http.StatusOK, httpRec.Code)

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
		configTestDefinitionBuilder(configTestProviderID, false),
		configTestDefinitionBuilder("def_01K0TESTOTH00000000000001", false),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	other := performIntegrationConfigRequest(t, suite, testUser.UserCtx, "def_01K0TESTOTH00000000000001", handlers.IntegrationConfigPayload{
		DefinitionID:  "def_01K0TESTOTH00000000000001",
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "other-project", "serviceAccountEmail": "other@example.iam.gserviceaccount.com"})),
	})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestProviderID,
		CredentialRef: configTestCredentialRef,
		IntegrationID: other.IntegrationID,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderHealthFailureDoesNotPersistCredential() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderHealthFailure"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:definitionID/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{configTestDefinitionBuilder(configTestFailHealthProviderID, true)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body := mustMarshalConfigPayload(t, handlers.IntegrationConfigPayload{
		DefinitionID:  configTestFailHealthProviderID,
		CredentialRef: configTestCredentialRef,
		Body:          json.RawMessage(mustMarshalJSON(t, map[string]any{"projectId": "sample-project", "serviceAccountEmail": "svc@example.iam.gserviceaccount.com"})),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+configTestFailHealthProviderID+"/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// A PENDING installation row must be created even when the health check fails.
	// The credential must not be stored and the status must not advance to CONNECTED.
	records, err := suite.db.Integration.Query().
		Where(
			integration.OwnerIDEQ(testUser.OrganizationID),
			integration.DefinitionIDEQ(configTestFailHealthProviderID),
		).
		All(testUser.UserCtx)
	assert.NoError(t, err)
	assert.Len(t, records, 1, "expected one PENDING installation row after failed setup")
	assert.Equal(t, enums.IntegrationStatusPending, records[0].Status)

	_, credOk, credErr := suite.h.IntegrationsRuntime.LoadCredential(testUser.UserCtx, records[0], types.NewCredentialSlotID("config_test"))
	assert.NoError(t, credErr)
	assert.False(t, credOk, "credential must not be stored after a failed health check")
}

func configTestDefinitionBuilder(definitionID string, failHealth bool) registry.Builder {
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
					ValidationOperation: configHealthCheckOperation.Name(),
					Disconnect: &types.DisconnectRegistration{
						CredentialRef: configTestCredentialRef,
						Description:   "Remove the persisted config test credential and disconnect this installation.",
					},
				},
			},
			Operations: []types.OperationRegistration{
				{
					Name:         configHealthCheckOperation.Name(),
					Description:  "Validate the config test installation",
					Topic:        types.NewDefinitionRef(definitionID).OperationTopic(configHealthCheckOperation.Name()),
					Policy:       types.ExecutionPolicy{Inline: true},
					ConfigSchema: configHealthSchema,
					Handle:       healthHandler,
				},
			},
		}, nil
	})
}

func userInputOnlyTestDefinitionBuilder(definitionID string) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          definitionID,
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
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationConfigResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	return resp
}

func mustMarshalConfigPayload(t *testing.T, payload handlers.IntegrationConfigPayload) []byte {
	t.Helper()

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	return body
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()

	body, err := json.Marshal(value)
	assert.NoError(t, err)

	return body
}

func decodeClientConfigField(t *testing.T, raw json.RawMessage, key string) string {
	t.Helper()

	document, err := jsonx.ToMap(raw)
	assert.NoError(t, err)

	value, _ := document[key].(string)

	return value
}
