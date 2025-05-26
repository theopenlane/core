package models_test

import (
	"errors"
	"testing"

	"github.com/theopenlane/core/pkg/models"
	"gotest.tools/v3/assert"
)

func TestValidateCronExpression(t *testing.T) {

	tt := []struct {
		name    string
		cron    string
		wantErr bool
	}{
		{
			name:    "valid hourly runs",
			cron:    "0 * * * *",
			wantErr: false,
		},
		{
			name:    "valid run every 30 minutes",
			cron:    "0,30 * * * *",
			wantErr: false,
		},
		{
			name:    "too frequent runs (5 minutes)",
			cron:    "*/5 * * * *",
			wantErr: true,
		},
		{
			name:    "too frequent uneven gaps (0,20,50)",
			cron:    "0,20,50 * * * *",
			wantErr: true,
		},
		{
			name:    "invalid syntax",
			cron:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			err := models.ValidateCronExpression(tt.cron)

			assert.Assert(t, (err != nil) == tt.wantErr)
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:    "valid https url",
			input:   "https://example.com/path",
			want:    "https://example.com/path",
			wantErr: nil,
		},
		{
			name:    "http url rejected",
			input:   "http://example.com",
			wantErr: models.ErrHTTPSOnlyURL,
		},
		{
			name:    "localhost rejected",
			input:   "https://localhost",
			wantErr: models.ErrLocalHostNotAllowed,
		},
		{
			name:    "127.0.0.1 rejected",
			input:   "https://127.0.0.1",
			wantErr: models.ErrLocalHostNotAllowed,
		},
		{
			name:    "loopback IPv6 rejected",
			input:   "https://[::1]",
			wantErr: models.ErrNoLoopbackAddressAllowed,
		},
		{
			name:    "invalid url format",
			input:   "https://%%%",
			wantErr: models.ErrInvalidURL,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: models.ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := models.ValidateURL(tt.input)

			if tt.wantErr != nil {
				assert.Assert(t, errors.Is(err, tt.wantErr))
				return
			}

			assert.Equal(t, got, tt.want)
		})
	}
}
