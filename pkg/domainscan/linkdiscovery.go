package domainscan

import (
	"context"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

// maxLinkScanBodyBytes caps the homepage read. A footer is in the served markup of any page
// that has one, so there is no reason to pull a whole single-page application bundle
const maxLinkScanBodyBytes int64 = 2 << 20

// hrefPattern captures the target of every anchor in the document, single or double quoted
var hrefPattern = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)

// linkFallbackPath is tried when the homepage names none of the documents, for the rare site
// that keeps its footer links on a legal hub instead of every page
const linkFallbackPath = "legal"

// GatherComplianceLinks finds a site's published legal documents by reading the links on its
// homepage, usually linked in the footer. domain may be a bare host or a URL
func GatherComplianceLinks(ctx context.Context, domain string) []ComplianceLink {
	homepage, ok := absoluteURL(domain)
	if !ok {
		return nil
	}

	if links := complianceLinksAt(ctx, homepage); len(links) > 0 {
		return links
	}

	fallback, ok := subpathURL(homepage, linkFallbackPath)
	if !ok {
		return nil
	}

	return complianceLinksAt(ctx, fallback)
}

// complianceLinksAt reads one page and returns the compliance documents its links name
func complianceLinksAt(ctx context.Context, pageURL string) []ComplianceLink {
	body, finalURL, ok := fetchBody(ctx, pageURL, maxLinkScanBodyBytes, nil)
	if !ok {
		return nil
	}

	return complianceLinksFromHTML(body, finalURL)
}

// complianceLinksFromHTML maps every anchor in the document onto a compliance document type,
// resolving relative targets against base and keeping the first link found for each type
func complianceLinksFromHTML(body, base string) []ComplianceLink {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil
	}

	seen := make(map[string]struct{})

	var links []ComplianceLink

	for _, match := range hrefPattern.FindAllStringSubmatch(body, -1) {
		href := strings.TrimSpace(match[1])

		docType := complianceTypeFromHref(href)
		if docType == "" {
			continue
		}

		if _, ok := seen[docType]; ok {
			continue
		}

		target, err := url.Parse(href)
		if err != nil {
			continue
		}

		seen[docType] = struct{}{}
		links = append(links, ComplianceLink{URL: baseURL.ResolveReference(target).String(), Type: docType})
	}

	return links
}

// complianceTypeFromHref resolves a link target to a compliance document type using its last
// path segment, so "/legal/cookie-policy" resolves to cookie_policy and "/legal/privacy" to
// privacy_policy via the same aliases the extraction output goes through
func complianceTypeFromHref(href string) string {
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "mailto:") {
		return ""
	}

	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}

	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")

	segment := segments[len(segments)-1]
	if segment == "" {
		return ""
	}

	segment = strings.TrimSuffix(segment, ".html")
	if !IsCanonicalComplianceType(segment) {
		return ""
	}

	return NormalizeComplianceType(segment)
}

// statusHostLabels are the subdomain labels a company serves its own status page under
var statusHostLabels = []string{"status", "uptime", "health"}

// statusPageHosts are the hosted status page providers, whose domains identify a status page
// wherever it is linked from
var statusPageHosts = []string{
	"statuspage.io",
	"instatus.com",
	"status.io",
	"statuspal.io",
	"betterstack.com",
	"uptime.com",
	"freshstatus.io",
	"sorryapp.com",
	"cachet.io",
	"openstatus.dev",
}

// GatherStatusPage finds a site's public status page.
//
// A status page is a host rather than a path, and almost always status.<domain>, so that is
// probed first: one request settles it for most companies, with no dependence on the page
// linking it anywhere. Scanning the homepage's links is the fallback, for the company on a
// hosted provider such as statuspage.io or one using a different label
func GatherStatusPage(ctx context.Context, domain string) string {
	homepage, ok := absoluteURL(domain)
	if !ok {
		return ""
	}

	if candidate, ok := statusPageURL(homepage); ok {
		if resolved, reachable := urlReachable(ctx, candidate); reachable {
			return resolved
		}
	}

	body, finalURL, ok := fetchBody(ctx, homepage, maxLinkScanBodyBytes, nil)
	if !ok {
		return ""
	}

	return statusPageFromHTML(body, finalURL)
}

// statusPageFromHTML returns the first link in the document whose host identifies a status page
func statusPageFromHTML(body, base string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}

	for _, match := range hrefPattern.FindAllStringSubmatch(body, -1) {
		href := strings.TrimSpace(match[1])
		if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "mailto:") {
			continue
		}

		target, err := url.Parse(href)
		if err != nil {
			continue
		}

		resolved := baseURL.ResolveReference(target)
		if isStatusPageHost(resolved.Hostname()) {
			return resolved.String()
		}
	}

	return ""
}

// isStatusPageHost reports whether host names a status page, either by its leading label or by
// belonging to a hosted status page provider
func isStatusPageHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if host == "" {
		return false
	}

	label, _, found := strings.Cut(host, ".")
	if found && slices.Contains(statusHostLabels, label) {
		return true
	}

	for _, provider := range statusPageHosts {
		if host == provider || strings.HasSuffix(host, "."+provider) {
			return true
		}
	}

	return false
}
