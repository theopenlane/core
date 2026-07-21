package domainscan

import (
	"net/url"
	"sort"
	"strings"

	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
)

// legalEntitySuffixes are corporate suffixes stripped from a raw company/organization name
var legalEntitySuffixes = []string{
	", Inc.", ", Inc", " Inc.", " Inc",
	", LLC", " LLC",
	", Ltd.", " Ltd.", " Ltd",
	", Limited", " Limited",
	" GmbH",
	", Corporation", " Corporation", " Corp.", " Corp",
	", Plc", " Plc", ", PLC", " PLC",
	" S.A.", " S.p.A.", " B.V.", " AG", " Co.",
}

// stripLegalEntitySuffixes repeatedly strips any trailing legalEntitySuffixes match until none remain
func stripLegalEntitySuffixes(name string) string {
	name = strings.TrimSpace(name)

	for {
		stripped := name

		for _, suffix := range legalEntitySuffixes {
			stripped = strings.TrimSuffix(stripped, suffix)
		}

		stripped = strings.TrimSpace(strings.TrimSuffix(stripped, ","))

		if stripped == name {
			return name
		}

		name = stripped
	}
}

// buildAssets reports the DNS records and IP addresses resolved during the scan, annotating each
// IP with its ASN/organization when known, plus any internal subdomains discovered by the
// domainscan enrichment. result may be nil, in which case only enrichment-derived assets are included
func buildAssets(result *url_scanner.ScanGetResponse, enrichment Enrichment) *Assets {
	var apexDomain, pageASNOrg string
	if result != nil {
		apexDomain = result.Task.ApexDomain
		pageASNOrg = result.Page.Asnname
	}

	assets := Assets{DNSRecords: append(buildDNSRecords(result, enrichment), buildInternalDomains(enrichment, apexDomain, pageASNOrg)...)}

	if result != nil && len(result.Lists.IPs) > 0 {
		asnByIP := make(map[string]url_scanner.ScanGetResponseMetaProcessorsASNData, len(result.Meta.Processors.ASN.Data))
		for _, a := range result.Meta.Processors.ASN.Data {
			asnByIP[a.IP] = a
		}

		ipAddresses := make([]IPAddress, 0, len(result.Lists.IPs))
		for _, ip := range result.Lists.IPs {
			entry := IPAddress{Address: ip}

			if a, ok := asnByIP[ip]; ok {
				entry.ASN = a.ASN
				entry.Org = a.Description
			}

			ipAddresses = append(ipAddresses, entry)
		}

		assets.IPAddresses = ipAddresses
	}

	if assets.IsEmpty() {
		return nil
	}

	return &assets
}

// buildDNSRecords combines the scan's resolved A records with the domainscan enrichment's NS
// records for the scanned domain's apex. result may be nil, in which case only the enrichment's
// NS records are included
func buildDNSRecords(result *url_scanner.ScanGetResponse, enrichment Enrichment) []DNSRecord {
	var dnsRecords []DNSRecord

	if result != nil {
		dnsRecords = make([]DNSRecord, 0, len(result.Lists.Domains))

		for _, d := range result.Lists.Domains {
			record := DNSRecord{Domain: d, Type: "A"}
			record.Vendor = hostVendorName(d, result.Task.ApexDomain, result.Page.Asnname)

			dnsRecords = append(dnsRecords, record)
		}
	}

	if enrichment.DNS != nil {
		for _, ns := range enrichment.DNS.NSHosts {
			record := DNSRecord{Domain: ns, Type: "NS"}
			record.Vendor, _ = vendorNameFromHostname(ns)

			dnsRecords = append(dnsRecords, record)
		}
	}

	return dnsRecords
}

// buildInternalDomains collects the subdomains the domainscan enrichment discovered, deduped and
// sorted, each attributed to whoever fronts it, matched by its registrable root domain the same
// way buildDNSRecords attributes its own entries
func buildInternalDomains(enrichment Enrichment, apexDomain, pageASNOrg string) []DNSRecord {
	seen := map[string]bool{}

	var hosts []string

	add := func(host string) {
		host = strings.ToLower(strings.TrimSpace(host))
		if host == "" || seen[host] {
			return
		}

		seen[host] = true

		hosts = append(hosts, host)
	}

	if enrichment.Company != nil {
		if u, err := url.Parse(enrichment.Company.StatusPageURL); err == nil {
			add(u.Hostname())
		}

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

	sort.Strings(hosts)

	records := make([]DNSRecord, 0, len(hosts))
	for _, host := range hosts {
		records = append(records, DNSRecord{Domain: host, Type: "internal", Vendor: hostVendorName(host, apexDomain, pageASNOrg)})
	}

	return records
}

// hostVendorName returns the vendor a DNS record's hostname belongs to. Third-party hosts are
// matched by domain; the scanned site's own apex (or a subdomain of it) has no distinguishing
// hostname to match on, so it falls back to whoever's fronting it per the scan's ASN data (e.g.
// Cloudflare proxying the customer's own domain)
func hostVendorName(host, apexDomain, pageASNOrg string) string {
	domain, ok := icannRegistrableDomain(host)
	if !ok {
		return ""
	}

	if domain != apexDomain {
		return domainVendorName(domain)
	}

	return vendorNameFromASNOrg(pageASNOrg)
}

// vendorNameFromASNOrg derives a vendor display name from an ASN organization name (e.g.
// "Cloudflare, Inc." -> "Cloudflare"), canonicalizing it against known vendor aliases when possible
func vendorNameFromASNOrg(org string) string {
	name := stripLegalEntitySuffixes(org)
	if name == "" {
		return ""
	}

	if canonical, ok := vendorCanonicalNames[strings.ToLower(name)]; ok {
		return canonical
	}

	// ASN registry org names often carry a trailing "-<code>" (e.g. "AMAZON-02", "AKAMAI-ASN1");
	// retry against known aliases with it stripped before falling back to the name as-is
	if base, _, found := strings.Cut(name, "-"); found {
		if canonical, ok := vendorCanonicalNames[strings.ToLower(base)]; ok {
			return canonical
		}
	}

	return name
}
