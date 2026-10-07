package urlx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/httpsling/httpclient"
)

func TestRequirePublicHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		wantErr error
	}{
		{name: "public hostname", rawURL: "https://hooks.example.com/path"},
		{name: "public ipv4", rawURL: "https://93.184.216.34/hook"},
		{name: "public ipv6", rawURL: "https://[2606:4700::1111]/hook"},
		{name: "localhost", rawURL: "http://localhost:8080", wantErr: ErrNonPublicHost},
		{name: "localhost uppercase trailing dot", rawURL: "http://LOCALHOST./", wantErr: ErrNonPublicHost},
		{name: "localhost subdomain", rawURL: "http://api.localhost/", wantErr: ErrNonPublicHost},
		{name: "ipv4 loopback", rawURL: "http://127.0.0.1/", wantErr: ErrNonPublicHost},
		{name: "ipv4 loopback range", rawURL: "http://127.1.2.3/", wantErr: ErrNonPublicHost},
		{name: "ipv4 unspecified", rawURL: "http://0.0.0.0/", wantErr: ErrNonPublicHost},
		{name: "ipv4 this network", rawURL: "http://0.1.2.3/", wantErr: ErrNonPublicHost},
		{name: "rfc1918 10/8", rawURL: "http://10.0.0.5/", wantErr: ErrNonPublicHost},
		{name: "rfc1918 172.16/12", rawURL: "http://172.20.1.1/", wantErr: ErrNonPublicHost},
		{name: "rfc1918 192.168/16", rawURL: "http://192.168.1.1/", wantErr: ErrNonPublicHost},
		{name: "link local metadata", rawURL: "http://169.254.169.254/latest/meta-data", wantErr: ErrNonPublicHost},
		{name: "cgnat", rawURL: "http://100.64.0.1/", wantErr: ErrNonPublicHost},
		{name: "benchmark range", rawURL: "http://198.18.0.1/", wantErr: ErrNonPublicHost},
		{name: "broadcast", rawURL: "http://255.255.255.255/", wantErr: ErrNonPublicHost},
		{name: "multicast", rawURL: "http://224.0.0.1/", wantErr: ErrNonPublicHost},
		{name: "ipv6 loopback", rawURL: "http://[::1]/", wantErr: ErrNonPublicHost},
		{name: "ipv6 unspecified", rawURL: "http://[::]/", wantErr: ErrNonPublicHost},
		{name: "ipv6 unique local", rawURL: "http://[fd00::1]/", wantErr: ErrNonPublicHost},
		{name: "ipv6 link local with zone", rawURL: "http://[fe80::1%25eth0]/", wantErr: ErrNonPublicHost},
		{name: "ipv4 mapped loopback", rawURL: "http://[::ffff:127.0.0.1]/", wantErr: ErrNonPublicHost},
		{name: "nat64 embedded private", rawURL: "http://[64:ff9b::a00:1]/", wantErr: ErrNonPublicHost},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := ParseAbsolute(tc.rawURL)
			require.NoError(t, err)

			err = RequirePublicHost(parsed)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestPublicOnlyBlocksLoopbackDial(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	client, err := httpclient.New(PublicOnly())
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	if resp != nil {
		resp.Body.Close()
	}

	assert.ErrorIs(t, err, ErrNonPublicHost)
}
