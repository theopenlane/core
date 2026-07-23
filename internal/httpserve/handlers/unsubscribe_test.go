package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// TestUnsubscribeHandler verifies the unsubscribe endpoint marks the subscriber unsubscribed (which the
// update hook also deactivates) on a valid token, and rejects a request without a token
func (suite *HandlerTestSuite) TestUnsubscribeHandler() {
	t := suite.T()

	suite.registerTestHandler("POST", "unsubscribe", suite.h.UnsubscribeHandler)

	allowCtx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	t.Run("happy path unsubscribes an active subscriber", func(t *testing.T) {
		suite.ClearTestData()

		sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")
		require.NoError(t, suite.db.Subscriber.UpdateOneID(sub.ID).SetVerifiedEmail(true).SetActive(true).Exec(allowCtx))

		target := fmt.Sprintf("/unsubscribe?token=%s", sub.Token)
		req := httptest.NewRequest(http.MethodPost, target, nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		res := recorder.Result()
		defer res.Body.Close()

		var out *models.UnsubscribeResponse
		require.NoError(t, json.NewDecoder(res.Body).Decode(&out))

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.NotEmpty(t, out.Message)

		updated, err := suite.db.Subscriber.Get(allowCtx, sub.ID)
		require.NoError(t, err)
		assert.True(t, updated.Unsubscribed)
		assert.False(t, updated.Active)
	})

	t.Run("happy path unsubscribes a trust center subscriber", func(t *testing.T) {
		suite.ClearTestData()

		// the create hook provisions the trust center's live setting (allow_subscribers defaults true)
		tc, err := suite.db.TrustCenter.Create().
			SetSlug("audit-unsub").
			SetOwnerID(testUser1.OrganizationID).
			Save(testUser1.UserCtx)
		require.NoError(t, err)
		t.Cleanup(func() { _ = suite.db.TrustCenter.DeleteOneID(tc.ID).Exec(allowCtx) })

		sub := suite.createTestSubscriber(t, tc.ID, gofakeit.Email(), "")
		require.NoError(t, suite.db.Subscriber.UpdateOneID(sub.ID).SetVerifiedEmail(true).SetActive(true).Exec(allowCtx))

		target := fmt.Sprintf("/unsubscribe?token=%s", sub.Token)
		req := httptest.NewRequest(http.MethodPost, target, nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		updated, err := suite.db.Subscriber.Get(allowCtx, sub.ID)
		require.NoError(t, err)
		assert.True(t, updated.Unsubscribed)
		assert.False(t, updated.Active)
		require.NotNil(t, updated.TrustCenterID)
		assert.Equal(t, tc.ID, *updated.TrustCenterID)
	})

	t.Run("missing token is rejected", func(t *testing.T) {
		suite.ClearTestData()

		req := httptest.NewRequest(http.MethodPost, "/unsubscribe", nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("unknown token is rejected", func(t *testing.T) {
		suite.ClearTestData()

		req := httptest.NewRequest(http.MethodPost, "/unsubscribe?token=not-a-real-token", nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("reusing a token after the subscriber is already unsubscribed is idempotent", func(t *testing.T) {
		suite.ClearTestData()

		sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")
		require.NoError(t, suite.db.Subscriber.UpdateOneID(sub.ID).SetVerifiedEmail(true).SetActive(true).Exec(allowCtx))

		target := fmt.Sprintf("/unsubscribe?token=%s", sub.Token)

		// first unsubscribe
		req := httptest.NewRequest(http.MethodPost, target, nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// reuse the same token against an already-unsubscribed subscriber: idempotent success with a
		// distinct "already unsubscribed" acknowledgment, state unchanged
		req2 := httptest.NewRequest(http.MethodPost, target, nil)
		recorder2 := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder2, req2)
		assert.Equal(t, http.StatusOK, recorder2.Code)

		res2 := recorder2.Result()
		defer res2.Body.Close()

		var out2 *models.UnsubscribeResponse
		require.NoError(t, json.NewDecoder(res2.Body).Decode(&out2))
		assert.Contains(t, out2.Message, "already unsubscribed")

		updated, err := suite.db.Subscriber.Get(allowCtx, sub.ID)
		require.NoError(t, err)
		assert.True(t, updated.Unsubscribed)
		assert.False(t, updated.Active)
	})
}
