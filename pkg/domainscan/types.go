package domainscan

// JSONSchemaProperty describes a single property in a JSON schema
type JSONSchemaProperty struct {
	Type        string                        `json:"type"`
	Description string                        `json:"description,omitempty"`
	Items       *JSONSchemaProperty           `json:"items,omitempty"`
	Properties  map[string]JSONSchemaProperty `json:"properties,omitempty"`
}

// JSONSchema is the JSON schema for structured extraction
type JSONSchema struct {
	Type       string                        `json:"type"`
	Properties map[string]JSONSchemaProperty `json:"properties"`
}

// ResponseFormat specifies JSON schema extraction
type ResponseFormat struct {
	Type   string     `json:"type"`
	Schema JSONSchema `json:"json_schema"`
}

// CompanyProfile is the company information extracted from a website by
// the browser rendering AI
type CompanyProfile struct {
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Industry         string      `json:"industry"`
	Products         []Product   `json:"products"`
	Location         string      `json:"location"`
	EmployeeRange    string      `json:"employee_range"`
	FoundedYear      string      `json:"founded_year"`
	EstimatedRevenue string      `json:"estimated_revenue"`
	SocialLinks      SocialLinks `json:"social_links"`
	Customers        []string    `json:"customers"`
	Technologies     []string    `json:"technologies"`
	// StatusPageURL is the company's public status/uptime page, if found
	StatusPageURL string `json:"status_page_url,omitempty"`
	// SubdomainLinks are other subdomains of the same company's domain linked
	// from the page (e.g. console.<domain>, app.<domain>, docs.<domain>)
	SubdomainLinks []string `json:"subdomain_links,omitempty"`
	// SSOSupported indicates the company's product advertises SSO support
	SSOSupported bool `json:"sso_supported"`
	// MFASupported indicates the company's product advertises MFA support
	MFASupported bool `json:"mfa_supported"`
}

// Product describes a single product or service offered by the company
type Product struct {
	Name string `json:"name"`
	// Summary is a brief, 1-2 sentence description of the product
	Summary string `json:"summary"`
	// FullDescription is a more thorough description of the product, drawn
	// from documentation, feature pages, or other technical content when available
	FullDescription string `json:"full_description"`
}

// SocialLinks holds a company's social media and community profile URLs
type SocialLinks struct {
	LinkedIn  string `json:"linkedin"`
	Twitter   string `json:"twitter"`
	GitHub    string `json:"github"`
	Discord   string `json:"discord"`
	Instagram string `json:"instagram"`
	YouTube   string `json:"youtube"`
	Facebook  string `json:"facebook"`
}

// CompliancePage holds structured compliance information extracted from a single page
type CompliancePage struct {
	// URL is the page URL that was analyzed
	URL string `json:"url"`
	// PageType categorizes the compliance document (e.g., privacy_policy, terms_of_service, trust_center, dpa, soc2_report, security, subprocessors, gdpr, cookie_policy)
	PageType string `json:"page_type"`
	// Title is the page title
	Title string `json:"title"`
	// Summary is a brief description of the page content
	Summary string `json:"summary"`
	// Frameworks lists compliance frameworks or certifications mentioned (e.g., SOC 2, ISO 27001, GDPR, HIPAA)
	Frameworks []string `json:"frameworks"`
	// SOC2Certified indicates whether the page claims SOC 2 (Type I or Type II) certification or compliance
	SOC2Certified bool `json:"soc2_certified"`
	// LastUpdated is the last updated or effective date mentioned on the page
	LastUpdated string `json:"last_updated"`
	// DownloadLinks contains URLs for downloadable reports or documents found on the page
	DownloadLinks []string `json:"download_links"`
	// Subprocessors lists third-party vendors or sub-processors mentioned on the page
	Subprocessors []string `json:"subprocessors,omitempty"`
	// Controls lists individual security or compliance controls/practices mentioned, such as
	// those shown on a trust center page (e.g., MFA enforced, data encrypted at rest and in
	// transit, background checks performed, vulnerability scanning, penetration testing, incident response plan)
	Controls []string `json:"controls,omitempty"`
	// TrustCenterHostedBy is the platform serving the company's trust center, if
	// one was found (e.g., Vanta, Drata, SafeBase, Whistic, Conveyor, TrustArc,
	// or "self-hosted") — itself a vendor/technology signal for the company
	TrustCenterHostedBy string `json:"trust_center_hosted_by,omitempty"`
	// Documents lists compliance documents found on a trust center, each
	// flagged as public (directly accessible) or gated (requires a request/NDA)
	Documents []TrustDocument `json:"documents,omitempty"`
	// ComplianceLinks lists URLs to other compliance documents found on the page, each categorized by type
	ComplianceLinks []ComplianceLink `json:"compliance_links,omitempty"`
}

// ComplianceLink is a single compliance-related link found on a page, categorized by type
type ComplianceLink struct {
	// URL is the link target
	URL string `json:"url"`
	// Type categorizes the linked document (e.g., privacy_policy, terms_of_service, trust_center, dpa, soc2_report, security, subprocessors, gdpr, cookie_policy, or other)
	Type string `json:"type"`
}

// TrustDocument is a single compliance document or report listed on a trust center
type TrustDocument struct {
	// Name is the document title (e.g., "SOC 2 Type II Report")
	Name string `json:"name"`
	// URL is a direct link to the document, if one is available without requesting access
	URL string `json:"url,omitempty"`
	// Public indicates the document can be viewed or downloaded without requesting access or signing an NDA
	Public bool `json:"public"`
}

// TrustCenterPage holds structured information extracted from a trust center or trust portal page
type TrustCenterPage struct {
	// HostedBy is the platform or vendor hosting this trust center (e.g.,
	// Vanta, Drata, SafeBase, Whistic, Conveyor, TrustArc, OneTrust), or
	// "self-hosted" if it appears to be a custom-built page
	HostedBy string `json:"hosted_by"`
	// Frameworks lists compliance frameworks or certifications listed
	Frameworks []string `json:"frameworks"`
	// SOC2Certified indicates whether the trust center claims SOC 2 (Type I or Type II) certification or compliance
	SOC2Certified bool `json:"soc2_certified"`
	// Controls lists individual security or compliance controls or practices listed
	Controls []string `json:"controls"`
	// Documents lists compliance documents or reports listed, each flagged public or gated
	Documents []TrustDocument `json:"documents"`
	// Subprocessors lists third-party vendors or sub-processors mentioned
	Subprocessors []string `json:"subprocessors,omitempty"`
}

// DNSVendorInfo holds DNS-derived signals about a domain: mail routing/authentication at the
// apex, records on conventional subdomains (see commonSubdomains), and vendors those point to
type DNSVendorInfo struct {
	// SPFRecord is the raw SPF TXT record found at the domain's apex, if any
	SPFRecord string `json:"spf_record,omitempty"`
	// DMARCRecord is the raw DMARC TXT record found at _dmarc.<domain>, if any
	DMARCRecord string `json:"dmarc_record,omitempty"`
	// DMARCPolicy is the enforcement policy from the DMARC record's p= tag (none, quarantine, or reject)
	DMARCPolicy string `json:"dmarc_policy,omitempty"`
	// MXHosts lists the mail exchanger hostnames for the domain's apex, in preference order
	MXHosts []string `json:"mx_hosts,omitempty"`
	// NSHosts lists the authoritative nameserver hostnames for the domain's apex, revealing
	// the DNS hosting vendor (e.g. Cloudflare, Route 53, NS1)
	NSHosts []string `json:"ns_hosts,omitempty"`
	// TXTRecords lists every TXT record found at the domain's apex
	TXTRecords []string `json:"txt_records,omitempty"`
	// DKIMSelectors lists DKIM selector labels (see commonDKIMSelectors) found on the domain's apex
	DKIMSelectors []string `json:"dkim_selectors,omitempty"`
	// Subdomains holds MX, TXT, CNAME, and DKIM selector records found on conventional
	// mail/vendor subdomains (see commonSubdomains), probed up to two levels deep
	Subdomains []SubdomainDNSInfo `json:"subdomains,omitempty"`
	// CertSubdomains lists subdomains discovered from certificate transparency logs
	// (crt.sh) — a passive lookup against a public third-party log rather than a probe
	// of the target's own infrastructure, so these are inventory only and not otherwise
	// resolved or vendor-attributed
	CertSubdomains []string `json:"cert_subdomains,omitempty"`
	// Vendors lists vendors identified from the MX, SPF, TXT, CNAME, DKIM selector, and
	// NS records found at the apex and on Subdomains
	Vendors []DNSVendor `json:"vendors,omitempty"`
}

// DNSVendor is a single vendor identified from a DNS record. URL is only
// populated when the signal pointed at an actual hostname (MX, SPF include,
// or CNAME target); DKIM selector and TXT verification-tag signals carry a
// name only, since the selector/tag itself isn't a domain
type DNSVendor struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// SubdomainDNSInfo holds the DNS records found for a single conventional subdomain probed alongside the apex
type SubdomainDNSInfo struct {
	// Host is the fully probed hostname, e.g. "mail.example.com"
	Host string `json:"host"`
	// MXHosts lists the mail exchanger hostnames found for this subdomain, if any
	MXHosts []string `json:"mx_hosts,omitempty"`
	// TXTRecords lists the TXT records found for this subdomain, if any
	TXTRecords []string `json:"txt_records,omitempty"`
	// CNAME is the canonical name this subdomain resolves to, if it's an
	// alias to another domain (e.g. a hosted checkout, status page, or
	// support portal pointed at a vendor's own domain)
	CNAME string `json:"cname,omitempty"`
	// DKIMSelectors lists DKIM selector labels (see commonDKIMSelectors)
	// found configured on this subdomain, e.g. "resend" for
	// "resend._domainkey.mail.example.com"
	DKIMSelectors []string `json:"dkim_selectors,omitempty"`
}
