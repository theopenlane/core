package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/tokens"
)

// TestRefreshTokenDiagnosticsExtractsSafeMetadata verifies refresh failure logs include safe JWT metadata
func TestRefreshTokenDiagnosticsExtractsSafeMetadata(t *testing.T) {
	now := time.Date(2026, 4, 26, 13, 46, 15, 0, time.UTC)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": float64(now.Add(time.Hour).Unix()),
		"iat": float64(now.Unix()),
		"nbf": float64(now.Add(time.Minute).Unix()),
	})
	token.Header["kid"] = "test-kid"

	tokenString, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	diag := refreshTokenDiagnostics(tokenString, tokens.ErrTokenSignatureInvalid)

	require.Equal(t, "signature_invalid", diag.category)
	require.Equal(t, "test-kid", diag.kid)
	require.Equal(t, "HS256", diag.alg)
	require.Empty(t, diag.parseError)
	require.Equal(t, "2026-04-26T14:46:15Z", diag.exp)
	require.Equal(t, "2026-04-26T13:46:15Z", diag.iat)
	require.Equal(t, "2026-04-26T13:47:15Z", diag.nbf)
}

// TestRefreshTokenErrorCategoryMatchesWrappedJWTMessages verifies common JWT error text is categorized
func TestRefreshTokenErrorCategoryMatchesWrappedJWTMessages(t *testing.T) {
	require.Equal(t, "signature_invalid", refreshTokenErrorCategory(errors.New("token signature is invalid: ed25519: verification error")))
	require.Equal(t, "not_valid_yet", refreshTokenErrorCategory(errors.New("token has invalid claims: token is expired, token is not valid yet")))
	require.Equal(t, "unknown_signing_key", refreshTokenErrorCategory(errors.New("token is unverifiable: error while executing keyfunc: unknown signing key")))
}
