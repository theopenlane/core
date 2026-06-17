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
)

// controlReportTestData holds entity IDs seeded by seedControlReportTestData
type controlReportTestData struct {
	primaryControlID   string
	secondaryControlID string
	tertiaryControlID  string
	subcontrolID       string
	evidenceID         string
	policyID           string
	// forwardMappingID maps primary → secondary
	forwardMappingID string
	// reverseMappingID maps secondary → primary, creating a duplicate reference to secondary
	// that the deduplication logic should collapse to a single relatedControl entry
	reverseMappingID string
	// tertiaryMappingID maps primary → tertiary, adding one more unique related control
	tertiaryMappingID string
}

// seedControlReportTestData enriches primaryControlID with a subcontrol, linked evidence,
// an internal policy, and three mapped controls to exercise deduplication.
// All three controls must already exist in the context's org.
//
// Mappings created:
//   - forward:  primary → secondary
//   - reverse:  secondary → primary  (secondary appears twice, deduped to one relatedControl)
//   - tertiary: primary → tertiary   (unique second related control)
//
// Expected relatedControls for primary: [secondary, tertiary] (2 unique entries).
func seedControlReportTestData(ctx context.Context, t *testing.T, primaryControlID, secondaryControlID, tertiaryControlID string) *controlReportTestData {
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

	return &controlReportTestData{
		primaryControlID:   primaryControlID,
		secondaryControlID: secondaryControlID,
		tertiaryControlID:  tertiaryControlID,
		subcontrolID:       sc.ID,
		evidenceID:         ev.ID,
		policyID:           policy.ID,
		forwardMappingID:   forward.ID,
		reverseMappingID:   reverse.ID,
		tertiaryMappingID:  tertiary.ID,
	}
}

func TestQueryControlReports(t *testing.T) {
	t.Parallel()

	localTestOrg := suite.seedOrgOwner(t)
	orgUser := suite.seedOrgOwner(t)

	orgOwnedCount := int64(11)
	systemOwnedCount := int64(3)
	controlIDs := []string{}

	for range orgOwnedCount {
		control := (&ControlBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)
		controlIDs = append(controlIDs, control.ID)
	}

	// system-owned controls must not appear in controlReports results
	for range systemOwnedCount {
		(&ControlBuilder{client: suite.client, SystemOwned: lo.ToPtr(true)}).MustNew(localTestOrg.owner.UserCtx, t)
	}

	// enrich the first three org-owned controls with associated data so enrichment paths are exercised
	richData := seedControlReportTestData(localTestOrg.owner.UserCtx, t, controlIDs[0], controlIDs[1], controlIDs[2])

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
			name:            "happy path, with last set",
			last:            lo.ToPtr(int64(3)),
			ctx:             localTestOrg.owner.UserCtx,
			expectedResults: 3,
		},
		{
			name:            "first set over max (10 in test)",
			first:           &orgOwnedCount,
			ctx:             localTestOrg.owner.UserCtx,
			expectedResults: testutils.MaxResultLimit,
		},
		{
			name:            "last set over max (10 in test)",
			last:            &orgOwnedCount,
			ctx:             localTestOrg.owner.UserCtx,
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
						// deduplication collapses it to one entry, plus tertiary = 2 total
						assert.Check(t, is.Len(edge.Node.RelatedControls, 2))
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

	control1 := (&ControlBuilder{client: suite.client, Category: cat1}).MustNew(localTestOrg.owner.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client, Category: cat1}).MustNew(localTestOrg.owner.UserCtx, t)
	control3 := (&ControlBuilder{client: suite.client, Category: cat2}).MustNew(localTestOrg.owner.UserCtx, t)
	(&ControlBuilder{client: suite.client}).MustNew(localTestOrg.owner.UserCtx, t)

	// system-owned controls must not appear in results regardless of category
	(&ControlBuilder{client: suite.client, Category: cat1, SystemOwned: lo.ToPtr(true)}).MustNew(localTestOrg.owner.UserCtx, t)
	(&ControlBuilder{client: suite.client, Category: cat2, SystemOwned: lo.ToPtr(true)}).MustNew(localTestOrg.owner.UserCtx, t)

	// enrich control1 with associated data so enrichment paths are exercised;
	// control3 is used as the tertiary to confirm a second unique related control
	richData := seedControlReportTestData(localTestOrg.owner.UserCtx, t, control1.ID, control2.ID, control3.ID)

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
							// deduplication collapses it to one entry, plus tertiary = 2 total
							assert.Check(t, is.Len(c.RelatedControls, 2))
						}
					}

					catCounts[cat.Category] = len(cat.Controls)
					totalControls += len(cat.Controls)
				}

				assert.Check(t, is.Equal(4, totalControls))
				assert.Check(t, is.Equal(2, catCounts[cat1]))
				assert.Check(t, is.Equal(1, catCounts[cat2]))
				assert.Check(t, is.Equal(1, catCounts[""]))
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
