package domainscan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/sync/errgroup"
)

// Enrichment holds the best-effort pkg/domainscan results for a single
// domain
type Enrichment struct {
	// Company is the base company profile
	Company *CompanyProfile
	// Compliance is compliance specific data such as soc2 attestation, subprocessors, etc
	Compliance *CompliancePage
	// DNS includes DNS probed data
	DNS *DNSVendorInfo
}

// EnrichmentErrors holds the per-lookup errors from GatherEnrichment, each nil on success
type EnrichmentErrors struct {
	Company    error
	Compliance error
	DNS        error
}

// GatherEnrichment runs the company profile, compliance, and DNS vendor
// lookups for domain concurrently, bounded by timeout. Each lookup is
// best-effort
func (c *Config) GatherEnrichment(ctx context.Context, domain string, timeout time.Duration) (Enrichment, EnrichmentErrors) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		enrichment Enrichment
		errs       EnrichmentErrors
		g          errgroup.Group
	)

	g.Go(func() error {
		if profile, err := c.GetCompanyData(ctx, domain); err != nil {
			errs.Company = err
		} else {
			enrichment.Company = profile
		}

		return nil
	})

	g.Go(func() error {
		if compliance, err := c.GetComplianceData(ctx, domain); err != nil {
			errs.Compliance = err
		} else {
			enrichment.Compliance = compliance
		}

		return nil
	})

	g.Go(func() error {
		if dnsInfo, err := GetDNSVendorInfo(ctx, domain); err != nil {
			errs.DNS = err
		} else {
			enrichment.DNS = dnsInfo
		}

		return nil
	})

	_ = g.Wait() // per-lookup errors are captured in errs above; this never fails

	return enrichment, errs
}

// vendorGroup accumulates every signal (wappalyzer detections, third-party
// request domains) that resolves to the same vendor
type vendorGroup struct {
	name       string
	url        string
	categories map[string]bool
}

// registrableDomain returns the root domain from a URL using the public
// suffix list, so vendor detections pointing at the same site can be grouped together
func registrableDomain(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil || u.Hostname() == "" {
		return ""
	}

	domain, ok := icannRegistrableDomain(u.Hostname())
	if !ok {
		return ""
	}

	return domain
}

// icannRegistrableDomain returns the registrable domain for host, using only ICANN (public)
// suffixes rather than the full public suffix list
func icannRegistrableDomain(host string) (string, bool) {
	if suffix, icann := publicsuffix.PublicSuffix(host); !icann {
		return suffix, suffix != ""
	}

	domain, err := publicsuffix.EffectiveTLDPlusOne(host)

	return domain, err == nil
}

// domainVendorName derives a display name from a registrable domain: an override
// from vendorDomainNames if the domain has one, otherwise capitalizing its first
// label, e.g. "google.com" -> "Google", falling back to vendorCanonicalNames if
// that naive label is itself a known alsoKnownAs alias (e.g. "hubspotemail.net" -> "Hubspot")
func domainVendorName(domain string) string {
	if override, ok := vendorDomainNames[domain]; ok {
		return override
	}

	label := strings.SplitN(domain, ".", 2)[0] //nolint:mnd
	if label == "" {
		return domain
	}

	if canonical, ok := vendorCanonicalNames[strings.ToLower(label)]; ok {
		return canonical
	}

	return strings.ToUpper(label[:1]) + label[1:]
}

// vendorNameForURL derives a vendor's grouping domain and display name from rawURL,
// preferring an exact vendorHostNames match (for hosts that share a registrable
// domain with another product, e.g. admin.google.com vs cloud.google.com) before
// collapsing to the registrable domain via domainVendorName
func vendorNameForURL(rawURL string) (name, domain string) {
	domain = registrableDomain(rawURL)
	if domain == "" {
		return "", ""
	}

	if u, err := url.Parse(rawURL); err == nil {
		if override, ok := vendorHostNames[strings.ToLower(u.Hostname())]; ok {
			return override, domain
		}
	}

	return domainVendorName(domain), domain
}

// vendorGroups accumulates vendor signals (wappalyzer detections, third-party
// request domains) keyed by registrable domain or vendor name, preserving
// first-seen order so the output list is stable
type vendorGroups struct {
	byKey  map[string]*vendorGroup
	byName map[string]*vendorGroup
	order  []string
}

func newVendorGroups() *vendorGroups {
	return &vendorGroups{byKey: map[string]*vendorGroup{}, byName: map[string]*vendorGroup{}}
}

// add folds one signal into the vendor group for key, creating the group if it doesn't exist yet
func (g *vendorGroups) add(key, name, url string, categories []string) {
	group, ok := g.byKey[key]

	if !ok {
		lowerName := strings.ToLower(name)

		if existing, ok := g.byName[lowerName]; ok {
			group = existing
			g.byKey[key] = group
		} else {
			group = &vendorGroup{name: name, url: url, categories: map[string]bool{}}
			g.byKey[key] = group
			g.byName[lowerName] = group
			g.order = append(g.order, key)
		}
	}

	if (group.url == "" || group.url == "Unknown") && url != "" && url != "Unknown" {
		group.url = url
	}

	for _, c := range categories {
		group.categories[c] = true
	}
}

// finalize converts the accumulated groups into the report's vendor list, in first-seen order
func (g *vendorGroups) finalize() []map[string]any {
	vendors := make([]map[string]any, 0, len(g.order))

	for _, key := range g.order {
		group := g.byKey[key]

		categories := make([]string, 0, len(group.categories))
		for c := range group.categories {
			categories = append(categories, c)
		}

		sort.Strings(categories)

		entry := map[string]any{"name": group.name}

		if group.url != "" {
			entry["url"] = group.url
		}

		if len(categories) > 0 {
			entry["categories"] = categories
		}

		vendors = append(vendors, entry)
	}

	return vendors
}

// ReportConfig configures how BuildScanReport classifies vendors versus
// technologies and which vendor names it always excludes
type ReportConfig struct {
	// NonVendorCategories lists wappalyzer categories treated as technologies
	// instead of vendors when building an onboarding domain scan report
	NonVendorCategories []string `json:"nonvendorcategories" koanf:"nonvendorcategories" default:"[Miscellaneous,JavaScript frameworks,JavaScript libraries,Static site generator]"`
	// DeniedVendorNames lists vendor names to always exclude from an onboarding domain scan report's vendor list
	DeniedVendorNames []string `json:"deniedvendornames" koanf:"deniedvendornames" default:"[rfc-editor,ajax,website-files,http/3,googletagmanager,cloudflareinsights,googlesyndication,gstatic,hcaptcha,googleapis,hsforms,hs-scripts,hscollectedforms,hsts]"`
	// ScanTTL is the cache TTL, in seconds, for Browser Rendering requests issued during domain scan enrichment
	ScanTTL int `json:"scanttl" koanf:"scanttl" default:"86400"`
}

// BuildScanReport combines a Cloudflare URL Scanner result with the Enrichment gathered by GatherEnrichment into a single report
// unified vendors/technologies, assets, findings, meta, platform, systems, and compliance sections
func BuildScanReport(result *url_scanner.ScanGetResponse, enrichment Enrichment, nonVendorCategories, deniedVendorNames []string) map[string]any {
	data := map[string]any{
		"external_scan_id": result.Task.UUID,
		"url":              result.Task.URL,
	}

	vendors, technologies := buildVendorsAndTechnologies(result, enrichment, nonVendorCategories, deniedVendorNames)
	if len(vendors) > 0 {
		data["vendors"] = vendors
	}

	if len(technologies) > 0 {
		data["technologies"] = technologies
	}

	if assets := buildAssets(result, enrichment); assets != nil {
		data["assets"] = assets
	}

	if trustCenter := buildTrustCenterSettings(result); trustCenter != nil {
		data["trust_center_settings"] = trustCenter
	}

	data["findings"] = buildFindings(result, enrichment)

	if meta := buildMeta(result); meta != nil {
		data["meta"] = meta
	}

	if platform := buildPlatform(enrichment); platform != nil {
		data["platform"] = platform
	}

	if systems := buildSystems(enrichment); len(systems) > 0 {
		data["systems"] = systems
	}

	if compliance := buildComplianceSection(enrichment); compliance != nil {
		data["compliance"] = compliance
	}

	return data
}

// buildVendorsAndTechnologies gets vendors and technologies from the scan combining Wappa
// data with third-party request domains, then folds in vendor signals from the domainscan enrichment
// deduped by the same keying, and drops any vendor matching deniedVendorNames
func buildVendorsAndTechnologies(result *url_scanner.ScanGetResponse, enrichment Enrichment, nonVendorCategories, deniedVendorNames []string) (vendors, technologies []map[string]any) {
	nonVendorCategorySet := make(map[string]bool, len(nonVendorCategories))
	for _, c := range nonVendorCategories {
		nonVendorCategorySet[strings.ToLower(c)] = true
	}

	deniedVendorNameSet := make(map[string]bool, len(deniedVendorNames))
	for _, name := range deniedVendorNames {
		deniedVendorNameSet[strings.ToLower(name)] = true
	}

	groups, technologies := groupWappaDetections(result.Meta.Processors.Wappa.Data, nonVendorCategorySet, deniedVendorNameSet)

	mergeRequestVendors(result.Data.Requests, result.Task.ApexDomain, groups)

	mergeEnrichmentVendors(enrichment, groups)

	vendors = filterDeniedVendors(groups.finalize(), deniedVendorNames)
	vendors = filterRedundantGoogle(vendors)

	return vendors, technologies
}

// filterRedundantGoogle drops the generic "Google" vendor entry when a more specific Google
// product (Google Workspace, Google Cloud, Google Drive, etc.) is also present in the list
func filterRedundantGoogle(vendors []map[string]any) []map[string]any {
	hasSpecificGoogle := false

	for _, vendor := range vendors {
		name, _ := vendor["name"].(string)
		if name != "Google" && strings.HasPrefix(name, "Google ") {
			hasSpecificGoogle = true
			break
		}
	}

	if !hasSpecificGoogle {
		return vendors
	}

	filtered := make([]map[string]any, 0, len(vendors))

	for _, vendor := range vendors {
		name, _ := vendor["name"].(string)
		if name == "Google" {
			continue
		}

		filtered = append(filtered, vendor)
	}

	return filtered
}

// maxVendorNameWords bounds how many words an LLM-extracted vendor/subprocessor name can have
// before it's treated as prose rather than an actual name and dropped
const maxVendorNameWords = 6

// looksLikeVendorName reports whether name is plausibly an actual company/product name rather
// than an explanatory sentence the extraction model returned in place of a real vendor list entry
func looksLikeVendorName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, "?") || strings.HasSuffix(name, "!") {
		return false
	}

	return len(strings.Fields(name)) <= maxVendorNameWords
}

// filterDeniedVendors drops any vendor whose name matches deniedVendorNames (case-insensitive)
func filterDeniedVendors(vendors []map[string]any, deniedVendorNames []string) []map[string]any {
	denied := make(map[string]bool, len(deniedVendorNames))
	for _, name := range deniedVendorNames {
		denied[strings.ToLower(name)] = true
	}

	filtered := make([]map[string]any, 0, len(vendors))

	for _, vendor := range vendors {
		name, _ := vendor["name"].(string)
		if denied[strings.ToLower(name)] {
			continue
		}

		filtered = append(filtered, vendor)
	}

	return filtered
}

// mergeEnrichmentVendors folds vendor signals from the domainscan enrichment
// into groups: detected technologies, compliance subprocessors (which already
// include a trust center's hosting vendor), a linked GitHub org, and DNS-derived vendors
func mergeEnrichmentVendors(enrichment Enrichment, groups *vendorGroups) {
	addNamedVendor := func(name, url string) {
		if name == "" {
			return
		}

		if canonical, ok := vendorCanonicalNames[strings.ToLower(name)]; ok {
			name = canonical
		}

		if url == "" {
			if domain, ok := vendorNameDomains[strings.ToLower(name)]; ok {
				url = "https://" + domain
			} else {
				url = "Unknown"
			}
		}

		groups.add("name:"+strings.ToLower(name), name, url, nil)
	}

	if enrichment.Company != nil {
		for _, tech := range enrichment.Company.Technologies {
			if looksLikeVendorName(tech) {
				addNamedVendor(tech, "")
			}
		}

		if enrichment.Company.SocialLinks.GitHub != "" {
			groups.add("name:github", "GitHub", "https://github.com", nil)
		}
	}

	if enrichment.Compliance != nil {
		for _, subprocessor := range enrichment.Compliance.Subprocessors {
			if looksLikeVendorName(subprocessor) {
				addNamedVendor(subprocessor, "")
			}
		}
	}

	if enrichment.DNS != nil {
		for _, vendor := range enrichment.DNS.Vendors {
			addNamedVendor(vendor.Name, vendor.URL)
		}
	}
}

// groupWappaDetections splits wappalyzer detections into vendor groups (keyed
// by registrable domain when known) and technologies, per nonVendorCategorySet
func groupWappaDetections(wappaData []url_scanner.ScanGetResponseMetaProcessorsWappaData, nonVendorCategorySet, deniedVendorNameSet map[string]bool) (groups *vendorGroups, technologies []map[string]any) {
	groups = newVendorGroups()
	technologies = make([]map[string]any, 0, len(wappaData))

	for _, w := range wappaData {
		if deniedVendorNameSet[strings.ToLower(w.App)] {
			continue
		}

		categories := make([]string, len(w.Categories))
		isTechnology := false

		for i, c := range w.Categories {
			categories[i] = c.Name
			if nonVendorCategorySet[strings.ToLower(c.Name)] {
				isTechnology = true
			}
		}

		if isTechnology {
			url := w.Website
			if url == "" {
				url = "Unknown"
			}

			technologies = append(technologies, map[string]any{
				"name":       w.App,
				"url":        url,
				"categories": categories,
			})

			continue
		}

		name := w.App
		if canonical, ok := vendorCanonicalNames[strings.ToLower(name)]; ok {
			name = canonical
		}

		if domain := registrableDomain(w.Website); domain != "" {
			groups.add(domain, name, "https://"+domain, categories)
		} else {
			groups.add("name:"+strings.ToLower(name), name, "Unknown", categories)
		}
	}

	return groups, technologies
}

// mergeRequestVendors supplements the wappalyzer-derived vendor groups with
// vendors identified from the domains actually contacted while the page loaded
func mergeRequestVendors(requests []url_scanner.ScanGetResponseDataRequest, apexDomain string, groups *vendorGroups) {
	for _, req := range requests {
		rawURL := req.Response.Response.URL
		if rawURL == "" {
			rawURL = req.Request.Request.URL
		}

		name, domain := vendorNameForURL(rawURL)
		if domain == "" || domain == apexDomain {
			continue
		}

		groups.add(domain, name, "https://"+domain, nil)
	}
}

// buildDNSRecords combines the scan's resolved A records with the domainscan enrichment's NS records for the scanned domain's apex
func buildDNSRecords(result *url_scanner.ScanGetResponse, enrichment Enrichment) []map[string]string {
	dnsRecords := make([]map[string]string, 0, len(result.Lists.Domains))

	for _, d := range result.Lists.Domains {
		dnsRecords = append(dnsRecords, map[string]string{"domain": d, "type": "A"})
	}

	if enrichment.DNS != nil {
		for _, ns := range enrichment.DNS.NSHosts {
			dnsRecords = append(dnsRecords, map[string]string{"domain": ns, "type": "NS"})
		}
	}

	return dnsRecords
}

// buildAssets reports the DNS records and IP addresses resolved during the scan, annotating each IP with its ASN/organization when known,
// plus any internal subdomains discovered by the domainscan enrichment
func buildAssets(result *url_scanner.ScanGetResponse, enrichment Enrichment) map[string]any {
	assets := map[string]any{}

	if dnsRecords := buildDNSRecords(result, enrichment); len(dnsRecords) > 0 {
		assets["dns_records"] = dnsRecords
	}

	if len(result.Lists.IPs) > 0 {
		asnByIP := make(map[string]url_scanner.ScanGetResponseMetaProcessorsASNData, len(result.Meta.Processors.ASN.Data))
		for _, a := range result.Meta.Processors.ASN.Data {
			asnByIP[a.IP] = a
		}

		ipAddresses := make([]map[string]string, 0, len(result.Lists.IPs))
		for _, ip := range result.Lists.IPs {
			entry := map[string]string{"address": ip}

			if a, ok := asnByIP[ip]; ok {
				entry["asn"] = a.ASN
				entry["org"] = a.Description
			}

			ipAddresses = append(ipAddresses, entry)
		}

		assets["ip_addresses"] = ipAddresses
	}

	if internalDomains := buildInternalDomains(enrichment); len(internalDomains) > 0 {
		assets["internal_domains"] = internalDomains
	}

	if len(assets) == 0 {
		return nil
	}

	return assets
}

// buildInternalDomains collects the subdomains the domainscan enrichment
// discovered (trust/security/compliance/status subdomains, subdomains linked
// from the company's own site, DNS-probed subdomains, certificate transparency
// subdomains, and hosts linked from compliance pages), deduped and sorted
func buildInternalDomains(enrichment Enrichment) []string {
	seen := map[string]bool{}

	var domains []string

	add := func(host string) {
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" || seen[host] {
			return
		}

		seen[host] = true

		domains = append(domains, host)
	}

	if enrichment.Company != nil && enrichment.Company.StatusPageURL != "" {
		if u, err := url.Parse(enrichment.Company.StatusPageURL); err == nil {
			add(u.Hostname())
		}
	}

	if enrichment.Company != nil {
		for _, link := range enrichment.Company.SubdomainLinks {
			if u, ok := normalizeURL(link); ok {
				add(u.Hostname())
			}
		}
	}

	if enrichment.Compliance != nil {
		for _, link := range enrichment.Compliance.ComplianceLinks {
			if u, err := url.Parse(link.URL); err == nil {
				add(u.Hostname())
			}
		}
	}

	if enrichment.DNS != nil {
		for _, sub := range enrichment.DNS.Subdomains {
			add(sub.Host)
		}
	}

	sort.Strings(domains)

	return domains
}

// buildTrustCenterSettings extracts the site favicon from the scan's page
// data, returning nil when no favicon was captured
func buildTrustCenterSettings(result *url_scanner.ScanGetResponse) map[string]any {
	type pageFavicon struct {
		URL  string `json:"url"`
		Hash string `json:"hash"`
	}

	type pageData struct {
		Favicon pageFavicon `json:"favicon"`
	}

	favicon := result.Page.JSON.RawJSON()
	if favicon == "" {
		return nil
	}

	var page pageData
	if err := json.Unmarshal([]byte(favicon), &page); err != nil || page.Favicon.Hash == "" {
		return nil
	}

	return map[string]any{
		"favicon": map[string]string{
			"url":  page.Favicon.URL,
			"hash": page.Favicon.Hash,
		},
	}
}

// buildFindings reports the scan's overall security verdict, any failing
// agent-readiness checks, and any expected compliance links that weren't found
func buildFindings(result *url_scanner.ScanGetResponse, enrichment Enrichment) map[string]any {
	findings := map[string]any{
		"security_violations": result.Verdicts.Overall.Categories,
		"risks":               result.Verdicts.Overall.Tags,
	}

	if result.Verdicts.Overall.Malicious {
		findings["is_malicious"] = true
	}

	if agentReadiness := buildAgentReadinessFindings(result.Meta.Processors.AgentReadiness); agentReadiness != nil {
		findings["agent_readiness"] = agentReadiness
	}

	if missing := buildMissingComplianceLinks(enrichment); missing != "" {
		findings["missing_compliance_links"] = missing
	}

	return findings
}

// expectedComplianceLinkTypes are the compliance document types a company
// site is generally expected to publish
var expectedComplianceLinkTypes = []string{"privacy_policy", "terms_of_service", "trust_center", "dpa", "security", "cookie_policy"}

// buildMissingComplianceLinks renders a GitHub-flavored Markdown task list, one unchecked
// item per expectedComplianceLinkTypes entry not found in the domainscan enrichment's
// compliance links, so it surfaces as one actionable checklist finding
func buildMissingComplianceLinks(enrichment Enrichment) string {
	if enrichment.Compliance == nil {
		return ""
	}

	found := make(map[string]bool, len(enrichment.Compliance.ComplianceLinks))
	for _, link := range enrichment.Compliance.ComplianceLinks {
		found[link.Type] = true
	}

	if enrichment.Compliance.PageType != "" {
		found[enrichment.Compliance.PageType] = true
	}

	items := make([]string, 0, len(expectedComplianceLinkTypes))

	for _, t := range expectedComplianceLinkTypes {
		if !found[t] {
			items = append(items, fmt.Sprintf("- [ ] %s", t))
		}
	}

	return strings.Join(items, "\n")
}

// buildAgentReadinessFindings reports the failing checks from the scan's agent-readiness assessment
// (e.g. missing markdown negotiation, no MCP server card)
func buildAgentReadinessFindings(processor url_scanner.ScanGetResponseMetaProcessorsAgentReadiness) map[string]any {
	raw := processor.JSON.RawJSON()
	if raw == "" {
		return nil
	}

	var parsed struct {
		Level     int64          `json:"level"`
		LevelName string         `json:"levelName"`
		Checks    map[string]any `json:"checks"`
	}

	if err := json.Unmarshal([]byte(raw), &parsed); err != nil || len(parsed.Checks) == 0 {
		return nil
	}

	failedChecks := []map[string]any{}
	walkAgentReadinessChecks(parsed.Checks, "", &failedChecks)

	if len(failedChecks) == 0 {
		return nil
	}

	sort.Slice(failedChecks, func(i, j int) bool {
		return failedChecks[i]["check"].(string) < failedChecks[j]["check"].(string)
	})

	return map[string]any{
		"level":      parsed.Level,
		"level_name": parsed.LevelName,
		"checklist":  buildAgentReadinessChecklistMarkdown(failedChecks),
		"reference":  agentReadinessReferenceURL,
	}
}

// agentReadinessReferenceURL links to Cloudflare's writeup of what the agent-readiness
// assessment measures and why, for context alongside the failed-check checklist
const agentReadinessReferenceURL = "https://blog.cloudflare.com/agent-readiness/"

// buildAgentReadinessChecklistMarkdown renders failedChecks as a single GitHub-flavored
// Markdown task list, one unchecked item per failing check, so the assessment surfaces
// as one finding instead of one per check
func buildAgentReadinessChecklistMarkdown(failedChecks []map[string]any) string {
	items := make([]string, 0, len(failedChecks))

	for _, c := range failedChecks {
		items = append(items, fmt.Sprintf("- [ ] %s", fmt.Sprint(c["message"])))
	}

	return strings.Join(items, "\n")
}

// walkAgentReadinessChecks recursively descends a generic agent-readiness check result
func walkAgentReadinessChecks(node map[string]any, path string, failedChecks *[]map[string]any) {
	status, hasStatus := node["status"].(string)
	message, hasMessage := node["message"].(string)

	if hasStatus && hasMessage {
		if status == "fail" {
			*failedChecks = append(*failedChecks, map[string]any{
				"check":   path,
				"message": message,
			})
		}

		return
	}

	for key, value := range node {
		child, ok := value.(map[string]any)
		if !ok {
			continue
		}

		childPath := key
		if path != "" {
			childPath = path + "." + key
		}

		walkAgentReadinessChecks(child, childPath, failedChecks)
	}
}

// buildMeta collects scan-level metadata (radar rank, categories, geolocation), returning nil when none of it was present in the scan
func buildMeta(result *url_scanner.ScanGetResponse) map[string]any {
	meta := map[string]any{}

	if len(result.Meta.Processors.RadarRank.Data) > 0 && result.Meta.Processors.RadarRank.Data[0].Rank > 0 {
		meta["rank"] = int(result.Meta.Processors.RadarRank.Data[0].Rank)
	}

	if len(result.Meta.Processors.URLCategories.Data) > 0 {
		urlCategories := make([]string, 0, len(result.Meta.Processors.URLCategories.Data))
		for _, c := range result.Meta.Processors.URLCategories.Data {
			urlCategories = append(urlCategories, c.Name)
		}

		meta["url_categories"] = urlCategories
	}

	if len(result.Meta.Processors.DomainCategories.Data) > 0 {
		domainCategories := make([]string, 0, len(result.Meta.Processors.DomainCategories.Data))
		for _, c := range result.Meta.Processors.DomainCategories.Data {
			domainCategories = append(domainCategories, c.Name)
		}

		meta["domain_categories"] = domainCategories
	}

	if len(result.Meta.Processors.Geoip.Data) > 0 {
		geo := result.Meta.Processors.Geoip.Data[0]

		meta["geolocation"] = map[string]any{
			"city":         geo.Geoip.City,
			"country":      geo.Geoip.Country,
			"country_name": geo.Geoip.CountryName,
			"region":       geo.Geoip.Region,
			"latitude":     geo.Geoip.Ll[0],
			"longitude":    geo.Geoip.Ll[1],
		}
	}

	if len(meta) == 0 {
		return nil
	}

	return meta
}

// buildPlatform shapes the scanned company's profile, using field names that mirror Openlane's Platform object (name/description)
func buildPlatform(enrichment Enrichment) map[string]any {
	company := enrichment.Company
	if company == nil {
		return nil
	}

	platform := map[string]any{
		"name":              company.Name,
		"description":       company.Description,
		"industry":          company.Industry,
		"location":          company.Location,
		"employee_range":    company.EmployeeRange,
		"founded_year":      company.FoundedYear,
		"estimated_revenue": company.EstimatedRevenue,
		"sso_supported":     company.SSOSupported,
		"mfa_supported":     company.MFASupported,
		"social_links":      company.SocialLinks,
	}

	if company.StatusPageURL != "" {
		platform["status_page_url"] = company.StatusPageURL
	}

	if len(company.Customers) > 0 {
		platform["customers"] = company.Customers
	}

	return platform
}

// buildSystems shapes each of the company's systems, using field names that
// mirror Openlane's SystemDetail object (system_name/description)
func buildSystems(enrichment Enrichment) []map[string]any {
	if enrichment.Company == nil {
		return nil
	}

	systems := make([]map[string]any, 0, len(enrichment.Company.Systems))

	for _, s := range enrichment.Company.Systems {
		description := s.FullDescription
		if description == "" {
			description = s.Summary
		}

		systems = append(systems, map[string]any{
			"system_name": s.Name,
			"description": description,
		})
	}

	return systems
}

// buildComplianceSection reports the company's compliance posture gathered by the domainscan enrichment
func buildComplianceSection(enrichment Enrichment) map[string]any {
	compliance := enrichment.Compliance
	if compliance == nil {
		return nil
	}

	section := map[string]any{
		"frameworks": compliance.Frameworks,
		"is_soc2":    compliance.SOC2Certified,
		"controls":   compliance.Controls,
	}

	if compliance.TrustCenterHostedBy != "" {
		section["trust_center_hosted_by"] = compliance.TrustCenterHostedBy
	}

	if len(compliance.Documents) > 0 {
		section["documents"] = compliance.Documents
	}

	return section
}
