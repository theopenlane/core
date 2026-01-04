package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

// TestFrameworkUpgradeDiff tests the diff calculation between two standard revisions
func TestFrameworkUpgradeDiff(t *testing.T) {
	ctx := setContext(systemAdminUser.UserCtx, suite.client.db)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// Create a unique shortName for this test
	shortName := "test-fw-" + ulids.New().String()

	// Create standard v1.0.0
	standardV1 := suite.client.db.Standard.Create().
		SetName("Test Framework V1").
		SetShortName(shortName).
		SetRevision("v1.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	// Create controls for v1.0.0
	ctlV1_1 := suite.client.db.Control.Create().
		SetStandardID(standardV1.ID).
		SetRefCode("CTL-001").
		SetTitle("Control 1 Original").
		SetDescription("Original description").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	ctlV1_2 := suite.client.db.Control.Create().
		SetStandardID(standardV1.ID).
		SetRefCode("CTL-002").
		SetTitle("Control 2").
		SetDescription("Will be removed").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	ctlV1_3 := suite.client.db.Control.Create().
		SetStandardID(standardV1.ID).
		SetRefCode("CTL-003").
		SetTitle("Control 3 Unchanged").
		SetDescription("Stays same").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// Clone controls to test user's org
	_, err := suite.client.api.CreateControlsByClone(testUser1.UserCtx, testclient.CloneControlInput{
		OwnerID:    &testUser1.OrganizationID,
		ControlIDs: []string{ctlV1_1.ID, ctlV1_2.ID, ctlV1_3.ID},
	})
	assert.NilError(t, err)

	// Create standard v2.0.0
	standardV2 := suite.client.db.Standard.Create().
		SetName("Test Framework V2").
		SetShortName(shortName).
		SetRevision("v2.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	// CTL-001 - updated title
	suite.client.db.Control.Create().
		SetStandardID(standardV2.ID).
		SetRefCode("CTL-001").
		SetTitle("Control 1 Updated").
		SetDescription("Original description").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// CTL-002 removed (not in v2)

	// CTL-003 - unchanged
	suite.client.db.Control.Create().
		SetStandardID(standardV2.ID).
		SetRefCode("CTL-003").
		SetTitle("Control 3 Unchanged").
		SetDescription("Stays same").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// CTL-004 - new control
	suite.client.db.Control.Create().
		SetStandardID(standardV2.ID).
		SetRefCode("CTL-004").
		SetTitle("Control 4 New").
		SetDescription("New in v2").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// Create standard v3.0.0
	standardV3 := suite.client.db.Standard.Create().
		SetName("Test Framework V3").
		SetShortName(shortName).
		SetRevision("v3.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	// CTL-001 - updated title
	suite.client.db.Control.Create().
		SetStandardID(standardV3.ID).
		SetRefCode("CTL-001").
		SetTitle("Control 1 Updated").
		SetDescription("Original description").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// CTL-003 - unchanged
	suite.client.db.Control.Create().
		SetStandardID(standardV3.ID).
		SetRefCode("CTL-003").
		SetTitle("Control 3 Unchanged").
		SetDescription("Stays same").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// CTL-004 - new control carried forward
	suite.client.db.Control.Create().
		SetStandardID(standardV3.ID).
		SetRefCode("CTL-004").
		SetTitle("Control 4 New").
		SetDescription("New in v2").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// CTL-005 - new control in v3
	suite.client.db.Control.Create().
		SetStandardID(standardV3.ID).
		SetRefCode("CTL-005").
		SetTitle("Control 5 New").
		SetDescription("New in v3").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	testCases := []struct {
		name              string
		standardID        string
		targetRevision    *string
		client            *testclient.TestClient
		ctx               context.Context
		expectedAdded     int
		expectedUpdated   int
		expectedRemoved   int
		expectedUnchanged int
		errorMsg          string
	}{
		{
			name:              "happy path - calculate diff v1 to latest",
			standardID:        standardV1.ID,
			targetRevision:    nil,
			client:            suite.client.api,
			ctx:               testUser1.UserCtx,
			expectedAdded:     2, // CTL-004, CTL-005
			expectedUpdated:   1, // CTL-001
			expectedRemoved:   1, // CTL-002
			expectedUnchanged: 1, // CTL-003
		},
		{
			name:              "using PAT",
			standardID:        standardV1.ID,
			targetRevision:    nil,
			client:            suite.client.apiWithPAT,
			ctx:               context.Background(),
			expectedAdded:     2,
			expectedUpdated:   1,
			expectedRemoved:   1,
			expectedUnchanged: 1,
		},
		{
			name:           "standard not found",
			standardID:     ulids.New().String(),
			targetRevision: lo.ToPtr("v2.0.0"),
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			errorMsg:       notFoundErrorMsg,
		},
		{
			name:           "target revision not found",
			standardID:     standardV1.ID,
			targetRevision: lo.ToPtr("v99.0.0"),
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
			errorMsg:       "not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.GetFrameworkUpgradeDiffs(tc.ctx, testclient.FrameworkUpgradeDiffInput{
				StandardID:     tc.standardID,
				TargetRevision: tc.targetRevision,
			})

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// Verify counts
			assert.Check(t, is.Len(resp.FrameworkUpgradeDiff.AddedControls, tc.expectedAdded))
			assert.Check(t, is.Len(resp.FrameworkUpgradeDiff.UpdatedControls, tc.expectedUpdated))
			assert.Check(t, is.Len(resp.FrameworkUpgradeDiff.RemovedControls, tc.expectedRemoved))
			assert.Check(t, is.Len(resp.FrameworkUpgradeDiff.UnchangedControls, tc.expectedUnchanged))

			// Verify summary
			assert.Check(t, is.Equal(int64(tc.expectedAdded), resp.FrameworkUpgradeDiff.Summary.AddedCount))
			assert.Check(t, is.Equal(int64(tc.expectedUpdated), resp.FrameworkUpgradeDiff.Summary.UpdatedCount))
			assert.Check(t, is.Equal(int64(tc.expectedRemoved), resp.FrameworkUpgradeDiff.Summary.RemovedCount))
			assert.Check(t, is.Equal(int64(tc.expectedUnchanged), resp.FrameworkUpgradeDiff.Summary.UnchangedCount))

			// Verify changes are detected
			if tc.expectedUpdated > 0 {
				assert.Check(t, len(resp.FrameworkUpgradeDiff.UpdatedControls[0].Changes) > 0)
			}

			if tc.targetRevision == nil {
				assert.Assert(t, resp.FrameworkUpgradeDiff.TargetStandard.Revision != nil)
				assert.Check(t, is.Equal("v3.0.0", *resp.FrameworkUpgradeDiff.TargetStandard.Revision))
			}
		})
	}

	// Cleanup
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: []string{standardV1.ID, standardV2.ID, standardV3.ID}}).MustDelete(systemAdminUser.UserCtx, t)
}

// TestApplyFrameworkUpgrade tests applying an upgrade and verifying the changes
func TestApplyFrameworkUpgrade(t *testing.T) {
	ctx := setContext(systemAdminUser.UserCtx, suite.client.db)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	shortName := "apply-fw-" + ulids.New().String()

	// Create v1.0.0
	standardV1 := suite.client.db.Standard.Create().
		SetName("Apply Framework V1").
		SetShortName(shortName).
		SetRevision("v1.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	ctlV1_1 := suite.client.db.Control.Create().
		SetStandardID(standardV1.ID).
		SetRefCode("ACTL-001").
		SetTitle("Apply Ctl 1 Original").
		SetDescription("Will update").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	ctlV1_2 := suite.client.db.Control.Create().
		SetStandardID(standardV1.ID).
		SetRefCode("ACTL-002").
		SetTitle("Apply Ctl 2").
		SetDescription("Will remove").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// Clone to org
	cloneResp, err := suite.client.api.CreateControlsByClone(testUser1.UserCtx, testclient.CloneControlInput{
		OwnerID:    &testUser1.OrganizationID,
		ControlIDs: []string{ctlV1_1.ID, ctlV1_2.ID},
	})
	assert.NilError(t, err)

	orgCtl1ID := cloneResp.CreateControlsByClone.Controls[0].ID
	orgCtl2ID := cloneResp.CreateControlsByClone.Controls[1].ID

	// Create v2.0.0
	standardV2 := suite.client.db.Standard.Create().
		SetName("Apply Framework V2").
		SetShortName(shortName).
		SetRevision("v2.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	// ACTL-001 - updated
	suite.client.db.Control.Create().
		SetStandardID(standardV2.ID).
		SetRefCode("ACTL-001").
		SetTitle("Apply Ctl 1 Updated").
		SetDescription("Updated in v2").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// ACTL-002 removed

	// ACTL-003 - new
	suite.client.db.Control.Create().
		SetStandardID(standardV2.ID).
		SetRefCode("ACTL-003").
		SetTitle("Apply Ctl 3 New").
		SetDescription("New in v2").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// Create v3.0.0
	standardV3 := suite.client.db.Standard.Create().
		SetName("Apply Framework V3").
		SetShortName(shortName).
		SetRevision("v3.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	// ACTL-001 - updated in v3
	suite.client.db.Control.Create().
		SetStandardID(standardV3.ID).
		SetRefCode("ACTL-001").
		SetTitle("Apply Ctl 1 Updated v3").
		SetDescription("Updated in v3").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// ACTL-003 - carried forward
	suite.client.db.Control.Create().
		SetStandardID(standardV3.ID).
		SetRefCode("ACTL-003").
		SetTitle("Apply Ctl 3 New").
		SetDescription("New in v2").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	// ACTL-004 - new in v3
	suite.client.db.Control.Create().
		SetStandardID(standardV3.ID).
		SetRefCode("ACTL-004").
		SetTitle("Apply Ctl 4 New").
		SetDescription("New in v3").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	testCases := []struct {
		name           string
		standardID     string
		targetRevision *string
		client         *testclient.TestClient
		ctx            context.Context
	}{
		{
			name:           "happy path - apply latest upgrade",
			standardID:     standardV1.ID,
			targetRevision: nil,
			client:         suite.client.api,
			ctx:            testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.ApplyFrameworkUpgrade(tc.ctx, testclient.ApplyFrameworkUpgradeInput{
				StandardID:     tc.standardID,
				TargetRevision: tc.targetRevision,
			})

			assert.NilError(t, err)
			assert.Check(t, resp.ApplyFrameworkUpgrade.Success)

			// Verify summary
			assert.Check(t, is.Equal(int64(2), resp.ApplyFrameworkUpgrade.Summary.AddedCount))
			assert.Check(t, is.Equal(int64(1), resp.ApplyFrameworkUpgrade.Summary.UpdatedCount))
			assert.Check(t, is.Equal(int64(1), resp.ApplyFrameworkUpgrade.Summary.RemovedCount))

			// Verify actual changes
			updatedCtl, err := tc.client.GetControlByID(tc.ctx, orgCtl1ID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal("Updated in v3", *updatedCtl.Control.Description))

			// Verify deletion
			_, err = tc.client.GetControlByID(tc.ctx, orgCtl2ID)
			assert.ErrorContains(t, err, notFoundErrorMsg)

			// Verify new controls added
			controls, err := tc.client.GetControls(tc.ctx, nil, nil, nil, nil, &testclient.ControlWhereInput{
				OwnerID:   lo.ToPtr(testUser1.OrganizationID),
				RefCodeIn: []string{"ACTL-003", "ACTL-004"},
			}, nil)
			assert.NilError(t, err)
			assert.Check(t, is.Len(controls.Controls.Edges, 2))

			upgradedCtl := suite.client.db.Control.GetX(allowCtx, orgCtl1ID)
			assert.Assert(t, upgradedCtl.ReferenceFrameworkRevision != nil)
			assert.Check(t, is.Equal("v3.0.0", *upgradedCtl.ReferenceFrameworkRevision))

			newControls, err := suite.client.db.Control.Query().
				Where(
					control.OwnerID(testUser1.OrganizationID),
					control.RefCodeIn("ACTL-003", "ACTL-004"),
					control.DeletedAtIsNil(),
				).
				All(allowCtx)
			assert.NilError(t, err)
			assert.Check(t, is.Len(newControls, 2))
			for _, newControl := range newControls {
				assert.Assert(t, newControl.ReferenceFrameworkRevision != nil)
				assert.Check(t, is.Equal("v3.0.0", *newControl.ReferenceFrameworkRevision))
			}
		})
	}

	// Cleanup
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: []string{standardV1.ID, standardV2.ID, standardV3.ID}}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestApplyFrameworkUpgradeSubcontrolsAndMappings(t *testing.T) {
	ctx := setContext(systemAdminUser.UserCtx, suite.client.db)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	shortName := "apply-fw-submap-" + ulids.New().String()

	standardV1 := suite.client.db.Standard.Create().
		SetName("Apply Framework Sub V1").
		SetShortName(shortName).
		SetRevision("v1.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	ctlV1_1 := suite.client.db.Control.Create().
		SetStandardID(standardV1.ID).
		SetRefCode("SM-CTL-001").
		SetTitle("Submap Control 1").
		SetDescription("Control v1").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	ctlV1_2 := suite.client.db.Control.Create().
		SetStandardID(standardV1.ID).
		SetRefCode("SM-CTL-002").
		SetTitle("Submap Control 2").
		SetDescription("Control v1").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	suite.client.db.Subcontrol.Create().
		SetControlID(ctlV1_1.ID).
		SetRefCode("SM-SCL-001").
		SetTitle("Subcontrol 1 Original").
		SetDescription("Old subcontrol").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	suite.client.db.Subcontrol.Create().
		SetControlID(ctlV1_1.ID).
		SetRefCode("SM-SCL-002").
		SetTitle("Subcontrol 2 Remove").
		SetDescription("Removed in v2").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	_, err := suite.client.api.CreateControlsByClone(testUser1.UserCtx, testclient.CloneControlInput{
		OwnerID:    &testUser1.OrganizationID,
		ControlIDs: []string{ctlV1_1.ID, ctlV1_2.ID},
	})
	assert.NilError(t, err)

	orgCtl1 := suite.client.db.Control.Query().
		Where(
			control.OwnerID(testUser1.OrganizationID),
			control.RefCode("SM-CTL-001"),
			control.DeletedAtIsNil(),
		).
		OnlyX(allowCtx)

	orgCtl2 := suite.client.db.Control.Query().
		Where(
			control.OwnerID(testUser1.OrganizationID),
			control.RefCode("SM-CTL-002"),
			control.DeletedAtIsNil(),
		).
		OnlyX(allowCtx)

	(&MappedControlBuilder{
		client:         suite.client,
		FromControlIDs: []string{orgCtl1.ID},
		ToControlIDs:   []string{orgCtl2.ID},
	}).MustNew(testUser1.UserCtx, t)

	standardV2 := suite.client.db.Standard.Create().
		SetName("Apply Framework Sub V2").
		SetShortName(shortName).
		SetRevision("v2.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	ctlV2_1 := suite.client.db.Control.Create().
		SetStandardID(standardV2.ID).
		SetRefCode("SM-CTL-001").
		SetTitle("Submap Control 1 Updated").
		SetDescription("Control v2").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	ctlV2_3 := suite.client.db.Control.Create().
		SetStandardID(standardV2.ID).
		SetRefCode("SM-CTL-003").
		SetTitle("Submap Control 3 New").
		SetDescription("New in v2").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	suite.client.db.Subcontrol.Create().
		SetControlID(ctlV2_1.ID).
		SetRefCode("SM-SCL-001").
		SetTitle("Subcontrol 1 Updated").
		SetDescription("Updated subcontrol").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	suite.client.db.Subcontrol.Create().
		SetControlID(ctlV2_1.ID).
		SetRefCode("SM-SCL-003").
		SetTitle("Subcontrol 3 New").
		SetDescription("New subcontrol").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	suggestedFrameworkMapping := suite.client.db.MappedControl.Create().
		AddFromControlIDs(ctlV2_1.ID).
		AddToControlIDs(ctlV2_3.ID).
		SetSource(enums.MappingSourceSuggested).
		SaveX(allowCtx)

	resp, err := suite.client.api.ApplyFrameworkUpgrade(testUser1.UserCtx, testclient.ApplyFrameworkUpgradeInput{
		StandardID:     standardV1.ID,
		TargetRevision: nil,
	})
	assert.NilError(t, err)
	assert.Check(t, resp.ApplyFrameworkUpgrade.Success)

	updatedSubcontrol := suite.client.db.Subcontrol.Query().
		Where(
			subcontrol.OwnerID(testUser1.OrganizationID),
			subcontrol.ControlID(orgCtl1.ID),
			subcontrol.RefCode("SM-SCL-001"),
			subcontrol.DeletedAtIsNil(),
		).
		OnlyX(allowCtx)
	assert.Check(t, is.Equal("Subcontrol 1 Updated", updatedSubcontrol.Title))

	addedExists, err := suite.client.db.Subcontrol.Query().
		Where(
			subcontrol.OwnerID(testUser1.OrganizationID),
			subcontrol.ControlID(orgCtl1.ID),
			subcontrol.RefCode("SM-SCL-003"),
			subcontrol.DeletedAtIsNil(),
		).
		Exist(allowCtx)
	assert.NilError(t, err)
	assert.Assert(t, addedExists)

	removedExists, err := suite.client.db.Subcontrol.Query().
		Where(
			subcontrol.OwnerID(testUser1.OrganizationID),
			subcontrol.ControlID(orgCtl1.ID),
			subcontrol.RefCode("SM-SCL-002"),
			subcontrol.DeletedAtIsNil(),
		).
		Exist(allowCtx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(false, removedExists))

	orgCtl3 := suite.client.db.Control.Query().
		Where(
			control.OwnerID(testUser1.OrganizationID),
			control.RefCode("SM-CTL-003"),
			control.DeletedAtIsNil(),
		).
		OnlyX(allowCtx)

	orgMappings, err := suite.client.db.MappedControl.Query().
		Where(mappedcontrol.OwnerID(testUser1.OrganizationID)).
		WithFromControls().
		WithToControls().
		All(allowCtx)
	assert.NilError(t, err)

	foundManual := false
	foundSuggested := false
	manualMappingID := ""
	suggestedMappingID := ""

	for _, mapping := range orgMappings {
		fromIDs := lo.Map(mapping.Edges.FromControls, func(c *generated.Control, _ int) string { return c.ID })
		toIDs := lo.Map(mapping.Edges.ToControls, func(c *generated.Control, _ int) string { return c.ID })

		if mapping.Source == enums.MappingSourceManual &&
			lo.Contains(fromIDs, orgCtl1.ID) &&
			lo.Contains(toIDs, orgCtl2.ID) {
			foundManual = true
			manualMappingID = mapping.ID
		}

		if mapping.Source == enums.MappingSourceSuggested &&
			lo.Contains(fromIDs, orgCtl1.ID) &&
			lo.Contains(toIDs, orgCtl3.ID) {
			foundSuggested = true
			suggestedMappingID = mapping.ID
		}
	}

	assert.Check(t, !foundManual)
	assert.Check(t, foundSuggested)

	mappingIDs := lo.Filter([]string{manualMappingID, suggestedMappingID}, func(id string, _ int) bool { return id != "" })
	if len(mappingIDs) > 0 {
		(&Cleanup[*generated.MappedControlDeleteOne]{client: suite.client.db.MappedControl, IDs: mappingIDs}).MustDelete(testUser1.UserCtx, t)
	}
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{orgCtl1.ID, orgCtl3.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.MappedControlDeleteOne]{client: suite.client.db.MappedControl, ID: suggestedFrameworkMapping.ID}).MustDelete(systemAdminUser.UserCtx, t)

	// Cleanup
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: []string{standardV1.ID, standardV2.ID}}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestApplyFrameworkUpgradeBackfillsReferenceRevision(t *testing.T) {
	ctx := setContext(systemAdminUser.UserCtx, suite.client.db)
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	shortName := "apply-fw-ref-" + ulids.New().String()

	standardV1 := suite.client.db.Standard.Create().
		SetName("Apply Framework Ref V1").
		SetShortName(shortName).
		SetRevision("v1.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	ctlV1 := suite.client.db.Control.Create().
		SetStandardID(standardV1.ID).
		SetRefCode("REF-CTL-001").
		SetTitle("Ref Control 1").
		SetDescription("Legacy control without reference revision").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	cloneResp, err := suite.client.api.CreateControlsByClone(testUser1.UserCtx, testclient.CloneControlInput{
		OwnerID:    &testUser1.OrganizationID,
		ControlIDs: []string{ctlV1.ID},
	})
	assert.NilError(t, err)

	orgCtlID := cloneResp.CreateControlsByClone.Controls[0].ID

	err = suite.client.db.Control.UpdateOneID(orgCtlID).
		ClearReferenceFrameworkRevision().
		Exec(allowCtx)
	assert.NilError(t, err)

	standardV2 := suite.client.db.Standard.Create().
		SetName("Apply Framework Ref V2").
		SetShortName(shortName).
		SetRevision("v2.0.0").
		SetIsPublic(true).
		SetFramework("Test").
		SaveX(allowCtx)

	suite.client.db.Control.Create().
		SetStandardID(standardV2.ID).
		SetRefCode("REF-CTL-001").
		SetTitle("Ref Control 1").
		SetDescription("Legacy control without reference revision").
		SetSource(enums.ControlSourceFramework).
		SaveX(allowCtx)

	_, err = suite.client.api.ApplyFrameworkUpgrade(testUser1.UserCtx, testclient.ApplyFrameworkUpgradeInput{
		StandardID:     standardV1.ID,
		TargetRevision: nil,
	})
	assert.NilError(t, err)

	upgraded := suite.client.db.Control.GetX(allowCtx, orgCtlID)
	assert.Assert(t, upgraded.ReferenceFrameworkRevision != nil)
	assert.Check(t, is.Equal("v2.0.0", *upgraded.ReferenceFrameworkRevision))

	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: []string{standardV1.ID, standardV2.ID}}).MustDelete(systemAdminUser.UserCtx, t)
}
