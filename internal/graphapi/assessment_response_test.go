package graphapi_test

import (
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestAssessmentDueDateUpdateSyncsToResponses(t *testing.T) {
	assessment := (&th.AssessmentBuilder{
		Client:              suite.Client,
		ResponseDueDuration: int64(time.Hour.Seconds()),
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	assessmentResponse := (&th.AssessmentResponseBuilder{
		Client:       suite.Client,
		AssessmentID: assessment.ID,
		OwnerID:      assessment.OwnerID,
	}).MustNew(th.SharedTestUser1.UserCtx, t)

	assert.Assert(t, !assessmentResponse.DueDate.IsZero())

	_, err := suite.Client.API.UpdateAssessment(th.SharedTestUser1.UserCtx, assessment.ID, testclient.UpdateAssessmentInput{
		ResponseDueDuration: lo.ToPtr(int64((2 * time.Hour).Seconds())),
	})
	assert.NilError(t, err)

	resp, err := suite.Client.API.GetAssessmentResponseByID(th.SharedTestUser1.UserCtx, assessmentResponse.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp.AssessmentResponse.DueDate != nil)
	assert.Assert(t, resp.AssessmentResponse.DueDate.After(assessmentResponse.DueDate))

	(&th.Cleanup[*generated.AssessmentResponseDeleteOne]{Client: suite.Client.DB.AssessmentResponse, ID: assessmentResponse.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.AssessmentDeleteOne]{Client: suite.Client.DB.Assessment, ID: assessment.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TemplateDeleteOne]{Client: suite.Client.DB.Template, ID: assessment.TemplateID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
