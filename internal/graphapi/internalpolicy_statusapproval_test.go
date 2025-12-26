package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/core/common/enums"
)

func TestCreateInternalPolicyStatusApproval(t *testing.T) {
	// Create approver and delegate groups
	approverGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	emptyGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Add testUser1 to both approver and delegate groups (to test both paths)
	(&GroupMemberBuilder{client: suite.client, UserID: testUser1.ID, GroupID: approverGroup.ID}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: testUser1.ID, GroupID: delegateGroup.ID}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		userContext     context.Context
		approverID      *string
		delegateID      *string
		status          enums.DocumentStatus
		requireApproval bool
		expectError     bool
		errorContains   string
	}{
		{
			name:            "happy path: create with APPROVED status and user in approver group",
			client:          suite.client.api,
			userContext:     testUser1.UserCtx,
			approverID:      &approverGroup.ID,
			status:          enums.DocumentApproved,
			requireApproval: true,
			expectError:     false,
		},
		{
			name:            "happy path: create with APPROVED status and user in delegate group",
			client:          suite.client.api,
			userContext:     testUser1.UserCtx,
			delegateID:      &delegateGroup.ID,
			status:          enums.DocumentApproved,
			requireApproval: true,
			expectError:     false,
		},
		{
			name:            "fail: create with APPROVED status but user not in approver group",
			client:          suite.client.api,
			userContext:     testUser1.UserCtx,
			approverID:      &emptyGroup.ID, // testUser1 is NOT in emptyGroup
			status:          enums.DocumentApproved,
			requireApproval: true,
			expectError:     true,
			errorContains:   "you must be in the approver group to mark as approved",
		},
		{
			name:            "happy path: create with APPROVED status but user not in approver group but approval not required",
			client:          suite.client.api,
			userContext:     testUser1.UserCtx,
			approverID:      &emptyGroup.ID, // testUser1 is NOT in emptyGroup
			status:          enums.DocumentApproved,
			requireApproval: false,
			expectError:     false,
		},
		{
			name:            "fail: create with APPROVED status but no approver group set",
			client:          suite.client.api,
			userContext:     testUser1.UserCtx,
			status:          enums.DocumentApproved,
			requireApproval: true,
			expectError:     true,
			errorContains:   "you must be in the approver group to mark as approved",
		},
		{
			name:            "happy path: create with DRAFT status and no approver group",
			client:          suite.client.api,
			userContext:     testUser1.UserCtx,
			status:          enums.DocumentDraft,
			requireApproval: true,
			expectError:     false,
		},
		{
			name:            "happy path: create with PUBLISHED status and no approver group",
			client:          suite.client.api,
			userContext:     testUser1.UserCtx,
			status:          enums.DocumentPublished,
			requireApproval: true,
			expectError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := testclient.CreateInternalPolicyInput{
				Name:             gofakeit.AppName(),
				Status:           &tc.status,
				ApprovalRequired: &tc.requireApproval,
			}

			if tc.approverID != nil {
				input.ApproverID = tc.approverID
			}

			if tc.delegateID != nil {
				input.DelegateID = tc.delegateID
			}

			resp, err := tc.client.CreateInternalPolicy(tc.userContext, input)

			if tc.expectError {
				assert.ErrorContains(t, err, tc.errorContains)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, *resp.CreateInternalPolicy.InternalPolicy.Status == tc.status)

			// Cleanup
			(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, ID: resp.CreateInternalPolicy.InternalPolicy.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// Cleanup
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{approverGroup.ID, delegateGroup.ID, emptyGroup.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestUpdateInternalPolicyStatusApproval(t *testing.T) {
	// Create approver and delegate groups
	approverGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	delegateGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	emptyGroup := (&GroupBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Add testUser1 to both approver and delegate groups
	(&GroupMemberBuilder{client: suite.client, UserID: testUser1.ID, GroupID: approverGroup.ID}).MustNew(testUser1.UserCtx, t)
	(&GroupMemberBuilder{client: suite.client, UserID: testUser1.ID, GroupID: delegateGroup.ID}).MustNew(testUser1.UserCtx, t)

	// Create policies with DRAFT status - use MustNew which bypasses hooks
	policy1 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	policy2 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	policy3 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	policy4 := (&InternalPolicyBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	policy5 := (&InternalPolicyBuilder{client: suite.client, SkipApprovalRequirement: true}).MustNew(testUser1.UserCtx, t)

	// Set approver/delegate groups using direct database access (bypasses authorization but keeps user context)
	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	suite.client.db.InternalPolicy.UpdateOneID(policy1.ID).SetApproverID(approverGroup.ID).SetStatus(enums.DocumentDraft).SaveX(allowCtx)
	suite.client.db.InternalPolicy.UpdateOneID(policy2.ID).SetDelegateID(delegateGroup.ID).SetStatus(enums.DocumentDraft).SaveX(allowCtx)
	suite.client.db.InternalPolicy.UpdateOneID(policy3.ID).SetStatus(enums.DocumentDraft).SaveX(allowCtx)                              // no approver group
	suite.client.db.InternalPolicy.UpdateOneID(policy4.ID).SetApproverID(emptyGroup.ID).SetStatus(enums.DocumentDraft).SaveX(allowCtx) // user not in group
	suite.client.db.InternalPolicy.UpdateOneID(policy5.ID).SetApproverID(emptyGroup.ID).SetStatus(enums.DocumentDraft).SaveX(allowCtx) // approval not required so should pass

	testCases := []struct {
		name          string
		client        *testclient.TestClient
		policyID      string
		userContext   context.Context
		newStatus     *enums.DocumentStatus
		expectError   bool
		errorContains string
	}{
		{
			name:        "happy path: update to APPROVED status with user in approver group",
			client:      suite.client.api,
			policyID:    policy1.ID,
			userContext: testUser1.UserCtx,
			newStatus:   lo.ToPtr(enums.DocumentApproved),
			expectError: false,
		},
		{
			name:        "happy path: update to APPROVED status with user in delegate group",
			client:      suite.client.api,
			policyID:    policy2.ID,
			userContext: testUser1.UserCtx,
			newStatus:   lo.ToPtr(enums.DocumentApproved),
			expectError: false,
		},
		{
			name:          "fail: update to APPROVED status but user not in approver group",
			client:        suite.client.api,
			policyID:      policy4.ID, // policy4 has emptyGroup, testUser1 is not a member
			userContext:   testUser1.UserCtx,
			newStatus:     lo.ToPtr(enums.DocumentApproved),
			expectError:   true,
			errorContains: "you must be in the approver group to mark as approved",
		},
		{
			name:          "fail: update to APPROVED status but no approver group set",
			client:        suite.client.api,
			policyID:      policy3.ID,
			userContext:   testUser1.UserCtx,
			newStatus:     lo.ToPtr(enums.DocumentApproved),
			expectError:   true,
			errorContains: "you must be in the approver group to mark as approved",
		},
		{
			name:        "happy path: update other status without check",
			client:      suite.client.api,
			policyID:    policy3.ID,
			userContext: testUser1.UserCtx,
			newStatus:   lo.ToPtr(enums.DocumentPublished),
			expectError: false,
		},
		{
			name:        "happy path: update to APPROVED status but approval not required",
			client:      suite.client.api,
			policyID:    policy5.ID,
			userContext: testUser1.UserCtx,
			newStatus:   lo.ToPtr(enums.DocumentApproved),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := testclient.UpdateInternalPolicyInput{}

			if tc.newStatus != nil {
				input.Status = tc.newStatus
			}

			resp, err := tc.client.UpdateInternalPolicy(tc.userContext, tc.policyID, input)

			if tc.expectError {
				assert.ErrorContains(t, err, tc.errorContains)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			if tc.newStatus != nil {
				assert.Check(t, *resp.UpdateInternalPolicy.InternalPolicy.Status == *tc.newStatus)
			}
		})
	}

	// Cleanup
	(&Cleanup[*generated.InternalPolicyDeleteOne]{client: suite.client.db.InternalPolicy, IDs: []string{policy1.ID, policy2.ID, policy3.ID, policy4.ID, policy5.ID}}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.GroupDeleteOne]{client: suite.client.db.Group, IDs: []string{approverGroup.ID, delegateGroup.ID, emptyGroup.ID}}).MustDelete(testUser1.UserCtx, t)
}
