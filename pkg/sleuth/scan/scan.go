package scan

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/pkg/sleuth/certx"
	"github.com/theopenlane/core/pkg/sleuth/dnsx"
	"github.com/theopenlane/core/pkg/sleuth/tech"
)

// DomainScanReport aggregates the results of multiple reconnaissance operations
// performed against a domain.
type DomainScanReport struct {
	DNS          dnsx.DNSRecordsReport   `json:"dns"`
	Certificates certx.CertsReport       `json:"certificates"`
	Technologies map[string]tech.AppInfo `json:"technologies"`
}

// ScanDomain performs a set of lightweight reconnaissance operations against the
// provided domain and returns a DomainScanReport containing the results.
func ScanDomain(ctx context.Context, domain string) (DomainScanReport, error) {
	var report DomainScanReport

	dnsClient, err := dnsx.NewDNSX(dnsx.WithOutputCDN(true))
	if err != nil {
		return report, fmt.Errorf("creating dns client: %w", err)
	}

	dnsReport, err := dnsClient.GetDomainDNSRecords(domain)
	if err != nil {
		return report, err
	}
	report.DNS = dnsReport

	certReport, err := certx.GetDomainCerts(ctx, domain)
	if err != nil {
		// non fatal but log the error in the report
		report.Certificates.Errors = append(report.Certificates.Errors, err.Error())
	} else {
		report.Certificates = certReport
	}

	techClient, err := tech.NewTech("https://" + domain)
	if err == nil {
		apps, err := techClient.GetTech()
		if err == nil {
			report.Technologies = apps
		}
	}

	return report, nil
}
