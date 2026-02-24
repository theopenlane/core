package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/echox/middleware/echocontext"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/iam/auth"
)

func (suite *HandlerTestSuite) TestCreateTrustCenterAnonymousJWT() {
	t := suite.T()

	// setup handler
	// Create operation for CreateTrustCenterAnonymousJWT
	operation := suite.createImpersonationOperation("CreateTrustCenterAnonymousJWT", "Create trust center anonymous JWT")
	suite.registerTestHandler("POST", "trustcenter/auth/anonymous", operation, suite.h.CreateTrustCenterAnonymousJWT)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)
	ctx = auth.WithCaller(ctx, auth.NewTrustCenterBootstrapCaller(""))
	mappableDomain, err := suite.db.MappableDomain.Create().
		SetName("trust.openlane.io").
		SetZoneID("1234").
		Save(ctx)
	require.NoError(t, err)

	customDomain, err := suite.db.CustomDomain.Create().
		SetCnameRecord("trust.meow.org").
		SetMappableDomainID(mappableDomain.ID).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)

	require.NoError(t, err)
	previewDomain, err := suite.db.CustomDomain.Create().
		SetCnameRecord("preview.meow.org").
		SetMappableDomainID(mappableDomain.ID).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)
	slug := "meow"

	trustCenterWithCD, err := suite.db.TrustCenter.Create().
		SetSlug(slug).
		SetOwnerID(testUser1.OrganizationID).
		SetCustomDomainID(customDomain.ID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)
	_, err = suite.db.TrustCenter.UpdateOneID(trustCenterWithCD.ID).
		SetPreviewDomainID(previewDomain.ID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	trustCenterNoCD, err := suite.db.TrustCenter.Create().
		SetSlug(slug).
		SetOwnerID(testUser2.OrganizationID).
		Save(testUser2.UserCtx)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		referer        string
		expectedStatus int
		expectedError  string
		expectSuccess  bool
	}{
		{
			name:           "happy path - default domain with slug",
			referer:        fmt.Sprintf("https://trust.openlane.com/%s", trustCenterNoCD.Slug),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "happy path - default domain with slug when custom domain exists",
			referer:        fmt.Sprintf("https://trust.openlane.com/%s", trustCenterWithCD.Slug),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "happy path - default domain with slug and path",
			referer:        fmt.Sprintf("https://trust.openlane.com/%s", trustCenterNoCD.Slug),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "happy path - custom domain",
			referer:        fmt.Sprintf("https://%s/any/path", customDomain.CnameRecord),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "happy path - custom domain normalized",
			referer:        fmt.Sprintf("https://%s./any/path", strings.ToUpper(customDomain.CnameRecord)),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "happy path - preview domain",
			referer:        fmt.Sprintf("https://%s/preview/path", previewDomain.CnameRecord),
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "missing referer",
			referer:        "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "referer is required",
		},
		{
			name:           "invalid referer URL",
			referer:        "not-a-valid-url",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid referer URL",
		},
		{
			name:           "referer with no hostname",
			referer:        "/just/a/path",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid referer URL",
		},
		{
			name:           "default domain missing slug",
			referer:        "https://trust.openlane.com/",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "slug is required in the path for default trust center domain",
		},
		{
			name:           "default domain empty slug",
			referer:        "https://trust.openlane.com//some/path",
			expectedStatus: http.StatusUnauthorized,

			expectedError: "slug is required in the path for default trust center domain",
		},
		{
			name:           "default domain nonexistent slug",
			referer:        "https://trust.openlane.com/nonexistent-slug",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "trust center not found",
		},
		{
			name:           "custom domain not found",
			referer:        "https://nonexistent.example.com/path",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "trust center not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := fmt.Sprintf("/trustcenter/auth/anonymous?referer=%s", tc.referer)

			req := httptest.NewRequest(http.MethodPost, target, nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			var out *models.CreateTrustCenterAnonymousJWTResponse

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			if tc.expectSuccess {
				require.NotNil(t, out)
				assert.NotEmpty(t, out.AccessToken, "access token should not be empty")
				assert.NotEmpty(t, out.RefreshToken, "refresh token should not be empty")
				assert.NotEmpty(t, out.Session, "session should not be empty")
				assert.Equal(t, "Bearer", out.TokenType, "token type should be bearer")
			} else {
				// For error cases, the response structure might be different
				// Let's check if we can decode it as a generic error response
				if out == nil {
					var errorResp map[string]interface{}
					res.Body.Close()
					req2 := httptest.NewRequest(http.MethodPost, target, nil)
					recorder2 := httptest.NewRecorder()
					suite.e.ServeHTTP(recorder2, req2)
					res2 := recorder2.Result()
					defer res2.Body.Close()

					if err := json.NewDecoder(res2.Body).Decode(&errorResp); err == nil {
						if errorMsg, ok := errorResp["error"].(string); ok {
							assert.Contains(t, errorMsg, tc.expectedError)
						}
					}
				}
			}
		})
	}
	suite.db.TrustCenter.DeleteOneID(trustCenterNoCD.ID).Exec(ctx)
	suite.db.TrustCenter.DeleteOneID(trustCenterWithCD.ID).Exec(ctx)
	suite.db.CustomDomain.DeleteOneID(customDomain.ID).Exec(ctx)
	suite.db.CustomDomain.DeleteOneID(previewDomain.ID).Exec(ctx)
}
