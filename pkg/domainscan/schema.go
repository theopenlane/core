package domainscan

import (
	"github.com/theopenlane/core/pkg/jsonx"
)

// Vendor is one detected vendor entry in a scan report
type Vendor struct {
	// Name is the vendor's canonical display name
	Name string `json:"name"`
	// LegalName is the vendor's raw, un-canonicalized legal entity name (e.g. "Cloudflare, Inc"),
	// set only when it differs from Name
	LegalName string `json:"legal_name,omitempty"`
	// URL is the vendor's site or product URL, if known
	URL string `json:"url,omitempty"`
	// Categories are the vendor's detected categories
	Categories []string `json:"categories,omitempty"`
}

// Technology is one detected technology entry in a scan report
type Technology struct {
	// Name is the technology's display name
	Name string `json:"name"`
	// URL is the technology's site or product URL, if known
	URL string `json:"url,omitempty"`
	// Categories are the technology's detected categories
	Categories []string `json:"categories,omitempty"`
}

// IPAddress is one resolved IP entry in a scan report's assets section
type IPAddress struct {
	// Address is the resolved IP address
	Address string `json:"address"`
	// ASN is the autonomous system number announcing the address, if known
	ASN string `json:"asn,omitempty"`
	// Org is the organization associated with the ASN, if known
	Org string `json:"org,omitempty"`
}

// DNSRecord is one DNS record entry in a scan report's assets section
type DNSRecord struct {
	// Domain is the DNS record's hostname
	Domain string `json:"domain"`
	// Type is the DNS record type (e.g. "A", "NS")
	Type string `json:"type"`
	// Vendor is the vendor this hostname was attributed to, if known
	Vendor string `json:"vendor,omitempty"`
}

// Assets is the assets section of a scan report
type Assets struct {
	// DNSRecords are the DNS records resolved during the scan, plus any subdomains discovered by
	// the domainscan enrichment (type "internal")
	DNSRecords []DNSRecord `json:"dns_records,omitempty"`
	// IPAddresses are the IP addresses resolved during the scan
	IPAddresses []IPAddress `json:"ip_addresses,omitempty"`
}

// IsEmpty reports whether none of the assets sections were populated
func (a Assets) IsEmpty() bool {
	return len(a.DNSRecords) == 0 && len(a.IPAddresses) == 0
}

// AgentReadinessFinding is one domain's agent-readiness assessment finding.
// Domain is only set once the finding has been merged across every domain in a submission
type AgentReadinessFinding struct {
	// Level is the scan's agent-readiness score
	Level int64 `json:"level"`
	// LevelName is the human-readable name for Level (e.g. "Bot-Aware")
	LevelName string `json:"level_name"`
	// Checklist is a GitHub-flavored Markdown task list of the failing checks
	Checklist string `json:"checklist"`
	// Reference links to Cloudflare's writeup of what the assessment measures
	Reference string `json:"reference"`
	// Domain is the domain this finding was raised against, set only once merged across domains
	Domain string `json:"domain,omitempty"`
}

// Findings is the findings section of a scan report
type Findings struct {
	// SecurityViolations are the scan's overall verdict categories
	SecurityViolations []string `json:"security_violations,omitempty"`
	// Risks are the scan's overall verdict tags
	Risks []string `json:"risks,omitempty"`
	// IsMalicious reports whether the scan's overall verdict flagged the site as malicious
	IsMalicious bool `json:"is_malicious,omitempty"`
	// MissingComplianceLinks is a GitHub-flavored Markdown task list of expected compliance
	// document types not found on the site
	MissingComplianceLinks string `json:"missing_compliance_links,omitempty"`
	// AgentReadiness is the scan's agent-readiness assessment findings
	AgentReadiness []AgentReadinessFinding `json:"agent_readiness,omitempty"`
}

// Geolocation is the scanned site's resolved geographic location
type Geolocation struct {
	// City is the resolved city name
	City string `json:"city,omitempty"`
	// Country is the resolved country code
	Country string `json:"country,omitempty"`
	// CountryName is the resolved country's full name
	CountryName string `json:"country_name,omitempty"`
	// Region is the resolved region or state
	Region string `json:"region,omitempty"`
	// Latitude is the resolved latitude
	Latitude float64 `json:"latitude,omitempty"`
	// Longitude is the resolved longitude
	Longitude float64 `json:"longitude,omitempty"`
}

// Meta is scan-level metadata: radar rank, categories, and geolocation
type Meta struct {
	// Rank is the site's Cloudflare Radar rank, if known
	Rank int `json:"rank,omitempty"`
	// URLCategories are the scanned URL's detected categories
	URLCategories []string `json:"url_categories,omitempty"`
	// DomainCategories are the scanned domain's detected categories
	DomainCategories []string `json:"domain_categories,omitempty"`
	// Geolocation is the scanned site's resolved geographic location
	Geolocation *Geolocation `json:"geolocation,omitempty"`
}

// IsEmpty reports whether none of the meta sections were populated
func (m Meta) IsEmpty() bool {
	return m.Rank == 0 && len(m.URLCategories) == 0 && len(m.DomainCategories) == 0 && m.Geolocation == nil
}

// Platform is the scanned company's profile, using field names that mirror Openlane's Platform object
type Platform struct {
	// Name is the company's name
	Name string `json:"name"`
	// Description is the company's description
	Description string `json:"description"`
	// Industry is the company's industry
	Industry string `json:"industry"`
	// Location is the company's headquarters location
	Location string `json:"location"`
	// EmployeeRange is the company's estimated employee count range
	EmployeeRange string `json:"employee_range"`
	// FoundedYear is the year the company was founded
	FoundedYear string `json:"founded_year"`
	// EstimatedRevenue is the company's estimated revenue
	EstimatedRevenue string `json:"estimated_revenue"`
	// SSOSupported indicates the company's product advertises SSO support
	SSOSupported bool `json:"sso_supported"`
	// MFASupported indicates the company's product advertises MFA support
	MFASupported bool `json:"mfa_supported"`
	// SocialLinks holds the company's social media and community profile URLs
	SocialLinks SocialLinks `json:"social_links"`
	// StatusPageURL is the company's public status/uptime page, if found
	StatusPageURL string `json:"status_page_url,omitempty"`
	// Customers are customer names or logos found on the site
	Customers []string `json:"customers,omitempty"`
	// ProvidedServices are the services or product categories the company itself provides
	ProvidedServices []string `json:"provided_services,omitempty"`
	// AuthMethods are the authentication methods the company's product advertises support for
	AuthMethods []string `json:"auth_methods,omitempty"`
}

// SystemEntry is one of the company's systems, using field names that mirror Openlane's SystemDetail object
type SystemEntry struct {
	// SystemName is the system's name
	SystemName string `json:"system_name"`
	// Description is the system's description
	Description string `json:"description"`
}

// Compliance is the company's compliance posture gathered by the domainscan enrichment
type Compliance struct {
	// Frameworks lists compliance frameworks or certifications mentioned (e.g., SOC 2, ISO 27001, GDPR, HIPAA)
	Frameworks []string `json:"frameworks,omitempty"`
	// IsSOC2 indicates the company claims SOC 2 (Type I or Type II) certification or compliance
	IsSOC2 bool `json:"is_soc2,omitempty"`
	// Controls lists individual security or compliance controls or practices mentioned
	Controls []string `json:"controls,omitempty"`
	// TrustCenterHostedBy is the platform serving the company's trust center, if one was found
	TrustCenterHostedBy string `json:"trust_center_hosted_by,omitempty"`
	// Documents lists compliance documents found on a trust center
	Documents []TrustDocument `json:"documents,omitempty"`
}

// Registrar is the domain's WHOIS registration data
type Registrar struct {
	// DNSSEC reports whether DNSSEC is enabled for the domain
	DNSSEC bool `json:"dnssec,omitempty"`
	// Registrar is the domain's registrar of record
	Registrar string `json:"registrar,omitempty"`
	// CreatedDate is when the domain was registered, RFC3339
	CreatedDate string `json:"created_date,omitempty"`
	// ExpirationDate is when the domain's registration expires, RFC3339
	ExpirationDate string `json:"expiration_date,omitempty"`
	// UpdatedDate is when the domain's registration was last updated, RFC3339
	UpdatedDate string `json:"updated_date,omitempty"`
	// Nameservers are the domain's authoritative nameservers per WHOIS
	Nameservers []string `json:"nameservers,omitempty"`
	// Status lists the domain's registry status codes (e.g. clientTransferProhibited)
	Status []string `json:"status,omitempty"`
}

// Favicon is the scanned site's favicon
type Favicon struct {
	// URL is the favicon's URL
	URL string `json:"url,omitempty"`
	// Hash is the favicon's content hash
	Hash string `json:"hash,omitempty"`
}

// Branding is visual branding data captured from the scan, usable for org avatars, trust
// center logos, or similar presentational purposes
type Branding struct {
	// Favicon is the scanned site's favicon
	Favicon Favicon `json:"favicon"`
}

// ScanReport is BuildScanReport's typed output for a single domain
type ScanReport struct {
	// ExternalScanID is the Cloudflare URL Scanner task id
	ExternalScanID string `json:"external_scan_id,omitempty"`
	// URL is the scanned URL
	URL string `json:"url,omitempty"`
	// Vendors are the vendors detected during the scan
	Vendors []Vendor `json:"vendors,omitempty"`
	// Technologies are the technologies detected during the scan
	Technologies []Technology `json:"technologies,omitempty"`
	// Assets are the DNS records, IP addresses, and internal domains resolved during the scan
	Assets *Assets `json:"assets,omitempty"`
	// Branding is visual branding data captured from the scan
	Branding *Branding `json:"branding,omitempty"`
	// Findings are the scan's security verdict, agent-readiness assessment, and compliance gaps
	Findings Findings `json:"findings"`
	// Meta is scan-level metadata (radar rank, categories, geolocation)
	Meta *Meta `json:"meta,omitempty"`
	// Platform is the scanned company's profile
	Platform *Platform `json:"platform,omitempty"`
	// Systems are the company's systems discovered by the domainscan enrichment
	Systems []SystemEntry `json:"systems,omitempty"`
	// Compliance is the company's compliance posture gathered by the domainscan enrichment
	Compliance *Compliance `json:"compliance,omitempty"`
	// Registrar is the domain's WHOIS registration data
	Registrar *Registrar `json:"registrar,omitempty"`
}

// Report is the JSON-schema-described payload attached to a domain scan completion Notification.
// Scans holds one entry per domain in the originating submission (a single entry
// for a one-off scan, one entry per domain for an org-settings batch)
// every other section is the union of that data across every successfully completed domain
type Report struct {
	// Scans is one entry per domain in the originating submission, carrying its own identity and outcome
	Scans []Result `json:"scans"`
	// Vendors is the union of vendors detected across every completed domain, deduped by name
	Vendors []Vendor `json:"vendors,omitempty"`
	// Technologies is the union of technologies detected across every completed domain, deduped by name
	Technologies []Technology `json:"technologies,omitempty"`
	// Assets is the union of DNS records, IP addresses, and internal domains across every completed domain
	Assets *Assets `json:"assets,omitempty"`
	// Branding is the first non-empty branding section found across completed domains
	Branding *Branding `json:"branding,omitempty"`
	// Findings is the union of findings across every completed domain
	Findings Findings `json:"findings,omitempty"`
	// Meta is the first non-empty meta section found across completed domains
	Meta *Meta `json:"meta,omitempty"`
	// Platform is the first non-empty platform section found across completed domains
	Platform *Platform `json:"platform,omitempty"`
	// Systems is the union of systems detected across every completed domain, deduped by name
	Systems []SystemEntry `json:"systems,omitempty"`
	// Compliance is the first non-empty compliance section found across completed domains
	Compliance *Compliance `json:"compliance,omitempty"`
	// Registrar is the first non-empty WHOIS registration section found across completed domains
	Registrar *Registrar `json:"registrar,omitempty"`
}

// Result is one domain's outcome within a DomainScanReport
type Result struct {
	// Domain is the scanned domain
	Domain string `json:"domain"`
	// InternalScanID is the Scan record id this result belongs to
	InternalScanID string `json:"internal_scan_id"`
	// ExternalScanID is the Cloudflare URL Scanner task id, present when the scan completed
	ExternalScanID string `json:"external_scan_id,omitempty"`
	// URL is the scanned URL, present when the scan completed
	URL string `json:"url,omitempty"`
	// Status is "completed" or "failed"
	Status string `json:"status"`
}

// DomainScanReportSchema is the reflected JSON schema for DomainScanReport, the shape of a
// domain scan completion Notification's data
var DomainScanReportSchema = jsonx.SchemaFrom[Report]()
