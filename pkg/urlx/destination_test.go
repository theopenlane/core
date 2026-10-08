package urlx

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPublicAddr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		addr     string
		expected bool
	}{
		{addr: "8.8.8.8", expected: true},
		{addr: "2606:4700:4700::1111", expected: true},
		{addr: "127.0.0.1", expected: false},
		{addr: "127.0.0.2", expected: false},
		{addr: "10.1.2.3", expected: false},
		{addr: "172.16.0.1", expected: false},
		{addr: "192.168.1.1", expected: false},
		{addr: "169.254.169.254", expected: false},
		{addr: "100.64.0.1", expected: false},
		{addr: "0.0.0.0", expected: false},
		{addr: "255.255.255.255", expected: false},
		{addr: "224.0.0.1", expected: false},
		{addr: "::1", expected: false},
		{addr: "::", expected: false},
		{addr: "fe80::1", expected: false},
		{addr: "fd00:ec2::254", expected: false},
		{addr: "::ffff:169.254.169.254", expected: false},
		{addr: "64:ff9b::a9fe:a9fe", expected: false},
		{addr: "2002:a9fe:a9fe::1", expected: false},
		{addr: "192.0.0.170", expected: false},
		{addr: "192.0.2.10", expected: false},
		{addr: "198.18.0.1", expected: false},
		{addr: "198.51.100.7", expected: false},
		{addr: "203.0.113.5", expected: false},
		{addr: "100::1", expected: false},
		{addr: "2001:db8::1", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, IsPublicAddr(netip.MustParseAddr(tt.addr)))
		})
	}
}

func TestValidatePublicURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url     string
		wantErr error
	}{
		{url: "https://example.com/webhook"},
		{url: "https://8.8.8.8/hook"},
		{url: "https://hooks.example.com./path"},
		{url: "http://metadata.google.internal/computeMetadata/v1/", wantErr: ErrNonPublicDestination},
		{url: "http://METADATA.GOOGLE.INTERNAL./", wantErr: ErrNonPublicDestination},
		{url: "http://metadata/computeMetadata/v1/", wantErr: ErrNonPublicDestination},
		{url: "http://localhost:8080/", wantErr: ErrNonPublicDestination},
		{url: "http://api.localhost/", wantErr: ErrNonPublicDestination},
		{url: "http://printer.local/", wantErr: ErrNonPublicDestination},
		{url: "http://169.254.169.254/latest/meta-data/", wantErr: ErrNonPublicDestination},
		{url: "http://[::1]:8080/", wantErr: ErrNonPublicDestination},
		{url: "http://[::ffff:127.0.0.1]/", wantErr: ErrNonPublicDestination},
		{url: "http://2130706433/", wantErr: ErrNonPublicDestination},
		{url: "https://[2606:4700::1111]/hook"},
		{url: "http://LOCALHOST./", wantErr: ErrNonPublicDestination},
		{url: "http://127.1.2.3/", wantErr: ErrNonPublicDestination},
		{url: "http://0.1.2.3/", wantErr: ErrNonPublicDestination},
		{url: "http://172.20.1.1/", wantErr: ErrNonPublicDestination},
		{url: "http://[fd00::1]/", wantErr: ErrNonPublicDestination},
		{url: "http://[fe80::1%25eth0]/", wantErr: ErrNonPublicDestination},
		{url: "http://[64:ff9b::a00:1]/", wantErr: ErrNonPublicDestination},
		{url: "ftp://example.com/", wantErr: ErrUnsupportedScheme},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()

			_, err := ValidatePublicURL(tt.url)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestPublicOnlyBlocksLoopbackDial(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client, err := NewHTTPClient(PublicOnly())
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	assert.ErrorIs(t, err, ErrNonPublicDestination)
}
