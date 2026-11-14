package clientip

import (
	"crypto/subtle"
	"net"
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"
)

const (
	// ForwardedClientIPHeader carries the originating visitor IP asserted by a trusted forwarding proxy
	ForwardedClientIPHeader = "X-Openlane-Client-IP"
	// ForwarderSecretHeader carries the shared secret a trusted forwarding proxy presents
	ForwarderSecretHeader = "X-Openlane-Forwarder-Secret" //nolint:gosec // header name, not a credential

	// remoteAddrSource names the socket peer as an entry in a header precedence list
	remoteAddrSource = "remoteaddr"
	// forwardedForSource names X-Forwarded-For as an entry in a header precedence list
	forwardedForSource = "x-forwarded-for"
)

// DefaultHeaders is the header precedence used to derive the originating client IP
// prefers CF-Connecting-IP, then True-Client-IP, falls back to the RemoteAddr
var DefaultHeaders = []string{"CF-Connecting-IP", "True-Client-IP", "RemoteAddr"}

// Options controls how the client IP is resolved
type Options struct {
	// ForwarderSecret is the shared secret trusted forwarders present; empty disables ForwardedClientIPHeader
	ForwarderSecret string
	// Headers is the precedence list consulted after the trusted forwarder header; empty uses DefaultHeaders
	Headers []string
	// ForwardedIndexFromBehind selects which X-Forwarded-For entry is used, counted from the last one
	ForwardedIndexFromBehind int
}

// TrustedForwardedIP returns the client IP asserted in ForwardedClientIPHeader when the request has the forwarder secret in ForwarderSecretHeader
// it returns an empty string when no secret is configured, the secret does not match, or the asserted value is not a valid IP
func TrustedForwardedIP(req *http.Request, secret string) string {
	if secret == "" {
		return ""
	}

	presented := req.Header.Get(ForwarderSecretHeader)
	if subtle.ConstantTimeCompare([]byte(presented), []byte(secret)) != 1 {
		return ""
	}

	ip := net.ParseIP(strings.TrimSpace(req.Header.Get(ForwardedClientIPHeader)))
	if ip == nil {
		return ""
	}

	return ip.String()
}

// Resolve returns the originating client IP: the trusted forwarder's asserted IP when its secret matches, otherwise
// the first non-empty source in the header precedence list. It returns an empty string when no source yields a value
func Resolve(req *http.Request, opts Options) string {
	if ip := TrustedForwardedIP(req, opts.ForwarderSecret); ip != "" {
		return ip
	}

	headers := opts.Headers
	if len(headers) == 0 {
		headers = DefaultHeaders
	}

	for _, header := range headers {
		switch strings.ToLower(header) {
		case remoteAddrSource:
			if req.RemoteAddr == "" {
				continue
			}

			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				return req.RemoteAddr
			}

			return ip
		case forwardedForSource:
			forwarded := req.Header.Get(echo.HeaderXForwardedFor)
			if forwarded == "" {
				continue
			}

			parts := strings.Split(forwarded, ",")
			for i, part := range parts {
				parts[i] = strings.TrimSpace(part)
			}

			index := max(len(parts)-1-opts.ForwardedIndexFromBehind, 0)

			if parts[index] != "" {
				return parts[index]
			}
		default:
			if value := req.Header.Get(header); value != "" {
				return value
			}
		}
	}

	return ""
}

// Extractor returns an echo IPExtractor backed by Resolve, so c.RealIP() (request logs, audit records)
// resolves the client IP the same way as the rate limiter
func Extractor(opts Options) echo.IPExtractor {
	return func(req *http.Request) string {
		if ip := Resolve(req, opts); ip != "" {
			return ip
		}

		ra, _, _ := net.SplitHostPort(req.RemoteAddr)

		return ra
	}
}
