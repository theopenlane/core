package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	apimodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestOrgInviteAcceptHandler(t *testing.T) {
	suite.setupTest(t)

	// add handler
	suite.registerTestHandler("GET", "invite", suite.h.OrganizationInviteAccept)

	// bypass auth
	ctx := context.Background()
	ctx = auth.WithInternalOperationContext(testUser1.UserCtx)

	group, err := suite.db.Group.Create().
		SetName("Test Group").
		SetDescription("This is a test group").
		SetOwnerID(testUser1.OrganizationID).
		Save(ctx)
	assert.NilError(t, err)

	var (
		groot  = "groot@theopenlane.io"
		mantis = "mantis@theopenlane.io"
	)

	// recipient test data
	recipient, err := suite.db.User.Create().
		SetEmail(groot).
		SetFirstName("Groot").
		SetLastName("JustGroot").
		SetAuthProvider(enums.AuthProviderGoogle).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		Save(ctx)
	assert.NilError(t, err)

	userSetting, err := recipient.Setting(ctx)
	assert.NilError(t, err)

	suite.enableModules(recipient.ID, userSetting.Edges.DefaultOrg.ID,
		[]models.OrgModule{models.CatalogBaseModule})

	mantisRecipient, err := suite.db.User.Create().
		SetEmail(mantis).
		SetFirstName("Mantis").
		SetLastName("JustMantis").
		SetAuthProvider(enums.AuthProviderGoogle).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		Save(ctx)
	assert.NilError(t, err)

	mantisSetting, err := mantisRecipient.Setting(ctx)
	assert.NilError(t, err)

	suite.enableModules(mantisRecipient.ID, mantisSetting.Edges.DefaultOrg.ID,
		[]models.OrgModule{models.CatalogBaseModule})

	testCases := []struct {
		name           string
		email          string
		recipientID    string
		recipientOrgID string
		tokenSet       bool
		apiToken       bool
		wantErr        bool
		wantCode       int
		errMsg         string
	}{
		{
			name:           "api token cannot accept invite",
			email:          "rocket@theopenlane.io",
			recipientID:    recipient.ID,
			recipientOrgID: userSetting.Edges.DefaultOrg.ID,
			tokenSet:       true,
			apiToken:       true,
			wantErr:        true,
			wantCode:       http.StatusUnauthorized,
			errMsg:         handlers.ErrAuthenticationRequired.Error(),
		},
		{
			name:           "happy path",
			email:          groot,
			recipientID:    recipient.ID,
			recipientOrgID: userSetting.Edges.DefaultOrg.ID,
			tokenSet:       true,
		},
		{
			name:           "happy path, invite email case does not match user email",
			email:          "Mantis@TheOpenLane.io",
			recipientID:    mantisRecipient.ID,
			recipientOrgID: mantisSetting.Edges.DefaultOrg.ID,
			tokenSet:       true,
		},
		{
			name:           "missing token",
			email:          groot,
			recipientID:    recipient.ID,
			recipientOrgID: userSetting.Edges.DefaultOrg.ID,
			tokenSet:       false,
			wantErr:        true,
			wantCode:       http.StatusBadRequest,
			errMsg:         "token is required",
		},
		{
			name:           "emails do not match token",
			email:          "drax@theopenlane.io",
			recipientID:    recipient.ID,
			recipientOrgID: userSetting.Edges.DefaultOrg.ID,
			tokenSet:       true,
			wantErr:        true,
			wantCode:       http.StatusBadRequest,
			errMsg:         "could not verify email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.ClearTestData()

			ctx := auth.WithInternalOperationContext(testUser1.UserCtx)

			target := "/invite"
			if tc.tokenSet {
				inv, err := suite.db.Invite.Create().
					SetRecipient(tc.email).
					AddGroupIDs(group.ID).
					Save(ctx)
				assert.NilError(t, err)
				target = fmt.Sprintf("/invite?token=%s", inv.Token)
			}

			req := httptest.NewRequest(http.MethodGet, target, nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			reqCtx := auth.NewTestContextWithOrgID(tc.recipientID, tc.recipientOrgID)
			if tc.apiToken {
				reqCtx = auth.WithCaller(reqCtx, &auth.Caller{
					SubjectID:          ulids.New().String(),
					OrganizationID:     tc.recipientOrgID,
					AuthenticationType: auth.APITokenAuthentication,
				})
			}

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(reqCtx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *apimodels.InviteResponse

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			if tc.wantErr {
				assert.Check(t, is.Equal(tc.wantCode, recorder.Code))

				assert.Check(t, is.DeepEqual(tc.errMsg, out.Error))

				return
			}

			assert.Check(t, is.Equal(http.StatusCreated, recorder.Code))
			assert.Check(t, is.DeepEqual(testUser1.OrganizationID, out.JoinedOrgID))
			assert.Check(t, is.Equal(strings.ToLower(tc.email), out.Email))

			// update context to reflect the joined org for subsequent queries
			recipientCtx := auth.NewTestContextWithOrgID(tc.recipientID, testUser1.OrganizationID)

			// Test the default org is updated
			user, err := suite.api.GetUserByID(recipientCtx, tc.recipientID)
			assert.NilError(t, err)
			assert.Assert(t, user != nil)
			assert.Assert(t, user.User.Setting.DefaultOrg != nil)

			assert.Check(t, is.DeepEqual(testUser1.OrganizationID, user.User.Setting.DefaultOrg.ID))

			// test the user was added to the group
			group, err := suite.api.GetGroupByID(recipientCtx, group.ID)
			assert.Check(t, group != nil)

			foundMember := false
			for _, member := range group.Group.Members.Edges {
				if member.Node.User.ID == tc.recipientID {
					foundMember = true
					break
				}
			}

			assert.Check(t, foundMember, "expected user to be a member of the group")

			// verify the invite and acceptance emails were sent through the mock sender
			suite.WaitForEvents()

			msgs := suite.mockEmailSender().Messages()
			assert.Assert(t, is.Len(msgs, 2))

			// check that both expected subject snippets appear across the sent messages
			expectedSnippets := []string{
				"Join Your Teammate",
				"You've been added to an Organization",
			}

			found := make(map[string]bool)

			for _, msg := range msgs {
				for _, snippet := range expectedSnippets {
					if strings.Contains(msg.Subject, snippet) || strings.Contains(msg.HTML, snippet) || strings.Contains(msg.Text, snippet) {
						found[snippet] = true
						break
					}
				}
			}

			assert.Check(t, is.Len(found, 2))

			for _, snippet := range expectedSnippets {
				assert.Check(t, found[snippet], "expected snippet not found: %s", snippet)
			}
		})
	}
}

func TestOrgInviteAcceptHandler_ExistingMemberNoReInvitation(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("GET", "invite", suite.h.OrganizationInviteAccept)

	testUser := suite.userBuilder(context.Background())

	// bypass auth
	ctx := context.Background()
	ctx = auth.WithInternalOperationContext(testUser.UserCtx)

	var groot = "groot2@theopenlane.io"

	// recipient test data
	recipient, err := suite.db.User.Create().
		SetEmail(groot).
		SetFirstName("Groot").
		SetLastName("JustGroot Again").
		SetAuthProvider(enums.AuthProviderGoogle).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		Save(ctx)
	assert.NilError(t, err)

	userSetting, err := recipient.Setting(ctx)
	assert.NilError(t, err)

	recipientCtx := auth.NewTestContextWithOrgID(recipient.ID, userSetting.Edges.DefaultOrg.ID)

	suite.enableModules(recipient.ID, userSetting.Edges.DefaultOrg.ID,
		[]models.OrgModule{models.CatalogBaseModule})

	// invite user to org
	inv, err := suite.db.Invite.Create().
		SetRecipient(groot).
		Save(ctx)
	assert.NilError(t, err)

	target := fmt.Sprintf("/invite?token=%s", inv.Token)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	recorder := httptest.NewRecorder()

	// accept the invitation
	suite.e.ServeHTTP(recorder, req.WithContext(recipientCtx))
	res := recorder.Result()
	defer res.Body.Close()

	assert.Check(t, is.Equal(http.StatusCreated, recorder.Code))

	// create another invite for the same user
	_, err = suite.db.Invite.Create().
		SetRecipient(groot).
		Save(ctx)

	assert.Assert(t, err != nil)
	assert.Check(t, is.Contains(err.Error(), hooks.ErrUserAlreadyOrgMember.Error()))
}

func TestOrgInviteAcceptHandler_OwnerRoleBecomesSuperAdmin(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler(http.MethodGet, "invite", suite.h.OrganizationInviteAccept)

	internalCtx := auth.WithInternalOperationContext(testUser1.UserCtx)

	email := "newowner@theopenlane.io"

	recipient, err := suite.db.User.Create().
		SetEmail(email).
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetAuthProvider(enums.AuthProviderCredentials).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		Save(internalCtx)
	assert.NilError(t, err)

	setting, err := recipient.Setting(internalCtx)
	assert.NilError(t, err)

	ctx := auth.NewTestContextWithOrgID(recipient.ID, setting.Edges.DefaultOrg.ID)

	suite.enableModules(recipient.ID, setting.Edges.DefaultOrg.ID,
		[]models.OrgModule{models.CatalogBaseModule})

	invite, err := suite.db.Invite.Create().
		SetRecipient(email).
		SetRole(enums.RoleOwner).
		SetOwnershipTransfer(true).
		Save(internalCtx)
	assert.NilError(t, err)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/invite?token=%s", invite.Token), nil)
	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req.WithContext(ctx))
	res := recorder.Result()
	defer res.Body.Close()

	assert.Assert(t, is.Equal(http.StatusCreated, recorder.Code))

	membership, err := suite.db.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(testUser1.OrganizationID),
			orgmembership.UserID(testUser1.ID),
		).
		Only(internalCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.RoleSuperAdmin, membership.Role))

	newOwner, err := suite.db.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(testUser1.OrganizationID),
			orgmembership.UserID(recipient.ID),
		).
		Only(internalCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.RoleOwner, newOwner.Role))
}
