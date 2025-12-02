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

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"
)

func (suite *HandlerTestSuite) TestACMESolverHandler() {
	t := suite.T()

	// setup handler
	// Create operation for ACMESolverHandler
	operation := suite.createImpersonationOperation("ACMESolverHandler", "ACME solver handler")
	suite.registerTestHandler("GET", "/.well-known/acme-challenge/:path", operation, suite.h.ACMESolverHandler)

	ec := echocontext.NewTestEchoContext().Request().Context()

	ctx := privacy.DecisionContext(ec, privacy.Allow)
	ctx = contextx.With(ctx, auth.AcmeSolverContextKey{})

	// Test data
	testPath := gofakeit.UUID()
	testValue := gofakeit.UUID()
	nonExistentPath := gofakeit.UUID()

	// Create a DNS verification record with ACME challenge data
	// Use ExecContext to bypass privacy policies completely
	_, err := suite.db.DNSVerification.Create().
		SetCloudflareHostnameID(gofakeit.UUID()).
		SetDNSTxtRecord("_acme-challenge.example.com").
		SetDNSTxtValue(gofakeit.UUID()).
		SetDNSVerificationStatus(enums.DNSVerificationStatusPending).
		SetAcmeChallengePath(testPath).
		SetExpectedAcmeChallengeValue(testValue).
		SetAcmeChallengeStatus(enums.SSLVerificationStatusInitializing).
		SetOwnerID(testUser1.OrganizationID).
		Save(ctx)
	require.NoError(t, err)

	// Create a deleted DNS verification record (should not be found)
	_, err = suite.db.DNSVerification.Create().
		SetCloudflareHostnameID(gofakeit.UUID()).
		SetDNSTxtRecord("_acme-challenge.deleted.com").
		SetDNSTxtValue(gofakeit.UUID()).
		SetDNSVerificationStatus(enums.DNSVerificationStatusPending).
		SetAcmeChallengePath("deleted-path").
		SetExpectedAcmeChallengeValue("deleted-value").
		SetAcmeChallengeStatus(enums.SSLVerificationStatusInitializing).
		SetOwnerID(testUser1.OrganizationID).
		SetDeletedAt(time.Now()).
		SetDeletedBy("test").
		Save(ctx)
	require.NoError(t, err)

	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		expectedValue  string
		expectError    bool
	}{
		{
			name:           "successful ACME challenge lookup",
			path:           testPath,
			expectedStatus: http.StatusOK,
			expectedValue:  testValue,
			expectError:    false,
		},
		{
			name:           "ACME challenge not found",
			path:           nonExistentPath,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "deleted ACME challenge should not be found",
			path:           "deleted-path",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := fmt.Sprintf("/.well-known/acme-challenge/%s", tc.path)

			req := httptest.NewRequest(http.MethodGet, target, nil)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectError {
				// For error cases, check that we get an error response
				var out *rout.Reply
				if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
					t.Error("error parsing error response", err)
				}
				require.NotNil(t, out)
				assert.False(t, out.Success)
			} else {
				// For success cases, check that we get the expected ACME challenge value as plain text
				body := recorder.Body.String()
				assert.Equal(t, tc.expectedValue, body)
			}
		})
	}
}
