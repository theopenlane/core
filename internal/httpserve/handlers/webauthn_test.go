package handlers_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	echo "github.com/theopenlane/echox"
	iamwebauthn "github.com/theopenlane/iam/providers/webauthn"
	"github.com/theopenlane/iam/sessions"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	coremodels "github.com/theopenlane/core/common/models"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/internal/ent/generated/webauthn"
	"github.com/theopenlane/iam/auth"
)

const (
	passkeyTestRPID   = "localhost"
	passkeyTestOrigin = "http://localhost"
)

type testPasskey struct {
	credentialID []byte
	key          *ecdsa.PrivateKey
}

func (suite *HandlerTestSuite) TestBeginWebauthnRegistration() {
	t := suite.T()

	suite.enableTestWebauthn(t)
	suite.registerTestHandler("POST", "registration/options", suite.h.BeginWebauthnRegistration)

	existingUser := suite.userBuilder(context.Background())
	newEmail := "passkey-" + strings.ToLower(ulids.New().String()) + "@theopenlane.io"

	testCases := []struct {
		name         string
		email        string
		ctx          context.Context
		expectedCode int
	}{
		{
			name:         "authenticated user registers own passkey",
			email:        existingUser.UserInfo.Email,
			ctx:          existingUser.UserCtx,
			expectedCode: http.StatusOK,
		},
		{
			name:         "authenticated user cannot register for another email",
			email:        newEmail,
			ctx:          existingUser.UserCtx,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "unauthenticated request for existing user is rejected",
			email:        existingUser.UserInfo.Email,
			ctx:          context.Background(),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "unauthenticated request for new email is rejected",
			email:        newEmail,
			ctx:          context.Background(),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:  "api token cannot register a passkey",
			email: existingUser.UserInfo.Email,
			ctx: auth.WithCaller(context.Background(), &auth.Caller{
				SubjectID:          existingUser.ID,
				OrganizationID:     existingUser.OrganizationID,
				AuthenticationType: auth.APITokenAuthentication,
			}),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:  "personal access token cannot register a passkey",
			email: existingUser.UserInfo.Email,
			ctx: auth.WithCaller(context.Background(), &auth.Caller{
				SubjectID:          existingUser.ID,
				OrganizationID:     existingUser.OrganizationID,
				AuthenticationType: auth.PATAuthentication,
			}),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:  "anonymous trust center caller cannot register a passkey",
			email: existingUser.UserInfo.Email,
			ctx: auth.WithCaller(context.Background(), &auth.Caller{
				SubjectID:          "anon_trustcenter_" + ulids.New().String(),
				OrganizationID:     existingUser.OrganizationID,
				OrganizationRole:   auth.AnonymousRole,
				AuthenticationType: auth.JWTAuthentication,
				Capabilities:       auth.CapTrustCenterAnonymous,
			}),
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(models.WebauthnRegistrationRequest{Email: tc.email, Name: "Passkey User"})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/registration/options", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			suite.e.ServeHTTP(recorder, req.WithContext(tc.ctx))

			require.Equal(t, tc.expectedCode, recorder.Code)

			if tc.expectedCode != http.StatusOK {
				return
			}

			var out models.WebauthnBeginRegistrationResponse
			require.NoError(t, json.NewDecoder(recorder.Body).Decode(&out))
			require.NotNil(t, out.CredentialCreation)
			require.NotEmpty(t, out.Session)
		})
	}

	internalCtx := auth.WithInternalOperationContext(context.Background())

	exists, err := suite.db.User.Query().Where(user.Email(newEmail)).Exist(internalCtx)
	require.NoError(t, err)
	require.False(t, exists)
}

func (suite *HandlerTestSuite) TestFinishWebauthnRegistration() {
	t := suite.T()

	suite.enableTestWebauthn(t)
	suite.registerTestHandler("POST", "registration/options", suite.h.BeginWebauthnRegistration)
	suite.registerTestHandler("POST", "registration/verification", suite.h.FinishWebauthnRegistration)

	testCases := []struct {
		name            string
		finishAsOther   bool
		finishLoggedOut bool
		expectedCode    int
	}{
		{
			name:         "authenticated user registers a passkey",
			expectedCode: http.StatusOK,
		},
		{
			name:          "another authenticated user cannot finish the registration",
			finishAsOther: true,
			expectedCode:  http.StatusUnauthorized,
		},
		{
			name:            "unauthenticated request cannot finish the registration",
			finishLoggedOut: true,
			expectedCode:    http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			owner := suite.userBuilderWithInput(context.Background(), &userInput{confirmedUser: true})

			finishCtx := owner.UserCtx
			if tc.finishAsOther {
				finishCtx = suite.userBuilderWithInput(context.Background(), &userInput{confirmedUser: true}).UserCtx
			}

			if tc.finishLoggedOut {
				finishCtx = context.Background()
			}

			rec := suite.registerPasskey(t, owner, finishCtx)

			assert.Equal(t, tc.expectedCode, rec.Code)

			count, err := suite.db.Webauthn.Query().Where(webauthn.OwnerID(owner.ID)).Count(owner.UserCtx)
			require.NoError(t, err)

			if tc.expectedCode != http.StatusOK {
				assert.Equal(t, 0, count)
				return
			}

			assert.Equal(t, 1, count)
			assert.NotContains(t, rec.Body.String(), "access_token")
		})
	}
}

func (suite *HandlerTestSuite) TestFinishWebauthnLogin() {
	t := suite.T()

	suite.enableTestWebauthn(t)
	suite.registerTestHandler("POST", "authentication/options", suite.h.BeginWebauthnLogin)
	suite.registerTestHandler("POST", "authentication/verification", suite.h.FinishWebauthnLogin)

	testCases := []struct {
		name           string
		confirmed      bool
		wrongKey       bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "confirmed email can log in with a passkey",
			confirmed:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unconfirmed email cannot log in with a passkey",
			confirmed:      false,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "verify your account",
		},
		{
			name:           "assertion signed with the wrong key is rejected",
			confirmed:      true,
			wrongKey:       true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := suite.userBuilderWithInput(context.Background(), &userInput{confirmedUser: tc.confirmed})
			passkey := suite.createTestPasskey(t, user)

			if tc.wrongKey {
				passkey.key = newTestPasskeyKey(t)
			}

			rec := suite.loginWithPasskey(t, user, passkey)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedError != "" {
				assert.Contains(t, rec.Body.String(), tc.expectedError)
			}
		})
	}
}

func (suite *HandlerTestSuite) TestHasValidSSOSession() {
	t := suite.T()
	ctx := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())

	user := suite.userBuilder(ctx.Request().Context())

	set := map[string]any{
		sessions.UserTypeKey: enums.AuthProviderOIDC.String(),
		sessions.UserIDKey:   user.ID,
	}

	c, err := suite.h.SessionConfig.SaveAndStoreSession(ctx.Request().Context(), ctx.Response().Writer, set, user.ID)
	require.NoError(t, err)

	token, err := sessions.SessionToken(c)
	require.NoError(t, err)

	ctx.Request().AddCookie(sessions.NewDevSessionCookie(token))

	ok := suite.h.HasValidSSOSession(ctx, user.ID)
	require.True(t, ok)
}

func (suite *HandlerTestSuite) TestHasValidSSOSessionInvalid() {
	t := suite.T()
	ctx := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), httptest.NewRecorder())

	user := suite.userBuilder(ctx.Request().Context())

	// no cookie set
	ok := suite.h.HasValidSSOSession(ctx, user.ID)
	require.False(t, ok)
}

func (suite *HandlerTestSuite) enableTestWebauthn(t *testing.T) {
	t.Helper()

	previous := suite.h.OauthProvider.Webauthn

	suite.h.OauthProvider.Webauthn.MaxDevices = 10
	suite.h.WebAuthn = iamwebauthn.NewWithConfig(iamwebauthn.ProviderConfig{
		Enabled:        true,
		DisplayName:    "Openlane",
		RelyingPartyID: passkeyTestRPID,
		RequestOrigins: []string{passkeyTestOrigin},
		Timeout:        time.Minute,
	})

	t.Cleanup(func() {
		suite.h.WebAuthn = nil
		suite.h.OauthProvider.Webauthn = previous
	})
}

func (suite *HandlerTestSuite) registerPasskey(t *testing.T, owner testUserDetails, finishCtx context.Context) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(models.WebauthnRegistrationRequest{Email: owner.UserInfo.Email})
	require.NoError(t, err)

	beginReq := httptest.NewRequest(http.MethodPost, "/registration/options", bytes.NewReader(body))
	beginReq.Header.Set("Content-Type", "application/json")

	beginRec := httptest.NewRecorder()
	suite.e.ServeHTTP(beginRec, beginReq.WithContext(owner.UserCtx))
	require.Equal(t, http.StatusOK, beginRec.Code)

	var begin models.WebauthnBeginRegistrationResponse
	require.NoError(t, json.NewDecoder(beginRec.Body).Decode(&begin))
	require.NotNil(t, begin.CredentialCreation)

	clientData, err := json.Marshal(map[string]any{
		"type":      "webauthn.create",
		"challenge": begin.Response.Challenge.String(),
		"origin":    passkeyTestOrigin,
	})
	require.NoError(t, err)

	credentialID := []byte(ulids.New().String())

	attestation, err := webauthncbor.Marshal(map[string]any{
		"fmt":      "none",
		"attStmt":  map[string]any{},
		"authData": passkeyAttestedAuthenticatorData(t, credentialID, newTestPasskeyKey(t)),
	})
	require.NoError(t, err)

	finish := models.WebauthnRegistrationFinishRequest{
		ID:    base64.RawURLEncoding.EncodeToString(credentialID),
		RawID: base64.RawURLEncoding.EncodeToString(credentialID),
		Type:  "public-key",
	}
	finish.Response.AttestationObject = base64.RawURLEncoding.EncodeToString(attestation)
	finish.Response.ClientDataJSON = base64.RawURLEncoding.EncodeToString(clientData)

	finishBody, err := json.Marshal(finish)
	require.NoError(t, err)

	finishReq := httptest.NewRequest(http.MethodPost, "/registration/verification", bytes.NewReader(finishBody))
	finishReq.Header.Set("Content-Type", "application/json")

	for _, cookie := range beginRec.Result().Cookies() {
		finishReq.AddCookie(cookie)
	}

	finishRec := httptest.NewRecorder()
	suite.e.ServeHTTP(finishRec, finishReq.WithContext(finishCtx))

	return finishRec
}

func (suite *HandlerTestSuite) createTestPasskey(t *testing.T, user testUserDetails) testPasskey {
	t.Helper()

	key := newTestPasskeyKey(t)
	credentialID := []byte(ulids.New().String())

	err := suite.db.Webauthn.Create().
		SetOwnerID(user.ID).
		SetCredentialID(credentialID).
		SetPublicKey(passkeyPublicKey(t, key)).
		SetAttestationType("none").
		SetAaguid(coremodels.ToAAGUID(make([]byte, 16))).
		SetSignCount(0).
		SetTransports([]string{"internal"}).
		Exec(user.UserCtx)
	require.NoError(t, err)

	return testPasskey{credentialID: credentialID, key: key}
}

func (suite *HandlerTestSuite) loginWithPasskey(t *testing.T, user testUserDetails, passkey testPasskey) *httptest.ResponseRecorder {
	t.Helper()

	beginReq := httptest.NewRequest(http.MethodPost, "/authentication/options", bytes.NewReader([]byte(`{}`)))
	beginReq.Header.Set("Content-Type", "application/json")

	beginRec := httptest.NewRecorder()
	suite.e.ServeHTTP(beginRec, beginReq)
	require.Equal(t, http.StatusOK, beginRec.Code)

	var begin models.WebauthnBeginLoginResponse
	require.NoError(t, json.NewDecoder(beginRec.Body).Decode(&begin))
	require.NotNil(t, begin.CredentialAssertion)

	clientData, err := json.Marshal(map[string]any{
		"type":      "webauthn.get",
		"challenge": begin.Response.Challenge.String(),
		"origin":    passkeyTestOrigin,
	})
	require.NoError(t, err)

	authData := passkeyAuthenticatorData()
	clientDataHash := sha256.Sum256(clientData)
	digest := sha256.Sum256(append(append([]byte{}, authData...), clientDataHash[:]...))

	signature, err := ecdsa.SignASN1(rand.Reader, passkey.key, digest[:])
	require.NoError(t, err)

	finish := models.WebauthnLoginFinishRequest{
		ID:    base64.RawURLEncoding.EncodeToString(passkey.credentialID),
		RawID: base64.RawURLEncoding.EncodeToString(passkey.credentialID),
		Type:  "public-key",
	}
	finish.Response.AuthenticatorData = base64.RawURLEncoding.EncodeToString(authData)
	finish.Response.ClientDataJSON = base64.RawURLEncoding.EncodeToString(clientData)
	finish.Response.Signature = base64.RawURLEncoding.EncodeToString(signature)
	finish.Response.UserHandle = base64.RawURLEncoding.EncodeToString([]byte(user.ID))

	body, err := json.Marshal(finish)
	require.NoError(t, err)

	finishReq := httptest.NewRequest(http.MethodPost, "/authentication/verification", bytes.NewReader(body))
	finishReq.Header.Set("Content-Type", "application/json")

	for _, cookie := range beginRec.Result().Cookies() {
		finishReq.AddCookie(cookie)
	}

	finishRec := httptest.NewRecorder()
	suite.e.ServeHTTP(finishRec, finishReq)

	return finishRec
}

func newTestPasskeyKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	return key
}

func passkeyPublicKey(t *testing.T, key *ecdsa.PrivateKey) []byte {
	t.Helper()

	point, err := key.PublicKey.Bytes()
	require.NoError(t, err)

	publicKey, err := webauthncbor.Marshal(webauthncose.EC2PublicKeyData{
		PublicKeyData: webauthncose.PublicKeyData{
			KeyType:   int64(webauthncose.EllipticKey),
			Algorithm: int64(webauthncose.AlgES256),
		},
		Curve:  int64(webauthncose.P256),
		XCoord: point[1:33],
		YCoord: point[33:65],
	})
	require.NoError(t, err)

	return publicKey
}

func passkeyAuthenticatorData() []byte {
	rpIDHash := sha256.Sum256([]byte(passkeyTestRPID))

	authData := append([]byte{}, rpIDHash[:]...)
	authData = append(authData, byte(protocol.FlagUserPresent|protocol.FlagUserVerified))

	return append(authData, 0, 0, 0, 0)
}

func passkeyAttestedAuthenticatorData(t *testing.T, credentialID []byte, key *ecdsa.PrivateKey) []byte {
	t.Helper()

	rpIDHash := sha256.Sum256([]byte(passkeyTestRPID))

	authData := append([]byte{}, rpIDHash[:]...)
	authData = append(authData, byte(protocol.FlagUserPresent|protocol.FlagUserVerified|protocol.FlagAttestedCredentialData))
	authData = append(authData, 0, 0, 0, 0)
	authData = append(authData, make([]byte, 16)...)
	authData = binary.BigEndian.AppendUint16(authData, uint16(len(credentialID))) //nolint:gosec

	authData = append(authData, credentialID...)

	return append(authData, passkeyPublicKey(t, key)...)
}
