package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/internal/testutils"
	"github.com/theopenlane/utils/ulids"
)

// controlReportTestData holds entity IDs seeded by seedControlReportTestData
type controlReportTestData struct {
	primaryControlID   string
	secondaryControlID string
	tertiaryControlID  string
	subcontrolID       string
	evidenceID         string
	policyID           string
	controlOwnerID     string // group assigned as owner of the primary control
	// forwardMappingID maps primary → secondary
	forwardMappingID string
	// reverseMappingID maps secondary → primary, creating a duplicate reference to secondary
	// that the deduplication logic should collapse to a single relatedControl entry
	reverseMappingID string
	// tertiaryMappingID maps primary → tertiary, adding one more unique related control
	tertiaryMappingID string

	// system mapping fields: system-owned controls are mapped together; org controls with
	// matching refCode+framework confirm that system mappings surface in relatedControls
	sysMapOrgSourceID string // org control matching sysControlA (the one to query)
	sysMapOrgTargetID string // org control matching sysControlB (expected in relatedControls)
	sysMapID          string // the system-owned MappedControl

	// ctrlToSubcontrolRelatedID is a subcontrol mapped directly to primaryControlID;
	// it must appear in primary's relatedControls with IsSubcontrol: true
	ctrlToSubcontrolRelatedID string
}

// seedControlReportTestData enriches primaryControlID with a subcontrol, linked evidence,
// an internal policy, and three org-owned mapped controls to exercise deduplication.
// It also creates a system-owned mapping between two new system controls and matching
// org controls to verify that system mappings surface in relatedControls.
// All three input controls must already exist in the context's org.
//
// Org mappings on primary:
//   - forward:  primary → secondary
//   - reverse:  secondary → primary  (secondary appears twice, deduped to one relatedControl)
//   - tertiary: primary → tertiary   (unique second related control)
//
// System mapping:
//   - sysControlA and sysControlB are system-owned with unique refCode+"SOC2" framework
//   - sysMapOrgSource/sysMapOrgTarget are org-owned with the same refCode+framework
//   - querying sysMapOrgSource's relatedControls should return sysMapOrgTarget via the system mapping
//
// Expected relatedControls for primary: [secondary, tertiary] (2 unique entries).
// Expected relatedControls for sysMapOrgSource: [sysMapOrgTarget] (1 entry via system mapping).
func seedControlReportTestData(ctx context.Context, t *testing.T, primaryControlID, secondaryControlID, tertiaryControlID, controlOwnerGroupID string) *controlReportTestData {
	t.Helper()

	sc := (&SubcontrolBuilder{client: suite.client, ControlID: primaryControlID}).MustNew(ctx, t)

	ev := (&EvidenceBuilder{client: suite.client, ControlID: primaryControlID}).MustNew(ctx, t)
	(&EvidenceBuilder{
		client:    suite.client,
		ControlID: primaryControlID,
		Status:    lo.ToPtr(enums.EvidenceStatusAuditorApproved),
	}).MustNew(ctx, t)

	policy := (&InternalPolicyBuilder{client: suite.client}).MustNew(ctx, t)
	dbCtx := setContext(ctx, suite.client.db)
	requireNoError(t, suite.client.db.InternalPolicy.UpdateOneID(policy.ID).AddControlIDs(primaryControlID).Exec(dbCtx))

	// forward: primary → secondary
	forward := (&MappedControlBuilder{
		client:         suite.client,
		FromControlIDs: []string{primaryControlID},
		ToControlIDs:   []string{secondaryControlID},
	}).MustNew(ctx, t)

	// reverse: secondary → primary (secondary now appears in two MappedControl records;
	// processMappedControlResults deduplicates via the refCode::framework map key)
	reverse := (&MappedControlBuilder{
		client:         suite.client,
		FromControlIDs: []string{secondaryControlID},
		ToControlIDs:   []string{primaryControlID},
	}).MustNew(ctx, t)

	// tertiary: primary → tertiary (distinct ref code, adds one more unique related control)
	tertiary := (&MappedControlBuilder{
		client:         suite.client,
		FromControlIDs: []string{primaryControlID},
		ToControlIDs:   []string{tertiaryControlID},
	}).MustNew(ctx, t)

	// control → subcontrol mapping: a subcontrol of tertiary is mapped directly to primary
	scOfTertiary := (&SubcontrolBuilder{client: suite.client, ControlID: tertiaryControlID}).MustNew(ctx, t)
	(&MappedControlBuilder{
		client:          suite.client,
		FromControlIDs:  []string{primaryControlID},
		ToSubcontrolIDs: []string{scOfTertiary.ID},
	}).MustNew(ctx, t)

	// system mapping scenario: unique refCodes per call prevent cross-test interference
	sysRefA := ulids.New().String()
	sysRefB := ulids.New().String()
	sysFramework := lo.ToPtr("SOC 2")

	sysControlA := (&ControlBuilder{
		client:             suite.client,
		RefCode:            sysRefA,
		ReferenceFramework: sysFramework,
	}).MustNew(sharedSystemAdminUser.UserCtx, t)

	sysControlB := (&ControlBuilder{
		client:             suite.client,
		RefCode:            sysRefB,
		ReferenceFramework: sysFramework,
	}).MustNew(sharedSystemAdminUser.UserCtx, t)

	// system-owned mapped control linking sysControlA → sysControlB
	sysMap := (&MappedControlBuilder{
		client:         suite.client,
		FromControlIDs: []string{sysControlA.ID},
		ToControlIDs:   []string{sysControlB.ID},
	}).MustNew(sharedSystemAdminUser.UserCtx, t)

	// org-owned controls that mirror the system controls by refCode+framework;
	// getOrgMappedControlsInfo resolves sysControlB's refCode to orgTgt
	orgSrc := (&ControlBuilder{
		client:             suite.client,
		RefCode:            sysRefA,
		ReferenceFramework: sysFramework,
	}).MustNew(ctx, t)

	orgTgt := (&ControlBuilder{
		client:             suite.client,
		RefCode:            sysRefB,
		ReferenceFramework: sysFramework,
	}).MustNew(ctx, t)

	return &controlReportTestData{
		primaryControlID:          primaryControlID,
		secondaryControlID:        secondaryControlID,
		tertiaryControlID:         tertiaryControlID,
		subcontrolID:              sc.ID,
		evidenceID:                ev.ID,
		policyID:                  policy.ID,
		controlOwnerID:            controlOwnerGroupID,
		forwardMappingID:          forward.ID,
		reverseMappingID:          reverse.ID,
		tertiaryMappingID:         tertiary.ID,
		sysMapOrgSourceID:         orgSrc.ID,
		sysMapOrgTargetID:         orgTgt.ID,
		sysMapID:                  sysMap.ID,
		ctrlToSubcontrolRelatedID: scOfTertiary.ID,
	}
}

func TestQueryControlReports(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.seedFreshOrgUsers(t)
	orgUser := suite.seedOrgOwner(t)

	// create 8 filler controls first (oldest) so that the enriched controls and the
	// system-mapping controls created inside the seed fall in the first page (CreatedAt DESC)
	for range 8 {
		(&ControlBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)
	}

	// system-owned controls must not appear in controlReports results (resolver filters SystemOwned: false);
	// the hook sets system_owned = true automatically when it sees a system admin caller
	for range 3 {
		(&ControlBuilder{client: suite.client}).MustNew(sharedSystemAdminUser.UserCtx, t)
	}

	// create primary/secondary/tertiary after the fillers so they appear in the first page
	ownerGroup := (&GroupBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)
	primary := (&ControlBuilder{client: suite.client, ControlOwnerID: ownerGroup.ID}).MustNew(localTestOrg.owner.UserCtx, t)
	secondary := (&ControlBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)
	tertiary := (&ControlBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)

	// seed adds 2 more org controls (sysMapOrgSource, sysMapOrgTarget) → 8+3+2 = 13 total
	orgOwnedCount := int64(13)
	richData := seedControlReportTestData(localTestOrg.owner.UserCtx, t, primary.ID, secondary.ID, tertiary.ID, ownerGroup.ID)

	testCases := []struct {
		name            string
		first           *int64
		last            *int64
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			ctx:             localTestOrg.owner.UserCtx,
			expectedResults: testutils.MaxResultLimit,
		},
		{
			name:            "happy path, with first set",
			first:           lo.ToPtr(int64(5)),
			ctx:             localTestOrg.owner.UserCtx,
			expectedResults: 5,
		},
		{
			name:            "happy path, with last set by admin",
			last:            lo.ToPtr(int64(3)),
			ctx:             localTestOrg.admin.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "first set over max (10 in test) by member",
			first:           &orgOwnedCount,
			ctx:             localTestOrg.member.UserCtx,
			expectedResults: testutils.MaxResultLimit,
		},
		{
			name:            "last set over max (10 in test) by auditor",
			last:            &orgOwnedCount,
			ctx:             localTestOrg.auditor.UserCtx,
			expectedResults: testutils.MaxResultLimit,
		},
		{
			name:            "another org, no results returned",
			ctx:             orgUser.owner.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			if tc.first != nil || tc.last != nil {
				resp, err := suite.client.api.GetControlReports(tc.ctx, tc.first, tc.last, nil, nil, nil, nil)
				assert.NilError(t, err)
				assert.Check(t, resp != nil)

				assert.Check(t, is.Len(resp.ControlReports.Edges, tc.expectedResults))
				assert.Check(t, is.Equal(orgOwnedCount, resp.ControlReports.TotalCount))

				if tc.last != nil {
					assert.Check(t, resp.ControlReports.PageInfo.HasPreviousPage)
				} else {
					assert.Check(t, resp.ControlReports.PageInfo.HasNextPage)
				}

				return
			}

			resp, err := suite.client.api.GetAllControlReports(tc.ctx)
			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			assert.Check(t, is.Len(resp.ControlReports.Edges, tc.expectedResults))

			if tc.expectedResults > 0 {
				assert.Check(t, is.Equal(orgOwnedCount, resp.ControlReports.TotalCount))
				assert.Check(t, resp.ControlReports.PageInfo.HasNextPage)

				for _, edge := range resp.ControlReports.Edges {
					assert.Check(t, edge.Node != nil)
					assert.Check(t, len(edge.Node.ID) != 0)
					assert.Check(t, len(edge.Node.RefCode) != 0)
					assert.Check(t, edge.Node.Status != nil)
					assert.Check(t, edge.Node.EvidenceStatus != nil)
					assert.Check(t, edge.Node.LinkedPolicies != nil)
					assert.Check(t, edge.Node.RelatedControls != nil)

					// verify enrichment fields for the control that has seeded data
					if edge.Node.ID == richData.primaryControlID {
						assert.Check(t, is.Len(edge.Node.Subcontrols, 1))
						assert.Check(t, is.Equal(edge.Node.Subcontrols[0].ID, richData.subcontrolID))
						assert.Check(t, is.Equal(int64(2), edge.Node.EvidenceStatus.TotalCount))
						assert.Check(t, is.Equal(int64(1), edge.Node.EvidenceStatus.ApprovedCount))
						assert.Check(t, is.Len(edge.Node.EvidenceStatus.CountByStatus, 2))
						assert.Check(t, is.Equal(int64(1), edge.Node.LinkedPolicies.TotalCount))
						// secondary appears in both the forward and reverse MappedControl records;
						// deduplication collapses it to one entry; tertiary and the
						// directly-mapped subcontrol each add one more = 3 total
						assert.Check(t, is.Len(edge.Node.RelatedControls, 3))
						assert.Check(t, edge.Node.ControlOwner != nil)
						assert.Check(t, is.Equal(richData.controlOwnerID, edge.Node.ControlOwner.ID))

						// the subcontrol mapped directly to primary must appear with IsSubcontrol: true
						var foundSubcontrolRelated bool
						for _, rc := range edge.Node.RelatedControls {
							if rc.ID == richData.ctrlToSubcontrolRelatedID {
								foundSubcontrolRelated = true
								assert.Check(t, rc.IsSubcontrol)
							}
						}
						assert.Check(t, foundSubcontrolRelated)
					}

					// org control matching sysControlA should surface sysMapOrgTarget via system mapping
					if edge.Node.ID == richData.sysMapOrgSourceID {
						assert.Check(t, is.Len(edge.Node.RelatedControls, 1))
					}
				}
			} else {
				assert.Check(t, is.Equal(int64(0), resp.ControlReports.TotalCount))
				assert.Check(t, !resp.ControlReports.PageInfo.HasNextPage)
			}
		})
	}

	cleanupOrganizationDataWithContext(localTestOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(orgUser.owner.UserCtx, t)
}

func TestQueryControlReportsByCategory(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.seedOrgOwner(t)
	orgUser := suite.seedOrgOwner(t)

	cat1 := "Access Control"
	cat2 := "Availability"

	ownerGroup := (&GroupBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)
	control1 := (&ControlBuilder{client: suite.client, Category: cat1, ControlOwnerID: ownerGroup.ID}).MustNew(localTestOrg.owner.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, Category: cat1}).MustNew(localTestOrg.owner.UserCtx, t)
	control3 := (&ControlBuilder{client: suite.client, Category: cat2}).MustNew(localTestOrg.owner.UserCtx, t)
	(&ControlBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)

	// system-owned controls must not appear in results regardless of category;
	// the hook sets system_owned = true automatically when it sees a system admin caller
	(&ControlBuilder{client: suite.client, Category: cat1}).MustNew(sharedSystemAdminUser.UserCtx, t)
	(&ControlBuilder{client: suite.client, Category: cat2}).MustNew(sharedSystemAdminUser.UserCtx, t)

	// enrich control1 with associated data so enrichment paths are exercised;
	// control3 is used as the tertiary to confirm a second unique related control
	richData := seedControlReportTestData(localTestOrg.owner.UserCtx, t, control1.ID, control2.ID, control3.ID, ownerGroup.ID)

	testCases := []struct {
		name               string
		ctx                context.Context
		where              *testclient.ControlWhereInput
		expectedCategories int
	}{
		{
			name:               "happy path, returns all categories including uncategorized",
			ctx:                localTestOrg.owner.UserCtx,
			expectedCategories: 3,
		},
		{
			name:               "happy path, filter by category",
			ctx:                localTestOrg.owner.UserCtx,
			where:              &testclient.ControlWhereInput{Category: lo.ToPtr(cat1)},
			expectedCategories: 1,
		},
		{
			name:               "another org, no categories returned",
			ctx:                orgUser.owner.UserCtx,
			expectedCategories: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := suite.client.api.GetControlReportsByCategory(tc.ctx, tc.where)
			assert.NilError(t, err)
			assert.Check(t, resp != nil)

			assert.Check(t, is.Len(resp.ControlReportsByCategory, tc.expectedCategories))

			if tc.expectedCategories == 0 {
				return
			}

			if tc.where == nil {
				totalControls := 0
				catCounts := map[string]int{}
				for _, cat := range resp.ControlReportsByCategory {
					assert.Check(t, is.Equal(int(cat.TotalCount), len(cat.Controls)))

					for _, c := range cat.Controls {
						assert.Check(t, len(c.ID) != 0)
						assert.Check(t, len(c.RefCode) != 0)
						assert.Check(t, c.Status != nil)
						assert.Check(t, c.EvidenceStatus != nil)
						assert.Check(t, c.LinkedPolicies != nil)
						assert.Check(t, c.RelatedControls != nil)

						// verify enrichment fields for the control that has seeded data
						if c.ID == richData.primaryControlID {
							assert.Check(t, is.Len(c.Subcontrols, 1))
							assert.Check(t, is.Equal(c.Subcontrols[0].ID, richData.subcontrolID))
							assert.Check(t, is.Equal(int64(2), c.EvidenceStatus.TotalCount))
							assert.Check(t, is.Equal(int64(1), c.EvidenceStatus.ApprovedCount))
							assert.Check(t, is.Len(c.EvidenceStatus.CountByStatus, 2))
							assert.Check(t, is.Equal(int64(1), c.LinkedPolicies.TotalCount))
							// secondary appears in both the forward and reverse MappedControl records;
							// deduplication collapses it to one entry; tertiary and the
							// directly-mapped subcontrol each add one more = 3 total
							assert.Check(t, is.Len(c.RelatedControls, 3))
							assert.Check(t, c.ControlOwner != nil)
							assert.Check(t, is.Equal(richData.controlOwnerID, c.ControlOwner.ID))

							// the subcontrol mapped directly to primary must appear with IsSubcontrol: true
							var foundSubcontrolRelated bool
							for _, rc := range c.RelatedControls {
								if rc.ID == richData.ctrlToSubcontrolRelatedID {
									foundSubcontrolRelated = true
									assert.Check(t, rc.IsSubcontrol)
								}
							}
							assert.Check(t, foundSubcontrolRelated)
						}

						// org control matching sysControlA should surface sysMapOrgTarget via system mapping
						if c.ID == richData.sysMapOrgSourceID {
							assert.Check(t, is.Len(c.RelatedControls, 1))
						}
					}

					catCounts[cat.Category] = len(cat.Controls)
					totalControls += len(cat.Controls)
				}

				// seed adds orgSrc and orgTgt (no category) on top of the 4 directly created controls:
				// 2 (cat1) + 1 (cat2) + 1 (no cat) + orgSrc + orgTgt = 6
				assert.Check(t, is.Equal(6, totalControls))
				assert.Check(t, is.Equal(2, catCounts[cat1]))
				assert.Check(t, is.Equal(1, catCounts[cat2]))
				assert.Check(t, is.Equal(3, catCounts[""]))
			}

			if tc.where != nil && tc.where.Category != nil {
				assert.Check(t, is.Equal(cat1, resp.ControlReportsByCategory[0].Category))
				assert.Check(t, is.Equal(int64(2), resp.ControlReportsByCategory[0].TotalCount))
			}
		})
	}

	cleanupOrganizationDataWithContext(localTestOrg.owner.UserCtx, t)
	cleanupOrganizationDataWithContext(orgUser.owner.UserCtx, t)
}

// TestControlReportRelatedControlsScoping verifies two properties of relatedControls (and the
// linkedPolicies derived from it):
//
//   - same-side siblings are not related: a mapping only asserts a relationship between its
//     from-side and its to-side, so two controls sharing the from-side are not related to each other
//   - mappings are not chased transitively: if sibA → target and target → hop2 are two separate
//     mappings, sibA's related controls are target only, never hop2
//
// Because linkedPolicies unions the policies of a control's related controls, the sibling's and
// the transitive control's policies must not leak into sibA's linkedPolicies.
func TestControlReportRelatedControlsScoping(t *testing.T) {
	t.Parallel()

	org := suite.seedOrgOwner(t)
	ctx := org.owner.UserCtx

	// sibA and sibB share the from-side of one mapping; target is its lone to-side control
	sibA := (&ControlBuilder{client: suite.client}).MustNew(ctx, t)
	sibB := (&ControlBuilder{client: suite.client}).MustNew(ctx, t)
	target := (&ControlBuilder{client: suite.client}).MustNew(ctx, t)
	// hop2 is a second hop reachable only by chasing target's mapping transitively
	hop2 := (&ControlBuilder{client: suite.client}).MustNew(ctx, t)

	// mapping 1: [sibA, sibB] -> [target]
	(&MappedControlBuilder{
		client:         suite.client,
		FromControlIDs: []string{sibA.ID, sibB.ID},
		ToControlIDs:   []string{target.ID},
	}).MustNew(ctx, t)

	// mapping 2: [target] -> [hop2]
	(&MappedControlBuilder{
		client:         suite.client,
		FromControlIDs: []string{target.ID},
		ToControlIDs:   []string{hop2.ID},
	}).MustNew(ctx, t)

	// the sibling and the transitive hop each get a distinct linked policy; target's own policy
	// lets us prove only the direct related control's policy surfaces on sibA
	siblingPolicy := (&InternalPolicyBuilder{client: suite.client}).MustNew(ctx, t)
	transitivePolicy := (&InternalPolicyBuilder{client: suite.client}).MustNew(ctx, t)
	targetPolicy := (&InternalPolicyBuilder{client: suite.client}).MustNew(ctx, t)

	dbCtx := setContext(ctx, suite.client.db)
	requireNoError(t, suite.client.db.InternalPolicy.UpdateOneID(siblingPolicy.ID).AddControlIDs(sibB.ID).Exec(dbCtx))
	requireNoError(t, suite.client.db.InternalPolicy.UpdateOneID(transitivePolicy.ID).AddControlIDs(hop2.ID).Exec(dbCtx))
	requireNoError(t, suite.client.db.InternalPolicy.UpdateOneID(targetPolicy.ID).AddControlIDs(target.ID).Exec(dbCtx))

	resp, err := suite.client.api.GetAllControlReports(ctx)
	assert.NilError(t, err)
	assert.Check(t, resp != nil)

	var checkedSibA bool
	for _, edge := range resp.ControlReports.Edges {
		if edge.Node.ID != sibA.ID {
			continue
		}

		checkedSibA = true

		// related controls is the opposite side of sibA's own mapping only: target, never the
		// same-side sibling sibB and never the transitive hop2
		assert.Check(t, is.Len(edge.Node.RelatedControls, 1))
		assert.Check(t, is.Equal(target.ID, edge.Node.RelatedControls[0].ID))

		for _, rc := range edge.Node.RelatedControls {
			assert.Check(t, rc.ID != sibB.ID, "same-side sibling must not be a related control")
			assert.Check(t, rc.ID != hop2.ID, "transitively mapped control must not be a related control")
		}

		// linkedPolicies is derived from relatedControls, so only the target's policy surfaces,
		// not the sibling's or the transitive hop's
		assert.Check(t, is.Equal(int64(1), edge.Node.LinkedPolicies.TotalCount))
		assert.Check(t, is.Len(edge.Node.LinkedPolicies.InternalPolicies, 1))
		assert.Check(t, is.Equal(targetPolicy.ID, edge.Node.LinkedPolicies.InternalPolicies[0].ID))
	}

	assert.Check(t, checkedSibA)

	cleanupOrganizationDataWithContext(ctx, t)
}
