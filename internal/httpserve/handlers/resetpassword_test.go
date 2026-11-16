package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/passwd"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/core/v2/internal/testutils"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestResetPasswordHandler(t *testing.T) {
	suite := suite.setupTest(t)

	// a functional blacklist is required for cutoff revocation to take effect
	tm, err := testutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
	assert.NilError(t, err)

	tm.WithBlacklist(tokens.NewRedisTokenBlacklist(suite.sharedRedisClient, "reset-test:"))

	suite.h.TokenManager = tm

	// setup request request
	suite.registerTestHandler("POST", "password-reset", suite.h.ResetPassword)

	ec := echocontext.NewTestEchoContext().Request().Context()

	var newPassword = "6z9Fqc-E-9v32NsJzLNU" //nolint:gosec

	expiredTTL := time.Now().AddDate(0, 0, -1).Format(time.RFC3339Nano)

	testCases := []struct {
		name           string
		email          string
		newPassword    string
		tokenSet       bool
		tokenProvided  string
		ttl            string
		expectedResp   string
		expectedStatus int
		from           string
	}{
		{
			name:           "happy path",
			email:          "kelsier@theopenlane.io",
			tokenSet:       true,
			newPassword:    newPassword,
			from:           "mitb@theopenlane.io",
			expectedResp:   emptyResponse,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "bad token (user not found)",
			email:          "eventure@theopenlane.io",
			tokenSet:       true,
			tokenProvided:  "thisisnotavalidtoken",
			newPassword:    newPassword,
			from:           "notactuallyanemail",
			expectedResp:   "password reset token invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "weak password",
			email:          "sazed@theopenlane.io",
			tokenSet:       true,
			newPassword:    "weak1",
			from:           "nottodaysatan",
			expectedResp:   "password is too weak",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "same password",
			email:          "sventure@theopenlane.io",
			tokenSet:       true,
			newPassword:    validPassword,
			from:           "mmhmm",
			expectedResp:   "password was already used",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing token",
			email:          "dockson@theopenlane.io",
			tokenSet:       false,
			newPassword:    newPassword,
			from:           "yadayadayada",
			expectedResp:   "token is required",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "expired reset token",
			email:          "tensoon@theopenlane.io",
			newPassword:    "6z9Fqc-E-9v32NsJzLNP",
			tokenSet:       true,
			from:           "zonkertons",
			ttl:            expiredTTL,
			expectedResp:   "reset token is expired, please request a new token using forgot-password",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// create user in the database
			rt, _, err := suite.createUserWithResetToken(t, ec, tc.email, tc.ttl)
			assert.NilError(t, err)

			userID := rt.OwnerID

			// issue an access token and a session before the reset; iat has second precision and a token
			// minted in the cutoff second stays valid, so the token is backdated instead of waiting
			oldClaims := &tokens.Claims{Subject: userID, UserID: userID}
			oldToken, err := tm.CreateAccessToken(oldClaims)
			assert.NilError(t, err)

			oldClaims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))

			oldAccess, err := tm.Sign(oldToken)
			assert.NilError(t, err)

			sessionRec := httptest.NewRecorder()
			_, err = suite.h.SessionConfig.CreateAndStoreSession(ec, sessionRec, userID)
			assert.NilError(t, err)

			sessionReq := httptest.NewRequest(http.MethodGet, "/", nil)
			for _, c := range sessionRec.Result().Cookies() {
				sessionReq.AddCookie(c)
			}

			session, err := suite.h.SessionConfig.SessionManager.Get(sessionReq, suite.h.SessionConfig.CookieConfig.Name)
			assert.NilError(t, err)

			sessionID := suite.h.SessionConfig.SessionManager.GetSessionIDFromCookie(session)

			exists, err := suite.h.SessionConfig.RedisStore.Exists(ec, sessionID)
			assert.NilError(t, err)
			assert.Assert(t, is.Equal(int64(1), exists))

			pwResetJSON := models.ResetPasswordRequest{
				Password: tc.newPassword,
			}

			if tc.tokenSet {
				pwResetJSON.Token = rt.Token
				if tc.tokenProvided != "" {
					pwResetJSON.Token = tc.tokenProvided
				}
			}

			body, err := json.Marshal(pwResetJSON)
			if err != nil {
				assert.NilError(t, err)
			}

			suite.ClearTestData()

			req := httptest.NewRequest(http.MethodPost, "/password-reset", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")

			// Set writer for tests that write on the response
			recorder := httptest.NewRecorder()

			// Using the ServerHTTP on echo will trigger the router and middleware
			suite.e.ServeHTTP(recorder, req)

			// get result
			res := recorder.Result()
			defer res.Body.Close()

			// check status
			assert.Check(t, is.Equal(tc.expectedStatus, recorder.Code))

			var out *models.ResetPasswordResponse

			// parse request body
			if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
				t.Error("error parsing response", err)
			}

			if tc.expectedStatus != http.StatusOK {
				assert.Check(t, is.Contains(out.Error, tc.expectedResp))

				// a failed reset leaves existing credentials intact
				_, err = tm.Verify(oldAccess)
				assert.Check(t, err == nil)
			} else {
				// tokens issued before the reset are rejected and the session is gone
				_, err = tm.Verify(oldAccess)
				assert.Check(t, is.ErrorIs(err, tokens.ErrTokenInvalid))

				exists, err := suite.h.SessionConfig.RedisStore.Exists(ec, sessionID)
				assert.NilError(t, err)
				assert.Check(t, is.Equal(int64(0), exists))

				// a pair created after the reset verifies
				newAccess, _, err := tm.CreateTokenPair(&tokens.Claims{Subject: userID, UserID: userID})
				assert.NilError(t, err)

				_, err = tm.Verify(newAccess)
				assert.NilError(t, err)

				suite.WaitForEvents()

				msgs := suite.mockEmailSender().Messages()
				assert.Assert(t, len(msgs) > 0)
				assert.Check(t, is.Contains(msgs[len(msgs)-1].Subject, "Password Reset"))
				assert.Assert(t, is.DeepEqual([]string{tc.email}, msgs[len(msgs)-1].To))
			}
		})
	}
}

// createUserWithResetToken creates a user with a valid reset token and returns the token, user id, and error if one occurred
// errRevocation simulates a redis failure while revoking credentials during a password reset
var errRevocation = errors.New("revocation failed")

// failingCutoffBlacklist wraps a TokenBlacklist but always fails RevokeIssuedBefore
type failingCutoffBlacklist struct {
	tokens.TokenBlacklist
}

// RevokeIssuedBefore always returns an error
func (failingCutoffBlacklist) RevokeIssuedBefore(context.Context, string, time.Time, time.Duration) error {
	return errRevocation
}

// failingUserSessionsStore wraps a PersistentStore but always fails DeleteUserSessions
type failingUserSessionsStore struct {
	sessions.PersistentStore
}

// DeleteUserSessions always returns an error
func (failingUserSessionsStore) DeleteUserSessions(context.Context, string) error {
	return errRevocation
}

// TestResetPasswordHandlerRevocationFailure verifies a reset whose credential revocation fails is rolled
// back, leaving the password and the reset token as they were so the reset can be retried
func TestResetPasswordHandlerRevocationFailure(t *testing.T) {
	testCases := []struct {
		name  string
		email string
		setup func(t *testing.T, suite *HandlerTestSuite)
	}{
		{
			name:  "token cutoff revocation fails",
			email: "vin@theopenlane.io",
			setup: func(t *testing.T, suite *HandlerTestSuite) {
				tm, err := testutils.CreateTokenManager(-15 * time.Minute) //nolint:mnd
				assert.NilError(t, err)

				tm.WithBlacklist(failingCutoffBlacklist{tokens.NewRedisTokenBlacklist(suite.sharedRedisClient, "reset-fail-test:")})

				suite.h.TokenManager = tm
			},
		},
		{
			name:  "session destruction fails",
			email: "elend@theopenlane.io",
			setup: func(_ *testing.T, suite *HandlerTestSuite) {
				// copy the config so the failing store does not leak into other tests
				sc := *suite.h.SessionConfig
				sc.RedisStore = failingUserSessionsStore{sc.RedisStore}

				suite.h.SessionConfig = &sc
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := suite.setupTest(t)

			tc.setup(t, suite)

			suite.registerTestHandler("POST", "password-reset", suite.h.ResetPassword)

			ec := echocontext.NewTestEchoContext().Request().Context()

			rt, userID, err := suite.createUserWithResetToken(t, ec, tc.email, "")
			assert.NilError(t, err)

			body, err := json.Marshal(models.ResetPasswordRequest{
				Password: "6z9Fqc-E-9v32NsJzLNU",
				Token:    rt.Token,
			})
			assert.NilError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/password-reset", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()

			suite.e.ServeHTTP(recorder, req)

			assert.Check(t, is.Equal(http.StatusInternalServerError, recorder.Code))

			internalCtx := auth.WithInternalOperationContext(ec)

			// the password change was rolled back
			user := suite.db.User.GetX(internalCtx, userID)

			valid, err := passwd.VerifyDerivedKey(*user.Password, validPassword)
			assert.NilError(t, err)
			assert.Check(t, valid)

			// the reset token was not consumed
			stored := suite.db.PasswordResetToken.GetX(internalCtx, rt.ID)
			assert.Check(t, stored.TTL.After(time.Now()))

			// no success email was sent
			suite.WaitForEvents()
			assert.Check(t, is.Len(suite.mockEmailSender().Messages(), 0))
		})
	}
}

func (suite *HandlerTestSuite) createUserWithResetToken(t *testing.T, ec context.Context, email string, ttl string) (*ent.PasswordResetToken, string, error) {
	ctx := auth.WithInternalOperationContext(ec)

	userSetting := suite.db.UserSetting.Create().
		SetEmailConfirmed(true).
		SaveX(ctx)

	u := suite.db.User.Create().
		SetFirstName(gofakeit.FirstName()).
		SetLastName(gofakeit.LastName()).
		SetEmail(email).
		SetPassword(validPassword).
		SetSetting(userSetting).
		SetLastLoginProvider(enums.AuthProviderCredentials).
		SetLastSeen(time.Now()).
		SaveX(ctx)

	user := handlers.User{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		ID:        u.ID,
	}

	// create token
	if err := user.CreatePasswordResetToken(); err != nil {
		return nil, "", err
	}

	if ttl != "" {
		user.PasswordResetExpires.String = ttl
	}

	// set expiry if provided in test case
	expiry, err := time.Parse(time.RFC3339Nano, user.PasswordResetExpires.String)
	if err != nil {
		return nil, "", err
	}

	// store token in db
	pr := suite.db.PasswordResetToken.Create().
		SetOwner(u).
		SetToken(user.PasswordResetToken.String).
		SetEmail(user.Email).
		SetSecret(user.PasswordResetSecret).
		SetTTL(expiry).
		SaveX(ctx)

	return pr, u.ID, nil
}
