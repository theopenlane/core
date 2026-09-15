package domainscan

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func complianceLinksByType(links []ComplianceLink) map[string]string {
	out := make(map[string]string, len(links))

	for _, link := range links {
		out[link.Type] = link.URL
	}

	return out
}

func TestComplianceLinksFromHTML(t *testing.T) {
	// the real theopenlane.io footer, whose documents every earlier mechanism reported missing
	html := `<footer>
	<a class="footer-anchor" href="/legal/cookie-policy">Cookie Policy</a>
	<a class="footer-anchor" href="/legal/cookie-settings">Cookie Settings</a>
	<a class="footer-anchor" href="/legal/terms-of-service">Terms of Service</a>
	<a class="footer-anchor" href="/legal/privacy">Privacy</a>
	</footer>`

	got := complianceLinksByType(complianceLinksFromHTML(html, "https://www.theopenlane.io"))

	assert.Check(t, is.Equal("https://www.theopenlane.io/legal/privacy", got["privacy_policy"]))
	assert.Check(t, is.Equal("https://www.theopenlane.io/legal/terms-of-service", got["terms_of_service"]))
	assert.Check(t, is.Equal("https://www.theopenlane.io/legal/cookie-policy", got["cookie_policy"]))

	// cookie-settings is a preferences dialog rather than a published policy, so it must not
	// satisfy the cookie_policy check on its own
	assert.Check(t, is.Len(got, 3))
}

func TestComplianceLinksFromHTMLSpellingVariants(t *testing.T) {
	html := `
	<a href="/privacy-policy">Privacy</a>
	<a href='/terms-of-use'>Terms</a>
	<a href="https://example.com/legal/dpa">DPA</a>
	<a href="/legal/subprocessors">Subprocessors</a>
	<a href="/trust">Trust Center</a>
	<a href="/security">Security</a>
	<a href="/cookies">Cookies</a>`

	got := complianceLinksByType(complianceLinksFromHTML(html, "https://example.com"))

	for _, want := range []string{"privacy_policy", "terms_of_service", "dpa", "subprocessors", "trust_center", "security", "cookie_policy"} {
		assert.Check(t, got[want] != "", "expected a link of type %s", want)
	}
}

func TestComplianceLinksFromHTMLResolvesTargets(t *testing.T) {
	tests := []struct {
		name string
		html string
		base string
		want string
	}{
		{
			name: "relative to a directory",
			html: `<a href="privacy">p</a>`,
			base: "https://example.com/legal/",
			want: "https://example.com/legal/privacy",
		},
		{
			name: "root relative",
			html: `<a href="/legal/privacy">p</a>`,
			base: "https://example.com/anything",
			want: "https://example.com/legal/privacy",
		},
		{
			name: "absolute on another host",
			html: `<a href="https://trust.example.com/privacy">p</a>`,
			base: "https://example.com",
			want: "https://trust.example.com/privacy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := complianceLinksByType(complianceLinksFromHTML(tt.html, tt.base))
			assert.Check(t, is.Equal(tt.want, got["privacy_policy"]))
		})
	}
}

func TestComplianceTypeFromHref(t *testing.T) {
	tests := []struct {
		href string
		want string
	}{
		{href: "/legal/privacy", want: "privacy_policy"},
		{href: "/legal/privacy/", want: "privacy_policy"},
		{href: "/legal/privacy?lang=en", want: "privacy_policy"},
		{href: "/legal/privacy#your-rights", want: "privacy_policy"},
		{href: "/legal/privacy.html", want: "privacy_policy"},
		{href: "/Legal/Privacy-Policy", want: "privacy_policy"},
		{href: "/legal/terms-of-service", want: "terms_of_service"},
		{href: "/legal/cookie-policy", want: "cookie_policy"},
		// a preferences dialog, not a document
		{href: "/legal/cookie-settings", want: ""},
		// the segment has to be the document, not merely mention it
		{href: "/blog/privacy-is-good", want: ""},
		{href: "/pricing", want: ""},
		{href: "#top", want: ""},
		{href: "mailto:legal@example.com", want: ""},
		{href: "/legal", want: ""},
		{href: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.href, func(t *testing.T) {
			assert.Check(t, is.Equal(tt.want, complianceTypeFromHref(tt.href)))
		})
	}
}

func TestComplianceLinksFromHTMLKeepsFirstOfEachType(t *testing.T) {
	links := complianceLinksFromHTML(`<a href="/legal/privacy">a</a><a href="/privacy">b</a>`, "https://example.com")

	assert.Check(t, is.Len(links, 1))
	assert.Check(t, is.Equal("https://example.com/legal/privacy", links[0].URL))
}

func TestIsStatusPageHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{host: "status.theopenlane.io", want: true},
		{host: "status.example.com", want: true},
		{host: "uptime.example.com", want: true},
		{host: "health.example.com", want: true},
		{host: "STATUS.Example.com", want: true},
		{host: "status.example.com.", want: true},
		// hosted providers identify a status page wherever it is linked from
		{host: "acme.statuspage.io", want: true},
		{host: "acme.instatus.com", want: true},
		{host: "status.io", want: true},
		// a page about status is not a status page, and the label has to lead the host
		{host: "example.com", want: false},
		{host: "www.example.com", want: false},
		{host: "api.status.example.com", want: false},
		{host: "notstatuspage.io", want: false},
		{host: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			assert.Check(t, is.Equal(tt.want, isStatusPageHost(tt.host)))
		})
	}
}

func TestStatusPageFromHTML(t *testing.T) {
	// theopenlane.io links its status page behind live text rather than the word "status"
	openlane := `<footer><a href="https://status.theopenlane.io">All services online</a></footer>`
	assert.Check(t, is.Equal("https://status.theopenlane.io", statusPageFromHTML(openlane, "https://www.theopenlane.io")))

	hosted := `<a href="https://acme.statuspage.io/">Service status</a>`
	assert.Check(t, is.Equal("https://acme.statuspage.io/", statusPageFromHTML(hosted, "https://acme.com")))

	// nothing that looks like a status host
	assert.Check(t, is.Equal("", statusPageFromHTML(`<a href="/pricing">Pricing</a>`, "https://example.com")))
}
