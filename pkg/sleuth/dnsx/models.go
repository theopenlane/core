package dnsx

// DNSRecord represents a DNS record with its associated properties
type DNSRecord struct {
	Name  string     `json:"name"`
	TTL   int        `json:"ttl"`
	Type  string     `json:"type"`
	Value string     `json:"value"`
	IP    *IPAddress `json:"ip,omitempty"`
	CDN   string     `json:"cdn,omitempty"`
}

// IPAddress represents an IP address with its associated properties
type IPAddress struct {
	IP        string `json:"ip"`
	RDNS      string `json:"rdns"`
	Dedicated bool   `json:"dedicated"`
}

// DNSRecords represents a collection of DNS records of various types
type DNSRecords struct {
	A     []*DNSRecord `json:"a,omitempty"`
	AAAA  []*DNSRecord `json:"aaaa,omitempty"`
	MX    []*DNSRecord `json:"mx,omitempty"`
	Txt   []*DNSRecord `json:"txt,omitempty"`
	NS    []*DNSRecord `json:"ns,omitempty"`
	CNAME []*DNSRecord `json:"cname,omitempty"`
	SPF   []*DNSRecord `json:"spf,omitempty"`
	DMARC []*DNSRecord `json:"dmarc,omitempty"`
	CAA   []*DNSRecord `json:"caa,omitempty"`
	SSHFP []*DNSRecord `json:"sshfp,omitempty"`
	DS    []*DNSRecord `json:"ds,omitempty"`
	URI   []*DNSRecord `json:"uri,omitempty"`
	HTTPS []*DNSRecord `json:"https,omitempty"`
	SMIME []*DNSRecord `json:"smime,omitempty"`
	SPKI  []*DNSRecord `json:"spki,omitempty"`
	ALIAS []*DNSRecord `json:"alias,omitempty"`
	PTR   []*DNSRecord `json:"ptr,omitempty"`
}

// DNSRecordsReport represents a report of DNS records for a specific domain
type DNSRecordsReport struct {
	Domain          string      `json:"domain"`
	DNSRecords      *DNSRecords `json:"dnsRecords,omitempty"`
	DMARCDomain     *string     `json:"dmarcDomain,omitempty"`
	DMARCDNSRecords *DNSRecords `json:"dmarcDNSRecords,omitempty"`
	DKIMDomain      *string     `json:"dkimDomain,omitempty"`
	DKIMDNSRecords  *DNSRecords `json:"dkimDNSRecords,omitempty"`
	CDNName         *string     `json:"cdnName,omitempty"`
	Errors          []string    `json:"errors,omitempty"`
}

// DNSSubenumReport represents a report of DNS subdomain enumeration for a specific domain
type DNSSubenumReport struct {
	Domain          string         `json:"domain"`
	EnumerationType DNSSubenumType `json:"enumerationType"`
	Subdomains      []string       `json:"subdomains,omitempty"`
	Errors          []string       `json:"errors,omitempty"`
}

// DNSSubenumType represents the type of DNS subdomain enumeration
type DNSSubenumType string

const (
	DNSSubenumTypeBrute   DNSSubenumType = "BRUTE"
	DNSSubenumTypePassive DNSSubenumType = "PASSIVE"
)

// DomainTakeover represents a potential domain takeover vulnerability
type DomainTakeover struct {
	Target       string     `json:"target"`
	StatusCode   int        `json:"statusCode"`
	ResponseBody string     `json:"responseBody"`
	Domain       string     `json:"domain"`
	CNAME        string     `json:"cname"`
	Services     []*Service `json:"services,omitempty"`
}

// DomainTakeoverReport represents a report of potential domain takeover vulnerabilities
type DomainTakeoverReport struct {
	DomainTakeovers []*DomainTakeover `json:"domainTakeovers,omitempty"`
	Errors          []string          `json:"errors,omitempty"`
}

// Fingerprint represents a system fingerprint to detect domain takeovers
type Fingerprint struct {
	CICDPass      bool     `json:"cicdPass"`
	CNAME         []string `json:"cname,omitempty"`
	Discussion    string   `json:"discussion"`
	Documentation string   `json:"documentation"`
	Fingerprint   string   `json:"fingerprint"`
	HTTPStatus    *int     `json:"httpStatus,omitempty"`
	NXDomain      bool     `json:"nxDomain"`
	Service       string   `json:"service"`
	Status        string   `json:"status"`
	Vulnerable    bool     `json:"vulnerable"`
}

// Service represents a service associated with a potential domain takeover vulnerability
type Service struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	Vulnerable  bool   `json:"vulnerable"`
}
