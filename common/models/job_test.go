package models_test

import (
	"errors"
	"testing"

	"github.com/theopenlane/core/common/models"
	"gotest.tools/v3/assert"
)

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
