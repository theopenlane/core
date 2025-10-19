package corejobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTokenPrefix(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected bool
	}{
		{
			name:     "valid personal access token",
			token:    "tolp_test123456",
			expected: true,
		},
		{
			name:     "valid API token",
			token:    "tola_apitoken123",
			expected: true,
		},
		{
			name:     "valid job runner token",
			token:    "runner_jobtoken456",
			expected: true,
		},
		{
			name:     "invalid token prefix",
			token:    "invalid_prefix",
			expected: false,
		},
		{
			name:     "empty token",
			token:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateTokenPrefix(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestOpenlaneConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  OpenlaneConfig
		wantErr error
	}{
		{
			name: "valid config",
			config: OpenlaneConfig{
				OpenlaneAPIHost:  "https://api.openlane.io",
				OpenlaneAPIToken: "tolp_validtoken123",
			},
			wantErr: nil,
		},
		{
			name: "missing host",
			config: OpenlaneConfig{
				OpenlaneAPIHost:  "",
				OpenlaneAPIToken: "tolp_validtoken123",
			},
			wantErr: ErrOpenlaneHostMissing,
		},
		{
			name: "missing token",
			config: OpenlaneConfig{
				OpenlaneAPIHost:  "https://api.openlane.com",
				OpenlaneAPIToken: "",
			},
			wantErr: ErrOpenlaneTokenMissing,
		},
		{
			name: "invalid token prefix",
			config: OpenlaneConfig{
				OpenlaneAPIHost:  "https://api.openlane.com",
				OpenlaneAPIToken: "invalid_prefix_token",
			},
			wantErr: ErrOpenlaneTokenMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
