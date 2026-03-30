// package graphapi
package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestQueryEmailTemplate(t *testing.T) {
	// create an emailTemplate to be queried using testUser1
	emailTemplate := (&EmailTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the EmailTemplate
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: emailTemplate.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: emailTemplate.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: emailTemplate.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "email template not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "email template not found, using not authorized user from another org",
			queryID:  emailTemplate.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetEmailTemplateByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.EmailTemplate.ID))
			assert.Check(t, is.Equal(emailTemplate.Name, resp.EmailTemplate.Name))
		})
	}

	(&Cleanup[*generated.EmailTemplateDeleteOne]{client: suite.client.db.EmailTemplate, ID: emailTemplate.ID}).MustDelete(testUser1.UserCtx, t)
}

// func TestQueryEmailTemplates(t *testing.T) {
// 	// create multiple EmailTemplates to be queried using testUser1
// 	EmailTemplate1 := (&EmailTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
// 	EmailTemplate2 := (&EmailTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

// 	testCases := []struct {
// 		name            string
// 		client          *testclient.TestClient
// 		ctx             context.Context
// 		expectedResults int
// 	}{
// 		{
// 			name:            "happy path",
// 			client:          suite.client.api,
// 			ctx:             testUser1.UserCtx,
// 			expectedResults: 2,
// 		},
// 		{
// 			name:            "happy path, using read only user of the same org",
// 			client:          suite.client.api,
// 			ctx:             viewOnlyUser.UserCtx,
// 			expectedResults: 2,
// 		},
// 		{
// 			name:            "happy path, using api token",
// 			client:          suite.client.apiWithToken,
// 			ctx:             context.Background(),
// 			expectedResults: 2,
// 		},
// 		{
// 			name:            "happy path, using pat",
// 			client:          suite.client.apiWithPAT,
// 			ctx:             context.Background(),
// 			expectedResults: 2,
// 		},
// 		{
// 			name:            "another user, no EmailTemplates should be returned",
// 			client:          suite.client.api,
// 			ctx:             testUser2.UserCtx,
// 			expectedResults: 0,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run("List "+tc.name, func(t *testing.T) {
// 			resp, err := tc.client.GetAllEmailTemplates(tc.ctx)
// 			assert.NilError(t, err)
// 			assert.Assert(t, resp != nil)

// 			assert.Check(t, is.Len(resp.EmailTemplates.Edges, tc.expectedResults))
// 		})
// 	}

// 	(&Cleanup[*generated.EmailTemplateDeleteOne]{client: suite.client.db.EmailTemplate, IDs: []string{EmailTemplate1.ID, EmailTemplate2.ID}}).MustDelete(testUser1.UserCtx, t)
// }

// func TestMutationCreateEmailTemplate(t *testing.T) {
// 	testCases := []struct {
// 		name        string
// 		request     testclient.CreateEmailTemplateInput
// 		client      *testclient.TestClient
// 		ctx         context.Context
// 		expectedErr string
// 	}{
// 		{
// 			name:    "happy path, minimal input",
// 			request: testclient.CreateEmailTemplateInput{
// 				// add minimal input for the EmailTemplate
// 			},
// 			client: suite.client.api,
// 			ctx:    testUser1.UserCtx,
// 		},
// 		{
// 			name:    "happy path, all input",
// 			request: testclient.CreateEmailTemplateInput{
// 				// add all input for the EmailTemplate
// 			},
// 			client: suite.client.api,
// 			ctx:    testUser1.UserCtx,
// 		},
// 		{
// 			name:    "happy path, using pat",
// 			request: testclient.CreateEmailTemplateInput{
// 				// add input for the EmailTemplate
// 			},
// 			client: suite.client.apiWithPAT,
// 			ctx:    context.Background(),
// 		},
// 		{
// 			name:    "happy path, using api token",
// 			request: testclient.CreateEmailTemplateInput{
// 				// add input for the EmailTemplate
// 			},
// 			client: suite.client.apiWithToken,
// 			ctx:    context.Background(),
// 		},
// 		{
// 			name:    "user not authorized, not enough permissions",
// 			request: testclient.CreateEmailTemplateInput{
// 				// add all input for the EmailTemplate
// 			},
// 			client:      suite.client.api,
// 			ctx:         viewOnlyUser.UserCtx,
// 			expectedErr: notAuthorizedErrorMsg,
// 		},
// 		// add additional test cases for the EmailTemplate
// 		//   - add test cases for required fields not being provided
// 		//   - add test cases for invalid input
// 		{
// 			name:        "missing required field",
// 			request:     testclient.CreateEmailTemplateInput{},
// 			client:      suite.client.api,
// 			ctx:         testUser1.UserCtx,
// 			expectedErr: "value is less than the required length",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run("Create "+tc.name, func(t *testing.T) {
// 			resp, err := tc.client.CreateEmailTemplate(tc.ctx, tc.request)
// 			if tc.expectedErr != "" {
// 				assert.ErrorContains(t, err, tc.expectedErr)

// 				return
// 			}

// 			assert.NilError(t, err)
// 			assert.Assert(t, resp != nil)

// 			// check required fields

// 			// check optional fields with if checks if they were provided or not

// 			// cleanup each EmailTemplate created
// 			(&Cleanup[*generated.EmailTemplateDeleteOne]{client: suite.client.db.EmailTemplate, ID: resp.CreateEmailTemplate.EmailTemplate.ID}).MustDelete(testUser1.UserCtx, t)
// 		})
// 	}
// }

// func TestMutationUpdateEmailTemplate(t *testing.T) {
// 	EmailTemplate := (&EmailTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

// 	testCases := []struct {
// 		name        string
// 		request     testclient.UpdateEmailTemplateInput
// 		client      *testclient.TestClient
// 		ctx         context.Context
// 		expectedErr string
// 	}{
// 		{
// 			name:    "happy path, update field",
// 			request: testclient.UpdateEmailTemplateInput{
// 				// add field to update
// 			},
// 			client: suite.client.api,
// 			ctx:    testUser1.UserCtx,
// 		},
// 		{
// 			name:    "happy path, update multiple fields",
// 			request: testclient.UpdateEmailTemplateInput{
// 				// add fields to update
// 			},
// 			client: suite.client.apiWithPAT,
// 			ctx:    context.Background(),
// 		},
// 		// add additional test update cases for the EmailTemplate
// 		{
// 			name:    "update not allowed, not enough permissions",
// 			request: testclient.UpdateEmailTemplateInput{
// 				// add field to update
// 			},
// 			client:      suite.client.api,
// 			ctx:         viewOnlyUser.UserCtx,
// 			expectedErr: notAuthorizedErrorMsg,
// 		},
// 		{
// 			name:    "update not allowed, no permissions",
// 			request: testclient.UpdateEmailTemplateInput{
// 				// add field to update
// 			},
// 			client:      suite.client.api,
// 			ctx:         testUser2.UserCtx,
// 			expectedErr: notFoundErrorMsg,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run("Update "+tc.name, func(t *testing.T) {
// 			resp, err := tc.client.UpdateEmailTemplate(tc.ctx, EmailTemplate.ID, tc.request)
// 			if tc.expectedErr != "" {
// 				assert.ErrorContains(t, err, tc.expectedErr)

// 				return
// 			}

// 			assert.NilError(t, err)
// 			assert.Assert(t, resp != nil)
// 			// add checks for the updated fields if they were set in the request
// 		})
// 	}

// 	(&Cleanup[*generated.EmailTemplateDeleteOne]{client: suite.client.db.EmailTemplate, ID: EmailTemplate.ID}).MustDelete(testUser1.UserCtx, t)
// }

// func TestMutationDeleteEmailTemplate(t *testing.T) {
// 	// create EmailTemplates to be deleted
// 	EmailTemplate1 := (&EmailTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
// 	EmailTemplate2 := (&EmailTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
// 	EmailTemplate3 := (&EmailTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

// 	testCases := []struct {
// 		name        string
// 		idToDelete  string
// 		client      *testclient.TestClient
// 		ctx         context.Context
// 		expectedErr string
// 	}{
// 		{
// 			name:        "not found, delete",
// 			idToDelete:  EmailTemplate1.ID,
// 			client:      suite.client.api,
// 			ctx:         testUser2.UserCtx,
// 			expectedErr: notFoundErrorMsg,
// 		},
// 		{
// 			name:        "not authorized, delete",
// 			idToDelete:  EmailTemplate1.ID,
// 			client:      suite.client.api,
// 			ctx:         viewOnlyUser.UserCtx,
// 			expectedErr: notAuthorizedErrorMsg,
// 		},
// 		{
// 			name:       "happy path, delete",
// 			idToDelete: EmailTemplate1.ID,
// 			client:     suite.client.api,
// 			ctx:        testUser1.UserCtx,
// 		},
// 		{
// 			name:        "already deleted, not found",
// 			idToDelete:  EmailTemplate1.ID,
// 			client:      suite.client.api,
// 			ctx:         testUser1.UserCtx,
// 			expectedErr: "not found",
// 		},
// 		{
// 			name:       "happy path, delete using personal access token",
// 			idToDelete: EmailTemplate2.ID,
// 			client:     suite.client.apiWithPAT,
// 			ctx:        context.Background(),
// 		},
// 		{
// 			name:       "happy path, delete using api token",
// 			idToDelete: EmailTemplate3.ID,
// 			client:     suite.client.apiWithToken,
// 			ctx:        context.Background(),
// 		},
// 		{
// 			name:        "unknown id, not found",
// 			idToDelete:  ulids.New().String(),
// 			client:      suite.client.api,
// 			ctx:         testUser1.UserCtx,
// 			expectedErr: notFoundErrorMsg,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run("Delete "+tc.name, func(t *testing.T) {
// 			resp, err := tc.client.DeleteEmailTemplate(tc.ctx, tc.idToDelete)
// 			if tc.expectedErr != "" {
// 				assert.ErrorContains(t, err, tc.expectedErr)

// 				return
// 			}

// 			assert.NilError(t, err)
// 			assert.Assert(t, resp != nil)
// 			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteEmailTemplate.DeletedID))
// 		})
// 	}
// }
