package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestQueryInternalPolicy(t *testing.T) {
	// create an InternalPolicy to be queried using th.SharedTestUser1
	internalPolicy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// setup a blocked group with a view only user
	blockedGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	(&th.GroupMemberBuilder{Client: suite.Client, UserID: th.SharedViewOnlyUser.ID, GroupID: blockedGroup.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	internalPolicy2 := (&th.InternalPolicyBuilder{Client: suite.Client, BlockedGroupIDs: []string{blockedGroup.ID}}).MustNew(th.SharedTestUser1.UserCtx, t)
	anonymousContext := th.CreateAnonymousTrustCenterContext(ulids.New().String(), th.SharedTestUser1.OrganizationID)

	// add test cases for querying the internal policy
	testCases := []struct {
		name               string
		queryID            string
		client             *testclient.TestClient
		ctx                context.Context
		errorMsg           string
		updateBlockedGroup bool
	}{
		{
			name:    "happy path",
			queryID: internalPolicy.ID,
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: internalPolicy.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:               "happy path, read only user but blocked",
			queryID:            internalPolicy2.ID,
			client:             suite.Client.API,
			ctx:                th.SharedViewOnlyUser.UserCtx,
			errorMsg:           th.NotFoundErrorMsg, // should not be able to access the policy due to blocked group
			updateBlockedGroup: true,
		},
		{
			name:    "happy path, read only user no longer blocked",
			queryID: internalPolicy2.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: internalPolicy.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "internalPolicy not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "internal policy not found, using not authorized user",
			queryID:  internalPolicy.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "no access, anonymous user",
			client:   suite.Client.API,
			ctx:      anonymousContext,
			queryID:  internalPolicy.ID,
			errorMsg: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetInternalPolicyByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				if tc.updateBlockedGroup {
					_, err := suite.Client.API.UpdateInternalPolicy(th.SharedTestUser1.UserCtx, internalPolicy2.ID,
						testclient.UpdateInternalPolicyInput{
							RemoveBlockedGroupIDs: []string{blockedGroup.ID},
						})
					assert.NilError(t, err)
				}

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.InternalPolicy.ID))
			assert.Check(t, len(resp.InternalPolicy.Name) != 0)
		})
	}

	// cleanup
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{internalPolicy.ID, internalPolicy2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryInternalPolicies(t *testing.T) {
	// create multiple policies to be queried using th.SharedTestUser1
	ip1 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	ip2 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// setup a blocked group with a view only user
	blockedGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	(&th.GroupMemberBuilder{Client: suite.Client, UserID: th.SharedViewOnlyUser.ID, GroupID: blockedGroup.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	ip3 := (&th.InternalPolicyBuilder{Client: suite.Client, BlockedGroupIDs: []string{blockedGroup.ID}}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name               string
		client             *testclient.TestClient
		ctx                context.Context
		updateBlockedGroup bool
		expectedResults    int
	}{
		{
			name:            "happy path",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 3,
		},
		{
			name:               "happy path, using read only user of the same org, one policy blocked",
			client:             suite.Client.API,
			ctx:                th.SharedViewOnlyUser.UserCtx,
			expectedResults:    2,    // should not see the policy that is blocked for them
			updateBlockedGroup: true, // update the blocked group to allow the view only user to see the policy
		},
		{
			name:            "happy path, using read only user of the same org, no blocked group",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 3, // should now see all policies after removing the blocked group
		},
		{
			name:            "happy path, using api token",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "happy path, using pat",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "another user, no policies should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllInternalPolicies(tc.ctx, nil, nil, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.InternalPolicies.Edges, tc.expectedResults))

			if tc.updateBlockedGroup {
				// do it the opposite, remove the policy from the group
				_, err := suite.Client.API.UpdateGroup(th.SharedTestUser1.UserCtx, blockedGroup.ID,
					testclient.UpdateGroupInput{
						RemoveInternalPolicyBlockedGroupIDs: []string{ip3.ID},
					},
				)

				assert.NilError(t, err)
			}
		})
	}

	// delete created policies
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{ip1.ID, ip2.ID, ip3.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateInternalPolicy(t *testing.T) {
	// create a system owned standard with a control
	systemStandard := (&th.StandardBuilder{Client: suite.Client}).MustNew(th.SharedSystemAdminUser.UserCtx, t)
	// create a control and add it to the system standard
	systemControl := (&th.ControlBuilder{Client: suite.Client, StandardID: systemStandard.ID}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	anotherGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// group for the view only user
	groupMember := (&th.GroupMemberBuilder{Client: suite.Client, UserID: th.SharedViewOnlyUser.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// approver and delegator groups for the test user
	approverGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	delegateGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// edges to add
	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	subcontrol := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	task := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name                       string
		request                    testclient.CreateInternalPolicyInput
		addGroupToOrg              bool
		controlEdgeShouldBeCreated bool
		client                     *testclient.TestClient
		ctx                        context.Context
		expectedErr                string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all input except edges",
			request: testclient.CreateInternalPolicyInput{
				Name:       "Releasing a new version",
				Status:     &enums.DocumentDraft,
				Revision:   lo.ToPtr("v1.1.0"),
				Details:    lo.ToPtr("do stuff"),
				ApproverID: &approverGroup.ID,
				DelegateID: &delegateGroup.ID,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, long details",
			request: testclient.CreateInternalPolicyInput{
				Name:       "Releasing a new version",
				Status:     &enums.DocumentDraft,
				Revision:   lo.ToPtr("v1.1.0"),
				Details:    lo.ToPtr(gofakeit.Sentence()),
				ApproverID: &approverGroup.ID,
				DelegateID: &delegateGroup.ID,
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, with control edges",
			request: testclient.CreateInternalPolicyInput{
				Name:          "Releasing a new version",
				Status:        &enums.DocumentDraft,
				Revision:      lo.ToPtr("v1.1.0"),
				Details:       lo.ToPtr("do stuff"),
				ControlIDs:    []string{control.ID},
				SubcontrolIDs: []string{subcontrol.ID},
				TaskIDs:       []string{task.ID},
			},
			client:                     suite.Client.API,
			ctx:                        th.SharedTestUser1.UserCtx,
			controlEdgeShouldBeCreated: true,
		},
		{
			name: "should not be allowed to add system standard control",
			request: testclient.CreateInternalPolicyInput{
				Name:       "Releasing a new version",
				Status:     &enums.DocumentDraft,
				ControlIDs: []string{systemControl.ID},
			},
			client:                     suite.Client.API,
			ctx:                        th.SharedTestUser1.UserCtx,
			expectedErr:                th.NotFoundErrorMsg,
			controlEdgeShouldBeCreated: false, // user does not have edit access to the control, it is owned by the system
		},
		{
			name: "happy path, add editor group",
			request: testclient.CreateInternalPolicyInput{
				Name:      "Test Policy",
				EditorIDs: []string{th.SharedTestUser1.GroupID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, add same task to another policy",
			request: testclient.CreateInternalPolicyInput{
				Name:    "Test Policy",
				TaskIDs: []string{task.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, add same control to another policy",
			request: testclient.CreateInternalPolicyInput{
				Name:       "Test Policy",
				ControlIDs: []string{control.ID},
			},
			client:                     suite.Client.API,
			ctx:                        th.SharedTestUser1.UserCtx,
			controlEdgeShouldBeCreated: true,
		},
		{
			name: "happy path, add same sub control to another policy",
			request: testclient.CreateInternalPolicyInput{
				Name:          "Test Policy",
				SubcontrolIDs: []string{subcontrol.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "add editor group, again - ensures the same group can be added to multiple policies",
			request: testclient.CreateInternalPolicyInput{
				Name:            "Test Policy",
				EditorIDs:       []string{th.SharedTestUser1.GroupID},
				BlockedGroupIDs: []string{anotherGroup.ID},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateInternalPolicyInput{
				Name:    "Test Internal Policy",
				OwnerID: &th.SharedTestUser1.OrganizationID,
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: testclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			addGroupToOrg: true,
			client:        suite.Client.API,
			ctx:           th.SharedViewOnlyUser.UserCtx,
		},
		{
			name: "missing required field",
			request: testclient.CreateInternalPolicyInput{
				Details: lo.ToPtr("instructions on how to release a new version"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.Client.API.UpdateOrganization(th.SharedTestUser1.UserCtx, th.SharedTestUser1.OrganizationID,
					testclient.UpdateOrganizationInput{
						AddInternalPolicyCreatorIDs: []string{groupMember.GroupID},
					}, nil, nil)
				assert.NilError(t, err)
			}

			resp, err := tc.client.CreateInternalPolicy(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateInternalPolicy.InternalPolicy.Name))

			assert.Check(t, len(resp.CreateInternalPolicy.InternalPolicy.DisplayID) != 0)
			assert.Check(t, is.Contains(resp.CreateInternalPolicy.InternalPolicy.DisplayID, "PLC-"))

			// check optional fields with if checks if they were provided or not
			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.CreateInternalPolicy.InternalPolicy.Status))
			} else {
				assert.Check(t, is.Equal(enums.DocumentDraft, *resp.CreateInternalPolicy.InternalPolicy.Status))
			}

			if tc.request.Revision != nil {
				assert.Check(t, is.Equal(*tc.request.Revision, *resp.CreateInternalPolicy.InternalPolicy.Revision))
			} else {
				assert.Check(t, is.Equal(models.DefaultRevision, *resp.CreateInternalPolicy.InternalPolicy.Revision))
			}

			if tc.request.Details != nil {
				assert.Check(t, is.DeepEqual(tc.request.Details, resp.CreateInternalPolicy.InternalPolicy.Details))
				assert.Check(t, resp.CreateInternalPolicy.InternalPolicy.Summary != nil)
			} else {
				assert.Check(t, is.Equal(*resp.CreateInternalPolicy.InternalPolicy.Details, ""))
			}

			if tc.request.ApproverID != nil {
				assert.Check(t, resp.CreateInternalPolicy.InternalPolicy.ID != "")
				assert.Check(t, is.Equal(*tc.request.ApproverID, resp.CreateInternalPolicy.InternalPolicy.Approver.ID))
			} else {
				assert.Check(t, resp.CreateInternalPolicy.InternalPolicy.Approver == nil)
			}

			if tc.request.DelegateID != nil {
				assert.Check(t, is.Equal(*tc.request.DelegateID, resp.CreateInternalPolicy.InternalPolicy.Delegate.ID))
			} else {
				assert.Check(t, resp.CreateInternalPolicy.InternalPolicy.Delegate == nil)
			}

			if tc.request.ControlIDs != nil {
				for _, controlID := range tc.request.ControlIDs {
					controlFound := false
					for _, edge := range resp.CreateInternalPolicy.InternalPolicy.Controls.Edges {
						if controlID == edge.Node.ID {
							controlFound = true
							break
						}
					}

					assert.Check(t, is.Equal(controlFound, tc.controlEdgeShouldBeCreated))
				}
			}

			// cleanup
			(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{resp.CreateInternalPolicy.InternalPolicy.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}

	// cleanup
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{control.ID, subcontrol.ControlID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, IDs: []string{subcontrol.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, IDs: []string{task.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{anotherGroup.ID, groupMember.GroupID, approverGroup.ID, delegateGroup.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)

	// cleanup the system standard and control
	(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, IDs: []string{systemStandard.ID}}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
}

func TestMutationUpdateInternalPolicy(t *testing.T) {
	makeSlate := func(children ...any) []any {
		return []any{
			map[string]any{
				"type":     "paragraph",
				"children": children,
			},
		}
	}

	internalPolicy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	internalPolicyAdminUser := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedAdminUser.UserCtx, t)

	// create a viewer user and add them to the same organization as th.SharedTestUser1
	// also add them to the same group as th.SharedTestUser1, this should still allow them to edit the policy
	// despite not not being an organization admin
	anotherViewerUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &anotherViewerUser, enums.RoleMember, th.SharedTestUser1.OrganizationID)

	(&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherViewerUser.ID, GroupID: th.SharedTestUser1.GroupID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// group admins should also have edit permissions when added to the group
	anotherViewerGroupAdminUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &anotherViewerGroupAdminUser, enums.RoleMember, th.SharedTestUser1.OrganizationID)

	(&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherViewerGroupAdminUser.ID, GroupID: th.SharedTestUser1.GroupID, Role: enums.RoleAdmin.String()}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	blockGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	(&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherViewerUser.ID, GroupID: blockGroup.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	// edges to add
	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	subcontrol := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	task := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name             string
		policyID         string
		request          testclient.UpdateInternalPolicyInput
		client           *testclient.TestClient
		ctx              context.Context
		expectedErr      string
		expectedRevision string
	}{
		{
			name:     "happy path, update details field",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Details: lo.ToPtr(gofakeit.Sentence()),
			},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			expectedRevision: "v0.1.0", // details updated, should be a minor update
		},
		{
			name:     "happy path, update details field on policy created by another user",
			policyID: internalPolicyAdminUser.ID,
			request: testclient.UpdateInternalPolicyInput{
				Details: lo.ToPtr(gofakeit.Sentence()),
			},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx, // org owner should always be able to update the policy
			expectedRevision: "v0.1.0",                   // details updated, should be a minor update (different policy than test 1)
		},
		{
			name:     "happy path, update name field",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Name:         lo.ToPtr("Updated InternalPolicy Name"),
				AddEditorIDs: []string{th.SharedTestUser1.GroupID}, // add the group to the editor groups for subsequent tests
			},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			expectedRevision: "v0.1.1", // no details updated, should be a patch update
		},
		{
			name:     "happy path, update multiple fields",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Status:           &enums.DocumentPublished,
				Details:          lo.ToPtr("Updated details"),
				RevisionBump:     &models.Major,
				AddControlIDs:    []string{control.ID},
				AddSubcontrolIDs: []string{subcontrol.ID},
				AddTaskIDs:       []string{task.ID},
			},
			client:           suite.Client.APIWithPAT,
			ctx:              context.Background(),
			expectedRevision: "v1.0.0", // details updated, but revision bump set to major
		},
		{
			name:     "member allowed to add comment",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "This is a comment from a member user",
				},
			},
			client:           suite.Client.API,
			ctx:              th.SharedViewOnlyUser.UserCtx,
			expectedRevision: "v1.0.1", // only comment added, should be a patch update
		},
		{
			name:     "member not allowed to update details",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				AddComment: &testclient.CreateNoteInput{
					Text: "This is a comment from a member user",
				},
				DetailsJSON: makeSlate(map[string]any{"text": "hello"}), // should not be allowed to update the details, only add a comment
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:     "update not allowed, not enough permissions",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated InternalPolicy Name"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:     "update allowed, org admins have edit access to all policies in the org",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client:           suite.Client.API,
			ctx:              th.SharedAdminUser.UserCtx,
			expectedRevision: "v1.0.2", // no details updated, should be a patch update
		},
		{
			name:     "member update allowed, user in editor group",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client:           suite.Client.API,
			ctx:              anotherViewerUser.UserCtx, // user assigned to the group which has editor permissions
			expectedRevision: "v1.0.3",                  // no details updated, should be a patch update
		},
		{
			name:     "member update allowed, user in editor group as admin",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again by Group Admin"),
			},
			client:           suite.Client.API,
			ctx:              anotherViewerGroupAdminUser.UserCtx, // user assigned to the group which has editor permissions as admin
			expectedRevision: "v1.0.4",                            // no details updated, should be a patch update
		},
		{
			name:     "happy path, block the group from editing",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				AddBlockedGroupIDs: []string{blockGroup.ID}, // block the group
			},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			expectedRevision: "v1.0.5", // no details updated, should be a patch update
		},
		{
			name:     "member update no longer allowed, user in blocked group",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client:      suite.Client.API,
			ctx:         anotherViewerUser.UserCtx, // user assigned to the group which was blocked
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:     "happy path, remove the group",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				RemoveEditorIDs: []string{th.SharedTestUser1.GroupID}, // remove the group from the editor groups
			},
			client:           suite.Client.API,
			ctx:              th.SharedTestUser1.UserCtx,
			expectedRevision: "v1.0.6", // no details updated, should be a patch update
		},
		{
			name:     "update not allowed, editor group was removed",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again Again"),
			},
			client:      suite.Client.API,
			ctx:         anotherViewerUser.UserCtx, // user assigned to the group which no longer has editor permissions
			expectedErr: th.NotFoundErrorMsg,       // TODO: this will change back to not authorized on the new permissions branch
		},
		{
			name:     "update not allowed, no permissions",
			policyID: internalPolicy.ID,
			request: testclient.UpdateInternalPolicyInput{
				Details: lo.ToPtr("Updated details"),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateInternalPolicy(tc.ctx, tc.policyID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check updated fields
			if tc.request.Name != nil {
				assert.Check(t, is.Equal(*tc.request.Name, resp.UpdateInternalPolicy.InternalPolicy.Name))
			}

			if tc.request.Details != nil {
				assert.Check(t, resp.UpdateInternalPolicy.InternalPolicy.Summary != nil)
			}

			if tc.request.Status != nil {
				assert.Check(t, is.Equal(*tc.request.Status, *resp.UpdateInternalPolicy.InternalPolicy.Status))
			}

			assert.Check(t, is.Equal(*&tc.expectedRevision, *resp.UpdateInternalPolicy.InternalPolicy.Revision))

			if tc.request.RevisionBump == &models.Major {
				assert.Check(t, is.Equal("v1.0.0", *resp.UpdateInternalPolicy.InternalPolicy.Revision))
			}

			if tc.request.Details != nil {
				assert.Check(t, is.DeepEqual(tc.request.Details, resp.UpdateInternalPolicy.InternalPolicy.Details))
			}
		})
	}

	// cleanup
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{internalPolicy.ID, internalPolicyAdminUser.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: []string{control.ID, subcontrol.ControlID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, IDs: []string{subcontrol.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, IDs: []string{task.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, IDs: []string{blockGroup.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteInternalPolicy(t *testing.T) {
	// create internal policies to be deleted
	internalPolicy1 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	internalPolicy2 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  internalPolicy1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: internalPolicy1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  internalPolicy1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: internalPolicy2.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteInternalPolicy(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteInternalPolicy.DeletedID))
		})
	}
}

func TestMutationRoleChangesCanAccessPolicy(t *testing.T) {
	cases := []struct {
		name string
		role string
	}{
		{
			name: "policy manager",
			role: "policy_manager",
		},
		{
			name: "compliance manager",
			role: "compliance_manager",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			localTestOrg := suite.SeedFreshMinimalOrgUsers(t, false)
			defer th.CleanupOrganizationDataWithContext(localTestOrg.Owner.UserCtx, t)

			policyToUpdate := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(localTestOrg.Owner.UserCtx, t)
			policyToDelete := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(localTestOrg.Owner.UserCtx, t)

			_, err := suite.Client.API.UpdateInternalPolicy(localTestOrg.Member.UserCtx, policyToUpdate.ID, testclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("member cannot edit or delete without policy role"),
			})
			assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)

			_, err = suite.Client.API.DeleteInternalPolicy(localTestOrg.Member.UserCtx, policyToDelete.ID)
			assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)

			// manually add the tuple that the api does
			tuple := fgax.GetTupleKey(fgax.TupleRequest{
				SubjectID:   localTestOrg.Member.ID,
				SubjectType: "user",
				ObjectID:    localTestOrg.Owner.OrganizationID,
				ObjectType:  "organization",
				Relation:    tc.role,
			})

			_, err = suite.Client.DB.Authz.WriteTupleKeys(localTestOrg.Owner.UserCtx, []fgax.TupleKey{tuple}, nil)
			assert.NilError(t, err)

			updatedName := tc.role + " can edit or delete any policy"

			updateResp, err := suite.Client.API.UpdateInternalPolicy(localTestOrg.Member.UserCtx, policyToUpdate.ID, testclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr(updatedName),
			})
			assert.NilError(t, err)
			assert.Assert(t, updateResp != nil)
			assert.Check(t, is.Equal(updatedName, updateResp.UpdateInternalPolicy.InternalPolicy.Name))

			deleteResp, err := suite.Client.API.DeleteInternalPolicy(localTestOrg.Member.UserCtx, policyToDelete.ID)
			assert.NilError(t, err)
			assert.Assert(t, deleteResp != nil)
			assert.Check(t, is.Equal(policyToDelete.ID, deleteResp.DeleteInternalPolicy.DeletedID))
		})
	}
}

func TestMutationUpdateBulkInternalPolicy(t *testing.T) {
	// create internal policies to be updated
	policy1 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	policy2 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	policy3 := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	control := (&th.ControlBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	subcontrol := (&th.SubcontrolBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	task := (&th.TaskBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// create another user and add them to the same organization and group as th.SharedTestUser1
	// this will allow us to test the group editor permissions
	anotherAdminUser := suite.UserBuilder(context.Background(), t)
	suite.AddUserToOrganization(th.SharedTestUser1.UserCtx, t, &anotherAdminUser, enums.RoleAdmin, th.SharedTestUser1.OrganizationID)

	groupMember := (&th.GroupMemberBuilder{Client: suite.Client, UserID: anotherAdminUser.ID}).MustNew(th.SharedTestUser1.UserCtx, t)

	policyAnotherUser := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser2.UserCtx, t)

	// ensure the user does not currently have access to update the policy
	res, err := suite.Client.API.UpdateBulkInternalPolicy(th.SharedTestUser2.UserCtx, []string{policy1.ID}, testclient.UpdateInternalPolicyInput{
		Status: lo.ToPtr(enums.DocumentPublished),
	})

	assert.Assert(t, is.Nil(err))
	// make sure nothing was updated
	assert.Equal(t, len(res.UpdateBulkInternalPolicy.InternalPolicies), 0)

	testCases := []struct {
		name                 string
		ids                  []string
		input                testclient.UpdateInternalPolicyInput
		client               *testclient.TestClient
		ctx                  context.Context
		expectedErr          string
		expectedUpdatedCount int
	}{
		{
			name: "happy path, update status on multiple policies",
			ids:  []string{policy1.ID, policy2.ID, policy3.ID},
			input: testclient.UpdateInternalPolicyInput{
				Status: &enums.DocumentPublished,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedUpdatedCount: 3,
		},
		{
			name: "happy path, editor permissions and revision bump",
			ids:  []string{policy1.ID, policy2.ID},
			input: testclient.UpdateInternalPolicyInput{
				AddEditorIDs: []string{groupMember.GroupID},
				RevisionBump: &models.Major,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedUpdatedCount: 2,
		},
		{
			name:        "empty ids array",
			ids:         []string{},
			input:       testclient.UpdateInternalPolicyInput{Details: lo.ToPtr("test")},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "ids is required",
		},
		{
			name: "mixed success and failure - some policies not authorized",
			ids:  []string{policy1.ID, policyAnotherUser.ID}, // second policy should fail authorization
			input: testclient.UpdateInternalPolicyInput{
				Status: &enums.DocumentDraft,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedUpdatedCount: 1, // only policy1 should be updated
		},
		{
			name: "update not allowed, no permissions to policies",
			ids:  []string{policy1.ID},
			input: testclient.UpdateInternalPolicyInput{
				Status: &enums.DocumentPublished,
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser2.UserCtx,
			expectedUpdatedCount: 0, // should not find any policies to update
		},
		{
			name: "update multiple policies with controls and tasks",
			ids:  []string{policy1.ID, policy2.ID, policy3.ID},
			input: testclient.UpdateInternalPolicyInput{
				Details:          lo.ToPtr("Updated details for all policies"),
				AddControlIDs:    []string{control.ID},
				AddSubcontrolIDs: []string{subcontrol.ID},
				AddTaskIDs:       []string{task.ID},
			},
			client:               suite.Client.API,
			ctx:                  th.SharedTestUser1.UserCtx,
			expectedUpdatedCount: 3,
		},
	}

	for _, tc := range testCases {
		t.Run("Bulk Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateBulkInternalPolicy(tc.ctx, tc.ids, tc.input)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.UpdateBulkInternalPolicy.InternalPolicies, tc.expectedUpdatedCount))
			assert.Check(t, is.Len(resp.UpdateBulkInternalPolicy.UpdatedIDs, tc.expectedUpdatedCount))

			// verify all returned policies have the expected values from tc.input
			for _, policy := range resp.UpdateBulkInternalPolicy.InternalPolicies {
				if tc.input.Name != nil {
					assert.Check(t, is.Equal(*tc.input.Name, policy.Name))
				}

				if tc.input.Status != nil {
					assert.Check(t, is.Equal(*tc.input.Status, *policy.Status))
				}

				if tc.input.Tags != nil {
					assert.Check(t, is.DeepEqual(tc.input.Tags, policy.Tags))
				}

				if tc.input.RevisionBump == &models.Minor {
					assert.Check(t, is.Equal("v0.1.0", *policy.Revision))
				}

				if tc.input.RevisionBump == &models.Major {
					assert.Check(t, is.Equal("v1.0.0", *policy.Revision))
				}

				if len(tc.input.AddEditorIDs) > 0 {
					// ensure the user has access to the policy now
					res, err := suite.Client.API.UpdateInternalPolicy(anotherAdminUser.UserCtx, policy.ID, testclient.UpdateInternalPolicyInput{
						Tags: []string{"bulk-test-tag"},
					})
					assert.NilError(t, err)
					assert.Check(t, res != nil)
					assert.Check(t, is.Equal(policy.ID, res.UpdateInternalPolicy.InternalPolicy.ID))
				}

				// ensure the org owner has access to the policy that was updated
				checkResp, err := suite.Client.API.GetInternalPolicyByID(th.SharedTestUser1.UserCtx, policy.ID)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(policy.ID, checkResp.InternalPolicy.ID))
			}

			// verify that the returned IDs match the ones that were actually updated
			for _, updatedID := range resp.UpdateBulkInternalPolicy.UpdatedIDs {
				found := false
				for _, expectedID := range tc.ids {
					if expectedID == updatedID {
						found = true
						break
					}
				}
				assert.Check(t, found, "Updated ID %s should be in the original request", updatedID)
			}
		})
	}

	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, IDs: []string{policy1.ID, policy2.ID, policy3.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: policyAnotherUser.ID}).MustDelete(th.SharedTestUser2.UserCtx, t)
	(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, ID: control.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.SubcontrolDeleteOne]{Client: suite.Client.DB.Subcontrol, ID: subcontrol.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.TaskDeleteOne]{Client: suite.Client.DB.Task, ID: task.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	(&th.Cleanup[*generated.GroupDeleteOne]{Client: suite.Client.DB.Group, ID: groupMember.GroupID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryInternalPolicyDiscussionComments(t *testing.T) {
	internalPolicy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	ownerComment := "comment from the policy owner"
	memberComment := "comment from the view only member"
	adminComment := "comment from the org admin"

	// the owner starts the discussion with the first comment
	resp, err := suite.Client.API.UpdateInternalPolicy(th.SharedTestUser1.UserCtx, internalPolicy.ID, testclient.UpdateInternalPolicyInput{
		AddDiscussion: &testclient.CreateDiscussionInput{
			AddComment: &testclient.CreateNoteInput{
				Text: ownerComment,
			},
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, is.Len(resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges, 1))

	discussionID := resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node.ID

	// a view only member and an org admin both reply on the same discussion
	for _, tc := range []struct {
		ctx     context.Context
		comment string
	}{
		{ctx: th.SharedViewOnlyUser.UserCtx, comment: memberComment},
		{ctx: th.SharedAdminUser.UserCtx, comment: adminComment},
	} {
		_, err = suite.Client.API.UpdateInternalPolicy(tc.ctx, internalPolicy.ID, testclient.UpdateInternalPolicyInput{
			UpdateDiscussion: &testclient.UpdateDiscussionsInput{
				ID: discussionID,
				Input: &testclient.UpdateDiscussionInput{
					AddComment: &testclient.CreateNoteInput{
						Text: tc.comment,
					},
				},
			},
		})
		assert.NilError(t, err)
	}

	// the view only member can see the policy, so they should see every comment on the
	// discussion, not just the ones they authored
	policyResp, err := suite.Client.API.GetInternalPolicyByID(th.SharedViewOnlyUser.UserCtx, internalPolicy.ID)
	assert.NilError(t, err)
	assert.Assert(t, is.Len(policyResp.InternalPolicy.Discussions.Edges, 1))

	comments := lo.Map(policyResp.InternalPolicy.Discussions.Edges[0].Node.Comments.Edges,
		func(edge *testclient.GetInternalPolicyByID_InternalPolicy_Discussions_Edges_Node_Comments_Edges, _ int) string {
			return edge.Node.Text
		})

	assert.Assert(t, is.Len(comments, 3))
	assert.Check(t, is.Contains(comments, ownerComment))
	assert.Check(t, is.Contains(comments, memberComment))
	assert.Check(t, is.Contains(comments, adminComment))

	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: internalPolicy.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationResolveInternalPolicyDiscussion(t *testing.T) {
	internalPolicy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// the org admin starts a discussion on the policy
	resp, err := suite.Client.API.UpdateInternalPolicy(th.SharedAdminUser.UserCtx, internalPolicy.ID, testclient.UpdateInternalPolicyInput{
		AddDiscussion: &testclient.CreateDiscussionInput{
			AddComment: &testclient.CreateNoteInput{
				Text: "this thread should be resolvable by its author",
			},
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, is.Len(resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges, 1))

	discussionID := resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node.ID

	// org admins can edit the policy, so they should be able to resolve a discussion on it
	resolved, err := suite.Client.API.UpdateDiscussion(th.SharedAdminUser.UserCtx, discussionID, testclient.UpdateDiscussionInput{
		IsResolved: lo.ToPtr(true),
	})
	assert.NilError(t, err)
	assert.Check(t, resolved.UpdateDiscussion.Discussion.IsResolved)

	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: internalPolicy.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationResolveOwnInternalPolicyDiscussion(t *testing.T) {
	internalPolicy := (&th.InternalPolicyBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// a view only member starts a discussion on a policy they cannot edit
	resp, err := suite.Client.API.UpdateInternalPolicy(th.SharedViewOnlyUser.UserCtx, internalPolicy.ID, testclient.UpdateInternalPolicyInput{
		AddDiscussion: &testclient.CreateDiscussionInput{
			AddComment: &testclient.CreateNoteInput{
				Text: "a member should be able to resolve their own thread",
			},
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, is.Len(resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges, 1))

	discussionID := resp.UpdateInternalPolicy.InternalPolicy.Discussions.Edges[0].Node.ID

	// the creator owns the discussion, so they can resolve it without edit access to the policy
	resolved, err := suite.Client.API.UpdateDiscussion(th.SharedViewOnlyUser.UserCtx, discussionID, testclient.UpdateDiscussionInput{
		IsResolved: lo.ToPtr(true),
	})
	assert.NilError(t, err)
	assert.Check(t, resolved.UpdateDiscussion.Discussion.IsResolved)

	// another user with only view access cannot resolve someone else's thread
	_, err = suite.Client.API.UpdateDiscussion(th.SharedAuditorUser.UserCtx, discussionID, testclient.UpdateDiscussionInput{
		IsResolved: lo.ToPtr(false),
	})
	assert.ErrorContains(t, err, th.NotAuthorizedErrorMsg)

	(&th.Cleanup[*generated.InternalPolicyDeleteOne]{Client: suite.Client.DB.InternalPolicy, ID: internalPolicy.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}
