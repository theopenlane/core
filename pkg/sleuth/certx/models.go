package certx

// CertificateRecord represents all of the information for a single x509 certificate record
type CertificateRecord struct {
	IssuerCAID     int    `json:"issuer_ca_id" yaml:"issuer_ca_id"`
	IssuerName     string `json:"issuer_name" yaml:"issuer_name"`
	CommonName     string `json:"common_name" yaml:"common_name"`
	NameValue      string `json:"name_value" yaml:"name_value"`
	ID             int64  `json:"id" yaml:"id"`
	EntryTimestamp string `json:"entry_timestamp" yaml:"entry_timestamp"`
	NotBefore      string `json:"not_before" yaml:"not_before"`
	NotAfter       string `json:"not_after" yaml:"not_after"`
	SerialNumber   string `json:"serial_number" yaml:"serial_number"`
	ResultCount    int    `json:"result_count" yaml:"result_count"`
}

// CertsReport represents the report of all certificates for a given domain
type CertsReport struct {
	Domain       string              `json:"domain" yaml:"domain"`
	Certificates []CertificateRecord `json:"certificates" yaml:"certificates"`
	Errors       []string            `json:"errors" yaml:"errors"`
}
