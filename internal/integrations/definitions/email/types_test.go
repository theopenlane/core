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
			name:     "resend with its key",
			cfg:      RuntimeEmailConfig{Provider: "resend", Resend: ResendConfig{ResendCredential: ResendCredential{APIKey: "key"}}, FromEmail: "test@example.com"},
			expected: true,
		},
		{
			name:     "postmark with its token",
			cfg:      RuntimeEmailConfig{Provider: "postmark", Postmark: PostmarkConfig{PostmarkCredential: PostmarkCredential{ServerToken: "token"}}, FromEmail: "test@example.com"},
			expected: true,
		},
		{
			name:     "resend selected but only postmark configured",
			cfg:      RuntimeEmailConfig{Provider: "resend", Postmark: PostmarkConfig{PostmarkCredential: PostmarkCredential{ServerToken: "token"}}, FromEmail: "test@example.com"},
			expected: false,
		},
		{
			name:     "missing provider",
			cfg:      RuntimeEmailConfig{Resend: ResendConfig{ResendCredential: ResendCredential{APIKey: "key"}}, FromEmail: "test@example.com"},
			expected: false,
		},
		{
			name:     "missing from email",
			cfg:      RuntimeEmailConfig{Provider: "resend", Resend: ResendConfig{ResendCredential: ResendCredential{APIKey: "key"}}},
			expected: false,
		},
		{
			name:     "mock needs no credential",
			cfg:      RuntimeEmailConfig{Provider: ProviderMock, FromEmail: "test@example.com"},
			expected: true,
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
