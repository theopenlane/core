//go:build test

package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestIntegrationConfigPayloadValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload IntegrationConfigPayload
		wantErr bool
	}{
		{
			name:    "missing definition ID",
			payload: IntegrationConfigPayload{},
			wantErr: true,
		},
		{
			name: "valid with body and credential ref",
			payload: IntegrationConfigPayload{
				DefinitionID:  "def_test",
				CredentialRef: types.NewCredentialSlotID("test"),
				Body:          json.RawMessage(`{"key":"value"}`),
			},
			wantErr: false,
		},
		{
			name: "body present without credential ref",
			payload: IntegrationConfigPayload{
				DefinitionID: "def_test",
				Body:         json.RawMessage(`{"key":"value"}`),
			},
			wantErr: true,
		},
		{
			name: "definition only with no body",
			payload: IntegrationConfigPayload{
				DefinitionID: "def_test",
			},
			wantErr: false,
		},
		{
			name: "empty body with credential ref is valid",
			payload: IntegrationConfigPayload{
				DefinitionID:  "def_test",
				CredentialRef: types.NewCredentialSlotID("test"),
			},
			wantErr: false,
		},
		{
			name: "null body does not require credential ref",
			payload: IntegrationConfigPayload{
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
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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
			req:     IntegrationAuthStartRequest{CredentialRef: types.NewCredentialSlotID("test")},
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
				CredentialRef: types.NewCredentialSlotID("test"),
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.req.Validate()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRefreshInstallationCredentialRequestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     RefreshInstallationCredentialRequest
		wantErr bool
	}{
		{
			name:    "missing installation ID",
			req:     RefreshInstallationCredentialRequest{},
			wantErr: true,
		},
		{
			name:    "valid",
			req:     RefreshInstallationCredentialRequest{InstallationID: "01HXYZ1234567890ABCDEFGHJ"},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.req.Validate()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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
