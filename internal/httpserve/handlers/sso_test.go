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
	"github.com/stretchr/testify/assert"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/zitadel/oidc/v3/pkg/client"
	oidccrypto "github.com/zitadel/oidc/v3/pkg/crypto"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

func (suite *HandlerTestSuite) TestWebfingerHandler() {
	t := suite.T()

	// Create operation for WebfingerHandler
	webfingerOp := suite.createImpersonationOperation("WebfingerHandler", "Webfinger handler")
	suite.registerTestHandler("GET", ".well-known/webfinger", webfingerOp, suite.h.WebfingerHandler)

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: lo.ToPtr(true),
		IdentityProvider:              lo.ToPtr(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:         lo.ToPtr("http://example.com"),
		MultifactorAuthEnforced:       lo.ToPtr(true),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.db.UserSetting.Update().Where(usersetting.UserID(testUser1.ID)).SetDefaultOrgID(org.ID).ExecX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=org:"+org.ID, nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var out models.SSOStatusReply
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	log.Error().Err(errors.New("output")).Interface("out", out).Msg("WebfingerHandler output")
	assert.True(t, out.Enforced)
	assert.True(t, out.OrgTFAEnforced)
	assert.Equal(t, org.ID, out.OrganizationID)

	emailReq := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=acct:"+testUser1.UserInfo.Email, nil)
	emailRec := httptest.NewRecorder()
	suite.e.ServeHTTP(emailRec, emailReq)
	assert.Equal(t, http.StatusOK, emailRec.Code)
	var emailOut models.SSOStatusReply
	assert.NoError(t, json.NewDecoder(emailRec.Body).Decode(&emailOut))
	assert.True(t, emailOut.Enforced)
	assert.True(t, emailOut.OrgTFAEnforced)
	assert.Equal(t, org.ID, emailOut.OrganizationID)
}

func (suite *HandlerTestSuite) TestWebfingerHandlerTFAOnly() {
	t := suite.T()

	// Create operation for WebfingerHandler
	webfingerOp := suite.createImpersonationOperation("WebfingerHandler", "Webfinger handler")
	suite.registerTestHandler("GET", ".well-known/webfinger", webfingerOp, suite.h.WebfingerHandler)

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	// Create organization setting with only TFA enforced, not SSO
	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		MultifactorAuthEnforced: lo.ToPtr(true),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.db.UserSetting.Update().Where(usersetting.UserID(testUser1.ID)).SetDefaultOrgID(org.ID).ExecX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=org:"+org.ID, nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var out models.SSOStatusReply
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	assert.False(t, out.Enforced)      // SSO not enforced
	assert.True(t, out.OrgTFAEnforced) // TFA enforced
	assert.Equal(t, org.ID, out.OrganizationID)
}

func (suite *HandlerTestSuite) TestWebfingerHandlerNotFound() {
	// Create operation for WebfingerHandler
	webfingerOp := suite.createImpersonationOperation("WebfingerHandler", "Webfinger handler")
	suite.registerTestHandler("GET", ".well-known/webfinger", webfingerOp, suite.h.WebfingerHandler)

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
	assert.NoError(t, err)

	der := x509.MarshalPKCS1PrivateKey(priv)
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})

	signer, err := client.NewSignerFromPrivateKeyByte(pemKey, "test-kid")
	assert.NoError(t, err)

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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
		claimsMap := make(map[string]interface{})
		b, _ := json.Marshal(claims)
		_ = json.Unmarshal(b, &claimsMap)
		claimsMap["sub"] = "1234"
		// Inject email and email_verified claims for compatibility with callback handler
		claimsMap["email"] = m.email
		claimsMap["email_verified"] = true

		info := &oidc.UserInfo{
			Subject: "1234",
			UserInfoProfile: oidc.UserInfoProfile{
				Name:    m.name,
				Picture: m.picture,
			},
			UserInfoEmail: oidc.UserInfoEmail{
				Email:         m.email,
				EmailVerified: true,
			},
		}
		claims.SetUserInfo(info)

		// Re-marshal claimsMap to JSON and sign as JWT
		claimsBytes, _ := json.Marshal(claimsMap)
		var claimsInterface map[string]interface{}
		_ = json.Unmarshal(claimsBytes, &claimsInterface)
		raw, err := oidccrypto.Sign(claimsInterface, m.signer)
		assert.NoError(t, err)

		resp := map[string]interface{}{
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

func (suite *HandlerTestSuite) TestSSOLoginAndCallback() {
	t := suite.T()

	// Create operations for SSO handlers
	ssoLoginOp := suite.createImpersonationOperation("SSOLoginHandler", "SSO login handler")
	ssoCallbackOp := suite.createImpersonationOperation("SSOCallbackHandler", "SSO callback handler")
	suite.registerTestHandler("GET", "v1/sso/login", ssoLoginOp, suite.h.SSOLoginHandler)
	suite.registerTestHandler("GET", "v1/sso/callback", ssoCallbackOp, suite.h.SSOCallbackHandler)

	oidc := newMockOIDCServer(t,
		withExpectedCode("code123"),
		withClientSecret("secret"),
		withUserInfo("sso@example.com", "SSO User", "https://example.com/avatar.png"),
	)
	defer oidc.Close()

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
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
		IdentityProviderLoginEnforced: lo.ToPtr(true),
		IdentityProvider:              lo.ToPtr(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:         &discovery,
		IdentityProviderClientID:      lo.ToPtr("client"),
		IdentityProviderClientSecret:  lo.ToPtr("secret"),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      ulids.New().String(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.db.OrganizationSetting.UpdateOneID(setting.ID).
		SetOrganizationID(org.ID).
		ExecX(ctx)

	ctxTargetOrg := auth.NewTestContextWithOrgID(ssoUser.ID, org.ID)
	ctxTargetOrg = privacy.DecisionContext(ctxTargetOrg, privacy.Allow)
	testUserCtx := ent.NewContext(ctxTargetOrg, suite.db)

	suite.db.OrgMembership.Create().SetInput(generated.CreateOrgMembershipInput{
		OrganizationID: org.ID,
		UserID:         ssoUser.ID,
		Role:           &enums.RoleMember,
	}).ExecX(testUserCtx)

	suite.db.UserSetting.Update().Where(usersetting.UserID(ssoUser.ID)).SetDefaultOrgID(org.ID).ExecX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/sso/login?organization_id="+org.ID, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

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

	assert.NotNil(t, state)
	assert.NotNil(t, nonce)

	oidc.nonce = nonce.Value

	// Set cookies as a single Cookie header instead of AddCookie
	cookieHeader := "state=" + state.Value + "; nonce=" + nonce.Value + "; organization_id=" + org.ID
	cbReq := httptest.NewRequest(http.MethodGet, "/v1/sso/callback?code=code123&state="+url.QueryEscape(state.Value)+"&organization_id="+org.ID, nil)
	cbReq.Header.Set("Cookie", cookieHeader)
	cbRec := httptest.NewRecorder()

	suite.e.ServeHTTP(cbRec, cbReq)
	assert.Equal(t, http.StatusOK, cbRec.Code)
	var out models.LoginReply
	assert.NoError(t, json.NewDecoder(cbRec.Body).Decode(&out))
	assert.True(t, out.Success)
}
