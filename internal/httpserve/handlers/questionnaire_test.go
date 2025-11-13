package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/enums"
	models "github.com/theopenlane/core/pkg/openapi"
)

func (suite *HandlerTestSuite) TestGetQuestionnaire() {
	t := suite.T()

	operation := suite.createImpersonationOperation("GetQuestionnaire", "Get questionnaire template configuration")
	suite.registerAuthenticatedTestHandler("GET", "/questionnaire", operation, suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	templateType := enums.Document
	jsonConfig := map[string]any{
		"title":       "Test Assessment Template",
		"description": "A test questionnaire template",
		"questions": []map[string]any{
			{
				"id":       "q1",
				"question": "What is your name?",
				"type":     "text",
			},
			{
				"id":       "q2",
				"question": "What is your email?",
				"type":     "email",
			},
		},
	}

	template, err := suite.db.Template.Create().
		SetName("Test Assessment Template").
		SetTemplateType(templateType).
		SetJsonconfig(jsonConfig).
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

	testEmail := "test@example.com"

	// Create an assessment response for the email - need questionnaire context and authenticated user
	questionnaireCtx := contextx.With(ctx, auth.QuestionnaireContextKey{})
	questionnaireCtx = auth.WithAuthenticatedUser(questionnaireCtx, &auth.AuthenticatedUser{
		SubjectID:      testUser1.ID,
		OrganizationID: testUser1.OrganizationID,
	})
	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(testUser1.OrganizationID).
		Save(questionnaireCtx)
	require.NoError(t, err)

	authData, err := suite.h.AuthManager.GenerateAnonymousQuestionnaireSession(ctx, httptest.NewRecorder(), assessment.OwnerID, assessment.ID, testEmail)
	require.NoError(t, err)
	require.NotNil(t, authData)

	testCases := []struct {
		name           string
		accessToken    string
		expectedStatus int
		expectSuccess  bool
		expectedError  string
		validateJSON   func(*testing.T, *models.GetQuestionnaireResponse)
	}{
		{
			name:           "successful questionnaire retrieval",
			accessToken:    authData.AccessToken,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			validateJSON: func(t *testing.T, resp *models.GetQuestionnaireResponse) {
				assert.NotNil(t, resp.Jsonconfig)
				assert.Equal(t, "Test Assessment Template", resp.Jsonconfig["title"])
				assert.Equal(t, "A test questionnaire template", resp.Jsonconfig["description"])
				questions, ok := resp.Jsonconfig["questions"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, questions, 2)
			},
		},
		{
			name:           "missing authorization",
			accessToken:    "",
			expectedStatus: http.StatusUnauthorized,
			expectSuccess:  false,
			expectedError:  "unauthorized",
		},
		{
			name:           "invalid token",
			accessToken:    "invalid_token",
			expectedStatus: http.StatusUnauthorized,
			expectSuccess:  false,
			expectedError:  "unauthorized",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
			req.Header.Set("Content-Type", "application/json")

			if tc.accessToken != "" {
				req.Header.Set("Authorization", "Bearer "+tc.accessToken)
			}

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectSuccess {
				var out *models.GetQuestionnaireResponse

				err := json.NewDecoder(res.Body).Decode(&out)
				require.NoError(t, err, "error parsing response")
				require.NotNil(t, out)

				if tc.validateJSON != nil {
					tc.validateJSON(t, out)
				}
			} else {
				var errorResp map[string]interface{}
				err := json.NewDecoder(res.Body).Decode(&errorResp)
				if err == nil && tc.expectedError != "" {
					if errorMsg, ok := errorResp["error"].(string); ok {
						assert.Contains(t, errorMsg, tc.expectedError)
					}
				}
			}
		})
	}

	// Cleanup
	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(ctx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}

func (suite *HandlerTestSuite) TestGetQuestionnaireNoTemplate() {
	t := suite.T()

	operation := suite.createImpersonationOperation("GetQuestionnaire", "Get questionnaire template configuration")
	suite.registerAuthenticatedTestHandler("GET", "/questionnaire", operation, suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment No Template").
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(ctx)
	require.NoError(t, err)

	testEmail := "notemplate@example.com"

	// Create an assessment response for the email - need questionnaire context and authenticated user
	questionnaireCtx := contextx.With(ctx, auth.QuestionnaireContextKey{})
	questionnaireCtx = auth.WithAuthenticatedUser(questionnaireCtx, &auth.AuthenticatedUser{
		SubjectID:      testUser1.ID,
		OrganizationID: testUser1.OrganizationID,
	})
	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(testUser1.OrganizationID).
		Save(questionnaireCtx)
	require.NoError(t, err)

	authData, err := suite.h.AuthManager.GenerateAnonymousQuestionnaireSession(ctx, httptest.NewRecorder(), assessment.OwnerID, assessment.ID, testEmail)
	require.NoError(t, err)
	require.NotNil(t, authData)

	req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authData.AccessToken)

	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	// Cleanup
	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(ctx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
}

func (suite *HandlerTestSuite) TestGetQuestionnaireAlreadyCompleted() {
	t := suite.T()

	operation := suite.createImpersonationOperation("GetQuestionnaire", "Get questionnaire template configuration")
	suite.registerAuthenticatedTestHandler("GET", "/questionnaire", operation, suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	templateType := enums.Document
	jsonConfig := map[string]any{
		"title":       "Test Assessment Template",
		"description": "A test questionnaire template",
		"questions": []map[string]any{
			{
				"id":       "q1",
				"question": "What is your name?",
				"type":     "text",
			},
		},
	}

	template, err := suite.db.Template.Create().
		SetName("Test Assessment Template Already Completed").
		SetTemplateType(templateType).
		SetJsonconfig(jsonConfig).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment Already Completed").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	testEmail := "completed-get@example.com"
	completedAt := time.Now().Add(-1 * time.Hour)

	// Create questionnaire context with anonymous user - needed for both DocumentData and AssessmentResponse
	anonUser := &auth.AnonymousQuestionnaireUser{
		SubjectID:      fmt.Sprintf("anon_questionnaire_%s", uuid.New().String()),
		SubjectEmail:   testEmail,
		OrganizationID: testUser1.OrganizationID,
		AssessmentID:   assessment.ID,
	}
	questionnaireCtx := contextx.With(ctx, auth.QuestionnaireContextKey{})
	questionnaireCtx = auth.WithAnonymousQuestionnaireUser(questionnaireCtx, anonUser)
	questionnaireCtx = auth.WithAuthenticatedUser(questionnaireCtx, &auth.AuthenticatedUser{
		SubjectID:      testUser1.ID,
		OrganizationID: testUser1.OrganizationID,
	})

	documentData, err := suite.db.DocumentData.Create().
		SetTemplateID(template.ID).
		SetOwnerID(testUser1.OrganizationID).
		SetData(map[string]any{"q1": "previous answer"}).
		Save(questionnaireCtx)
	require.NoError(t, err)

	// Create AssessmentResponse - this will trigger email job insertion
	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(testUser1.OrganizationID).
		Save(questionnaireCtx)
	require.NoError(t, err)

	// Verify the email job was inserted
	job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
		[]rivertest.ExpectedJob{
			{
				Args: jobs.EmailArgs{},
			},
		})
	require.NotNil(t, job)

	// Update the assessment response to completed status using allowCtx (like the handler does)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.QuestionnaireContextKey{})
	allowCtx = auth.WithAnonymousQuestionnaireUser(allowCtx, anonUser)

	assessmentResponse, err = suite.db.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		SetCompletedAt(completedAt).
		SetDocumentDataID(documentData.ID).
		Save(allowCtx)
	require.NoError(t, err)

	anonUserID := fmt.Sprintf("anon_questionnaire_%s", uuid.New().String())
	claims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:       anonUserID,
		OrgID:        assessment.OwnerID,
		AssessmentID: assessment.ID,
		Email:        testEmail,
	}

	accessToken, _, err := suite.h.DBClient.TokenManager.CreateTokenPair(claims)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var errorResp map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&errorResp)
	require.NoError(t, err)

	if errorMsg, ok := errorResp["error"].(string); ok {
		assert.Contains(t, errorMsg, "already been completed")
	}

	// Cleanup
	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(ctx)
	suite.db.DocumentData.DeleteOneID(documentData.ID).Exec(ctx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}

func (suite *HandlerTestSuite) TestSubmitQuestionnaire() {
	t := suite.T()

	operation := suite.createImpersonationOperation("SubmitQuestionnaire", "Submit questionnaire response data")
	suite.registerAuthenticatedTestHandler("POST", "/questionnaire", operation, suite.h.SubmitQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	templateType := enums.Document
	jsonConfig := map[string]any{
		"title":       "Test Assessment Template",
		"description": "A test questionnaire template",
		"questions": []map[string]any{
			{
				"id":       "q1",
				"question": "What is your name?",
				"type":     "text",
			},
			{
				"id":       "q2",
				"question": "What is your email?",
				"type":     "email",
			},
		},
	}

	template, err := suite.db.Template.Create().
		SetName("Test Assessment Template For Submit").
		SetTemplateType(templateType).
		SetJsonconfig(jsonConfig).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment For Submit").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	testEmail := "test@example.com"

	anonUserID := fmt.Sprintf("anon_questionnaire_%s", uuid.New().String())
	claims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:       anonUserID,
		OrgID:        assessment.OwnerID,
		AssessmentID: assessment.ID,
		Email:        testEmail,
	}

	accessToken, _, err := suite.h.DBClient.TokenManager.CreateTokenPair(claims)
	require.NoError(t, err)

	authData := &models.AuthData{
		AccessToken: accessToken,
	}

	submissionData := models.SubmitQuestionnaireRequest{
		Data: map[string]any{
			"q1":               "John Doe",
			"q2":               "john.doe@example.com",
			"additional_notes": "Some additional information",
		},
	}

	testCases := []struct {
		name           string
		accessToken    string
		requestBody    models.SubmitQuestionnaireRequest
		expectedStatus int
		expectSuccess  bool
		expectedError  string
		validateJSON   func(*testing.T, *models.SubmitQuestionnaireResponse)
	}{
		{
			name:           "successful questionnaire submission",
			accessToken:    authData.AccessToken,
			requestBody:    submissionData,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			validateJSON: func(t *testing.T, resp *models.SubmitQuestionnaireResponse) {
				assert.NotEmpty(t, resp.DocumentDataID)
				assert.Equal(t, "COMPLETED", resp.Status)
				assert.NotEmpty(t, resp.CompletedAt)
			},
		},
		{
			name:           "missing authorization",
			accessToken:    "",
			requestBody:    submissionData,
			expectedStatus: http.StatusUnauthorized,
			expectSuccess:  false,
			expectedError:  "unauthorized",
		},
		{
			name:        "missing data",
			accessToken: authData.AccessToken,
			requestBody: models.SubmitQuestionnaireRequest{
				Data: map[string]any{},
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			expectedError:  "missing questionnaire data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			bodyBytes, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			if tc.accessToken != "" {
				req.Header.Set("Authorization", "Bearer "+tc.accessToken)
			}

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectSuccess {
				var out *models.SubmitQuestionnaireResponse

				err := json.NewDecoder(res.Body).Decode(&out)
				require.NoError(t, err, "error parsing response")
				require.NotNil(t, out)

				if tc.validateJSON != nil {
					tc.validateJSON(t, out)
				}

				documentData, err := suite.db.DocumentData.Get(ctx, out.DocumentDataID)
				require.NoError(t, err)
				assert.NotNil(t, documentData)
				assert.Equal(t, template.ID, documentData.TemplateID)
				assert.Equal(t, assessment.OwnerID, documentData.OwnerID)

				// Note: AssessmentResponse is created by the handler during submission
				// We don't pre-create it in test setup to avoid triggering email hooks
			} else {
				var errorResp map[string]interface{}
				err := json.NewDecoder(res.Body).Decode(&errorResp)
				if err == nil && tc.expectedError != "" {
					if errorMsg, ok := errorResp["error"].(string); ok {
						assert.Contains(t, errorMsg, tc.expectedError)
					}
				}
			}
		})
	}

	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}

func (suite *HandlerTestSuite) TestSubmitQuestionnaireAlreadyCompleted() {
	t := suite.T()

	operation := suite.createImpersonationOperation("SubmitQuestionnaire", "Submit questionnaire response data")
	suite.registerAuthenticatedTestHandler("POST", "/questionnaire", operation, suite.h.SubmitQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	templateType := enums.Document
	jsonConfig := map[string]any{
		"title": "Test Assessment Template",
	}

	template, err := suite.db.Template.Create().
		SetName("Test Assessment Template Already Completed").
		SetTemplateType(templateType).
		SetJsonconfig(jsonConfig).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment Already Completed").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	testEmail := "completed@example.com"
	completedAt := time.Now().Add(-1 * time.Hour)

	// Create questionnaire context with anonymous user - needed for both DocumentData and AssessmentResponse
	anonUser := &auth.AnonymousQuestionnaireUser{
		SubjectID:      fmt.Sprintf("anon_questionnaire_%s", uuid.New().String()),
		SubjectEmail:   testEmail,
		OrganizationID: testUser1.OrganizationID,
		AssessmentID:   assessment.ID,
	}
	questionnaireCtx := contextx.With(ctx, auth.QuestionnaireContextKey{})
	questionnaireCtx = auth.WithAnonymousQuestionnaireUser(questionnaireCtx, anonUser)
	questionnaireCtx = auth.WithAuthenticatedUser(questionnaireCtx, &auth.AuthenticatedUser{
		SubjectID:      testUser1.ID,
		OrganizationID: testUser1.OrganizationID,
	})

	documentData, err := suite.db.DocumentData.Create().
		SetTemplateID(template.ID).
		SetOwnerID(testUser1.OrganizationID).
		SetData(map[string]any{"q1": "previous answer"}).
		Save(questionnaireCtx)
	require.NoError(t, err)

	// Create AssessmentResponse - this will trigger email job insertion
	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(testUser1.OrganizationID).
		Save(questionnaireCtx)
	require.NoError(t, err)

	// Verify the email job was inserted
	job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
		[]rivertest.ExpectedJob{
			{
				Args: jobs.EmailArgs{},
			},
		})
	require.NotNil(t, job)

	// Update the assessment response to completed status using allowCtx (like the handler does)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = contextx.With(allowCtx, auth.QuestionnaireContextKey{})
	allowCtx = auth.WithAnonymousQuestionnaireUser(allowCtx, anonUser)

	assessmentResponse, err = suite.db.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		SetCompletedAt(completedAt).
		SetDocumentDataID(documentData.ID).
		Save(allowCtx)
	require.NoError(t, err)

	anonUserID := fmt.Sprintf("anon_questionnaire_%s", uuid.New().String())
	claims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:       anonUserID,
		OrgID:        assessment.OwnerID,
		AssessmentID: assessment.ID,
		Email:        testEmail,
	}

	accessToken, _, err := suite.h.DBClient.TokenManager.CreateTokenPair(claims)
	require.NoError(t, err)

	authData := &models.AuthData{
		AccessToken: accessToken,
	}

	submissionData := models.SubmitQuestionnaireRequest{
		Data: map[string]any{
			"q1": "New answer - should be rejected",
		},
	}

	bodyBytes, err := json.Marshal(submissionData)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authData.AccessToken)

	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var errorResp map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&errorResp)
	require.NoError(t, err)

	if errorMsg, ok := errorResp["error"].(string); ok {
		assert.Contains(t, errorMsg, "already been completed")
	}

	unchangedDocData, err := suite.db.DocumentData.Get(ctx, documentData.ID)
	require.NoError(t, err)
	assert.Equal(t, "previous answer", unchangedDocData.Data["q1"])

	// Cleanup
	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(ctx)
	suite.db.DocumentData.DeleteOneID(documentData.ID).Exec(ctx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}
