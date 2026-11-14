package clientip_test

import (
	"net/http/httptest"
	"testing"

	echo "github.com/theopenlane/echox"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/middleware/clientip"
)

func TestExtractor(t *testing.T) {
	t.Parallel()

	const secret = "forwarder-secret"

	tests := []struct {
		name     string
		opts     clientip.Options
		headers  map[string]string
		expected string
	}{
		{
			name: "trusted forwarder ip wins over cloudflare",
			opts: clientip.Options{ForwarderSecret: secret},
			headers: map[string]string{
				clientip.ForwarderSecretHeader:   secret,
				clientip.ForwardedClientIPHeader: "203.0.113.7",
				clientip.HeaderCFConnectingIP:    "198.51.100.1",
			},
			expected: "203.0.113.7",
		},
		{
			name: "wrong secret falls back to cloudflare",
			opts: clientip.Options{ForwarderSecret: secret},
			headers: map[string]string{
				clientip.ForwarderSecretHeader:   "nope",
				clientip.ForwardedClientIPHeader: "203.0.113.7",
				clientip.HeaderCFConnectingIP:    "198.51.100.1",
			},
			expected: "198.51.100.1",
		},
		{
			name: "no configured secret ignores the forwarded header",
			opts: clientip.Options{},
			headers: map[string]string{
				clientip.ForwarderSecretHeader:   "",
				clientip.ForwardedClientIPHeader: "203.0.113.7",
				clientip.HeaderCFConnectingIP:    "198.51.100.1",
			},
			expected: "198.51.100.1",
		},
		{
			name: "invalid forwarded ip falls back",
			opts: clientip.Options{ForwarderSecret: secret},
			headers: map[string]string{
				clientip.ForwarderSecretHeader:   secret,
				clientip.ForwardedClientIPHeader: "not-an-ip",
				clientip.HeaderCFConnectingIP:    "198.51.100.1",
			},
			expected: "198.51.100.1",
		},
		{
			name: "client supplied x-forwarded-for is ignored by default",
			opts: clientip.Options{},
			headers: map[string]string{
				echo.HeaderXForwardedFor: "6.6.6.6, 198.51.100.1",
			},
			expected: "192.0.2.1",
		},
		{
			name: "true-client-ip used when cloudflare header absent",
			opts: clientip.Options{},
			headers: map[string]string{
				clientip.HeaderTrueClientIP: "198.51.100.9",
			},
			expected: "198.51.100.9",
		},
		{
			name: "configured x-forwarded-for counts from the end",
			opts: clientip.Options{Headers: []string{"X-Forwarded-For"}, ForwardedIndexFromBehind: 1},
			headers: map[string]string{
				echo.HeaderXForwardedFor: "6.6.6.6, 198.51.100.1, 10.0.0.1",
			},
			expected: "198.51.100.1",
		},
		{
			name:     "no headers uses the socket peer",
			opts:     clientip.Options{ForwarderSecret: secret},
			headers:  map[string]string{},
			expected: "192.0.2.1",
		},
		{
			name:     "configured headers without remoteaddr still fall back to the socket peer",
			opts:     clientip.Options{Headers: []string{clientip.HeaderCFConnectingIP}},
			headers:  map[string]string{},
			expected: "192.0.2.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "192.0.2.1:1234"

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			assert.Check(t, is.Equal(tc.expected, clientip.Extractor(tc.opts)(req)))
		})
	}
}
