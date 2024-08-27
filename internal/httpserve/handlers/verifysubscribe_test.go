package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/rShetty/asyncwait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mock_fga "github.com/theopenlane/iam/fgax/mockery"

	"github.com/theopenlane/utils/emails"
	"github.com/theopenlane/utils/emails/mock"
	"github.com/theopenlane/utils/ulids"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
)

func (suite *HandlerTestSuite) TestVerifySubscribeHandler() {
	t := suite.T()

	// add handler
	suite.e.GET("subscribe/verify", suite.h.VerifySubscriptionHandler)

	// bypass auth
	ctx := context.Background()
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	mock_fga.WriteAny(t, suite.fga)

	// setup test data
	user := suite.db.User.Create().
		SetEmail(gofakeit.Email()).
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SaveX(ctx)

	reqCtx, err := userContextWithID(user.ID)
	require.NoError(t, err)

	ctx = privacy.DecisionContext(reqCtx, privacy.Allow)

	input := openlaneclient.CreateOrganizationInput{
		Name: "mitb",
	}

	org, err := suite.api.CreateOrganization(ctx, input)
	require.NoError(t, err)

	expiredTTL := time.Now().AddDate(0, 0, -1).Format(time.RFC3339Nano)

	testCases := []struct {
		name           string
		email          string
		ttl            string
		tokenSet       bool
		emailExpected  bool
		expectedStatus int
	}{
		{
			name:           "happy path, new subscriber",
			email:          gofakeit.Email(),
			tokenSet:       true,
			emailExpected:  false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "expired token",
			email:          gofakeit.Email(),
			tokenSet:       true,
			ttl:            expiredTTL,
			emailExpected:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			email:          gofakeit.Email(),
			tokenSet:       false,
			emailExpected:  false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer mock_fga.ClearMocks(suite.fga)

			sent := time.Now()

			mock.ResetEmailMock()

			sub := suite.createTestSubscriber(t, org.CreateOrganization.Organization.ID, tc.email, tc.ttl)

			target := "/subscribe/verify"
			if tc.tokenSet {
				target = fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)
			}

			req := httptest.NewRequest(http.MethodGet, target, nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *models.VerifySubscribeReply

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				assert.NotEmpty(t, out.Message)
			}

			// Test that one verify email was sent to each user
			messages := []*mock.EmailMetadata{
				{
					To:        tc.email,
					From:      "mitb@theopenlane.io",
					Subject:   fmt.Sprintf(emails.Subscribed, "mitb"),
					Timestamp: sent,
				},
			}

			// wait for messages
			predicate := func() bool {
				return suite.h.TaskMan.GetQueueLength() == 0
			}
			successful := asyncwait.NewAsyncWait(maxWaitInMillis, pollIntervalInMillis).Check(predicate)

			if successful != true {
				t.Errorf("max wait of email send")
			}

			if tc.emailExpected {
				mock.CheckEmails(t, messages)
			}
		})
	}
}

// createTestSubscriber is a helper to create a test subscriber
func (suite *HandlerTestSuite) createTestSubscriber(t *testing.T, orgID, email, ttl string) *ent.Subscriber {
	user := handlers.User{
		Email: email,
	}

	// create token
	if err := user.CreateVerificationToken(); err != nil {
		require.NoError(t, err)
	}

	if ttl != "" {
		user.EmailVerificationExpires.String = ttl
	}

	expires, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	if err != nil {
		require.NoError(t, err)
	}

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests
	ctx, err := auth.NewTestContextWithOrgID(ulids.New().String(), orgID)
	require.NoError(t, err)

	reqCtx := privacy.DecisionContext(ctx, privacy.Allow)

	// store token in db
	return suite.db.Subscriber.Create().
		SetToken(user.EmailVerificationToken.String).
		SetEmail(user.Email).
		SetSecret(user.EmailVerificationSecret).
		SetTTL(expires).
		SaveX(reqCtx)
}
