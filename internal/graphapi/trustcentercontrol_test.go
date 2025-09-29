package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

func TestMutationCreateTrustCenterControl(t *testing.T) {
	// Create test data
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Create control and trust center for different organization
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.CreateTrustCenterControlInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path - create trust center control",
			request: testclient.CreateTrustCenterControlInput{
				TrustCenterID: &trustCenter.ID,
				ControlID:     control.ID,
				Tags:          []string{"test", "fga"},
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path - different organization",
			request: testclient.CreateTrustCenterControlInput{
				TrustCenterID: &trustCenter2.ID,
				ControlID:     control2.ID,
			},
			client: suite.client.api,
			ctx:    testUser2.UserCtx,
		},
		{
			name: "not authorized - view only user",
			request: testclient.CreateTrustCenterControlInput{
				TrustCenterID: &trustCenter.ID,
				ControlID:     control.ID,
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "not authorized - different organization user",
			request: testclient.CreateTrustCenterControlInput{
				TrustCenterID: &trustCenter.ID,
				ControlID:     control.ID,
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "control not found",
			request: testclient.CreateTrustCenterControlInput{
				TrustCenterID: &trustCenter.ID,
				ControlID:     "non-existent-control-id",
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg, // FGA returns unauthorized for non-existent resources
		},
		{
			name: "trust center not found",
			request: testclient.CreateTrustCenterControlInput{
				TrustCenterID: lo.ToPtr("non-existent-trust-center-id"),
				ControlID:     control.ID,
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg, // FGA returns unauthorized for non-existent resources
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateTrustCenterControl(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			if tc.request.TrustCenterID != nil && resp.CreateTrustCenterControl.TrustCenterControl.TrustCenterID != nil {
				assert.Check(t, is.Equal(*tc.request.TrustCenterID, *resp.CreateTrustCenterControl.TrustCenterControl.TrustCenterID))
			}
			assert.Check(t, is.Equal(tc.request.ControlID, resp.CreateTrustCenterControl.TrustCenterControl.ControlID))

			if tc.request.Tags != nil {
				assert.Check(t, is.DeepEqual(tc.request.Tags, resp.CreateTrustCenterControl.TrustCenterControl.Tags))
			}

			// Clean up
			(&Cleanup[*generated.TrustCenterControlDeleteOne]{
				client: suite.client.db.TrustCenterControl,
				ID:     resp.CreateTrustCenterControl.TrustCenterControl.ID,
			}).MustDelete(tc.ctx, t)
		})
	}

	// Clean up test data
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control2.ID}).MustDelete(testUser2.UserCtx, t)
}

func TestMutationDeleteTrustCenterControl(t *testing.T) {
	// Create test data
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenterControl := (&TrustCenterControlBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
		ControlID:     control.ID,
	}).MustNew(testUser1.UserCtx, t)

	// Create trust center control for different organization
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	trustCenterControl2 := (&TrustCenterControlBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter2.ID,
		ControlID:     control2.ID,
	}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:       "happy path - delete trust center control",
			idToDelete: trustCenterControl.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "not authorized - view only user",
			idToDelete:  trustCenterControl2.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not authorized - different organization user",
			idToDelete:  trustCenterControl2.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "trust center control not found",
			idToDelete:  "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteTrustCenterControl(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteTrustCenterControl.DeletedID))
		})
	}

	// Clean up remaining test data
	(&Cleanup[*generated.TrustCenterControlDeleteOne]{client: suite.client.db.TrustCenterControl, ID: trustCenterControl2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control2.ID}).MustDelete(testUser2.UserCtx, t)
}

// TestTrustCenterControlFGATupleCreation tests that the FGA tuples are created correctly
func TestTrustCenterControlFGATupleCreation(t *testing.T) {
	// Create test data
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// Test 1: Create trust center control association
	resp, err := suite.client.api.CreateTrustCenterControl(testUser1.UserCtx, testclient.CreateTrustCenterControlInput{
		TrustCenterID: &trustCenter.ID,
		ControlID:     control.ID,
		Tags:          []string{"fga", "test"},
	})
	assert.NilError(t, err)
	assert.Assert(t, resp != nil)

	trustCenterControlID := resp.CreateTrustCenterControl.TrustCenterControl.ID

	// Test 2: Verify the FGA tuple was created by checking access
	t.Run("verify FGA tuple creation", func(t *testing.T) {
		// Create anonymous user context to test public access
		anonymousCtx := createAnonymousTrustCenterContext(trustCenter.ID, testUser1.OrganizationID)
		anonUser, ok := auth.AnonymousTrustCenterUserFromContext(anonymousCtx)
		assert.Check(t, ok, "should have anonymous user in context")

		// Check if the anonymous user can view the control via FGA
		// This verifies that the trust_center_association tuple was created correctly
		checkReq := fgax.AccessCheck{
			SubjectID:   anonUser.SubjectID,
			SubjectType: "user",
			ObjectID:    control.ID,
			ObjectType:  "control",
			Relation:    "can_view",
		}

		allowed, err := suite.client.db.Authz.CheckAccess(anonymousCtx, checkReq)
		assert.NilError(t, err)
		assert.Check(t, allowed, "anonymous user should be able to view control via trust center association")
	})

	// Test 3: Test FGA check directly
	t.Run("verify FGA check for public access", func(t *testing.T) {
		// Create anonymous user context
		anonymousCtx := createAnonymousTrustCenterContext(trustCenter.ID, testUser1.OrganizationID)
		anonUser, ok := auth.AnonymousTrustCenterUserFromContext(anonymousCtx)
		assert.Check(t, ok, "should have anonymous user in context")

		// Check if the anonymous user can view the control via FGA
		checkReq := fgax.AccessCheck{
			SubjectID:   anonUser.SubjectID,
			SubjectType: "user",
			ObjectID:    control.ID,
			ObjectType:  "control",
			Relation:    "can_view",
		}

		allowed, err := suite.client.db.Authz.CheckAccess(anonymousCtx, checkReq)
		assert.NilError(t, err)
		assert.Check(t, allowed, "anonymous user should be able to view control via FGA")
	})

	// Clean up
	(&Cleanup[*generated.TrustCenterControlDeleteOne]{client: suite.client.db.TrustCenterControl, ID: trustCenterControlID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryTrustCenterControl(t *testing.T) {
	// Create test data
	trustCenter := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	control := (&ControlBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	trustCenterControl := (&TrustCenterControlBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter.ID,
		ControlID:     control.ID,
		Tags:          []string{"query", "test"},
	}).MustNew(testUser1.UserCtx, t)

	// Create trust center control for different organization
	trustCenter2 := (&TrustCenterBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	control2 := (&ControlBuilder{client: suite.client}).MustNew(testUser2.UserCtx, t)
	trustCenterControl2 := (&TrustCenterControlBuilder{
		client:        suite.client,
		TrustCenterID: trustCenter2.ID,
		ControlID:     control2.ID,
	}).MustNew(testUser2.UserCtx, t)

	testCases := []struct {
		name        string
		queryID     string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path - query trust center control",
			queryID: trustCenterControl.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path - admin user",
			queryID: trustCenterControl.ID,
			client:  suite.client.api,
			ctx:     adminUser.UserCtx,
		},
		{
			name:    "happy path - view only user can view trust center controls",
			queryID: trustCenterControl.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:        "not authorized - different organization user",
			queryID:     trustCenterControl.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "trust center control not found",
			queryID:     "non-existent-id",
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetTrustCenterControlByID(tc.ctx, tc.queryID)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.queryID, resp.TrustCenterControl.ID))

			// Check TrustCenterID (it's a pointer to string)
			if resp.TrustCenterControl.TrustCenterID != nil {
				assert.Check(t, is.Equal(trustCenter.ID, *resp.TrustCenterControl.TrustCenterID))
			}
			assert.Check(t, is.Equal(control.ID, resp.TrustCenterControl.ControlID))
		})
	}

	// Clean up
	(&Cleanup[*generated.TrustCenterControlDeleteOne]{client: suite.client.db.TrustCenterControl, ID: trustCenterControl.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterControlDeleteOne]{client: suite.client.db.TrustCenterControl, ID: trustCenterControl2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.TrustCenterDeleteOne]{client: suite.client.db.TrustCenter, ID: trustCenter2.ID}).MustDelete(testUser2.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ControlDeleteOne]{client: suite.client.db.Control, ID: control2.ID}).MustDelete(testUser2.UserCtx, t)
}
