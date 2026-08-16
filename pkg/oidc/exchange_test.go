package oidc

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/theopenlane/iam/tokens"
)

const (
	// testIssuer is the platform issuer external identity providers federate against
	testIssuer = "https://api.theopenlane.io"
	// testJWKSEndpoint is the JWKS location advertised by the discovery document
	testJWKSEndpoint = "https://api.theopenlane.io/.well-known/jwks.json"
	// testAudience is a representative GCP workload identity pool provider resource name
	testAudience = "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/openlane/providers/openlane"
	// testOrganizationID is the organization the assertions are requested for
	testOrganizationID = "01JQZX9K8N7M6P5R4T3V2W1Y0Z"
	// testScope is the OAuth scope requested during exchanges
	testScope = "https://www.googleapis.com/auth/cloud-platform"
	// testRSAKeySize is the smallest RSA key size the signing key loader accepts
	testRSAKeySize = 2048
)

// testTokenManager builds a token manager signing with the supplied key
func testTokenManager(t *testing.T, key crypto.Signer) *tokens.TokenManager {
	t.Helper()

	manager, err := tokens.NewWithKey(key, tokens.NewConfig(
		tokens.WithIssuer(testIssuer),
		tokens.WithAudience(testIssuer),
	))
	require.NoError(t, err)

	return manager
}

// parseAssertion verifies the assertion signature and returns its claims
func parseAssertion(t *testing.T, signed string, key crypto.Signer) jwt.MapClaims {
	t.Helper()

	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(signed, claims, func(*jwt.Token) (any, error) {
		return key.Public(), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	require.NoError(t, err)
	require.True(t, token.Valid)

	return claims
}

// testTokenSource builds a federation token source pointed at the supplied STS endpoint
func testTokenSource(t *testing.T, key *rsa.PrivateKey, endpoint string) oauth2.TokenSource {
	t.Helper()

	source, err := NewTokenSource(context.Background(), FederationSource{
		Manager:        testTokenManager(t, key),
		OrganizationID: testOrganizationID,
		Audience:       testAudience,
		Scopes:         []string{testScope},
		Endpoint:       endpoint,
	})
	require.NoError(t, err)

	return source
}

// TestNewTokenSource verifies the token source constructor validation
func TestNewTokenSource(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, testRSAKeySize)
	require.NoError(t, err)

	t.Run("requires an organization", func(t *testing.T) {
		_, err := NewTokenSource(context.Background(), FederationSource{Audience: testAudience})
		require.ErrorIs(t, err, ErrOrganizationIDRequired)
	})

	t.Run("requires an audience", func(t *testing.T) {
		_, err := NewTokenSource(context.Background(), FederationSource{OrganizationID: testOrganizationID})
		require.ErrorIs(t, err, ErrAudienceRequired)
	})

	t.Run("requires an endpoint", func(t *testing.T) {
		_, err := NewTokenSource(context.Background(), FederationSource{
			Manager:        testTokenManager(t, key),
			OrganizationID: testOrganizationID,
			Audience:       testAudience,
		})
		require.ErrorIs(t, err, ErrEndpointRequired)
	})

	t.Run("requires an RS256 signing key at build time", func(t *testing.T) {
		_, edKey, keyErr := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, keyErr)

		_, err := NewTokenSource(context.Background(), FederationSource{
			Manager:        testTokenManager(t, edKey),
			OrganizationID: testOrganizationID,
			Audience:       testAudience,
			Endpoint:       "https://sts.example.com/v1/token",
		})
		require.ErrorIs(t, err, tokens.ErrSigningAlgorithmMismatch)
	})
}

// TestTokenExchange verifies the RFC 8693 request sent and the response parsed
func TestTokenExchange(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, testRSAKeySize)
	require.NoError(t, err)

	t.Run("sends the exchange request and parses the granted token", func(t *testing.T) {
		var form url.Values

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, r.ParseForm())

			form = r.Form

			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{
				"access_token": "ya29.test-granted-token",
				"issued_token_type": "urn:ietf:params:oauth:token-type:access_token",
				"token_type": "Bearer",
				"expires_in": 3600
			}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		token, err := testTokenSource(t, key, server.URL).Token()
		require.NoError(t, err)

		assert.Equal(t, "ya29.test-granted-token", token.AccessToken)
		assert.Equal(t, "Bearer", token.TokenType)
		assert.WithinDuration(t, time.Now().Add(time.Hour), token.Expiry, time.Minute)

		// the exchange identifies the relying party with the assertion as subject token
		assert.Equal(t, "urn:ietf:params:oauth:grant-type:token-exchange", form.Get("grant_type"))
		assert.Equal(t, "urn:ietf:params:oauth:token-type:jwt", form.Get("subject_token_type"))
		assert.Equal(t, "urn:ietf:params:oauth:token-type:access_token", form.Get("requested_token_type"))
		assert.Equal(t, testAudience, form.Get("audience"))
		assert.Equal(t, testScope, form.Get("scope"))

		// pins zitadel's serialization of absent actor fields as empty parameters
		assert.Contains(t, form, "actor_token")
		assert.Empty(t, form.Get("actor_token"))

		// the assertion carries server-derived identity claims, verifiable with the issuer's public key
		claims := parseAssertion(t, form.Get("subject_token"), key)
		assert.Equal(t, testIssuer, claims["iss"])
		assert.Equal(t, testOrganizationID, claims["sub"])
		assert.Equal(t, testAudience, claims["aud"])
		assert.Equal(t, testOrganizationID, claims["organization_id"])

		expiry, err := claims.GetExpirationTime()
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now().Add(defaultAssertionLifetime), expiry.Time, time.Minute)
	})

	t.Run("returns the exchange error when the STS rejects the assertion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(`{"error":"invalid_request","error_description":"invalid audience"}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		_, err := testTokenSource(t, key, server.URL).Token()
		require.ErrorIs(t, err, ErrExchangeFailed)
		require.ErrorContains(t, err, "invalid audience")
	})
}
