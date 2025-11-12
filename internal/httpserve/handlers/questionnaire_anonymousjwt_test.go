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

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	models "github.com/theopenlane/core/pkg/openapi"
)

func (suite *HandlerTestSuite) TestCreateQuestionnaireAnonymousJWT() {
	t := suite.T()

	// setup handler
	// Create operation for CreateQuestionnaireAnonymousJWT
	operation := suite.createImpersonationOperation("CreateQuestionnaireAnonymousJWT", "Create questionnaire anonymous JWT")
	suite.registerTestHandler("POST", "questionnaire/auth/anonymous", operation, suite.h.CreateQuestionnaireAnonymousJWT)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	// Create a test template
	templateType := enums.Document
	template, err := suite.db.Template.Create().
		SetName("Test Assessment Template").
		SetTemplateType(templateType).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	// Create a test assessment
	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		assessmentID   string
		expectedStatus int
		expectSuccess  bool
		expectedError  string
	}{
		{
			name:           "successful anonymous JWT creation",
			assessmentID:   assessment.ID,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "assessment not found",
			assessmentID:   "nonexistent-assessment-id",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "assessment not found",
		},
		{
			name:           "missing assessment ID",
			assessmentID:   "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "assessment_id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"assessment_id":"%s"}`, tc.assessmentID)
			req := httptest.NewRequest(http.MethodPost, "/questionnaire/auth/anonymous", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			var out *models.CreateQuestionnaireAnonymousJWTResponse

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
				if out == nil {
					var errorResp map[string]interface{}
					res.Body.Close()
					req2 := httptest.NewRequest(http.MethodPost, "/questionnaire/auth/anonymous", strings.NewReader(body))
					req2.Header.Set("Content-Type", "application/json")
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

	// Cleanup
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}

