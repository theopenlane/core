package domainscan

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// maxSubdomainDepth bounds how many levels of commonSubdomains are nested (e.g. mail.<apex> -> send.mail.<apex>)
const maxSubdomainDepth = 2

// GetDNSVendorInfo resolves MX/SPF/DMARC at the apex, probes commonSubdomains for MX/TXT/CNAME/DKIM,
// and derives vendor names from whatever's found rather than a maintained lookup table
func GetDNSVendorInfo(ctx context.Context, rawURL string) (*DNSVendorInfo, error) {
	host, ok := apexDomain(rawURL)
	if !ok {
		return nil, fmt.Errorf("could not determine domain from %q", rawURL)
	}

	info := &DNSVendorInfo{}

	seen := make(map[string]int)
	addVendor := func(name, url string) {
		if name == "" {
			return
		}

		if idx, ok := seen[name]; ok {
			if url != "" && info.Vendors[idx].URL == "" {
				info.Vendors[idx].URL = url
			}

			return
		}

		seen[name] = len(info.Vendors)
		info.Vendors = append(info.Vendors, DNSVendor{Name: name, URL: url})
	}

	// apex records can't be a CNAME per the DNS spec, so only MX/TXT are looked up
	apexSignals := lookupHostSignals(ctx, host, false)
	info.MXHosts = apexSignals.MXHosts
	info.TXTRecords = apexSignals.TXTRecords
	info.SPFRecord = apexSignals.SPFRecord
	collectVendorsFromSignals(apexSignals, addVendor)

	if nsRecords, err := net.DefaultResolver.LookupNS(ctx, fqdn(host)); err == nil {
		for _, ns := range nsRecords {
			nsHost := strings.TrimSuffix(ns.Host, ".")
			info.NSHosts = append(info.NSHosts, nsHost)
			addHostnameVendor(nsHost, addVendor)
		}
	}

	info.DKIMSelectors = probeDKIMSelectors(ctx, host)
	for _, selector := range info.DKIMSelectors {
		if vendor, ok := dkimSelectorVendor(selector); ok {
			addVendor(vendor, "")
		}
	}

	info.Subdomains = scanSubdomains(ctx, host, maxSubdomainDepth, addVendor)

	if dmarc, ok := lookupTXTPrefix(ctx, "_dmarc."+host, "v=dmarc1"); ok {
		info.DMARCRecord = dmarc
		info.DMARCPolicy = dmarcTagValue(dmarc, "p")
	}

	if hosts, err := certTransparencySubdomains(ctx, host); err == nil {
		info.CertSubdomains = hosts
	}

	return info, nil
}

// scanSubdomains probes each commonSubdomains entry under parentHost for DNS activity, reporting vendors to addVendor
func scanSubdomains(ctx context.Context, parentHost string, maxDepth int, addVendor func(name, url string)) []SubdomainDNSInfo {
	if maxDepth <= 0 {
		return nil
	}

	var results []SubdomainDNSInfo

	for _, label := range commonSubdomains {
		host := label + "." + parentHost

		sig := lookupHostSignals(ctx, host, true)
		selectors := probeDKIMSelectors(ctx, host)

		if len(sig.MXHosts) == 0 && len(sig.TXTRecords) == 0 && sig.CNAME == "" && len(selectors) == 0 {
			continue
		}

		collectVendorsFromSignals(sig, addVendor)

		for _, selector := range selectors {
			if vendor, ok := dkimSelectorVendor(selector); ok {
				addVendor(vendor, "")
			}
		}

		results = append(results, SubdomainDNSInfo{
			Host:          host,
			MXHosts:       sig.MXHosts,
			TXTRecords:    sig.TXTRecords,
			CNAME:         sig.CNAME,
			DKIMSelectors: selectors,
		})

		results = append(results, scanSubdomains(ctx, host, maxDepth-1, addVendor)...)
	}

	return results
}

// probeDKIMSelectors checks host for each of commonDKIMSelectors, returning the labels that resolve
func probeDKIMSelectors(ctx context.Context, host string) []string {
	var found []string

	for _, selector := range commonDKIMSelectors {
		name := selector.label + "._domainkey." + host

		if records, err := net.DefaultResolver.LookupTXT(ctx, fqdn(name)); err == nil && len(records) > 0 {
			found = append(found, selector.label)
		}
	}

	return found
}

// dkimSelectorVendor looks up the vendor for a resolved DKIM selector label
func dkimSelectorVendor(label string) (string, bool) {
	for _, selector := range commonDKIMSelectors {
		if selector.label == label {
			return selector.vendor, selector.vendor != ""
		}
	}

	return "", false
}

// hostSignals holds the raw DNS records found for a single host
type hostSignals struct {
	MXHosts    []string
	TXTRecords []string
	SPFRecord  string
	CNAME      string
}

// lookupHostSignals resolves MX and TXT records for host, plus its CNAME target if includeCNAME is set
func lookupHostSignals(ctx context.Context, host string, includeCNAME bool) hostSignals {
	var sig hostSignals

	if mxRecords, err := net.DefaultResolver.LookupMX(ctx, fqdn(host)); err == nil {
		for _, mx := range mxRecords {
			sig.MXHosts = append(sig.MXHosts, strings.TrimSuffix(mx.Host, "."))
		}
	}

	if records, err := net.DefaultResolver.LookupTXT(ctx, fqdn(host)); err == nil {
		sig.TXTRecords = records

		for _, record := range records {
			if strings.HasPrefix(strings.ToLower(record), "v=spf1") {
				sig.SPFRecord = record
				break
			}
		}
	}

	if includeCNAME {
		if cname, ok := lookupCNAMETarget(ctx, host); ok {
			sig.CNAME = cname
		}
	}

	return sig
}

// lookupCNAMETarget returns host's canonical name if it's actually an alias
func lookupCNAMETarget(ctx context.Context, host string) (string, bool) {
	cname, err := net.DefaultResolver.LookupCNAME(ctx, fqdn(host))
	if err != nil {
		return "", false
	}

	cname = strings.TrimSuffix(cname, ".")
	if strings.EqualFold(cname, host) {
		return "", false
	}

	return cname, true
}

// addHostnameVendor derives a vendor name from host's registrable domain and reports it to addVendor
func addHostnameVendor(host string, addVendor func(name, url string)) {
	name, domain := vendorNameFromHostname(host)
	if name == "" {
		return
	}

	url := ""
	if domain != "" {
		url = "https://" + domain
	}

	addVendor(name, url)
}

// collectVendorsFromSignals derives vendor names from a host's MX, SPF, TXT, and CNAME records,
// passing along the registrable domain behind the MX/SPF/CNAME hostname as the vendor's URL when
// one was resolved
func collectVendorsFromSignals(sig hostSignals, addVendor func(name, url string)) {
	for _, mxHost := range sig.MXHosts {
		addHostnameVendor(mxHost, addVendor)
	}

	if sig.SPFRecord != "" {
		for _, spfHost := range spfIncludeHosts(sig.SPFRecord) {
			addHostnameVendor(spfHost, addVendor)
		}
	}

	for _, record := range sig.TXTRecords {
		if name, ok := verificationVendorName(record); ok {
			addVendor(name, "")
		}
	}

	if sig.CNAME != "" {
		addHostnameVendor(sig.CNAME, addVendor)
	}
}

// lookupTXTPrefix returns the first TXT record on host whose value starts with prefix (case-insensitively)
func lookupTXTPrefix(ctx context.Context, host, prefix string) (string, bool) {
	records, err := net.DefaultResolver.LookupTXT(ctx, fqdn(host))
	if err != nil {
		return "", false
	}

	for _, record := range records {
		if strings.HasPrefix(strings.ToLower(record), prefix) {
			return record, true
		}
	}

	return "", false
}

// fqdn ensures host has a trailing dot, bypassing resolv.conf's ndots search-domain expansion
// (otherwise a bare name like "example.com" wastes round trips against every k8s cluster search suffix)
func fqdn(host string) string {
	if strings.HasSuffix(host, ".") {
		return host
	}

	return host + "."
}

// dmarcTagValue extracts the value of a tag (e.g. "p" for policy) from a semicolon-delimited DMARC record
func dmarcTagValue(record, tag string) string {
	for _, part := range strings.Split(record, ";") {
		name, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok && strings.EqualFold(strings.TrimSpace(name), tag) {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

// spfIncludeHosts extracts the hostnames referenced by "include:" mechanisms in an SPF record
func spfIncludeHosts(record string) []string {
	var hosts []string

	for _, field := range strings.Fields(record) {
		if host, ok := strings.CutPrefix(field, "include:"); ok {
			hosts = append(hosts, host)
		}
	}

	return hosts
}

// vendorNameFromHostname derives a display name from a hostname's
// registrable domain by capitalizing its first label, e.g.
// "aspmx.l.google.com" -> "Google", "google.com", also returning that
// registrable domain so callers can use it as the vendor's URL
func vendorNameFromHostname(host string) (name, domain string) {
	host = strings.TrimSuffix(host, ".")

	domain, ok := icannRegistrableDomain(host)
	if !ok {
		return "", ""
	}

	if override, ok := vendorHostNames[strings.ToLower(host)]; ok {
		return override, domain
	}

	if override, ok := vendorDomainNames[domain]; ok {
		return override, domain
	}

	label := strings.SplitN(domain, ".", 2)[0]

	return titleCaseLabel(label), domain
}

// titleCaseLabel capitalizes the first letter of a string
func titleCaseLabel(s string) string {
	if s == "" {
		return ""
	}

	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// verificationVendorName recognizes a "tag=token" or "tag:token" TXT record (e.g.
// "google-site-verification=abc123") and derives the vendor name from the tag's first hyphen-delimited part
func verificationVendorName(record string) (string, bool) {
	key, value, ok := strings.Cut(record, "=")
	if !ok {
		key, value, ok = strings.Cut(record, ":")
	}

	if !ok {
		return "", false
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if key == "" || value == "" || len(key) > 40 {
		return "", false
	}

	if strings.ContainsAny(key, " \t") || strings.ContainsAny(value, " \t") {
		return "", false
	}

	switch strings.ToLower(key) {
	case "v", "spf", "dmarc", "dkim":
		// mail-protocol tags, not vendor verification tags
		return "", false
	}

	label := strings.SplitN(key, "-", 2)[0]

	return titleCaseLabel(label), label != ""
}
