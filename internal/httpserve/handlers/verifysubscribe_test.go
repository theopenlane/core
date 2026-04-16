package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/httpserve/handlers"
)

func (suite *HandlerTestSuite) TestVerifySubscribeHandler() {
	t := suite.T()

	// add handler
	// Create operation for VerifySubscriptionHandler
	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("GET", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

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
			suite.ClearTestData()

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

			// verify email was sent through the mock sender
			if tc.emailExpected {
				msgs := suite.mockEmailSender().Messages()
				require.NotEmpty(t, msgs)
				assert.Contains(t, msgs[0].Subject, "subscribed")
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
	err := user.CreateVerificationToken()
	require.NoError(t, err)

	if ttl != "" {
		user.EmailVerificationExpires.String = ttl
	}

	expires, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	require.NoError(t, err)

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests
	reqCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	// store token in db
	return suite.db.Subscriber.Create().
		SetToken(user.EmailVerificationToken.String).
		SetEmail(user.Email).
		SetSecret(user.EmailVerificationSecret).
		SetTTL(expires).
		SaveX(reqCtx)
}
