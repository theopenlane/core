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
	suite.registerTestHandler("POST", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

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

			req := httptest.NewRequest(http.MethodPost, target, nil)

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
// who unsubscribed is rejected and does NOT re-subscribe them. Re-subscription must go through the
// createSubscriber mutation (covered in the graphapi tests), not the verify endpoint
func (suite *HandlerTestSuite) TestVerifySubscribeDoesNotResurrectUnsubscribed() {
	t := suite.T()

	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("POST", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")

	// opted out before confirming: token unused, but the unsubscribed guard still refuses to activate them
	require.NoError(t, suite.db.Subscriber.UpdateOneID(sub.ID).SetUnsubscribed(true).SetActive(false).Exec(allowCtx))

	target := fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)
	req := httptest.NewRequest(http.MethodPost, target, nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// no resurrection: still unsubscribed, inactive, and unverified
	updated, err := suite.db.Subscriber.Get(allowCtx, sub.ID)
	require.NoError(t, err)
	assert.True(t, updated.Unsubscribed)
	assert.False(t, updated.Active)
	assert.False(t, updated.VerifiedEmail)
}

// TestVerifySubscribeTrustCenterSubscriber confirms the handler verifies a subscriber scoped to a trust
// center: the org caller set in the handler resolves to the trust center's owning org, satisfying the
// org-ownership pre-policy on the update
func (suite *HandlerTestSuite) TestVerifySubscribeTrustCenterSubscriber() {
	t := suite.T()

	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("POST", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	// creating the trust center provisions its live setting (allow_subscribers defaults true), so a
	// subscriber scoped to it is permitted
	tc, err := suite.db.TrustCenter.Create().
		SetSlug("audit-verify").
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = suite.db.TrustCenter.DeleteOneID(tc.ID).Exec(allowCtx) })

	sub := suite.createTestSubscriber(t, tc.ID, gofakeit.Email(), "")

	target := fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)
	req := httptest.NewRequest(http.MethodPost, target, nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	updated, err := suite.db.Subscriber.Get(allowCtx, sub.ID)
	require.NoError(t, err)
	assert.True(t, updated.VerifiedEmail)
	assert.True(t, updated.Active)
	require.NotNil(t, updated.TrustCenterID)
	assert.Equal(t, tc.ID, *updated.TrustCenterID)
}

// TestVerifySubscribeUnknownToken rejects a verify request whose token matches no subscriber
func (suite *HandlerTestSuite) TestVerifySubscribeUnknownToken() {
	t := suite.T()

	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("POST", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	req := httptest.NewRequest(http.MethodPost, "/subscribe/verify?token=not-a-real-token", nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestVerifySubscribeTokenSingleUse confirms a verify token works once (200) and a replay is rejected
// (400) with no further change to the subscriber
func (suite *HandlerTestSuite) TestVerifySubscribeTokenSingleUse() {
	t := suite.T()

	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("POST", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")

	target := fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)

	// first use confirms the subscriber
	firstReq := httptest.NewRequest(http.MethodPost, target, nil)
	firstRec := httptest.NewRecorder()
	suite.e.ServeHTTP(firstRec, firstReq)
	assert.Equal(t, http.StatusOK, firstRec.Code)

	confirmed, err := suite.db.Subscriber.Get(allowCtx, sub.ID)
	require.NoError(t, err)
	require.True(t, confirmed.VerifiedEmail)
	require.True(t, confirmed.Active)

	// replaying the same token is rejected
	secondReq := httptest.NewRequest(http.MethodPost, target, nil)
	secondRec := httptest.NewRecorder()
	suite.e.ServeHTTP(secondRec, secondReq)
	assert.Equal(t, http.StatusBadRequest, secondRec.Code)

	// the replay changed nothing
	after, err := suite.db.Subscriber.Get(allowCtx, sub.ID)
	require.NoError(t, err)
	assert.True(t, after.VerifiedEmail)
	assert.True(t, after.Active)
	assert.False(t, after.Unsubscribed)
	assert.Equal(t, confirmed.SendAttempts, after.SendAttempts)
}

// TestVerifySubscribeTrustCenterExpiredResend confirms the auto-resent verification email for a trust
// center subscriber carries the rotated token in its links, not the token it just replaced
func (suite *HandlerTestSuite) TestVerifySubscribeTrustCenterExpiredResend() {
	t := suite.T()

	operation := suite.createImpersonationOperation("VerifySubscriptionHandler", "Verify subscription")
	suite.registerTestHandler("POST", "subscribe/verify", operation, suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	tc, err := suite.db.TrustCenter.Create().
		SetSlug("expired-resend").
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = suite.db.TrustCenter.DeleteOneID(tc.ID).Exec(allowCtx) })

	expiredTTL := time.Now().AddDate(0, 0, -1).Format(time.RFC3339Nano)
	sub := suite.createTestSubscriber(t, tc.ID, gofakeit.Email(), expiredTTL)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/subscribe/verify?token=%s", sub.Token), nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	updated, err := suite.db.Subscriber.Get(allowCtx, sub.ID)
	require.NoError(t, err)
	require.NotEqual(t, sub.Token, updated.Token)

	suite.WaitForEvents()

	msgs := suite.mockEmailSender().Messages()
	require.NotEmpty(t, msgs)

	resent := msgs[len(msgs)-1]
	assert.Contains(t, resent.HTML, updated.Token)
	assert.NotContains(t, resent.HTML, sub.Token)
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

	sub, err := builder.Save(reqCtx)
	require.NoError(t, err)

	// the create hook rotates token/ttl/secret, so an explicit ttl must be re-applied after the save
	if ttl != "" {
		sub, err = suite.db.Subscriber.UpdateOneID(sub.ID).SetTTL(expires).Save(reqCtx)
		require.NoError(t, err)
	}

	return sub
}
