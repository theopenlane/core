package sleuth

// DNSRecord represents a DNS record with its associated properties
type DNSRecord struct {
	Name  string `json:"name" url:"name"`
	TTL   int    `json:"ttl" url:"ttl"`
	Type  string `json:"type" url:"type"`
	Value string `json:"value" url:"value"`
}

// DNSRecords represents a collection of DNS records of various types
type DNSRecords struct {
	A     []*DNSRecord `json:"a,omitempty" url:"a,omitempty"`
	AAAA  []*DNSRecord `json:"aaaa,omitempty" url:"aaaa,omitempty"`
	MX    []*DNSRecord `json:"mx,omitempty" url:"mx,omitempty"`
	Txt   []*DNSRecord `json:"txt,omitempty" url:"txt,omitempty"`
	NS    []*DNSRecord `json:"ns,omitempty" url:"ns,omitempty"`
	CNAME []*DNSRecord `json:"cname,omitempty" url:"cname,omitempty"`
}

// DNSRecordsReport represents a report of DNS records for a specific domain
type DNSRecordsReport struct {
	Domain          string      `json:"domain" url:"domain"`
	DNSRecords      *DNSRecords `json:"dnsRecords,omitempty" url:"dnsRecords,omitempty"`
	DMARCDomain     *string     `json:"dmarcDomain,omitempty" url:"dmarcDomain,omitempty"`
	DMARCDNSRecords *DNSRecords `json:"dmarcDNSRecords,omitempty" url:"dmarcDNSRecords,omitempty"`
	DKIMDomain      *string     `json:"dkimDomain,omitempty" url:"dkimDomain,omitempty"`
	DKIMDNSRecords  *DNSRecords `json:"dkimDNSRecords,omitempty" url:"dkimDNSRecords,omitempty"`
	Errors          []string    `json:"errors,omitempty" url:"errors,omitempty"`
}

// DNSSubenumReport represents a report of DNS subdomain enumeration for a specific domain
type DNSSubenumReport struct {
	Domain          string         `json:"domain" url:"domain"`
	EnumerationType DNSSubenumType `json:"enumerationType" url:"enumerationType"`
	Subdomains      []string       `json:"subdomains,omitempty" url:"subdomains,omitempty"`
	Errors          []string       `json:"errors,omitempty" url:"errors,omitempty"`
}

// DNSSubenumType represents the type of DNS subdomain enumeration
type DNSSubenumType string

const (
	DNSSubenumTypeBrute   DNSSubenumType = "BRUTE"
	DNSSubenumTypePassive DNSSubenumType = "PASSIVE"
)

// DomainTakeover represents a potential domain takeover vulnerability
type DomainTakeover struct {
	Target       string     `json:"target" url:"target"`
	StatusCode   int        `json:"statusCode" url:"statusCode"`
	ResponseBody string     `json:"responseBody" url:"responseBody"`
	Domain       string     `json:"domain" url:"domain"`
	CNAME        string     `json:"cname" url:"cname"`
	Services     []*Service `json:"services,omitempty" url:"services,omitempty"`
}

// DomainTakeoverReport represents a report of potential domain takeover vulnerabilities
type DomainTakeoverReport struct {
	DomainTakeovers []*DomainTakeover `json:"domainTakeovers,omitempty" url:"domainTakeovers,omitempty"`
	Errors          []string          `json:"errors,omitempty" url:"errors,omitempty"`
}

// Fingerprint represents a system fingerprint to detect domain takeovers
type Fingerprint struct {
	CICDPass      bool     `json:"cicdPass" url:"cicdPass"`
	CNAME         []string `json:"cname,omitempty" url:"cname,omitempty"`
	Discussion    string   `json:"discussion" url:"discussion"`
	Documentation string   `json:"documentation" url:"documentation"`
	Fingerprint   string   `json:"fingerprint" url:"fingerprint"`
	HTTPStatus    *int     `json:"httpStatus,omitempty" url:"httpStatus,omitempty"`
	NXDomain      bool     `json:"nxDomain" url:"nxDomain"`
	Service       string   `json:"service" url:"service"`
	Status        string   `json:"status" url:"status"`
	Vulnerable    bool     `json:"vulnerable" url:"vulnerable"`
}

// Service represents a service associated with a potential domain takeover vulnerability
type Service struct {
	Name        string `json:"name" url:"name"`
	Fingerprint string `json:"fingerprint" url:"fingerprint"`
	Vulnerable  bool   `json:"vulnerable" url:"vulnerable"`
}
