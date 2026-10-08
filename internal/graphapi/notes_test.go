package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	entfinding "github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/note"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestMutationUpdateNoteForTask(t *testing.T) {
	task := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path with PAT",
			request: testclient.UpdateTaskInput{
				AddComment: &testclient.CreateNoteInput{
					Text:    "This is a test note using PAT",
					OwnerID: &th.SharedTestUser1.OrganizationID,
				},
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "missing required field - text",
			request: testclient.UpdateTaskInput{
				AddComment: &testclient.CreateNoteInput{},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "task not found",
			request: testclient.UpdateTaskInput{
				AddComment: &testclient.CreateNoteInput{
					Text:    "This is a test note",
					OwnerID: &th.SharedTestUser1.OrganizationID,
				},
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx, // wrong user
			expectedErr: th.NotAuthorizedErrorMsg,
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
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, ID: task.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationAddNoteForReview(t *testing.T) {
	createResp, err := suite.Client.API.CreateReview(th.SharedTestUser1.UserCtx, testclient.CreateReviewInput{
		Title: "This is a review",
	})
	assert.NilError(t, err)
	assert.Assert(t, createResp != nil)

	id := createResp.GetCreateReview().GetReview().ID

	tt := []struct {
		name    string
		ctx     context.Context
		comment string
	}{
		{
			name:    "user can create comment",
			ctx:     th.SharedViewOnlyUser.UserCtx,
			comment: "This is a review comment from a user",
		},
		{
			name:    "admin can create comment",
			ctx:     th.SharedAdminUser.UserCtx,
			comment: "This is a review comment from an admin",
		},
		{
			name:    "owner can create comment",
			ctx:     th.SharedTestUser1.UserCtx,
			comment: "This is a review comment from an owner",
		},
		{
			name:    "auditor can create comment",
			ctx:     th.SharedAuditorUser.UserCtx,
			comment: "This is a review comment from an auditor",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.UpdateReview(tc.ctx, id, testclient.UpdateReviewInput{
				AddComment: &testclient.CreateNoteInput{
					Text: tc.comment,
				},
			}, nil)

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(id, resp.GetUpdateReview().GetReview().ID))
		})
	}

	(&th.Cleanup[*generated.ReviewDeleteOne]{Client: suite.Client.DB.Review, ID: id}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationAddNoteForControl(t *testing.T) {
	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// ensure view only user caan see the control
	_, err := suite.Client.API.GetControlByID(th.SharedViewOnlyUser.UserCtx, control.ID)
	assert.NilError(t, err)
	assert.Assert(t, control.ID != "")

	testCases := []struct {
		name        string
		request     testclient.UpdateControlInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, member can add comment",
			request: testclient.UpdateControlInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "This is a test note",
				},
			},
			client: suite.Client.API,
			ctx:    th.SharedViewOnlyUser.UserCtx,
		},
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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, minimal input",
			request: testclient.UpdateControlInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "This is a test note",
				},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path with PAT",
			request: testclient.UpdateControlInput{
				AddComment: &testclient.CreateNoteInput{
					Text:    "This is a test note using PAT",
					OwnerID: &th.SharedTestUser1.OrganizationID,
				},
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},

		{
			name: "missing required field - text",
			request: testclient.UpdateControlInput{
				AddComment: &testclient.CreateNoteInput{},
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for idx, tc := range testCases {
		t.Run("Add Note "+tc.name, func(t *testing.T) {
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
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: control.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationUpdateDiscussionForControl(t *testing.T) {
	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add initial discussion to update
	resp, err := suite.Client.API.UpdateControl(th.SharedTestUser1.UserCtx, control.ID, testclient.UpdateControlInput{
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
	updateResp, err := suite.Client.API.UpdateControl(th.SharedTestUser1.UserCtx, control.ID, testclient.UpdateControlInput{
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
	updateComment, err := suite.Client.API.UpdateControlComment(th.SharedTestUser1.UserCtx, noteToUpdateID, testclient.UpdateNoteInput{
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

	updateResp2, err := suite.Client.API.UpdateControl(th.SharedTestUser1.UserCtx, control.ID, testclient.UpdateControlInput{
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
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: control.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationUpdateDiscussionForPolicy(t *testing.T) {
	policy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add initial discussion to update
	resp, err := suite.Client.API.UpdateInternalPolicy(th.SharedTestUser1.UserCtx, policy.ID, testclient.UpdateInternalPolicyInput{
		AddDiscussion: &testclient.CreateDiscussionInput{
			ExternalID: lo.ToPtr("DISC-22402"),
			IsResolved: lo.ToPtr(false),
			AddComment: &testclient.CreateNoteInput{
				Text: "This is a test note as part of a discussion",
			},
		},
	})

	assert.NilError(t, err)
	assert.Assert(t, resp != nil)
	assert.Assert(t, resp.UpdateInternalPolicy.InternalPolicy.Comments.Edges != nil)
	assert.Assert(t, len(resp.UpdateInternalPolicy.InternalPolicy.Comments.Edges) == 1)

	assert.Assert(t, resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges != nil)
	assert.Assert(t, len(resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges) != 0)
	assert.Assert(t, len(resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node.Comments.Edges) != 0)
	assert.Check(t, is.Equal("This is a test note as part of a discussion", resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node.Comments.Edges[0].Node.Text))

	discussionID := resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node.ID

	// now update the discussion by adding another comment
	updateResp, err := suite.Client.API.UpdateInternalPolicy(th.SharedTestUser1.UserCtx, policy.ID, testclient.UpdateInternalPolicyInput{
		UpdateDiscussion: &testclient.UpdateDiscussionsInput{
			ID: discussionID,
			Input: &testclient.UpdateDiscussionInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "This is an additional comment in the discussion for the policy",
				},
			},
		},
	})

	assert.NilError(t, err)
	assert.Assert(t, updateResp != nil)
	assert.Assert(t, updateResp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges != nil)
	assert.Assert(t, len(updateResp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges) != 0)
	assert.Assert(t, updateResp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node.Comments.Edges != nil)
	assert.Assert(t, len(updateResp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node.Comments.Edges) == 2)

	// make sure the policy also has the comments linked
	assert.Assert(t, updateResp.UpdateInternalPolicy.InternalPolicy.Comments.Edges != nil)
	assert.Assert(t, len(updateResp.UpdateInternalPolicy.InternalPolicy.Comments.Edges) == 2)

	updatedDiscussion := updateResp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node
	assert.Assert(t, len(updatedDiscussion.Comments.Edges) == 2)
	assert.Check(t, is.Equal("This is an additional comment in the discussion for the policy", updatedDiscussion.Comments.Edges[1].Node.Text))

	// now lets try to update the second comment in the discussion
	noteToUpdateID := updatedDiscussion.Comments.Edges[1].Node.ID
	updatedText := "This is an updated additional comment in the discussion for the policy"
	updateComment, err := suite.Client.API.UpdateInternalPolicyComment(th.SharedTestUser1.UserCtx, noteToUpdateID, testclient.UpdateNoteInput{
		Text: &updatedText,
	})

	assert.NilError(t, err)
	assert.Assert(t, updateComment != nil)
	assert.Assert(t, updateComment.UpdateInternalPolicyComment.InternalPolicy.Comments.Edges != nil)
	for _, edge := range updateComment.UpdateInternalPolicyComment.InternalPolicy.Comments.Edges {
		if edge.Node.ID == noteToUpdateID {
			assert.Check(t, is.Equal(updatedText, edge.Node.Text))
		}
	}

	// ensure its the same on the discussion side
	for _, discEdge := range updateComment.UpdateInternalPolicyComment.InternalPolicy.Discussions.Edges {
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

	updateResp2, err := suite.Client.API.UpdateInternalPolicy(th.SharedTestUser1.UserCtx, policy.ID, testclient.UpdateInternalPolicyInput{
		UpdateDiscussion: &testclient.UpdateDiscussionsInput{
			ID: discussionID,
			Input: &testclient.UpdateDiscussionInput{
				RemoveCommentIDs: []string{noteToRemoveID},
			},
		},
	})

	assert.NilError(t, err)
	assert.Assert(t, updateResp2 != nil)
	assert.Assert(t, updateResp2.UpdateInternalPolicy.InternalPolicy.Discussions.Edges != nil)
	assert.Assert(t, len(updateResp2.UpdateInternalPolicy.InternalPolicy.Discussions.Edges) != 0)

	updatedDiscussion2 := updateResp2.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node
	assert.Assert(t, len(updatedDiscussion2.Comments.Edges) == 1)
	// this will be the updated comment
	assert.Check(t, is.Equal(updatedText, updatedDiscussion2.Comments.Edges[0].Node.Text))

	// clean up
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: policy.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteNoteForTask(t *testing.T) {
	userTask := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	createResp, err := suite.Client.API.UpdateTask(th.SharedTestUser1.UserCtx, userTask.ID, testclient.UpdateTaskInput{
		AddComment: &testclient.CreateNoteInput{
			Text: "Here is my comment",
		},
	})

	assert.NilError(t, err)

	assert.Assert(t, createResp != nil)
	assert.Assert(t, len(createResp.UpdateTask.Task.Comments.Edges) != 0)
	noteID := createResp.UpdateTask.Task.Comments.Edges[0].Node.ID

	_, err = suite.Client.API.DeleteNote(th.SharedTestUser1.UserCtx, noteID)
	assert.NilError(t, err)

	// cleanup task
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, ID: userTask.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteTaskNotes(t *testing.T) {
	task := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

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
				createResp, err := suite.Client.API.UpdateTask(th.SharedTestUser1.UserCtx, task.ID, testclient.UpdateTaskInput{
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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path with PAT",
			request: func() testclient.UpdateTaskInput {
				// create a note to delete
				createResp, err := suite.Client.API.UpdateTask(th.SharedTestUser1.UserCtx, task.ID, testclient.UpdateTaskInput{
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
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "comment not found",
			request: func() testclient.UpdateTaskInput {
				return testclient.UpdateTaskInput{
					DeleteComment: &[]string{"non-existent-id"}[0],
				}
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
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
			assert.Check(t, is.ErrorContains(err, th.NotFoundErrorMsg))
		})
	}

	// clean up
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, ID: task.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryNote(t *testing.T) {
	task := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	createResp, err := suite.Client.API.UpdateTask(th.SharedTestUser1.UserCtx, task.ID, testclient.UpdateTaskInput{
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
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name:   "happy path with PAT",
			noteID: noteID,
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name:        "note not found",
			noteID:      "non-existent-id",
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "note not found",
		},
		{
			name:        "unauthorized user",
			noteID:      noteID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
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
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, ID: task.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

const ownerCommentText = "comment from the owner"

func TestMutationCommentsForVulnerability(t *testing.T) {
	vuln := createVulnerability(t, th.SharedTestUser1.UserCtx, "Commentable Vulnerability")

	testCases := []struct {
		name        string
		client      *testclient.TestClient
		ctx         context.Context
		comment     string
		expectedErr string
	}{
		{
			name:    "org owner can comment",
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			comment: ownerCommentText,
		},
		{
			name:    "view only user can comment",
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
			comment: "comment from a view only user",
		},
		{
			name:    "pat can comment",
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
			comment: "comment from a pat",
		},
		{
			name:    "api token can comment",
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
			comment: "comment from an api token",
		},
		{
			name:        "user in another org cannot comment",
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			comment:     "comment from another org",
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateVulnerability(tc.ctx, vuln.ID, testclient.UpdateVulnerabilityInput{
				AddComment: &testclient.CreateNoteInput{Text: tc.comment},
			})

			expectedCount := 1
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				expectedCount = 0
			} else {
				assert.NilError(t, err)
				assert.Check(t, is.Equal(vuln.ID, resp.UpdateVulnerability.Vulnerability.ID))
			}

			count, err := suite.Client.DB.Note.Query().
				Where(note.HasVulnerabilityWith(vulnerability.ID(vuln.ID)), note.Text(tc.comment)).
				Count(th.SharedTestUser1.UserCtx)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(expectedCount, count))
		})
	}

	viewerDenied := []struct {
		name  string
		input testclient.UpdateVulnerabilityInput
	}{
		{name: "set a field", input: testclient.UpdateVulnerabilityInput{DisplayName: lo.ToPtr("renamed by viewer")}},
		{name: "clear a field", input: testclient.UpdateVulnerabilityInput{ClearDisplayName: lo.ToPtr(true)}},
		{name: "clear all comments", input: testclient.UpdateVulnerabilityInput{ClearComments: lo.ToPtr(true)}},
		{name: "add a comment while clearing a field", input: testclient.UpdateVulnerabilityInput{
			AddComment:       &testclient.CreateNoteInput{Text: "sneaky comment"},
			ClearDisplayName: lo.ToPtr(true),
		}},
	}

	for _, tc := range viewerDenied {
		t.Run("view only user cannot "+tc.name, func(t *testing.T) {
			_, err := suite.Client.API.UpdateVulnerability(th.SharedViewOnlyUser.UserCtx, vuln.ID, tc.input)
			assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)
		})
	}

	t.Run("view only user cannot delete the vulnerability", func(t *testing.T) {
		_, err := suite.Client.API.DeleteVulnerability(th.SharedViewOnlyUser.UserCtx, vuln.ID)
		assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)
	})

	unchanged, err := suite.Client.DB.Vulnerability.Query().Where(vulnerability.ID(vuln.ID)).WithComments().Only(th.SharedTestUser1.UserCtx)
	assert.NilError(t, err)
	assert.Check(t, unchanged.DisplayName != "")
	assert.Check(t, is.Len(unchanged.Edges.Comments, 4))

	ownerComment, err := suite.Client.DB.Note.Query().
		Where(note.HasVulnerabilityWith(vulnerability.ID(vuln.ID)), note.Text(ownerCommentText)).
		Only(th.SharedTestUser1.UserCtx)
	assert.NilError(t, err)

	t.Run("view only user cannot detach another user's comment", func(t *testing.T) {
		_, err := suite.Client.API.UpdateVulnerability(th.SharedViewOnlyUser.UserCtx, vuln.ID, testclient.UpdateVulnerabilityInput{
			RemoveCommentIDs: []string{ownerComment.ID},
		})
		assert.Check(t, err != nil)

		stillAttached, err := suite.Client.DB.Note.Query().Where(note.ID(ownerComment.ID), note.HasVulnerability()).Exist(th.SharedTestUser1.UserCtx)
		assert.NilError(t, err)
		assert.Check(t, stillAttached)
	})

	t.Run("owner can update their comment", func(t *testing.T) {
		resp, err := suite.Client.API.UpdateVulnerabilityComment(th.SharedTestUser1.UserCtx, ownerComment.ID, testclient.UpdateNoteInput{
			Text: lo.ToPtr("edited comment from the owner"),
		}, nil)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(vuln.ID, resp.UpdateVulnerabilityComment.Vulnerability.ID))

		edited := lo.Filter(resp.UpdateVulnerabilityComment.Vulnerability.Comments.Edges, func(e *testclient.UpdateVulnerabilityComment_UpdateVulnerabilityComment_Vulnerability_Comments_Edges, _ int) bool {
			return e.Node.ID == ownerComment.ID
		})
		assert.Assert(t, is.Len(edited, 1))
		assert.Check(t, is.Equal("edited comment from the owner", edited[0].Node.Text))
	})

	t.Run("user in another org cannot update the comment", func(t *testing.T) {
		_, err := suite.Client.API.UpdateVulnerabilityComment(th.SharedTestUser2.UserCtx, ownerComment.ID, testclient.UpdateNoteInput{
			Text: lo.ToPtr("hijacked"),
		}, nil)
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	t.Run("owner can delete the comment", func(t *testing.T) {
		_, err := suite.Client.API.UpdateVulnerability(th.SharedTestUser1.UserCtx, vuln.ID, testclient.UpdateVulnerabilityInput{
			DeleteComment: &ownerComment.ID,
		})
		assert.NilError(t, err)

		exists, err := suite.Client.DB.Note.Query().Where(note.ID(ownerComment.ID)).Exist(th.SharedTestUser1.UserCtx)
		assert.NilError(t, err)
		assert.Check(t, !exists)
	})

	(&th.Cleanup[*generated.VulnerabilityDeleteOne]{Client: suite.Client.DB.Vulnerability, ID: vuln.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCommentsForFinding(t *testing.T) {
	finding := createFinding(t, th.SharedTestUser1.UserCtx, "Commentable Finding")

	testCases := []struct {
		name        string
		client      *testclient.TestClient
		ctx         context.Context
		comment     string
		expectedErr string
	}{
		{
			name:    "org owner can comment",
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
			comment: ownerCommentText,
		},
		{
			name:    "view only user can comment",
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
			comment: "comment from a view only user",
		},
		{
			name:    "pat can comment",
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
			comment: "comment from a pat",
		},
		{
			name:    "api token can comment",
			client:  suite.Client.APIWithToken,
			ctx:     context.Background(),
			comment: "comment from an api token",
		},
		{
			name:        "user in another org cannot comment",
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			comment:     "comment from another org",
			expectedErr: th.NotAuthorizedErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateFinding(tc.ctx, finding.ID, testclient.UpdateFindingInput{
				AddComment: &testclient.CreateNoteInput{Text: tc.comment},
			})

			expectedCount := 1
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				expectedCount = 0
			} else {
				assert.NilError(t, err)
				assert.Check(t, is.Equal(finding.ID, resp.UpdateFinding.Finding.ID))
			}

			count, err := suite.Client.DB.Note.Query().
				Where(note.HasFindingWith(entfinding.ID(finding.ID)), note.Text(tc.comment)).
				Count(th.SharedTestUser1.UserCtx)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(expectedCount, count))
		})
	}

	viewerDenied := []struct {
		name  string
		input testclient.UpdateFindingInput
	}{
		{name: "set a field", input: testclient.UpdateFindingInput{DisplayName: lo.ToPtr("renamed by viewer")}},
		{name: "clear a field", input: testclient.UpdateFindingInput{ClearDisplayName: lo.ToPtr(true)}},
		{name: "clear all comments", input: testclient.UpdateFindingInput{ClearComments: lo.ToPtr(true)}},
		{name: "add a comment while clearing a field", input: testclient.UpdateFindingInput{
			AddComment:       &testclient.CreateNoteInput{Text: "sneaky comment"},
			ClearDisplayName: lo.ToPtr(true),
		}},
	}

	for _, tc := range viewerDenied {
		t.Run("view only user cannot "+tc.name, func(t *testing.T) {
			_, err := suite.Client.API.UpdateFinding(th.SharedViewOnlyUser.UserCtx, finding.ID, tc.input)
			assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)
		})
	}

	t.Run("view only user cannot delete the finding", func(t *testing.T) {
		_, err := suite.Client.API.DeleteFinding(th.SharedViewOnlyUser.UserCtx, finding.ID)
		assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)
	})

	unchanged, err := suite.Client.DB.Finding.Query().Where(entfinding.ID(finding.ID)).WithComments().Only(th.SharedTestUser1.UserCtx)
	assert.NilError(t, err)
	assert.Check(t, unchanged.DisplayName != "")
	assert.Check(t, is.Len(unchanged.Edges.Comments, 4))

	ownerComment, err := suite.Client.DB.Note.Query().
		Where(note.HasFindingWith(entfinding.ID(finding.ID)), note.Text(ownerCommentText)).
		Only(th.SharedTestUser1.UserCtx)
	assert.NilError(t, err)

	t.Run("view only user cannot detach another user's comment", func(t *testing.T) {
		_, err := suite.Client.API.UpdateFinding(th.SharedViewOnlyUser.UserCtx, finding.ID, testclient.UpdateFindingInput{
			RemoveCommentIDs: []string{ownerComment.ID},
		})
		assert.Check(t, err != nil)

		stillAttached, err := suite.Client.DB.Note.Query().Where(note.ID(ownerComment.ID), note.HasFinding()).Exist(th.SharedTestUser1.UserCtx)
		assert.NilError(t, err)
		assert.Check(t, stillAttached)
	})

	t.Run("owner can update their comment", func(t *testing.T) {
		resp, err := suite.Client.API.UpdateFindingComment(th.SharedTestUser1.UserCtx, ownerComment.ID, testclient.UpdateNoteInput{
			Text: lo.ToPtr("edited comment from the owner"),
		}, nil)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(finding.ID, resp.UpdateFindingComment.Finding.ID))

		edited := lo.Filter(resp.UpdateFindingComment.Finding.Comments.Edges, func(e *testclient.UpdateFindingComment_UpdateFindingComment_Finding_Comments_Edges, _ int) bool {
			return e.Node.ID == ownerComment.ID
		})
		assert.Assert(t, is.Len(edited, 1))
		assert.Check(t, is.Equal("edited comment from the owner", edited[0].Node.Text))
	})

	t.Run("user in another org cannot update the comment", func(t *testing.T) {
		_, err := suite.Client.API.UpdateFindingComment(th.SharedTestUser2.UserCtx, ownerComment.ID, testclient.UpdateNoteInput{
			Text: lo.ToPtr("hijacked"),
		}, nil)
		assert.ErrorContains(t, err, th.NotFoundErrorMsg)
	})

	t.Run("owner can delete the comment", func(t *testing.T) {
		_, err := suite.Client.API.UpdateFinding(th.SharedTestUser1.UserCtx, finding.ID, testclient.UpdateFindingInput{
			DeleteComment: &ownerComment.ID,
		})
		assert.NilError(t, err)

		exists, err := suite.Client.DB.Note.Query().Where(note.ID(ownerComment.ID)).Exist(th.SharedTestUser1.UserCtx)
		assert.NilError(t, err)
		assert.Check(t, !exists)
	})

	(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: finding.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
