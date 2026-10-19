package githubapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// githubSignature computes the X-Hub-Signature-256 value GitHub sends for payload under secret
func githubSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func signedGitHubRequest(signature string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	if signature != "" {
		req.Header.Set(githubSignatureHeader, signature)
	}

	return req
}

func TestAppVerify(t *testing.T) {
	t.Parallel()

	payload := []byte(`{"zen":"keep it logically awesome"}`)

	tests := []struct {
		name      string
		secret    string
		signature string
		payload   []byte
		wantErr   error
	}{
		{name: "valid signature", secret: "secret", signature: githubSignature("secret", payload), payload: payload},
		{name: "missing secret fails closed", secret: "", signature: githubSignature("secret", payload), payload: payload, wantErr: ErrWebhookSecretMissing},
		{name: "missing header", secret: "secret", payload: payload, wantErr: ErrWebhookSignatureMissing},
		{name: "wrong secret", secret: "secret", signature: githubSignature("other", payload), payload: payload, wantErr: ErrWebhookSignatureMismatch},
		{name: "tampered payload", secret: "secret", signature: githubSignature("secret", payload), payload: []byte(`{"zen":"tampered"}`), wantErr: ErrWebhookSignatureMismatch},
		{name: "malformed header", secret: "secret", signature: "sha1=abc", payload: payload, wantErr: ErrWebhookSignatureMismatch},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := App{Config: Config{WebhookSecret: tc.secret}}.Verify(types.WebhookInboundRequest{
				Request: signedGitHubRequest(tc.signature),
				Payload: tc.payload,
			})

			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
