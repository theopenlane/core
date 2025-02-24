package graphapi_test

import (
	"context"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
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
			assert.NotEmpty(t, resp.Task.Description)
			assert.NotEmpty(t, resp.Task.Status)
		})
	}
}

func (suite *GraphTestSuite) TestQueryTasks() {
	t := suite.T()

	(&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no entities should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllTasks(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Tasks.Edges, tc.expectedResults)
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
				Title:       "test-task",
				Description: lo.ToPtr("test description"),
				Status:      &enums.TaskStatusInProgress,
				Category:    lo.ToPtr("evidence upload"),
				Details:     lo.ToPtr("do all the things for the thing"),
				Due:         lo.ToPtr(time.Now().Add(time.Hour * 24)),
				AssigneeID:  &viewOnlyUser.ID, // assign the task to another user
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
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
			name: "missing title, but display name provided",
			request: openlaneclient.CreateTaskInput{
				Description: lo.ToPtr("makin' a list, checkin' it twice"),
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

			if tc.request.Description == nil {
				assert.Empty(t, resp.CreateTask.Task.Description)
			} else {
				assert.Equal(t, tc.request.Description, resp.CreateTask.Task.Description)
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
				assert.WithinDuration(t, *tc.request.Due, *resp.CreateTask.Task.Due, 10*time.Second)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateTask() {
	t := suite.T()

	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	group := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	taskCommentID := ""

	testCases := []struct {
		name                 string
		request              *openlaneclient.UpdateTaskInput
		updateCommentRequest *openlaneclient.UpdateNoteInput
		client               *openlaneclient.OpenlaneClient
		ctx                  context.Context
		expectedErr          string
	}{
		{
			name: "happy path, update description",
			request: &openlaneclient.UpdateTaskInput{
				Description: lo.ToPtr(("makin' a list, checkin' it twice")),
				AssigneeID:  &adminUser.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, add comment",
			request: &openlaneclient.UpdateTaskInput{
				AddComment: &openlaneclient.CreateNoteInput{
					Text: "matt is the best",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, update comment",
			updateCommentRequest: &openlaneclient.UpdateNoteInput{
				Text: lo.ToPtr("sarah is better"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, delete comment",
			request: &openlaneclient.UpdateTaskInput{
				DeleteComment: &taskCommentID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update category",
			request: &openlaneclient.UpdateTaskInput{
				Category: lo.ToPtr("risk review"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update assignee",
			request: &openlaneclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(viewOnlyUser.ID),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update assignee to same user, should not error",
			request: &openlaneclient.UpdateTaskInput{
				AssigneeID: lo.ToPtr(viewOnlyUser.ID),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update status and details",
			request: &openlaneclient.UpdateTaskInput{
				Status:  &enums.TaskStatusInProgress,
				Details: lo.ToPtr("do all the things for the thing"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add to group",
			request: &openlaneclient.UpdateTaskInput{
				AddGroupIDs: []string{group.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
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
				commentResp, err = suite.client.api.UpdateTaskComment(testUser1.UserCtx, taskCommentID, *tc.updateCommentRequest)

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

				if tc.request.Description != nil {
					assert.Equal(t, *tc.request.Description, *resp.UpdateTask.Task.Description)
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

				if tc.request.AddComment != nil {
					assert.NotEmpty(t, resp.UpdateTask.Task.Comments)
					assert.Equal(t, tc.request.AddComment.Text, resp.UpdateTask.Task.Comments[0].Text)

					// there should only be one comment
					require.Len(t, resp.UpdateTask.Task.Comments, 1)
					taskCommentID = resp.UpdateTask.Task.Comments[0].ID

					// user shouldn't be able to see the comment
					checkResp, err := suite.client.api.GetNoteByID(viewOnlyUser.UserCtx, taskCommentID)
					require.Error(t, err)
					assert.Nil(t, checkResp)

					// user should be able to see the comment because they are an the assignee
					checkResp, err = suite.client.api.GetNoteByID(adminUser.UserCtx, taskCommentID)
					require.Error(t, err)
					assert.Nil(t, checkResp)

					// org owner should be able to see the comment
					checkResp, err = suite.client.api.GetNoteByID(testUser1.UserCtx, taskCommentID)
					require.NoError(t, err)
					assert.NotNil(t, checkResp)
				} else if tc.request.DeleteComment != nil {
					// should not have any comments
					require.Len(t, resp.UpdateTask.Task.Comments, 0)
				}
			} else if tc.updateCommentRequest != nil {
				require.NotNil(t, commentResp)

				// should only have the original comment
				require.Len(t, commentResp.UpdateTaskComment.Task.Comments, 1)
				assert.Equal(t, *tc.updateCommentRequest.Text, commentResp.UpdateTaskComment.Task.Comments[0].Text)
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
