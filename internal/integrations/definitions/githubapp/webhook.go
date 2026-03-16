package githubapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

const githubSignatureHeader = "X-Hub-Signature-256"

// Webhook verifies inbound GitHub App webhook signatures
type Webhook struct {
	// Config holds the operator-supplied webhook verification settings
	Config Config
}

// Verify validates the HMAC-SHA256 signature on an inbound GitHub webhook request
func (w Webhook) Verify(ctx context.Context, request *http.Request) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if w.Config.WebhookSecret == "" {
		return ErrWebhookSecretMissing
	}

	signature := request.Header.Get(githubSignatureHeader)
	if signature == "" {
		return ErrWebhookSignatureMissing
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return ErrWebhookBodyRead
	}

	mac := hmac.New(sha256.New, []byte(w.Config.WebhookSecret))
	_, _ = mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return ErrWebhookSignatureMismatch
	}

	return nil
}
