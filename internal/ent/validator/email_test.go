package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailVerificationConfig_VerifyEmailAddress(t *testing.T) {
	tests := []struct {
		name      string
		config    EmailVerificationConfig
		email     string
		wantValid bool
		wantErr   bool
		emptyVal  bool
	}{
		{
			name: "valid corporate email",
			config: EmailVerificationConfig{
				Enabled: true,
				AllowedEmailTypes: AllowedEmailTypes{
					Disposable: false,
					Free:       false,
					Role:       false,
				},
			},
			email:     "mitb@theopenlane.io",
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "invalid syntax",
			config: EmailVerificationConfig{
				Enabled: true,
				AllowedEmailTypes: AllowedEmailTypes{
					Disposable: false,
					Free:       false,
					Role:       false,
				},
			},
			email:     "lasso",
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "disposable email not allowed",
			config: EmailVerificationConfig{
				Enabled: true,
				AllowedEmailTypes: AllowedEmailTypes{
					Disposable: false,
					Free:       true,
					Role:       true,
				},
			},
			email:     "b23f6e80d0d86@temp-mail.org",
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "disposable email allowed",
			config: EmailVerificationConfig{
				Enabled: true,
				AllowedEmailTypes: AllowedEmailTypes{
					Disposable: true,
					Free:       true,
					Role:       true,
				},
			},
			email:     "b23f6e80d0d86@temp-mail.org",
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "free email not allowed",
			config: EmailVerificationConfig{
				Enabled: true,
				AllowedEmailTypes: AllowedEmailTypes{
					Disposable: true,
					Free:       false,
					Role:       true,
				},
			},
			email:     "user@gmail.com",
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "role email not allowed",
			config: EmailVerificationConfig{
				Enabled: true,
				AllowedEmailTypes: AllowedEmailTypes{
					Disposable: true,
					Free:       true,
					Role:       false,
				},
			},
			email:     "support@theopenlane.io",
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "verification error on domain lookup",
			config: EmailVerificationConfig{
				Enabled: true,
				AllowedEmailTypes: AllowedEmailTypes{
					Disposable: true,
					Free:       true,
					Role:       true,
				},
			},
			email:     "user@company.com",
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "not enabled",
			config: EmailVerificationConfig{
				Enabled: false,
			},
			email:     "user@company.com",
			wantValid: true,
			wantErr:   false,
			emptyVal:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := tt.config.NewVerifier()
			got, res, err := verifier.VerifyEmailAddress(tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantValid, got)

			if tt.wantValid && !tt.emptyVal {
				assert.NotNil(t, res)
			} else {
				assert.Nil(t, res)
			}
		})
	}
}
