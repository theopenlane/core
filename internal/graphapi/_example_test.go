package graphapi_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/openlaneclient"
)

// This file provides examples for testing the basic CRUD operations for the Openlane API for a given object
// you can start by replacing `OBJECT` with the object you are testing, e.g. `OBJECT`

// For org owned objects, the setup should be pretty similar to the examples below
// For objects with complex relationships, you may need to create additional objects to test the relationships

// Seed data for test users is created in seed_test.go, refer to the file for details
// the seeded data includes:
// - testUser1 (org owner), viewOnlyUser (read-only access), and adminUser (view and edit access)
//   as part of one organization
// - testUser2 part of another organization
// if you find yourself needing additional test users, you can add them to the seed data so they
// are available for all tests

// Additionally, there are clients that can be used based on JWT, personal access token, or API tokens
// All tests should include the cross section of all clients and users to ensure proper access control

// Before getting started, you'll need to add an OBJECTBuilder to the models_test.go file
// and add the necessary fields to the OBJECTBuilder struct
// Example:
// type OBJECTBuilder struct {
// 	client *client

// 	// Fields
// 	Name string
// }

// // MustNew OBJECT builder is used to create, without authz checks, OBJECTs in the database
// func (e *OBJECTBuilder) MustNew(ctx context.Context, t *testing.T) *ent.OBJECT {
// 	ctx = privacy.DecisionContext(ctx, privacy.Allow)

// 	if e.Name == "" {
// 		e.Name = gofakeit.AppName()
// 	}

// 	OBJECT := e.client.db.OBJECT.Create().
// 		SetName(e.Name).
// 		SaveX(ctx)

// 	return OBJECT
// }

func TestQueryOBJECT(t *testing.T) {

	// create an OBJECT to be queried using testUser1
	OBJECT := (&OBJECTBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the OBJECT
	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: OBJECT.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: OBJECT.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: OBJECT.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "OBJECT not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "OBJECT not found, using not authorized user",
			queryID:  OBJECT.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetOBJECTByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			require.NotEmpty(t, resp.OBJECT)

			assert.Equal(t, tc.queryID, resp.OBJECT.ID)
			// add additional assertions for the object
			// e.g.
			// assert.NotEmpty(t, resp.OBJECT.Name)
		})
	}
}

func TestQueryOBJECTs(t *testing.T) {

	// create multiple objects to be queried using testUser1
	(&OBJECTBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	(&OBJECTBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             testUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.client.api,
			ctx:             viewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.client.apiWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.client.apiWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no OBJECTs should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllOBJECTs(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.OBJECTs.Edges, tc.expectedResults)
		})
	}
}

func TestMutationCreateOBJECT(t *testing.T) {

	testCases := []struct {
		name        string
		request     openlaneclient.CreateOBJECTInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path, minimal input",
			request: openlaneclient.CreateOBJECTInput{
				// add minimal input for the OBJECT
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:    "happy path, all input",
			request: openlaneclient.CreateOBJECTInput{
				// add all input for the OBJECT
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:    "happy path, using pat",
			request: openlaneclient.CreateOBJECTInput{
				// add input for the OBJECT
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name:    "happy path, using api token",
			request: openlaneclient.CreateOBJECTInput{
				// add input for the OBJECT
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name:    "user not authorized, not enough permissions",
			request: openlaneclient.CreateOBJECTInput{
				// add all input for the OBJECT
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		// add additional test cases for the OBJECT
		//   - add test cases for required fields not being provided
		//   - add test cases for invalid input
		{
			name:        "missing required field",
			request:     openlaneclient.CreateOBJECTInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateOBJECT(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields

			// check optional fields with if checks if they were provided or not
		})
	}
}

func TestMutationUpdateOBJECT(t *testing.T) {

	OBJECT := (&OBJECTBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateOBJECTInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:    "happy path, update field",
			request: openlaneclient.UpdateOBJECTInput{
				// add field to update
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:    "happy path, update multiple fields",
			request: openlaneclient.UpdateOBJECTInput{
				// add fields to update
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		// add additional test update cases for the OBJECT
		{
			name:    "update not allowed, not enough permissions",
			request: openlaneclient.UpdateOBJECTInput{
				// add field to update
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:    "update not allowed, no permissions",
			request: openlaneclient.UpdateOBJECTInput{
				// add field to update
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateOBJECT(tc.ctx, OBJECT.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// add checks for the updated fields if they were set in the request
		})
	}
}

func TestMutationDeleteOBJECT(t *testing.T) {

	// create objects to be deleted
	OBJECT1 := (&OBJECTBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	OBJECT2 := (&OBJECTBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not authorized, delete",
			idToDelete:  OBJECT1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: OBJECT1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  OBJECT1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: OBJECT2.ID,
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
			resp, err := tc.client.DeleteOBJECT(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteOBJECT.DeletedID)
		})
	}
}
