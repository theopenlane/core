package domainscan

import (
	"context"
	"net/http"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

// userAgent identifies this scanner to servers it probes
const userAgent = "theopenlane-domainscan/1.0"

// trustCenterCandidateSubdomains are subdomain prefixes commonly used for a
// company's trust/security/compliance portal, tried in this order.
var trustCenterCandidateSubdomains = []string{"trust", "security", "compliance"}

// normalizeURL parses rawURL, assuming an https:// scheme if none is given
func normalizeURL(rawURL string) (*url.URL, bool) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, false
	}

	if u.Scheme == "" {
		u, err = url.Parse("https://" + rawURL)
		if err != nil {
			return nil, false
		}
	}

	if u.Hostname() == "" {
		return nil, false
	}

	return u, true
}

// apexDomain returns the registrable (eTLD+1) domain for rawURL, e.g.
// "www.mail.example.co.uk" -> "example.co.uk"
func apexDomain(rawURL string) (string, bool) {
	parsed, ok := normalizeURL(rawURL)
	if !ok {
		return "", false
	}

	host, err := publicsuffix.EffectiveTLDPlusOne(parsed.Hostname())
	if err != nil {
		return "", false
	}

	return host, true
}

// trustCenterURLs derives candidate trust center URLs for rawURL, one per
// entry in trustCenterCandidateSubdomains (e.g. trust.<domain>, security.<domain>).
func trustCenterURLs(rawURL string) ([]string, bool) {
	parsed, ok := normalizeURL(rawURL)
	if !ok {
		return nil, false
	}

	host, err := publicsuffix.EffectiveTLDPlusOne(parsed.Hostname())
	if err != nil {
		return nil, false
	}

	urls := make([]string, 0, len(trustCenterCandidateSubdomains))

	for _, sub := range trustCenterCandidateSubdomains {
		u := *parsed
		u.Host = sub + "." + host
		u.Path = ""
		u.RawQuery = ""
		u.Fragment = ""

		urls = append(urls, u.String())
	}

	return urls, true
}

// statusPageURL derives a status.<domain> URL from the given domain
func statusPageURL(rawURL string) (string, bool) {
	parsed, ok := normalizeURL(rawURL)
	if !ok {
		return "", false
	}

	host, err := publicsuffix.EffectiveTLDPlusOne(parsed.Hostname())
	if err != nil {
		return "", false
	}

	status := *parsed
	status.Host = "status." + host
	status.Path = ""
	status.RawQuery = ""
	status.Fragment = ""

	return status.String(), true
}

// subpathURL returns the URL formed by pointing rawURL at path
func subpathURL(rawURL, path string) (string, bool) {
	parsed, ok := normalizeURL(rawURL)
	if !ok {
		return "", false
	}

	parsed.Path = "/" + path
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), true
}

// newHeadRequest builds a HEAD request against rawURL with the scanner's User-Agent set
func newHeadRequest(ctx context.Context, rawURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

// urlReachable does a lightweight HEAD request to rawURL and reports whether
// it resolves to a non-error response, returning the final URL after any redirects.
func urlReachable(ctx context.Context, rawURL string) (string, bool) {
	req, err := newHeadRequest(ctx, rawURL)
	if err != nil {
		return "", false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", false
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", false
	}

	if resp.Request != nil && resp.Request.URL != nil {
		return resp.Request.URL.String(), true
	}

	return rawURL, true
}

// resolveRedirectTarget follows rawURL's HTTP redirect chain via a lightweight HEAD request and returns the origin
func resolveRedirectTarget(ctx context.Context, rawURL string) string {
	req, err := newHeadRequest(ctx, rawURL)
	if err != nil {
		return rawURL
	}

	resp, err := http.DefaultClient.Do(req)
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
