package domainscan

import (
	"net/url"
	"sort"
	"strings"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
	"golang.org/x/net/publicsuffix"
)

// vendorGroup accumulates every signal (wappalyzer detections, third-party
// request domains) that resolves to the same vendor
type vendorGroup struct {
	name       string
	legalName  string
	url        string
	categories map[string]bool
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
func (g *vendorGroups) add(key, name, legalName, url string, categories []string) {
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

	if group.legalName == "" && legalName != "" && legalName != group.name {
		group.legalName = legalName
	}

	for _, c := range categories {
		group.categories[c] = true
	}
}

// finalize converts the accumulated groups into the report's vendor list, in first-seen order
func (g *vendorGroups) finalize() []Vendor {
	vendors := make([]Vendor, 0, len(g.order))

	for _, key := range g.order {
		group := g.byKey[key]

		var categories []string
		for c := range group.categories {
			categories = append(categories, c)
		}

		sort.Strings(categories)

		vendors = append(vendors, Vendor{Name: group.name, LegalName: group.legalName, URL: group.url, Categories: categories})
	}

	return vendors
}

// buildVendorsAndTechnologies gets vendors and technologies from the scan combining Wappa
// data with third-party request domains, then folds in vendor signals from the domainscan enrichment
// deduped by the same keying, and drops any vendor matching deniedVendorNames. result may be nil,
// in which case only the enrichment-derived vendors are included
func buildVendorsAndTechnologies(result *url_scanner.ScanGetResponse, enrichment Enrichment, nonVendorCategories, deniedVendorNames []string) (vendors []Vendor, technologies []Technology) {
	nonVendorCategorySet := make(map[string]bool, len(nonVendorCategories))
	for _, c := range nonVendorCategories {
		nonVendorCategorySet[strings.ToLower(c)] = true
	}

	deniedVendorNameSet := make(map[string]bool, len(deniedVendorNames))
	for _, name := range deniedVendorNames {
		deniedVendorNameSet[strings.ToLower(name)] = true
	}

	var groups *vendorGroups

	if result != nil {
		groups, technologies = groupWappaDetections(result.Meta.Processors.Wappa.Data, nonVendorCategorySet, deniedVendorNameSet)
		mergeRequestVendors(result.Data.Requests, result.Task.ApexDomain, groups)
	} else {
		groups = newVendorGroups()
	}

	mergeEnrichmentVendors(enrichment, groups)

	vendors = filterDeniedVendors(groups.finalize(), deniedVendorNames)
	vendors = filterRedundantGoogle(vendors)

	return vendors, technologies
}

// groupWappaDetections splits wappalyzer detections into vendor groups (keyed
// by registrable domain when known) and technologies, per nonVendorCategorySet
func groupWappaDetections(wappaData []url_scanner.ScanGetResponseMetaProcessorsWappaData, nonVendorCategorySet, deniedVendorNameSet map[string]bool) (groups *vendorGroups, technologies []Technology) {
	groups = newVendorGroups()
	technologies = make([]Technology, 0, len(wappaData))

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

			technologies = append(technologies, Technology{
				Name:       w.App,
				URL:        url,
				Categories: categories,
			})

			continue
		}

		name := w.App
		if canonical, ok := vendorCanonicalNames[strings.ToLower(name)]; ok {
			name = canonical
		}

		if domain := registrableDomain(w.Website); domain != "" {
			groups.add(domain, name, "", "https://"+domain, categories)
		} else {
			groups.add("name:"+strings.ToLower(name), name, "", "Unknown", categories)
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

		groups.add(domain, name, "", "https://"+domain, nil)
	}
}

// mergeEnrichmentVendors folds vendor signals from the domainscan enrichment
// into groups: detected technologies, compliance subprocessors (which already
// include a trust center's hosting vendor), a linked GitHub org, and DNS-derived vendors
func mergeEnrichmentVendors(enrichment Enrichment, groups *vendorGroups) {
	addNamedVendor := func(rawName, url string) {
		if rawName == "" {
			return
		}

		name := stripLegalEntitySuffixes(rawName)

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

		groups.add("name:"+strings.ToLower(name), name, rawName, url, nil)
	}

	if enrichment.Company != nil {
		for _, tech := range enrichment.Company.Technologies {
			if looksLikeVendorName(tech) {
				addNamedVendor(tech, "")
			}
		}

		if enrichment.Company.SocialLinks.GitHub != "" {
			groups.add("name:github", "GitHub", "", "https://github.com", nil)
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

// filterRedundantGoogle drops the generic "Google" vendor entry when a more specific Google
// product (Google Workspace, Google Cloud, Google Drive, etc.) is also present in the list
func filterRedundantGoogle(vendors []Vendor) []Vendor {
	hasSpecificGoogle := false

	for _, vendor := range vendors {
		if vendor.Name != "Google" && strings.HasPrefix(vendor.Name, "Google ") {
			hasSpecificGoogle = true
			break
		}
	}

	if !hasSpecificGoogle {
		return vendors
	}

	filtered := make([]Vendor, 0, len(vendors))

	for _, vendor := range vendors {
		if vendor.Name == "Google" {
			continue
		}

		filtered = append(filtered, vendor)
	}

	return filtered
}

// filterDeniedVendors drops any vendor whose name matches deniedVendorNames (case-insensitive)
func filterDeniedVendors(vendors []Vendor, deniedVendorNames []string) []Vendor {
	denied := make(map[string]bool, len(deniedVendorNames))
	for _, name := range deniedVendorNames {
		denied[strings.ToLower(name)] = true
	}

	filtered := make([]Vendor, 0, len(vendors))

	for _, vendor := range vendors {
		if denied[strings.ToLower(vendor.Name)] {
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

// RegistrableDomain returns the registrable domain (eTLD+1) for host, e.g. "app.hubspot.com" ->
// "hubspot.com", so a subdomain can be matched against the vendor that owns its root domain
func RegistrableDomain(host string) (string, bool) {
	return icannRegistrableDomain(host)
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
