package domainscan

import (
	"context"
	"net/http"
	"sync"

	"github.com/theopenlane/httpsling"
	"golang.org/x/net/publicsuffix"

	"github.com/theopenlane/core/pkg/urlx"
)

// userAgent identifies this scanner to servers it probes
const userAgent = "theopenlane-domainscan/1.0"

// trustCenterCandidateSubdomains are subdomain prefixes commonly used for a
// company's trust/security/compliance portal, tried in this order.
var trustCenterCandidateSubdomains = []string{"trust", "security", "compliance"}

// apexDomain returns the registrable (eTLD+1) domain for rawURL, e.g.
// "www.mail.example.co.uk" -> "example.co.uk"
func apexDomain(rawURL string) (string, bool) {
	parsed, err := urlx.Parse(rawURL)
	if err != nil {
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
	parsed, err := urlx.Parse(rawURL)
	if err != nil {
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
	parsed, err := urlx.Parse(rawURL)
	if err != nil {
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

// urlReachable does a lightweight HEAD request to rawURL and reports whether
// it resolves to a non-error response, returning the final URL after any redirects.
func urlReachable(ctx context.Context, rawURL string) (string, bool) {
	requester, err := scanRequester()
	if err != nil {
		return "", false
	}

	resp, err := requester.SendWithContext(ctx, httpsling.Head(rawURL))
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
