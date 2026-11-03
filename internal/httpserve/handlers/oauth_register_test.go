package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestOauthRegister(t *testing.T) {
	suite.setupTest(t)

	// add login handler
	suite.registerTestHandler("POST", "oauth/register", suite.h.OauthRegister)

	ensureUserAbsent := func(t *testing.T, email string) {
		t.Helper()

		ctx := auth.WithInternalOperationContext(testUser1.UserCtx)
		_, _ = suite.db.User.Delete().Where(user.Email(email)).Exec(ctx)
	}

	// Helper keeps the handler call consistent across subtests.
	send := func(t *testing.T, registerJSON models.OauthTokenRequest) (*httptest.ResponseRecorder, *models.LoginResponse) {
		body, err := json.Marshal(registerJSON)
		if err != nil {
			assert.NilError(t, err)
		}

		req := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(string(body)))
		req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		res := recorder.Result()
		defer res.Body.Close()

		var out *models.LoginResponse
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

		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))
		assert.Check(t, is.DeepEqual(rout.ErrorCode(""), out.ErrorCode))
		assert.Check(t, out.AccessToken != "")
		assert.Check(t, out.RefreshToken != "")
		assert.Check(t, out.Success)
		assert.Check(t, !out.TFAEnabled) // we did not setup the user to have TFA
		assert.Check(t, is.Equal("Bearer", out.TokenType))

		suite.WaitForEvents()

		msgs := suite.mockEmailSender().Messages()
		assert.Assert(t, len(msgs) > 0)
		assert.Check(t, is.Contains(msgs[0].Subject, "Welcome to"))
		assert.Check(t, is.DeepEqual([]string{email}, msgs[0].To))
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
		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))

		// clear the job table so we can assert no new welcome email is sent
		suite.ClearTestData()

		recorder, out = send(t, registerJSON)
		assert.Check(t, is.Equal(http.StatusOK, recorder.Code))
		assert.Check(t, out.Success)

		suite.WaitForEvents()

		msgs := suite.mockEmailSender().Messages()
		assert.Check(t, is.Len(msgs, 0))
	})

	t.Run("github requires enrollment when the orgs enforces tfa", func(t *testing.T) {
		suite.ClearTestData()
		suite.registerAuthenticatedTestHandler(http.MethodPost, tfaValidatePath, suite.h.ValidateTOTP)

		email := "antman@theopenlane.io"

		data := models.OauthTokenRequest{
			Name:             "User with TFA",
			Email:            email,
			AuthProvider:     enums.AuthProviderGitHub.String(),
			ExternalUserID:   "123456",
			ExternalUserName: "scarletwitch",
			ClientToken:      "gh_thistokenisvalid",
		}

		recorder, out := send(t, data)
		assert.Assert(t, is.Equal(http.StatusOK, recorder.Code))
		assert.Assert(t, out.Success)

		assert.Assert(t, out.AccessToken != "")

		orgID, ok := accessTokenClaims(t, out.AccessToken)["org"].(string)
		assert.Assert(t, ok)

		ctx := generated.NewContext(auth.WithInternalOperationContext(testUser1.UserCtx), suite.db)

		user, err := suite.db.User.Query().Where(user.Email(email)).WithSetting().Only(ctx)
		assert.NilError(t, err)
		assert.Assert(t, !user.Edges.Setting.IsTfaEnabled)

		err = suite.db.OrganizationSetting.Update().
			Where(organizationsetting.OrganizationID(orgID)).
			SetMultifactorAuthEnforced(true).
			Exec(ctx)
		assert.NilError(t, err)

		recorder, out = send(t, data)
		assert.Assert(t, is.Equal(http.StatusOK, recorder.Code))
		assert.Check(t, out.Success)
		assert.Check(t, out.TFASetupRequired)
		assert.Check(t, !out.TFAEnabled)
		assert.Assert(t, out.AccessToken != "")
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
		assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))
		assert.Check(t, is.DeepEqual(handlers.InvalidInputErrCode, out.ErrorCode))
	})

	t.Run("sso enforced default org redirects to sso instead of issuing tokens", func(t *testing.T) {
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

		recorder, _ := send(t, registerJSON)
		assert.Assert(t, is.Equal(http.StatusOK, recorder.Code))

		org := suite.ssoEnforcedOrg()

		antman, err := suite.db.User.Query().Where(user.Email(email)).WithSetting().Only(auth.WithInternalOperationContext(testUser1.UserCtx))
		assert.NilError(t, err)

		memberCtx := generated.NewContext(auth.WithInternalOperationContext(auth.NewTestContextWithOrgID(antman.ID, org.ID)), suite.db)

		assert.NilError(t, suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
			OrganizationID: org.ID,
			UserID:         antman.ID,
			Role:           &enums.RoleMember,
		}).Exec(memberCtx))

		assert.NilError(t, suite.db.UserSetting.UpdateOneID(antman.Edges.Setting.ID).SetDefaultOrgID(org.ID).Exec(memberCtx))

		body, err := json.Marshal(registerJSON)
		assert.NilError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(string(body)))
		req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

		rec := httptest.NewRecorder()
		suite.e.ServeHTTP(rec, req)

		assert.Check(t, is.Equal(http.StatusFound, rec.Code))
		assert.Check(t, !strings.Contains(rec.Body.String(), "access_token"))
	})
}
