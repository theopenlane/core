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
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
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

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	testEmail := "test@example.com"

	questionnaireCtx := auth.WithCaller(ctx, auth.NewQuestionnaireCaller(testUser1.OrganizationID, testUser1.ID, "", testEmail))
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
			expectedError:  "no authorization header",
		},
		{
			name:           "invalid token",
			accessToken:    "invalid_token",
			expectedStatus: http.StatusUnauthorized,
			expectSuccess:  false,
			expectedError:  "token is malformed",
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

	templateType := enums.Document
	jsonConfig := map[string]any{
		"title":       "Test Assessment Template Missing",
		"description": "A test questionnaire template that will be deleted",
		"questions": []map[string]any{
			{
				"id":       "q1",
				"question": "What is your name?",
				"type":     "text",
			},
		},
	}

	template, err := suite.db.Template.Create().
		SetName("Test Assessment Template Missing").
		SetTemplateType(templateType).
		SetJsonconfig(jsonConfig).
		SetKind(enums.TemplateKindQuestionnaire).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment No Template").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	err = suite.db.Template.DeleteOneID(template.ID).Exec(testUser1.UserCtx)
	require.NoError(t, err)

	testEmail := "notemplate@example.com"

	questionnaireCtx := auth.WithCaller(ctx, auth.NewQuestionnaireCaller(testUser1.OrganizationID, testUser1.ID, "", testEmail))
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

	assert.Equal(t, http.StatusOK, recorder.Code)

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

	anonUser := auth.NewQuestionnaireCaller(testUser1.OrganizationID, fmt.Sprintf("anon_questionnaire_%s", ulids.New().String()), "", testEmail)
	questionnaireCtx := auth.WithCaller(ctx, anonUser)
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	documentData, err := suite.db.DocumentData.Create().
		SetTemplateID(template.ID).
		SetOwnerID(testUser1.OrganizationID).
		SetData(map[string]any{"q1": "previous answer"}).
		Save(questionnaireCtx)
	require.NoError(t, err)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(testUser1.OrganizationID).
		Save(questionnaireCtx)
	require.NoError(t, err)

	job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
		[]rivertest.ExpectedJob{
			{
				Args: jobs.EmailArgs{},
			},
		})
	require.NotNil(t, job)

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, anonUser)
	allowCtx = auth.ActiveAssessmentIDKey.Set(allowCtx, assessment.ID)

	assessmentResponse, err = suite.db.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		SetCompletedAt(completedAt).
		SetDocumentDataID(documentData.ID).
		Save(allowCtx)
	require.NoError(t, err)

	anonUserID := fmt.Sprintf("anon_questionnaire_%s", ulids.New().String())
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

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(ctx)
	suite.db.DocumentData.DeleteOneID(documentData.ID).Exec(ctx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}

func (suite *HandlerTestSuite) TestGetQuestionnaireReturnsDraftData() {
	t := suite.T()

	submitOp := suite.createImpersonationOperation("SubmitQuestionnaire", "Submit questionnaire response data")
	suite.registerAuthenticatedTestHandler("POST", "/questionnaire", submitOp, suite.h.SubmitQuestionnaire)

	getOp := suite.createImpersonationOperation("GetQuestionnaire", "Get questionnaire template configuration")
	suite.registerAuthenticatedTestHandler("GET", "/questionnaire", getOp, suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	jsonConfig := map[string]any{
		"title": "Draft Fetch Test",
		"questions": []map[string]any{
			{"id": "q1", "question": "Name?", "type": "text"},
			{"id": "q2", "question": "Role?", "type": "text"},
		},
	}

	template, err := suite.db.Template.Create().
		SetName("Draft Fetch Template").
		SetTemplateType(enums.Document).
		SetJsonconfig(jsonConfig).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Draft Fetch Assessment").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	testEmail := "draftfetch@example.com"
	anonUserID := fmt.Sprintf("anon_questionnaire_%s", assessment.ID)

	anonUser := auth.NewQuestionnaireCaller(testUser1.OrganizationID, anonUserID, "", testEmail)
	questionnaireCtx := auth.WithCaller(ctx, anonUser)
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(testUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusSent).
		Save(questionnaireCtx)
	require.NoError(t, err)

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

	// fetch before any draft - no saved data
	t.Run("no saved data before draft", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var out models.GetQuestionnaireResponse
		err := json.NewDecoder(recorder.Result().Body).Decode(&out)
		require.NoError(t, err)

		assert.NotNil(t, out.Jsonconfig)
		assert.Equal(t, "Draft Fetch Test", out.Jsonconfig["title"])
		assert.Nil(t, out.SavedData)
	})

	// save a draft
	var documentDataID string

	t.Run("save draft", func(t *testing.T) {
		draftReq := models.SubmitQuestionnaireRequest{
			Data:    map[string]any{"q1": "Alice", "q2": "Engineer"},
			IsDraft: true,
		}

		bodyBytes, err := json.Marshal(draftReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)

		var out models.SubmitQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, "DRAFT", out.Status)

		documentDataID = out.DocumentDataID
	})

	// fetch after draft - saved data returned alongside form config
	t.Run("saved data returned after draft", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var out models.GetQuestionnaireResponse
		err := json.NewDecoder(recorder.Result().Body).Decode(&out)
		require.NoError(t, err)

		assert.NotNil(t, out.Jsonconfig)
		assert.Equal(t, "Draft Fetch Test", out.Jsonconfig["title"])

		require.NotNil(t, out.SavedData)
		assert.Equal(t, "Alice", out.SavedData["q1"])
		assert.Equal(t, "Engineer", out.SavedData["q2"])
	})

	if documentDataID != "" {
		suite.db.DocumentData.DeleteOneID(documentDataID).Exec(questionnaireCtx)
	}

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(questionnaireCtx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(questionnaireCtx)
	suite.db.Template.DeleteOneID(template.ID).Exec(questionnaireCtx)
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

	anonUserID := fmt.Sprintf("anon_questionnaire_%s", ulids.New().String())
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

	anonUser := auth.NewQuestionnaireCaller(assessment.OwnerID, anonUserID, "", testEmail)
	questionnaireCtx := auth.WithCaller(ctx, anonUser)
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(testUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusSent).
		Save(questionnaireCtx)
	require.NoError(t, err)

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
			expectedError:  "no authorization header",
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

	var createdDocumentDataIDs []string

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

				documentData, err := suite.db.DocumentData.Get(questionnaireCtx, out.DocumentDataID)
				require.NoError(t, err)
				assert.NotNil(t, documentData)
				assert.Equal(t, assessment.OwnerID, documentData.OwnerID)

				updatedResponse, err := suite.db.AssessmentResponse.Get(questionnaireCtx, assessmentResponse.ID)
				require.NoError(t, err)
				assert.Equal(t, enums.AssessmentResponseStatusCompleted, updatedResponse.Status)
				assert.Equal(t, documentData.ID, updatedResponse.DocumentDataID)
				assert.NotZero(t, updatedResponse.CompletedAt)

				createdDocumentDataIDs = append(createdDocumentDataIDs, documentData.ID)
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

	for _, docID := range createdDocumentDataIDs {
		err := suite.db.DocumentData.DeleteOneID(docID).Exec(questionnaireCtx)
		require.NoError(t, err)
	}

	err = suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(questionnaireCtx)
	require.NoError(t, err)
	err = suite.db.Assessment.DeleteOneID(assessment.ID).Exec(questionnaireCtx)
	require.NoError(t, err)
	err = suite.db.Template.DeleteOneID(template.ID).Exec(questionnaireCtx)
	require.NoError(t, err)
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

	anonUser := auth.NewQuestionnaireCaller(testUser1.OrganizationID, fmt.Sprintf("anon_questionnaire_%s", ulids.New().String()), "", testEmail)
	questionnaireCtx := auth.WithCaller(ctx, anonUser)
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	documentData, err := suite.db.DocumentData.Create().
		SetTemplateID(template.ID).
		SetOwnerID(testUser1.OrganizationID).
		SetData(map[string]any{"q1": "previous answer"}).
		Save(questionnaireCtx)
	require.NoError(t, err)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(testUser1.OrganizationID).
		Save(questionnaireCtx)
	require.NoError(t, err)

	job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
		[]rivertest.ExpectedJob{
			{
				Args: jobs.EmailArgs{},
			},
		})
	require.NotNil(t, job)

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	allowCtx = auth.WithCaller(allowCtx, anonUser)
	allowCtx = auth.ActiveAssessmentIDKey.Set(allowCtx, assessment.ID)

	assessmentResponse, err = suite.db.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		SetCompletedAt(completedAt).
		SetDocumentDataID(documentData.ID).
		Save(allowCtx)
	require.NoError(t, err)

	anonUserID := fmt.Sprintf("anon_questionnaire_%s", ulids.New().String())
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

	unchangedDocData, err := suite.db.DocumentData.Get(allowCtx, documentData.ID)
	require.NoError(t, err)
	assert.Equal(t, "previous answer", unchangedDocData.Data["q1"])

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(allowCtx)
	suite.db.DocumentData.DeleteOneID(documentData.ID).Exec(allowCtx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(allowCtx)
	suite.db.Template.DeleteOneID(template.ID).Exec(allowCtx)
}

func (suite *HandlerTestSuite) TestSubmitQuestionnaireAuthenticatedUser() {
	t := suite.T()

	operation := suite.createImpersonationOperation("SubmitQuestionnaire", "Submit questionnaire response data")
	suite.registerTestHandler("POST", "/questionnaire", operation, suite.h.SubmitQuestionnaire)

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
		SetName("Test Assessment Template Auth User").
		SetTemplateType(templateType).
		SetJsonconfig(jsonConfig).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment Auth User").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testUser1.UserInfo.Email).
		SetOwnerID(testUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusSent).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	submissionData := models.SubmitQuestionnaireRequest{
		AssessmentID: assessment.ID,
		Data: map[string]any{
			"q1": "Jane Doe",
		},
	}

	submissionDataMissingID := models.SubmitQuestionnaireRequest{
		Data: map[string]any{
			"q1": "Jane Doe",
		},
	}

	testCases := []struct {
		name           string
		requestBody    models.SubmitQuestionnaireRequest
		expectedStatus int
		expectSuccess  bool
		expectedError  string
		validateJSON   func(*testing.T, *models.SubmitQuestionnaireResponse)
	}{
		{
			name:           "successful questionnaire submission authenticated user",
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
			name:           "missing assessment id authenticated user",
			requestBody:    submissionDataMissingID,
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			expectedError:  "missing assessment ID",
		},
	}

	var documentDataIDs []string

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			bodyBytes, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			reqCtx := testUser1.UserCtx
			reqCtx = auth.WithCaller(reqCtx, &auth.Caller{
				SubjectID:      testUser1.ID,
				SubjectEmail:   testUser1.UserInfo.Email,
				OrganizationID: testUser1.OrganizationID,
			})

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req.WithContext(reqCtx))

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

				documentData, err := suite.db.DocumentData.Get(testUser1.UserCtx, out.DocumentDataID)
				require.NoError(t, err)
				assert.NotNil(t, documentData)
				assert.Equal(t, assessment.OwnerID, documentData.OwnerID)

				updatedResponse, err := suite.db.AssessmentResponse.Get(testUser1.UserCtx, assessmentResponse.ID)
				require.NoError(t, err)
				assert.Equal(t, enums.AssessmentResponseStatusCompleted, updatedResponse.Status)
				assert.Equal(t, documentData.ID, updatedResponse.DocumentDataID)
				assert.NotZero(t, updatedResponse.CompletedAt)

				documentDataIDs = append(documentDataIDs, documentData.ID)

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

	for _, docID := range documentDataIDs {
		err := suite.db.DocumentData.DeleteOneID(docID).Exec(testUser1.UserCtx)
		require.NoError(t, err)
	}

	err = suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(testUser1.UserCtx)
	require.NoError(t, err)
	err = suite.db.Assessment.DeleteOneID(assessment.ID).Exec(testUser1.UserCtx)
	require.NoError(t, err)
	err = suite.db.Template.DeleteOneID(template.ID).Exec(testUser1.UserCtx)
	require.NoError(t, err)
}

func (suite *HandlerTestSuite) TestSubmitQuestionnaireDraft() {
	t := suite.T()

	operation := suite.createImpersonationOperation("SubmitQuestionnaire", "Submit questionnaire response data")
	suite.registerAuthenticatedTestHandler("POST", "/questionnaire", operation, suite.h.SubmitQuestionnaire)

	getOperation := suite.createImpersonationOperation("GetQuestionnaire", "Get questionnaire template configuration")
	suite.registerAuthenticatedTestHandler("GET", "/questionnaire", getOperation, suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	template, err := suite.db.Template.Create().
		SetName("Test Assessment Template Draft").
		SetTemplateType(enums.Document).
		SetJsonconfig(map[string]any{"title": "Draft Test"}).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment Draft").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	testEmail := "draft@example.com"
	anonUserID := fmt.Sprintf("anon_questionnaire_%s", assessment.ID)

	anonUser := auth.NewQuestionnaireCaller(testUser1.OrganizationID, anonUserID, "", testEmail)
	questionnaireCtx := auth.WithCaller(ctx, anonUser)
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(testUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusSent).
		Save(questionnaireCtx)
	require.NoError(t, err)

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

	var documentDataID string

	t.Run("save draft", func(t *testing.T) {
		draftReq := models.SubmitQuestionnaireRequest{
			Data:    map[string]any{"q1": "partial answer"},
			IsDraft: true,
		}

		bodyBytes, err := json.Marshal(draftReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var out models.SubmitQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		require.NoError(t, err)

		assert.NotEmpty(t, out.DocumentDataID)
		assert.Equal(t, "DRAFT", out.Status)
		assert.Empty(t, out.CompletedAt)

		documentDataID = out.DocumentDataID

		updated, err := suite.db.AssessmentResponse.Get(questionnaireCtx, assessmentResponse.ID)
		require.NoError(t, err)
		assert.Equal(t, enums.AssessmentResponseStatusDraft, updated.Status)
		assert.Equal(t, documentDataID, updated.DocumentDataID)
	})

	t.Run("update draft with new data", func(t *testing.T) {
		draftReq := models.SubmitQuestionnaireRequest{
			Data:    map[string]any{"q1": "updated partial answer", "q2": "new field"},
			IsDraft: true,
		}

		bodyBytes, err := json.Marshal(draftReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var out models.SubmitQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, documentDataID, out.DocumentDataID)
		assert.Equal(t, "DRAFT", out.Status)

		docData, err := suite.db.DocumentData.Get(questionnaireCtx, documentDataID)
		require.NoError(t, err)
		assert.Equal(t, "updated partial answer", docData.Data["q1"])
		assert.Equal(t, "new field", docData.Data["q2"])
	})

	t.Run("get questionnaire returns saved draft data", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var out models.GetQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		require.NoError(t, err)

		assert.NotNil(t, out.SavedData)
		assert.Equal(t, "updated partial answer", out.SavedData["q1"])
		assert.Equal(t, "new field", out.SavedData["q2"])
	})

	t.Run("complete from draft", func(t *testing.T) {
		submitReq := models.SubmitQuestionnaireRequest{
			Data: map[string]any{"q1": "final answer", "q2": "final field"},
		}

		bodyBytes, err := json.Marshal(submitReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var out models.SubmitQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, documentDataID, out.DocumentDataID)
		assert.Equal(t, "COMPLETED", out.Status)
		assert.NotEmpty(t, out.CompletedAt)

		updated, err := suite.db.AssessmentResponse.Get(questionnaireCtx, assessmentResponse.ID)
		require.NoError(t, err)
		assert.Equal(t, enums.AssessmentResponseStatusCompleted, updated.Status)
		assert.NotZero(t, updated.CompletedAt)

		docData, err := suite.db.DocumentData.Get(questionnaireCtx, documentDataID)
		require.NoError(t, err)
		assert.Equal(t, "final answer", docData.Data["q1"])
	})

	t.Run("cannot draft after completed", func(t *testing.T) {
		draftReq := models.SubmitQuestionnaireRequest{
			Data:    map[string]any{"q1": "should fail"},
			IsDraft: true,
		}

		bodyBytes, err := json.Marshal(draftReq)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	if documentDataID != "" {
		suite.db.DocumentData.DeleteOneID(documentDataID).Exec(questionnaireCtx)
	}

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(questionnaireCtx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(questionnaireCtx)
	suite.db.Template.DeleteOneID(template.ID).Exec(questionnaireCtx)
}
