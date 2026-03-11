//go:build test

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	integrationspec "github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderSuccess() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderSuccess"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]integrationspec.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	credFields := map[string]any{
		"projectId":           "sample-project",
		"serviceAccountEmail": "svc@example.iam.gserviceaccount.com",
		"audience":            "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
	}

	credJSON, err := json.Marshal(credFields)
	require.NoError(t, err)

	body, err := json.Marshal(handlers.IntegrationConfigPayload{
		Provider: "gcpscc",
		Body:     handlers.IntegrationConfigBody(credJSON),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcpscc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp handlers.IntegrationConfigResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "gcpscc", resp.Provider)

	stored := suite.db.Integration.Query().
		Where(
			integration.OwnerID(testUser.OrganizationID),
			integration.Kind("gcpscc"),
		).
		WithSecrets().
		OnlyX(testUser.UserCtx)

	require.NotNil(t, stored)
	assert.Len(t, stored.Edges.Secrets, 1)

	credential := stored.Edges.Secrets[0]
	assert.Contains(t, string(credential.CredentialSet.ProviderData), "projectId")
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderInvalidPayload() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderInvalid"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]integrationspec.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	// missing required projectId field
	credFields := map[string]any{
		"serviceAccountEmail": "svc@example.iam.gserviceaccount.com",
	}

	credJSON, err := json.Marshal(credFields)
	require.NoError(t, err)

	body, err := json.Marshal(handlers.IntegrationConfigPayload{
		Provider: "gcpscc",
		Body:     handlers.IntegrationConfigBody(credJSON),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcpscc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var resp struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.True(t, strings.TrimSpace(resp.Error) != "")
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderMissingProvider() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderMissingProvider"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]integrationspec.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	credFields := map[string]any{
		"projectId":           "sample-project",
		"serviceAccountEmail": "svc@example.iam.gserviceaccount.com",
		"audience":            "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
	}

	credJSON, err := json.Marshal(credFields)
	require.NoError(t, err)

	body, err := json.Marshal(handlers.IntegrationConfigPayload{
		Provider: "",
		Body:     handlers.IntegrationConfigBody(credJSON),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcpscc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUnknownProvider() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUnknown"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]integrationspec.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	credFields := map[string]any{
		"projectId":           "sample-project",
		"serviceAccountEmail": "svc@example.iam.gserviceaccount.com",
		"audience":            "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
	}

	credJSON, err := json.Marshal(credFields)
	require.NoError(t, err)

	body, err := json.Marshal(handlers.IntegrationConfigPayload{
		Provider: "unknown_provider",
		Body:     handlers.IntegrationConfigBody(credJSON),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/unknown_provider/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderEmptyPayload() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderEmpty"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]integrationspec.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.IntegrationConfigPayload{
		Provider: "gcpscc",
		Body:     handlers.IntegrationConfigBody(`{}`),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcpscc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUnauthorized() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUnauthorized"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]integrationspec.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
	})
	defer restore()

	credFields := map[string]any{
		"projectId":           "sample-project",
		"serviceAccountEmail": "svc@example.iam.gserviceaccount.com",
		"audience":            "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
	}

	credJSON, err := json.Marshal(credFields)
	require.NoError(t, err)

	body, err := json.Marshal(handlers.IntegrationConfigPayload{
		Provider: "gcpscc",
		Body:     handlers.IntegrationConfigBody(credJSON),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcpscc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderUpdateExisting() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderUpdate"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]integrationspec.ProviderSpec{
		types.ProviderType("gcpscc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	initialFields := map[string]any{
		"projectId":           "initial-project",
		"serviceAccountEmail": "initial@example.iam.gserviceaccount.com",
		"audience":            "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
	}

	credJSON, err := json.Marshal(initialFields)
	require.NoError(t, err)

	body, err := json.Marshal(handlers.IntegrationConfigPayload{
		Provider: "gcpscc",
		Body:     handlers.IntegrationConfigBody(credJSON),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcpscc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	updatedFields := map[string]any{
		"projectId":           "updated-project",
		"serviceAccountEmail": "updated@example.iam.gserviceaccount.com",
		"audience":            "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/pool/providers/provider",
	}

	credJSON, err = json.Marshal(updatedFields)
	require.NoError(t, err)

	body, err = json.Marshal(handlers.IntegrationConfigPayload{
		Provider: "gcpscc",
		Body:     handlers.IntegrationConfigBody(credJSON),
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/v1/integrations/gcpscc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec = httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	stored := suite.db.Integration.Query().
		Where(
			integration.OwnerID(testUser.OrganizationID),
			integration.Kind("gcpscc"),
		).
		WithSecrets().
		OnlyX(testUser.UserCtx)

	require.NotNil(t, stored)
	assert.Len(t, stored.Edges.Secrets, 1)

	credential := stored.Edges.Secrets[0]
	providerData, err := jsonx.ToMap(credential.CredentialSet.ProviderData)
	require.NoError(t, err)
	assert.Equal(t, "updated-project", providerData["projectId"])
}
