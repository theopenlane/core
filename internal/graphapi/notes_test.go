package graphapi_test

import (
	"context"
	"testing"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestMutationUpdateNote(t *testing.T) {
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateTaskInput
		client      openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.UpdateTaskInput{
				AddComment: &openlaneclient.CreateNoteInput{
					Text: "This is a test note",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path with PAT",
			request: openlaneclient.UpdateTaskInput{
				AddComment: &openlaneclient.CreateNoteInput{
					Text:    "This is a test note using PAT",
					OwnerID: &testUser1.OrganizationID,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "missing required field - text",
			request: openlaneclient.UpdateTaskInput{
				AddComment: &openlaneclient.CreateNoteInput{},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "owner id must be present with pat",
			request: openlaneclient.UpdateTaskInput{
				AddComment: &openlaneclient.CreateNoteInput{
					Text: "This is a test note using PAT",
				},
			},
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			expectedErr: "owner_id is required",
		},
		{
			name: "task not found",
			request: openlaneclient.UpdateTaskInput{
				AddComment: &openlaneclient.CreateNoteInput{
					Text:    "This is a test note",
					OwnerID: &testUser1.OrganizationID,
				},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx, //wrong user
			expectedErr: "task not found",
		},
	}

	for idx, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {

			resp, err := tc.client.UpdateTask(tc.ctx, task.ID, tc.request)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, len(resp.UpdateTask.Task.Comments.Edges) != 0)

			assert.Check(t, is.Equal(tc.request.AddComment.Text, resp.UpdateTask.Task.Comments.Edges[idx].Node.Text))

			noteID := resp.UpdateTask.Task.Comments.Edges[idx].Node.ID

			_, err = tc.client.GetNoteByID(tc.ctx, noteID)
			assert.NilError(t, err)
		})
	}

	// clean up
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: task.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteNote(t *testing.T) {
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     func() openlaneclient.UpdateTaskInput // changed to function to get fresh note ID each time
		client      openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			request: func() openlaneclient.UpdateTaskInput {
				createResp, err := suite.client.api.UpdateTask(testUser1.UserCtx, task.ID, openlaneclient.UpdateTaskInput{
					AddComment: &openlaneclient.CreateNoteInput{
						Text: "Note to be deleted",
					},
				})
				assert.NilError(t, err)
				assert.Assert(t, createResp != nil)
				assert.Assert(t, len(createResp.UpdateTask.Task.Comments.Edges) != 0)
				noteID := createResp.UpdateTask.Task.Comments.Edges[0].Node.ID
				return openlaneclient.UpdateTaskInput{
					DeleteComment: &noteID,
				}
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path with PAT",
			request: func() openlaneclient.UpdateTaskInput {
				// create a note to delete
				createResp, err := suite.client.api.UpdateTask(testUser1.UserCtx, task.ID, openlaneclient.UpdateTaskInput{
					AddComment: &openlaneclient.CreateNoteInput{
						Text: "Note to be deleted with PAT",
					},
				})
				assert.NilError(t, err)
				assert.Assert(t, createResp != nil)
				assert.Assert(t, len(createResp.UpdateTask.Task.Comments.Edges) != 0)
				noteID := createResp.UpdateTask.Task.Comments.Edges[0].Node.ID
				return openlaneclient.UpdateTaskInput{
					DeleteComment: &noteID,
				}
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "comment not found",
			request: func() openlaneclient.UpdateTaskInput {
				return openlaneclient.UpdateTaskInput{
					DeleteComment: &[]string{"non-existent-id"}[0],
				}
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "comment not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			request := tc.request() // get fresh request with new note
			resp, err := tc.client.UpdateTask(tc.ctx, task.ID, request)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			noteID := *request.DeleteComment
			_, err = tc.client.GetNoteByID(tc.ctx, noteID)
			assert.Check(t, is.ErrorContains(err, ""))
			assert.ErrorContains(t, err, "note not found")
		})
	}

	// clean up
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: task.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryNote(t *testing.T) {
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	createResp, err := suite.client.api.UpdateTask(testUser1.UserCtx, task.ID, openlaneclient.UpdateTaskInput{
		AddComment: &openlaneclient.CreateNoteInput{
			Text: "Note for querying",
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, createResp != nil)
	assert.Assert(t, len(createResp.UpdateTask.Task.Comments.Edges) != 0)
	noteID := createResp.UpdateTask.Task.Comments.Edges[0].Node.ID

	testCases := []struct {
		name        string
		noteID      string
		client      openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:   "happy path",
			noteID: noteID,
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:   "happy path with PAT",
			noteID: noteID,
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:        "note not found",
			noteID:      "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "note not found",
		},
		{
			name:        "unauthorized user",
			noteID:      noteID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: "note not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Query "+tc.name, func(t *testing.T) {
			note, err := tc.client.GetNoteByID(tc.ctx, tc.noteID)
			if tc.expectedErr != "" {

				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(note))
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, note != nil)
			assert.Check(t, is.Equal(tc.noteID, note.Note.ID))
			assert.Check(t, is.Equal("Note for querying", note.Note.Text))
		})
	}

	// clean up
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: task.ID}).MustDelete(testUser1.UserCtx, t)
}
