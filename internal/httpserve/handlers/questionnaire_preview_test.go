package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func (suite *HandlerTestSuite) TestGetQuestionnairePreviewAndRealResponse() {
	t := suite.T()

	suite.registerAuthenticatedTestHandler("GET", "/questionnaire", suite.h.GetQuestionnaire)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	jsonConfig := map[string]any{
		"title": "Preview vs Real Template",
		"questions": []map[string]any{
			{"id": "q1", "question": "What is your name?", "type": "text"},
		},
	}

	template, err := suite.db.Template.Create().
		SetName("Preview vs Real Template").
		SetTemplateType(enums.Document).
		SetJsonconfig(jsonConfig).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	assert.NilError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetName("Preview vs Real Assessment").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	assert.NilError(t, err)

	testEmail := "preview-vs-real@example.com"

	anonUser := auth.NewQuestionnaireCaller(testUser1.OrganizationID, fmt.Sprintf("anon_questionnaire_%s", assessment.ID), "", testEmail)
	questionnaireCtx := auth.WithCaller(ctx, anonUser)
	questionnaireCtx = auth.ActiveAssessmentIDKey.Set(questionnaireCtx, assessment.ID)

	allowCtx := auth.WithCaller(ctx, anonUser)
	allowCtx = auth.ActiveAssessmentIDKey.Set(allowCtx, assessment.ID)

	realResponse := suite.createResponseWithData(questionnaireCtx, allowCtx, assessment.ID, testEmail, template.ID, false, "real answer")
	testResponse := suite.createResponseWithData(questionnaireCtx, allowCtx, assessment.ID, testEmail, template.ID, true, "test answer")

	testCases := []struct {
		name     string
		preview  bool
		wantData string
	}{
		{name: "real token retrieves the real response", preview: false, wantData: "real answer"},
		{name: "preview token retrieves the test response", preview: true, wantData: "test answer"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			anonUserID := fmt.Sprintf("anon_questionnaire_%s", assessment.ID)
			accessToken, _, err := suite.h.DBClient.TokenManager.CreateTokenPair(&tokens.Claims{
				RegisteredClaims:  jwt.RegisteredClaims{Subject: anonUserID},
				UserID:            anonUserID,
				OrgID:             assessment.OwnerID,
				AssessmentID:      assessment.ID,
				Email:             testEmail,
				AssessmentPreview: tc.preview,
			})
			assert.NilError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/questionnaire", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			recorder := httptest.NewRecorder()
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, recorder.Code, http.StatusOK)

			var out *models.GetQuestionnaireResponse
			err = json.NewDecoder(res.Body).Decode(&out)
			assert.NilError(t, err)
			assert.Assert(t, out != nil)
			assert.Equal(t, out.SavedData["q1"], tc.wantData)
		})
	}

	suite.db.AssessmentResponse.DeleteOneID(realResponse).Exec(ctx)
	suite.db.AssessmentResponse.DeleteOneID(testResponse).Exec(ctx)
	suite.db.Assessment.DeleteOneID(assessment.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}

// createResponseWithData creates an assessment response (test or real) and attaches document data
func (suite *HandlerTestSuite) createResponseWithData(createCtx, allowCtx context.Context, assessmentID, email, templateID string, isTest bool, marker string) string {
	t := suite.T()

	create := suite.db.AssessmentResponse.Create().
		SetAssessmentID(assessmentID).
		SetEmail(email).
		SetOwnerID(testUser1.OrganizationID)

	if isTest {
		create.SetIsTest(true)
	}

	response, err := create.Save(createCtx)
	assert.NilError(t, err)

	documentData, err := suite.db.DocumentData.Create().
		SetTemplateID(templateID).
		SetOwnerID(testUser1.OrganizationID).
		SetData(map[string]any{"q1": marker}).
		Save(allowCtx)
	assert.NilError(t, err)

	_, err = suite.db.AssessmentResponse.UpdateOneID(response.ID).
		SetDocumentDataID(documentData.ID).
		Save(allowCtx)
	assert.NilError(t, err)

	return response.ID
}
