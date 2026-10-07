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

	"github.com/samber/lo"
	"github.com/theopenlane/httpsling/httpclient"
)

const (
	// localhostName is the reserved loopback hostname, including its subdomains per RFC 6761
	localhostName = "localhost"
	// publicDialTimeout bounds connection establishment for public-only clients
	publicDialTimeout = 30 * time.Second
	// publicDialKeepAlive is the TCP keep-alive period for public-only clients
	publicDialKeepAlive = 30 * time.Second
)

// nonPublicPrefixes are ranges netip reports as global unicast that are still not publicly routable;
// NAT64 prefixes are included because they embed (and can translate to) arbitrary IPv4 targets
var nonPublicPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
}

// IsPublicAddr reports whether addr is a publicly routable unicast address
func IsPublicAddr(addr netip.Addr) bool {
	addr = addr.Unmap()

	if !addr.IsGlobalUnicast() || addr.IsPrivate() {
		return false
	}

	return !lo.SomeBy(nonPublicPrefixes, func(p netip.Prefix) bool { return p.Contains(addr) })
}

// RequirePublicHost returns ErrNonPublicHost when u targets localhost or an IP literal that is not
// publicly routable; hostnames are not resolved, PublicOnly enforces resolved addresses at dial time
func RequirePublicHost(u *url.URL) error {
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")

	if host == localhostName || strings.HasSuffix(host, "."+localhostName) {
		return ErrNonPublicHost
	}

	if addr, err := netip.ParseAddr(host); err == nil && !IsPublicAddr(addr) {
		return ErrNonPublicHost
	}

	return nil
}

// PublicOnly returns an httpclient option that refuses connections to addresses that are not publicly
// routable. The check runs against the resolved IP of every dial, so DNS answers and redirects cannot
// reach internal hosts; environment proxies are disabled since dialing a proxy would bypass the check
func PublicOnly() httpclient.Option {
	return httpclient.TransportOption(func(t *http.Transport) error {
		t.Proxy = nil
		t.DialContext = (&net.Dialer{
			Timeout:   publicDialTimeout,
			KeepAlive: publicDialKeepAlive,
			Control:   publicDialControl,
		}).DialContext

		return nil
	})
}

// publicDialControl rejects dials whose resolved address is not publicly routable
func publicDialControl(_, address string, _ syscall.RawConn) error {
	addrPort, err := netip.ParseAddrPort(address)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNonPublicHost, err)
	}

	if !IsPublicAddr(addrPort.Addr()) {
		return fmt.Errorf("%w: %s", ErrNonPublicHost, addrPort.Addr())
	}

	return nil
}
