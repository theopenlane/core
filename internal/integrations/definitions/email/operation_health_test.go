package email

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/newman/providers/mock"
)

func TestHealthCheck_NilSender(t *testing.T) {
	client := &Client{
		Sender: nil,
		Config: RuntimeEmailConfig{},
	}

	_, err := HealthCheck{}.Run(context.Background(), client)

	require.ErrorIs(t, err, ErrSenderNotConfigured)
}

func TestHealthCheck_ConfiguredSender(t *testing.T) {
	mockSender, err := mock.New("")
	require.NoError(t, err)

	client := &Client{
		Sender: mockSender,
		Config: RuntimeEmailConfig{
			Provider:  "mock",
			FromEmail: "noreply@test.com",
		},
	}

	result, err := HealthCheck{}.Run(context.Background(), client)
	require.NoError(t, err)

	var data map[string]any
	require.NoError(t, json.Unmarshal(result, &data))

	assert.Equal(t, "mock", data["provider"])
	assert.Equal(t, "noreply@test.com", data["fromEmail"])
}
