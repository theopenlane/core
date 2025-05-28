package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func TestQueryTask(t *testing.T) {
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: task.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: task.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     notFoundErrorMsg,
			queryID:  "notfound",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTaskByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.Task.ID))
			assert.Check(t, len(resp.Task.Title) != 0)
			assert.Check(t, resp.Task.Details != nil)
			assert.Check(t, len(resp.Task.Status) != 0)
		})
	}

	// cleanup
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: task.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTasks(t *testing.T) {
	// create a bunch to test the pagination with different users
	// works with overfetching
	numTasks := 10
	org1TaskIDs := []string{}
	org2TaskIDs := []string{}
	for range numTasks {
		t1 := (&TaskBuilder{client: suite.client, Due: gofakeit.Date()}).MustNew(testUser1.UserCtx, t)
		t2 := (&TaskBuilder{client: suite.client, Due: gofakeit.Date()}).MustNew(viewOnlyUser2.UserCtx, t)
		t3 := (&TaskBuilder{client: suite.client, Due: gofakeit.Date()}).MustNew(adminUser.UserCtx, t)
		org1TaskIDs = append(org1TaskIDs, t1.ID, t2.ID, t3.ID)

		t4 := (&TaskBuilder{client: suite.client, Due: gofakeit.Date()}).MustNew(testUser2.UserCtx, t)
		org2TaskIDs = append(org2TaskIDs, t4.ID)
	}

	userCtxPersonalOrg := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.PersonalOrgID)

	// add a task for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	taskPersonal := (&TaskBuilder{client: suite.client, AssigneeID: testUser1.ID}).MustNew(userCtxPersonalOrg, t)

	risk := (&RiskBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	taskWithRisk := (&TaskBuilder{client: suite.client, RiskID: risk.ID}).MustNew(testUser1.UserCtx, t)

	org1TaskIDs = append(org1TaskIDs, taskWithRisk.ID)

	var (
		startCursorDue     *string
		startCursorCreated *string
	)

	first := 10
	testCases := []struct {
		name            string
		orderBy         []*openlaneclient.TaskOrder
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
		setCursor       bool
		useCursor       bool
		totalCount      int64
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: first,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date, page 1",
			orderBy:         []*openlaneclient.TaskOrder{{Field: openlaneclient.TaskOrderFieldDue, Direction: openlaneclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*openlaneclient.TaskOrder{{Field: openlaneclient.TaskOrderFieldDue, Direction: openlaneclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*openlaneclient.TaskOrder{{Field: openlaneclient.TaskOrderFieldDue, Direction: openlaneclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by due date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*openlaneclient.TaskOrder{{Field: openlaneclient.TaskOrderFieldDue, Direction: openlaneclient.OrderDirectionDesc}},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 1,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date, page 1",
			orderBy:         []*openlaneclient.TaskOrder{{Field: openlaneclient.TaskOrderFieldCreatedAt, Direction: openlaneclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date and cursor, page 2",
			useCursor:       true,
			orderBy:         []*openlaneclient.TaskOrder{{Field: openlaneclient.TaskOrderFieldCreatedAt, Direction: openlaneclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date and cursor, page 3",
			useCursor:       true,
			orderBy:         []*openlaneclient.TaskOrder{{Field: openlaneclient.TaskOrderFieldCreatedAt, Direction: openlaneclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: first,
			setCursor:       true,
			totalCount:      31,
		},
		{
			name:            "happy path, with order by created date and cursor, page 4",
			useCursor:       true,
			orderBy:         []*openlaneclient.TaskOrder{{Field: openlaneclient.TaskOrderFieldCreatedAt, Direction: openlaneclient.OrderDirectionAsc}},
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 1,
			totalCount:      31,
		},
		{
			name:            "happy path, view only user",
			client:          suite.client.api,
			ctx:             viewOnlyUser2.UserCtx,
			expectedResults: first,
			totalCount:      10,
		},
		{
			name:            "happy path, admin user",
			client:          suite.client.api,
			ctx:             adminUser.UserCtx,
			expectedResults: first,
			totalCount:      11,
		},
		{
			name:            "happy path, using pat - which should have access to all tasks because its authorized to the personal org",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: first,
			totalCount:      32,
		},
		{
			name:            "another user, no entities should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: first,
			totalCount:      10,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			firstInput := int64(first)

			var after *string

			if tc.useCursor {
				if tc.orderBy[0].Field == openlaneclient.TaskOrderFieldDue {
					after = startCursorDue
				} else if tc.orderBy[0].Field == openlaneclient.TaskOrderFieldCreatedAt {
					after = startCursorCreated
				}
			}

			resp, err := tc.client.GetTasks(tc.ctx, &firstInput, nil, after, nil, nil, tc.orderBy)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Tasks.Edges, tc.expectedResults))
			assert.Check(t, is.Equal(tc.totalCount, resp.Tasks.TotalCount))

			if tc.setCursor {
				// set the start cursor for the next test case
				assert.Assert(t, resp.Tasks.PageInfo.HasNextPage)
				assert.Assert(t, resp.Tasks.PageInfo.EndCursor != nil)

				if tc.orderBy[0].Field == openlaneclient.TaskOrderFieldDue {
					startCursorDue = resp.Tasks.PageInfo.EndCursor
				} else if tc.orderBy[0].Field == openlaneclient.TaskOrderFieldCreatedAt {
					startCursorCreated = resp.Tasks.PageInfo.EndCursor
				}
			} else if tc.useCursor {
				// if we are using the cursor, but not setting it, we should not have a next page
				assert.Check(t, !(resp.Tasks.PageInfo.HasNextPage))

				// it should still have an end cursor
				assert.Check(t, resp.Tasks.PageInfo.EndCursor != nil)
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: taskPersonal.ID}).MustDelete(userCtxPersonalOrg, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: org1TaskIDs}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: org2TaskIDs}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationCreateTask(t *testing.T) {
	om := (&OrgMemberBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	userCtx := auth.NewTestContextWithOrgID(om.UserID, om.OrganizationID)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateTaskInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateTaskInput{
				Title: "test-task",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateTaskInput{
				Title:      "test-task",
				Details:    lo.ToPtr("test details of the task"),
				Status:     &enums.TaskStatusInProgress,
				Category:   lo.ToPtr("evidence upload"),
				Due:        lo.ToPtr(models.DateTime(time.Now().Add(time.Hour * 24))),
				AssigneeID: &om.UserID, // assign the task to another user
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "create with assignee not in org should fail",
			request: openlaneclient.CreateTaskInput{
				Title:      "test-task",
				AssigneeID: &testUser2.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "user not in organization",
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateTaskInput{
				Title:   "test-task",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateTaskInput{
				Title: "test-task",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "missing title, but display name provided",
			request: openlaneclient.CreateTaskInput{
				Details: lo.ToPtr("makin' a list, checkin' it twice"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTask(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.Title, resp.CreateTask.Task.Title))

			assert.Check(t, len(resp.CreateTask.Task.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.CreateTask.Task.DisplayID, "TSK-"))

			assert.Check(t, resp.CreateTask.Task.OwnerID != nil)

			if tc.request.Details == nil {
				assert.Check(t, is.Equal(*resp.CreateTask.Task.Details, ""))
			} else {
				assert.Check(t, is.Equal(*tc.request.Details, *resp.CreateTask.Task.Details))
			}

			if tc.request.Status == nil {
				assert.Check(t, is.Equal(enums.TaskStatusOpen, resp.CreateTask.Task.Status))
			} else {
				assert.Check(t, is.Equal(*tc.request.Status, resp.CreateTask.Task.Status))
			}

			if tc.request.Details == nil {
				assert.Check(t, is.Equal(*resp.CreateTask.Task.Details, ""))
			} else {
				assert.Check(t, is.Equal(*tc.request.Details, *resp.CreateTask.Task.Details))
			}

			if tc.request.Category == nil {
				assert.Check(t, is.Equal(*resp.CreateTask.Task.Category, ""))
			} else {
				assert.Check(t, is.Equal(*tc.request.Category, *resp.CreateTask.Task.Category))
			}

			if tc.request.Due == nil {
				assert.Check(t, resp.CreateTask.Task.Due == nil)
			} else {
				assert.Assert(t, resp.CreateTask.Task.Due != nil)
				diff := time.Time(*resp.CreateTask.Task.Due).Sub(time.Time(*tc.request.Due))
				assert.Check(t, diff >= -10*time.Second && diff <= 10*time.Second, "time difference is not within 10 seconds")
			}

			// when using an API token, the assigner is not set
			if tc.client == suite.client.apiWithToken {
				assert.Check(t, is.Nil(resp.CreateTask.Task.Assigner))
			} else {
				// otherwise it defaults to the authorized user
				assert.Check(t, resp.CreateTask.Task.Assigner != nil)
				assert.Check(t, is.Equal(testUser1.ID, resp.CreateTask.Task.Assigner.ID))
			}

			if tc.request.AssigneeID == nil {
				assert.Check(t, is.Nil(resp.CreateTask.Task.Assignee))
			} else {
				assert.Assert(t, resp.CreateTask.Task.Assignee != nil)
				assert.Check(t, is.Equal(*tc.request.AssigneeID, resp.CreateTask.Task.Assignee.ID))

				// make sure the assignee can see the task
				taskResp, err := suite.client.api.GetTaskByID(userCtx, resp.CreateTask.Task.ID)
				assert.NilError(t, err)
				assert.Check(t, taskResp != nil)

				// make sure the another org member cannot see the task
				taskResp, err = suite.client.api.GetTaskByID(adminUser.UserCtx, resp.CreateTask.Task.ID)

				assert.Check(t, is.Nil(taskResp))
			}

			// cleanup
			(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: resp.CreateTask.Task.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// cleanup
	(&Cleanup[*generated.OrgMembershipDeleteOne]{client: suite.client.db.OrgMembership, ID: om.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateTask(t *testing.T) {
	task := (&TaskBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	group := (&GroupBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	pngFile, err := objects.NewUploadFile("testdata/uploads/logo.png")
	assert.NilError(t, err)

	pdfFile, err := objects.NewUploadFile("testdata/uploads/hello.pdf")
	assert.NilError(t, err)

	taskCommentID := ""

	assignee := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &assignee, enums.RoleMember, testUser1.OrganizationID)

	// add parents to ensure permissions are inherited
	risk := (&RiskBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	taskRisk := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// make sure the user cannot can see the task before they are the assigner
	taskResp, err := suite.client.api.GetTaskByID(viewOnlyUser2.UserCtx, task.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)
	assert.Check(t, is.Nil(taskResp))

	// make sure the user cannot can see the task before they are the assignee
	taskResp, err = suite.client.api.GetTaskByID(assignee.UserCtx, task.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)
	assert.Check(t, is.Nil(taskResp))

	// make sure the user cannot see the task before the risk is added
	taskResp, err = suite.client.api.GetTaskByID(adminUser.UserCtx, taskRisk.ID)
	assert.ErrorContains(t, err, notFoundErrorMsg)
	assert.Check(t, is.Nil(taskResp))

	// NOTE: the tests and checks are ordered due to dependencies between updates
	// if you update cases, they will most likely need to be added to the end of the list
	testCases := []struct {
		name                 string
		taskID               string
		request              *openlaneclient.UpdateTaskInput
		updateCommentRequest *openlaneclient.UpdateNoteInput
		files                []*graphql.Upload
		client               *openlaneclient.OpenlaneClient
		ctx                  context.Context
		expectedErr          string
	}{
		{
			name:   "happy path, update details",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				Details:    lo.ToPtr(("makin' a list, checkin' it twice")),
				AssigneeID: &adminUser.ID,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path, add comment",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				AddComment: &openlaneclient.CreateNoteInput{
					Text: "matt is the best",
				},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path, update comment with files",
			taskID: task.ID,
			updateCommentRequest: &openlaneclient.UpdateNoteInput{
				Text: lo.ToPtr("sarah is better"),
			},
			files: []*graphql.Upload{
				{
					File:        pngFile.File,
					Filename:    pngFile.Filename,
					Size:        pngFile.Size,
					ContentType: pngFile.ContentType,
				},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path, update comment with file using PAT",
			taskID: task.ID,
			updateCommentRequest: &openlaneclient.UpdateNoteInput{
				Text: lo.ToPtr("sarah is still better"),
			},
			files: []*graphql.Upload{
				{
					File:        pdfFile.File,
					Filename:    pdfFile.Filename,
					Size:        pdfFile.Size,
					ContentType: pdfFile.ContentType,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:   "happy path, delete comment",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				DeleteComment: &taskCommentID,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "happy path, add risk",
			taskID: taskRisk.ID,
			request: &openlaneclient.UpdateTaskInput{
				AddRiskIDs: []string{risk.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:   "update category using pat of owner",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				Category: lo.ToPtr("risk review"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:   "update assignee to user not in org should fail",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(testUser2.ID),
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: "user not in organization",
		},
		{
			name:   "update assignee to view only user",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(assignee.ID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "update assignee to same user, should not error",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(assignee.ID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "update status and details",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				Status:  &enums.TaskStatusInProgress,
				Details: lo.ToPtr("do all the things for the thing"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "add to group",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				AddGroupIDs: []string{group.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "update assigner to another org member, this user should still be able to see it because they originally created it",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				AssignerID: lo.ToPtr(viewOnlyUser2.ID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "clear assignee",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				ClearAssignee: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name:   "clear assigner",
			taskID: task.ID,
			request: &openlaneclient.UpdateTaskInput{
				ClearAssigner: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			var (
				err         error
				resp        *openlaneclient.UpdateTask
				commentResp *openlaneclient.UpdateTaskComment
			)

			if tc.request != nil {
				resp, err = tc.client.UpdateTask(tc.ctx, tc.taskID, *tc.request)
			} else if tc.updateCommentRequest != nil {
				if len(tc.files) > 0 {
					expectUploadNillable(t, suite.client.objectStore.Storage, tc.files)
				}

				commentResp, err = suite.client.api.UpdateTaskComment(testUser1.UserCtx, taskCommentID, *tc.updateCommentRequest, tc.files)
			}

			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)

			if tc.request != nil {
				assert.Assert(t, resp != nil)

				if tc.request.Details != nil {
					assert.Check(t, is.Equal(*tc.request.Details, *resp.UpdateTask.Task.Details))
				}

				if tc.request.Status != nil {
					assert.Check(t, is.Equal(*tc.request.Status, resp.UpdateTask.Task.Status))
				}

				if tc.request.Details != nil {
					assert.Check(t, is.DeepEqual(tc.request.Details, resp.UpdateTask.Task.Details))
				}

				if tc.request.Category != nil {
					assert.Check(t, is.Equal(*tc.request.Category, *resp.UpdateTask.Task.Category))
				}

				if tc.request.ClearAssignee != nil {
					assert.Check(t, is.Nil(resp.UpdateTask.Task.Assignee))

					// the previous assignee should no longer be able to see the task
					taskResp, err := suite.client.api.GetTaskByID(assignee.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, is.ErrorContains(err, notFoundErrorMsg))
					assert.Check(t, is.Nil(taskResp))
				}

				if tc.request.ClearAssigner != nil {
					assert.Check(t, is.Nil(resp.UpdateTask.Task.Assignee))

					// the previous assigner should no longer be able to see the task
					taskResp, err := suite.client.api.GetTaskByID(viewOnlyUser2.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, is.ErrorContains(err, notFoundErrorMsg))
					assert.Check(t, is.Nil(taskResp))
				}

				if tc.request.AddRiskIDs != nil {
					taskResp, err := suite.client.api.GetTaskByID(adminUser.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, is.Nil(err))
					assert.Check(t, is.Equal(taskResp.Task.ID, tc.taskID))
				}

				if tc.request.AssignerID != nil {
					assert.Check(t, resp.UpdateTask.Task.Assigner != nil)
					assert.Check(t, is.Equal(*tc.request.AssignerID, resp.UpdateTask.Task.Assigner.ID))

					// make sure the assigner can see the task
					taskResp, err := suite.client.api.GetTaskByID(viewOnlyUser2.UserCtx, resp.UpdateTask.Task.ID)
					assert.Check(t, err)
					assert.Check(t, taskResp != nil)

					// make sure the original creator can still see the task
					taskResp, err = suite.client.api.GetTaskByID(adminUser.UserCtx, resp.UpdateTask.Task.ID)
					assert.NilError(t, err)
					assert.Check(t, taskResp != nil)
				}

				if tc.request.AddComment != nil {
					assert.Check(t, len(resp.UpdateTask.Task.Comments.Edges) != 0)
					assert.Check(t, is.Equal(tc.request.AddComment.Text, resp.UpdateTask.Task.Comments.Edges[0].Node.Text))

					// there should only be one comment
					assert.Assert(t, is.Len(resp.UpdateTask.Task.Comments.Edges, 1))
					taskCommentID = resp.UpdateTask.Task.Comments.Edges[0].Node.ID

					// user shouldn't be able to see the comment
					checkResp, err := suite.client.api.GetNoteByID(assignee.UserCtx, taskCommentID)
					assert.Check(t, is.ErrorContains(err, notFoundErrorMsg))
					assert.Check(t, is.Nil(checkResp))

					// user should be able to see the comment since they created the task
					checkResp, err = suite.client.api.GetNoteByID(adminUser.UserCtx, taskCommentID)
					assert.Check(t, err)
					assert.Check(t, checkResp != nil)

					// org owner should be able to see the comment
					checkResp, err = suite.client.api.GetNoteByID(testUser1.UserCtx, taskCommentID)
					assert.Check(t, err)
					assert.Check(t, checkResp != nil)
				} else if tc.request.DeleteComment != nil {
					// should not have any comments
					assert.Check(t, is.Len(resp.UpdateTask.Task.Comments.Edges, 0))
				}
			} else if tc.updateCommentRequest != nil {
				assert.Assert(t, commentResp != nil)

				// should only have the original comment
				assert.Assert(t, is.Len(commentResp.UpdateTaskComment.Task.Comments.Edges, 1))
				assert.Check(t, is.Equal(*tc.updateCommentRequest.Text, commentResp.UpdateTaskComment.Task.Comments.Edges[0].Node.Text))
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: task.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: group.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteTask(t *testing.T) {
	task1 := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	task2 := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete task",
			idToDelete:  task1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete task",
			idToDelete: task1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "task already deleted, not found",
			idToDelete:  task1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "task not found",
		},
		{
			name:       "happy path, delete task using personal access token",
			idToDelete: task2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown task, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTask(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTask.DeletedID))
		})
	}
}
