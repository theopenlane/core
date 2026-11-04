//go:build test

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestRunIntegrationOperationHealthCheckInline(t *testing.T) {
	suite := suite.setupParallelTest(t)

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
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))

	var resp handlers.RunIntegrationOperationResponse
	assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Check(t, resp.Success)
	assert.Check(t, is.Equal("ok", resp.Status))
	assert.Check(t, is.DeepEqual(opTestHealthCheckOperation.Name(), resp.Operation))
	assert.Check(t, is.Contains(resp.Summary, "Integration operation completed"))
}

func TestRunIntegrationOperationInlinePolicy(t *testing.T) {
	suite := suite.setupParallelTest(t)

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
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestInlineDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))

	var resp handlers.RunIntegrationOperationResponse
	assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Check(t, resp.Success)
	assert.Check(t, is.Equal("ok", resp.Status))
	assert.Check(t, is.DeepEqual(opTestRepoSyncOperation.Name(), resp.Operation))
	assert.Check(t, is.Contains(resp.Summary, "Integration operation completed"))
}

func TestRunIntegrationOperationQueuedAsync(t *testing.T) {
	suite := suite.setupParallelTest(t)

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
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))

	var resp handlers.RunIntegrationOperationResponse
	assert.NilError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Check(t, resp.Success)
	assert.Check(t, is.Equal("queued", resp.Status))
	assert.Check(t, is.DeepEqual(opTestRepoSyncOperation.Name(), resp.Operation))
	assert.Check(t, is.Contains(resp.Summary, "queued"))
	assert.Check(t, len(resp.Details) > 0)
}

func TestRunIntegrationOperationUnauthorized(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Assert(t, is.Equal(http.StatusUnauthorized, rec.Code))
}

func TestRunIntegrationOperationInvalidProvider(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/def_nonexistent/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestRunIntegrationOperationMissingOperationName(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	testUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		Body: handlers.RunIntegrationOperationBody{},
	})
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestRunIntegrationOperationUnknownOperation(t *testing.T) {
	suite := suite.setupParallelTest(t)

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
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestRunIntegrationOperationInvalidConfig(t *testing.T) {
	suite := suite.setupParallelTest(t)

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
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/v1/integrations/%s/operations?integration_id=%s", operationTestDefinitionID, integrationID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestRunIntegrationOperationInstallationNotFound(t *testing.T) {
	suite := suite.setupParallelTest(t)

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
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(testUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
}

func TestRunIntegrationOperationOtherOrgInstallationNotFound(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	owner := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})
	otherUser := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, owner.UserCtx, owner.OrganizationID, operationTestDefinitionID)

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: integrationID,
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(otherUser.UserCtx))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
	assert.Check(t, is.Contains(rec.Body.String(), handlers.ErrIntegrationNotFound.Error()))
}

func TestRunIntegrationOperationNoActiveOrgUnauthorized(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{operationTestDefinitionBuilder(operationTestDefinitionID, false)})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	owner := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	integrationID := suite.createOperationTestIntegration(t, owner.UserCtx, owner.OrganizationID, operationTestDefinitionID)

	// a system admin bypasses the org query filter, so the handler itself must refuse a caller with no active organization
	noOrgCtx := auth.NewTestContextWithOrgID(owner.ID, "", auth.WithCapabilities(auth.CapabilitiesForSystemAdmin(true)))
	noOrgCtx = ent.NewContext(noOrgCtx, suite.db)

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: integrationID,
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+integrationID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(noOrgCtx))

	assert.Assert(t, is.Equal(http.StatusUnauthorized, rec.Code))
}

func TestRunIntegrationOperationAnonymousTokenUnauthorized(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerAuthenticatedTestHandler(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	restore := suite.withDefinitionRuntime(t, []registry.Builder{
		operationTestDefinitionBuilder(operationTestInlineDefinitionID, true),
		operationTestDefinitionBuilder(operationTestDefinitionID, false),
	})
	defer restore()

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	owner := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	inlineIntegrationID := suite.createOperationTestIntegration(t, owner.UserCtx, owner.OrganizationID, operationTestInlineDefinitionID)
	queuedIntegrationID := suite.createOperationTestIntegration(t, owner.UserCtx, owner.OrganizationID, operationTestDefinitionID)

	questionnaireToken, err := urlx.GenerateAnonTokenURL(reqCtx, suite.h.DBClient.TokenManager, nil, url.URL{}, urlx.AnonTokenRequest{
		Prefix:    authmanager.AnonQuestionnaireJWTPrefix,
		SubjectID: ulids.New().String(),
		OrgID:     owner.OrganizationID,
		Email:     "respondent@example.com",
		Duration:  suite.h.DBClient.TokenManager.Config().AssessmentAccessDuration,
		ExtraClaims: func(c *tokens.Claims) {
			c.AssessmentID = ulids.New().String()
		},
	})
	assert.NilError(t, err)

	trustCenterSession, err := suite.h.AuthManager.GenerateAnonymousTrustCenterSession(reqCtx, httptest.NewRecorder(), owner.OrganizationID, ulids.New().String())
	assert.NilError(t, err)

	testCases := []struct {
		name          string
		accessToken   string
		definitionID  string
		integrationID string
	}{
		{name: "questionnaire token inline operation", accessToken: questionnaireToken.AccessToken, definitionID: operationTestInlineDefinitionID, integrationID: inlineIntegrationID},
		{name: "questionnaire token queued operation", accessToken: questionnaireToken.AccessToken, definitionID: operationTestDefinitionID, integrationID: queuedIntegrationID},
		{name: "trust center token inline operation", accessToken: trustCenterSession.AccessToken, definitionID: operationTestInlineDefinitionID, integrationID: inlineIntegrationID},
		{name: "trust center token queued operation", accessToken: trustCenterSession.AccessToken, definitionID: operationTestDefinitionID, integrationID: queuedIntegrationID},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
				IntegrationID: tc.integrationID,
				Body: handlers.RunIntegrationOperationBody{
					Operation: opTestRepoSyncOperation.Name(),
				},
			})
			assert.NilError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+tc.definitionID+"/operations?integration_id="+tc.integrationID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tc.accessToken)

			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req)

			assert.Assert(t, is.Equal(http.StatusUnauthorized, rec.Code))
		})
	}
}

func TestRunIntegrationOperationSystemDefinitionNotInvocable(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	owner := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	def, ok := suite.h.IntegrationsRuntime.Registry().Definition(operationTestSystemDefinitionID)
	assert.Assert(t, ok)
	assert.Assert(t, !def.Installable())

	// the configure handler refuses non-installable definitions, so this row can only exist by bypassing that surface
	record, _, err := suite.h.IntegrationsRuntime.EnsureInstallation(owner.UserCtx, owner.OrganizationID, "", def, nil, nil)
	assert.NilError(t, err)

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: record.ID,
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestSystemSweepOperation.Name(),
		},
	})
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestSystemDefinitionID+"/operations?integration_id="+record.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(owner.UserCtx))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
	assert.Check(t, is.Contains(rec.Body.String(), handlers.ErrIntegrationNotFound.Error()))
}

func TestRunIntegrationOperationSystemOwnedInstallationNotFound(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerRouteOnce(http.MethodPost, operationTestPath, suite.h.RunIntegrationOperation)

	reqCtx := echocontext.NewTestEchoContext().Request().Context()
	owner := suite.userBuilderWithInput(reqCtx, &userInput{confirmedUser: true})

	// a full-system admin with no organization creates an ownerless row, which the system-owned hook flags
	systemCtx := auth.NewTestContextWithOrgID(owner.ID, "", auth.WithCapabilities(auth.CapSystemAdmin|auth.CapBypassOrgFilter|auth.CapInternalOperation|auth.CapSystemSweep))
	systemCtx = ent.NewContext(systemCtx, suite.db)

	record, err := suite.db.Integration.Create().
		SetName("system installation").
		SetDefinitionID(operationTestDefinitionID).
		SetStatus(enums.IntegrationStatusConnected).
		Save(systemCtx)
	assert.NilError(t, err)
	assert.Assert(t, record.SystemOwned)
	assert.Assert(t, is.Len(record.OwnerID, 0))

	body, err := json.Marshal(handlers.RunIntegrationOperationRequest{
		IntegrationID: record.ID,
		Body: handlers.RunIntegrationOperationBody{
			Operation: opTestHealthCheckOperation.Name(),
		},
	})
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/"+operationTestDefinitionID+"/operations?integration_id="+record.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req.WithContext(owner.UserCtx))

	assert.Assert(t, is.Equal(http.StatusBadRequest, rec.Code))
	assert.Check(t, is.Contains(rec.Body.String(), handlers.ErrIntegrationNotFound.Error()))
}

func (suite *HandlerTestSuite) createOperationTestIntegration(t *testing.T, ctx context.Context, orgID, definitionID string) string {
	t.Helper()

	def, ok := suite.h.IntegrationsRuntime.Registry().Definition(definitionID)
	assert.Assert(t, ok)

	rec, _, err := suite.h.IntegrationsRuntime.EnsureInstallation(ctx, orgID, "", def, nil, nil)
	assert.NilError(t, err)

	credential := types.CredentialSet{
		Data: json.RawMessage(`{"token":"test-token"}`),
	}

	err = suite.h.IntegrationsRuntime.ReconcileCredential(ctx, rec, operationTestCredentialRef, credential)
	assert.NilError(t, err)

	return rec.ID
}
