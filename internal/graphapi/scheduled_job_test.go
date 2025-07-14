package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryScheduledJob(t *testing.T) {

	job := (&ScheduledJobBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	runner := (&JobRunnerBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	firstScheduledJob := (&ControlScheduledJobBuilder{
		client:        suite.client,
		JobID:         job.ID,
		Configuration: models.JobConfiguration{},
		JobRunnerID:   runner.ID,
	}).MustNew(testUser1.UserCtx, t)

	secondScheduledJob := (&ControlScheduledJobBuilder{
		client:        suite.client,
		JobID:         job.ID,
		Configuration: models.JobConfiguration{},
		JobRunnerID:   runner.ID,
	}).MustNew(testUser1.UserCtx, t)

	secondJob := (&ScheduledJobBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	thirdScheduledJob := (&ControlScheduledJobBuilder{
		client:        suite.client,
		JobID:         secondJob.ID,
		Configuration: models.JobConfiguration{},
		JobRunnerID:   runner.ID,
	}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name          string
		userID        string
		client        *openlaneclient.OpenlaneClient
		ctx           context.Context
		errorMsg      string
		expectedCount int
	}{
		{
			name:          "happy path user",
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			expectedCount: 2,
		},
		{
			name:          "happy path user with pat",
			client:        suite.client.apiWithPAT,
			ctx:           context.Background(),
			expectedCount: 2,
		},
		{
			name:          "happy path second user",
			client:        suite.client.api,
			ctx:           testUser2.UserCtx,
			expectedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {

			resp, err := tc.client.GetAllControlScheduledJobs(tc.ctx)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.ControlScheduledJobs.Edges, tc.expectedCount))
		})
	}

	(&Cleanup[*generated.JobRunnerDeleteOne]{
		client: suite.client.db.JobRunner,
		IDs:    []string{runner.ID},
	}).MustDelete(systemAdminUser.UserCtx, t)

	(&Cleanup[*generated.ControlScheduledJobDeleteOne]{
		client: suite.client.db.ControlScheduledJob,
		IDs:    []string{firstScheduledJob.ID, secondScheduledJob.ID},
	}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.ControlScheduledJobDeleteOne]{
		client: suite.client.db.ControlScheduledJob,
		IDs:    []string{thirdScheduledJob.ID},
	}).MustDelete(testUser2.UserCtx, t)
}

func TestControlScheduledJob(t *testing.T) {

	job := (&ScheduledJobBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	runner := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	control := (&ControlBuilder{client: suite.client, Name: "Test Control"}).MustNew(testUser1.UserCtx, t)

	subControl := (&SubcontrolBuilder{client: suite.client, ControlID: control.ID, Name: "Test Control"}).
		MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name       string
		ctx        context.Context
		client     *openlaneclient.OpenlaneClient
		jobBuilder ControlScheduledJobBuilder
		errorMsg   string
	}{
		{
			name:   "happy path - create scheduled job with runner",
			ctx:    testUser1.UserCtx,
			client: suite.client.api,
			jobBuilder: ControlScheduledJobBuilder{
				client:        suite.client,
				JobID:         job.ID,
				Configuration: models.JobConfiguration{},
				JobRunnerID:   runner.ID,
				ControlIDs:    []string{control.ID},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			job := tc.jobBuilder.MustNew(tc.ctx, t)

			assert.Equal(t, job.JobID, tc.jobBuilder.JobID)

			if tc.jobBuilder.JobRunnerID != "" {
				assert.Equal(t, job.JobRunnerID, tc.jobBuilder.JobRunnerID)
			}

			if tc.jobBuilder.Cron != nil {
				assert.Equal(t, *job.Cron, *tc.jobBuilder.Cron)
			}

			if len(tc.jobBuilder.ControlIDs) > 0 {
				controls, err := job.QueryControls().All(setContext(tc.ctx, suite.client.db))
				assert.NilError(t, err)
				assert.Equal(t, len(controls), len(tc.jobBuilder.ControlIDs))
			}

			(&Cleanup[*generated.ControlScheduledJobDeleteOne]{
				client: suite.client.db.ControlScheduledJob,
				ID:     job.ID,
			}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.JobRunnerDeleteOne]{
		client: suite.client.db.JobRunner,
		ID:     runner.ID,
	}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.SubcontrolDeleteOne]{
		client: suite.client.db.Subcontrol,
		ID:     subControl.ID,
	}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.ControlDeleteOne]{
		client: suite.client.db.Control,
		ID:     control.ID,
	}).MustDelete(testUser1.UserCtx, t)
}
