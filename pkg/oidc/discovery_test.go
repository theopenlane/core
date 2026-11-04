package oidc

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/tokens"
)

// TestDiscoveryDocument verifies the discovery document external providers resolve
func TestDiscoveryDocument(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, testRSAKeySize)
	assert.NilError(t, err)

	manager, err := tokens.NewWithKey(key, tokens.NewConfig(
		tokens.WithIssuer(testIssuer),
		tokens.WithAudience(testIssuer),
		func(c *tokens.Config) { c.JWKSEndpoint = testJWKSEndpoint },
	))
	assert.NilError(t, err)

	doc := DiscoveryDocument(manager)

	assert.Check(t, is.Equal(testIssuer, doc.Issuer))
	assert.Check(t, is.Equal(testJWKSEndpoint, doc.JwksURI))
	assert.Check(t, is.DeepEqual([]string{"id_token"}, doc.ResponseTypesSupported))
	assert.Check(t, is.DeepEqual([]string{"public"}, doc.SubjectTypesSupported))
	assert.Check(t, is.DeepEqual([]string{jwt.SigningMethodRS256.Alg()}, doc.IDTokenSigningAlgValuesSupported))
}
