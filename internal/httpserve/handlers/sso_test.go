package handlers_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/rs/zerolog/log"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/zitadel/oidc/v3/pkg/client"
	oidccrypto "github.com/zitadel/oidc/v3/pkg/crypto"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/internal/ent/generated/usersetting"
)

func TestWebfingerHandler(t *testing.T) {
	suite := suite.setupParallelTest(t).withTestUsers()

	suite.registerTestHandler("GET", ".well-known/webfinger", suite.h.WebfingerHandler)

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		MultifactorAuthEnforced: new(true),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.enforceSSOOnSetting(ctx, setting.ID)

	suite.db.UserSetting.Update().Where(usersetting.UserID(suite.testUser1.ID)).SetDefaultOrgID(org.ID).ExecX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=org:"+org.ID, nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	assert.Check(t, is.Equal(http.StatusOK, rec.Code))
	var out models.SSOStatusResponse
	assert.Check(t, json.NewDecoder(rec.Body).Decode(&out))
	log.Error().Err(errors.New("output")).Interface("out", out).Msg("WebfingerHandler output")
	assert.Check(t, out.Enforced)
	assert.Check(t, out.OrgTFAEnforced)
	assert.Check(t, is.DeepEqual(org.ID, out.OrganizationID))

	emailReq := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=acct:"+suite.testUser1.UserInfo.Email, nil)
	emailRec := httptest.NewRecorder()
	suite.e.ServeHTTP(emailRec, emailReq)
	assert.Check(t, is.Equal(http.StatusOK, emailRec.Code))
	var emailOut models.SSOStatusResponse
	assert.Check(t, json.NewDecoder(emailRec.Body).Decode(&emailOut))
	// suite.testUser1 is the organization owner, so the per-account lookup reflects the owner SSO exemption;
	// the org level lookup above still reports enforcement. TFA enforcement is independent of exemption
	assert.Check(t, !emailOut.Enforced)
	assert.Check(t, emailOut.OrgTFAEnforced)
	assert.Check(t, emailOut.IsOrgOwner)
	assert.Check(t, is.DeepEqual(org.ID, emailOut.OrganizationID))
}

func TestWebfingerHandlerTFAOnly(t *testing.T) {
	suite := suite.setupParallelTest(t).withTestUsers()

	suite.registerTestHandler("GET", ".well-known/webfinger", suite.h.WebfingerHandler)

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	// Create organization setting with only TFA enforced, not SSO
	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		MultifactorAuthEnforced: new(true),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.db.UserSetting.Update().Where(usersetting.UserID(suite.testUser1.ID)).SetDefaultOrgID(org.ID).ExecX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=org:"+org.ID, nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	assert.Check(t, is.Equal(http.StatusOK, rec.Code))
	var out models.SSOStatusResponse
	assert.Check(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Check(t, !out.Enforced)      // SSO not enforced
	assert.Check(t, out.OrgTFAEnforced) // TFA enforced
	assert.Check(t, is.DeepEqual(org.ID, out.OrganizationID))
}

func TestWebfingerHandlerNotFound(t *testing.T) {
	suite := suite.setupParallelTest(t)
	suite.registerTestHandler("GET", ".well-known/webfinger", suite.h.WebfingerHandler)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=acct:"+gofakeit.Email(), nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
}

// mockOIDCServer is a minimal OIDC provider used for testing the SSO flow.
// It exposes discovery, token and JWK endpoints. The ID tokens are signed
// with an in-memory RSA key.
type mockOIDCServer struct {
	server         *httptest.Server
	signer         jose.Signer
	privKey        *rsa.PrivateKey
	keyID          string
	expectedCode   string
	expectedSecret string
	nonce          string
	email          string
	name           string
	picture        string
}

type oidcServerOption func(*mockOIDCServer)

func withExpectedCode(code string) oidcServerOption {
	return func(m *mockOIDCServer) { m.expectedCode = code }
}

func withClientSecret(secret string) oidcServerOption {
	return func(m *mockOIDCServer) { m.expectedSecret = secret }
}

func withUserInfo(email, name, picture string) oidcServerOption {
	return func(m *mockOIDCServer) {
		m.email = email
		m.name = name
		m.picture = picture
	}
}

func newMockOIDCServer(t *testing.T, opts ...oidcServerOption) *mockOIDCServer {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Check(t, err)

	der := x509.MarshalPKCS1PrivateKey(priv)
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})

	signer, err := client.NewSignerFromPrivateKeyByte(pemKey, "test-kid")
	assert.Check(t, err)

	m := &mockOIDCServer{
		keyID:          "test-kid",
		signer:         signer,
		privKey:        priv,
		email:          "sso@example.com",
		name:           "SSO User",
		picture:        "https://example.com/avatar.png",
		expectedSecret: "secret",
	}

	for _, opt := range opts {
		opt(m)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                m.server.URL,
			"authorization_endpoint":                m.server.URL + "/auth",
			"token_endpoint":                        m.server.URL + "/token",
			"jwks_uri":                              m.server.URL + "/keys",
			"response_types_supported":              []string{"code"},
			"grant_types_supported":                 []string{"authorization_code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		})
	})

	mux.HandleFunc("/keys", func(w http.ResponseWriter, _ *http.Request) {
		jwk := jose.JSONWebKey{Key: &m.privKey.PublicKey, Algorithm: string(jose.RS256), Use: "sig", KeyID: m.keyID}
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}})
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json") // Ensure JSON content type for oauth2 client
		_ = r.ParseForm()
		if m.expectedCode != "" && r.Form.Get("code") != m.expectedCode {
			http.Error(w, "invalid code", http.StatusBadRequest)
			return
		}

		clientID := r.Form.Get("client_id")
		secret := r.Form.Get("client_secret")
		if clientID == "" || secret == "" {
			if id, pw, ok := r.BasicAuth(); ok {
				if clientID == "" {
					clientID = id
				}
				if secret == "" {
					secret = pw
				}
			}
		}

		if m.expectedSecret != "" && secret != m.expectedSecret {
			http.Error(w, "invalid client secret", http.StatusUnauthorized)
			return
		}

		claims := oidc.NewIDTokenClaims(
			m.server.URL,
			"1234",
			[]string{clientID},
			time.Now().Add(time.Hour),
			time.Now(),
			m.nonce,
			"",
			nil,
			clientID,
			0,
		)
		// Manually set the sub claim since the constructor does not include it in the JWT
		claimsMap := make(map[string]any)
		b, _ := json.Marshal(claims)
		_ = json.Unmarshal(b, &claimsMap)
		claimsMap["sub"] = "1234"
		// Inject email and email_verified claims for compatibility with callback handler
		claimsMap["email"] = m.email
		claimsMap["email_verified"] = true

		info := &oidc.UserInfo{
			Subject:       "1234",
			Name:          m.name,
			Picture:       m.picture,
			Email:         m.email,
			EmailVerified: true,
		}
		claims.SetUserInfo(info)

		// Re-marshal claimsMap to JSON and sign as JWT
		claimsBytes, _ := json.Marshal(claimsMap)
		var claimsInterface map[string]any
		_ = json.Unmarshal(claimsBytes, &claimsInterface)
		raw, err := oidccrypto.Sign(claimsInterface, m.signer)
		assert.Check(t, err)

		resp := map[string]any{
			"access_token": "access-token",
			"id_token":     raw,
			"token_type":   "Bearer",
			"expires_in":   3600,
		}

		_ = json.NewEncoder(w).Encode(resp)
	})

	m.server = httptest.NewServer(mux)
	return m
}

func (m *mockOIDCServer) Close() { m.server.Close() }

// TestSSOInitiateHandler verifies the public, slug-keyed SSO entry point resolves the organization, returns
// the identity provider redirect URL, and sets the state/nonce/organization_id cookies the callback needs,
// without requiring the caller to be authenticated or a member
func TestSSOInitiateHandler(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerTestHandler("GET", "v1/orgs/:slug_name/sso", suite.h.SSOInitiateHandler)

	oidc := newMockOIDCServer(t)
	defer oidc.Close()

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	discovery := oidc.server.URL + "/.well-known/openid-configuration"
	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProvider:             new(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:        &discovery,
		IdentityProviderClientID:     new("client"),
		IdentityProviderClientSecret: new("secret"),
	}).Save(ctx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ctx)
	assert.NilError(t, err)

	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(setting.ID).SetOrganizationID(org.ID).Exec(ctx))

	assert.Assert(t, len(org.SlugName) > 0, "the organization create hook should derive a slug from the name")

	// a known slug resolves the organization and returns the IdP redirect plus the flow cookies
	req := httptest.NewRequest(http.MethodGet, "/v1/orgs/"+org.SlugName+"/sso", nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))

	var out models.SSOLoginResponse
	assert.NilError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.Check(t, out.Success)
	assert.Check(t, len(out.RedirectURI) > 0, "should return the identity provider redirect URL")
	assert.Check(t, is.Contains(out.RedirectURI, oidc.server.URL), "redirect should point at the configured identity provider")

	cookies := map[string]string{}
	for _, c := range rec.Result().Cookies() {
		cookies[c.Name] = c.Value
	}
	assert.Check(t, len(cookies["state"]) > 0, "state cookie should be set for the callback to validate")
	assert.Check(t, len(cookies["nonce"]) > 0, "nonce cookie should be set for the callback to validate")
	assert.Check(t, is.DeepEqual(org.ID, cookies["organization_id"]), "organization_id cookie should be resolved from the slug")

	// an unknown slug returns not found rather than starting a flow
	missingReq := httptest.NewRequest(http.MethodGet, "/v1/orgs/this-slug-does-not-exist/sso", nil)
	missingRec := httptest.NewRecorder()
	suite.e.ServeHTTP(missingRec, missingReq)
	assert.Check(t, is.Equal(http.StatusNotFound, missingRec.Code))
}

func TestSSOLoginAndCallback(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerTestHandler("GET", "v1/sso/login", suite.h.SSOLoginHandler)
	suite.registerTestHandler("GET", "v1/sso/callback", suite.h.SSOCallbackHandler)

	oidc := newMockOIDCServer(t,
		withExpectedCode("code123"),
		withClientSecret("secret"),
		withUserInfo("sso@example.com", "SSO User", "https://example.com/avatar.png"),
	)
	defer oidc.Close()

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	ssoUser := suite.db.User.Create().
		SetEmail("sso@example.com").
		SetFirstName("SSO").
		SetLastName("User").
		SetLastLoginProvider(enums.AuthProviderOIDC).
		SetLastSeen(time.Now()).
		SaveX(ctx)

	discovery := oidc.server.URL + "/.well-known/openid-configuration"
	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: new(true),
		IdentityProvider:              new(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:         &discovery,
		IdentityProviderClientID:      new("client"),
		IdentityProviderClientSecret:  new("secret"),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		ExecX(ctx)

	ctxTargetOrg := auth.NewTestContextWithOrgID(ssoUser.ID, org.ID)
	ctxTargetOrg = auth.WithInternalOperationContext(ctxTargetOrg)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

	suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         ssoUser.ID,
		Role:           &enums.RoleMember,
	}).ExecX(testUserCtx)

	suite.db.UserSetting.Update().Where(usersetting.UserID(ssoUser.ID)).SetDefaultOrgID(org.ID).ExecX(ctx)

	ownPAT, err := suite.db.PersonalAccessToken.Create().
		SetName(ulids.New().String()).
		AddOrganizationIDs(org.ID).
		Save(testUserCtx)
	assert.NilError(t, err)

	otherUserPAT, err := suite.db.PersonalAccessToken.Create().
		SetName(ulids.New().String()).
		AddOrganizationIDs(suite.testUser1.OrganizationID).
		Save(ctx)
	assert.NilError(t, err)

	otherOrgAPIToken, err := suite.db.APIToken.Create().
		SetName(ulids.New().String()).
		Save(ent.NewContext(auth.NewTestContextWithOrgID(suite.testUser1.ID, suite.testUser1.OrganizationID), suite.db))
	assert.NilError(t, err)

	testCases := []struct {
		name         string
		tokenID      string
		tokenType    string
		expectedCode int
	}{
		{
			name:         "no token",
			expectedCode: http.StatusOK,
		},
		{
			name:         "own personal access token is authorized",
			tokenID:      ownPAT.ID,
			tokenType:    "personal",
			expectedCode: http.StatusOK,
		},
		{
			name:         "another user's personal access token is rejected",
			tokenID:      otherUserPAT.ID,
			tokenType:    "personal",
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "another organization's api token is rejected",
			tokenID:      otherOrgAPIToken.ID,
			tokenType:    "api",
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/sso/login?organization_id="+org.ID, nil)
			rec := httptest.NewRecorder()
			suite.e.ServeHTTP(rec, req)

			assert.Check(t, is.Equal(http.StatusOK, rec.Code))

			cookies := rec.Result().Cookies()
			var state, nonce *http.Cookie
			for _, c := range cookies {
				switch c.Name {
				case "state":
					state = c
				case "nonce":
					nonce = c
				}
			}

			assert.Assert(t, state != nil)
			assert.Assert(t, nonce != nil)

			oidc.nonce = nonce.Value

			// Set cookies as a single Cookie header instead of AddCookie
			cookieHeader := "state=" + state.Value + "; nonce=" + nonce.Value + "; organization_id=" + org.ID
			if tc.tokenID != "" {
				cookieHeader += "; token_id=" + tc.tokenID + "; token_type=" + tc.tokenType
			}

			cbReq := httptest.NewRequest(http.MethodGet, "/v1/sso/callback?code=code123&state="+url.QueryEscape(state.Value)+"&organization_id="+org.ID, nil)
			cbReq.Header.Set("Cookie", cookieHeader)
			cbRec := httptest.NewRecorder()

			suite.e.ServeHTTP(cbRec, cbReq)
			assert.Check(t, is.Equal(tc.expectedCode, cbRec.Code))

			if tc.expectedCode != http.StatusOK {
				return
			}

			var out models.LoginResponse
			assert.Check(t, json.NewDecoder(cbRec.Body).Decode(&out))
			assert.Check(t, out.Success)

			if tc.tokenID != "" {
				pat, err := suite.db.PersonalAccessToken.Get(testUserCtx, tc.tokenID)
				assert.NilError(t, err)
				assert.Check(t, is.Contains(pat.SSOAuthorizations, org.ID))
			}
		})
	}
}

// TestSSOCallbackJITProvisioning verifies that when SSO login is enforced and just-in-time provisioning is
// enabled, a user who authenticates against the configured identity provider but is not yet a member is
// provisioned into the organization as a member during the callback
type ssoScenario struct {
	email       string
	enforce     bool
	jit         bool
	jitDomains  []string
	preexisting bool
	cookieOrgID string
}

// ssoJITScenario runs the full SSO login + callback flow for a non-member user with the given email
// against an organization configured with the provided enforcement and JIT settings, and reports whether
// the user was provisioned into the organization. Handlers must be registered by the caller
func (suite *HandlerTestSuite) ssoJITScenario(t *testing.T, sc ssoScenario) (bool, int) {
	t.Helper()

	oidc := newMockOIDCServer(t,
		withExpectedCode("code123"),
		withClientSecret("secret"),
		withUserInfo(sc.email, "JIT User", ""),
	)
	defer oidc.Close()

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	if sc.preexisting {
		_, err := suite.db.User.Create().
			SetEmail(sc.email).
			SetFirstName("JIT").
			SetLastName("User").
			SetLastLoginProvider(enums.AuthProviderOIDC).
			SetLastSeen(time.Now()).
			Save(ctx)
		assert.NilError(t, err)
	}

	discovery := oidc.server.URL + "/.well-known/openid-configuration"
	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProvider:                new(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:           &discovery,
		IdentityProviderClientID:        new("client"),
		IdentityProviderClientSecret:    new("secret"),
		IdentityProviderJitProvisioning: &sc.jit,
		JitAllowedEmailDomains:          sc.jitDomains,
	}).Save(ctx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ctx)
	assert.NilError(t, err)

	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		SetIdentityProviderAuthTested(true).
		Exec(ctx))

	if sc.enforce {
		assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(setting.ID).
			SetIdentityProviderLoginEnforced(true).
			Exec(ctx))
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/sso/login?organization_id="+org.ID, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))

	var state, nonce *http.Cookie
	for _, c := range rec.Result().Cookies() {
		switch c.Name {
		case "state":
			state = c
		case "nonce":
			nonce = c
		}
	}
	assert.Assert(t, state != nil)
	assert.Assert(t, nonce != nil)

	oidc.nonce = nonce.Value

	cookieOrgID := lo.CoalesceOrEmpty(sc.cookieOrgID, org.ID)

	cookieHeader := "state=" + state.Value + "; nonce=" + nonce.Value + "; organization_id=" + cookieOrgID
	cbReq := httptest.NewRequest(http.MethodGet, "/v1/sso/callback?code=code123&state="+url.QueryEscape(state.Value)+"&organization_id="+org.ID, nil)
	cbReq.Header.Set("Cookie", cookieHeader)
	cbRec := httptest.NewRecorder()
	suite.e.ServeHTTP(cbRec, cbReq)

	// JIT provisioning runs before session generation, so the membership decision is observable even when
	// an enforced org rejects the session for a user it did not provision
	member, err := suite.db.OrgMembership.Query().
		Where(orgmembership.HasUserWith(user.EmailEqualFold(sc.email)), orgmembership.OrganizationID(org.ID)).
		Exist(ctx)
	assert.NilError(t, err)

	return member, cbRec.Code
}

// TestSSOCallbackJITPermutations covers the JIT provisioning conditions: the domain allowlist (in-list vs
// not-in-list) and that JIT is skipped when the toggle is off or SSO is not enforced
func TestSSOCallbackJITPermutations(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerTestHandler("GET", "v1/sso/login", suite.h.SSOLoginHandler)
	suite.registerTestHandler("GET", "v1/sso/callback", suite.h.SSOCallbackHandler)

	// allowlist match: enforced + JIT on + domain in jit_allowed_email_domains => provisioned
	member, code := suite.ssoJITScenario(t, ssoScenario{email: "match@allowed-jit.com", enforce: true, jit: true, jitDomains: []string{"allowed-jit.com"}})
	assert.Check(t, member, "a directory user whose domain is in the JIT allowlist should be provisioned")
	assert.Check(t, is.Equal(http.StatusOK, code), "a provisioned user should complete sign-in")

	// allowlist miss: enforced + JIT on but domain NOT in the allowlist => not provisioned, and the
	// callback returns a clean forbidden response rather than a server error
	member, code = suite.ssoJITScenario(t, ssoScenario{email: "stranger@other-jit.com", enforce: true, jit: true, jitDomains: []string{"allowed-jit.com"}})
	assert.Check(t, !member, "a directory user whose domain is not in the JIT allowlist should not be provisioned")
	assert.Check(t, is.Equal(http.StatusForbidden, code), "an unprovisioned user on an enforced org should get a forbidden response, not a server error")

	// JIT toggle off: enforced but JIT disabled => not provisioned, clean forbidden response
	member, code = suite.ssoJITScenario(t, ssoScenario{email: "nojit@example-jit.com", enforce: true})
	assert.Check(t, !member, "JIT should not provision when the toggle is off")
	assert.Check(t, is.Equal(http.StatusForbidden, code), "an unprovisioned user on an enforced org should get a forbidden response, not a server error")

	// not enforced: JIT on but SSO not enforced => not provisioned (membership comes from auto-join instead)
	member, _ = suite.ssoJITScenario(t, ssoScenario{email: "unenforced@example-jit.com", jit: true})
	assert.Check(t, !member, "JIT should not provision when SSO is not enforced")

	member, code = suite.ssoJITScenario(t, ssoScenario{email: "existing@allowed-jit.com", enforce: true, jit: true, preexisting: true})
	assert.Check(t, !member)
	assert.Check(t, is.Equal(http.StatusForbidden, code))

	member, code = suite.ssoJITScenario(t, ssoScenario{email: "mismatch@allowed-jit.com", enforce: true, jit: true, cookieOrgID: ulids.New().String()})
	assert.Check(t, !member)
	assert.Check(t, is.Equal(http.StatusBadRequest, code))
}

func TestSSOCallbackJITProvisioning(t *testing.T) {
	suite := suite.setupParallelTest(t)

	suite.registerTestHandler("GET", "v1/sso/login", suite.h.SSOLoginHandler)
	suite.registerTestHandler("GET", "v1/sso/callback", suite.h.SSOCallbackHandler)

	oidc := newMockOIDCServer(t,
		withExpectedCode("code123"),
		withClientSecret("secret"),
		withUserInfo("jit@example.com", "JIT User", ""),
	)
	defer oidc.Close()

	ctx := auth.WithInternalOperationContext(suite.testUser1.UserCtx)
	ctx = ent.NewContext(ctx, suite.db)

	discovery := oidc.server.URL + "/.well-known/openid-configuration"
	setting, err := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProvider:             new(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:        &discovery,
		IdentityProviderClientID:     new("client"),
		IdentityProviderClientSecret: new("secret"),
	}).Save(ctx)
	assert.NilError(t, err)

	org, err := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).Save(ctx)
	assert.NilError(t, err)

	// mark the connection tested, then enforce SSO; JIT provisioning is left at its enabled default
	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		SetIdentityProviderAuthTested(true).
		Exec(ctx))
	assert.NilError(t, suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetIdentityProviderLoginEnforced(true).
		Exec(ctx))

	req := httptest.NewRequest(http.MethodGet, "/v1/sso/login?organization_id="+org.ID, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)
	assert.Assert(t, is.Equal(http.StatusOK, rec.Code))

	var state, nonce *http.Cookie
	for _, c := range rec.Result().Cookies() {
		switch c.Name {
		case "state":
			state = c
		case "nonce":
			nonce = c
		}
	}
	assert.Assert(t, state != nil)
	assert.Assert(t, nonce != nil)

	oidc.nonce = nonce.Value

	cookieHeader := "state=" + state.Value + "; nonce=" + nonce.Value + "; organization_id=" + org.ID
	cbReq := httptest.NewRequest(http.MethodGet, "/v1/sso/callback?code=code123&state="+url.QueryEscape(state.Value)+"&organization_id="+org.ID, nil)
	cbReq.Header.Set("Cookie", cookieHeader)
	cbRec := httptest.NewRecorder()
	suite.e.ServeHTTP(cbRec, cbReq)
	assert.Assert(t, is.Equal(http.StatusOK, cbRec.Code))

	// the callback provisioned the membership just-in-time, as a member
	membership, err := suite.db.OrgMembership.Query().
		Where(orgmembership.HasUserWith(user.EmailEqualFold("jit@example.com")), orgmembership.OrganizationID(org.ID)).
		Only(ctx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(enums.RoleMember, membership.Role))
}
