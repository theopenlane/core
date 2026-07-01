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

			sub := suite.createTestSubscriber(t, "", tc.email, tc.ttl)

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
				suite.WaitForEvents()

				msgs := suite.mockEmailSender().Messages()
				require.NotEmpty(t, msgs)
				assert.Contains(t, msgs[0].Subject, "subscription")
			}
		})
	}
}

// TestVerifySubscribeDoesNotResurrectUnsubscribed verifies that replaying a verify link for a contact
// who unsubscribed does NOT re-subscribe them: they stay unsubscribed and inactive. Re-subscription must
// go through the createSubscriber mutation (covered in the graphapi tests), not the verify endpoint
func (suite *HandlerTestSuite) TestVerifySubscribeDoesNotResurrectUnsubscribed() {
	t := suite.T()

	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("GET", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")

	// the contact verified previously and then opted out
	suite.db.Subscriber.UpdateOneID(sub.ID).SetVerifiedEmail(true).SetUnsubscribed(true).SetActive(false).ExecX(allowCtx)

	target := fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	// no resurrection: still unsubscribed and inactive
	updated := suite.db.Subscriber.GetX(allowCtx, sub.ID)
	assert.True(t, updated.Unsubscribed)
	assert.False(t, updated.Active)
}

// TestVerifySubscribeTrustCenterSubscriber confirms the handler verifies a subscriber scoped to a trust
// center: the org caller set in the handler resolves to the trust center's owning org, satisfying the
// org-ownership pre-policy on the update
func (suite *HandlerTestSuite) TestVerifySubscribeTrustCenterSubscriber() {
	t := suite.T()

	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("GET", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	// creating the trust center provisions its live setting (allow_subscribers defaults true), so a
	// subscriber scoped to it is permitted
	tc := suite.db.TrustCenter.Create().
		SetSlug("audit-verify").
		SetOwnerID(testUser1.OrganizationID).
		SaveX(testUser1.UserCtx)
	t.Cleanup(func() { _ = suite.db.TrustCenter.DeleteOneID(tc.ID).Exec(allowCtx) })

	sub := suite.createTestSubscriber(t, tc.ID, gofakeit.Email(), "")

	target := fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	updated := suite.db.Subscriber.GetX(allowCtx, sub.ID)
	assert.True(t, updated.VerifiedEmail)
	assert.True(t, updated.Active)
	require.NotNil(t, updated.TrustCenterID)
	assert.Equal(t, tc.ID, *updated.TrustCenterID)
}

// TestVerifySubscribeUnknownToken rejects a verify request whose token matches no subscriber
func (suite *HandlerTestSuite) TestVerifySubscribeUnknownToken() {
	t := suite.T()

	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("GET", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	req := httptest.NewRequest(http.MethodGet, "/subscribe/verify?token=not-a-real-token", nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestVerifySubscribeAlreadyVerified confirms re-hitting verify for an already active subscriber is a
// safe no-op: still 200 and the status is unchanged
func (suite *HandlerTestSuite) TestVerifySubscribeAlreadyVerified() {
	t := suite.T()

	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("GET", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")
	suite.db.Subscriber.UpdateOneID(sub.ID).SetVerifiedEmail(true).SetActive(true).ExecX(allowCtx)

	target := fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	updated := suite.db.Subscriber.GetX(allowCtx, sub.ID)
	assert.True(t, updated.VerifiedEmail)
	assert.True(t, updated.Active)
	assert.False(t, updated.Unsubscribed)
}

// createTestSubscriber is a helper to create a test subscriber. When trustCenterID is non-empty the
// subscriber is scoped to that trust center, otherwise it is an organization-level subscriber
func (suite *HandlerTestSuite) createTestSubscriber(t *testing.T, trustCenterID, email, ttl string) *ent.Subscriber {
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
	builder := suite.db.Subscriber.Create().
		SetToken(user.EmailVerificationToken.String).
		SetEmail(user.Email).
		SetSecret(user.EmailVerificationSecret).
		SetTTL(expires)

	if trustCenterID != "" {
		builder = builder.SetTrustCenterID(trustCenterID)
	}

	return builder.SaveX(reqCtx)
}
