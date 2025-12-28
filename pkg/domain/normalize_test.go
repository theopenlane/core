package domain

import (
	"errors"
	"testing"
)

func TestNormalizeHostname(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "raw hostname",
			input: "trust.example.com",
			want:  "trust.example.com",
		},
		{
			name:  "url with path and mixed case",
			input: "https://Trust.Example.com/path",
			want:  "trust.example.com",
		},
		{
			name:  "hostname with trailing dot",
			input: "Trust.Example.com.",
			want:  "trust.example.com",
		},
		{
			name:  "hostname with port",
			input: "trust.example.com:8443",
			want:  "trust.example.com",
		},
		{
			name:    "empty input",
			input:   "  ",
			wantErr: ErrEmptyHostname,
		},
		{
			name:    "invalid input",
			input:   "http://@",
			wantErr: ErrInvalidHostname,
		},
		{
			name:    "invalid url escape",
			input:   "http://example.com/%zz",
			wantErr: ErrInvalidHostname,
		},
		{
			name:    "scheme without hostname",
			input:   "http://",
			wantErr: ErrInvalidHostname,
		},
		{
			name:    "invalid port after scheme applied",
			input:   "example.com:bad",
			wantErr: ErrInvalidHostname,
		},
		{
			name:    "host trims to empty",
			input:   ".",
			wantErr: ErrInvalidHostname,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeHostname(tc.input)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.wantErr)
				}
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
