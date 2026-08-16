package urlx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr error
	}{
		{
			name:   "bare host gets https scheme",
			rawURL: "example.com",
			want:   "https://example.com",
		},
		{
			name:   "explicit http scheme is preserved",
			rawURL: "http://example.com",
			want:   "http://example.com",
		},
		{
			name:   "path and query are preserved",
			rawURL: "example.com/path?x=1",
			want:   "https://example.com/path?x=1",
		},
		{
			name:   "bare host with port",
			rawURL: "example.com:8443",
			want:   "https://example.com:8443",
		},
		{
			name:   "protocol relative input",
			rawURL: "//example.com/path",
			want:   "https://example.com/path",
		},
		{
			name:   "surrounding whitespace is trimmed",
			rawURL: "  example.com  ",
			want:   "https://example.com",
		},
		{
			name:    "empty input",
			rawURL:  "   ",
			wantErr: ErrEmptyURL,
		},
		{
			name:    "scheme without host",
			rawURL:  "https://",
			wantErr: ErrMissingHost,
		},
		{
			name:    "userinfo without host",
			rawURL:  "http://@",
			wantErr: ErrMissingHost,
		},
		{
			name:    "invalid url escape",
			rawURL:  "http://example.com/%zz",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "invalid port after scheme applied",
			rawURL:  "example.com:bad",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "non http scheme",
			rawURL:  "ftp://example.com",
			wantErr: ErrUnsupportedScheme,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tc.rawURL)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got.String())
		})
	}
}

func TestParseAbsolute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr error
	}{
		{
			name:   "https url with path",
			rawURL: "https://example.com/manifest.json",
			want:   "https://example.com/manifest.json",
		},
		{
			name:   "http url with port",
			rawURL: "http://example.com:8080/manifest.json",
			want:   "http://example.com:8080/manifest.json",
		},
		{
			name:    "schemeless input",
			rawURL:  "example.com/manifest.json",
			wantErr: ErrUnsupportedScheme,
		},
		{
			name:    "non http scheme",
			rawURL:  "ftp://example.com",
			wantErr: ErrUnsupportedScheme,
		},
		{
			name:    "scheme without host",
			rawURL:  "https://",
			wantErr: ErrMissingHost,
		},
		{
			name:    "empty input",
			rawURL:  "",
			wantErr: ErrEmptyURL,
		},
		{
			name:    "invalid url escape",
			rawURL:  "https://example.com/%zz",
			wantErr: ErrInvalidURL,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseAbsolute(tc.rawURL)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got.String())
		})
	}
}

func TestNormalizeHostname(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr error
	}{
		{
			name:   "raw hostname",
			rawURL: "trust.example.com",
			want:   "trust.example.com",
		},
		{
			name:   "url with path and mixed case",
			rawURL: "https://Trust.Example.com/path",
			want:   "trust.example.com",
		},
		{
			name:   "hostname with trailing dot",
			rawURL: "Trust.Example.com.",
			want:   "trust.example.com",
		},
		{
			name:   "hostname with port",
			rawURL: "trust.example.com:8443",
			want:   "trust.example.com",
		},
		{
			name:    "empty input",
			rawURL:  "  ",
			wantErr: ErrEmptyURL,
		},
		{
			name:    "userinfo without host",
			rawURL:  "http://@",
			wantErr: ErrMissingHost,
		},
		{
			name:    "invalid url escape",
			rawURL:  "http://example.com/%zz",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "scheme without hostname",
			rawURL:  "http://",
			wantErr: ErrMissingHost,
		},
		{
			name:    "invalid port after scheme applied",
			rawURL:  "example.com:bad",
			wantErr: ErrInvalidURL,
		},
		{
			name:    "host trims to empty",
			rawURL:  ".",
			wantErr: ErrMissingHost,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeHostname(tc.rawURL)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWithoutQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		rawURL string
		want   string
	}{
		{
			name:   "query and fragment are stripped",
			rawURL: "https://example.com/webhook?token=abc#frag",
			want:   "https://example.com/webhook",
		},
		{
			name:   "url without query is unchanged",
			rawURL: "https://example.com/webhook",
			want:   "https://example.com/webhook",
		},
		{
			name:   "unparsable input is returned unchanged",
			rawURL: "https://example.com/%zz",
			want:   "https://example.com/%zz",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, WithoutQuery(tc.rawURL))
		})
	}
}
