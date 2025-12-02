package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/shared/models"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestQueryScheduledJob(t *testing.T) {
	job := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	runner := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	firstScheduledJob := (&ScheduledJobBuilder{
		client:        suite.client,
		JobID:         job.ID,
		Configuration: models.JobConfiguration{},
		JobRunnerID:   runner.ID,
	}).MustNew(testUser1.UserCtx, t)

	secondScheduledJob := (&ScheduledJobBuilder{
		client:        suite.client,
		JobID:         job.ID,
		Configuration: models.JobConfiguration{},
		JobRunnerID:   runner.ID,
	}).MustNew(testUser1.UserCtx, t)

	secondJob := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	runner2 := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	thirdScheduledJob := (&ScheduledJobBuilder{
		client:        suite.client,
		JobID:         secondJob.ID,
		Configuration: models.JobConfiguration{},
		JobRunnerID:   runner2.ID,
	}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name          string
		userID        string
		client        *testclient.TestClient
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
			name:          "happy path admin user",
			client:        suite.client.api,
			ctx:           testUser1.UserCtx,
			expectedCount: 2,
		},
		{
			name:          "happy path view only user",
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
			resp, err := tc.client.GetAllScheduledJobs(tc.ctx)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.ScheduledJobs.Edges, tc.expectedCount))
		})
	}

	(&Cleanup[*generated.JobRunnerDeleteOne]{
		client: suite.client.db.JobRunner,
		IDs:    []string{runner.ID},
	}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.ScheduledJobDeleteOne]{
		client: suite.client.db.ScheduledJob,
		IDs:    []string{firstScheduledJob.ID, secondScheduledJob.ID},
	}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.ScheduledJobDeleteOne]{
		client: suite.client.db.ScheduledJob,
		IDs:    []string{thirdScheduledJob.ID},
	}).MustDelete(testUser2.UserCtx, t)

	(&Cleanup[*generated.JobRunnerDeleteOne]{
		client: suite.client.db.JobRunner,
		IDs:    []string{runner2.ID},
	}).MustDelete(testUser2.UserCtx, t)

	(&Cleanup[*generated.JobTemplateDeleteOne]{
		client: suite.client.db.JobTemplate,
		IDs:    []string{job.ID},
	}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.JobTemplateDeleteOne]{
		client: suite.client.db.JobTemplate,
		IDs:    []string{secondJob.ID},
	}).MustDelete(testUser2.UserCtx, t)
}

func TestScheduledJobs(t *testing.T) {
	job := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	runner := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client, RefCode: "Test Control"}).MustNew(testUser1.UserCtx, t)
	subControl := (&SubcontrolBuilder{client: suite.client, ControlID: control.ID, Name: "Test Control"}).
		MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name       string
		ctx        context.Context
		client     *testclient.TestClient
		jobBuilder ScheduledJobBuilder
		errorMsg   string
	}{
		{
			name:   "happy path - create scheduled job with runner",
			ctx:    testUser1.UserCtx,
			client: suite.client.api,
			jobBuilder: ScheduledJobBuilder{
				client:        suite.client,
				JobID:         job.ID,
				Configuration: models.JobConfiguration{},
				JobRunnerID:   runner.ID,
				ControlIDs:    []string{control.ID},
			},
		},
		{
			name:   "happy path - create scheduled job with runner by admin",
			ctx:    adminUser.UserCtx,
			client: suite.client.api,
			jobBuilder: ScheduledJobBuilder{
				client:        suite.client,
				JobID:         job.ID,
				Configuration: models.JobConfiguration{},
				JobRunnerID:   runner.ID,
				ControlIDs:    []string{control.ID},
			},
		},
		{
			name:   "create scheduled job with runner by view only user should fail",
			ctx:    viewOnlyUser.UserCtx,
			client: suite.client.api,
			jobBuilder: ScheduledJobBuilder{
				client:        suite.client,
				JobID:         job.ID,
				Configuration: models.JobConfiguration{},
				ControlIDs:    []string{control.ID},
			},
			errorMsg: notAuthorizedErrorMsg,
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

			(&Cleanup[*generated.ScheduledJobDeleteOne]{
				client: suite.client.db.ScheduledJob,
				ID:     job.ID,
			}).MustDelete(tc.ctx, t)
		})
	}

	(&Cleanup[*generated.JobRunnerDeleteOne]{
		client: suite.client.db.JobRunner,
		ID:     runner.ID,
	}).MustDelete(testUser1.UserCtx, t)

	(&Cleanup[*generated.JobTemplateDeleteOne]{
		client: suite.client.db.JobTemplate,
		ID:     job.ID,
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

func TestMutationCreateScheduledJob(t *testing.T) {
	job := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	// jobSystemOwned := (&JobTemplateBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)

	runner := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client, RefCode: "Test Control"}).MustNew(testUser1.UserCtx, t)
	subControl := (&SubcontrolBuilder{client: suite.client, ControlID: control.ID, Name: "Test Control"}).
		MustNew(testUser1.UserCtx, t)

	job2 := (&JobTemplateBuilder{client: suite.client, Cron: "0 0 0 * * *"}).MustNew(testUser2.UserCtx, t)

	cron := "0 0 0 * * *"
	invalidCron := "0 0 * * *" // invalid cron syntax (requires 5 parts), should fail validation

	testCases := []struct {
		name        string
		request     testclient.CreateScheduledJobInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input, cron inherited from job template",
			request: testclient.CreateScheduledJobInput{
				JobTemplateID: job.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		// TODO: see comment on schema, public tuples need to be implemented for this to work
		// {
		// 	name: "happy path, minimal input, cron inherited from job template, system owned job",
		// 	request: testclient.CreateScheduledJobInput{
		// 		JobTemplateID: jobSystemOwned.ID,
		// 	},
		// 	client: suite.client.api,
		// 	ctx:    testUser1.UserCtx,
		// },
		{
			name: "happy path, all input",
			request: testclient.CreateScheduledJobInput{
				JobTemplateID: job.ID,
				Cron:          &cron,
				JobRunnerID:   &runner.ID,
				ControlIDs:    []string{control.ID},
				SubcontrolIDs: []string{subControl.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateScheduledJobInput{
				JobTemplateID: job.ID,
				Cron:          &cron,
				OwnerID:       &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateScheduledJobInput{
				JobTemplateID: job.ID,
				Cron:          &cron,
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateScheduledJobInput{
				JobTemplateID: job.ID,
				Cron:          &cron,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user not authorized, job not in organization",
			request: testclient.CreateScheduledJobInput{
				JobTemplateID: job.ID,
				Cron:          &cron,
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user not authorized, job runner not in organization",
			request: testclient.CreateScheduledJobInput{
				JobTemplateID: job2.ID,
				Cron:          &cron,
				JobRunnerID:   &runner.ID,
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "missing required field, job template id",
			request: testclient.CreateScheduledJobInput{
				Cron: &cron,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "invalid input, cron",
			request: testclient.CreateScheduledJobInput{
				JobTemplateID: job.ID,
				Cron:          &invalidCron,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "invalid cron syntax",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateScheduledJob(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, is.Equal(resp.CreateScheduledJob.ScheduledJob.JobID, tc.request.JobTemplateID))

			if tc.request.Cron != nil {
				assert.Check(t, is.Equal(*resp.CreateScheduledJob.ScheduledJob.Cron, *tc.request.Cron))
			} else {
				// should fall back to the job template cron if it has one
				if resp.CreateScheduledJob.ScheduledJob.Cron != nil {
					assert.Check(t, is.Equal(*resp.CreateScheduledJob.ScheduledJob.Cron, job.Cron.String()))
				}
			}

			// check optional fields with if checks if they were provided
			if tc.request.JobRunnerID != nil {
				assert.Check(t, is.Equal(*resp.CreateScheduledJob.ScheduledJob.JobRunnerID, *tc.request.JobRunnerID))
			} else {
				assert.Check(t, *resp.CreateScheduledJob.ScheduledJob.JobRunnerID == "")
			}

			if len(tc.request.ControlIDs) > 0 {
				assert.Check(t, is.Equal(len(resp.CreateScheduledJob.ScheduledJob.Controls.Edges), len(tc.request.ControlIDs)))
			} else {
				assert.Check(t, is.Equal(len(resp.CreateScheduledJob.ScheduledJob.Controls.Edges), 0))
			}

			if len(tc.request.SubcontrolIDs) > 0 {
				assert.Check(t, is.Equal(len(resp.CreateScheduledJob.ScheduledJob.Subcontrols.Edges), len(tc.request.SubcontrolIDs)))
			} else {
				assert.Check(t, is.Equal(len(resp.CreateScheduledJob.ScheduledJob.Subcontrols.Edges), 0))
			}

			// cleanup each ScheduledJob created
			(&Cleanup[*generated.ScheduledJobDeleteOne]{client: suite.client.db.ScheduledJob, ID: resp.CreateScheduledJob.ScheduledJob.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// cleanup each JobTemplate created
	(&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, ID: job.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, ID: job2.ID}).MustDelete(testUser2.UserCtx, t)
	// (&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, ID: jobSystemOwned.ID}).MustDelete(systemAdminUser.UserCtx, t)

	// cleanup each JobRunner created
	(&Cleanup[*generated.JobRunnerDeleteOne]{client: suite.client.db.JobRunner, ID: runner.ID}).MustDelete(testUser1.UserCtx, t)

	// cleanup each Control created
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)

	// cleanup each Subcontrol created
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, ID: subControl.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateScheduledJob(t *testing.T) {
	job := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control1 := (&ControlBuilder{client: suite.client, RefCode: "TC-1"}).MustNew(testUser1.UserCtx, t)

	// ensure we can create two scheduled jobs with the same job template id
	scheduledJob := (&ScheduledJobBuilder{client: suite.client, JobID: job.ID, ControlIDs: []string{control1.ID}}).MustNew(testUser1.UserCtx, t)
	scheduledJob2 := (&ScheduledJobBuilder{client: suite.client, JobID: job.ID}).MustNew(testUser1.UserCtx, t)

	runner := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	anotherRunner := (&JobRunnerBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, RefCode: "TC-2"}).MustNew(testUser1.UserCtx, t)
	subControl := (&SubcontrolBuilder{client: suite.client, ControlID: control2.ID, Name: "SCT-1"}).
		MustNew(testUser1.UserCtx, t)

	newCron := "1 1 0 * * *"
	anotherCron := "0 0 1 * * *"
	invalidCron := "0 0 0 * *"
	testCases := []struct {
		name           string
		request        testclient.UpdateScheduledJobInput
		scheduledJobID string
		client         *testclient.TestClient
		ctx            context.Context
		expectedErr    string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateScheduledJobInput{
				Cron: &newCron,
			},
			scheduledJobID: scheduledJob.ID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateScheduledJobInput{
				AddControlIDs:    []string{control2.ID},
				AddSubcontrolIDs: []string{subControl.ID},
				JobRunnerID:      &runner.ID,
				JobTemplateID:    &job.ID,
			},
			scheduledJobID: scheduledJob.ID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
		},
		{
			name: "happy path, update multiple fields with pat",
			request: testclient.UpdateScheduledJobInput{
				Cron: &anotherCron,
			},
			scheduledJobID: scheduledJob2.ID,
			client:         suite.client.apiWithPAT,
			ctx:            context.Background(),
		},
		{
			name: "happy path, update multiple fields with api token",
			request: testclient.UpdateScheduledJobInput{
				JobRunnerID: &anotherRunner.ID,
			},
			scheduledJobID: scheduledJob.ID,
			client:         suite.client.apiWithToken,
			ctx:            context.Background(),
		},
		{
			name: "update not allowed, not enough permissions",
			request: testclient.UpdateScheduledJobInput{
				Cron: &newCron,
			},
			scheduledJobID: scheduledJob.ID,
			client:         suite.client.api,
			ctx:            viewOnlyUser.UserCtx,
			expectedErr:    notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateScheduledJobInput{
				Cron: &newCron,
			},
			scheduledJobID: scheduledJob.ID,
			client:         suite.client.api,
			ctx:            testUser2.UserCtx,
			expectedErr:    notFoundErrorMsg,
		},
		{
			name: "invalid input, cron",
			request: testclient.UpdateScheduledJobInput{
				Cron: &invalidCron,
			},
			scheduledJobID: scheduledJob.ID,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			expectedErr:    "invalid cron syntax",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateScheduledJob(tc.ctx, tc.scheduledJobID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// add checks for the updated fields if they were set in the request
			if tc.request.Cron != nil {
				assert.Check(t, is.Equal(*resp.UpdateScheduledJob.ScheduledJob.Cron, *tc.request.Cron))
			}

			if tc.request.JobRunnerID != nil {
				assert.Check(t, is.Equal(*resp.UpdateScheduledJob.ScheduledJob.JobRunnerID, *tc.request.JobRunnerID))
			}

			if len(tc.request.AddControlIDs) > 0 {
				expectedCount := len(tc.request.AddControlIDs) + 1 // add one because the scheduled job was created with a control already
				assert.Check(t, is.Equal(len(resp.UpdateScheduledJob.ScheduledJob.Controls.Edges), expectedCount))
			}

			if len(tc.request.AddSubcontrolIDs) > 0 {
				assert.Check(t, is.Equal(len(resp.UpdateScheduledJob.ScheduledJob.Subcontrols.Edges), len(tc.request.AddSubcontrolIDs)))
			}

			if tc.request.JobTemplateID != nil {
				assert.Check(t, is.Equal(resp.UpdateScheduledJob.ScheduledJob.JobID, *tc.request.JobTemplateID))
			}
		})
	}

	(&Cleanup[*generated.ScheduledJobDeleteOne]{client: suite.client.db.ScheduledJob, ID: scheduledJob.ID}).MustDelete(testUser1.UserCtx, t)

	// cleanup each JobTemplate created
	(&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, ID: job.ID}).MustDelete(testUser1.UserCtx, t)

	// cleanup each JobRunner created
	(&Cleanup[*generated.JobRunnerDeleteOne]{client: suite.client.db.JobRunner, ID: runner.ID}).MustDelete(testUser1.UserCtx, t)

	// cleanup each Control created
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID}}).MustDelete(testUser1.UserCtx, t)

	// cleanup each Subcontrol created
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, ID: subControl.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteScheduledJob(t *testing.T) {
	// create scheduled jobs to be deleted
	job := (&JobTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	scheduledJob1 := (&ScheduledJobBuilder{client: suite.client, JobID: job.ID}).MustNew(testUser1.UserCtx, t)
	scheduledJob2 := (&ScheduledJobBuilder{client: suite.client, JobID: job.ID}).MustNew(testUser1.UserCtx, t)
	scheduledJob3 := (&ScheduledJobBuilder{client: suite.client, JobID: job.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not found, delete",
			idToDelete:  scheduledJob1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  scheduledJob1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: scheduledJob1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  scheduledJob1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: scheduledJob2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: scheduledJob3.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteScheduledJob(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteScheduledJob.DeletedID))
		})
	}

	// cleanup each JobTemplate created
	(&Cleanup[*generated.JobTemplateDeleteOne]{client: suite.client.db.JobTemplate, ID: job.ID}).MustDelete(testUser1.UserCtx, t)
}
