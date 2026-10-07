package email

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/newman/providers/mock"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestVerify_NilSender(t *testing.T) {
	client := &Client{
		Sender: nil,
		Config: RuntimeEmailConfig{},
	}

	_, err := verify(context.Background(), types.ConnectionRequest[Credential]{}, client)

	require.ErrorIs(t, err, ErrSenderNotConfigured)
}

func TestVerify_ConfiguredSender(t *testing.T) {
	mockSender, err := mock.New("")
	require.NoError(t, err)

	client := &Client{
		Sender: mockSender,
		Config: RuntimeEmailConfig{
			Provider:  "mock",
			FromEmail: "noreply@test.com",
		},
	}

	metadata, err := verify(context.Background(), types.ConnectionRequest[Credential]{}, client)
	require.NoError(t, err)

	assert.Equal(t, "mock", metadata.Provider)
	assert.Equal(t, "noreply@test.com", metadata.FromEmail)
}
