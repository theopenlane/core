package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestMutationCreateFindingControl(t *testing.T) {
	finding := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Create Finding Control Finding")

	control1 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control3 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control4 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		controlID   string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:      "happy path, owner creates finding control",
			controlID: control1.ID,
			client:    suite.client.api,
			ctx:       sharedTestUser1.UserCtx,
		},
		{
			name:      "happy path, admin creates finding control",
			controlID: control2.ID,
			client:    suite.client.api,
			ctx:       sharedAdminUser.UserCtx,
		},
		{
			name:        "not authorized, view only member creates finding control",
			controlID:   control3.ID,
			client:      suite.client.api,
			ctx:         sharedViewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:      "happy path, auditor creates finding control",
			controlID: control4.ID,
			client:    suite.client.api,
			ctx:       sharedAuditorUser.UserCtx,
		},
	}

	fcIDs := make([]string, 0, len(testCases))

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateFindingControl(tc.ctx, testclient.CreateFindingControlInput{
				FindingID: finding.ID,
				ControlID: tc.controlID,
			})
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.controlID, resp.CreateFindingControl.FindingControl.ControlID))

			fcIDs = append(fcIDs, resp.CreateFindingControl.FindingControl.ID)
		})
	}

	(&Cleanup[*generated.FindingControlDeleteOne]{client: suite.client.db.FindingControl, IDs: fcIDs}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID, control4.ID}}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: finding.ID}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestMutationUpdateFindingControl(t *testing.T) {
	finding := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Update Finding Control Finding")

	control1 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control3 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control4 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	createFC := func(ctx context.Context, controlID string) string {
		resp, err := suite.client.api.CreateFindingControl(ctx, testclient.CreateFindingControlInput{
			FindingID: finding.ID,
			ControlID: controlID,
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		return resp.CreateFindingControl.FindingControl.ID
	}

	fc1 := createFC(sharedTestUser1.UserCtx, control1.ID)
	fc2 := createFC(sharedTestUser1.UserCtx, control2.ID)
	fc3 := createFC(sharedTestUser1.UserCtx, control3.ID)
	fc4 := createFC(sharedTestUser1.UserCtx, control4.ID)

	testCases := []struct {
		name        string
		id          string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:   "happy path, owner updates finding control",
			id:     fc1,
			client: suite.client.api,
			ctx:    sharedTestUser1.UserCtx,
		},
		{
			name:   "happy path, admin updates finding control",
			id:     fc2,
			client: suite.client.api,
			ctx:    sharedAdminUser.UserCtx,
		},
		{
			name:        "not authorized, view only member cannot update a finding control",
			id:          fc3,
			client:      suite.client.api,
			ctx:         sharedViewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:   "happy path, auditor updates finding control",
			id:     fc4,
			client: suite.client.api,
			ctx:    sharedAuditorUser.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateFindingControl(tc.ctx, tc.id, testclient.UpdateFindingControlInput{
				Source: lo.ToPtr("updated-source"),
			})
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal("updated-source", *resp.UpdateFindingControl.FindingControl.Source))
		})
	}

	(&Cleanup[*generated.FindingControlDeleteOne]{client: suite.client.db.FindingControl, IDs: []string{fc1, fc2, fc3, fc4}}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID, control4.ID}}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: finding.ID}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestMutationDeleteFindingControl(t *testing.T) {
	finding := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Delete Finding Control Finding")

	control1 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control3 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control4 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	createFC := func(ctx context.Context, controlID string) string {
		resp, err := suite.client.api.CreateFindingControl(ctx, testclient.CreateFindingControlInput{
			FindingID: finding.ID,
			ControlID: controlID,
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		return resp.CreateFindingControl.FindingControl.ID
	}

	fc1 := createFC(sharedTestUser1.UserCtx, control1.ID)
	fc2 := createFC(sharedTestUser1.UserCtx, control2.ID)
	fc3 := createFC(sharedTestUser1.UserCtx, control3.ID)
	fc4 := createFC(sharedTestUser1.UserCtx, control4.ID)

	testCases := []struct {
		name        string
		id          string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:   "happy path, owner deletes finding control",
			id:     fc1,
			client: suite.client.api,
			ctx:    sharedTestUser1.UserCtx,
		},
		{
			name:   "happy path, admin deletes finding control",
			id:     fc2,
			client: suite.client.api,
			ctx:    sharedAdminUser.UserCtx,
		},
		{
			name:        "not authorized, view only member cannot delete a finding control",
			id:          fc3,
			client:      suite.client.api,
			ctx:         sharedViewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:   "happy path, auditor deletes finding control",
			id:     fc4,
			client: suite.client.api,
			ctx:    sharedAuditorUser.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteFindingControl(tc.ctx, tc.id)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.id, resp.DeleteFindingControl.DeletedID))
		})
	}

	// fc3 was never deleted since the view only member wasn't authorized
	(&Cleanup[*generated.FindingControlDeleteOne]{client: suite.client.db.FindingControl, ID: fc3}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID, control4.ID}}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: finding.ID}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestMutationDeleteBulkFindingControl(t *testing.T) {
	finding := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Bulk Delete Finding Control Finding")

	control1 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control3 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	createFC := func(ctx context.Context, controlID string) string {
		resp, err := suite.client.api.CreateFindingControl(ctx, testclient.CreateFindingControlInput{
			FindingID: finding.ID,
			ControlID: controlID,
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		return resp.CreateFindingControl.FindingControl.ID
	}

	fc1 := createFC(sharedTestUser1.UserCtx, control1.ID)
	fc2 := createFC(sharedTestUser1.UserCtx, control2.ID)
	fc3 := createFC(sharedTestUser1.UserCtx, control3.ID)

	testCases := []struct {
		name                 string
		idsToDelete          []string
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
		expectedDeletedCount int
	}{
		{
			name:                 "happy path, owner deletes finding control",
			idsToDelete:          []string{fc1},
			client:               suite.client.api,
			ctx:                  sharedTestUser1.UserCtx,
			expectedDeletedCount: 1,
		},
		{
			name:                 "happy path, admin deletes finding control",
			idsToDelete:          []string{fc2},
			client:               suite.client.api,
			ctx:                  sharedAdminUser.UserCtx,
			expectedDeletedCount: 1,
		},
		{
			name:                 "not authorized, view only member cannot delete a finding control without edit access to the control",
			idsToDelete:          []string{fc3},
			client:               suite.client.api,
			ctx:                  sharedViewOnlyUser.UserCtx,
			expectedDeletedCount: 0,
		},
		{
			name:                 "happy path, auditor deletes a finding control they created",
			idsToDelete:          []string{fc3},
			client:               suite.client.api,
			ctx:                  sharedAuditorUser.UserCtx,
			expectedDeletedCount: 1,
		},
	}

	for _, tc := range testCases {
		t.Run("Bulk Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteBulkFindingControl(tc.ctx, tc.idsToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.DeleteBulkFindingControl.DeletedIDs, tc.expectedDeletedCount))
		})
	}

	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID, control3.ID}}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: finding.ID}).MustDelete(sharedTestUser1.UserCtx, t)
}
