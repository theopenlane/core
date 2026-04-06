// package graphapi
package graphapi_test

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/testclient"
	"github.com/theopenlane/utils/ulids"
)

func TestQueryEmailBranding(t *testing.T) {
	// create an email branding to be queried using testUser1
	emailBranding := (&EmailBrandingBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	// add test cases for querying the email branding
	testCases := []struct {
		name     string
		queryID  string
		client   *testclient.TestClient
		ctx      context.Context
		errorMsg string
	}{
		{
			name:    "happy path",
			queryID: emailBranding.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: emailBranding.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: emailBranding.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "email branding not found, invalid ID",
			queryID:  "invalid",
			client:   suite.client.api,
			ctx:      testUser1.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
		{
			name:     "email branding not found, using not authorized user from another org",
			queryID:  emailBranding.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetEmailBrandingByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.queryID, resp.EmailBranding.ID))
			assert.Check(t, is.Equal(emailBranding.Name, resp.EmailBranding.Name))
		})
	}

	(&Cleanup[*generated.EmailBrandingDeleteOne]{client: suite.client.db.EmailBranding, ID: emailBranding.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryEmailBrandings(t *testing.T) {
	// create multiple email branding to be queried using testUser1
	emailBranding1 := (&EmailBrandingBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	emailBranding2 := (&EmailBrandingBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
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
			name:            "another user, no email branding should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEmailBrandings(tc.ctx, nil, nil, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.EmailBrandings.Edges, tc.expectedResults))
		})
	}

	(&Cleanup[*generated.EmailBrandingDeleteOne]{client: suite.client.db.EmailBranding, IDs: []string{emailBranding1.ID, emailBranding2.ID}}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateEmailBranding(t *testing.T) {
	// add email template to link to email branding
	emailTemplate := (&EmailTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	testCases := []struct {
		name        string
		request     testclient.CreateEmailBrandingInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateEmailBrandingInput{
				Name: "Email Branding Name " + ulids.New().String(),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateEmailBrandingInput{
				Name:             "Email Branding Name " + ulids.New().String(),
				PrimaryColor:     lo.ToPtr("#068721"),
				BrandName:        lo.ToPtr("Brand Name"),
				LogoRemoteURL:    lo.ToPtr("https://example.com/logo.png"),
				SecondaryColor:   lo.ToPtr("#FFFFFF"),
				BackgroundColor:  lo.ToPtr("#F0F0F0"),
				TextColor:        lo.ToPtr("#000000"),
				ButtonColor:      lo.ToPtr("#9AC01D"),
				ButtonTextColor:  lo.ToPtr("#000000"),
				LinkColor:        lo.ToPtr("#777777"),
				FontFamily:       &enums.FontCourierBold,
				EmailTemplateIDs: []string{emailTemplate.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateEmailBrandingInput{
				Name: "Email Branding Name " + ulids.New().String(),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateEmailBrandingInput{
				Name:             "Email Branding Name " + ulids.New().String(),
				EmailTemplateIDs: []string{emailTemplate.ID},
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateEmailBrandingInput{
				Name: "Email Branding Name " + ulids.New().String(),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "missing required field, name",
			request:     testclient.CreateEmailBrandingInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateEmailBranding(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.Name, resp.CreateEmailBranding.EmailBranding.Name))

			if tc.request.PrimaryColor != nil {
				assert.Check(t, is.Equal(*tc.request.PrimaryColor, *resp.CreateEmailBranding.EmailBranding.PrimaryColor))
			} else {
				assert.Check(t, is.Equal("", *resp.CreateEmailBranding.EmailBranding.PrimaryColor))
			}

			if tc.request.BrandName != nil {
				assert.Check(t, is.Equal(*tc.request.BrandName, *resp.CreateEmailBranding.EmailBranding.BrandName))
			} else {
				assert.Check(t, is.Equal("", *resp.CreateEmailBranding.EmailBranding.BrandName))
			}

			if tc.request.LogoRemoteURL != nil {
				assert.Check(t, is.Equal(*tc.request.LogoRemoteURL, *resp.CreateEmailBranding.EmailBranding.LogoRemoteURL))
			} else {
				assert.Check(t, resp.CreateEmailBranding.EmailBranding.LogoRemoteURL == nil)
			}

			if tc.request.SecondaryColor != nil {
				assert.Check(t, is.Equal(*tc.request.SecondaryColor, *resp.CreateEmailBranding.EmailBranding.SecondaryColor))
			} else {
				assert.Check(t, is.Equal("", *resp.CreateEmailBranding.EmailBranding.SecondaryColor))
			}

			if tc.request.BackgroundColor != nil {
				assert.Check(t, is.Equal(*tc.request.BackgroundColor, *resp.CreateEmailBranding.EmailBranding.BackgroundColor))
			} else {
				assert.Check(t, is.Equal("", *resp.CreateEmailBranding.EmailBranding.BackgroundColor))
			}

			if tc.request.TextColor != nil {
				assert.Check(t, is.Equal(*tc.request.TextColor, *resp.CreateEmailBranding.EmailBranding.TextColor))
			} else {
				assert.Check(t, is.Equal("", *resp.CreateEmailBranding.EmailBranding.TextColor))
			}

			if tc.request.ButtonColor != nil {
				assert.Check(t, is.Equal(*tc.request.ButtonColor, *resp.CreateEmailBranding.EmailBranding.ButtonColor))
			} else {
				assert.Check(t, is.Equal("", *resp.CreateEmailBranding.EmailBranding.ButtonColor))
			}

			if tc.request.ButtonTextColor != nil {
				assert.Check(t, is.Equal(*tc.request.ButtonTextColor, *resp.CreateEmailBranding.EmailBranding.ButtonTextColor))
			} else {
				assert.Check(t, is.Equal("", *resp.CreateEmailBranding.EmailBranding.ButtonTextColor))
			}

			if tc.request.LinkColor != nil {
				assert.Check(t, is.Equal(*tc.request.LinkColor, *resp.CreateEmailBranding.EmailBranding.LinkColor))
			} else {
				assert.Check(t, is.Equal("", *resp.CreateEmailBranding.EmailBranding.LinkColor))
			}

			if tc.request.FontFamily != nil {
				assert.Check(t, is.Equal(*tc.request.FontFamily, *resp.CreateEmailBranding.EmailBranding.FontFamily))
			} else {
				assert.Check(t, is.Equal(enums.FontHelvetica, *resp.CreateEmailBranding.EmailBranding.FontFamily))
			}

			// cleanup each email branding created
			(&Cleanup[*generated.EmailBrandingDeleteOne]{client: suite.client.db.EmailBranding, ID: resp.CreateEmailBranding.EmailBranding.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}

	// cleanup email template created
	(&Cleanup[*generated.EmailTemplateDeleteOne]{client: suite.client.db.EmailTemplate, ID: emailTemplate.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationUpdateEmailBranding(t *testing.T) {
	emailBranding := (&EmailBrandingBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	emailTemplate := (&EmailTemplateBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateEmailBrandingInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateEmailBrandingInput{
				Name:                lo.ToPtr("Updated Email Branding Name " + ulids.New().String()),
				AddEmailTemplateIDs: []string{emailTemplate.ID},
			},
			client: suite.client.api,
			ctx:    adminUser.UserCtx,
		},
		{
			name: "happy path, update multiple fields with PAT",
			request: testclient.UpdateEmailBrandingInput{
				Name:            lo.ToPtr("Updated Email Branding Name " + ulids.New().String()),
				BackgroundColor: lo.ToPtr("#F0F0F0"),
				TextColor:       lo.ToPtr("#000000"),
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions",
			request: testclient.UpdateEmailBrandingInput{
				Name: lo.ToPtr("Updated Email Branding Name " + ulids.New().String()),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateEmailBrandingInput{
				Name: lo.ToPtr("Updated Email Branding Name " + ulids.New().String()),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateEmailBranding(tc.ctx, emailBranding.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Name != nil {
				assert.Check(t, is.Equal(*tc.request.Name, resp.UpdateEmailBranding.EmailBranding.Name))
			}

			if tc.request.BackgroundColor != nil {
				assert.Check(t, is.Equal(*tc.request.BackgroundColor, *resp.UpdateEmailBranding.EmailBranding.BackgroundColor))
			}

			if tc.request.TextColor != nil {
				assert.Check(t, is.Equal(*tc.request.TextColor, *resp.UpdateEmailBranding.EmailBranding.TextColor))
			}
		})
	}

	(&Cleanup[*generated.EmailBrandingDeleteOne]{client: suite.client.db.EmailBranding, ID: emailBranding.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.EmailTemplateDeleteOne]{client: suite.client.db.EmailTemplate, ID: emailTemplate.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteEmailBranding(t *testing.T) {
	// create email branding to be deleted
	emailBranding1 := (&EmailBrandingBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	emailBranding2 := (&EmailBrandingBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	emailBranding3 := (&EmailBrandingBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not found, delete",
			idToDelete:  emailBranding1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  emailBranding1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: emailBranding1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  emailBranding1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: emailBranding2.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: emailBranding3.ID,
			client:     suite.client.apiWithToken,
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
			resp, err := tc.client.DeleteEmailBranding(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteEmailBranding.DeletedID))
		})
	}
}
