package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/echox/middleware/echocontext"

	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/common/models"
	apimodels "github.com/theopenlane/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/testutils"
)

func (suite *HandlerTestSuite) TestRefreshHandler() {
	t := suite.T()

	// add modules for the user
	modulesEnabled := []models.OrgModule{models.CatalogBaseModule, models.CatalogComplianceModule, models.CatalogEntityManagementModule}

	suite.enableModules(testUser1.ID, testUser1.OrganizationID, modulesEnabled)

	// add handler
	// Create operation for RefreshHandler
	operation := suite.createImpersonationOperation("RefreshHandler", "Refresh authentication token")
	suite.registerTestHandler("POST", "refresh", operation, suite.h.RefreshHandler)

	// Set full overlap of the refresh and access token so the refresh token is immediately valid
	tm, err := testutils.CreateTokenManager(-60 * time.Minute) //nolint:mnd
	require.NoError(t, err)

	suite.h.TokenManager = tm

	ec := echocontext.NewTestEchoContext().Request().Context()

	// set privacy allow in order to allow the creation of the users without
	// authentication in the tests
	ec = privacy.DecisionContext(ec, privacy.Allow)

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
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: user.ID,
		},
		UserID: user.ID,
		OrgID:  testUser1.OrganizationID,
	}

	_, refresh, err := tm.CreateTokenPair(claims)
	require.NoError(t, err)

	// ensure refresh token is valid by waiting
	time.Sleep(1 * time.Second)

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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			refreshJSON := apimodels.RefreshRequest{
				RefreshToken: tc.refresh,
			}

			body, err := json.Marshal(refreshJSON)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(string(body)))
			req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			res := recorder.Result()
			defer res.Body.Close()

			var out *apimodels.RefreshReply

			// parse request body
			err = json.NewDecoder(res.Body).Decode(&out)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			if tc.expectedStatus == http.StatusOK {
				assert.True(t, out.Success)

				jwt, err := tokens.ParseUnverifiedTokenClaims(out.AccessToken)
				require.NoError(t, err)

				// ensure the claims contain the modules
				assert.NotEmpty(t, jwt.Modules)
			} else {
				assert.Contains(t, out.Error, tc.expectedErrMessage)
			}
		})
	}
}
