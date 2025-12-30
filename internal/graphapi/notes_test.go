package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestMutationUpdateNoteForTask(t *testing.T) {
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateTaskInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.UpdateTaskInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "This is a test note",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path with PAT",
			request: testclient.UpdateTaskInput{
				AddComment: &testclient.CreateNoteInput{
					Text:    "This is a test note using PAT",
					OwnerID: &testUser1.OrganizationID,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "missing required field - text",
			request: testclient.UpdateTaskInput{
				AddComment: &testclient.CreateNoteInput{},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "task not found",
			request: testclient.UpdateTaskInput{
				AddComment: &testclient.CreateNoteInput{
					Text:    "This is a test note",
					OwnerID: &testUser1.OrganizationID,
				},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx, // wrong user
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for idx, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateTask(tc.ctx, task.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

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

func TestMutationAddNoteForControl(t *testing.T) {
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateControlInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, add discussion",
			request: testclient.UpdateControlInput{
				AddDiscussion: &testclient.CreateDiscussionInput{
					ExternalID: lo.ToPtr("DISC-12345"),
					IsResolved: lo.ToPtr(false),
					AddComment: &testclient.CreateNoteInput{
						Text: "This is a test note as part of a discussion",
					},
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, minimal input",
			request: testclient.UpdateControlInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "This is a test note",
				},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path with PAT",
			request: testclient.UpdateControlInput{
				AddComment: &testclient.CreateNoteInput{
					Text:    "This is a test note using PAT",
					OwnerID: &testUser1.OrganizationID,
				},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "missing required field - text",
			request: testclient.UpdateControlInput{
				AddComment: &testclient.CreateNoteInput{},
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "control not found",
			request: testclient.UpdateControlInput{
				AddComment: &testclient.CreateNoteInput{
					Text:    "This is a test note",
					OwnerID: &testUser1.OrganizationID,
				},
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx, // wrong user
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for idx, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateControl(tc.ctx, control.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Assert(t, resp.UpdateControl.Control.Comments.Edges != nil)
			// we want to make sure we have at least idx+1 comments, this means you need to keep the test order
			// intact so all failures are after all successes
			assert.Assert(t, len(resp.UpdateControl.Control.Comments.Edges) > idx)

			if tc.request.AddComment != nil {
				assert.Check(t, is.Equal(tc.request.AddComment.Text, resp.UpdateControl.Control.Comments.Edges[idx].Node.Text))
			}

			if tc.request.AddDiscussion != nil {
				assert.Assert(t, resp.UpdateControl.Control.Discussions.Edges != nil)
				assert.Assert(t, len(resp.UpdateControl.Control.Discussions.Edges) != 0)
				assert.Assert(t, len(resp.UpdateControl.Control.Discussions.Edges[0].Node.Comments.Edges) != 0)
				assert.Check(t, is.Equal(tc.request.AddDiscussion.AddComment.Text, resp.UpdateControl.Control.Discussions.Edges[0].Node.Comments.Edges[0].Node.Text))
			}

			noteID := resp.UpdateControl.Control.Comments.Edges[idx].Node.ID

			_, err = tc.client.GetNoteByID(tc.ctx, noteID)
			assert.NilError(t, err)
		})
	}

	// clean up
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateDiscussionForControl(t *testing.T) {
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add initial discussion to update
	resp, err := suite.client.api.UpdateControl(testUser1.UserCtx, control.ID, testclient.UpdateControlInput{
		AddDiscussion: &testclient.CreateDiscussionInput{
			ExternalID: lo.ToPtr("DISC-22401"),
			IsResolved: lo.ToPtr(false),
			AddComment: &testclient.CreateNoteInput{
				Text: "This is a test note as part of a discussion",
			},
		},
	})

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Assert(t, resp.UpdateControl.Control.Comments.Edges != nil)
	assert.Assert(t, len(resp.UpdateControl.Control.Comments.Edges) == 1)

	assert.Assert(t, resp.UpdateControl.Control.Discussions.Edges != nil)
	assert.Assert(t, len(resp.UpdateControl.Control.Discussions.Edges) != 0)
	assert.Assert(t, len(resp.UpdateControl.Control.Discussions.Edges[0].Node.Comments.Edges) != 0)
	assert.Check(t, is.Equal("This is a test note as part of a discussion", resp.UpdateControl.Control.Discussions.Edges[0].Node.Comments.Edges[0].Node.Text))

	discussionID := resp.UpdateControl.Control.Discussions.Edges[0].Node.ID

	// now update the discussion by adding another comment
	updateResp, err := suite.client.api.UpdateControl(testUser1.UserCtx, control.ID, testclient.UpdateControlInput{
		UpdateDiscussion: &testclient.UpdateDiscussionsInput{
			ID: discussionID,
			Input: &testclient.UpdateDiscussionInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "This is an additional comment in the discussion",
				},
			},
		},
	})

	assert.NilError(t, err)
	assert.Assert(t, updateResp != nil)
	assert.Assert(t, updateResp.UpdateControl.Control.Discussions.Edges != nil)
	assert.Assert(t, len(updateResp.UpdateControl.Control.Discussions.Edges) != 0)
	assert.Assert(t, updateResp.UpdateControl.Control.Discussions.Edges[0].Node.Comments.Edges != nil)
	assert.Assert(t, len(updateResp.UpdateControl.Control.Discussions.Edges[0].Node.Comments.Edges) == 2)

	// make sure the control also has the comments linked
	assert.Assert(t, updateResp.UpdateControl.Control.Comments.Edges != nil)
	assert.Assert(t, len(updateResp.UpdateControl.Control.Comments.Edges) == 2)

	updatedDiscussion := updateResp.UpdateControl.Control.Discussions.Edges[0].Node
	assert.Assert(t, len(updatedDiscussion.Comments.Edges) == 2)
	assert.Check(t, is.Equal("This is an additional comment in the discussion", updatedDiscussion.Comments.Edges[1].Node.Text))

	// now lets try to update the second comment in the discussion
	noteToUpdateID := updatedDiscussion.Comments.Edges[1].Node.ID
	updatedText := "This is an updated additional comment in the discussion"
	updateComment, err := suite.client.api.UpdateControlComment(testUser1.UserCtx, noteToUpdateID, testclient.UpdateNoteInput{
		Text: &updatedText,
	})

	assert.NilError(t, err)
	assert.Assert(t, updateComment != nil)
	assert.Assert(t, updateComment.UpdateControlComment.Control.Comments.Edges != nil)
	for _, edge := range updateComment.UpdateControlComment.Control.Comments.Edges {
		if edge.Node.ID == noteToUpdateID {
			assert.Check(t, is.Equal(updatedText, edge.Node.Text))
		}
	}

	// ensure its the same on the discussion side
	for _, discEdge := range updateComment.UpdateControlComment.Control.Discussions.Edges {
		if discEdge.Node.ID == discussionID {
			for _, commentEdge := range discEdge.Node.Comments.Edges {
				if commentEdge.Node.ID == noteToUpdateID {
					assert.Check(t, is.Equal(updatedText, commentEdge.Node.Text))
				}
			}
		}
	}

	// now lets try to remove a comment from the discussion
	noteToRemoveID := updatedDiscussion.Comments.Edges[0].Node.ID

	updateResp2, err := suite.client.api.UpdateControl(testUser1.UserCtx, control.ID, testclient.UpdateControlInput{
		UpdateDiscussion: &testclient.UpdateDiscussionsInput{
			ID: discussionID,
			Input: &testclient.UpdateDiscussionInput{
				RemoveCommentIDs: []string{noteToRemoveID},
			},
		},
	})

	assert.NilError(t, err)
	assert.Assert(t, updateResp2 != nil)
	assert.Assert(t, updateResp2.UpdateControl.Control.Discussions.Edges != nil)
	assert.Assert(t, len(updateResp2.UpdateControl.Control.Discussions.Edges) != 0)

	updatedDiscussion2 := updateResp2.UpdateControl.Control.Discussions.Edges[0].Node
	assert.Assert(t, len(updatedDiscussion2.Comments.Edges) == 1)
	// this will be the updated comment
	assert.Check(t, is.Equal(updatedText, updatedDiscussion2.Comments.Edges[0].Node.Text))

	// clean up
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteNoteForTask(t *testing.T) {
	userTask := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	createResp, err := suite.client.api.UpdateTask(testUser1.UserCtx, userTask.ID, testclient.UpdateTaskInput{
		AddComment: &testclient.CreateNoteInput{
			Text: "Here is my comment",
		},
	})

	assert.NilError(t, err)

	assert.Assert(t, createResp != nil)
	assert.Assert(t, len(createResp.UpdateTask.Task.Comments.Edges) != 0)
	noteID := createResp.UpdateTask.Task.Comments.Edges[0].Node.ID

	_, err = suite.client.api.DeleteNote(testUser1.UserCtx, noteID)
	assert.NilError(t, err)

	// cleanup task
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: userTask.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteTaskNotes(t *testing.T) {
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     func() testclient.UpdateTaskInput // changed to function to get fresh note ID each time
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			request: func() testclient.UpdateTaskInput {
				createResp, err := suite.client.api.UpdateTask(testUser1.UserCtx, task.ID, testclient.UpdateTaskInput{
					AddComment: &testclient.CreateNoteInput{
						Text: "Note to be deleted",
					},
				})
				assert.NilError(t, err)
				assert.Assert(t, createResp != nil)
				assert.Assert(t, len(createResp.UpdateTask.Task.Comments.Edges) != 0)
				noteID := createResp.UpdateTask.Task.Comments.Edges[0].Node.ID
				return testclient.UpdateTaskInput{
					DeleteComment: &noteID,
				}
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path with PAT",
			request: func() testclient.UpdateTaskInput {
				// create a note to delete
				createResp, err := suite.client.api.UpdateTask(testUser1.UserCtx, task.ID, testclient.UpdateTaskInput{
					AddComment: &testclient.CreateNoteInput{
						Text: "Note to be deleted with PAT",
					},
				})
				assert.NilError(t, err)
				assert.Assert(t, createResp != nil)
				assert.Assert(t, len(createResp.UpdateTask.Task.Comments.Edges) != 0)
				noteID := createResp.UpdateTask.Task.Comments.Edges[0].Node.ID
				return testclient.UpdateTaskInput{
					DeleteComment: &noteID,
				}
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "comment not found",
			request: func() testclient.UpdateTaskInput {
				return testclient.UpdateTaskInput{
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
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			noteID := *request.DeleteComment
			_, err = tc.client.GetNoteByID(tc.ctx, noteID)
			assert.Check(t, is.ErrorContains(err, notFoundErrorMsg))
		})
	}

	// clean up
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, ID: task.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryNote(t *testing.T) {
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	createResp, err := suite.client.api.UpdateTask(testUser1.UserCtx, task.ID, testclient.UpdateTaskInput{
		AddComment: &testclient.CreateNoteInput{
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
		client      *testclient.TestClient
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
