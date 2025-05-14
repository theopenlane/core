package graphapi_test

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func TestQueryContact(t *testing.T) {
	contact := (&ContactBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		expected *ent.Contact
		errorMsg string
	}{
		{
			name:    "happy path contact",
			queryID: contact.ID,
			client:  suite.client.api,
			ctx:     testUser1.UserCtx,
		},
		{
			name:    "happy path contact, view only user",
			queryID: contact.ID,
			client:  suite.client.api,
			ctx:     viewOnlyUser.UserCtx,
		},
		{
			name:     "contact not returned, no access",
			queryID:  contact.ID,
			client:   suite.client.api,
			ctx:      testUser2.UserCtx,
			errorMsg: "contact not found",
		},
		{
			name:    "happy path contact, with api token",
			queryID: contact.ID,
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
		},
		{
			name:    "happy path contact, with pat",
			queryID: contact.ID,
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetContactByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
		})
	}

	(&Cleanup[*generated.ContactDeleteOne]{client: suite.client.db.Contact, ID: contact.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestQueryContacts(t *testing.T) {
	contact1 := (&ContactBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	contact2 := (&ContactBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

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
			name:            "happy path, view only user",
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
			name:            "another user, no contacts should be returned",
			client:          suite.client.api,
			ctx:             testUser2.UserCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			resp, err := tc.client.GetAllContacts(tc.ctx)
			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Check(t, is.Len(resp.Contacts.Edges, tc.expectedResults))
		})
	}

	(&Cleanup[*generated.ContactDeleteOne]{client: suite.client.db.Contact, ID: contact1.ID}).MustDelete(testUser1.UserCtx, t)
	(&Cleanup[*generated.ContactDeleteOne]{client: suite.client.db.Contact, ID: contact2.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationCreateContact(t *testing.T) {
	testCases := []struct {
		name        string
		request     openlaneclient.CreateContactInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateContactInput{
				FullName: "Aemond Targaryen",
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "view only user cannot create",
			request: openlaneclient.CreateContactInput{
				FullName: "Aemond Targaryen",
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateContactInput{
				FullName: "Rhaenys Targaryen",
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateContactInput{
				FullName: "Aegon Targaryen",
				OwnerID:  &testUser1.OrganizationID,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateContactInput{
				FullName:    "Aemond Targaryen",
				Email:       lo.ToPtr("Atargarygen@dragon.com"),
				PhoneNumber: lo.ToPtr(gofakeit.Phone()),
				Title:       lo.ToPtr("Prince of the Targaryen Dynasty"),
				Company:     lo.ToPtr("Targaryen Dynasty"),
				Status:      &enums.UserStatusOnboarding,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "missing required field, name",
			request: openlaneclient.CreateContactInput{
				Email: lo.ToPtr("atargarygen@dragon.com"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateContact(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			assert.Equal(t, tc.request.FullName, resp.CreateContact.Contact.FullName)

			if tc.request.Email == nil {
				assert.Equal(t, *resp.CreateContact.Contact.Email, "")
			} else {
				assert.Equal(t, strings.ToLower(*tc.request.Email), *resp.CreateContact.Contact.Email)
			}

			if tc.request.PhoneNumber == nil {
				assert.Equal(t, *resp.CreateContact.Contact.PhoneNumber, "")
			} else {
				assert.Equal(t, *tc.request.PhoneNumber, *resp.CreateContact.Contact.PhoneNumber)
			}

			if tc.request.Address == nil {
				assert.Equal(t, *resp.CreateContact.Contact.Address, "")
			} else {
				assert.Equal(t, *tc.request.Address, *resp.CreateContact.Contact.Address)
			}

			if tc.request.Title == nil {
				assert.Equal(t, *resp.CreateContact.Contact.Title, "")
			} else {
				assert.Equal(t, *tc.request.Title, *resp.CreateContact.Contact.Title)
			}

			if tc.request.Company == nil {
				assert.Equal(t, *resp.CreateContact.Contact.Company, "")
			} else {
				assert.Equal(t, *tc.request.Company, *resp.CreateContact.Contact.Company)
			}

			// status should default to active
			if tc.request.Status == nil {
				assert.Equal(t, enums.UserStatusActive, resp.CreateContact.Contact.Status)
			} else {
				assert.Equal(t, *tc.request.Status, resp.CreateContact.Contact.Status)
			}

			(&Cleanup[*generated.ContactDeleteOne]{client: suite.client.db.Contact, ID: resp.CreateContact.Contact.ID}).MustDelete(testUser1.UserCtx, t)
		})
	}
}

func TestMutationUpdateContact(t *testing.T) {
	contact := (&ContactBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateContactInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, update name",
			request: openlaneclient.UpdateContactInput{
				FullName: lo.ToPtr("Alicent Hightower"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "view only user cannot update",
			request: openlaneclient.UpdateContactInput{
				PhoneNumber: lo.ToPtr(gofakeit.Phone()),
			},
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name: "no access, cannot update",
			request: openlaneclient.UpdateContactInput{
				PhoneNumber: lo.ToPtr(gofakeit.Phone()),
			},
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name: "update phone number, using api token",
			request: openlaneclient.UpdateContactInput{
				PhoneNumber: lo.ToPtr(gofakeit.Phone()),
			},
			client: suite.client.apiWithToken,
			ctx:    context.Background(),
		},
		{
			name: "update status, using personal access token",
			request: openlaneclient.UpdateContactInput{
				Status: &enums.UserStatusInactive,
			},
			client: suite.client.apiWithPAT,
			ctx:    context.Background(),
		},
		{
			name: "update email",
			request: openlaneclient.UpdateContactInput{
				Email: lo.ToPtr("a.hightower@dragon.net"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update phone number, invalid",
			request: openlaneclient.UpdateContactInput{
				PhoneNumber: lo.ToPtr("not a phone number"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: rout.InvalidField("phone_number").Error(),
		},
		{
			name: "update email, invalid",
			request: openlaneclient.UpdateContactInput{
				Email: lo.ToPtr("a.hightower"),
			},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "validator failed for field",
		},
		{
			name: "update title",
			request: openlaneclient.UpdateContactInput{
				Title: lo.ToPtr("Queen of the Seven Kingdoms"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "update company",
			request: openlaneclient.UpdateContactInput{
				Company: lo.ToPtr("House Targaryen"),
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			resp, err := tc.client.UpdateContact(tc.ctx, contact.ID, tc.request)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			if tc.request.PhoneNumber != nil {
				assert.Equal(t, *tc.request.PhoneNumber, *resp.UpdateContact.Contact.PhoneNumber)
			}

			if tc.request.Email != nil {
				assert.Equal(t, *tc.request.Email, *resp.UpdateContact.Contact.Email)
			}

			if tc.request.FullName != nil {
				assert.Equal(t, *tc.request.FullName, resp.UpdateContact.Contact.FullName)
			}

			if tc.request.Title != nil {
				assert.Equal(t, *tc.request.Title, *resp.UpdateContact.Contact.Title)
			}

			if tc.request.Company != nil {
				assert.Equal(t, *tc.request.Company, *resp.UpdateContact.Contact.Company)
			}

			if tc.request.Status != nil {
				assert.Equal(t, *tc.request.Status, resp.UpdateContact.Contact.Status)
			}
		})
	}

	(&Cleanup[*generated.ContactDeleteOne]{client: suite.client.db.Contact, ID: contact.ID}).MustDelete(testUser1.UserCtx, t)
}

func TestMutationDeleteContact(t *testing.T) {
	contact1 := (&ContactBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	contact2 := (&ContactBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)
	contact3 := (&ContactBuilder{client: suite.client}).MustNew(testUser1.UserCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name:        "not allowed to delete, not enough permissions",
			idToDelete:  contact1.ID,
			client:      suite.client.api,
			ctx:         viewOnlyUser.UserCtx,
			expectedErr: notAuthorizedErrorMsg,
		},
		{
			name:        "not allowed to delete, no access to object",
			idToDelete:  contact1.ID,
			client:      suite.client.api,
			ctx:         testUser2.UserCtx,
			expectedErr: notFoundErrorMsg,
		},
		{
			name:       "happy path, delete contact",
			idToDelete: contact1.ID,
			client:     suite.client.api,
			ctx:        testUser1.UserCtx,
		},
		{
			name:        "contact already deleted, not found",
			idToDelete:  contact1.ID,
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "contact not found",
		},
		{
			name:       "happy path, delete contact using api token",
			idToDelete: contact2.ID,
			client:     suite.client.apiWithToken,
			ctx:        context.Background(),
		},
		{
			name:       "happy path, delete contact using pat",
			idToDelete: contact3.ID,
			client:     suite.client.apiWithPAT,
			ctx:        context.Background(),
		},
		{
			name:        "unknown contact, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "contact not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			resp, err := tc.client.DeleteContact(tc.ctx, tc.idToDelete)
			if tc.expectedErr != "" {
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Check(t, is.Nil(resp))

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)
			assert.Equal(t, tc.idToDelete, resp.DeleteContact.DeletedID)
		})
	}
}
