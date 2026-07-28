package graphapi_test

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestAssessmentDueDateUpdateSyncsToResponses(t *testing.T) {
	assessment := (&AssessmentBuilder{
		client:              suite.client,
		ResponseDueDuration: int64(time.Hour.Seconds()),
	}).MustNew(sharedTestUser1.UserCtx, t)

	assessmentResponse := (&AssessmentResponseBuilder{
		client:       suite.client,
		AssessmentID: assessment.ID,
		OwnerID:      assessment.OwnerID,
	}).MustNew(sharedTestUser1.UserCtx, t)

	assert.Assert(t, !assessmentResponse.DueDate.IsZero())

	_, err := suite.client.api.UpdateAssessment(sharedTestUser1.UserCtx, assessment.ID, testclient.UpdateAssessmentInput{
		ResponseDueDuration: lo.ToPtr(int64((2 * time.Hour).Seconds())),
	})
	assert.NilError(t, err)

	resp, err := suite.client.api.GetAssessmentResponseByID(sharedTestUser1.UserCtx, assessmentResponse.ID)
	assert.NilError(t, err)
	assert.Assert(t, resp.AssessmentResponse.DueDate != nil)
	assert.Assert(t, resp.AssessmentResponse.DueDate.After(assessmentResponse.DueDate))

	(&Cleanup[*generated.AssessmentResponseDeleteOne]{client: suite.client.db.AssessmentResponse, ID: assessmentResponse.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.AssessmentDeleteOne]{client: suite.client.db.Assessment, ID: assessment.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.TemplateDeleteOne]{client: suite.client.db.Template, ID: assessment.TemplateID}).MustDelete(sharedTestUser1.UserCtx, t)
}
