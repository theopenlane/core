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
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestVerifySubscribeHandler() {
	t := suite.T()

	// add handler
	suite.e.GET("subscribe/verify", suite.h.VerifySubscriptionHandler)

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
			sub := suite.createTestSubscriber(t, testUser1.OrganizationID, tc.email, tc.ttl)

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

			// ensure email job was created
			// the first job is the subscriber verification email and the following is the subscribed email
			if tc.emailExpected {
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
				assert.Contains(t, string(job[0].EncodedArgs), "Thank you for subscribing") // first email is the invite email
				assert.Contains(t, string(job[1].EncodedArgs), "You've been subscribed to") // second email is the accepted invite email
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
