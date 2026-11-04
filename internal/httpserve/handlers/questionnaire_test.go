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
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/assessmentresponse"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationbrandingsetting"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestGetQuestionnaire(t *testing.T) {
	suite := suite.setupTest(t)

	suite.registerQuestionnaireTestHandler("GET", "/questionnaire", suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := auth.WithInternalOperationContext(ec)

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
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	_, err = suite.db.OrganizationBrandingSetting.Update().
		Where(organizationbrandingsetting.OrganizationID(suite.testUser1.OrganizationID)).
		SetRemoveAssessmentBranding(true).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	testEmail := "test@example.com"

	questionnaireCtx := auth.WithInternalOperationContext(auth.WithCaller(ctx, auth.NewQuestionnaireCaller(suite.testUser1.OrganizationID, suite.testUser1.ID, "", testEmail)))
	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(questionnaireCtx)
	assert.NilError(t, err)

	anonUserID := fmt.Sprintf("anon_questionnaire_%s", assessment.ID)

	accessToken, _, err := suite.h.DBClient.TokenManager.CreateTokenPair(&tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: anonUserID},
		UserID:           anonUserID,
		OrgID:            assessment.OwnerID,
		AssessmentID:     assessment.ID,
		Email:            testEmail,
	})
	assert.NilError(t, err)

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
			accessToken:    accessToken,
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			validateJSON: func(t *testing.T, resp *models.GetQuestionnaireResponse) {
				assert.Check(t, resp.Jsonconfig != nil)
				assert.Check(t, is.Equal("Test Assessment Template", resp.Jsonconfig["title"]))
				assert.Check(t, is.Equal("A test questionnaire template", resp.Jsonconfig["description"]))
				questions, ok := resp.Jsonconfig["questions"].([]interface{})
				assert.Check(t, ok)
				assert.Check(t, is.Len(questions, 2))
				assert.Check(t, resp.RemoveBranding)
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

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))

			if tc.expectSuccess {
				var out *models.GetQuestionnaireResponse

				err := json.NewDecoder(res.Body).Decode(&out)
				assert.NilError(t, err, "error parsing response")
				assert.Assert(t, out != nil)

				if tc.validateJSON != nil {
					tc.validateJSON(t, out)
				}
			} else {
				var errorResp map[string]interface{}
				err := json.NewDecoder(res.Body).Decode(&errorResp)
				if err == nil && tc.expectedError != "" {
					if errorMsg, ok := errorResp["error"].(string); ok {
						assert.Check(t, is.Contains(errorMsg, tc.expectedError))
					}
				}
			}
		})
	}

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(ctx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}

func TestGetQuestionnaireNoTemplate(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerQuestionnaireTestHandler("GET", "/questionnaire", suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := auth.WithInternalOperationContext(ec)

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
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment No Template").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	err = suite.db.Template.DeleteOneID(template.ID).Exec(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	testEmail := "notemplate@example.com"

	questionnaireCtx := auth.WithInternalOperationContext(auth.WithCaller(ctx, auth.NewQuestionnaireCaller(suite.testUser1.OrganizationID, suite.testUser1.ID, "", testEmail)))
	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(questionnaireCtx)
	assert.NilError(t, err)

	noTemplateAnonUserID := fmt.Sprintf("anon_questionnaire_%s", assessment.ID)

	noTemplateToken, _, err := suite.h.DBClient.TokenManager.CreateTokenPair(&tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: noTemplateAnonUserID},
		UserID:           noTemplateAnonUserID,
		OrgID:            assessment.OwnerID,
		AssessmentID:     assessment.ID,
		Email:            testEmail,
	})
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+noTemplateToken)

	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(ctx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
}

func TestGetQuestionnaireAlreadyCompleted(t *testing.T) {
	suite := suite.setupTest(t)

	suite.registerQuestionnaireTestHandler("GET", "/questionnaire", suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := auth.WithInternalOperationContext(ec)

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
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment Already Completed").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	testEmail := "completed-get@example.com"
	completedAt := time.Now().Add(-1 * time.Hour)

	anonUser := auth.NewQuestionnaireCaller(suite.testUser1.OrganizationID, fmt.Sprintf("anon_questionnaire_%s", ulids.New().String()), "", testEmail)
	questionnaireCtx := auth.WithInternalOperationContext(auth.WithCaller(ctx, anonUser))
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	documentData, err := suite.db.DocumentData.Create().
		SetTemplateID(template.ID).
		SetOwnerID(suite.testUser1.OrganizationID).
		SetData(map[string]any{"q1": "previous answer"}).
		Save(questionnaireCtx)
	assert.NilError(t, err)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(questionnaireCtx)
	assert.NilError(t, err)

	suite.WaitForEvents()

	msgs := suite.mockEmailSender().Messages()
	assert.Assert(t, is.Len(msgs, 1))

	internalCtx := auth.WithInternalOperationContext(auth.WithCaller(ctx, anonUser))
	internalCtx = auth.ActiveAssessmentIDKey.Set(internalCtx, assessment.ID)

	assessmentResponse, err = suite.db.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		SetCompletedAt(completedAt).
		SetDocumentDataID(documentData.ID).
		Save(internalCtx)
	assert.NilError(t, err)

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
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))

	var errorResp map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&errorResp)
	assert.NilError(t, err)

	if errorMsg, ok := errorResp["error"].(string); ok {
		assert.Check(t, is.Contains(errorMsg, "already been completed"))
	}

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(ctx)
	suite.db.DocumentData.DeleteOneID(documentData.ID).Exec(ctx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}

func TestQuestionnaireWithCampaign(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerQuestionnaireTestHandler("GET", "/questionnaire", suite.h.GetQuestionnaire)
	suite.registerQuestionnaireTestHandler("POST", "/questionnaire", suite.h.SubmitQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := auth.WithInternalOperationContext(ec)

	jsonConfig := map[string]any{
		"title": "Campaign Questionnaire",
		"questions": []map[string]any{
			{"id": "q1", "question": "What is your name?", "type": "text"},
		},
	}

	template, err := suite.db.Template.Create().
		SetName("Campaign Questionnaire Template").
		SetTemplateType(enums.Document).
		SetJsonconfig(jsonConfig).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Campaign Questionnaire Assessment").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	campaign, err := suite.db.Campaign.Create().
		SetName("Campaign Questionnaire Campaign").
		SetAssessmentID(assessment.ID).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	testEmail := "campaign@theopenlane.io"

	anonUser := auth.NewQuestionnaireCaller(suite.testUser1.OrganizationID, fmt.Sprintf("anon_questionnaire_%s", assessment.ID), "", testEmail)

	questionnaireCtx := auth.WithInternalOperationContext(auth.WithCaller(ctx, anonUser))
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	assessmentResponseFn := func(campaignID, marker string) (string, string) {
		create := suite.db.AssessmentResponse.Create().
			SetAssessmentID(assessment.ID).
			SetEmail(testEmail).
			SetOwnerID(suite.testUser1.OrganizationID)

		if campaignID != "" {
			create = create.SetCampaignID(campaignID)
		}

		response, err := create.Save(questionnaireCtx)
		assert.NilError(t, err)

		documentData, err := suite.db.DocumentData.Create().
			SetTemplateID(template.ID).
			SetOwnerID(suite.testUser1.OrganizationID).
			SetData(map[string]any{"q1": marker}).
			Save(questionnaireCtx)
		assert.NilError(t, err)

		_, err = suite.db.AssessmentResponse.UpdateOneID(response.ID).
			SetDocumentDataID(documentData.ID).
			Save(questionnaireCtx)
		assert.NilError(t, err)

		return response.ID, documentData.ID
	}

	responseWithoutCampaignID, responseWithoutCampaignDocID := assessmentResponseFn("", "orphan answer")
	campaignResponseWithID, campaignDocDataResponseID := assessmentResponseFn(campaign.ID, "here is my answer")

	anonUserID := fmt.Sprintf("anon_questionnaire_%s", assessment.ID)
	accessToken, _, err := suite.h.DBClient.TokenManager.CreateTokenPair(&tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: anonUserID},
		UserID:           anonUserID,
		OrgID:            assessment.OwnerID,
		AssessmentID:     assessment.ID,
		CampaignID:       campaign.ID,
		Email:            testEmail,
	})
	assert.NilError(t, err)

	t.Run("resolves the campaign linked assessment response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		var out models.GetQuestionnaireResponse
		err := json.NewDecoder(recorder.Result().Body).Decode(&out)
		assert.NilError(t, err)

		assert.Check(t, is.Equal("Campaign Questionnaire", out.Jsonconfig["title"]))
		assert.Check(t, is.Equal("here is my answer", out.SavedData["q1"]))
	})

	t.Run("campaign linked response can be submitted", func(t *testing.T) {
		submitReq := models.SubmitQuestionnaireRequest{
			Data: map[string]any{"q1": "here is my answer"},
		}

		bodyBytes, err := json.Marshal(submitReq)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		var out models.SubmitQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		assert.NilError(t, err)

		assert.Check(t, is.DeepEqual(campaignDocDataResponseID, out.DocumentDataID))
		assert.Check(t, is.Equal("COMPLETED", out.Status))
		assert.Check(t, len(out.CompletedAt) > 0)

		updated, err := suite.db.AssessmentResponse.Get(questionnaireCtx, campaignResponseWithID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(enums.AssessmentResponseStatusCompleted, updated.Status))

		doc, err := suite.db.DocumentData.Get(questionnaireCtx, campaignDocDataResponseID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("here is my answer", doc.Data["q1"]))

		// verify this is still in sent and only the response with a campaign was updated
		resp, err := suite.db.AssessmentResponse.Get(questionnaireCtx, responseWithoutCampaignID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(enums.AssessmentResponseStatusSent, resp.Status))
	})

	suite.db.DocumentData.DeleteOneID(campaignDocDataResponseID).Exec(questionnaireCtx)
	suite.db.DocumentData.DeleteOneID(responseWithoutCampaignDocID).Exec(questionnaireCtx)
	suite.db.AssessmentResponse.DeleteOneID(campaignResponseWithID).Exec(questionnaireCtx)
	suite.db.AssessmentResponse.DeleteOneID(responseWithoutCampaignID).Exec(questionnaireCtx)
	suite.db.Campaign.DeleteOneID(campaign.ID).Exec(questionnaireCtx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(questionnaireCtx)
	suite.db.Template.DeleteOneID(template.ID).Exec(questionnaireCtx)
}

func TestAnonymousQuestionnaireRejectsDraft(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerQuestionnaireTestHandler("POST", "/questionnaire", suite.h.SubmitQuestionnaire)
	suite.registerQuestionnaireTestHandler("GET", "/questionnaire", suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := auth.WithInternalOperationContext(ec)

	jsonConfig := map[string]any{
		"title": "Draft Rejection Test",
		"questions": []map[string]any{
			{"id": "q1", "question": "Name?", "type": "text"},
		},
	}

	template, err := suite.db.Template.Create().
		SetName("Draft Rejection Template").
		SetTemplateType(enums.Document).
		SetJsonconfig(jsonConfig).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Draft Rejection Assessment").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	// truly anonymous token — no email, like a shareable system-owned assessment link
	anonUserID := fmt.Sprintf("anon_questionnaire_%s", assessment.ID)

	anonClaims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:       anonUserID,
		OrgID:        assessment.OwnerID,
		AssessmentID: assessment.ID,
	}

	anonToken, _, err := suite.h.DBClient.TokenManager.CreateTokenPair(anonClaims)
	assert.NilError(t, err)

	t.Run("reject draft for anonymous user without email", func(t *testing.T) {
		draftReq := models.SubmitQuestionnaireRequest{
			Data:    map[string]any{"q1": "Alice"},
			IsDraft: true,
		}

		bodyBytes, err := json.Marshal(draftReq)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+anonToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Assert(t, is.Equal(http.StatusBadRequest, recorder.Code))

		var errorResp map[string]interface{}
		err = json.NewDecoder(recorder.Result().Body).Decode(&errorResp)
		assert.NilError(t, err)

		if errorMsg, ok := errorResp["error"].(string); ok {
			assert.Check(t, is.Contains(errorMsg, "cannot be saved as drafts"))
		}
	})

	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}

func TestIdentifiedQuestionnaireAllowsDraft(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerQuestionnaireTestHandler("POST", "/questionnaire", suite.h.SubmitQuestionnaire)
	suite.registerQuestionnaireTestHandler("GET", "/questionnaire", suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := auth.WithInternalOperationContext(ec)

	jsonConfig := map[string]any{
		"title": "Identified Draft Test",
		"questions": []map[string]any{
			{"id": "q1", "question": "Name?", "type": "text"},
			{"id": "q2", "question": "Role?", "type": "text"},
		},
	}

	template, err := suite.db.Template.Create().
		SetName("Identified Draft Template").
		SetTemplateType(enums.Document).
		SetJsonconfig(jsonConfig).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Identified Draft Assessment").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	testEmail := "identified-draft@example.com"
	anonUserID := fmt.Sprintf("anon_questionnaire_%s", assessment.ID)

	anonUser := auth.NewQuestionnaireCaller(suite.testUser1.OrganizationID, anonUserID, "", testEmail)
	questionnaireCtx := auth.WithInternalOperationContext(auth.WithCaller(ctx, anonUser))
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(suite.testUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusSent).
		Save(questionnaireCtx)
	assert.NilError(t, err)

	// identified token — has email, like a campaign-sent questionnaire link
	identifiedClaims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:       anonUserID,
		OrgID:        assessment.OwnerID,
		AssessmentID: assessment.ID,
		Email:        testEmail,
	}

	identifiedToken, _, err := suite.h.DBClient.TokenManager.CreateTokenPair(identifiedClaims)
	assert.NilError(t, err)

	t.Run("no saved data before draft", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+identifiedToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		var out models.GetQuestionnaireResponse
		err := json.NewDecoder(recorder.Result().Body).Decode(&out)
		assert.NilError(t, err)

		assert.Check(t, out.Jsonconfig != nil)
		assert.Check(t, is.Equal("Identified Draft Test", out.Jsonconfig["title"]))
		assert.Check(t, is.Nil(out.SavedData))
	})

	var documentDataID string

	t.Run("allow draft for identified user with email", func(t *testing.T) {
		draftReq := models.SubmitQuestionnaireRequest{
			Data:    map[string]any{"q1": "Alice", "q2": "Engineer"},
			IsDraft: true,
		}

		bodyBytes, err := json.Marshal(draftReq)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+identifiedToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Assert(t, is.Equal(http.StatusOK, recorder.Code))

		var out models.SubmitQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		assert.NilError(t, err)

		assert.Check(t, len(out.DocumentDataID) > 0)
		assert.Check(t, is.Equal("DRAFT", out.Status))
		assert.Check(t, is.Len(out.CompletedAt, 0))

		documentDataID = out.DocumentDataID

		updatedResponse, err := suite.db.AssessmentResponse.Get(questionnaireCtx, assessmentResponse.ID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(enums.AssessmentResponseStatusDraft, updatedResponse.Status))
		assert.Check(t, is.DeepEqual(documentDataID, updatedResponse.DocumentDataID))
	})

	t.Run("saved data returned after draft", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+identifiedToken)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		var out models.GetQuestionnaireResponse
		err := json.NewDecoder(recorder.Result().Body).Decode(&out)
		assert.NilError(t, err)

		assert.Check(t, out.SavedData != nil)
		assert.Check(t, is.Equal("Alice", out.SavedData["q1"]))
		assert.Check(t, is.Equal("Engineer", out.SavedData["q2"]))
	})

	if documentDataID != "" {
		suite.db.DocumentData.DeleteOneID(documentDataID).Exec(questionnaireCtx)
	}

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(questionnaireCtx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(questionnaireCtx)
	suite.db.Template.DeleteOneID(template.ID).Exec(questionnaireCtx)
}

func TestSubmitQuestionnaire(t *testing.T) {
	suite := suite.setupTest(t)

	suite.registerQuestionnaireTestHandler("POST", "/questionnaire", suite.h.SubmitQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := auth.WithInternalOperationContext(ec)

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
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment For Submit").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

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
	assert.NilError(t, err)

	authData := &models.AuthData{
		AccessToken: accessToken,
	}

	anonUser := auth.NewQuestionnaireCaller(assessment.OwnerID, anonUserID, "", testEmail)
	questionnaireCtx := auth.WithInternalOperationContext(auth.WithCaller(ctx, anonUser))
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(suite.testUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusSent).
		Save(questionnaireCtx)
	assert.NilError(t, err)

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
				assert.Check(t, len(resp.DocumentDataID) > 0)
				assert.Check(t, is.Equal("COMPLETED", resp.Status))
				assert.Check(t, len(resp.CompletedAt) > 0)
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
			assert.NilError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			if tc.accessToken != "" {
				req.Header.Set("Authorization", "Bearer "+tc.accessToken)
			}

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))

			if tc.expectSuccess {
				var out *models.SubmitQuestionnaireResponse

				err := json.NewDecoder(res.Body).Decode(&out)
				assert.NilError(t, err, "error parsing response")
				assert.Assert(t, out != nil)

				if tc.validateJSON != nil {
					tc.validateJSON(t, out)
				}

				documentData, err := suite.db.DocumentData.Get(questionnaireCtx, out.DocumentDataID)
				assert.NilError(t, err)
				assert.Check(t, documentData != nil)
				assert.Check(t, is.DeepEqual(assessment.OwnerID, documentData.OwnerID))

				updatedResponse, err := suite.db.AssessmentResponse.Get(questionnaireCtx, assessmentResponse.ID)
				assert.NilError(t, err)
				assert.Check(t, is.DeepEqual(enums.AssessmentResponseStatusCompleted, updatedResponse.Status))
				assert.Check(t, is.DeepEqual(documentData.ID, updatedResponse.DocumentDataID))
				assert.Check(t, !updatedResponse.CompletedAt.IsZero())

				createdDocumentDataIDs = append(createdDocumentDataIDs, documentData.ID)
			} else {
				var errorResp map[string]interface{}
				err := json.NewDecoder(res.Body).Decode(&errorResp)
				if err == nil && tc.expectedError != "" {
					if errorMsg, ok := errorResp["error"].(string); ok {
						assert.Check(t, is.Contains(errorMsg, tc.expectedError))
					}
				}
			}
		})
	}

	for _, docID := range createdDocumentDataIDs {
		err := suite.db.DocumentData.DeleteOneID(docID).Exec(questionnaireCtx)
		assert.NilError(t, err)
	}

	err = suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(questionnaireCtx)
	assert.NilError(t, err)
	err = suite.db.Assessment.DeleteOneID(assessment.ID).Exec(questionnaireCtx)
	assert.NilError(t, err)
	err = suite.db.Template.DeleteOneID(template.ID).Exec(questionnaireCtx)
	assert.NilError(t, err)
}

func TestSubmitQuestionnaireAlreadyCompleted(t *testing.T) {
	suite := suite.setupTest(t)

	suite.registerQuestionnaireTestHandler("POST", "/questionnaire", suite.h.SubmitQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := auth.WithInternalOperationContext(ec)

	templateType := enums.Document
	jsonConfig := map[string]any{
		"title": "Test Assessment Template",
	}

	template, err := suite.db.Template.Create().
		SetName("Test Assessment Template Already Completed").
		SetTemplateType(templateType).
		SetJsonconfig(jsonConfig).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment Already Completed").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	testEmail := "completed@example.com"
	completedAt := time.Now().Add(-1 * time.Hour)

	anonUser := auth.NewQuestionnaireCaller(suite.testUser1.OrganizationID, fmt.Sprintf("anon_questionnaire_%s", ulids.New().String()), "", testEmail)
	questionnaireCtx := auth.WithInternalOperationContext(auth.WithCaller(ctx, anonUser))
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	documentData, err := suite.db.DocumentData.Create().
		SetTemplateID(template.ID).
		SetOwnerID(suite.testUser1.OrganizationID).
		SetData(map[string]any{"q1": "previous answer"}).
		Save(questionnaireCtx)
	assert.NilError(t, err)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(testEmail).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(questionnaireCtx)
	assert.NilError(t, err)

	suite.WaitForEvents()

	msgs := suite.mockEmailSender().Messages()
	assert.Assert(t, is.Len(msgs, 1))

	internalCtx := auth.WithInternalOperationContext(auth.WithCaller(ctx, anonUser))
	internalCtx = auth.ActiveAssessmentIDKey.Set(internalCtx, assessment.ID)

	assessmentResponse, err = suite.db.AssessmentResponse.UpdateOneID(assessmentResponse.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		SetCompletedAt(completedAt).
		SetDocumentDataID(documentData.ID).
		Save(internalCtx)
	assert.NilError(t, err)

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
	assert.NilError(t, err)

	authData := &models.AuthData{
		AccessToken: accessToken,
	}

	submissionData := models.SubmitQuestionnaireRequest{
		Data: map[string]any{
			"q1": "New answer - should be rejected",
		},
	}

	bodyBytes, err := json.Marshal(submissionData)
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authData.AccessToken)

	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))

	var errorResp map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&errorResp)
	assert.NilError(t, err)

	if errorMsg, ok := errorResp["error"].(string); ok {
		assert.Check(t, is.Contains(errorMsg, "already been completed"))
	}

	unchangedDocData, err := suite.db.DocumentData.Get(internalCtx, documentData.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal("previous answer", unchangedDocData.Data["q1"]))

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(internalCtx)
	suite.db.DocumentData.DeleteOneID(documentData.ID).Exec(internalCtx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(internalCtx)
	suite.db.Template.DeleteOneID(template.ID).Exec(internalCtx)
}

func TestSubmitQuestionnaireCaller(t *testing.T) {
	suite := suite.setupParallelTest(t).withTestUsers()

	suite.registerTestHandler("POST", "/questionnaire", suite.h.SubmitQuestionnaire)

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
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment Auth User").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(suite.testUser1.UserInfo.Email).
		SetOwnerID(suite.testUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusSent).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

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

	otherOrgAssessment, err := suite.db.Assessment.Create().
		SetName("Other Org Assessment").
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetJsonconfig(map[string]any{"title": "Other Org Assessment"}).
		SetOwnerID(suite.testUser2.OrganizationID).
		Save(suite.testUser2.UserCtx)
	assert.NilError(t, err)

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
				assert.Check(t, len(resp.DocumentDataID) > 0)
				assert.Check(t, is.Equal("COMPLETED", resp.Status))
				assert.Check(t, len(resp.CompletedAt) > 0)
			},
		},
		{
			name:           "missing assessment id authenticated user",
			requestBody:    submissionDataMissingID,
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			expectedError:  "missing assessment ID",
		},
		{
			name: "assessment in another organization",
			requestBody: models.SubmitQuestionnaireRequest{
				AssessmentID: otherOrgAssessment.ID,
				Data:         submissionData.Data,
			},
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
			expectedError:  "assessment not found",
		},
		{
			name: "assessment that does not exist",
			requestBody: models.SubmitQuestionnaireRequest{
				AssessmentID: ulids.New().String(),
				Data:         submissionData.Data,
			},
			expectedStatus: http.StatusNotFound,
			expectSuccess:  false,
			expectedError:  "assessment not found",
		},
	}

	var documentDataIDs []string

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			bodyBytes, err := json.Marshal(tc.requestBody)
			assert.NilError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			reqCtx := suite.testUser1.UserCtx
			reqCtx = auth.WithCaller(reqCtx, &auth.Caller{
				SubjectID:      suite.testUser1.ID,
				SubjectEmail:   suite.testUser1.UserInfo.Email,
				OrganizationID: suite.testUser1.OrganizationID,
			})

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req.WithContext(reqCtx))

			res := recorder.Result()
			defer res.Body.Close()

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))

			if tc.expectSuccess {
				var out *models.SubmitQuestionnaireResponse

				err := json.NewDecoder(res.Body).Decode(&out)
				assert.NilError(t, err, "error parsing response")
				assert.Assert(t, out != nil)

				if tc.validateJSON != nil {
					tc.validateJSON(t, out)
				}

				documentData, err := suite.db.DocumentData.Get(suite.testUser1.UserCtx, out.DocumentDataID)
				assert.NilError(t, err)
				assert.Check(t, documentData != nil)
				assert.Check(t, is.DeepEqual(assessment.OwnerID, documentData.OwnerID))

				updatedResponse, err := suite.db.AssessmentResponse.Get(suite.testUser1.UserCtx, assessmentResponse.ID)
				assert.NilError(t, err)
				assert.Check(t, is.DeepEqual(enums.AssessmentResponseStatusCompleted, updatedResponse.Status))
				assert.Check(t, is.DeepEqual(documentData.ID, updatedResponse.DocumentDataID))
				assert.Check(t, !updatedResponse.CompletedAt.IsZero())

				documentDataIDs = append(documentDataIDs, documentData.ID)

			} else {
				var errorResp map[string]interface{}
				err := json.NewDecoder(res.Body).Decode(&errorResp)
				if err == nil && tc.expectedError != "" {
					if errorMsg, ok := errorResp["error"].(string); ok {
						assert.Check(t, is.Contains(errorMsg, tc.expectedError))
					}
				}
			}
		})
	}

	t.Run("org member with a response can submit", func(t *testing.T) {
		member := suite.userBuilder(context.Background())

		err := suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
			OrganizationID: suite.testUser1.OrganizationID,
			UserID:         member.ID,
			Role:           &enums.RoleMember,
		}).Exec(suite.testUser1.UserCtx)
		assert.NilError(t, err)

		memberResponse, err := suite.db.AssessmentResponse.Create().
			SetAssessmentID(assessment.ID).
			SetEmail(member.UserInfo.Email).
			SetOwnerID(suite.testUser1.OrganizationID).
			SetStatus(enums.AssessmentResponseStatusSent).
			Save(suite.testUser1.UserCtx)
		assert.NilError(t, err)

		t.Cleanup(func() {
			assert.NilError(t, suite.db.AssessmentResponse.DeleteOneID(memberResponse.ID).Exec(suite.testUser1.UserCtx))
		})

		bodyBytes, err := json.Marshal(submissionData)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		reqCtx := auth.WithCaller(member.UserCtx, &auth.Caller{
			SubjectID:      member.ID,
			SubjectEmail:   member.UserInfo.Email,
			OrganizationID: suite.testUser1.OrganizationID,
		})

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req.WithContext(reqCtx))

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		updatedResponse, err := suite.db.AssessmentResponse.Get(suite.testUser1.UserCtx, memberResponse.ID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(enums.AssessmentResponseStatusCompleted, updatedResponse.Status))

		if updatedResponse.DocumentDataID != "" {
			documentDataIDs = append(documentDataIDs, updatedResponse.DocumentDataID)
		}
	})

	rejectedCallers := []struct {
		name   string
		caller *auth.Caller
	}{
		{
			name: "anonymous trust center caller cannot submit",
			caller: &auth.Caller{
				SubjectID:        "anon_trustcenter_" + ulids.New().String(),
				OrganizationID:   suite.testUser1.OrganizationID,
				OrganizationRole: auth.AnonymousRole,
				Capabilities:     auth.CapTrustCenterAnonymous,
			},
		},
		{
			name: "api token cannot submit",
			caller: &auth.Caller{
				SubjectID:          ulids.New().String(),
				OrganizationID:     suite.testUser1.OrganizationID,
				OrganizationIDs:    []string{suite.testUser1.OrganizationID},
				AuthenticationType: auth.APITokenAuthentication,
			},
		},
		{
			name: "support access caller cannot submit",
			caller: &auth.Caller{
				SubjectID:          ulids.New().String(),
				OrganizationID:     suite.testUser1.OrganizationID,
				OrganizationIDs:    []string{suite.testUser1.OrganizationID},
				AuthenticationType: auth.JWTAuthentication,
				Capabilities:       auth.CapOrgSupport,
			},
		},
	}

	for _, rc := range rejectedCallers {
		t.Run(rc.name, func(t *testing.T) {
			before, err := suite.db.AssessmentResponse.Query().Where(assessmentresponse.AssessmentIDEQ(assessment.ID)).Count(suite.testUser1.UserCtx)
			assert.NilError(t, err)

			bodyBytes, err := json.Marshal(submissionData)
			assert.NilError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			suite.e.ServeHTTP(recorder, req.WithContext(auth.WithCaller(context.Background(), rc.caller)))

			assert.Check(t, is.Equal(http.StatusUnauthorized, recorder.Code))

			after, err := suite.db.AssessmentResponse.Query().Where(assessmentresponse.AssessmentIDEQ(assessment.ID)).Count(suite.testUser1.UserCtx)
			assert.NilError(t, err)
			assert.Check(t, is.DeepEqual(before, after))
		})
	}

	for _, docID := range documentDataIDs {
		err := suite.db.DocumentData.DeleteOneID(docID).Exec(suite.testUser1.UserCtx)
		assert.NilError(t, err)
	}

	err = suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(suite.testUser1.UserCtx)
	assert.NilError(t, err)
	err = suite.db.Assessment.DeleteOneID(assessment.ID).Exec(suite.testUser1.UserCtx)
	assert.NilError(t, err)
	err = suite.db.Template.DeleteOneID(template.ID).Exec(suite.testUser1.UserCtx)
	assert.NilError(t, err)
	err = suite.db.Assessment.DeleteOneID(otherOrgAssessment.ID).Exec(suite.testUser2.UserCtx)
	assert.NilError(t, err)
}

func TestSubmitQuestionnaireDraft(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerTestHandler("POST", "/questionnaire", suite.h.SubmitQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := auth.WithInternalOperationContext(ec)

	template, err := suite.db.Template.Create().
		SetName("Test Assessment Template Draft").
		SetTemplateType(enums.Document).
		SetJsonconfig(map[string]any{"title": "Draft Test"}).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Test Assessment Draft").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(suite.testUser1.OrganizationID).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	assessmentResponse, err := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessment.ID).
		SetEmail(suite.testUser1.UserInfo.Email).
		SetOwnerID(suite.testUser1.OrganizationID).
		SetStatus(enums.AssessmentResponseStatusSent).
		Save(suite.testUser1.UserCtx)
	assert.NilError(t, err)

	reqCtx := auth.WithCaller(ctx, &auth.Caller{
		SubjectID:      suite.testUser1.ID,
		SubjectEmail:   suite.testUser1.UserInfo.Email,
		OrganizationID: suite.testUser1.OrganizationID,
	})

	var documentDataID string

	t.Run("save draft", func(t *testing.T) {
		draftReq := models.SubmitQuestionnaireRequest{
			AssessmentID: assessment.ID,
			Data:         map[string]any{"q1": "partial answer"},
			IsDraft:      true,
		}

		bodyBytes, err := json.Marshal(draftReq)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req.WithContext(reqCtx))

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		var out models.SubmitQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		assert.NilError(t, err)

		assert.Check(t, len(out.DocumentDataID) > 0)
		assert.Check(t, is.Equal("DRAFT", out.Status))
		assert.Check(t, is.Len(out.CompletedAt, 0))

		documentDataID = out.DocumentDataID

		updated, err := suite.db.AssessmentResponse.Get(suite.testUser1.UserCtx, assessmentResponse.ID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(enums.AssessmentResponseStatusDraft, updated.Status))
		assert.Check(t, is.DeepEqual(documentDataID, updated.DocumentDataID))
	})

	t.Run("update draft with new data", func(t *testing.T) {
		draftReq := models.SubmitQuestionnaireRequest{
			AssessmentID: assessment.ID,
			Data:         map[string]any{"q1": "updated partial answer", "q2": "new field"},
			IsDraft:      true,
		}

		bodyBytes, err := json.Marshal(draftReq)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req.WithContext(reqCtx))

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		var out models.SubmitQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		assert.NilError(t, err)

		assert.Check(t, is.DeepEqual(documentDataID, out.DocumentDataID))
		assert.Check(t, is.Equal("DRAFT", out.Status))

		docData, err := suite.db.DocumentData.Get(suite.testUser1.UserCtx, documentDataID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("updated partial answer", docData.Data["q1"]))
		assert.Check(t, is.Equal("new field", docData.Data["q2"]))
	})

	t.Run("complete from draft", func(t *testing.T) {
		submitReq := models.SubmitQuestionnaireRequest{
			AssessmentID: assessment.ID,
			Data:         map[string]any{"q1": "final answer", "q2": "final field"},
		}

		bodyBytes, err := json.Marshal(submitReq)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req.WithContext(reqCtx))

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		var out models.SubmitQuestionnaireResponse
		err = json.NewDecoder(recorder.Result().Body).Decode(&out)
		assert.NilError(t, err)

		assert.Check(t, is.DeepEqual(documentDataID, out.DocumentDataID))
		assert.Check(t, is.Equal("COMPLETED", out.Status))
		assert.Check(t, len(out.CompletedAt) > 0)

		updated, err := suite.db.AssessmentResponse.Get(suite.testUser1.UserCtx, assessmentResponse.ID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(enums.AssessmentResponseStatusCompleted, updated.Status))
		assert.Check(t, !updated.CompletedAt.IsZero())

		docData, err := suite.db.DocumentData.Get(suite.testUser1.UserCtx, documentDataID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal("final answer", docData.Data["q1"]))
	})

	t.Run("cannot draft after completed", func(t *testing.T) {
		draftReq := models.SubmitQuestionnaireRequest{
			AssessmentID: assessment.ID,
			Data:         map[string]any{"q1": "should not be allowed"},
			IsDraft:      true,
		}

		bodyBytes, err := json.Marshal(draftReq)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req.WithContext(reqCtx))

		assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))
	})

	if documentDataID != "" {
		suite.db.DocumentData.DeleteOneID(documentDataID).Exec(suite.testUser1.UserCtx)
	}

	suite.db.AssessmentResponse.DeleteOneID(assessmentResponse.ID).Exec(suite.testUser1.UserCtx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(suite.testUser1.UserCtx)
	suite.db.Template.DeleteOneID(template.ID).Exec(suite.testUser1.UserCtx)
}
