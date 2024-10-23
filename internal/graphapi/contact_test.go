package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"
	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestQueryContact() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	contact := (&ContactBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name     string
		queryID  string
		allowed  bool
		client   *openlaneclient.OpenlaneClient
		ctx      context.Context
		expected *ent.Contact
		errorMsg string
	}{
		{
			name:     "happy path contact",
			allowed:  true,
			queryID:  contact.ID,
			client:   suite.client.api,
			ctx:      reqCtx,
			expected: contact,
		},
		{
			name:     "no access",
			allowed:  false,
			queryID:  contact.ID,
			client:   suite.client.api,
			ctx:      reqCtx,
			errorMsg: "not authorized",
		},
		{
			name:     "happy path contact, with api token",
			allowed:  true,
			queryID:  contact.ID,
			client:   suite.client.apiWithToken,
			ctx:      context.Background(),
			expected: contact,
		},
		{
			name:     "no access, using api token",
			allowed:  false,
			queryID:  contact.ID,
			client:   suite.client.apiWithToken,
			ctx:      context.Background(),
			errorMsg: "not authorized",
		},
		{
			name:     "happy path contact, with pat",
			allowed:  true,
			queryID:  contact.ID,
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			expected: contact,
		},
		{
			name:     "no access, using pat",
			allowed:  false,
			queryID:  contact.ID,
			client:   suite.client.apiWithPAT,
			ctx:      context.Background(),
			errorMsg: "not authorized",
		},
	}

	for _, tc := range testCases {
		t.Run("Get "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.GetContactByID(tc.ctx, tc.queryID)

			if tc.errorMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.errorMsg)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Contact)
		})
	}

	// delete created org and contact
	(&ContactCleanup{client: suite.client, ID: contact.ID}).MustDelete(reqCtx, t)
}

func (suite *GraphTestSuite) TestQueryContacts() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	_ = (&ContactBuilder{client: suite.client}).MustNew(reqCtx, t)
	_ = (&ContactBuilder{client: suite.client}).MustNew(reqCtx, t)

	otherUser := (&UserBuilder{client: suite.client}).MustNew(reqCtx, t)
	otherCtx, err := userContextWithID(otherUser.ID)
	require.NoError(t, err)

	testCases := []struct {
		name            string
		client          *openlaneclient.OpenlaneClient
		ctx             context.Context
		expectedResults int
	}{
		{
			name:            "happy path",
			client:          suite.client.api,
			ctx:             reqCtx,
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
			ctx:             otherCtx,
			expectedResults: 0,
		},
	}

	for _, tc := range testCases {
		t.Run("List "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			resp, err := tc.client.GetAllContacts(tc.ctx)
			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Len(t, resp.Contacts.Edges, tc.expectedResults)
		})
	}
}

func (suite *GraphTestSuite) TestMutationCreateContact() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	testCases := []struct {
		name        string
		request     openlaneclient.CreateContactInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateContactInput{
				FullName: "Aemond Targaryen",
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "happy path, using api token",
			request: openlaneclient.CreateContactInput{
				FullName: "Rhaenys Targaryen",
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "happy path, using pat",
			request: openlaneclient.CreateContactInput{
				FullName: "Aegon Targaryen",
				OwnerID:  &testOrgID,
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateContactInput{
				FullName:    "Aemond Targaryen",
				Email:       lo.ToPtr("atargarygen@dragon.com"),
				PhoneNumber: lo.ToPtr(gofakeit.Phone()),
				Title:       lo.ToPtr("Prince of the Targaryen Dynasty"),
				Company:     lo.ToPtr("Targaryen Dynasty"),
				Status:      &enums.UserStatusOnboarding,
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "do not create if not allowed",
			request: openlaneclient.CreateContactInput{
				FullName: "Halaena Targaryen",
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: create on contact",
		},
		{
			name: "missing required field, name",
			request: openlaneclient.CreateContactInput{
				Email: lo.ToPtr("atargarygen@dragon.com"),
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.CreateContact(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			assert.Equal(t, tc.request.FullName, resp.CreateContact.Contact.FullName)

			if tc.request.Email == nil {
				assert.Empty(t, resp.CreateContact.Contact.Email)
			} else {
				assert.Equal(t, *tc.request.Email, *resp.CreateContact.Contact.Email)
			}

			if tc.request.PhoneNumber == nil {
				assert.Empty(t, resp.CreateContact.Contact.PhoneNumber)
			} else {
				assert.Equal(t, *tc.request.PhoneNumber, *resp.CreateContact.Contact.PhoneNumber)
			}

			if tc.request.Address == nil {
				assert.Empty(t, resp.CreateContact.Contact.Address)
			} else {
				assert.Equal(t, *tc.request.Address, *resp.CreateContact.Contact.Address)
			}

			if tc.request.Title == nil {
				assert.Empty(t, resp.CreateContact.Contact.Title)
			} else {
				assert.Equal(t, *tc.request.Title, *resp.CreateContact.Contact.Title)
			}

			if tc.request.Company == nil {
				assert.Empty(t, resp.CreateContact.Contact.Company)
			} else {
				assert.Equal(t, *tc.request.Company, *resp.CreateContact.Contact.Company)
			}

			// status should default to active
			if tc.request.Status == nil {
				assert.Equal(t, enums.UserStatusActive, resp.CreateContact.Contact.Status)
			} else {
				assert.Equal(t, *tc.request.Status, resp.CreateContact.Contact.Status)
			}
		})
	}
}

func (suite *GraphTestSuite) TestMutationUpdateContact() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	contact := (&ContactBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		request     openlaneclient.UpdateContactInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		expectedErr string
	}{
		{
			name: "happy path, update name",
			request: openlaneclient.UpdateContactInput{
				FullName: lo.ToPtr("Alicent Hightower"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "update phone number, using api token",
			request: openlaneclient.UpdateContactInput{
				PhoneNumber: lo.ToPtr(gofakeit.Phone()),
			},
			client:  suite.client.apiWithToken,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "update status, using personal access token",
			request: openlaneclient.UpdateContactInput{
				Status: &enums.UserStatusInactive,
			},
			client:  suite.client.apiWithPAT,
			ctx:     context.Background(),
			allowed: true,
		},
		{
			name: "update email",
			request: openlaneclient.UpdateContactInput{
				Email: lo.ToPtr("a.hightower@dragon.net"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "update phone number, invalid",
			request: openlaneclient.UpdateContactInput{
				PhoneNumber: lo.ToPtr("not a phone number"),
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: rout.InvalidField("phone_number").Error(),
		},
		{
			name: "update email, invalid",
			request: openlaneclient.UpdateContactInput{
				Email: lo.ToPtr("a.hightower"),
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     true,
			expectedErr: "validator failed for field",
		},
		{
			name: "update title",
			request: openlaneclient.UpdateContactInput{
				Title: lo.ToPtr("Queen of the Seven Kingdoms"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "update company",
			request: openlaneclient.UpdateContactInput{
				Company: lo.ToPtr("House Targaryen"),
			},
			client:  suite.client.api,
			ctx:     reqCtx,
			allowed: true,
		},
		{
			name: "not allowed to update",
			request: openlaneclient.UpdateContactInput{
				Company: lo.ToPtr("House Hightower"),
			},
			client:      suite.client.api,
			ctx:         reqCtx,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: update on contact",
		},
	}

	for _, tc := range testCases {
		t.Run("Update "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization
			mock_fga.CheckAny(t, suite.client.fga, tc.allowed)

			resp, err := tc.client.UpdateContact(tc.ctx, contact.ID, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

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
}

func (suite *GraphTestSuite) TestMutationDeleteContact() {
	t := suite.T()

	// setup user context
	reqCtx, err := userContext()
	require.NoError(t, err)

	contact1 := (&ContactBuilder{client: suite.client}).MustNew(reqCtx, t)
	contact2 := (&ContactBuilder{client: suite.client}).MustNew(reqCtx, t)
	contact3 := (&ContactBuilder{client: suite.client}).MustNew(reqCtx, t)

	testCases := []struct {
		name        string
		idToDelete  string
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		allowed     bool
		checkAccess bool
		expectedErr string
	}{
		{
			name:        "not allowed to delete",
			idToDelete:  contact1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     false,
			expectedErr: "you are not authorized to perform this action: delete on contact",
		},
		{
			name:        "happy path, delete contact",
			idToDelete:  contact1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "contact already deleted, not found",
			idToDelete:  contact1.ID,
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "contact not found",
		},
		{
			name:        "happy path, delete contact using api token",
			idToDelete:  contact2.ID,
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "happy path, delete contact using pat",
			idToDelete:  contact3.ID,
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			checkAccess: true,
			allowed:     true,
		},
		{
			name:        "unknown contact, not found",
			idToDelete:  ulids.New().String(),
			client:      suite.client.api,
			ctx:         reqCtx,
			checkAccess: false,
			allowed:     true,
			expectedErr: "contact not found",
		},
	}

	for _, tc := range testCases {
		t.Run("Delete "+tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.client.fga)

			// check for edit permissions on the organization if contact exists
			if tc.checkAccess {
				mock_fga.CheckAny(t, suite.client.fga, tc.allowed)
			}

			resp, err := suite.client.api.DeleteContact(reqCtx, tc.idToDelete)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tc.idToDelete, resp.DeleteContact.DeletedID)
		})
	}
}
