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

	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/types"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/ent/generated/integration"
)

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderSuccess() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderSuccess"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	payload := models.IntegrationConfigRequest{
		ProjectID:           "sample-project",
		ServiceAccountEmail: "svc@example.iam.gserviceaccount.com",
	}

	body, err := json.Marshal(models.IntegrationConfigPayload{
		IntegrationConfigParams: models.IntegrationConfigParams{Provider: "gcp_scc"},
		Body:                    payload,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcp_scc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp models.IntegrationConfigResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.True(t, resp.Success)
	assert.Equal(t, "gcp_scc", resp.Provider)

	stored := suite.db.Integration.Query().
		Where(
			integration.OwnerID(testUser.OrganizationID),
			integration.Kind("gcp_scc"),
		).
		WithSecrets().
		OnlyX(testUser.UserCtx)

	require.NotNil(t, stored)
	assert.Len(t, stored.Edges.Secrets, 1)

	credential := stored.Edges.Secrets[0]
	assert.Contains(t, credential.CredentialSet.ProviderData, "projectId")
}

func (suite *HandlerTestSuite) TestConfigureIntegrationProviderInvalidPayload() {
	t := suite.T()

	op := openapi3.NewOperation()
	op.OperationID = "ConfigureIntegrationProviderInvalid"
	suite.registerRouteOnce(http.MethodPost, "/v1/integrations/:provider/config", op, suite.h.ConfigureIntegrationProvider)

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	payload := models.IntegrationConfigRequest{
		ServiceAccountEmail: "svc@example.iam.gserviceaccount.com",
	}

	body, err := json.Marshal(models.IntegrationConfigPayload{
		IntegrationConfigParams: models.IntegrationConfigParams{Provider: "gcp_scc"},
		Body:                    payload,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcp_scc/config", bytes.NewReader(body))
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

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	payload := models.IntegrationConfigRequest{
		ProjectID:           "sample-project",
		ServiceAccountEmail: "svc@example.iam.gserviceaccount.com",
	}

	body, err := json.Marshal(models.IntegrationConfigPayload{
		IntegrationConfigParams: models.IntegrationConfigParams{Provider: ""},
		Body:                    payload,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcp_scc/config", bytes.NewReader(body))
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

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	payload := models.IntegrationConfigRequest{
		ProjectID:           "sample-project",
		ServiceAccountEmail: "svc@example.iam.gserviceaccount.com",
	}

	body, err := json.Marshal(models.IntegrationConfigPayload{
		IntegrationConfigParams: models.IntegrationConfigParams{Provider: "unknown_provider"},
		Body:                    payload,
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

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(models.IntegrationConfigPayload{
		IntegrationConfigParams: models.IntegrationConfigParams{Provider: "gcp_scc"},
		Body:                    models.IntegrationConfigRequest{},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcp_scc/config", bytes.NewReader(body))
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

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
	})
	defer restore()

	payload := models.IntegrationConfigRequest{
		ProjectID:           "sample-project",
		ServiceAccountEmail: "svc@example.iam.gserviceaccount.com",
	}

	body, err := json.Marshal(models.IntegrationConfigPayload{
		IntegrationConfigParams: models.IntegrationConfigParams{Provider: "gcp_scc"},
		Body:                    payload,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcp_scc/config", bytes.NewReader(body))
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

	restore := suite.withIntegrationRegistry(t, map[types.ProviderType]config.ProviderSpec{
		types.ProviderType("gcp_scc"): gcpSCCSpec(),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	initialPayload := models.IntegrationConfigRequest{
		ProjectID:           "initial-project",
		ServiceAccountEmail: "initial@example.iam.gserviceaccount.com",
	}

	body, err := json.Marshal(models.IntegrationConfigPayload{
		IntegrationConfigParams: models.IntegrationConfigParams{Provider: "gcp_scc"},
		Body:                    initialPayload,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/gcp_scc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	updatedPayload := models.IntegrationConfigRequest{
		ProjectID:           "updated-project",
		ServiceAccountEmail: "updated@example.iam.gserviceaccount.com",
	}

	body, err = json.Marshal(models.IntegrationConfigPayload{
		IntegrationConfigParams: models.IntegrationConfigParams{Provider: "gcp_scc"},
		Body:                    updatedPayload,
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/v1/integrations/gcp_scc/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(testUser.UserCtx)

	rec = httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	stored := suite.db.Integration.Query().
		Where(
			integration.OwnerID(testUser.OrganizationID),
			integration.Kind("gcp_scc"),
		).
		WithSecrets().
		OnlyX(testUser.UserCtx)

	require.NotNil(t, stored)
	assert.Len(t, stored.Edges.Secrets, 1)

	credential := stored.Edges.Secrets[0]
	assert.Equal(t, "updated-project", credential.CredentialSet.ProviderData["projectId"])
}
