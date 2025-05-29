package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func TestQueryInternalPolicy(t *testing.T) {
	// create an InternalPolicy to be queried using testUser1
	internalPolicy := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// setup a blocked group with a view only user
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID, GroupID: blockedGroup.ID}).MustNew(testUser1.UserCtx, t)

	internalPolicy2 := (&InternalPolicyBuilder{client: suite.client, BlockedGroupIDs: []string{blockedGroup.ID}}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the internal policy
	testCases := []struct {
		name               string
		queryID            string
		client             *openlaneclient.OpenlaneClient
		ctx                context.Context
		errorMsg           string
		updateBlockedGroup bool
	}{
		{
			name:    "happy path",
			queryID: internalPolicy.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: internalPolicy.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:               "happy path, read only user but blocked",
			queryID:            internalPolicy2.ID,
			client:             suite.client.api,
			ctx:                viewOnlyUser.UserCtx,
			errorMsg:           notFoundErrorMsg, // should not be able to access the policy due to blocked group
			updateBlockedGroup: true,
		},
		{
			name:    "happy path, read only user no longer blocked",
			queryID: internalPolicy2.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: internalPolicy.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "internalPolicy not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "internal policy not found, using not authorized user",
			queryID:  internalPolicy.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetInternalPolicyByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				if tc.updateBlockedGroup {
					_, err := suite.client.api.UpdateInternalPolicy(testUser1.UserCtx, internalPolicy2.ID,
						openlaneclient.UpdateInternalPolicyInput{
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
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, IDs: []string{internalPolicy.ID, internalPolicy2.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryInternalPolicies(t *testing.T) {
	// create multiple policies to be queried using testUser1
	ip1 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	ip2 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// setup a blocked group with a view only user
	blockedGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID, GroupID: blockedGroup.ID}).MustNew(testUser1.UserCtx, t)

	ip3 := (&InternalPolicyBuilder{client: suite.client, BlockedGroupIDs: []string{blockedGroup.ID}}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name               string
		client             *openlaneclient.OpenlaneClient
		ctx                context.Context
		updateBlockedGroup bool
		expectedResults    int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 3,
		},
		{
			name:               "happy path, using read only user of the same org, one policy blocked",
			client:             suite.client.api,
			ctx:                viewOnlyUser.UserCtx,
			expectedResults:    2,    // should not see the policy that is blocked for them
			updateBlockedGroup: true, // update the blocked group to allow the view only user to see the policy
		},
		{
			name:            "happy path, using read only user of the same org, no blocked group",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 3, // should now see all policies after removing the blocked group
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 3,
		},
		{
			name:            "another user, no policies should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllInternalPolicies(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.InternalPolicies.Edges, tc.expectedResults))

			if tc.updateBlockedGroup {
				// do it the opposite, remove the policy from the group
				_, err := suite.client.api.UpdateGroup(testUser1.UserCtx, blockedGroup.ID,
					openlaneclient.UpdateGroupInput{
						RemoveInternalPolicyBlockedGroupIDs: []string{ip3.ID},
					},
				)

				assert.NilError(t, err)
			}
		})
	}

	// delete created policies
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, IDs: []string{ip1.ID, ip2.ID, ip3.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateInternalPolicy(t *testing.T) {
	// create a system owned standard with a control
	systemStandard := (&StandardBuilder{client: suite.client}).MustNew(systemAdminUser.UserCtx, t)
	// create a control and add it to the system standard
	systemControl := (&ControlBuilder{client: suite.client, StandardID: systemStandard.ID}).MustNew(systemAdminUser.UserCtx, t)

	anotherGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// group for the view only user
	groupMember := (&GroupMemberBuilder{client: suite.client, UserID: viewOnlyUser.ID}).MustNew(testUser1.UserCtx, t)

	// approver and delegator groups for the test user
	approverGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// edges to add
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subcontrol := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name                       string
		request                    openlaneclient.CreateInternalPolicyInput
		addGroupToOrg              bool
		controlEdgeShouldBeCreated bool
		client                     *openlaneclient.OpenlaneClient
		ctx                        context.Context
		expectedErr                string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input except edges",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:       "Releasing a new version",
				Status:     &enums.DocumentDraft,
				PolicyType: lo.ToPtr("sop"),
				Revision:   lo.ToPtr("v1.1.0"),
				Details:    lo.ToPtr("do stuff"),
				ApproverID: &approverGroup.ID,
				DelegateID: &delegateGroup.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, long details",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:       "Releasing a new version",
				Status:     &enums.DocumentDraft,
				PolicyType: lo.ToPtr("sop"),
				Revision:   lo.ToPtr("v1.1.0"),
				Details:    lo.ToPtr(gofakeit.Sentence(1000)),
				ApproverID: &approverGroup.ID,
				DelegateID: &delegateGroup.ID,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, with control edges",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:          "Releasing a new version",
				Status:        &enums.DocumentDraft,
				PolicyType:    lo.ToPtr("sop"),
				Revision:      lo.ToPtr("v1.1.0"),
				Details:       lo.ToPtr("do stuff"),
				ControlIDs:    []string{control.ID},
				SubcontrolIDs: []string{subcontrol.ID},
				TaskIDs:       []string{task.ID},
			},
			client:                     suite.client.api,
			ctx:                        testUser1.UserCtx,
			controlEdgeShouldBeCreated: true,
		},
		{
			name: "should not be allowed to add system standard control",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:       "Releasing a new version",
				Status:     &enums.DocumentDraft,
				ControlIDs: []string{systemControl.ID},
			},
			client:                     suite.client.api,
			ctx:                        testUser1.UserCtx,
			controlEdgeShouldBeCreated: false, // user does not have edit access to the control, it is owned by the system
		},
		{
			name: "happy path, add editor group",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:      "Test Policy",
				EditorIDs: []string{testUser1.GroupID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, add same task to another policy",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:    "Test Policy",
				TaskIDs: []string{task.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, add same control to another policy",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:       "Test Policy",
				ControlIDs: []string{control.ID},
			},
			client:                     suite.client.api,
			ctx:                        testUser1.UserCtx,
			controlEdgeShouldBeCreated: true,
		},
		{
			name: "happy path, add same sub control to another policy",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:          "Test Policy",
				SubcontrolIDs: []string{subcontrol.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "add editor group, again - ensures the same group can be added to multiple policies",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:            "Test Policy",
				EditorIDs:       []string{testUser1.GroupID},
				BlockedGroupIDs: []string{anotherGroup.ID},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateInternalPolicyInput{
				Name:    "Test Internal Policy",
				OwnerID: &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: openlaneclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "user now authorized, added to group with creator permissions",
			request: openlaneclient.CreateInternalPolicyInput{
				Name: "Test InternalPolicy",
			},
			addGroupToOrg: true,
			client:        suite.client.api,
			ctx:           viewOnlyUser.UserCtx,
		},
		{
			name: "missing required field",
			request: openlaneclient.CreateInternalPolicyInput{
				Details: lo.ToPtr("instructions on how to release a new version"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			if tc.addGroupToOrg {
				_, err := suite.client.api.UpdateOrganization(testUser1.UserCtx, testUser1.OrganizationID,
					openlaneclient.UpdateOrganizationInput{
						AddInternalPolicyCreatorIDs: []string{groupMember.GroupID},
					}, nil)
				assert.NilError(t, err)
			}

			log.Error().Str("control_id", systemControl.ID).Msg("system control id")
			resp, err := tc.client.CreateInternalPolicy(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

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

			if tc.request.PolicyType != nil {
				assert.Check(t, is.Equal(*tc.request.PolicyType, *resp.CreateInternalPolicy.InternalPolicy.PolicyType))
			} else {
				assert.Check(t, is.Equal(*resp.CreateInternalPolicy.InternalPolicy.PolicyType, ""))
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

					assert.Check(t, is.Equal(controlFound, tc.controlEdgeShouldBeCreated), "control not found in edges")
				}
			}

			// cleanup
			(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, IDs: []string{resp.CreateInternalPolicy.InternalPolicy.ID}}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// cleanup
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control.ID, subcontrol.ControlID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: []string{subcontrol.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: []string{task.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{anotherGroup.ID, groupMember.GroupID, approverGroup.ID, delegateGroup.ID}}).MustDelete(testUser1.UserCtx, t)

	// cleanup the system standard and control
	(&Cleanup[*generated.StandardDeleteOne]{client: suite.client.db.Standard, IDs: []string{systemStandard.ID}}).MustDelete(systemAdminUser.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{systemControl.ID}}).MustDelete(systemAdminUser.UserCtx, t)
}

func TestMutationUpdateInternalPolicy(t *testing.T) {
	internalPolicy := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	internalPolicyAdminUser := (&InternalPolicyBuilder{client: suite.client}).MustNew(adminUser.UserCtx, t)

	// create another admin user and add them to the same organization and group as testUser1
	// this will allow us to test the group editor permissions
	anotherAdminUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &anotherAdminUser, enums.RoleAdmin, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherAdminUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create a viewer user and add them to the same organization as testUser1
	// also add them to the same group as testUser1, this should still allow them to edit the policy
	// despite not not being an organization admin
	anotherViewerUser := suite.userBuilder(context.Background(), t)
	suite.addUserToOrganization(testUser1.UserCtx, t, &anotherViewerUser, enums.RoleMember, testUser1.OrganizationID)

	(&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: testUser1.GroupID}).MustNew(testUser1.UserCtx, t)

	// create one more group that will be used to test the blocked group permissions and add anotherViewerUser to it
	blockGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: anotherViewerUser.ID, GroupID: blockGroup.ID}).MustNew(testUser1.UserCtx, t)

	// edges to add
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	subcontrol := (&SubcontrolBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	task := (&TaskBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		policyID    string
		request     openlaneclient.UpdateInternalPolicyInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:     "happy path, update details field",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Details: lo.ToPtr(gofakeit.Sentence(200)),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:     "happy path, update details field on policy created by another user",
			policyID: internalPolicyAdminUser.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Details: lo.ToPtr(gofakeit.Sentence(200)),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx, // org owner should always be able to update the policy
		},
		{
			name:     "happy path, update name field",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Name:         lo.ToPtr("Updated InternalPolicy Name"),
				AddEditorIDs: []string{testUser1.GroupID}, // add the group to the editor groups for subsequent tests
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:     "happy path, update multiple fields",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Status:           &enums.DocumentPublished,
				Details:          lo.ToPtr("Updated details"),
				RevisionBump:     &models.Major,
				AddControlIDs:    []string{control.ID},
				AddSubcontrolIDs: []string{subcontrol.ID},
				AddTaskIDs:       []string{task.ID},
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:     "update not allowed, not enough permissions",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated InternalPolicy Name"),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:     "update allowed, user in editor group",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client: suite.client.api,
			ctx:    anotherAdminUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name:     "member update allowed, user in editor group",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client: suite.client.api,
			ctx:    anotherViewerUser.UserCtx, // user assigned to the group which has editor permissions
		},
		{
			name:     "happy path, block the group from editing",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				AddBlockedGroupIDs: []string{blockGroup.ID}, // block the group
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:     "member update no longer allowed, user in blocked group",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again"),
			},
			client:      suite.client.api,
			ctx:         anotherViewerUser.UserCtx, // user assigned to the group which was blocked
			expectedErr: notFoundErrorMsg,
		},
		{
			name:     "happy path, remove the group",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				RemoveEditorIDs: []string{testUser1.GroupID}, // remove the group from the editor groups
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:     "update not allowed, editor group was removed",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Name: lo.ToPtr("Updated Procedure Name Again Again"),
			},
			client:      suite.client.api,
			ctx:         anotherAdminUser.UserCtx, // user assigned to the group which no longer has editor permissions
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:     "update not allowed, no permissions",
			policyID: internalPolicy.ID,
			request: openlaneclient.UpdateInternalPolicyInput{
				Details: lo.ToPtr("Updated details"),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateInternalPolicy(tc.ctx, tc.policyID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

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

			if tc.request.PolicyType != nil {
				assert.Check(t, is.Equal(*tc.request.PolicyType, *resp.UpdateInternalPolicy.InternalPolicy.PolicyType))
			}

			if tc.request.Revision != nil {
				assert.Check(t, is.Equal(*tc.request.Revision, *resp.UpdateInternalPolicy.InternalPolicy.Revision))
			}

			if tc.request.RevisionBump == &models.Major {
				assert.Check(t, is.Equal("v1.0.0", *resp.UpdateInternalPolicy.InternalPolicy.Revision))
			}

			if tc.request.Details != nil {
				assert.Check(t, is.DeepEqual(tc.request.Details, resp.UpdateInternalPolicy.InternalPolicy.Details))
			}
		})
	}

	// cleanup
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, IDs: []string{internalPolicy.ID, internalPolicyAdminUser.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, IDs: []string{control.ID, subcontrol.ControlID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.SubcontrolDeleteOne]{client: suite.client.db.Subcontrol, IDs: []string{subcontrol.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TaskDeleteOne]{client: suite.client.db.Task, IDs: []string{task.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{blockGroup.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteInternalPolicy(t *testing.T) {
	// create internal policies to be deleted
	internalPolicy1 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	internalPolicy2 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  internalPolicy1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: internalPolicy1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  internalPolicy1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: internalPolicy2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown id, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteInternalPolicy(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteInternalPolicy.DeletedID))
		})
	}
}
