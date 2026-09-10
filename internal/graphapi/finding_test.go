package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestMutationCreateFinding(t *testing.T) {
	finding := createFinding(t, th.SharedTestUser1.UserCtx, "Existing Finding")

	editingGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	groupMember := (&th.GroupMemberBuilder{Client: suite.Client, GroupID: editingGroup.ID, UserID: th.SharedAuditorUser.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	firstControlFindingResp, err := suite.Client.API.CreateFinding(th.SharedTestUser1.UserCtx, testclient.CreateFindingInput{
		DisplayName: lo.ToPtr("First Control Finding"),
		ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
	})
	assert.NilError(t, err)
	assert.Check(t, firstControlFindingResp != nil)

	resp, err := suite.Client.API.CreateReview(th.SharedAuditorUser.UserCtx, testclient.CreateReviewInput{
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
				DisplayName:       lo.ToPtr("Finding"),
				ExternalID:        lo.ToPtr("finding-" + ulids.New().String()),
				ExternalOwnerID:   lo.ToPtr("external-owner"),
				OwnerID:           &th.SharedTestUser1.OrganizationID,
				AssignedToUserID:  &th.SharedViewOnlyUser.ID,
				AssignedToGroupID: &th.SharedTestUser1.GroupID,
				ReviewedByUserID:  &th.SharedViewOnlyUser.ID,
				ReviewedByGroupID: &th.SharedTestUser1.GroupID,
			},
		},
		{
			name: "happy path, minimal input",
			request: testclient.CreateFindingInput{
				DisplayName: lo.ToPtr("Finding"),
				ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
			},
		},
		{
			name: "happy path, auditor",
			request: testclient.CreateFindingInput{
				DisplayName: lo.ToPtr("Auditor"),
				ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
			},
			ctx: th.SharedAuditorUser.UserCtx,
		},
		{
			name: "happy path, auditor under review",
			request: testclient.CreateFindingInput{
				DisplayName: lo.ToPtr("Auditor Review Finding"),
				ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
				ReviewIDs:   []string{resp.CreateReview.Review.ID},
			},
			ctx: th.SharedAuditorUser.UserCtx,
		},
	}

	for _, tc := range tt {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = th.SharedTestUser1.UserCtx
			}

			resp, err := suite.Client.API.CreateFinding(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, len(resp.CreateFinding.Finding.ID) != 0)
			assert.Check(t, is.Equal(*tc.request.DisplayName, *resp.CreateFinding.Finding.DisplayName))
			assert.Check(t, is.Equal(*tc.request.ExternalID, *resp.CreateFinding.Finding.ExternalID))
			assert.Check(t, is.Equal(th.SharedTestUser1.OrganizationID, *resp.CreateFinding.Finding.OwnerID))

			if tc.request.Severity != nil {
				assert.Check(t, is.Equal(*tc.request.Severity, *resp.CreateFinding.Finding.Severity))
			}

			if tc.request.AssignedToUserID != nil {
				assert.Check(t, is.Equal(*tc.request.AssignedToUserID, *resp.CreateFinding.Finding.AssignedToUserID))
			} else {
				assert.Check(t, *resp.CreateFinding.Finding.AssignedToUserID == "", "expected AssignedToUserID to be empty but was %v", resp.CreateFinding.Finding.AssignedToUserID)
			}

			if tc.request.AssignedToGroupID != nil {
				assert.Check(t, is.Equal(*tc.request.AssignedToGroupID, *resp.CreateFinding.Finding.AssignedToGroupID))
			} else {
				assert.Check(t, *resp.CreateFinding.Finding.AssignedToGroupID == "", "expected AssignedToGroupID to be empty but was %v", resp.CreateFinding.Finding.AssignedToGroupID)
			}

			if tc.request.ReviewedByUserID != nil {
				assert.Check(t, is.Equal(*tc.request.ReviewedByUserID, *resp.CreateFinding.Finding.ReviewedByUserID))
			} else {
				assert.Check(t, *resp.CreateFinding.Finding.ReviewedByUserID == "", "expected ReviewedByUserID to be empty but was %v", resp.CreateFinding.Finding.ReviewedByUserID)
			}

			if tc.request.ReviewedByGroupID != nil {
				assert.Check(t, is.Equal(*tc.request.ReviewedByGroupID, *resp.CreateFinding.Finding.ReviewedByGroupID))
			} else {
				assert.Check(t, *resp.CreateFinding.Finding.ReviewedByGroupID == "", "expected ReviewedByGroupID to be empty but was %v", resp.CreateFinding.Finding.ReviewedByGroupID)
			}

			(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: resp.CreateFinding.Finding.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}

	(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: []string{finding.ID, firstControlFindingResp.CreateFinding.Finding.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.GroupMembershipDeleteOne]{Client: suite.Client.DB.GroupMembership, ID: groupMember.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, ID: editingGroup.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ReviewDeleteOne]{Client: suite.Client.DB.Review, ID: resp.CreateReview.Review.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
func TestMutationCreateFindingUnderLinkedObject(t *testing.T) {
	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	control2 := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	org2Control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	resp, err := suite.Client.API.CreateReview(th.SharedAuditorUser.UserCtx, testclient.CreateReviewInput{
		Title: "Review for auditor",
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	tt := []struct {
		name        string
		request     testclient.CreateFindingInput
		expectedErr string
	}{
		{
			name: "auditor can create under review",
			request: testclient.CreateFindingInput{
				DisplayName: lo.ToPtr("Finding review"),
				ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
				ReviewIDs:   []string{resp.CreateReview.Review.ID},
			},
		},
		{
			name: "auditor can create without linking to a parent object",
			request: testclient.CreateFindingInput{
				DisplayName: lo.ToPtr("finding not linked to an object"),
				ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.CreateFinding(th.SharedAuditorUser.UserCtx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, len(resp.CreateFinding.Finding.ID) != 0)
			assert.Check(t, is.Equal(*tc.request.DisplayName, *resp.CreateFinding.Finding.DisplayName))

			(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: resp.CreateFinding.Finding.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}

	// controls are linked to a finding after creation via the finding_control join, not
	// through CreateFindingInput, so exercise CreateFindingControl separately
	t.Run("auditor can create finding control under multiple controls", func(t *testing.T) {
		findingResp, err := suite.Client.API.CreateFinding(th.SharedAuditorUser.UserCtx, testclient.CreateFindingInput{
			DisplayName: lo.ToPtr("Finding multiple controls"),
			ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
			ReviewIDs:   []string{resp.CreateReview.Review.ID},
		})
		assert.NilError(t, err)
		assert.Assert(t, findingResp != nil)

		findingID := findingResp.CreateFinding.Finding.ID

		fcIDs := make([]string, 0, 2)

		for _, c := range []string{control.ID, control2.ID} {
			fcResp, err := suite.Client.API.CreateFindingControl(th.SharedAuditorUser.UserCtx, testclient.CreateFindingControlInput{
				FindingID: findingID,
				ControlID: c,
			})
			assert.NilError(t, err)
			assert.Assert(t, fcResp != nil)
			assert.Check(t, is.Equal(c, fcResp.CreateFindingControl.FindingControl.ControlID))

			fcIDs = append(fcIDs, fcResp.CreateFindingControl.FindingControl.ID)
		}

		// confirm the finding is now linked to both controls, not just the last one created
		ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

		entFinding, err := suite.Client.DB.Finding.Get(ctx, findingID)
		assert.NilError(t, err)

		linkedControls, err := entFinding.QueryControls().All(ctx)
		assert.NilError(t, err)
		assert.Check(t, is.Len(linkedControls, 2))

		(&th.Cleanup[*generated.FindingControlDeleteOne]{Client: suite.Client.DB.FindingControl, IDs: fcIDs}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: findingID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	t.Run("auditor cannot create finding control link to unauthorized control", func(t *testing.T) {
		findingResp, err := suite.Client.API.CreateFinding(th.SharedAuditorUser.UserCtx, testclient.CreateFindingInput{
			DisplayName: lo.ToPtr("Finding linked to control from org 2"),
			ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
			ReviewIDs:   []string{resp.CreateReview.Review.ID},
		})
		assert.NilError(t, err)
		assert.Assert(t, findingResp != nil)

		findingID := findingResp.CreateFinding.Finding.ID

		_, err = suite.Client.API.CreateFindingControl(th.SharedAuditorUser.UserCtx, testclient.CreateFindingControlInput{
			FindingID: findingID,
			ControlID: org2Control.ID,
		})
		assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)

		(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: findingID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{control.ID, control2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: org2Control.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.ReviewDeleteOne]{Client: suite.Client.DB.Review, ID: resp.CreateReview.Review.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateBulkFinding(t *testing.T) {

	finding := createFinding(t, th.SharedTestUser1.UserCtx, "Existing Bulk Finding")

	creationRequests := []*testclient.CreateFindingInput{
		{
			DisplayName:     lo.ToPtr("Bulk Finding 1"),
			ExternalID:      lo.ToPtr("finding-" + ulids.New().String()),
			ExternalOwnerID: lo.ToPtr("external-owner"),
			OwnerID:         &th.SharedTestUser1.OrganizationID,
		},
		{
			DisplayName:     lo.ToPtr("Bulk Finding 2"),
			ExternalID:      lo.ToPtr("finding-" + ulids.New().String()),
			ExternalOwnerID: lo.ToPtr("external-owner"),
			OwnerID:         &th.SharedTestUser1.OrganizationID,
		},
		{
			DisplayName:     lo.ToPtr("Bulk Finding 3"),
			ExternalID:      lo.ToPtr("finding-" + ulids.New().String()),
			ExternalOwnerID: lo.ToPtr("external-owner"),
			OwnerID:         &th.SharedTestUser1.OrganizationID,
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
			name: "auditor without parent is authorized",
			requests: []*testclient.CreateFindingInput{
				{
					DisplayName: lo.ToPtr("Auditor Bulk Finding Without Parent"),
					ExternalID:  lo.ToPtr("finding-" + ulids.New().String()),
					OwnerID:     &th.SharedTestUser1.OrganizationID,
				},
			},
			ctx:           th.SharedAuditorUser.UserCtx,
			expectedCount: 1,
		},
	}

	for _, tc := range tt {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = th.SharedTestUser1.UserCtx
			}

			resp, err := suite.Client.API.CreateBulkFinding(tc.ctx, tc.requests)
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

			(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: ids}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}

	(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: finding.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationUpdateFinding(t *testing.T) {
	finding := createFinding(t, th.SharedTestUser1.UserCtx, "Update Finding")

	duplicateFinding := createFinding(t, th.SharedTestUser1.UserCtx, "Duplicate Update Finding")

	anotherOrgFinding := createFinding(t, th.SharedTestUser2.UserCtx, "Unauthorized Update Finding")

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
				DisplayName:       lo.ToPtr("Updated Finding"),
				Description:       lo.ToPtr("Updated description"),
				Severity:          lo.ToPtr("critical"),
				Open:              lo.ToPtr(false),
				Tags:              []string{"updated", "finding"},
				AssignedToUserID:  &th.SharedViewOnlyUser.ID,
				AssignedToGroupID: &th.SharedTestUser1.GroupID,
				ReviewedByUserID:  &th.SharedViewOnlyUser.ID,
				ReviewedByGroupID: &th.SharedTestUser1.GroupID,
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
			name: "auditor can update existing finding",
			id:   finding.ID,
			request: testclient.UpdateFindingInput{
				DisplayName: lo.ToPtr("Auditor Update"),
			},
			ctx: th.SharedAuditorUser.UserCtx,
		},
		{
			name: "not authorized, valid id",
			id:   anotherOrgFinding.ID,
			request: testclient.UpdateFindingInput{
				DisplayName: lo.ToPtr("Unauthorized Update"),
			},
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "invalid id",
			id:   "invalid",
			request: testclient.UpdateFindingInput{
				DisplayName: lo.ToPtr("Invalid ID Update"),
			},
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name: "unknown id",
			id:   ulids.New().String(),
			request: testclient.UpdateFindingInput{
				DisplayName: lo.ToPtr("Unknown ID Update"),
			},
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = th.SharedTestUser1.UserCtx
			}

			resp, err := suite.Client.API.UpdateFinding(tc.ctx, tc.id, tc.request)
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

			if tc.request.AssignedToUserID != nil {
				assert.Check(t, is.Equal(*tc.request.AssignedToUserID, *resp.UpdateFinding.Finding.AssignedToUserID))
			}

			if tc.request.AssignedToGroupID != nil {
				assert.Check(t, is.Equal(*tc.request.AssignedToGroupID, *resp.UpdateFinding.Finding.AssignedToGroupID))
			}

			if tc.request.ReviewedByUserID != nil {
				assert.Check(t, is.Equal(*tc.request.ReviewedByUserID, *resp.UpdateFinding.Finding.ReviewedByUserID))
			}

			if tc.request.ReviewedByGroupID != nil {
				assert.Check(t, is.Equal(*tc.request.ReviewedByGroupID, *resp.UpdateFinding.Finding.ReviewedByGroupID))
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

	(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: []string{finding.ID, duplicateFinding.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: anotherOrgFinding.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
}

func TestMutationDeleteFinding(t *testing.T) {
	finding := createFinding(t, th.SharedTestUser1.UserCtx, "Delete Finding")
	anotherOrgFinding := createFinding(t, th.SharedTestUser2.UserCtx, "Unauthorized Delete Finding")
	auditorFinding := createFinding(t, th.SharedTestUser1.UserCtx, "Auditor Delete Finding")

	testCases := []struct {
		name        string
		idToDelete  string
		ctx         context.Context
		expectedErr string
	}{
		{
			name:       "auditor can delete existing finding",
			idToDelete: auditorFinding.ID,
			ctx:        th.SharedAuditorUser.UserCtx,
		},
		{
			name:        "not authorized, valid id",
			idToDelete:  anotherOrgFinding.ID,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "invalid id",
			idToDelete:  "invalid",
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "unknown id",
			idToDelete:  ulids.New().String(),
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path",
			idToDelete: finding.ID,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  finding.ID,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = th.SharedTestUser1.UserCtx
			}

			resp, err := suite.Client.API.DeleteFinding(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteFinding.DeletedID))

			_, err = suite.Client.API.GetFindingByID(th.SharedTestUser1.UserCtx, tc.idToDelete)
			assert.ErrorContains(t, err, th.NotFoundErrorMsg)
		})
	}

	(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: anotherOrgFinding.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
}

func TestMutationDeleteBulkFinding(t *testing.T) {
	finding1 := createFinding(t, th.SharedTestUser1.UserCtx, "Bulk Delete Finding 1")
	finding2 := createFinding(t, th.SharedTestUser1.UserCtx, "Bulk Delete Finding 2")
	finding3 := createFinding(t, th.SharedTestUser1.UserCtx, "Bulk Delete Finding 3")

	orgFinding := createFinding(t, th.SharedTestUser2.UserCtx, "Unauthorized Bulk Delete Finding")
	auditorFinding := createFinding(t, th.SharedTestUser1.UserCtx, "Auditor Bulk Delete Finding")

	testCases := []struct {
		name                 string
		idsToDelete          []string
		ctx                  context.Context
		expectedErr          string
		expectedDeletedCount int
	}{
		{
			name:                 "auditor can delete existing finding",
			idsToDelete:          []string{auditorFinding.ID},
			ctx:                  th.SharedAuditorUser.UserCtx,
			expectedDeletedCount: 1,
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
				tc.ctx = th.SharedTestUser1.UserCtx
			}

			resp, err := suite.Client.API.DeleteBulkFinding(tc.ctx, tc.idsToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Len(resp.DeleteBulkFinding.DeletedIDs, tc.expectedDeletedCount))

			for _, id := range resp.DeleteBulkFinding.DeletedIDs {
				assert.Check(t, is.Contains(tc.idsToDelete, id))

				_, err := suite.Client.API.GetFindingByID(th.SharedTestUser1.UserCtx, id)
				assert.ErrorContains(t, err, th.NotFoundErrorMsg)
			}
		})
	}

	(&th.Cleanup[*generated.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: orgFinding.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
}

func createFinding(t *testing.T, ctx context.Context, displayName string) *testclient.CreateFinding_CreateFinding_Finding {
	t.Helper()

	resp, err := suite.Client.API.CreateFinding(ctx, testclient.CreateFindingInput{
		DisplayName:     &displayName,
		ExternalID:      lo.ToPtr("finding-" + ulids.New().String()),
		ExternalOwnerID: lo.ToPtr("external-owner"),
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	return &resp.CreateFinding.Finding
}
