package models

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid https url",
			input:   "https://example.com/path",
			wantErr: nil,
		},
		{
			name:    "http url rejected",
			input:   "http://example.com",
			wantErr: ErrHTTPSOnlyURL,
		},
		{
			name:    "localhost rejected",
			input:   "https://localhost",
			wantErr: ErrLocalHostNotAllowed,
		},
		{
			name:    "127.0.0.1 rejected",
			input:   "https://127.0.0.1",
			wantErr: ErrLocalHostNotAllowed,
		},
		{
			name:    "loopback IPv6 rejected",
			input:   "https://[::1]",
			wantErr: ErrNoLoopbackAddressAllowed,
		},
		{
			name:    "invalid url format",
			input:   "https://%%%",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.input)

			if tt.wantErr != nil {
				assert.Assert(t, errors.Is(err, tt.wantErr))
				return
			}

			assert.NilError(t, err)
		})
	}
}
