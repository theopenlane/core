package oidc

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/tokens"
)

// TestDiscoveryDocument verifies the discovery document external providers resolve
func TestDiscoveryDocument(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, testRSAKeySize)
	require.NoError(t, err)

	manager, err := tokens.NewWithKey(key, tokens.NewConfig(
		tokens.WithIssuer(testIssuer),
		tokens.WithAudience(testIssuer),
		func(c *tokens.Config) { c.JWKSEndpoint = testJWKSEndpoint },
	))
	require.NoError(t, err)

	doc := DiscoveryDocument(manager)

	assert.Equal(t, testIssuer, doc.Issuer)
	assert.Equal(t, testJWKSEndpoint, doc.JwksURI)
	assert.Equal(t, []string{"id_token"}, doc.ResponseTypesSupported)
	assert.Equal(t, []string{"public"}, doc.SubjectTypesSupported)
	assert.Equal(t, []string{jwt.SigningMethodRS256.Alg()}, doc.IDTokenSigningAlgValuesSupported)
}
