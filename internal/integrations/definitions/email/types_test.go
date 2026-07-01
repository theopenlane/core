package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeEmailConfig_Provisioned(t *testing.T) {
	tests := []struct {
		name     string
		cfg      RuntimeEmailConfig
		expected bool
	}{
		{
			name:     "all fields present",
			cfg:      RuntimeEmailConfig{APIKey: "key", Provider: "resend", FromEmail: "test@example.com"},
			expected: true,
		},
		{
			name:     "missing api key",
			cfg:      RuntimeEmailConfig{Provider: "resend", FromEmail: "test@example.com"},
			expected: false,
		},
		{
			name:     "missing provider",
			cfg:      RuntimeEmailConfig{APIKey: "key", FromEmail: "test@example.com"},
			expected: false,
		},
		{
			name:     "missing from email",
			cfg:      RuntimeEmailConfig{APIKey: "key", Provider: "resend"},
			expected: false,
		},
		{
			name:     "zero value",
			cfg:      RuntimeEmailConfig{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.cfg.Provisioned())
		})
	}
}

func TestUserInput_ToRuntimeConfig(t *testing.T) {
	input := UserInput{
		FromEmail:      "from@example.com",
		CompanyName:    "Acme Corp",
		CompanyAddress: "123 Main St",
	}

	cfg := input.ToRuntimeConfig()

	require.Equal(t, input.FromEmail, cfg.FromEmail)
	assert.Equal(t, input.CompanyName, cfg.CompanyName)
	assert.Equal(t, input.CompanyAddress, cfg.CompanyAddress)
}
