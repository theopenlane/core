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

func TestQueryFindingControl(t *testing.T) {
	finding := createFinding(t, sharedTestUser1.UserCtx, "Query Finding Control Finding")
	control := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	resp, err := suite.client.api.CreateFindingControl(sharedTestUser1.UserCtx, testclient.CreateFindingControlInput{
		FindingID: finding.ID,
		ControlID: control.ID,
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	fc := resp.CreateFindingControl.FindingControl

	testCases := []struct {
		name     string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:   "happy path, owner queries finding control",
			client: suite.client.api,
			ctx:    sharedTestUser1.UserCtx,
		},
		{
			name:   "happy path, admin queries finding control",
			client: suite.client.api,
			ctx:    sharedAdminUser.UserCtx,
		},
		{
			name:     "no access, user of different org",
			client:   suite.client.api,
			ctx:      sharedTestUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "no access, api token of different org",
			client:   suite.client.apiWithTokenOrg2,
			ctx:      context.Background(),
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetFindingControlByID(tc.ctx, fc.ID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(fc.ID, resp.FindingControl.ID))
		})
	}

	(&Cleanup[*generated.FindingControlDeleteOne]{client: suite.client.db.FindingControl, ID: fc.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: finding.ID}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestQueryFindingControls(t *testing.T) {
	finding1 := createFinding(t, sharedTestUser1.UserCtx, "List Finding Control Finding 1")
	control1 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)

	createFC := func(ctx context.Context, findingID, controlID string) string {
		resp, err := suite.client.api.CreateFindingControl(ctx, testclient.CreateFindingControlInput{
			FindingID: findingID,
			ControlID: controlID,
		})
		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		return resp.CreateFindingControl.FindingControl.ID
	}

	fc1 := createFC(sharedTestUser1.UserCtx, finding1.ID, control1.ID)
	fc2 := createFC(sharedTestUser1.UserCtx, finding1.ID, control2.ID)

	// finding control that belongs to a different organization
	finding2 := createFinding(t, sharedTestUser2.UserCtx, "List Finding Control Finding 2")
	control3 := (&ControlBuilder{client: suite.client}).MustNew(sharedTestUser2.UserCtx, t)
	fc3 := createFC(sharedTestUser2.UserCtx, finding2.ID, control3.ID)

	t.Run("owner only sees finding controls within their organization", func(t *testing.T) {
		resp, err := suite.client.api.GetAllFindingControls(sharedTestUser1.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		ids := make([]string, 0, len(resp.FindingControls.Edges))
		for _, edge := range resp.FindingControls.Edges {
			ids = append(ids, edge.Node.ID)
		}

		assert.Check(t, lo.Contains(ids, fc1))
		assert.Check(t, lo.Contains(ids, fc2))
		assert.Check(t, !lo.Contains(ids, fc3))
	})

	t.Run("user of another org only sees their own organization's finding controls", func(t *testing.T) {
		resp, err := suite.client.api.GetAllFindingControls(sharedTestUser2.UserCtx)

		assert.NilError(t, err)
		assert.Assert(t, resp != nil)

		ids := make([]string, 0, len(resp.FindingControls.Edges))
		for _, edge := range resp.FindingControls.Edges {
			ids = append(ids, edge.Node.ID)
		}

		assert.Check(t, lo.Contains(ids, fc3))
		assert.Check(t, !lo.Contains(ids, fc1))
		assert.Check(t, !lo.Contains(ids, fc2))
	})

	(&Cleanup[*generated.FindingControlDeleteOne]{client: suite.client.db.FindingControl, IDs: []string{fc1, fc2}}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.FindingControlDeleteOne]{client: suite.client.db.FindingControl, ID: fc3}).MustDelete(sharedTestUser2.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control1.ID, control2.ID}}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control3.ID}).MustDelete(sharedTestUser2.UserCtx, t)
	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: finding1.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: finding2.ID}).MustDelete(sharedTestUser2.UserCtx, t)
}

func TestMutationCreateFindingControl(t *testing.T) {
	finding := createFinding(t, sharedTestUser1.UserCtx, "Create Finding Control Finding")

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
	finding := createFinding(t, sharedTestUser1.UserCtx, "Update Finding Control Finding")

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
	finding := createFinding(t, sharedTestUser1.UserCtx, "Delete Finding Control Finding")

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
	finding := createFinding(t, sharedTestUser1.UserCtx, "Bulk Delete Finding Control Finding")

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
