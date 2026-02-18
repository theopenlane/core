package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

func (suite *HandlerTestSuite) TestResendQuestionnaireEmail() {
	t := suite.T()

	operation := suite.createImpersonationOperation("ResendQuestionnaireEmail", "Resend questionnaire authentication email")
	suite.registerTestHandler("POST", "/questionnaire/resend", operation, suite.h.ResendQuestionnaireEmail)

	ec := echocontext.NewTestEchoContext().Request().Context()
	ctx := privacy.DecisionContext(ec, privacy.Allow)

	questionnaireCtx := contextx.With(ctx, auth.QuestionnaireContextKey{})
	questionnaireCtx = auth.WithAuthenticatedUser(questionnaireCtx, &auth.AuthenticatedUser{
		SubjectID:      testUser1.ID,
		OrganizationID: testUser1.OrganizationID,
	})

	template, err := suite.db.Template.Create().
		SetName("Resend Test Template").
		SetTemplateType(enums.Document).
		SetJsonconfig(map[string]any{"title": "Resend Test"}).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	assessmentObj, err := suite.db.Assessment.Create().
		SetName("Resend Test Assessment").
		SetTemplateID(template.ID).
		SetAssessmentType(enums.AssessmentTypeExternal).
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)

	send := func(t *testing.T, email, assessmentID string) (*httptest.ResponseRecorder, map[string]interface{}) {
		reqBody := models.ResendQuestionnaireRequest{
			Email:        email,
			AssessmentID: assessmentID,
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/questionnaire/resend", strings.NewReader(string(body)))
		req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

		recorder := httptest.NewRecorder()

		suite.e.ServeHTTP(recorder, req)

		res := recorder.Result()
		defer res.Body.Close()

		var out map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
			t.Error("error parsing response", err)
		}

		return recorder, out
	}

	t.Run("successful resend", func(t *testing.T) {
		suite.ClearTestData()

		testEmail := "resend-success@example.com"

		assessmentResp, err := suite.db.AssessmentResponse.Create().
			SetAssessmentID(assessmentObj.ID).
			SetEmail(testEmail).
			SetOwnerID(testUser1.OrganizationID).
			SetStatus(enums.AssessmentResponseStatusSent).
			Save(questionnaireCtx)
		require.NoError(t, err)

		rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
			[]rivertest.ExpectedJob{{Args: jobs.EmailArgs{}}})

		suite.ClearTestData()

		recorder, out := send(t, testEmail, assessmentObj.ID)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, true, out["success"])
		assert.NotEmpty(t, out["message"])

		rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
			[]rivertest.ExpectedJob{{Args: jobs.EmailArgs{}}})

		suite.db.AssessmentResponse.DeleteOneID(assessmentResp.ID).Exec(ctx)
	})

	t.Run("email not found returns 200", func(t *testing.T) {
		recorder, out := send(t, "nonexistent@example.com", assessmentObj.ID)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, true, out["success"])
	})

	t.Run("wrong assessment id returns 200", func(t *testing.T) {
		recorder, out := send(t, "resend-success@example.com", "01NONEXISTENT000000000000")

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, true, out["success"])
	})

	t.Run("missing email returns 400", func(t *testing.T) {
		recorder, _ := send(t, "", assessmentObj.ID)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("missing assessment_id returns 400", func(t *testing.T) {
		recorder, _ := send(t, "test@example.com", "")

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("completed assessment returns 200 without resend", func(t *testing.T) {
		suite.ClearTestData()

		completedEmail := "completed-resend@example.com"

		completedResp, err := suite.db.AssessmentResponse.Create().
			SetAssessmentID(assessmentObj.ID).
			SetEmail(completedEmail).
			SetOwnerID(testUser1.OrganizationID).
			Save(questionnaireCtx)
		require.NoError(t, err)

		rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
			[]rivertest.ExpectedJob{{Args: jobs.EmailArgs{}}})

		completedResp, err = suite.db.AssessmentResponse.UpdateOneID(completedResp.ID).
			SetStatus(enums.AssessmentResponseStatusCompleted).
			SetCompletedAt(time.Now().Add(-1 * time.Hour)).
			Save(questionnaireCtx)
		require.NoError(t, err)

		suite.ClearTestData()

		recorder, out := send(t, completedEmail, assessmentObj.ID)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, true, out["success"])

		suite.db.AssessmentResponse.DeleteOneID(completedResp.ID).Exec(ctx)
	})

	t.Run("overdue assessment returns 400", func(t *testing.T) {
		suite.ClearTestData()

		overdueEmail := "overdue-resend@example.com"

		overdueResp, err := suite.db.AssessmentResponse.Create().
			SetAssessmentID(assessmentObj.ID).
			SetEmail(overdueEmail).
			SetOwnerID(testUser1.OrganizationID).
			SetStatus(enums.AssessmentResponseStatusSent).
			SetDueDate(time.Now().Add(24 * time.Hour)).
			Save(questionnaireCtx)
		require.NoError(t, err)

		rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
			[]rivertest.ExpectedJob{{Args: jobs.EmailArgs{}}})

		overdueResp, err = suite.db.AssessmentResponse.UpdateOneID(overdueResp.ID).
			SetDueDate(time.Now().Add(-24 * time.Hour)).
			Save(questionnaireCtx)
		require.NoError(t, err)

		recorder, out := send(t, overdueEmail, assessmentObj.ID)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		if errMsg, ok := out["error"].(string); ok {
			assert.Contains(t, errMsg, "overdue")
		}

		suite.db.AssessmentResponse.DeleteOneID(overdueResp.ID).Exec(ctx)
	})

	t.Run("max attempts returns 429", func(t *testing.T) {
		suite.ClearTestData()

		maxEmail := "maxattempts-resend@example.com"

		maxResp, err := suite.db.AssessmentResponse.Create().
			SetAssessmentID(assessmentObj.ID).
			SetEmail(maxEmail).
			SetOwnerID(testUser1.OrganizationID).
			SetStatus(enums.AssessmentResponseStatusSent).
			SetSendAttempts(5).
			Save(questionnaireCtx)
		require.NoError(t, err)

		rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
			[]rivertest.ExpectedJob{{Args: jobs.EmailArgs{}}})

		suite.ClearTestData()

		recorder, out := send(t, maxEmail, assessmentObj.ID)

		assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
		if errMsg, ok := out["error"].(string); ok {
			assert.Contains(t, errMsg, "max attempts")
		}

		suite.db.AssessmentResponse.DeleteOneID(maxResp.ID).Exec(ctx)
	})

	// cleanup
	suite.db.Assessment.DeleteOneID(assessmentObj.ID).Exec(ctx)
	suite.db.Template.DeleteOneID(template.ID).Exec(ctx)
}
