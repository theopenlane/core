package keystore

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GCPWorkloadIdentityIssuer issues subject tokens for Google Workload Identity Federation.
type hushSecretLoader interface {
	LoadHushSecret(ctx context.Context, ownerID, name string) (string, error)
}

// GCPWorkloadIdentityIssuer issues subject tokens for Google Workload Identity Federation.
type GCPWorkloadIdentityIssuer struct {
	loader hushSecretLoader
	now    func() time.Time
}

// NewGCPWorkloadIdentityIssuer returns a WorkloadIdentityIssuer backed by the integration store.
func NewGCPWorkloadIdentityIssuer(store *Store) *GCPWorkloadIdentityIssuer {
	return &GCPWorkloadIdentityIssuer{
		loader: store,
		now:    time.Now,
	}
}

// IssueSubjectToken produces a signed JWT suitable for use as a Workload Identity Federation subject token.
func (i *GCPWorkloadIdentityIssuer) IssueSubjectToken(ctx context.Context, orgID string, spec *ProviderSpec, attrs map[string]string) (*WorkloadIdentitySubjectToken, error) {
	if i.loader == nil {
		return nil, fmt.Errorf("gcp workload identity issuer: store not configured")
	}

	provider := strings.TrimSpace(attrs["workloadIdentityProvider"])
	if provider == "" {
		return nil, fmt.Errorf("gcp workload identity issuer: workloadIdentityProvider metadata required")
	}

	credentialsJSON, err := i.resolveCredentialJSON(ctx, orgID, spec, attrs)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(credentialsJSON) == "" {
		return nil, fmt.Errorf("gcp workload identity issuer: workload identity credentials missing")
	}

	key, err := parseServiceAccountKey(credentialsJSON)
	if err != nil {
		return nil, err
	}

	now := i.now()
	claims := jwt.MapClaims{
		"iss": key.ClientEmail,
		"sub": key.ClientEmail,
		"aud": provider,
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
	}

	if project := strings.TrimSpace(attrs["workloadPoolProject"]); project != "" {
		claims["google"] = map[string]any{
			"workforce_pool_user_project": project,
		}
	}

	if audience := strings.TrimSpace(attrs["audience"]); audience != "" {
		claims["target_audience"] = audience
	} else if spec != nil && spec.WorkloadIdentity != nil && spec.WorkloadIdentity.Audience != "" {
		claims["target_audience"] = spec.WorkloadIdentity.Audience
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if key.PrivateKeyID != "" {
		token.Header["kid"] = key.PrivateKeyID
	}

	signed, err := token.SignedString(key.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("gcp workload identity issuer: sign subject token: %w", err)
	}

	return &WorkloadIdentitySubjectToken{
		Token: signed,
		Type:  "urn:ietf:params:oauth:token-type:jwt",
	}, nil
}

func (i *GCPWorkloadIdentityIssuer) resolveCredentialJSON(ctx context.Context, orgID string, spec *ProviderSpec, attrs map[string]string) (string, error) {
	if raw := strings.TrimSpace(attrs["credentialsJSON"]); raw != "" {
		return raw, nil
	}

	if ref := strings.TrimSpace(attrs["credentialsRef"]); ref != "" {
		name := normalizeSecretName(ref)
		value, err := i.loader.LoadHushSecret(ctx, orgID, name)
		if err != nil {
			return "", fmt.Errorf("gcp workload identity issuer: load credentials secret %q: %w", name, err)
		}
		return value, nil
	}

	// Fall back to the provider-derived secret name if metadata added it automatically.
	if spec != nil && spec.Name != "" {
		name := NewHelper(spec.Name, "").SecretName("credentials")
		if value, err := i.loader.LoadHushSecret(ctx, orgID, name); err == nil && strings.TrimSpace(value) != "" {
			return value, nil
		}
	}

	return "", nil
}

type serviceAccountKey struct {
	PrivateKey   *rsa.PrivateKey
	PrivateKeyID string
	ClientEmail  string
}

func parseServiceAccountKey(raw string) (*serviceAccountKey, error) {
	var payload struct {
		PrivateKey   string `json:"private_key"`
		PrivateKeyID string `json:"private_key_id"`
		ClientEmail  string `json:"client_email"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("gcp workload identity issuer: decode credentials json: %w", err)
	}
	if payload.PrivateKey == "" || payload.ClientEmail == "" {
		return nil, errors.New("gcp workload identity issuer: credentials missing private_key or client_email")
	}

	block, _ := pem.Decode([]byte(payload.PrivateKey))
	if block == nil {
		return nil, errors.New("gcp workload identity issuer: invalid private key encoding")
	}

	var parsed any
	var err error
	if parsed, err = x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
		if rsaKey, errPKCS1 := x509.ParsePKCS1PrivateKey(block.Bytes); errPKCS1 == nil {
			parsed = rsaKey
		} else {
			return nil, fmt.Errorf("gcp workload identity issuer: parse private key: %w", err)
		}
	}

	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("gcp workload identity issuer: private key must be RSA")
	}

	return &serviceAccountKey{
		PrivateKey:   rsaKey,
		PrivateKeyID: strings.TrimSpace(payload.PrivateKeyID),
		ClientEmail:  strings.TrimSpace(payload.ClientEmail),
	}, nil
}

func normalizeSecretName(ref string) string {
	trimmed := strings.TrimSpace(ref)
	trimmed = strings.TrimPrefix(trimmed, "hush://")
	trimmed = strings.TrimPrefix(trimmed, "secret://")
	if trimmed == "" {
		return ""
	}
	if strings.Contains(trimmed, "/") {
		parts := strings.Split(trimmed, "/")
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return trimmed
}
