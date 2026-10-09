package urlx

import (
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/theopenlane/httpsling/httpclient"
)

const (
	publicDialTimeout   = 30 * time.Second
	publicDialKeepAlive = 30 * time.Second
)

// nonPublicPrefixes contains special-purpose address ranges that netip considers
// global unicast but that should not be valid outbound webhook destinations.
var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),       // "this network", some stacks route it to localhost
	netip.MustParsePrefix("100.64.0.0/10"),   // carrier-grade NAT shared space, used by cloud internal services
	netip.MustParsePrefix("192.0.0.0/24"),    // IETF protocol assignments, e.g. DS-Lite and NAT64 discovery
	netip.MustParsePrefix("192.0.2.0/24"),    // TEST-NET-1 documentation
	netip.MustParsePrefix("198.18.0.0/15"),   // benchmarking, reused internally by labs and fake-IP proxies
	netip.MustParsePrefix("198.51.100.0/24"), // TEST-NET-2 documentation
	netip.MustParsePrefix("203.0.113.0/24"),  // TEST-NET-3 documentation
	netip.MustParsePrefix("240.0.0.0/4"),     // reserved for future use
	netip.MustParsePrefix("64:ff9b::/96"),    // NAT64 well-known prefix, embeds an IPv4 address
	netip.MustParsePrefix("64:ff9b:1::/48"),  // NAT64 local-use prefix, embeds an IPv4 address
	netip.MustParsePrefix("100::/64"),        // IPv6 discard-only
	netip.MustParsePrefix("2001::/32"),       // Teredo tunneling, embeds an IPv4 address
	netip.MustParsePrefix("2001:db8::/32"),   // IPv6 documentation
	netip.MustParsePrefix("2002::/16"),       // 6to4 tunneling, embeds an IPv4 address
}

// nonPublicHostSuffixes are hostname suffixes that only resolve inside private networks
var nonPublicHostSuffixes = []string{".internal", ".local", ".localhost", ".localdomain", ".home.arpa"}

// IsPublicAddr reports whether addr is a globally routable unicast address
func IsPublicAddr(addr netip.Addr) bool {
	addr = addr.Unmap()

	if !addr.IsValid() || !addr.IsGlobalUnicast() || addr.IsPrivate() {
		return false
	}

	for _, prefix := range nonPublicPrefixes {
		if prefix.Contains(addr) {
			return false
		}
	}

	return true
}

// ValidatePublicURL parses rawURL as an absolute http(s) URL and rejects hosts that are
// IP literals outside public address space or names that only resolve on private networks;
// the dial-time check in PublicOnly remains the authoritative guard
func ValidatePublicURL(rawURL string) (*url.URL, error) {
	parsed, err := ParseAbsolute(rawURL)
	if err != nil {
		return nil, err
	}

	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")

	if addr, err := netip.ParseAddr(host); err == nil {
		if !IsPublicAddr(addr) {
			return nil, fmt.Errorf("%w: %s", ErrNonPublicDestination, host)
		}

		return parsed, nil
	}

	// single-label names such as "metadata" or "localhost" resolve through internal search domains
	if !strings.Contains(host, ".") {
		return nil, fmt.Errorf("%w: %s", ErrNonPublicDestination, host)
	}

	for _, suffix := range nonPublicHostSuffixes {
		if strings.HasSuffix(host, suffix) {
			return nil, fmt.Errorf("%w: %s", ErrNonPublicDestination, host)
		}
	}

	return parsed, nil
}

// PublicOnly restricts a client to public destinations by checking every resolved address at
// dial time, which covers redirect hops and DNS rebinding; proxies are disabled so the check
// applies to the real destination rather than the proxy
func PublicOnly() httpclient.Option {
	return httpclient.TransportOption(func(t *http.Transport) error {
		t.Proxy = nil
		t.DialContext = (&net.Dialer{
			Timeout:   publicDialTimeout,
			KeepAlive: publicDialKeepAlive,
			Control:   publicOnlyControl,
		}).DialContext

		return nil
	})
}

func publicOnlyControl(_ string, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrNonPublicDestination, address)
	}

	addr, err := netip.ParseAddr(host)
	if err != nil || !IsPublicAddr(addr) {
		return fmt.Errorf("%w: %s", ErrNonPublicDestination, host)
	}

	return nil
}
