package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/httpserve/handlers"
)

func (suite *HandlerTestSuite) TestOauthRegister() {
	t := suite.T()

	// add login handler
	// Create operation for OauthRegister
	operation := suite.createImpersonationOperation("OauthRegister", "OAuth register")
	suite.registerTestHandler("POST", "oauth/register", operation, suite.h.OauthRegister)

	ensureUserAbsent := func(t *testing.T, email string) {
		t.Helper()

		ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
		_, _ = suite.db.User.Delete().Where(user.Email(email)).Exec(ctx)
	}

	// Helper keeps the handler call consistent across subtests.
	send := func(t *testing.T, registerJSON models.OauthTokenRequest) (*httptest.ResponseRecorder, *models.LoginReply) {
		body, err := json.Marshal(registerJSON)
		if err != nil {
			require.NoError(t, err)
		}

		req := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(string(body)))
		req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		res := recorder.Result()
		defer res.Body.Close()

		var out *models.LoginReply
		if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
			t.Error("error parsing response", err)
		}

		return recorder, out
	}

	t.Run("happy path, github", func(t *testing.T) {
		suite.ClearTestData()

		email := "antman@theopenlane.io"
		ensureUserAbsent(t, email)

		registerJSON := models.OauthTokenRequest{
			Name:             "Ant Man",
			Email:            email,
			AuthProvider:     enums.AuthProviderGitHub.String(),
			ExternalUserID:   "123456",
			ExternalUserName: "scarletwitch",
			ClientToken:      "gh_thistokenisvalid",
		}

		recorder, out := send(t, registerJSON)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, rout.ErrorCode(""), out.ErrorCode)
		assert.NotNil(t, out.AccessToken)
		assert.NotNil(t, out.RefreshToken)
		assert.True(t, out.Success)
		assert.False(t, out.TFAEnabled) // we did not setup the user to have TFA
		assert.Equal(t, "Bearer", out.TokenType)

		msgs := suite.mockEmailSender().Messages()
		require.NotEmpty(t, msgs)
		assert.Contains(t, msgs[0].Subject, "Welcome to")
		assert.Equal(t, []string{email}, msgs[0].To)
	})

	// Keep "same user" within a single subtest so job-queue assertions don't
	// depend on subtest ordering or prior DB state.
	t.Run("happy path, github, same user", func(t *testing.T) {
		suite.ClearTestData()

		email := "antman@theopenlane.io"
		ensureUserAbsent(t, email)

		registerJSON := models.OauthTokenRequest{
			Name:             "Ant Man",
			Email:            email,
			AuthProvider:     enums.AuthProviderGitHub.String(),
			ExternalUserID:   "123456",
			ExternalUserName: "scarletwitch",
			ClientToken:      "gh_thistokenisvalid",
		}

		recorder, out := send(t, registerJSON)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// clear the job table so we can assert no new welcome email is sent
		suite.ClearTestData()

		recorder, out = send(t, registerJSON)
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.True(t, out.Success)

		msgs := suite.mockEmailSender().Messages()
		assert.Empty(t, msgs)
	})

	t.Run("mismatch email", func(t *testing.T) {
		suite.ClearTestData()

		registerJSON := models.OauthTokenRequest{
			Name:             "Ant Man",
			Email:            "antman@marvel.com",
			AuthProvider:     enums.AuthProviderGitHub.String(),
			ExternalUserID:   "123456",
			ExternalUserName: "scarletwitch",
			ClientToken:      "gh_thistokenisvalid",
		}

		recorder, out := send(t, registerJSON)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, handlers.InvalidInputErrCode, out.ErrorCode)
	})
}
