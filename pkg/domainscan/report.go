package domainscan

import (
	"context"
	"time"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
	"golang.org/x/sync/errgroup"

	"github.com/theopenlane/core/pkg/jsonx"
)

// Enrichment holds the best-effort pkg/domainscan results for a single
// domain
type Enrichment struct {
	// Company is the base company profile
	Company *CompanyProfile `json:"company,omitempty"`
	// Compliance is compliance specific data such as soc2 attestation, subprocessors, etc
	Compliance *CompliancePage `json:"compliance,omitempty"`
	// DNS includes DNS probed data
	DNS *DNSVendorInfo `json:"dns,omitempty"`
}

// EnrichmentErrors holds the per-lookup errors from GatherEnrichment, each nil on success
type EnrichmentErrors struct {
	Company    error
	Compliance error
	DNS        error
}

// ReportConfig configures how BuildScanReport classifies vendors versus
// technologies and which vendor names it always excludes
type ReportConfig struct {
	// NonVendorCategories lists wappalyzer categories treated as technologies
	// instead of vendors when building an onboarding domain scan report
	NonVendorCategories []string `json:"nonvendorcategories" koanf:"nonvendorcategories" default:"[Miscellaneous,JavaScript frameworks,JavaScript libraries,Static site generator]"`
	// DeniedVendorNames lists vendor names to always exclude from an onboarding domain scan report's vendor list
	DeniedVendorNames []string `json:"deniedvendornames" koanf:"deniedvendornames" default:"[rfc-editor,ajax,website-files,http/3,googletagmanager,cloudflareinsights,googlesyndication,gstatic,hcaptcha,googleapis,hsforms,hs-scripts,hscollectedforms,hsts,hs-banner,hs-analytics]"`
	// ScanTTL is the cache TTL, in seconds, for Browser Rendering requests issued during domain scan enrichment
	ScanTTL int `json:"scanttl" koanf:"scanttl" default:"86400"`
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

// BuildScanReport combines a Cloudflare URL Scanner result with the Enrichment gathered by GatherEnrichment into a single report
// unified vendors/technologies, assets, findings, meta, platform, systems, and compliance sections
func BuildScanReport(result *url_scanner.ScanGetResponse, enrichment Enrichment, nonVendorCategories, deniedVendorNames []string) map[string]any {
	report := ScanReport{
		Findings: buildFindings(result, enrichment),
	}

	if result != nil {
		report.ExternalScanID = result.Task.UUID
		report.URL = result.Task.URL
	}

	report.Vendors, report.Technologies = buildVendorsAndTechnologies(result, enrichment, nonVendorCategories, deniedVendorNames)
	report.Assets = buildAssets(result, enrichment)
	report.Branding = buildBranding(result)
	report.Meta = buildMeta(result)
	report.Platform = buildPlatform(enrichment)
	report.Systems = buildSystems(enrichment)
	report.Compliance = buildComplianceSection(enrichment)

	data, _ := jsonx.ToMap(report)

	return data
}
