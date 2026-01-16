package graphapi

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestCheckOrigin(t *testing.T) {
	allowedOrigins := map[string]struct{}{
		"https://allowed.com":      {},
		"https://*.vercel.app":     {},
		"https://sub.allowed.com":  {},
		"https://allowed.com:8080": {},
		"https://allowed.com/":     {},
	}

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{
			name:   "Allowed origin",
			origin: "https://allowed.com",
			want:   true,
		},
		{
			name:   "not-allowed origin",
			origin: "https://notallowed.com",
			want:   false,
		},
		{
			name:   "wildcard subdomain match",
			origin: "https://openlane.vercel.app",
			want:   true,
		},
		{
			name:   "Origin is empty, not allowed",
			origin: "",
			want:   false,
		},
		{
			name:   "allow origin with trailing slash",
			origin: "https://allowed.com/",
			want:   true,
		},
		{
			name:   "allow origin with port",
			origin: "https://allowed.com:8080",
			want:   true,
		},
		{
			name:   "allowed origin with port",
			origin: "https://allowed.com:8080",
			want:   true,
		},
		{
			name:   "allow origin with subdomain",
			origin: "https://sub.allowed.com",
			want:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checkOrigin(tc.origin, allowedOrigins)
			assert.Check(t, is.Equal(got, tc.want))
		})
	}
}
