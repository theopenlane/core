// package graphapi
package graphapi_test

import (
	"context"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/utils/ulids"
)

// TestPreviewEmailTemplateGate verifies the preview resolver admits customer-selectable catalog
// entries and the system trust center update message while rejecting internal system emails
func TestPreviewEmailTemplateGate(t *testing.T) {
	tcOrg := th.CreateFreshOrgWithTrustCenter(t)

	// the system trust center update message is previewable so editors can see what subscribers receive
	resp, err := suite.Client.API.PreviewEmailTemplate(tcOrg.Owner.UserCtx, emaildef.TrustCenterUpdateTemplate, map[string]any{})
	assert.NilError(t, err)
	assert.Check(t, resp.PreviewEmailTemplate != "")

	// customer-selectable catalog entries remain previewable
	resp, err = suite.Client.API.PreviewEmailTemplate(tcOrg.Owner.UserCtx, emaildef.BrandedMessageOp.Name(), map[string]any{})
	assert.NilError(t, err)
	assert.Check(t, resp.PreviewEmailTemplate != "")

	// internal system emails are not previewable
	_, err = suite.Client.API.PreviewEmailTemplate(tcOrg.Owner.UserCtx, emaildef.SubprocessorNotificationOp.Name(), map[string]any{})
	assert.ErrorContains(t, err, "not a customer-selectable")
}

func validEmailTemplateDefaults() map[string]any {
	return map[string]any{
		"subject": "Test subject",
		"title":   "Test title",
		"intros":  []any{"Test body"},
	}
}

func TestQueryEmailTemplate(t *testing.T) {
	// create an email template to be queried using th.SharedTestUser1
	emailTemplate := (&th.EmailTemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	// add test cases for querying the email template
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
			client:  suite.Client.API,
			ctx:     th.SharedTestUser1.UserCtx,
		},
		{
			name:    "happy path, read only user",
			queryID: emailTemplate.ID,
			client:  suite.Client.API,
			ctx:     th.SharedViewOnlyUser.UserCtx,
		},
		{
			name:    "happy path using personal access token",
			queryID: emailTemplate.ID,
			client:  suite.Client.APIWithPAT,
			ctx:     context.Background(),
		},
		{
			name:     "email template not found, invalid ID",
			queryID:  "invalid",
			client:   suite.Client.API,
			ctx:      th.SharedTestUser1.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
		},
		{
			name:     "email template not found, using not authorized user from another org",
			queryID:  emailTemplate.ID,
			client:   suite.Client.API,
			ctx:      th.SharedTestUser2.UserCtx,
			errorMsg: th.NotFoundErrorMsg,
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

	(&th.Cleanup[*generated.EmailTemplateDeleteOne]{Client: suite.Client.DB.EmailTemplate, ID: emailTemplate.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestQueryEmailTemplates(t *testing.T) {
	// create multiple email templates to be queried using th.SharedTestUser1
	emailTemplate1 := (&th.EmailTemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	emailTemplate2 := (&th.EmailTemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name            string
		client          *testclient.TestClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser1.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using read only user of the same org",
			client:          suite.Client.API,
			ctx:             th.SharedViewOnlyUser.UserCtx,
			expectedResults: 2,
		},
		{
			name:            "happy path, using api token",
			client:          suite.Client.APIWithToken,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "happy path, using pat",
			client:          suite.Client.APIWithPAT,
			ctx:             context.Background(),
			expectedResults: 2,
		},
		{
			name:            "another user, no email templates should be returned",
			client:          suite.Client.API,
			ctx:             th.SharedTestUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllEmailTemplates(tc.ctx, nil, nil, nil, nil, nil)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.EmailTemplates.Edges, tc.expectedResults))
		})
	}

	(&th.Cleanup[*generated.EmailTemplateDeleteOne]{Client: suite.Client.DB.EmailTemplate, IDs: []string{emailTemplate1.ID, emailTemplate2.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationCreateEmailTemplate(t *testing.T) {
	testCases := []struct {
		name        string
		request     testclient.CreateEmailTemplateInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateEmailTemplateInput{
				Key:             emaildef.BrandedMessageOp.Name(),
				Name:            "Email Template Name " + ulids.New().String(),
				TemplateContext: &enums.TemplateContextCampaignRecipient,
				Defaults:        validEmailTemplateDefaults(),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: testclient.CreateEmailTemplateInput{
				Key:             emaildef.BrandedMessageOp.Name(),
				Name:            "Email Template Name " + ulids.New().String(),
				TemplateContext: &enums.TemplateContextTransactional,
				Description:     lo.ToPtr("This is a description for the email template"),
				Active:          lo.ToPtr(false),
				Version:         lo.ToPtr(int64(1)),
				Defaults: map[string]any{
					"subject":    "{{ .companyName }} — Welcome",
					"title":      "Welcome to {{ .companyName }}",
					"intros":     []any{"Hi {{ .firstName }}, thanks for joining."},
					"buttonText": "Get Started",
					"buttonLink": "{{ .rootURL }}/dashboard",
				},
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, using pat",
			request: testclient.CreateEmailTemplateInput{
				Key:             emaildef.BrandedMessageOp.Name(),
				Name:            "Email Template Name " + ulids.New().String(),
				TemplateContext: &enums.TemplateContextTransactional,
				Defaults:        validEmailTemplateDefaults(),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using api token",
			request: testclient.CreateEmailTemplateInput{
				Key:             emaildef.BrandedMessageOp.Name(),
				Name:            "Email Template Name " + ulids.New().String(),
				TemplateContext: &enums.TemplateContextCampaignRecipient,
				Defaults:        validEmailTemplateDefaults(),
			},
			client: suite.Client.APIWithToken,
			ctx:    context.Background(),
		},
		{
			name: "user not authorized, not enough permissions",
			request: testclient.CreateEmailTemplateInput{
				Key:             emaildef.BrandedMessageOp.Name(),
				Name:            "Email Template Name " + ulids.New().String(),
				TemplateContext: &enums.TemplateContextCampaignRecipient,
				Defaults:        validEmailTemplateDefaults(),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "missing required field, key",
			request: testclient.CreateEmailTemplateInput{
				Name:            "Email Template Name " + ulids.New().String(),
				TemplateContext: &enums.TemplateContextCampaignRecipient,
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "missing required field, name",
			request: testclient.CreateEmailTemplateInput{
				Key:             emaildef.BrandedMessageOp.Name(),
				TemplateContext: &enums.TemplateContextCampaignRecipient,
				Defaults:        validEmailTemplateDefaults(),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateEmailTemplate(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Equal(tc.request.Key, resp.CreateEmailTemplate.EmailTemplate.Key))
			assert.Check(t, is.Equal(tc.request.Name, resp.CreateEmailTemplate.EmailTemplate.Name))
			assert.Check(t, is.DeepEqual(tc.request.TemplateContext, resp.CreateEmailTemplate.EmailTemplate.TemplateContext))

			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.CreateEmailTemplate.EmailTemplate.Description))
			} else {
				assert.Check(t, is.Equal(*resp.CreateEmailTemplate.EmailTemplate.Description, ""))
			}

			if tc.request.Active != nil {
				assert.Check(t, is.Equal(*tc.request.Active, resp.CreateEmailTemplate.EmailTemplate.Active))
			} else {
				assert.Check(t, resp.CreateEmailTemplate.EmailTemplate.Active == true)
			}

			if tc.request.Version != nil {
				assert.Check(t, is.Equal(*tc.request.Version, resp.CreateEmailTemplate.EmailTemplate.Version))
			} else {
				assert.Check(t, resp.CreateEmailTemplate.EmailTemplate.Version == 1)
			}

			// cleanup each email template created
			(&th.Cleanup[*generated.EmailTemplateDeleteOne]{Client: suite.Client.DB.EmailTemplate, ID: resp.CreateEmailTemplate.EmailTemplate.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateEmailTemplate(t *testing.T) {
	emailTemplate := (&th.EmailTemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     testclient.UpdateEmailTemplateInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update field",
			request: testclient.UpdateEmailTemplateInput{
				Name: lo.ToPtr("Updated Email Template Name " + ulids.New().String()),
			},
			client: suite.Client.API,
			ctx:    th.SharedTestUser1.UserCtx,
		},
		{
			name: "happy path, update multiple fields",
			request: testclient.UpdateEmailTemplateInput{
				Name:        lo.ToPtr("Updated Email Template Name " + ulids.New().String()),
				Description: lo.ToPtr("Updated description for the email template"),
				Active:      lo.ToPtr(false),
			},
			client: suite.Client.APIWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update not allowed, not enough permissions",
			request: testclient.UpdateEmailTemplateInput{
				Name: lo.ToPtr("Updated Email Template Name " + ulids.New().String()),
			},
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name: "update not allowed, no permissions",
			request: testclient.UpdateEmailTemplateInput{
				Name: lo.ToPtr("Updated Email Template Name " + ulids.New().String()),
			},
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateEmailTemplate(tc.ctx, emailTemplate.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.Name != nil {
				assert.Check(t, is.Equal(*tc.request.Name, resp.UpdateEmailTemplate.EmailTemplate.Name))
			}
			if tc.request.Description != nil {
				assert.Check(t, is.Equal(*tc.request.Description, *resp.UpdateEmailTemplate.EmailTemplate.Description))
			}
			if tc.request.Active != nil {
				assert.Check(t, is.Equal(*tc.request.Active, resp.UpdateEmailTemplate.EmailTemplate.Active))
			}
		})
	}

	(&th.Cleanup[*generated.EmailTemplateDeleteOne]{Client: suite.Client.DB.EmailTemplate, ID: emailTemplate.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
}

func TestMutationDeleteEmailTemplate(t *testing.T) {
	// create email templates to be deleted
	emailTemplate1 := (&th.EmailTemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	emailTemplate2 := (&th.EmailTemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)
	emailTemplate3 := (&th.EmailTemplateBuilder{Client: suite.Client}).MustNew(th.SharedTestUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not found, delete",
			idToDelete:  emailTemplate1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser2.UserCtx,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "not authorized, delete",
			idToDelete:  emailTemplate1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedViewOnlyUser.UserCtx,
			expectedErr: th.NotAuthorizedErrorMsg,
		},
		{
			name:       "happy path, delete",
			idToDelete: emailTemplate1.ID,
			client:     suite.Client.API,
			ctx:        th.SharedTestUser1.UserCtx,
		},
		{
			name:        "already deleted, not found",
			idToDelete:  emailTemplate1.ID,
			client:      suite.Client.API,
			ctx:         th.SharedTestUser1.UserCtx,
			expectedErr: "not found",
		},
		{
			name:       "happy path, delete using personal access token",
			idToDelete: emailTemplate2.ID,
			client:     suite.Client.APIWithPAT,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete using api token",
			idToDelete: emailTemplate3.ID,
			client:     suite.Client.APIWithToken,
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
			resp, err := tc.client.DeleteEmailTemplate(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Check(t, is.Equal(tc.idToDelete, resp.DeleteEmailTemplate.DeletedID))
		})
	}
}
