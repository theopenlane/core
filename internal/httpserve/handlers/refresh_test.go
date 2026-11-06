package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	apimodels "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/testutils"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestRefreshHandler(t *testing.T) {
	suite := suite.setupParallelTest(t).withTestUsers()

	// add modules for the user
	modulesEnabled := []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule}

	suite.enableModules(suite.testUser1.ID, suite.testUser1.OrganizationID, modulesEnabled)

	// add handler
	suite.registerTestHandler("POST", "refresh", suite.h.RefreshHandler)

	// Set full overlap of the refresh and access token so the refresh token is immediately valid
	tm, err := testutils.CreateTokenManager(-60 * time.Minute) //nolint:mnd
	assert.NilError(t, err)

	suite.h.TokenManager = tm

	ec := echocontext.NewTestEchoContext().Request().Context()

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests
	ec = auth.WithInternalOperationContext(ec)

	// create user in the database
	validUser := gofakeit.Email()
	validPassword := gofakeit.Password(true, true, true, true, false, 20)

	userID := ulids.New().String()

	userSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ec)

	user := suite.db.User.Create().
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetEmail(validUser).
		SetPassword(validPassword).
		SetSetting(userSetting).
		SetID(userID).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SetSub(userID). // this is required to parse the refresh token
		SaveX(ec)

	claims := &tokens.Claims{
		Subject: user.ID,
		UserID:  user.ID,
		OrgID:   suite.testUser1.OrganizationID,
	}

	access, refresh, err := tm.CreateTokenPair(claims)
	assert.NilError(t, err)

	// Ensure refresh token is valid by waiting until its NotBefore time instead of a fixed sleep.
	nbf, err := tokens.NotBefore(refresh)
	assert.NilError(t, err)
	if !nbf.IsZero() {
		if wait := time.Until(nbf); wait > 0 {
			time.Sleep(wait + 10*time.Millisecond)
		}
	}

	testCases := []struct {
		name               string
		refresh            string
		expectedErrMessage string
		expectedStatus     int
	}{
		{
			name:           "happy path, valid credentials",
			refresh:        refresh,
			expectedStatus: http.StatusOK,
		},
		{
			name:               "empty refresh",
			refresh:            "",
			expectedStatus:     http.StatusBadRequest,
			expectedErrMessage: "refresh_token is required",
		},
		{
			name:           "access token cannot be used as a refresh token",
			refresh:        access,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			refreshJSON := apimodels.RefreshRequest{
				RefreshToken: tc.refresh,
			}

			body, err := json.Marshal(refreshJSON)
			assert.NilError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *apimodels.RefreshResponse

			// parse request body
			err = json.NewDecoder(res.Body).Decode(&out)
			assert.NilError(t, err)

			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))

			if tc.expectedStatus == http.StatusOK {
				assert.Check(t, out.Success)

				jwt, err := tokens.ParseUnverifiedTokenClaims(out.AccessToken)
				assert.NilError(t, err)

				// ensure the claims contain the modules
				assert.Check(t, len(jwt.Modules) > 0)
			} else {
				assert.Check(t, is.Contains(out.Error, tc.expectedErrMessage))
			}
		})
	}
}
