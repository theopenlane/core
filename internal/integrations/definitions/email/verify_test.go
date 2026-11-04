package email

import (
	"context"
	"errors"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/newman"
	"github.com/theopenlane/newman/providers/mock"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verifyFailsSender embeds newman.EmailSender so it satisfies the interface without
// implementing every method, and overrides Verify to always return an error
type verifyFailsSender struct {
	newman.EmailSender
}

// Verify always fails, simulating a provider rejecting the configured credentials
func (verifyFailsSender) Verify(context.Context) error {
	return errors.New("invalid api key")
}

func TestVerifySender_NilSender(t *testing.T) {
	client := &Client{
		Sender: nil,
		Config: RuntimeEmailConfig{},
	}

	_, err := verifySender[ResendCredential](providerResend)(context.Background(), types.ConnectionRequest[ResendCredential]{}, client)

	assert.ErrorIs(t, err, ErrSenderNotConfigured)
}

func TestVerifySender_ConfiguredSender(t *testing.T) {
	mockSender, err := mock.New("")
	assert.NilError(t, err)

	client := &Client{
		Sender: mockSender,
		Config: RuntimeEmailConfig{
			FromEmail: "noreply@test.com",
		},
	}

	metadata, err := verifySender[PostmarkCredential](providerPostmark)(context.Background(), types.ConnectionRequest[PostmarkCredential]{}, client)
	assert.NilError(t, err)

	assert.Check(t, is.Equal(providerPostmark, metadata.Provider))
	assert.Check(t, is.Equal("noreply@test.com", metadata.FromEmail))
}

func TestVerifySender_VerifyFailed(t *testing.T) {
	client := &Client{
		Sender: verifyFailsSender{},
		Config: RuntimeEmailConfig{
			FromEmail: "noreply@test.com",
		},
	}

	_, err := verifySender[ResendCredential](providerResend)(context.Background(), types.ConnectionRequest[ResendCredential]{}, client)

	assert.ErrorIs(t, err, ErrSenderVerifyFailed)
}
