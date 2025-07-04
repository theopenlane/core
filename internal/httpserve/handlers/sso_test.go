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
	"github.com/stretchr/testify/require"

	"github.com/zitadel/oidc/pkg/client"
	oidccrypto "github.com/zitadel/oidc/pkg/crypto"
	"github.com/zitadel/oidc/pkg/oidc"
	jose "gopkg.in/square/go-jose.v2"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/generated"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

func (suite *HandlerTestSuite) TestWebfingerHandler() {
	t := suite.T()

	suite.e.GET(".well-known/webfinger", suite.h.WebfingerHandler)

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: lo.ToPtr(true),
		IdentityProvider:              lo.ToPtr(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:         lo.ToPtr("http://example.com"),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.db.UserSetting.Update().Where(usersetting.UserID(testUser1.ID)).SetDefaultOrgID(org.ID).ExecX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=org:"+org.ID, nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out models.SSOStatusReply
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&out))
	log.Error().Err(errors.New("output")).Interface("out", out).Msg("WebfingerHandler output")
	assert.True(t, out.Enforced)
	assert.Equal(t, org.ID, out.OrganizationID)

	emailReq := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=acct:"+testUser1.UserInfo.Email, nil)
	emailRec := httptest.NewRecorder()
	suite.e.ServeHTTP(emailRec, emailReq)
	require.Equal(t, http.StatusOK, emailRec.Code)
	var emailOut models.SSOStatusReply
	require.NoError(t, json.NewDecoder(emailRec.Body).Decode(&emailOut))
	assert.True(t, emailOut.Enforced)
	assert.Equal(t, org.ID, emailOut.OrganizationID)
}

func (suite *HandlerTestSuite) TestWebfingerHandlerNotFound() {
	suite.e.GET(".well-known/webfinger", suite.h.WebfingerHandler)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/webfinger?resource=acct:"+gofakeit.Email(), nil)
	rec := httptest.NewRecorder()

	suite.e.ServeHTTP(rec, req)
}

// mockOIDCServer is a minimal OIDC provider used for testing the SSO flow.
// It exposes discovery, token and JWK endpoints. The ID tokens are signed
// with an in-memory RSA key.
type mockOIDCServer struct {
	server       *httptest.Server
	signer       jose.Signer
	privKey      *rsa.PrivateKey
	keyID        string
	expectedCode string
	nonce        string
	email        string
	name         string
	picture      string
}

type oidcServerOption func(*mockOIDCServer)

func withExpectedCode(code string) oidcServerOption {
	return func(m *mockOIDCServer) { m.expectedCode = code }
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
	require.NoError(t, err)

	der := x509.MarshalPKCS1PrivateKey(priv)
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})

	signer, err := client.NewSignerFromPrivateKeyByte(pemKey, "test-kid")
	require.NoError(t, err)

	m := &mockOIDCServer{
		keyID:   "test-kid",
		signer:  signer,
		privKey: priv,
		email:   "sso@example.com",
		name:    "SSO User",
		picture: "https://example.com/avatar.png",
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
		_ = r.ParseForm()
		if m.expectedCode != "" && r.Form.Get("code") != m.expectedCode {
			http.Error(w, "invalid code", http.StatusBadRequest)
			return
		}

		claims := oidc.NewIDTokenClaims(
			m.server.URL,
			"1234",
			[]string{r.Form.Get("client_id")},
			time.Now().Add(time.Hour),
			time.Now(),
			m.nonce,
			"",
			nil,
			r.Form.Get("client_id"),
			0,
		)

		info := oidc.NewUserInfo()
		info.SetEmail(m.email, true)
		info.SetName(m.name)
		info.SetPicture(m.picture)
		claims.SetUserinfo(info)

		raw, err := oidccrypto.Sign(claims, m.signer)
		require.NoError(t, err)

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

	suite.e.GET("v1/sso/login", suite.h.SSOLoginHandler)
	suite.e.GET("v1/sso/callback", suite.h.SSOCallbackHandler)

	oidc := newMockOIDCServer(t,
		withExpectedCode("code123"),
		withUserInfo("sso@example.com", "SSO User", "https://example.com/avatar.png"),
	)
	defer oidc.Close()

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)
	ctx = ent.NewContext(ctx, suite.db)

	discovery := oidc.server.URL + "/.well-known/openid-configuration"
	setting := suite.db.OrganizationSetting.Create().SetInput(generated.CreateOrganizationSettingInput{
		IdentityProviderLoginEnforced: lo.ToPtr(true),
		IdentityProvider:              lo.ToPtr(enums.SSOProviderOkta),
		OidcDiscoveryEndpoint:         &discovery,
		IdentityProviderClientID:      lo.ToPtr("client"),
		IdentityProviderClientSecret:  lo.ToPtr("secret"),
	}).SaveX(ctx)

	org := suite.db.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name:      gofakeit.Name(),
		SettingID: &setting.ID,
	}).SaveX(ctx)

	suite.db.UserSetting.Update().Where(usersetting.UserID(testUser1.ID)).SetDefaultOrgID(org.ID).ExecX(ctx)

	req := httptest.NewRequest(http.MethodGet, "/v1/sso/login?organization_id="+org.ID, nil)
	rec := httptest.NewRecorder()
	suite.e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	loc, err := rec.Result().Location()
	require.NoError(t, err)
	assert.Contains(t, loc.String(), oidc.server.URL+"/auth")

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

	require.NotNil(t, state)
	require.NotNil(t, nonce)

	oidc.nonce = nonce.Value

	cbReq := httptest.NewRequest(http.MethodGet, "/v1/sso/callback?code=code123&state="+url.QueryEscape(state.Value)+"&organization_id="+org.ID, nil)
	cbReq.AddCookie(state)
	cbReq.AddCookie(nonce)
	cbReq.AddCookie(&http.Cookie{Name: "organization_id", Value: org.ID})
	cbRec := httptest.NewRecorder()

	suite.e.ServeHTTP(cbRec, cbReq)

	require.Equal(t, http.StatusOK, cbRec.Code)
	var out models.LoginReply
	require.NoError(t, json.NewDecoder(cbRec.Body).Decode(&out))
	assert.True(t, out.Success)
}
