package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"

	models "github.com/theopenlane/core/common/openapi"
)

func (suite *HandlerTestSuite) TestDomainScanHandler() {
	t := suite.T()

	suite.registerTestHandler("POST", "domain-scan", suite.h.DomainScanHandler)

	doRequest := func(t *testing.T, req models.DomainScanRequest, ctx context.Context) *httptest.ResponseRecorder {
		t.Helper()

		body, err := json.Marshal(req)
		require.NoError(t, err)

		httpReq := httptest.NewRequest(http.MethodPost, "/domain-scan", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq = httpReq.WithContext(ctx)

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, httpReq)

		return rec
	}

	t.Run("missing domain is a bad request", func(t *testing.T) {
		testUser := suite.userBuilder(context.Background())

		rec := doRequest(t, models.DomainScanRequest{}, testUser.UserCtx)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("unauthenticated request is rejected", func(t *testing.T) {
		rec := doRequest(t, models.DomainScanRequest{Domain: "example.com"}, context.Background())
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("second request for the same org within the hour is rate limited", func(t *testing.T) {
		testUser := suite.userBuilder(context.Background())

		first := doRequest(t, models.DomainScanRequest{Domain: "example.com"}, testUser.UserCtx)
		require.Equal(t, http.StatusOK, first.Code)

		var reply models.DomainScanReply
		require.NoError(t, json.Unmarshal(first.Body.Bytes(), &reply))
		assert.True(t, reply.Reply.Success)
		assert.NotEmpty(t, reply.ScanID)

		second := doRequest(t, models.DomainScanRequest{Domain: "example.com"}, testUser.UserCtx)
		assert.Equal(t, http.StatusTooManyRequests, second.Code)
	})

	t.Run("a second domain for the same org still counts against the rate limit", func(t *testing.T) {
		testUser := suite.userBuilder(context.Background())

		first := doRequest(t, models.DomainScanRequest{Domain: "example.com"}, testUser.UserCtx)
		require.Equal(t, http.StatusOK, first.Code)

		second := doRequest(t, models.DomainScanRequest{Domain: "sub.example.com"}, testUser.UserCtx)
		assert.Equal(t, http.StatusTooManyRequests, second.Code)
	})

	t.Run("system admins bypass the rate limit", func(t *testing.T) {
		testUser := suite.userBuilder(context.Background())

		adminCtx := auth.WithCaller(testUser.UserCtx, &auth.Caller{
			SubjectID:      testUser.ID,
			OrganizationID: testUser.OrganizationID,
			Capabilities:   auth.CapSystemAdmin,
		})

		first := doRequest(t, models.DomainScanRequest{Domain: "example.com"}, adminCtx)
		assert.Equal(t, http.StatusOK, first.Code)

		second := doRequest(t, models.DomainScanRequest{Domain: "example.com"}, adminCtx)
		assert.Equal(t, http.StatusOK, second.Code)
	})
}
