package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
)

func TestMutationCreateReport(t *testing.T) {

	validQuery := `query ReportTasks($first: Int!) {
  tasks(first: $first) {
    edges {
      node {
        id
        title
        status
      }
    }
  }
}`

	variables := map[string]any{"first": float64(10)}

	tt := []struct {
		name                 string
		input                testclient.CreateReportInput
		client               *testclient.TestClient
		ctx                  context.Context
		expectedVisibility   enums.ReportVisibility
		expectedMembersCount int
		ownerID              string
		hasError             bool
	}{
		{
			name: "defaults to private with an owner membership",
			input: testclient.CreateReportInput{
				Name:      "My tasks",
				Query:     validQuery,
				Variables: variables,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedVisibility:   enums.ReportVisibilityPrivate,
			expectedMembersCount: 1,
		},
		{
			name: "can create a restricted report",
			input: testclient.CreateReportInput{
				Name:       "Shared tasks",
				Query:      validQuery,
				Variables:  variables,
				Visibility: &enums.ReportVisibilityRestricted,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedVisibility:   enums.ReportVisibilityRestricted,
			expectedMembersCount: 1,
		},
		{
			name: "organization level report",
			input: testclient.CreateReportInput{
				Name:       "Organization tasks",
				Query:      validQuery,
				Variables:  variables,
				Visibility: &enums.ReportVisibilityOrganization,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedVisibility:   enums.ReportVisibilityOrganization,
			expectedMembersCount: 1,
		},
		{
			name: "userless report requires explicit organization visibility",
			input: testclient.CreateReportInput{
				Name:       "tasks from api",
				Query:      validQuery,
				Variables:  variables,
				Visibility: &enums.ReportVisibilityOrganization,
			},
			client:             suite.Client.APIWithToken,
			ctx:                context.Background(),
			expectedVisibility: enums.ReportVisibilityOrganization,
			// since this was created by an api token, we do not have a "creator"
			expectedMembersCount: 0,
		},
		{
			name: "userless report cannot omit visibility",
			input: testclient.CreateReportInput{
				Name:      "tasks from api without visibility",
				Query:     validQuery,
				Variables: variables,
			},
			client:   suite.Client.APIWithToken,
			ctx:      context.Background(),
			hasError: true,
		},
		{
			name: "userless report cannot use private visibility",
			input: testclient.CreateReportInput{
				Name:       "private tasks from api",
				Query:      validQuery,
				Variables:  variables,
				Visibility: &enums.ReportVisibilityPrivate,
			},
			client:   suite.Client.APIWithToken,
			ctx:      context.Background(),
			hasError: true,
		},
		{
			name: "userless report cannot use restricted visibility",
			input: testclient.CreateReportInput{
				Name:       "restricted tasks from api",
				Query:      validQuery,
				Variables:  variables,
				Visibility: &enums.ReportVisibilityRestricted,
			},
			client:   suite.Client.APIWithToken,
			ctx:      context.Background(),
			hasError: true,
		},
		{
			name: "invalid query cannot be stored",
			input: testclient.CreateReportInput{
				Name: "Invalid report",
				Query: `query InvalidReport {
  tasks {
    edges {
      node {
        nonexistentField
      }
    }
  }
}`,
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			hasError: true,
		},
		{
			name: "invalid variables cannot be stored",
			input: testclient.CreateReportInput{
				Name:      "Invalid variables",
				Query:     validQuery,
				Variables: map[string]any{"first": "ten"},
			},
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			hasError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateReport(tc.ctx, tc.input)
			if tc.hasError {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err)

			t.Cleanup(func() {
				(&th.Cleanup[*generated.ReportDeleteOne]{Client: suite.Client.DB.Report, ID: resp.CreateReport.Report.ID}).
					MustDelete(th.SharedTestUser1.UserCtx, t)
			})

			assert.Check(t, is.Equal(lo.FromPtr(resp.CreateReport.Report.Visibility), tc.expectedVisibility))
			assert.Check(t, is.Equal(lo.FromPtr(resp.CreateReport.Report.ReportOwnerID), tc.ownerID))

			members, err := suite.Client.API.GetReportMembersByReportID(th.SharedTestUser1.UserCtx, resp.CreateReport.Report.ID)
			assert.NilError(t, err)

			assert.Assert(t, is.Len(members.Report.Members.Edges, tc.expectedMembersCount))
		})
	}
}

func TestMutationReport_Member(t *testing.T) {

	org := suite.SeedFreshMinimalOrgUsers(t, false)

	t.Cleanup(func() {
		th.CleanupOrganizationDataWithContext(org.Owner.UserCtx, t)
	})

	input := testclient.CreateReportInput{
		Name: "report created by a member",
		Query: `query ReportTasks {
  tasks(first: 10) {
    edges {
      node {
        id
        title
      }
    }
  }
}`,
	}

	resp, err := suite.Client.API.CreateReport(org.Member.UserCtx, input)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(resp.CreateReport.Report.Query, input.Query))
	assert.Check(t, is.Equal(lo.FromPtr(resp.CreateReport.Report.Visibility), enums.ReportVisibilityPrivate))
}

func TestMutationDeleteReport(t *testing.T) {

	report := (&th.ReportBuilder{Client: suite.Client, Visibility: enums.ReportVisibilityOrganization}).
		MustNew(th.SharedTestUser1.UserCtx, t)

	anotherReport := (&th.ReportBuilder{Client: suite.Client, Visibility: enums.ReportVisibilityOrganization}).
		MustNew(th.SharedTestUser1.UserCtx, t)

	orgReport, err := suite.Client.APIWithToken.CreateReport(context.Background(), testclient.CreateReportInput{
		Name:       "Ownerless report",
		Query:      report.Query,
		Visibility: &enums.ReportVisibilityOrganization,
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.AddUserToReportWithRole(th.SharedTestUser1.UserCtx, testclient.CreateReportMembershipInput{
		ReportID: report.ID,
		UserID:   th.SharedViewOnlyUser.ID,
		Role:     &enums.RoleMember,
	})
	assert.NilError(t, err)

	tt := []struct {
		name     string
		reportID string
		ctx      context.Context
		hasError bool
	}{
		{
			name:     "user from another org cannot delete",
			reportID: report.ID,
			ctx:      th.SharedTestUser2.UserCtx,
			hasError: true,
		},
		{
			name:     "report member cannot delete",
			reportID: report.ID,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			hasError: true,
		},
		{
			name:     "organization view only member cannot delete",
			reportID: report.ID,
			ctx:      th.SharedViewOnlyUser2.UserCtx,
			hasError: true,
		},
		{
			name:     "organization admin is not report owner",
			reportID: report.ID,
			ctx:      th.SharedAdminUser.UserCtx,
			hasError: true,
		},
		{
			name:     "report owner",
			reportID: report.ID,
			ctx:      th.SharedTestUser1.UserCtx,
		},
		{
			name:     "organization super admin can delete another user's report",
			reportID: anotherReport.ID,
			ctx:      th.SharedSuperAdminUser.UserCtx,
		},
		{
			name:     "organization viewer cannot delete an ownerless report",
			reportID: orgReport.CreateReport.Report.ID,
			ctx:      th.SharedViewOnlyUser.UserCtx,
			hasError: true,
		},
		{
			name:     "organization admin cannot delete an ownerless report",
			reportID: orgReport.CreateReport.Report.ID,
			ctx:      th.SharedAdminUser.UserCtx,
			hasError: true,
		},
		{
			name:     "organization super admin can delete an ownerless report",
			reportID: orgReport.CreateReport.Report.ID,
			ctx:      th.SharedSuperAdminUser.UserCtx,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := suite.Client.API.DeleteReport(tc.ctx, tc.reportID)
			if tc.hasError {
				assert.Assert(t, err != nil)
				_, err = suite.Client.API.GetReportByID(th.SharedTestUser1.UserCtx, tc.reportID)
				assert.NilError(t, err)
				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(resp.DeleteReport.DeletedID, tc.reportID))

		})
	}

	// make sure they were deleted
	where := &testclient.ReportWhereInput{
		IDIn: []string{report.ID, anotherReport.ID, orgReport.CreateReport.Report.ID},
	}

	reports, err := suite.Client.API.GetReports(th.SharedSuperAdminUser.UserCtx, nil, nil, nil, nil, nil, where)

	assert.NilError(t, err)
	assert.Check(t, is.Len(reports.Reports.Edges, 0))
}

func TestMutationReport_AddGroup(t *testing.T) {

	newUser := suite.UserBuilder(context.Background(), t)
	t.Cleanup(func() {
		th.CleanupOrganizationDataWithContext(newUser.UserCtx, t)
	})

	tt := []struct {
		name       string
		withEditor bool
		visibility enums.ReportVisibility
		hasError   bool
	}{
		{
			name:       "cannot add group to a private report",
			visibility: enums.ReportVisibilityPrivate,
			hasError:   true,
		},
		{
			name:       "cannot add editor group to a private report",
			visibility: enums.ReportVisibilityPrivate,
			withEditor: true,
			hasError:   true,
		},
		{
			name:       "can add group to a restricted report",
			visibility: enums.ReportVisibilityRestricted,
		},
		{
			name:       "can add editor group to a restricted report",
			visibility: enums.ReportVisibilityRestricted,
			withEditor: true,
		},
		{
			name:       "can add group to organization level report",
			visibility: enums.ReportVisibilityOrganization,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			report := (&th.ReportBuilder{Client: suite.Client, Visibility: tc.visibility}).MustNew(newUser.UserCtx, t)

			group, err := suite.Client.API.CreateGroup(newUser.UserCtx, testclient.CreateGroupInput{
				Name: tc.name,
			})
			assert.NilError(t, err)

			input := testclient.UpdateReportInput{
				AddViewerIDs: []string{group.CreateGroup.Group.ID},
			}

			if tc.withEditor {
				input.AddViewerIDs = nil
				input.AddEditorIDs = []string{group.CreateGroup.Group.ID}
			}

			_, err = suite.Client.API.UpdateReport(newUser.UserCtx, report.ID, input)
			assert.Equal(t, err != nil, tc.hasError)

			resp, err := suite.Client.API.GetReportByID(newUser.UserCtx, report.ID)
			assert.NilError(t, err)

			if tc.hasError {
				assert.Check(t, is.Len(resp.Report.Editors.Edges, 0))
				assert.Check(t, is.Len(resp.Report.Viewers.Edges, 0))
				return
			}

			if tc.withEditor {
				assert.Assert(t, is.Len(resp.Report.Editors.Edges, 1))
				assert.Check(t, is.Equal(resp.Report.Editors.Edges[0].Node.ID, group.CreateGroup.Group.ID))
				return
			}

			assert.Assert(t, is.Len(resp.Report.Viewers.Edges, 1))
			assert.Check(t, is.Equal(resp.Report.Viewers.Edges[0].Node.ID, group.CreateGroup.Group.ID))
		})
	}
}

func TestMutationReport_Group(t *testing.T) {

	org := suite.SeedFreshMinimalOrgUsers(t, false)

	t.Cleanup(func() {
		th.CleanupOrganizationDataWithContext(org.Owner.UserCtx, t)
	})

	editorGroupUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(org.Owner.UserCtx, t, &editorGroupUser, enums.RoleMember, org.Owner.OrganizationID)

	report := (&th.ReportBuilder{
		Client:     suite.Client,
		Visibility: enums.ReportVisibilityRestricted,
	}).MustNew(org.Owner.UserCtx, t)

	viewerGroup, err := suite.Client.API.CreateGroup(org.Owner.UserCtx, testclient.CreateGroupInput{
		Name:            "Report viewers",
		ReportViewerIDs: []string{report.ID},
	})
	assert.NilError(t, err)

	editorGroup, err := suite.Client.API.CreateGroup(org.Owner.UserCtx, testclient.CreateGroupInput{
		Name:            "Report editors",
		ReportEditorIDs: []string{report.ID},
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.AddUserToGroupWithRole(org.Owner.UserCtx, testclient.CreateGroupMembershipInput{
		GroupID: viewerGroup.CreateGroup.Group.ID,
		UserID:  org.Member.ID,
		Role:    &enums.RoleMember,
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.AddUserToGroupWithRole(org.Owner.UserCtx, testclient.CreateGroupMembershipInput{
		GroupID: editorGroup.CreateGroup.Group.ID,
		UserID:  editorGroupUser.ID,
		Role:    &enums.RoleMember,
	})
	assert.NilError(t, err)

	t.Run("viewer group can read but no edit access", func(t *testing.T) {
		_, err := suite.Client.API.GetReportByID(org.Member.UserCtx, report.ID)
		assert.NilError(t, err)

		_, err = suite.Client.API.UpdateReport(org.Member.UserCtx, report.ID, testclient.UpdateReportInput{
			Name: lo.ToPtr("Here is a new query name"),
		})
		assert.Assert(t, err != nil)
	})

	t.Run("editor group has read + edit access", func(t *testing.T) {
		_, err := suite.Client.API.GetReportByID(editorGroupUser.UserCtx, report.ID)
		assert.NilError(t, err)

		_, err = suite.Client.API.UpdateReport(editorGroupUser.UserCtx, report.ID, testclient.UpdateReportInput{
			Name: lo.ToPtr("someone in editor group updated this"),
		})
		assert.NilError(t, err)
	})

	// now move that same report into private visibility
	// then make sure the editors + viewer edges no longer exists and they cannot view/edit again
	_, err = suite.Client.API.UpdateReport(org.Owner.UserCtx, report.ID, testclient.UpdateReportInput{
		Visibility: &enums.ReportVisibilityPrivate,
	})
	assert.NilError(t, err)

	newReport, err := suite.Client.API.GetReportByID(org.Owner.UserCtx, report.ID)
	assert.NilError(t, err)

	assert.Check(t, is.Len(newReport.Report.Editors.Edges, 0))
	assert.Check(t, is.Len(newReport.Report.Viewers.Edges, 0))

	t.Run("existing viewer group lost read access", func(t *testing.T) {
		_, err := suite.Client.API.GetReportByID(org.Member.UserCtx, report.ID)
		assert.Assert(t, err != nil)

		_, err = suite.Client.API.UpdateReport(org.Member.UserCtx, report.ID, testclient.UpdateReportInput{
			Name: lo.ToPtr("failed update"),
		})
		assert.Assert(t, err != nil)
	})

	t.Run("previous editor group no longer has edit access", func(t *testing.T) {
		_, err := suite.Client.API.GetReportByID(editorGroupUser.UserCtx, report.ID)
		assert.Assert(t, err != nil)

		_, err = suite.Client.API.UpdateReport(editorGroupUser.UserCtx, report.ID, testclient.UpdateReportInput{
			Name: lo.ToPtr("Editor access was revoked"),
		})
		assert.Assert(t, err != nil)
	})
}

func TestMutationReportGroup_Permission(t *testing.T) {
	org := suite.SeedFreshMinimalOrgUsers(t, false)
	t.Cleanup(func() {
		th.CleanupOrganizationDataWithContext(org.Owner.UserCtx, t)
	})

	report := (&th.ReportBuilder{Client: suite.Client, Visibility: enums.ReportVisibilityRestricted}).
		MustNew(org.Owner.UserCtx, t)

	group, err := suite.Client.API.CreateGroup(org.Owner.UserCtx, testclient.CreateGroupInput{
		Name: "Group managed without report edit access",
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.AddUserToGroupWithRole(org.Owner.UserCtx, testclient.CreateGroupMembershipInput{
		GroupID: group.CreateGroup.Group.ID,
		UserID:  org.Admin.ID,
		Role:    &enums.RoleAdmin,
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.UpdateGroup(org.Admin.UserCtx, group.CreateGroup.Group.ID, testclient.UpdateGroupInput{
		Description: lo.ToPtr("Caller can edit this group"),
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.GetReportByID(org.Admin.UserCtx, report.ID)
	assert.Assert(t, err != nil)

	t.Run("no report access cannot grant viewer access", func(t *testing.T) {
		_, err := suite.Client.API.CreateGroup(org.Admin.UserCtx, testclient.CreateGroupInput{
			Name:            "no report access cannot grant viewer access",
			ReportViewerIDs: []string{report.ID},
		})
		assert.Assert(t, err != nil)

		_, err = suite.Client.API.UpdateGroup(org.Admin.UserCtx, group.CreateGroup.Group.ID, testclient.UpdateGroupInput{
			AddReportViewerIDs: []string{report.ID},
		})
		assert.Assert(t, err != nil)
	})

	t.Run("no report access cannot grant editor access", func(t *testing.T) {
		_, err := suite.Client.API.CreateGroup(org.Admin.UserCtx, testclient.CreateGroupInput{
			Name:            "no report access cannot grant editor access",
			ReportEditorIDs: []string{report.ID},
		})
		assert.Assert(t, err != nil)

		_, err = suite.Client.API.UpdateGroup(org.Admin.UserCtx, group.CreateGroup.Group.ID, testclient.UpdateGroupInput{
			AddReportEditorIDs: []string{report.ID},
		})
		assert.Assert(t, err != nil)
	})

	_, err = suite.Client.API.AddUserToReportWithRole(org.Owner.UserCtx, testclient.CreateReportMembershipInput{
		ReportID: report.ID,
		UserID:   org.Admin.ID,
		Role:     &enums.RoleMember,
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.GetReportByID(org.Admin.UserCtx, report.ID)
	assert.NilError(t, err)

	t.Run("report member cannot grant viewer access", func(t *testing.T) {
		_, err := suite.Client.API.CreateGroup(org.Admin.UserCtx, testclient.CreateGroupInput{
			Name:            "report member cannot grant viewer access",
			ReportViewerIDs: []string{report.ID},
		})
		assert.Assert(t, err != nil)

		_, err = suite.Client.API.UpdateGroup(org.Admin.UserCtx, group.CreateGroup.Group.ID, testclient.UpdateGroupInput{
			AddReportViewerIDs: []string{report.ID},
		})
		assert.Assert(t, err != nil)
	})

	t.Run("report member cannot grant editor access", func(t *testing.T) {
		_, err := suite.Client.API.CreateGroup(org.Admin.UserCtx, testclient.CreateGroupInput{
			Name:            "report member cannot grant editor access",
			ReportEditorIDs: []string{report.ID},
		})
		assert.Assert(t, err != nil)

		_, err = suite.Client.API.UpdateGroup(org.Admin.UserCtx, group.CreateGroup.Group.ID, testclient.UpdateGroupInput{
			AddReportEditorIDs: []string{report.ID},
		})
		assert.Assert(t, err != nil)
	})

	unchanged, err := suite.Client.API.GetReportByID(org.Owner.UserCtx, report.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Len(unchanged.Report.Editors.Edges, 0))
	assert.Check(t, is.Len(unchanged.Report.Viewers.Edges, 0))
}
