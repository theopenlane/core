package domainscan

import (
	"context"
	"net/http"
	"net/url"
	"sync"

	"github.com/theopenlane/httpsling"
	"golang.org/x/net/publicsuffix"

	"github.com/theopenlane/core/v2/pkg/urlx"
)

// userAgent identifies this scanner to servers it probes
const userAgent = "theopenlane-domainscan/1.0"

// trustCenterCandidateSubdomains are subdomain prefixes commonly used for a
// company's trust/security/compliance portal, tried in this order.
var trustCenterCandidateSubdomains = []string{"trust", "security", "compliance"}

// apexDomain returns the registrable (eTLD+1) domain for rawURL, e.g.
// "www.mail.example.co.uk" -> "example.co.uk"
func apexDomain(rawURL string) (string, bool) {
	_, host, ok := parseApex(rawURL)

	return host, ok
}

// parseApex parses rawURL and returns it alongside its registrable domain
func parseApex(rawURL string) (*url.URL, string, bool) {
	parsed, err := urlx.Parse(rawURL)
	if err != nil {
		return nil, "", false
	}

	host, err := publicsuffix.EffectiveTLDPlusOne(parsed.Hostname())
	if err != nil {
		return nil, "", false
	}

	return parsed, host, true
}

// absoluteURL returns rawURL with a scheme, so a bare host such as "example.com" becomes
// "https://example.com" and can be requested
func absoluteURL(rawURL string) (string, bool) {
	parsed, err := urlx.Parse(rawURL)
	if err != nil {
		return "", false
	}

	return parsed.String(), true
}

// subdomainURL returns rawURL pointed at sub.<apex>, with any path, query and fragment
// dropped, e.g. subdomainURL("https://www.example.com/pricing", "status") ->
// "https://status.example.com"
func subdomainURL(rawURL, sub string) (string, bool) {
	parsed, host, ok := parseApex(rawURL)
	if !ok {
		return "", false
	}

	derived := *parsed
	derived.Host = sub + "." + host
	derived.Path = ""
	derived.RawQuery = ""
	derived.Fragment = ""

	return derived.String(), true
}

// trustCenterURLs derives candidate trust center URLs for rawURL, one per
// entry in trustCenterCandidateSubdomains (e.g. trust.<domain>, security.<domain>).
func trustCenterURLs(rawURL string) ([]string, bool) {
	urls := make([]string, 0, len(trustCenterCandidateSubdomains))

	for _, sub := range trustCenterCandidateSubdomains {
		derived, ok := subdomainURL(rawURL, sub)
		if !ok {
			return nil, false
		}

		urls = append(urls, derived)
	}

	return urls, true
}

// statusPageURL derives a status.<domain> URL from the given domain
func statusPageURL(rawURL string) (string, bool) {
	return subdomainURL(rawURL, "status")
}

// subpathURL returns the URL formed by pointing rawURL at path
func subpathURL(rawURL, path string) (string, bool) {
	parsed, err := urlx.Parse(rawURL)
	if err != nil {
		return "", false
	}

	parsed.Path = "/" + path
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), true
}

// scanRequester lazily builds the shared prober with the scanner User-Agent set,
// constructed once so probes reuse pooled connections
var scanRequester = sync.OnceValues(func() (*httpsling.Requester, error) {
	return urlx.NewRequester(httpsling.Header(httpsling.HeaderUserAgent, userAgent))
})

// urlReachable reports whether rawURL is worth fetching, returning the final URL after any
// redirects
func urlReachable(ctx context.Context, rawURL string) (string, bool) {
	if final, verdict := probeURL(ctx, httpsling.Head(rawURL), rawURL); verdict == probeFound {
		return final, true
	}

	final, verdict := probeURL(ctx, httpsling.Get(rawURL), rawURL)

	switch verdict {
	case probeFound:
		return final, true
	case probeAbsent:
		return "", false
	default:
		// blocked, refused or unreachable by this prober: let the renderer decide
		return rawURL, true
	}
}

// probeVerdict is what a single reachability probe concluded about a URL
type probeVerdict int

const (
	// probeFound means the origin served the URL
	probeFound probeVerdict = iota
	// probeAbsent means the origin answered and the page is not there
	probeAbsent
	// probeBlocked means we got no answer about the path: a refused method, bot protection,
	// a rate limit, a server error or a transport failure
	probeBlocked
)

// probeURL sends a single reachability probe and reports the final URL and its verdict
func probeURL(ctx context.Context, option httpsling.Option, rawURL string) (string, probeVerdict) {
	requester, err := scanRequester()
	if err != nil {
		return "", probeBlocked
	}

	resp, err := requester.SendWithContext(ctx, option)
	if err != nil {
		return "", probeBlocked
	}

	defer resp.Body.Close() //nolint:errcheck

	switch {
	case resp.StatusCode == http.StatusNotFound, resp.StatusCode == http.StatusGone:
		return "", probeAbsent
	case resp.StatusCode >= http.StatusBadRequest:
		return "", probeBlocked
	}

	if resp.Request != nil && resp.Request.URL != nil {
		return resp.Request.URL.String(), probeFound
	}

	return rawURL, probeFound
}

// contentTypeGuard reports whether a response's media type is one the caller will accept
type contentTypeGuard func(mediaType string) bool

// fetchBody GETs rawURL and returns its body as text along with the URL it was finally served
// from, so relative references resolve against the right origin after a redirect. ok is false
// for a transport error, a non-200 status, a body over maxBytes, or a media type the guard
// rejects.
func fetchBody(ctx context.Context, rawURL string, maxBytes int64, accept contentTypeGuard) (body, finalURL string, ok bool) {
	requester, err := scanRequester()
	if err != nil {
		return "", "", false
	}

	resp, err := requester.SendWithContext(ctx, httpsling.Get(rawURL))
	if err != nil {
		return "", "", false
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()

		return "", "", false
	}

	if ct := resp.Header.Get(httpsling.HeaderContentType); ct != "" && accept != nil && !accept(ct) {
		_ = resp.Body.Close()

		return "", "", false
	}

	raw, err := urlx.ReadBody(resp, urlx.MaxSizeValidator(maxBytes))
	if err != nil {
		return "", "", false
	}

	finalURL = rawURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	return string(raw), finalURL, true
}

// resolveRedirectTarget follows rawURL's HTTP redirect chain via a lightweight HEAD request and returns the origin
func resolveRedirectTarget(ctx context.Context, rawURL string) string {
	requester, err := scanRequester()
	if err != nil {
		return rawURL
	}

	resp, err := requester.SendWithContext(ctx, httpsling.Head(rawURL))
	if err != nil {
		return rawURL
	}

	defer resp.Body.Close()

	if resp.Request == nil || resp.Request.URL == nil {
		return rawURL
	}

	final := *resp.Request.URL
	final.Path = ""
	final.RawQuery = ""
	final.Fragment = ""

	return final.String()
}
