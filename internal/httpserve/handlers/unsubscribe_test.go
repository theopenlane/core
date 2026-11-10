package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestUnsubscribeHandler verifies the unsubscribe endpoint marks the subscriber unsubscribed (which the
// update hook also deactivates) on a valid token, and rejects a request without a token
func TestUnsubscribeHandler(t *testing.T) {
	suite := suite.setupTest(t)

	suite.registerTestHandler("POST", "unsubscribe", suite.h.UnsubscribeHandler)

	internalCtx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)

	t.Run("happy path unsubscribes an active subscriber", func(t *testing.T) {
		suite.ClearTestData()

		sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")
		assert.NilError(t, suite.db.Subscriber.UpdateOneID(sub.ID).SetVerifiedEmail(true).SetActive(true).Exec(internalCtx))

		target := fmt.Sprintf("/unsubscribe?token=%s", sub.Token)
		req := httptest.NewRequest(http.MethodPost, target, nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		res := recorder.Result()
		defer res.Body.Close()

		var out *models.UnsubscribeResponse
		assert.NilError(t, json.NewDecoder(res.Body).Decode(&out))

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))
		assert.Check(t, len(out.Message) > 0)

		updated, err := suite.db.Subscriber.Get(internalCtx, sub.ID)
		assert.NilError(t, err)
		assert.Check(t, updated.Unsubscribed)
		assert.Check(t, !updated.Active)
	})

	t.Run("happy path unsubscribes a trust center subscriber", func(t *testing.T) {
		suite.ClearTestData()

		// the create hook provisions the trust center's live setting (allow_subscribers defaults true)
		tc, err := suite.db.TrustCenter.Create().
			SetSlug("audit-unsub").
			SetOwnerID(suite.testUser1.OrganizationID).
			Save(suite.testUser1.UserCtx)
		assert.NilError(t, err)
		t.Cleanup(func() { _ = suite.db.TrustCenter.DeleteOneID(tc.ID).Exec(internalCtx) })

		sub := suite.createTestSubscriber(t, tc.ID, gofakeit.Email(), "")
		assert.NilError(t, suite.db.Subscriber.UpdateOneID(sub.ID).SetVerifiedEmail(true).SetActive(true).Exec(internalCtx))

		target := fmt.Sprintf("/unsubscribe?token=%s", sub.Token)
		req := httptest.NewRequest(http.MethodPost, target, nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		updated, err := suite.db.Subscriber.Get(internalCtx, sub.ID)
		assert.NilError(t, err)
		assert.Check(t, updated.Unsubscribed)
		assert.Check(t, !updated.Active)
		assert.Assert(t, updated.TrustCenterID != nil)
		assert.Check(t, is.DeepEqual(tc.ID, *updated.TrustCenterID))
	})

	t.Run("missing token is rejected", func(t *testing.T) {
		suite.ClearTestData()

		req := httptest.NewRequest(http.MethodPost, "/unsubscribe", nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))
	})

	t.Run("unknown token is rejected", func(t *testing.T) {
		suite.ClearTestData()

		req := httptest.NewRequest(http.MethodPost, "/unsubscribe?token=not-a-real-token", nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))
	})

	t.Run("reusing a token after the subscriber is already unsubscribed is idempotent", func(t *testing.T) {
		suite.ClearTestData()

		sub := suite.createTestSubscriber(t, "", gofakeit.Email(), "")
		assert.NilError(t, suite.db.Subscriber.UpdateOneID(sub.ID).SetVerifiedEmail(true).SetActive(true).Exec(internalCtx))

		target := fmt.Sprintf("/unsubscribe?token=%s", sub.Token)

		// first unsubscribe
		req := httptest.NewRequest(http.MethodPost, target, nil)
		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)
		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		// reuse the same token against an already-unsubscribed subscriber: idempotent success with a
		// distinct "already unsubscribed" acknowledgment, state unchanged
		req2 := httptest.NewRequest(http.MethodPost, target, nil)
		recorder2 := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder2, req2)
		assert.Check(t, is.Equal(http.StatusOK, recorder2.Code))

		res2 := recorder2.Result()
		defer res2.Body.Close()

		var out2 *models.UnsubscribeResponse
		assert.NilError(t, json.NewDecoder(res2.Body).Decode(&out2))
		assert.Check(t, is.Contains(out2.Message, "already unsubscribed"))

		updated, err := suite.db.Subscriber.Get(internalCtx, sub.ID)
		assert.NilError(t, err)
		assert.Check(t, updated.Unsubscribed)
		assert.Check(t, !updated.Active)
	})
}
