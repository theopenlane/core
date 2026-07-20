package domainscan

import (
	"context"
	"time"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/intel"
)

// RegistrarInfo holds WHOIS registration data for a domain
type RegistrarInfo struct {
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
	// DNSSEC reports whether DNSSEC is enabled for the domain
	DNSSEC bool `json:"dnssec,omitempty"`
}

// GetRegistrarInfo fetches WHOIS registration data for domain via Cloudflare's Intel API
func (c *Config) GetRegistrarInfo(ctx context.Context, domain string) (*RegistrarInfo, error) {
	client := cloudflare.NewClient(c.clientOptions()...)

	resp, err := client.Intel.Whois.Get(ctx, intel.WhoisGetParams{
		AccountID: cloudflare.F(c.AccountID),
		Domain:    cloudflare.F(domain),
	})
	if err != nil {
		return nil, err
	}

	if !resp.Found {
		return nil, nil
	}

	info := &RegistrarInfo{
		Registrar:   resp.Registrar,
		Nameservers: resp.Nameservers,
		Status:      resp.Status,
		DNSSEC:      resp.DNSSEC,
	}

	if !resp.CreatedDate.IsZero() {
		info.CreatedDate = resp.CreatedDate.Format(time.RFC3339)
	}

	if !resp.ExpirationDate.IsZero() {
		info.ExpirationDate = resp.ExpirationDate.Format(time.RFC3339)
	}

	if !resp.UpdatedDate.IsZero() {
		info.UpdatedDate = resp.UpdatedDate.Format(time.RFC3339)
	}

	return info, nil
}
