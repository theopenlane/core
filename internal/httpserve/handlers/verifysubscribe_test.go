package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestVerifySubscribeHandler(t *testing.T) {
	suite.setupTest(t)

	// add handler
	suite.registerTestHandler("POST", "subscribe/verify", suite.h.VerifySubscriptionHandler)

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
			name:           "happy path, mixed case email",
			email:          "Star.Lord@TheOpenLane.io",
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

			var out *models.VerifySubscribeResponse

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))

			if tc.expectedStatus == http.StatusOK {
				assert.Check(t, len(out.Message) > 0)
			}

			// verify email was sent through the mock sender
			if tc.emailExpected {
				suite.WaitForEvents()

				msgs := suite.mockEmailSender().Messages()
				assert.Assert(t, len(msgs) > 0)
				assert.Check(t, is.Contains(msgs[0].Subject, "subscription"))
			}
		})
	}
}

// TestVerifySubscribeDoesNotResurrectUnsubscribed verifies that replaying a verify link for a contact
// who unsubscribed is rejected and does NOT re-subscribe them. Re-subscription must go through the
// createSubscriber mutation (covered in the graphapi tests), not the verify endpoint
func TestVerifySubscribeDoesNotResurrectUnsubscribed(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("POST", "subscribe/verify", suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	internalCtx := auth.WithInternalOperationContext(testUser1.UserCtx)
	sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")

	// opted out before confirming: token unused, but the unsubscribed guard still refuses to activate them
	assert.NilError(t, suite.db.Subscriber.UpdateOneID(sub.ID).SetUnsubscribed(true).SetActive(false).Exec(internalCtx))

	target := fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)
	req := httptest.NewRequest(http.MethodPost, target, nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))

	// no resurrection: still unsubscribed, inactive, and unverified
	updated, err := suite.db.Subscriber.Get(internalCtx, sub.ID)
	assert.NilError(t, err)
	assert.Check(t, updated.Unsubscribed)
	assert.Check(t, !updated.Active)
	assert.Check(t, !updated.VerifiedEmail)
}

// TestVerifySubscribeTrustCenterSubscriber confirms the handler verifies a subscriber scoped to a trust
// center: the org caller set in the handler resolves to the trust center's owning org, satisfying the
// org-ownership pre-policy on the update
func TestVerifySubscribeTrustCenterSubscriber(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("POST", "subscribe/verify", suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	internalCtx := auth.WithInternalOperationContext(testUser1.UserCtx)

	// creating the trust center provisions its live setting (allow_subscribers defaults true), so a
	// subscriber scoped to it is permitted
	tc, err := suite.db.TrustCenter.Create().
		SetSlug("audit-verify").
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	assert.NilError(t, err)
	t.Cleanup(func() { _ = suite.db.TrustCenter.DeleteOneID(tc.ID).Exec(internalCtx) })

	sub := suite.createTestSubscriber(t, tc.ID, gofakeit.Email(), "")

	target := fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)
	req := httptest.NewRequest(http.MethodPost, target, nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

	updated, err := suite.db.Subscriber.Get(internalCtx, sub.ID)
	assert.NilError(t, err)
	assert.Check(t, updated.VerifiedEmail)
	assert.Check(t, updated.Active)
	assert.Assert(t, updated.TrustCenterID != nil)
	assert.Check(t, is.DeepEqual(tc.ID, *updated.TrustCenterID))
}

// TestVerifySubscribeUnknownToken rejects a verify request whose token matches no subscriber
func TestVerifySubscribeUnknownToken(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("POST", "subscribe/verify", suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	req := httptest.NewRequest(http.MethodPost, "/subscribe/verify?token=not-a-real-token", nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))
}

// TestVerifySubscribeTokenSingleUse confirms a verify token works once (200) and a replay is rejected
// (400) with no further change to the subscriber
func TestVerifySubscribeTokenSingleUse(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("POST", "subscribe/verify", suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	internalCtx := auth.WithInternalOperationContext(testUser1.UserCtx)
	sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")

	target := fmt.Sprintf("/subscribe/verify?token=%s", sub.Token)

	// first use confirms the subscriber
	firstReq := httptest.NewRequest(http.MethodPost, target, nil)
	firstRec := httptest.NewRecorder()
	suite.e.ServeHTTP(firstRec, firstReq)
	assert.Check(t, is.Equal(http.StatusOK, firstRec.Code))

	confirmed, err := suite.db.Subscriber.Get(internalCtx, sub.ID)
	assert.NilError(t, err)
	assert.Assert(t, confirmed.VerifiedEmail)
	assert.Assert(t, confirmed.Active)

	// replaying the same token is rejected
	secondReq := httptest.NewRequest(http.MethodPost, target, nil)
	secondRec := httptest.NewRecorder()
	suite.e.ServeHTTP(secondRec, secondReq)
	assert.Check(t, is.Equal(http.StatusBadRequest, secondRec.Code))

	// the replay changed nothing
	after, err := suite.db.Subscriber.Get(internalCtx, sub.ID)
	assert.NilError(t, err)
	assert.Check(t, after.VerifiedEmail)
	assert.Check(t, after.Active)
	assert.Check(t, !after.Unsubscribed)
	assert.Check(t, is.DeepEqual(confirmed.SendAttempts, after.SendAttempts))
}

// TestVerifySubscribeTrustCenterExpiredResend confirms the auto-resent verification email for a trust
// center subscriber carries the rotated token in its links, not the token it just replaced
func TestVerifySubscribeTrustCenterExpiredResend(t *testing.T) {
	suite.setupTest(t)

	suite.registerTestHandler("POST", "subscribe/verify", suite.h.VerifySubscriptionHandler)

	suite.ClearTestData()

	internalCtx := auth.WithInternalOperationContext(testUser1.UserCtx)

	tc, err := suite.db.TrustCenter.Create().
		SetSlug("expired-resend").
		SetOwnerID(testUser1.OrganizationID).
		Save(testUser1.UserCtx)
	assert.NilError(t, err)
	t.Cleanup(func() { _ = suite.db.TrustCenter.DeleteOneID(tc.ID).Exec(internalCtx) })

	expiredTTL := time.Now().AddDate(0, 0, -1).Format(time.RFC3339Nano)
	sub := suite.createTestSubscriber(t, tc.ID, gofakeit.Email(), expiredTTL)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/subscribe/verify?token=%s", sub.Token), nil)
	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

	updated, err := suite.db.Subscriber.Get(internalCtx, sub.ID)
	assert.NilError(t, err)
	assert.Assert(t, sub.Token != updated.Token)

	suite.WaitForEvents()

	msgs := suite.mockEmailSender().Messages()
	assert.Assert(t, len(msgs) > 0)

	resent := msgs[len(msgs)-1]
	assert.Check(t, is.Contains(resent.HTML, updated.Token))
	assert.Check(t, !strings.Contains(resent.HTML, sub.Token))
}

// createTestSubscriber is a helper to create a test subscriber. When trustCenterID is non-empty the
// subscriber is scoped to that trust center, otherwise it is an organization-level subscriber
func (suite *HandlerTestSuite) createTestSubscriber(t *testing.T, trustCenterID, email, ttl string) *ent.Subscriber {
	user := handlers.User{
		Email: email,
	}

	// create token
	err := user.CreateVerificationToken()
	assert.NilError(t, err)

	if ttl != "" {
		user.EmailVerificationExpires.String = ttl
	}

	expires, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	assert.NilError(t, err)

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests
	reqCtx := auth.WithInternalOperationContext(testUser1.UserCtx)

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
	assert.NilError(t, err)

	// the create hook rotates token/ttl/secret, so an explicit ttl must be re-applied after the save
	if ttl != "" {
		sub, err = suite.db.Subscriber.UpdateOneID(sub.ID).SetTTL(expires).Save(reqCtx)
		assert.NilError(t, err)
	}

	return sub
}
