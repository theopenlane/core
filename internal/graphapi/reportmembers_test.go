package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

func TestMutation_ReportMembers(t *testing.T) {

	org := suite.SeedFreshMinimalOrgUsers(t, false)

	t.Cleanup(func() {
		th.CleanupOrganizationDataWithContext(org.Owner.UserCtx, t)
	})

	newMember := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(org.Owner.UserCtx, t, &newMember, enums.RoleMember, org.Owner.OrganizationID)

	report := (&th.ReportBuilder{
		Client:        suite.Client,
		Visibility:    enums.ReportVisibilityRestricted,
		ReportOwnerID: newMember.ID,
	}).MustNew(org.Owner.UserCtx, t)

	var membershipID string

	t.Run("adding a member", func(t *testing.T) {
		resp, err := suite.Client.API.AddUserToReportWithRole(org.Owner.UserCtx, testclient.CreateReportMembershipInput{
			ReportID: report.ID,
			UserID:   org.Member.ID,
			Role:     &enums.RoleMember,
		})
		assert.NilError(t, err)

		membershipID = resp.CreateReportMembership.ReportMembership.ID

		assert.Check(t, is.Equal(resp.CreateReportMembership.ReportMembership.Role, enums.RoleMember))
		assert.Check(t, is.Equal(resp.CreateReportMembership.ReportMembership.UserID, org.Member.ID))

		_, err = suite.Client.API.GetReportByID(org.Member.UserCtx, report.ID)
		assert.NilError(t, err)

		_, err = suite.Client.API.AddUserToReportWithRole(org.Member.UserCtx, testclient.CreateReportMembershipInput{
			ReportID: report.ID,
			UserID:   newMember.ID,
			Role:     &enums.RoleMember,
		})
		assert.Assert(t, err != nil)
	})

	t.Run("removing a member from a report revokes their access", func(t *testing.T) {
		_, err := suite.Client.API.GetReportByID(org.Member.UserCtx, report.ID)
		assert.NilError(t, err)

		removed, err := suite.Client.API.RemoveUserFromReport(org.Owner.UserCtx, membershipID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(removed.DeleteReportMembership.DeletedID, membershipID))

		_, err = suite.Client.API.GetReportByID(org.Member.UserCtx, report.ID)
		assert.Assert(t, err != nil)
	})

	t.Run("restricted visibility requires report membership", func(t *testing.T) {

		// move it to org level first and make sure org members can view
		// then flip to restricted after which they should lose their access
		_, err := suite.Client.API.UpdateReport(org.Owner.UserCtx, report.ID, testclient.UpdateReportInput{
			Visibility: &enums.ReportVisibilityOrganization,
		})
		assert.NilError(t, err)

		_, err = suite.Client.API.GetReportByID(newMember.UserCtx, report.ID)
		assert.NilError(t, err)

		_, err = suite.Client.API.UpdateReport(org.Owner.UserCtx, report.ID, testclient.UpdateReportInput{
			Visibility: &enums.ReportVisibilityRestricted,
		})
		assert.NilError(t, err)

		_, err = suite.Client.API.GetReportByID(newMember.UserCtx, report.ID)
		assert.Assert(t, err != nil)

		_, err = suite.Client.API.AddUserToReportWithRole(org.Owner.UserCtx, testclient.CreateReportMembershipInput{
			ReportID: report.ID,
			UserID:   newMember.ID,
			Role:     &enums.RoleMember,
		})
		assert.NilError(t, err)

		_, err = suite.Client.API.GetReportByID(newMember.UserCtx, report.ID)
		assert.NilError(t, err)
	})

	t.Run("changing report to private visibility removes all members automatically", func(t *testing.T) {
		_, err := suite.Client.API.AddUserToReportWithRole(org.Owner.UserCtx, testclient.CreateReportMembershipInput{
			ReportID: report.ID,
			UserID:   org.Member.ID,
			Role:     &enums.RoleOwner,
		})
		assert.NilError(t, err)

		// user should be able to view before we change to private then revocation should happen
		_, err = suite.Client.API.GetReportByID(org.Member.UserCtx, report.ID)
		assert.NilError(t, err)

		_, err = suite.Client.API.UpdateReport(org.Owner.UserCtx, report.ID, testclient.UpdateReportInput{
			Visibility: &enums.ReportVisibilityPrivate,
		})
		assert.NilError(t, err)

		members, err := suite.Client.API.GetReportMembersByReportID(org.Owner.UserCtx, report.ID)
		assert.NilError(t, err)

		// only the creator should be left
		assert.Assert(t, is.Len(members.Report.Members.Edges, 1))

		_, err = suite.Client.API.GetReportByID(org.Member.UserCtx, report.ID)
		assert.Assert(t, err != nil)
	})
}

func TestMutation_ChangeOwnerRole(t *testing.T) {

	tt := []struct {
		name       string
		visibility enums.ReportVisibility
		hasError   bool
	}{
		{
			name:       "owner can be demoted to member",
			visibility: enums.ReportVisibilityOrganization,
		},
		{
			name:       "restricted owner membership can be demoted to member",
			visibility: enums.ReportVisibilityRestricted,
		},
		{
			name:       "private report prevents membership updates",
			visibility: enums.ReportVisibilityPrivate,
			hasError:   true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			report := (&th.ReportBuilder{
				Client:        suite.Client,
				Visibility:    tc.visibility,
				ReportOwnerID: th.SharedViewOnlyUser.ID,
			}).MustNew(th.SharedTestUser1.UserCtx, t)

			t.Cleanup(func() {
				(&th.Cleanup[*generated.ReportDeleteOne]{Client: suite.Client.DB.Report, ID: report.ID}).
					MustDelete(th.SharedTestUser1.UserCtx, t)
			})

			members, err := suite.Client.API.GetReportMembersByReportID(th.SharedTestUser1.UserCtx, report.ID)
			assert.NilError(t, err)
			assert.Assert(t, is.Len(members.Report.Members.Edges, 1))

			_, err = suite.Client.API.UpdateReportMembership(th.SharedTestUser1.UserCtx, members.Report.Members.Edges[0].Node.ID,
				testclient.UpdateReportMembershipInput{Role: &enums.RoleMember})

			if tc.hasError {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err)

			resp, err := suite.Client.API.GetReportByID(th.SharedTestUser1.UserCtx, report.ID)
			assert.NilError(t, err)

			assert.Check(t, is.Equal(lo.FromPtr(resp.Report.ReportOwnerID), th.SharedViewOnlyUser.ID))
			assert.Check(t, is.Equal(lo.FromPtr(resp.Report.Visibility), tc.visibility))

			members, err = suite.Client.API.GetReportMembersByReportID(th.SharedTestUser1.UserCtx, report.ID)
			assert.NilError(t, err)

			assert.Assert(t, is.Len(members.Report.Members.Edges, 1))
		})
	}
}
