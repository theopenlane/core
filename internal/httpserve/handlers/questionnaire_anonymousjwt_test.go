package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	models "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/iam/auth"
)

func (suite *HandlerTestSuite) TestCreateQuestionnaireAnonymousJWT() {
	t := suite.T()

	operation := suite.createImpersonationOperation("CreateQuestionnaireAnonymousJWT", "Create questionnaire anonymous JWT")
	suite.registerTestHandler("POST", "questionnaire/auth/anonymous", operation, suite.h.CreateQuestionnaireAnonymousJWT)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)
	ctx = contextx.With(ctx, auth.QuestionnaireContextKey{})

	template, err := suite.db.Template.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Test Template").
		SetTemplateType(enums.Document).
		SetKind(enums.TemplateKindQuestionnaire).
		SetJsonconfig(map[string]interface{}{
			"title": "Sample Questionnaire",
			"questions": []map[string]interface{}{
				{
					"id":       "q1",
					"question": "Sample question",
					"type":     "text",
				},
			},
		}).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Test Assessment").
		SetTemplateID(template.ID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		assessmentID   string
		expectedStatus int
		expectedError  string
		expectSuccess  bool
	}{
		{
			name:           "happy path - valid assessment ID",
			assessmentID:   assessment.ID,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "missing assessment ID",
			assessmentID:   "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "missing assessment ID",
		},
		{
			name:           "nonexistent assessment ID",
			assessmentID:   "nonexistent-id",
			expectedStatus: http.StatusNotFound,
			expectedError:  "assessment not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := "/questionnaire/auth/anonymous"
			if tc.assessmentID != "" {
				target = fmt.Sprintf("%s?assessment_id=%s", target, tc.assessmentID)
			}

			req := httptest.NewRequest(http.MethodPost, target, nil)

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			var out *models.CreateQuestionnaireAnonymousJWTResponse

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

	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}
