package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestACMESolverHandler(t *testing.T) {
	suite.setupTest(t)

	// setup handler
	suite.registerTestHandler("GET", "/.well-known/acme-challenge/:path", suite.h.ACMESolverHandler)

	ec := echocontext.NewTestEchoContext().Request().Context()

	ctx := auth.WithOrgInternalCaller(ec, testUser1.OrganizationID)

	// Test data
	testPath := gofakeit.UUID()
	testValue := gofakeit.UUID()
	nonExistentPath := gofakeit.UUID()

	// Create a DNS verification record with ACME challenge data
	err := suite.db.DNSVerification.Create().
		SetCloudflareHostnameID(gofakeit.UUID()).
		SetDNSTxtRecord("_acme-challenge.example.com").
		SetDNSTxtValue(gofakeit.UUID()).
		SetDNSVerificationStatus(enums.DNSVerificationStatusPending).
		SetAcmeChallengePath(testPath).
		SetExpectedAcmeChallengeValue(testValue).
		SetAcmeChallengeStatus(enums.SSLVerificationStatusInitializing).
		SetOwnerID(testUser1.OrganizationID).
		Exec(ctx)
	assert.NilError(t, err)

	// Create a deleted DNS verification record (should not be found)
	err = suite.db.DNSVerification.Create().
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
		Exec(ctx)
	assert.NilError(t, err)

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

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))

			if tc.expectError {
				// For error cases, check that we get an error response
				var out *rout.Reply
				if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
					t.Error("error parsing error response", err)
				}
				assert.Assert(t, out != nil)
				assert.Check(t, !out.Success)
			} else {
				// For success cases, check that we get the expected ACME challenge value as plain text
				body := recorder.Body.String()
				assert.Check(t, is.DeepEqual(tc.expectedValue, body))
			}
		})
	}
}
