package domainscan

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/browser_rendering"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/theopenlane/httpsling"
	"golang.org/x/sync/errgroup"
)

const (
	domainScanTTL = 600 // 10 minutes in seconds
	// browserNavigationTimeout is the Puppeteer-level navigation timeout in milliseconds, controlling how long the browser waits for the waitUntil condition.
	// The default Puppeteer timeout is 30000ms; 45000ms provides additional headroom for slow initial page loads.
	browserNavigationTimeout = 45000
	// browserWaitUntil is the Puppeteer navigation completion strategy: "load" fires once HTML, CSS, and scripts are parsed,
	// which is more reliable for SPAs than "networkidle2" (some never drop below 2 open connections).
	browserWaitUntil = "load"
	// browserWaitForTimeout is an additional delay, in milliseconds, held  after the page load event fires and before extraction runs
	browserWaitForTimeout = 3000
	// trustCenterWaitForSelectorTimeout bounds how long we wait for trust center content to mount client-side before giving up and extracting
	// whatever is present; bestAttempt ensures this never hard-fails the request.
	trustCenterWaitForSelectorTimeout = 8000
	// browserRendering422MaxAttempts is the number of times to retry a browser rendering request after a 422 response, which Cloudflare
	// returns transiently when the target page fails to render (timeouts, blocked resources, etc.) rather than for a malformed request.
	browserRendering422MaxAttempts = 3
	// browserRendering422BaseDelay is the base delay between 422 retries
	browserRendering422BaseDelay = time.Second
)

// trustCenterContentSelector is a best-effort CSS selector for the content
// container on common trust center pages commonly render their compliance content into elements matching
// one of these class/id name fragments, or a <main> element once hydrated)
const trustCenterContentSelector = `main, [class*="control" i], [class*="framework" i], [class*="compliance" i], [class*="security" i], [id*="trust" i]`

// trustCenterSubpaths are common paths appended to a trust center's root URL and each fetched independently
var trustCenterSubpaths = []string{"", "controls", "compliance", "security", "documents", "subprocessors"}

// companyProfileSubpaths are common marketing/product pages fetched alongside the homepage when building a company
// profile, since details are frequently only mentioned on a dedicated page rather than the homepage itself
var companyProfileSubpaths = []string{"company", "pricing", "security", "legal", "contact", "about", "features", "platform", "docs"}

// Config holds the Cloudflare credentials used for browser rendering and browser-derived enrichment lookups
type Config struct {
	// APIToken use to authenticate into with the cloudflare API
	APIToken string
	// AccountID the APIToken is associated with
	AccountID string
}

// clientOptions builds the request options shared by every Cloudflare API call this package makes
func (c *Config) clientOptions() []option.RequestOption {
	return []option.RequestOption{
		option.WithAPIToken(c.APIToken),
		option.WithHeader(httpsling.HeaderContentType, httpsling.ContentTypeJSON),
	}
}

// GetComplianceData fetches compliance information from the given domain and, if it can derive one, from a trust.<domain> subdomain as well
func (c *Config) GetComplianceData(ctx context.Context, domain string) (*CompliancePage, error) {
	comp, err := c.fetchCompliancePage(ctx, domain)
	if err != nil {
		return nil, err
	}

	candidates, ok := trustCenterURLs(domain)
	if !ok {
		return comp, nil
	}

	for _, trustURL := range candidates {
		trustCenter, err := c.fetchTrustCenterPages(ctx, trustURL)
		if err != nil {
			// this candidate subdomain may not exist, not an actual error
			continue
		}

		return mergeTrustCenterIntoCompliancePage(comp, trustCenter, trustURL), nil
	}

	return comp, nil
}

// fetchCompliancePage runs the compliance prompt against a single URL and unmarshals the structured result into a CompliancePage
func (c *Config) fetchCompliancePage(ctx context.Context, url string) (*CompliancePage, error) {
	resp, err := c.browserRendering(ctx, url, promptCompliance, "")
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	comp := &CompliancePage{}
	if err := json.Unmarshal(data, comp); err != nil {
		return nil, err
	}

	return comp, nil
}

// fetchTrustCenterPages probes the trust center's root URL and a handful of
// common subpaths and merges everything it can reach into a single TrustCenterPage
func (c *Config) fetchTrustCenterPages(ctx context.Context, trustURL string) (*TrustCenterPage, error) {
	if resolved := resolveRedirectTarget(ctx, trustURL); resolved != trustURL {
		trustURL = resolved
	}

	var pages []*TrustCenterPage

	var firstErr error

	for _, sub := range trustCenterSubpaths {
		pageURL := strings.TrimRight(trustURL, "/")
		if sub != "" {
			pageURL += "/" + sub
		}

		tc, err := c.fetchTrustCenterPage(ctx, pageURL)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}

			continue
		}

		pages = append(pages, tc)
	}

	if len(pages) == 0 {
		return nil, firstErr
	}

	return mergeTrustCenterPages(pages...), nil
}

// fetchTrustCenterPage runs the trust center prompt against a single URL and
// unmarshals the structured result into a TrustCenterPage
func (c *Config) fetchTrustCenterPage(ctx context.Context, url string) (*TrustCenterPage, error) {
	resp, err := c.browserRendering(ctx, url, promptTrustCenter, "")
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	tc := &TrustCenterPage{}
	if err := json.Unmarshal(data, tc); err != nil {
		return nil, err
	}

	return tc, nil
}

// discoverSystemSubdomains probes systemSubdomainCandidates under url's apex domain for HTTP
// reachability concurrently, returning whichever resolve as confirmed evidence of real systems
func discoverSystemSubdomains(ctx context.Context, url string) []string {
	apex, ok := apexDomain(url)
	if !ok {
		return nil
	}

	resolved := make([]string, len(systemSubdomainCandidates))

	var g errgroup.Group

	for i, label := range systemSubdomainCandidates {
		g.Go(func() error {
			if final, ok := urlReachable(ctx, "https://"+label+"."+apex); ok {
				resolved[i] = final
			}

			return nil
		})
	}

	_ = g.Wait() // per-candidate failures just mean that subdomain isn't reachable; never fails overall

	found := make([]string, 0, len(resolved))

	for _, r := range resolved {
		if r != "" {
			found = append(found, r)
		}
	}

	return found
}

// GetCompanyData builds a company profile from url's homepage, merged with whichever of companyProfileSubpaths
// resolve on the same host
func (c *Config) GetCompanyData(ctx context.Context, url string) (*CompanyProfile, error) {
	var promptSuffix string

	if subdomains := discoverSystemSubdomains(ctx, url); len(subdomains) > 0 {
		promptSuffix = "\n\nThe following subdomains were confirmed reachable for this company (not just linked from the page, but actually verified live): " +
			strings.Join(subdomains, ", ") +
			". Treat each as strong evidence of a real, separate system and factor them into the systems list even if they aren't mentioned or linked anywhere on the rendered page."
	}

	homepage, err := c.fetchCompanyProfilePage(ctx, url, promptSuffix)
	if err != nil {
		return nil, err
	}

	pages := make([]*CompanyProfile, len(companyProfileSubpaths))

	var g errgroup.Group

	for i, sub := range companyProfileSubpaths {
		g.Go(func() error {
			pageURL, ok := subpathURL(url, sub)
			if !ok {
				return nil
			}

			resolved, reachable := urlReachable(ctx, pageURL)
			if !reachable {
				return nil
			}

			page, err := c.fetchCompanyProfilePage(ctx, resolved, "")
			if err != nil {
				// this subpath may have failed to render, not an actual error
				return nil
			}

			pages[i] = page

			return nil
		})
	}

	_ = g.Wait() // per-subpath failures just mean that page isn't reachable or didn't render; never fails overall

	profile := mergeCompanyProfiles(append([]*CompanyProfile{homepage}, pages...)...)

	if profile.StatusPageURL == "" {
		if candidate, ok := statusPageURL(url); ok {
			if resolved, exists := urlReachable(ctx, candidate); exists {
				profile.StatusPageURL = resolved
			}
		}
	}

	return profile, nil
}

// fetchCompanyProfilePage runs the company profile prompt against a single URL and unmarshals the structured result into a CompanyProfile
func (c *Config) fetchCompanyProfilePage(ctx context.Context, url, promptSuffix string) (*CompanyProfile, error) {
	resp, err := c.browserRendering(ctx, url, promptCompany, promptSuffix)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	profile := &CompanyProfile{}
	if err := json.Unmarshal(data, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// BrowserRendering will take a snapshot of a provided domain and return the extracted company profile.
// promptSuffix, if non-empty, is appended to the base prompt for kind, letting callers hand the model
// verified facts (e.g. confirmed-reachable subdomains) it can't discover from the rendered page alone
func (c *Config) browserRendering(ctx context.Context, target string, kind PromptType, promptSuffix string) (*browser_rendering.JsonNewResponse, error) {
	client := cloudflare.NewClient(c.clientOptions()...)

	if parsed, ok := normalizeURL(target); ok {
		target = parsed.String()
	}

	params := c.getBrowserRenderingJSONParams(target, getPrompt(kind)+promptSuffix, kind)

	var lastErr error

	for attempt := range browserRendering422MaxAttempts {
		if attempt > 0 {
			delay := browserRendering422BaseDelay * time.Duration(1<<(attempt-1))

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := client.BrowserRendering.Json.New(ctx, params)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		var apiErr *cloudflare.Error
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnprocessableEntity {
			return nil, err
		}
	}

	return nil, lastErr
}

// gotoOptions controls page navigation behavior in the browser rendering API
type gotoOptions struct {
	// WaitUntil controls when navigation is considered complete.
	// See browserWaitUntil for the selected strategy.
	WaitUntil string `json:"waitUntil,omitempty"`
	// Timeout is the navigation timeout in milliseconds; Puppeteer defaults to
	// 30000, set higher for heavy SPAs that take longer to reach network idle.
	Timeout int `json:"timeout,omitempty"`
}

// waitForSelectorOptions tells the browser rendering API to wait for a CSS
// selector to appear before extraction runs, for content that mounts
// client-side after the initial page load
type waitForSelectorOptions struct {
	// Selector is the CSS selector to wait for
	Selector string `json:"selector"`
	// Timeout is how long to wait for the selector, in milliseconds
	Timeout int `json:"timeout,omitempty"`
}

// getBrowserRenderingJSONParams gets the params required for the browser rendering JSON request
func (c *Config) getBrowserRenderingJSONParams(url string, prompt string, kind PromptType) browser_rendering.JsonNewParams {
	params := browser_rendering.JsonNewParams{
		AccountID: cloudflare.F(c.AccountID),
		CacheTTL:  cloudflare.Float(domainScanTTL),
	}

	var schema ResponseFormat

	switch kind {
	case promptCompliance:
		schema = buildCompliancePageSchema()
	case promptTrustCenter:
		schema = buildTrustCenterPageSchema()
	default:
		schema = buildCompanyProfileSchema()
	}

	body := browser_rendering.JsonNewParamsBody{
		URL:            cloudflare.String(url),
		Prompt:         cloudflare.String(prompt),
		ResponseFormat: cloudflare.F[any](schema),
		GotoOptions:    cloudflare.F[any](gotoOptions{WaitUntil: browserWaitUntil, Timeout: browserNavigationTimeout}),
		WaitForTimeout: cloudflare.Float(browserWaitForTimeout),
	}

	if kind == promptTrustCenter {
		body.WaitForSelector = cloudflare.F[any](waitForSelectorOptions{
			Selector: trustCenterContentSelector,
			Timeout:  trustCenterWaitForSelectorTimeout,
		})
		body.BestAttempt = cloudflare.Bool(true)
	}

	params.Body = body

	return params
}

// buildCompanyProfileSchema constructs the JSON schema for company profile extraction
func buildCompanyProfileSchema() ResponseFormat {
	return ResponseFormat{
		Type: "json_schema",
		Schema: JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchemaProperty{
				"name": {
					Type:        "string",
					Description: "The company or organization name",
				},
				"description": {
					Type:        "string",
					Description: "A brief description of what the company does, in 1-2 sentences",
				},
				"industry": {
					Type:        "string",
					Description: "The primary industry the company operates in",
				},
				"systems": {
					Type:        "array",
					Description: "Typically 2-5 distinct technical surfaces that make up the company's own product infrastructure, e.g. a web console/dashboard, public API, mobile app, CLI tool, or storage/database backend. Do NOT include the company's product modules, capabilities, or named marketing features as separate entries (e.g. things like 'Compliance Automation', 'Policy Management', 'Frameworks', 'Trust Center', 'Registry', 'Reporting', or similar named offerings are features within one web console, not separate systems, and must be excluded). A company with a single product should usually yield a single system, not one per feature it markets.",
					Items: &JSONSchemaProperty{
						Type:        "object",
						Description: "A single technical system, not a product module, capability, or marketing feature name",
						Properties: map[string]JSONSchemaProperty{
							"name":             {Type: "string", Description: "The system name (e.g. Console, API, Mobile App, Storage Backend) — never a marketing feature or module name"},
							"summary":          {Type: "string", Description: "A brief, 1-2 sentence description of what this system does"},
							"full_description": {Type: "string", Description: "A more thorough description of what this system does and what data or functionality it handles, drawn from documentation, architecture pages, or other technical content when available"},
						},
					},
				},
				"location": {
					Type:        "string",
					Description: "The headquarters or primary location of the company",
				},
				"employee_range": {
					Type:        "string",
					Description: "The approximate employee count range, such as 1-10, 11-50, 51-200, 201-500, 501-1000, 1001-5000, 5000+",
				},
				"founded_year": {
					Type:        "string",
					Description: "The year the company was founded or established, if mentioned on the website (e.g., 2015, 2020). Leave empty if not found.",
				},
				"estimated_revenue": {
					Type:        "string",
					Description: "The estimated annual revenue range if mentioned or inferable from the website (e.g., $1M-$10M, $10M-$50M, $50M-$100M, $100M+). Leave empty if not discoverable.",
				},
				"social_links": {
					Type:        "object",
					Description: "URLs for the company's social media and community profiles found in the website header, footer, or about page",
					Properties: map[string]JSONSchemaProperty{
						"linkedin":  {Type: "string", Description: "LinkedIn company page URL"},
						"twitter":   {Type: "string", Description: "Twitter/X profile URL"},
						"github":    {Type: "string", Description: "GitHub organization URL"},
						"discord":   {Type: "string", Description: "Discord server invite URL"},
						"instagram": {Type: "string", Description: "Instagram profile URL"},
						"youtube":   {Type: "string", Description: "YouTube channel URL"},
						"facebook":  {Type: "string", Description: "Facebook page URL"},
					},
				},
				"customers": {
					Type:        "array",
					Description: "Named customers, clients, or case study companies mentioned on the website (e.g., in logos, testimonials, or case studies). Only include company or organization names, not individual people.",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A customer or client company name",
					},
				},
				"technologies": {
					Type:        "array",
					Description: "Third-party SaaS tools, platforms, analytics services, and technology vendors detectable on the website (e.g., Google Analytics, Salesforce, HubSpot, Cloudflare, Intercom, Stripe, Segment, Zendesk). Only include vendor or product names, not web standards or protocols.",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A technology vendor or SaaS platform name",
					},
				},
				"status_page_url": {
					Type:        "string",
					Description: "The URL of the company's public status/uptime page, if linked anywhere on the site (e.g. status.<domain>, or a hosted platform like Statuspage.io, BetterStack, Instatus, or incident.io). Leave empty if none is found.",
				},
				"subdomain_links": {
					Type:        "array",
					Description: "URLs found in the page's navigation, footer, or body that point to other subdomains of this same company's domain (e.g. console.<domain>, app.<domain>, docs.<domain>, dashboard.<domain>) — the company's own other products or sections, not third-party vendor links.",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A URL pointing to another subdomain of this same company's domain",
					},
				},
				"sso_supported": {
					Type:        "boolean",
					Description: "True if the site advertises single sign-on (SSO) support for its product, false otherwise",
				},
				"mfa_supported": {
					Type:        "boolean",
					Description: "True if the site advertises multi-factor authentication (MFA) support for its product, false otherwise",
				},
			},
		},
	}
}

// buildCompliancePageSchema constructs the JSON schema for compliance page extraction
func buildCompliancePageSchema() ResponseFormat {
	return ResponseFormat{
		Type: "json_schema",
		Schema: JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchemaProperty{
				"url": {
					Type:        "string",
					Description: "The page URL that was analyzed",
				},
				"page_type": {
					Type:        "string",
					Description: "The type of compliance document: privacy_policy, terms_of_service, trust_center, dpa, soc2_report, security, subprocessors, gdpr, cookie_policy, or other",
				},
				"title": {
					Type:        "string",
					Description: "The page title",
				},
				"summary": {
					Type:        "string",
					Description: "A brief summary of the page content",
				},
				"frameworks": {
					Type:        "array",
					Description: "Compliance frameworks or certifications the company claims to have achieved (e.g., SOC 2 Type II, ISO 27001, GDPR, HIPAA, PCI DSS)",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A compliance framework or certification name",
					},
				},
				"soc2_certified": {
					Type:        "boolean",
					Description: "True if the page claims the company is SOC 2 (Type I or Type II) certified or compliant, false otherwise",
				},
				"last_updated": {
					Type:        "string",
					Description: "The last updated or effective date mentioned on the page",
				},
				"download_links": {
					Type:        "array",
					Description: "URLs to downloadable documents or reports found on the page",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A URL to a downloadable document or report",
					},
				},
				"subprocessors": {
					Type:        "array",
					Description: "Names of any subprocessors or third-party vendors listed on the page",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A subprocessor or third-party vendor name",
					},
				},
				"controls": {
					Type:        "array",
					Description: "Individual, concrete security or operational practices that this specific page actually states the company performs (for example, statements about access control, encryption, monitoring, testing, or training — the exact wording should reflect what's on the page, not a generic list). Do NOT put framework, certification, or compliance-standard names or statements here (e.g., do not include \"SOC 2 Certified\", \"ISO 27001 Certified\", \"HIPAA Compliant\") — those belong only in frameworks and soc2_certified. Return an empty array if the page does not explicitly describe any such practices; do not guess or fill in typical examples.",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A single concrete security or operational practice, not a framework or certification name",
					},
				},
				"compliance_links": {
					Type:        "array",
					Description: "URLs to other compliance documents such as privacy policies, terms of service, trust center pages, data processing agreements, or security pages found on this page",
					Items: &JSONSchemaProperty{
						Type:        "object",
						Description: "A single compliance-related link, categorized by document type",
						Properties: map[string]JSONSchemaProperty{
							"url":  {Type: "string", Description: "The link target"},
							"type": {Type: "string", Description: "The type of document the link points to: privacy_policy, terms_of_service, trust_center, dpa, soc2_report, security, subprocessors, gdpr, cookie_policy, or other"},
						},
					},
				},
			},
		},
	}
}

// buildTrustCenterPageSchema constructs the JSON schema for trust center page extraction
func buildTrustCenterPageSchema() ResponseFormat {
	return ResponseFormat{
		Type: "json_schema",
		Schema: JSONSchema{
			Type: "object",
			Properties: map[string]JSONSchemaProperty{
				"hosted_by": {
					Type:        "string",
					Description: "The platform or vendor hosting this trust center. Look for any \"Powered by\", footer, or page-title attribution naming the platform (e.g. Vanta, Drata, SafeBase, Whistic, Conveyor, TrustArc, OneTrust, SecurityPal, and others not listed here) rather than only matching these examples; use \"self-hosted\" only if no such attribution appears anywhere on the page.",
				},
				"frameworks": {
					Type:        "array",
					Description: "Compliance frameworks or certifications listed (e.g., SOC 2 Type II, ISO 27001, ISO 42001, GDPR, HIPAA, PCI DSS)",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A compliance framework or certification name",
					},
				},
				"soc2_certified": {
					Type:        "boolean",
					Description: "True if the trust center claims the company is SOC 2 (Type I or Type II) certified or compliant, false otherwise",
				},
				"controls": {
					Type:        "array",
					Description: "Individual, concrete security or operational practices that this specific page actually states the company performs (for example, statements about access control, encryption, monitoring, testing, or training — the exact wording should reflect what's on the page, not a generic list). Do NOT put framework, certification, or compliance-standard names or statements here (e.g., do not include \"SOC 2 Certified\", \"ISO 27001 Certified\", \"HIPAA Compliant\") — those belong only in frameworks and soc2_certified. Return an empty array if the page does not explicitly describe any such practices; do not guess or fill in typical examples.",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A single concrete security or operational practice, not a framework or certification name",
					},
				},
				"subprocessors": {
					Type:        "array",
					Description: "Names of any subprocessors or third-party vendors listed",
					Items: &JSONSchemaProperty{
						Type:        "string",
						Description: "A subprocessor or third-party vendor name",
					},
				},
			},
		},
	}
}
