package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestMutationCreateFinding(t *testing.T) {
	finding := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Existing Finding")

	editingGroup := (&GroupBuilder{client: suite.client}).MustNew(sharedTestUser1.UserCtx, t)
	groupMember := (&GroupMemberBuilder{client: suite.client, GroupID: editingGroup.ID, UserID: sharedAuditorUser.ID}).MustNew(sharedTestUser1.UserCtx, t)

	control := (&ControlBuilder{client: suite.client, ControlEditorGroupID: editingGroup.ID}).MustNew(sharedTestUser1.UserCtx, t)

	resp, err := suite.client.api.CreateReview(sharedAuditorUser.UserCtx, testclient.CreateReviewInput{
		Title: "Auditor finding for review",
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	tt := []struct {
		name        string
		request     testclient.CreateFindingInput
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			request: testclient.CreateFindingInput{
				DisplayName:     lo.ToPtr("Finding"),
				ExternalID:      lo.ToPtr("finding-" + ulids.New().String()),
				ExternalOwnerID: lo.ToPtr("external-owner"),
				OwnerID:         &sharedTestUser1.OrganizationID,
			},
		},
		{
			name: "happy path, minimal input",
			request: testclient.CreateFindingInput{
				DisplayName: lo.ToPtr("Finding"),
				ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
				OwnerID:     &sharedTestUser1.OrganizationID,
			},
		},
		{
			name: "happy path, auditor",
			request: testclient.CreateFindingInput{
				DisplayName: lo.ToPtr("Auditor"),
				ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
				ControlIDs:  []string{control.ID},
				OwnerID:     &sharedTestUser1.OrganizationID,
			},
			ctx: sharedAuditorUser.UserCtx,
		},
		{
			name: "happy path, auditor under review",
			request: testclient.CreateFindingInput{
				DisplayName: lo.ToPtr("Auditor Review Finding"),
				ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
				ReviewIDs:   []string{resp.CreateReview.Review.ID},
				OwnerID:     &sharedTestUser1.OrganizationID,
			},
			ctx: sharedAuditorUser.UserCtx,
		},
		{
			name: "unauthorized auditor",
			request: testclient.CreateFindingInput{
				DisplayName: lo.ToPtr("Auditor Finding Without Parent"),
				ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
				OwnerID:     &sharedTestUser1.OrganizationID,
			},
			ctx:         sharedAuditorUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range tt {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = sharedTestUser1.UserCtx
			}

			resp, err := suite.client.api.CreateFinding(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, len(resp.CreateFinding.Finding.ID) != 0)
			assert.Check(t, is.Equal(*tc.request.DisplayName, *resp.CreateFinding.Finding.DisplayName))
			assert.Check(t, is.Equal(*tc.request.ExternalID, *resp.CreateFinding.Finding.ExternalID))
			assert.Check(t, is.Equal(sharedTestUser1.OrganizationID, *resp.CreateFinding.Finding.OwnerID))

			if tc.request.Severity != nil {
				assert.Check(t, is.Equal(*tc.request.Severity, *resp.CreateFinding.Finding.Severity))
			}

			(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: resp.CreateFinding.Finding.ID}).MustDelete(sharedTestUser1.UserCtx, t)
		})
	}

	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: finding.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.GroupMembershipDeleteOne]{client: suite.client.db.GroupMembership, ID: groupMember.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, ID: editingGroup.ID}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.ReviewDeleteOne]{client: suite.client.db.Review, ID: resp.CreateReview.Review.ID}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestMutationCreateBulkFinding(t *testing.T) {

	finding := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Existing Bulk Finding")

	creationRequests := []*testclient.CreateFindingInput{
		{
			DisplayName:     lo.ToPtr("Bulk Finding 1"),
			ExternalID:      lo.ToPtr("finding-" + ulids.New().String()),
			ExternalOwnerID: lo.ToPtr("external-owner"),
			OwnerID:         &sharedTestUser1.OrganizationID,
		},
		{
			DisplayName:     lo.ToPtr("Bulk Finding 2"),
			ExternalID:      lo.ToPtr("finding-" + ulids.New().String()),
			ExternalOwnerID: lo.ToPtr("external-owner"),
			OwnerID:         &sharedTestUser1.OrganizationID,
		},
		{
			DisplayName:     lo.ToPtr("Bulk Finding 3"),
			ExternalID:      lo.ToPtr("finding-" + ulids.New().String()),
			ExternalOwnerID: lo.ToPtr("external-owner"),
			OwnerID:         &sharedTestUser1.OrganizationID,
		},
	}

	tt := []struct {
		name          string
		requests      []*testclient.CreateFindingInput
		ctx           context.Context
		expectedErr   string
		expectedCount int
	}{
		{
			name:          "happy path",
			requests:      creationRequests,
			expectedCount: 3,
		},
		{
			name:        "empty input",
			requests:    []*testclient.CreateFindingInput{},
			expectedErr: "input is required",
		},
		{
			name: "auditor without parent is not authorized",
			requests: []*testclient.CreateFindingInput{
				{
					DisplayName: lo.ToPtr("Auditor Bulk Finding Without Parent"),
					ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
					OwnerID:     &sharedTestUser1.OrganizationID,
				},
			},
			ctx:         sharedAuditorUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
	}

	for _, tc := range tt {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = sharedTestUser1.UserCtx
			}

			resp, err := suite.client.api.CreateBulkFinding(tc.ctx, tc.requests)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.CreateBulkFinding.Findings, tc.expectedCount))

			ids := make([]string, 0, len(resp.CreateBulkFinding.Findings))
			for i, finding := range resp.CreateBulkFinding.Findings {
				assert.Check(t, is.Equal(*tc.requests[i].DisplayName, *finding.DisplayName))
				assert.Check(t, is.Equal(*tc.requests[i].ExternalID, *finding.ExternalID))
				ids = append(ids, finding.ID)
			}

			(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, IDs: ids}).MustDelete(sharedTestUser1.UserCtx, t)
		})
	}

	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: finding.ID}).MustDelete(sharedTestUser1.UserCtx, t)
}

func TestMutationUpdateFinding(t *testing.T) {
	finding := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Update Finding")

	duplicateFinding := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Duplicate Update Finding")

	anotherOrgFinding := createFinding(t, sharedTestUser2.UserCtx, sharedTestUser2.OrganizationID, "Unauthorized Update Finding")

	testCases := []struct {
		name        string
		id          string
		request     testclient.UpdateFindingInput
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path",
			id:   finding.ID,
			request: testclient.UpdateFindingInput{
				DisplayName: lo.ToPtr("Updated Finding"),
				Description: lo.ToPtr("Updated description"),
				Severity:    lo.ToPtr("critical"),
				Open:        lo.ToPtr(false),
				Tags:        []string{"updated", "finding"},
			},
		},
		{
			name: "append list fields",
			id:   finding.ID,
			request: testclient.UpdateFindingInput{
				AppendCategories:       []string{"runtime"},
				AppendReferences:       []string{"https://example.com/finding"},
				AppendStepsToReproduce: []string{"deploy vulnerable config"},
				AppendTargets:          []string{"service-a"},
			},
		},
		{
			name: "duplicate external id",
			id:   duplicateFinding.ID,
			request: testclient.UpdateFindingInput{
				ExternalID:      finding.ExternalID,
				ExternalOwnerID: finding.ExternalOwnerID,
			},
			expectedErr: "already exists",
		},
		{
			name: "auditor cannot update existing finding",
			id:   finding.ID,
			request: testclient.UpdateFindingInput{
				DisplayName: lo.ToPtr("Auditor Update"),
			},
			ctx:         sharedAuditorUser.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "not authorized, valid id",
			id:   anotherOrgFinding.ID,
			request: testclient.UpdateFindingInput{
				DisplayName: lo.ToPtr("Unauthorized Update"),
			},
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "invalid id",
			id:   "invalid",
			request: testclient.UpdateFindingInput{
				DisplayName: lo.ToPtr("Invalid ID Update"),
			},
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "unknown id",
			id:   ulids.New().String(),
			request: testclient.UpdateFindingInput{
				DisplayName: lo.ToPtr("Unknown ID Update"),
			},
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = sharedTestUser1.UserCtx
			}

			resp, err := suite.client.api.UpdateFinding(tc.ctx, tc.id, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.id, resp.UpdateFinding.Finding.ID))

			if tc.request.DisplayName != nil {
				assert.Check(t, is.Equal(*tc.request.DisplayName, *resp.UpdateFinding.Finding.DisplayName))
			}

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateFinding.Finding.Description))
			}

			if tc.request.Severity != nil {
				assert.Check(t, is.Equal(*tc.request.Severity, *resp.UpdateFinding.Finding.Severity))
			}

			if tc.request.Open != nil {
				assert.Check(t, is.Equal(*tc.request.Open, *resp.UpdateFinding.Finding.Open))
			}

			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.UpdateFinding.Finding.Tags))
			}

			if len(tc.request.AppendCategories) > 0 {
				assert.Check(t, is.Contains(resp.UpdateFinding.Finding.Categories, tc.request.AppendCategories[0]))
			}

			if len(tc.request.AppendReferences) > 0 {
				assert.Check(t, is.Contains(resp.UpdateFinding.Finding.References, tc.request.AppendReferences[0]))
			}

			if len(tc.request.AppendStepsToReproduce) > 0 {
				assert.Check(t, is.Contains(resp.UpdateFinding.Finding.StepsToReproduce, tc.request.AppendStepsToReproduce[0]))
			}

			if len(tc.request.AppendTargets) > 0 {
				assert.Check(t, is.Contains(resp.UpdateFinding.Finding.Targets, tc.request.AppendTargets[0]))
			}
		})
	}

	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, IDs: []string{finding.ID, duplicateFinding.ID}}).MustDelete(sharedTestUser1.UserCtx, t)
	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: anotherOrgFinding.ID}).MustDelete(sharedTestUser2.UserCtx, t)
}

func TestMutationDeleteFinding(t *testing.T) {
	finding := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Delete Finding")
	anotherOrgFinding := createFinding(t, sharedTestUser2.UserCtx, sharedTestUser2.OrganizationID, "Unauthorized Delete Finding")

	testCases := []struct {
		name        string
		idToDelete  string
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "auditor cannot delete existing finding",
			idToDelete:  finding.ID,
			ctx:         sharedAuditorUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not authorized, valid id",
			idToDelete:  anotherOrgFinding.ID,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "invalid id",
			idToDelete:  "invalid",
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "unknown id",
			idToDelete:  ulids.New().String(),
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path",
			idToDelete: finding.ID,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  finding.ID,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = sharedTestUser1.UserCtx
			}

			resp, err := suite.client.api.DeleteFinding(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteFinding.DeletedID))

			_, err = suite.client.api.GetFindingByID(sharedTestUser1.UserCtx, tc.idToDelete)
			assert.ErrorContains(t, err, notFoundErrorMsg)
		})
	}

	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: anotherOrgFinding.ID}).MustDelete(sharedTestUser2.UserCtx, t)
}

func TestMutationDeleteBulkFinding(t *testing.T) {
	finding1 := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Bulk Delete Finding 1")
	finding2 := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Bulk Delete Finding 2")
	finding3 := createFinding(t, sharedTestUser1.UserCtx, sharedTestUser1.OrganizationID, "Bulk Delete Finding 3")

	orgFinding := createFinding(t, sharedTestUser2.UserCtx, sharedTestUser2.OrganizationID, "Unauthorized Bulk Delete Finding")

	testCases := []struct {
		name                 string
		idsToDelete          []string
		ctx                  context.Context
		expectedErr          string
		expectedDeletedCount int
	}{
		{
			name:                 "auditor cannot delete existing finding",
			idsToDelete:          []string{finding1.ID},
			ctx:                  sharedAuditorUser.UserCtx,
			expectedDeletedCount: 0,
		},
		{
			name:                 "happy path",
			idsToDelete:          []string{finding1.ID, finding2.ID, finding3.ID},
			expectedDeletedCount: 3,
		},
		{
			name:                 "not authorized, valid id",
			idsToDelete:          []string{orgFinding.ID},
			expectedDeletedCount: 0,
		},
		{
			name:        "empty ids",
			idsToDelete: []string{},
			expectedErr: "ids is required",
		},
		{
			name:                 "invalid id",
			idsToDelete:          []string{ulids.New().String()},
			expectedDeletedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("Bulk Delete "+tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = sharedTestUser1.UserCtx
			}

			resp, err := suite.client.api.DeleteBulkFinding(tc.ctx, tc.idsToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.DeleteBulkFinding.DeletedIDs, tc.expectedDeletedCount))

			for _, id := range resp.DeleteBulkFinding.DeletedIDs {
				assert.Check(t, is.Contains(tc.idsToDelete, id))

				_, err := suite.client.api.GetFindingByID(sharedTestUser1.UserCtx, id)
				assert.ErrorContains(t, err, notFoundErrorMsg)
			}
		})
	}

	(&Cleanup[*generated.FindingDeleteOne]{client: suite.client.db.Finding, ID: orgFinding.ID}).MustDelete(sharedTestUser2.UserCtx, t)
}

func createFinding(t *testing.T, ctx context.Context, ownerID string, displayName string) *testclient.CreateFinding_CreateFinding_Finding {
	t.Helper()

	resp, err := suite.client.api.CreateFinding(ctx, testclient.CreateFindingInput{
		DisplayName:     &displayName,
		ExternalID:      lo.ToPtr("finding-" + ulids.New().String()),
		ExternalOwnerID: lo.ToPtr("external-owner"),
		OwnerID:         &ownerID,
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	return &resp.CreateFinding.Finding
}
