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

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	apimodels "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/ent/hooks"
)

func (suite *HandlerTestSuite) TestOrgInviteAcceptHandler() {
	t := suite.T()

	// add handler
	// Create operation for OrganizationInviteAccept
	operation := suite.createImpersonationOperation("OrganizationInviteAccept", "Accept organization invite")
	suite.registerTestHandler("GET", "invite", operation, suite.h.OrganizationInviteAccept)

	// bypass auth
	ctx := context.Background()
	ctx = privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	group, err := suite.db.Group.Create().
		SetName("Test Group").
		SetDescription("This is a test group").
		SetOwnerID(testUser1.OrganizationID).
		Save(ctx)
	require.NoError(t, err)

	var groot = "groot@theopenlane.io"

	// recipient test data
	recipient, err := suite.db.User.Create().
		SetEmail(groot).
		SetFirstName("Groot").
		SetLastName("JustGroot").
		SetAuthProvider(enums.AuthProviderGoogle).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		Save(ctx)
	require.NoError(t, err)

	userSetting, err := recipient.Setting(ctx)
	require.NoError(t, err)

	recipientCtx := auth.NewTestContextWithOrgID(recipient.ID, userSetting.Edges.DefaultOrg.ID)

	suite.enableModules(recipient.ID, userSetting.Edges.DefaultOrg.ID,
		[]models.OrgModule{models.CatalogBaseModule})

	testCases := []struct {
		name     string
		email    string
		tokenSet bool
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "happy path",
			email:    groot,
			tokenSet: true,
		},
		{
			name:     "missing token",
			email:    groot,
			tokenSet: false,
			wantErr:  true,
			errMsg:   "token is required",
		},
		{
			name:     "emails do not match token",
			email:    "drax@theopenlane.io",
			tokenSet: true,
			wantErr:  true,
			errMsg:   "could not verify email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite.ClearTestData()

			ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

			target := "/invite"
			if tc.tokenSet {
				inv, err := suite.db.Invite.Create().
					SetRecipient(tc.email).
					AddGroupIDs(group.ID).
					Save(ctx)
				require.NoError(t, err)
				target = fmt.Sprintf("/invite?token=%s", inv.Token)
			}

			req := httptest.NewRequest(http.MethodGet, target, nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(recipientCtx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *apimodels.InviteReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			if tc.wantErr {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)

				assert.Equal(t, tc.errMsg, out.Error)

				return
			}

			assert.Equal(t, http.StatusCreated, recorder.Code)
			assert.Equal(t, testUser1.OrganizationID, out.JoinedOrgID)
			assert.Equal(t, tc.email, out.Email)

			// Test the default org is updated
			user, err := suite.api.GetUserByID(recipientCtx, recipient.ID)
			require.NoError(t, err)
			require.NotNil(t, user)
			require.NotNil(t, user.User.Setting.DefaultOrg)

			assert.Equal(t, testUser1.OrganizationID, user.User.Setting.DefaultOrg.ID)

			// test the user was added to the group
			group, err := suite.api.GetGroupByID(recipientCtx, group.ID)
			assert.NotNil(t, group)

			foundMember := false
			for _, member := range group.Group.Members.Edges {
				if member.Node.User.ID == recipient.ID {
					foundMember = true
					break
				}
			}

			assert.True(t, foundMember, "expected user to be a member of the group")

			// ensure the email jobs are created
			// there will be two because the first is the invite email
			// and the 2nd one is the accepted email
			job := rivertest.RequireManyInserted(context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
				[]rivertest.ExpectedJob{
					{
						Args: jobs.EmailArgs{},
					},
					{
						Args: jobs.EmailArgs{},
					},
				})
			require.NotNil(t, job)

			// We cannot determine the order of which they will be processed really especially for job 2 and 3
			// So just check and make sure they all contain these values at some point
			expectedSnippets := []string{
				"Join your team",
				"You've been added to an organization",
			}

			found := make(map[string]bool)

			for _, v := range job {
				for _, snippet := range expectedSnippets {
					if strings.Contains(string(v.EncodedArgs), snippet) {
						found[snippet] = true
						break
					}
				}
			}

			assert.Len(t, found, 2)

			for _, snippet := range expectedSnippets {
				assert.True(t, found[snippet], "expected snippet not found: %s", snippet)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestOrgInviteAcceptHandler_ExistingMemberNoReInvitation() {
	t := suite.T()

	operation := suite.createImpersonationOperation("OrganizationInviteAccept", "Accept organization invite")
	suite.registerTestHandler("GET", "invite", operation, suite.h.OrganizationInviteAccept)

	testUser := suite.userBuilder(context.Background())

	// bypass auth
	ctx := context.Background()
	ctx = privacy.DecisionContext(testUser.UserCtx, privacy.Allow)

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
	require.NoError(t, err)

	userSetting, err := recipient.Setting(ctx)
	require.NoError(t, err)

	recipientCtx := auth.NewTestContextWithOrgID(recipient.ID, userSetting.Edges.DefaultOrg.ID)

	suite.enableModules(recipient.ID, userSetting.Edges.DefaultOrg.ID,
		[]models.OrgModule{models.CatalogBaseModule})

	// invite user to org
	inv, err := suite.db.Invite.Create().
		SetRecipient(groot).
		Save(ctx)
	require.NoError(t, err)

	target := fmt.Sprintf("/invite?token=%s", inv.Token)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	recorder := httptest.NewRecorder()

	// accept the invitation
	suite.e.ServeHTTP(recorder, req.WithContext(recipientCtx))
	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, recorder.Code)

	// create another invite for the same user
	_, err = suite.db.Invite.Create().
		SetRecipient(groot).
		Save(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), hooks.ErrUserAlreadyOrgMember.Error())
}
