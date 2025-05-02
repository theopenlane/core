package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/core/pkg/testutils"
)

func (suite *GraphTestSuite) TestQueryTask() {
	t := suite.T()

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.queryID, resp.Task.ID)
			assert.NotEmpty(t, resp.Task.Title)
			assert.NotEmpty(t, resp.Task.Details)
			assert.NotEmpty(t, resp.Task.Status)
		})
	}
}

func (suite *GraphTestSuite) TestQueryTasks() {
	t := suite.T()

	// create a bunch to test the pagination with different users
	// works with overfetching
	numTasks := 10
	for range numTasks {
		(&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
		(&TaskBuilder{client: suite.client}).MustNew(viewOnlyUser.UserCtx, t)
		(&TaskBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
		(&TaskBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	}

	userCtxPersonalOrg := auth.NewTestContextWithOrgID(testUser1.ID, testUser1.PersonalOrgID)

	// add a task for the user to another org; this should not be returned for JWT auth, since it's
	// restricted to a single org. PAT auth would return it if both orgs are authorized on the token
	(&TaskBuilder{client: suite.client, AssigneeID: testUser1.ID}).MustNew(userCtxPersonalOrg, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
		totalCount      int64
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: testutils.MaxResultLimit,
			totalCount:      30,
		},
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: testutils.MaxResultLimit,
			totalCount:      10,
		},
		{
			name:            "happy path, using pat - which should have access to all tasks because its authorized to the personal org",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: testutils.MaxResultLimit,
			totalCount:      31,
		},
		{
			name:            "another user, no entities should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: testutils.MaxResultLimit,
			totalCount:      10,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			first := int64(10)
			resp, err := tc.client.GetTasks(tc.ctx, &first, nil, nil)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Tasks.Edges, tc.expectedResults)
			assert.Equal(t, tc.totalCount, resp.Tasks.TotalCount)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateTask() {
	t := suite.T()

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
				AssigneeID: &viewOnlyUser.ID, // assign the task to another user
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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.request.Title, resp.CreateTask.Task.Title)

			assert.NotEmpty(t, resp.CreateTask.Task.DisplayID)
			assert.Contains(t, resp.CreateTask.Task.DisplayID, "TSK-")

			assert.NotNil(t, resp.CreateTask.Task.OwnerID)

			if tc.request.Details == nil {
				assert.Empty(t, resp.CreateTask.Task.Details)
			} else {
				assert.Equal(t, tc.request.Details, resp.CreateTask.Task.Details)
			}

			if tc.request.Status == nil {
				assert.Equal(t, enums.TaskStatusOpen, resp.CreateTask.Task.Status)
			} else {
				assert.Equal(t, *tc.request.Status, resp.CreateTask.Task.Status)
			}

			if tc.request.Details == nil {
				assert.Empty(t, resp.CreateTask.Task.Details)
			} else {
				assert.Equal(t, tc.request.Details, resp.CreateTask.Task.Details)
			}

			if tc.request.Category == nil {
				assert.Empty(t, resp.CreateTask.Task.Category)
			} else {
				assert.Equal(t, tc.request.Category, resp.CreateTask.Task.Category)
			}

			if tc.request.Due == nil {
				assert.Empty(t, resp.CreateTask.Task.Due)
			} else {
				assert.WithinDuration(t, time.Time(*tc.request.Due), time.Time(*resp.CreateTask.Task.Due), 10*time.Second)
			}

			// when using an API token, the assigner is not set
			if tc.client == suite.client.apiWithToken {
				assert.Nil(t, resp.CreateTask.Task.Assigner)
			} else {
				// otherwise it defaults to the authorized user
				assert.NotNil(t, resp.CreateTask.Task.Assigner)
				assert.Equal(t, testUser1.ID, resp.CreateTask.Task.Assigner.ID)
			}

			if tc.request.AssigneeID == nil {
				assert.Nil(t, resp.CreateTask.Task.Assignee)
			} else {
				require.NotNil(t, resp.CreateTask.Task.Assignee)

				assert.Equal(t, *tc.request.AssigneeID, resp.CreateTask.Task.Assignee.ID)

				// make sure the assignee can see the task
				taskResp, err := suite.client.api.GetTaskByID(viewOnlyUser.UserCtx, resp.CreateTask.Task.ID)
				require.NoError(t, err)
				assert.NotNil(t, taskResp)

				// make sure the another org member cannot see the task
				taskResp, err = suite.client.api.GetTaskByID(adminUser.UserCtx, resp.CreateTask.Task.ID)
				require.Error(t, err)
				assert.Nil(t, taskResp)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateTask() {
	t := suite.T()

	task := (&TaskBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)
	group := (&GroupBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	pngFile, err := objects.NewUploadFile("testdata/uploads/logo.png")
	require.NoError(t, err)

	pdfFile, err := objects.NewUploadFile("testdata/uploads/hello.pdf")
	require.NoError(t, err)

	taskCommentID := ""

	// make sure the user cannot can see the task before they are the assigner
	taskResp, err := suite.client.api.GetTaskByID(viewOnlyUser2.UserCtx, task.ID)
	require.Error(t, err)
	assert.Nil(t, taskResp)

	// make sure the user cannot can see the task before they are the assignee
	taskResp, err = suite.client.api.GetTaskByID(viewOnlyUser.UserCtx, task.ID)
	require.Error(t, err)
	assert.Nil(t, taskResp)

	// NOTE: the tests and checks are ordered due to dependencies between updates
	// if you update cases, they will most likely need to be added to the end of the list
	testCases := []struct {
		name                 string
		request              *openlaneclient.UpdateTaskInput
		updateCommentRequest *openlaneclient.UpdateNoteInput
		files                []*graphql.Upload
		client               *openlaneclient.OpenlaneClient
		ctx                  context.Context
		expectedErr          string
	}{
		{
			name: "happy path, update details",
			request: &openlaneclient.UpdateTaskInput{
				Details:    lo.ToPtr(("makin' a list, checkin' it twice")),
				AssigneeID: &adminUser.ID,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, add comment",
			request: &openlaneclient.UpdateTaskInput{
				AddComment: &openlaneclient.CreateNoteInput{
					Text: "matt is the best",
				},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, update comment with files",
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
			name: "happy path, update comment with file using PAT",
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
			name: "happy path, delete comment",
			request: &openlaneclient.UpdateTaskInput{
				DeleteComment: &taskCommentID,
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "update category using pat of owner",
			request: &openlaneclient.UpdateTaskInput{
				Category: lo.ToPtr("risk review"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update assignee to user not in org should fail",
			request: &openlaneclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(testUser2.ID),
			},
			client:      suite.client.api,
			ctx:         adminUser.UserCtx,
			expectedErr: "user not in organization",
		},
		{
			name: "update assignee to view only user",
			request: &openlaneclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(viewOnlyUser.ID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "update assignee to same user, should not error",
			request: &openlaneclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(viewOnlyUser.ID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "update status and details",
			request: &openlaneclient.UpdateTaskInput{
				Status:  &enums.TaskStatusInProgress,
				Details: lo.ToPtr("do all the things for the thing"),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "add to group",
			request: &openlaneclient.UpdateTaskInput{
				AddGroupIDs: []string{group.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "update assigner to another org member, this user should still be able to see it because they originally created it",
			request: &openlaneclient.UpdateTaskInput{
				AssignerID: lo.ToPtr(viewOnlyUser2.ID),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "clear assignee",
			request: &openlaneclient.UpdateTaskInput{
				ClearAssignee: lo.ToPtr(true),
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "clear assigner",
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
				resp, err = tc.client.UpdateTask(tc.ctx, task.ID, *tc.request)
				if tc.expectedErr != "" {
					require.Error(t, err)
					assert.ErrorContains(t, err, tc.expectedErr)
					assert.Nil(t, resp)

					return
				}
			} else if tc.updateCommentRequest != nil {
				if len(tc.files) > 0 {
					expectUploadNillable(t, suite.client.objectStore.Storage, tc.files)
				}
				commentResp, err = suite.client.api.UpdateTaskComment(testUser1.UserCtx, taskCommentID, *tc.updateCommentRequest, tc.files)
			}

			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)

			if tc.request != nil {
				require.NotNil(t, resp)

				if tc.request.Details != nil {
					assert.Equal(t, *tc.request.Details, *resp.UpdateTask.Task.Details)
				}

				if tc.request.Status != nil {
					assert.Equal(t, *tc.request.Status, resp.UpdateTask.Task.Status)
				}

				if tc.request.Details != nil {
					assert.Equal(t, tc.request.Details, resp.UpdateTask.Task.Details)
				}

				if tc.request.Category != nil {
					assert.Equal(t, *tc.request.Category, *resp.UpdateTask.Task.Category)
				}

				if tc.request.ClearAssignee != nil {
					assert.Nil(t, resp.UpdateTask.Task.Assignee)

					// the previous assignee should no longer be able to see the task
					taskResp, err := suite.client.api.GetTaskByID(viewOnlyUser.UserCtx, resp.UpdateTask.Task.ID)
					assert.Error(t, err)
					assert.Nil(t, taskResp)
				}

				if tc.request.ClearAssigner != nil {
					assert.Nil(t, resp.UpdateTask.Task.Assignee)

					// the previous assigner should no longer be able to see the task
					taskResp, err := suite.client.api.GetTaskByID(viewOnlyUser2.UserCtx, resp.UpdateTask.Task.ID)
					assert.Error(t, err)
					assert.Nil(t, taskResp)
				}

				if tc.request.AssignerID != nil {
					assert.NotNil(t, resp.UpdateTask.Task.Assigner)
					assert.Equal(t, *tc.request.AssignerID, resp.UpdateTask.Task.Assigner.ID)

					// make sure the assigner can see the task
					taskResp, err := suite.client.api.GetTaskByID(viewOnlyUser2.UserCtx, resp.UpdateTask.Task.ID)
					assert.NoError(t, err)
					assert.NotNil(t, taskResp)

					// make sure the original creator can still see the task
					taskResp, err = suite.client.api.GetTaskByID(adminUser.UserCtx, resp.UpdateTask.Task.ID)
					require.NoError(t, err)
					assert.NotNil(t, taskResp)
				}

				if tc.request.AddComment != nil {
					assert.NotEmpty(t, resp.UpdateTask.Task.Comments.Edges)
					assert.Equal(t, tc.request.AddComment.Text, resp.UpdateTask.Task.Comments.Edges[0].Node.Text)

					// there should only be one comment
					require.Len(t, resp.UpdateTask.Task.Comments.Edges, 1)
					taskCommentID = resp.UpdateTask.Task.Comments.Edges[0].Node.ID

					// user shouldn't be able to see the comment
					checkResp, err := suite.client.api.GetNoteByID(viewOnlyUser.UserCtx, taskCommentID)
					assert.Error(t, err)
					assert.Nil(t, checkResp)

					// user should be able to see the comment since they created the task
					checkResp, err = suite.client.api.GetNoteByID(adminUser.UserCtx, taskCommentID)
					assert.NoError(t, err)
					assert.NotNil(t, checkResp)

					// org owner should be able to see the comment
					checkResp, err = suite.client.api.GetNoteByID(testUser1.UserCtx, taskCommentID)
					assert.NoError(t, err)
					assert.NotNil(t, checkResp)
				} else if tc.request.DeleteComment != nil {
					// should not have any comments
					assert.Len(t, resp.UpdateTask.Task.Comments.Edges, 0)
				}
			} else if tc.updateCommentRequest != nil {
				require.NotNil(t, commentResp)

				// should only have the original comment
				require.Len(t, commentResp.UpdateTaskComment.Task.Comments.Edges, 1)
				assert.Equal(t, *tc.updateCommentRequest.Text, commentResp.UpdateTaskComment.Task.Comments.Edges[0].Node.Text)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationDeleteTask() {
	t := suite.T()

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
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteTask.DeletedID)
		})
	}
}
