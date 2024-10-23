package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
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
	mock_fga.ListAny(t, suite.fga, []string{})

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

	mock_fga.ClearMocks(suite.fga)

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
			defer suite.ClearTestData()

			if !tc.wantErr {
				mock_fga.WriteAny(t, suite.fga)
			}

			ctx := privacy.DecisionContext(userCtx, privacy.Allow)

			invite := suite.db.Invite.Create().
				SetRecipient(tc.email).SaveX(ctx)

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

			mock_fga.ListAny(t, suite.fga, []string{"organization:" + org.CreateOrganization.Organization.ID})

			// Test the default org is updated
			user, err := suite.api.GetUserByID(recipientCtx, recipient.ID)
			require.NoError(t, err)
			require.NotNil(t, user)
			require.NotNil(t, user.User.Setting.DefaultOrg)

			assert.Equal(t, org.CreateOrganization.Organization.ID, user.User.Setting.DefaultOrg.ID)

			// ensure the email jobs are created
			// there will be two because the first is the invite email and the second is the accepted invite email
			job := rivertest.RequireManyInserted[*riverpgxv5.Driver](context.Background(), t, riverpgxv5.New(suite.db.Job.GetPool()),
				[]rivertest.ExpectedJob{
					{
						Args: jobs.EmailArgs{},
					},
					{
						Args: jobs.EmailArgs{},
					},
				})
			require.NotNil(t, job)
			assert.Contains(t, string(job[0].EncodedArgs), "Join your team")                       // first email is the invite email
			assert.Contains(t, string(job[1].EncodedArgs), "You've been added to an organization") // second email is the accepted invite email
		})
	}
}
