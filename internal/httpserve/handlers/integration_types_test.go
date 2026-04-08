package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigureIntegrationRequestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload ConfigureIntegrationRequest
		wantErr bool
	}{
		{
			name:    "missing definition ID",
			payload: ConfigureIntegrationRequest{},
			wantErr: true,
		},
		{
			name: "valid with body and credential ref",
			payload: ConfigureIntegrationRequest{
				DefinitionID:  "def_test",
				CredentialRef: "test",
				Body:          json.RawMessage(`{"key":"value"}`),
			},
			wantErr: false,
		},
		{
			name: "body present without credential ref",
			payload: ConfigureIntegrationRequest{
				DefinitionID: "def_test",
				Body:         json.RawMessage(`{"key":"value"}`),
			},
			wantErr: true,
		},
		{
			name: "definition only with no body",
			payload: ConfigureIntegrationRequest{
				DefinitionID: "def_test",
			},
			wantErr: false,
		},
		{
			name: "empty body with credential ref is valid",
			payload: ConfigureIntegrationRequest{
				DefinitionID:  "def_test",
				CredentialRef: "test",
			},
			wantErr: false,
		},
		{
			name: "empty object body does not require credential ref",
			payload: ConfigureIntegrationRequest{
				DefinitionID: "def_test",
				Body:         json.RawMessage(`{}`),
			},
			wantErr: false,
		},
		{
			name: "null body does not require credential ref",
			payload: ConfigureIntegrationRequest{
				DefinitionID: "def_test",
				Body:         json.RawMessage("null"),
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.payload.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIntegrationAuthStartRequestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     IntegrationAuthStartRequest
		wantErr bool
	}{
		{
			name:    "missing definition ID",
			req:     IntegrationAuthStartRequest{CredentialRef: "test"},
			wantErr: true,
		},
		{
			name:    "missing credential ref",
			req:     IntegrationAuthStartRequest{DefinitionID: "def_test"},
			wantErr: true,
		},
		{
			name: "valid",
			req: IntegrationAuthStartRequest{
				DefinitionID:  "def_test",
				CredentialRef: "test",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.req.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifyWebhookHMACSHA256(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"test":"data"}`)
	secret := "test-secret"

	tests := []struct {
		name      string
		secret    string
		signature string
		wantErr   error
	}{
		{
			name:    "empty secret",
			secret:  "",
			wantErr: errIntegrationWebhookSecretMissing,
		},
		{
			name:      "missing signature header",
			secret:    secret,
			signature: "",
			wantErr:   errIntegrationWebhookSignatureMissing,
		},
		{
			name:      "missing sha256 prefix",
			secret:    secret,
			signature: "abc123",
			wantErr:   errIntegrationWebhookSignatureMismatch,
		},
		{
			name:      "invalid hex in signature",
			secret:    secret,
			signature: "sha256=zzzz",
			wantErr:   errIntegrationWebhookSignatureMismatch,
		},
		{
			name:      "wrong signature",
			secret:    secret,
			signature: "sha256=0000000000000000000000000000000000000000000000000000000000000000",
			wantErr:   errIntegrationWebhookSignatureMismatch,
		},
		{
			name:   "valid signature",
			secret: secret,
			signature: func() string {
				return computeTestHMAC(secret, payload)
			}(),
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := &http.Request{Header: http.Header{}}
			if tc.signature != "" {
				req.Header.Set(integrationWebhookSignatureHeader, tc.signature)
			}

			err := verifyWebhookHMACSHA256(req, payload, tc.secret)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func computeTestHMAC(secret string, payload []byte) string {
	mac := hmacSHA256([]byte(secret), payload)
	return "sha256=" + hex.EncodeToString(mac)
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
