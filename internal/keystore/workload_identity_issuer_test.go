package keystore

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type mockSecretLoader struct {
	values map[string]string
}

func (m mockSecretLoader) LoadHushSecret(_ context.Context, _ string, name string) (string, error) {
	if v, ok := m.values[name]; ok {
		return v, nil
	}
	return "", fmt.Errorf("secret %s not found", name)
}

func TestGCPWorkloadIdentityIssuer_IssueSubjectToken(t *testing.T) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	pemKey := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	keyJSON, err := json.Marshal(map[string]any{
		"private_key":    string(pemKey),
		"private_key_id": "test-key",
		"client_email":   "svc@example.iam.gserviceaccount.com",
	})
	require.NoError(t, err)

	issuer := &GCPWorkloadIdentityIssuer{
		loader: mockSecretLoader{values: map[string]string{
			"gcp_scc_credentials": string(keyJSON),
		}},
		now: func() time.Time { return time.Unix(1_700_000_000, 0).UTC() },
	}

	spec := &ProviderSpec{Name: "gcp_scc"}
	token, err := issuer.IssueSubjectToken(context.Background(), "org-1", spec, map[string]string{
		"audience":                 "//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/pool/providers/provider",
		"workloadIdentityProvider": "//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/pool/providers/provider",
		"credentialsRef":           "gcp_scc_credentials",
	})
	require.NoError(t, err)
	require.Equal(t, "urn:ietf:params:oauth:token-type:jwt", token.Type)
	require.NotEmpty(t, token.Token)

	parsed, err := jwt.Parse(token.Token, func(t *jwt.Token) (any, error) {
		return &privateKey.PublicKey, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithoutClaimsValidation())
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	require.Equal(t, "//iam.googleapis.com/projects/123/locations/global/workloadIdentityPools/pool/providers/provider", claims["aud"])
	require.Equal(t, "svc@example.iam.gserviceaccount.com", claims["iss"])
	require.Equal(t, "svc@example.iam.gserviceaccount.com", claims["sub"])
}

func TestGCPWorkloadIdentityIssuer_MissingProvider(t *testing.T) {
	issuer := &GCPWorkloadIdentityIssuer{
		loader: mockSecretLoader{},
		now:    time.Now,
	}

	_, err := issuer.IssueSubjectToken(context.Background(), "org-1", &ProviderSpec{Name: "gcp_scc"}, map[string]string{})
	require.Error(t, err)
}
