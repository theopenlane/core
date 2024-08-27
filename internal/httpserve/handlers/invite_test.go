package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mock_fga "github.com/datumforge/fgax/mockery"
	"github.com/rShetty/asyncwait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/pkg/auth"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/utils/emails"
	"github.com/theopenlane/utils/emails/mock"
)

func (suite *HandlerTestSuite) TestOrgInviteAcceptHandler() {
	t := suite.T()

	// add handler
	suite.e.GET("invite", suite.h.OrganizationInviteAccept)

	// bypass auth
	ctx := context.Background()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	mock_fga.WriteAny(t, suite.fga)
	mock_fga.CheckAny(t, suite.fga, true)

	// setup test data
	requestor := suite.db.User.Create().
		SetEmail("rocket@theopenlane.io").
		SetFirstName("Rocket").
		SetLastName("Racoon").
		SaveX(ctx)

	reqCtx, err := userContextWithID(requestor.ID)
	require.NoError(t, err)

	input := openlaneclient.CreateOrganizationInput{
		Name: "avengers",
	}

	org, err := suite.api.CreateOrganization(reqCtx, input)
	require.NoError(t, err)

	var groot = "groot@theopenlane.io"

	// recipient test data
	recipient := suite.db.User.Create().
		SetEmail(groot).
		SetFirstName("Groot").
		SetLastName("JustGroot").
		SetAuthProvider(enums.AuthProviderGoogle).
		SaveX(ctx)

	userCtx, err := auth.NewTestContextWithOrgID(requestor.ID, org.CreateOrganization.Organization.ID)
	require.NoError(t, err)

	userSetting, err := recipient.Setting(ctx)
	require.NoError(t, err)

	recipientCtx, err := auth.NewTestContextWithOrgID(recipient.ID, userSetting.Edges.DefaultOrg.ID)
	require.NoError(t, err)

	testCases := []struct {
		name          string
		email         string
		tokenSet      bool
		emailExpected bool
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "happy path",
			email:         groot,
			emailExpected: true,
			tokenSet:      true,
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
			defer mock_fga.ClearMocks(suite.fga)

			sent := time.Now()

			mock.ResetEmailMock()

			ctx := privacy.DecisionContext(userCtx, privacy.Allow)

			invite := suite.db.Invite.Create().
				SetRecipient(tc.email).SaveX(ctx)

			// wait for messages so we don't have conflicts with the accept message
			predicate := func() bool {
				return suite.h.TaskMan.GetQueueLength() == 0
			}

			asyncwait.NewAsyncWait(maxWaitInMillis, pollIntervalInMillis).Check(predicate)

			mock.ResetEmailMock()

			target := "/invite"
			if tc.tokenSet {
				target = fmt.Sprintf("/invite?token=%s", invite.Token)
			}

			req := httptest.NewRequest(http.MethodGet, target, nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req.WithContext(recipientCtx))

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.InviteReply

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
			assert.Equal(t, org.CreateOrganization.Organization.ID, out.JoinedOrgID)
			assert.Equal(t, tc.email, out.Email)

			// Test the default org is updated
			user, err := suite.api.GetUserByID(recipientCtx, recipient.ID)
			require.NoError(t, err)
			require.NotNil(t, user)
			require.NotNil(t, user.User.Setting.DefaultOrg)

			assert.Equal(t, org.CreateOrganization.Organization.ID, user.User.Setting.DefaultOrg.ID)

			// Test that one email was sent for accepted invite
			messages := []*mock.EmailMetadata{
				{
					To:        tc.email,
					From:      "mitb@theopenlane.io",
					Subject:   emails.InviteBeenAccepted,
					Timestamp: sent,
				},
			}

			// wait for messages
			predicate = func() bool {
				return suite.h.TaskMan.GetQueueLength() == 0
			}
			successful := asyncwait.NewAsyncWait(maxWaitInMillis, pollIntervalInMillis).Check(predicate)

			if successful != true {
				t.Errorf("max wait of email send")
			}

			if tc.emailExpected {
				mock.CheckEmails(t, messages)
			} else {
				mock.CheckEmails(t, nil)
			}
		})
	}
}
