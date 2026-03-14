package githubapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/theopenlane/core/internal/ent/generated"
)

const (
	githubSignatureHeader = "X-Hub-Signature-256"
	githubEventHeader     = "X-GitHub-Event"
)

// verifyWebhook validates the HMAC-SHA256 signature on an inbound GitHub webhook request
func (d *def) verifyWebhook(_ context.Context, request *http.Request) error {
	if d.cfg.WebhookSecret == "" {
		return ErrWebhookSecretMissing
	}

	sig := request.Header.Get(githubSignatureHeader)
	if sig == "" {
		return errors.New("githubapp: missing webhook signature header")
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return fmt.Errorf("githubapp: failed to read webhook body: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(d.cfg.WebhookSecret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return errors.New("githubapp: webhook signature mismatch")
	}

	return nil
}

// resolveWebhook resolves an inbound GitHub webhook request to an integration record
func (d *def) resolveWebhook(_ context.Context, _ *http.Request) (*generated.Integration, error) {
	return nil, errors.New("githubapp: webhook resolution not implemented")
}

// handleWebhook dispatches an inbound GitHub webhook event
func (d *def) handleWebhook(_ context.Context, request *http.Request, _ *generated.Integration) error {
	event := strings.TrimSpace(request.Header.Get(githubEventHeader))
	if event == "" {
		return errors.New("githubapp: missing X-GitHub-Event header")
	}

	return fmt.Errorf("githubapp: unhandled event type %q", event)
}
