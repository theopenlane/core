package dnsx

import (
	"context"
	"fmt"
	"net"
	"slices"

	miekgdns "github.com/miekg/dns"
	"github.com/projectdiscovery/cdncheck"
	pddnsx "github.com/projectdiscovery/dnsx/libs/dnsx"
	"github.com/rs/zerolog/log"
)

// DNSX is a wrapper around the dnsx library for client initialization and functional options settings
type DNSX struct {
	// Resolver is the dnsx resolver
	Client *pddnsx.DNSX
	// Options are the dnsx options
	Options *Options
	// Records are the DNS records
	Records DNSRecordsReport
	// CDNCheck is the CDN check client
	CDNCheck *cdncheck.Client
}

// NewDNSX creates a new DNSX client
func NewDNSX(opts ...Option) (*DNSX, error) {
	options := NewOptions(opts...)

	client, err := pddnsx.New(pddnsx.Options{
		BaseResolvers:     options.BaseResolvers,
		MaxRetries:        options.MaxRetries,
		QuestionTypes:     options.QuestionTypes,
		Trace:             options.Trace,
		TraceMaxRecursion: options.TraceMaxRecursion,
		Hostsfile:         options.Hostsfile,
		OutputCDN:         options.OutputCDN,
		QueryAll:          options.QueryAll,
	})
	if err != nil {
		return nil, err
	}

	cdnClient, err := cdncheck.NewWithOpts(options.MaxRetries, options.BaseResolvers)
	if err != nil {
		return nil, err
	}

	return &DNSX{
		Client:   client,
		Options:  options,
		CDNCheck: cdnClient,
	}, nil
}

// Lookup performs a DNS lookup for the given host and returns the IP addresses
func (d *DNSX) Lookup(host string) ([]net.IP, error) {
	ips, err := d.Client.Lookup(host)
	if err != nil {
		return nil, err
	}

	targetIPs := make([]net.IP, 0, len(ips))

	for _, ii := range ips {
		ip := net.ParseIP(ii)
		if ip != nil {
			targetIPs = append(targetIPs, ip)
		}
	}

	return targetIPs, nil
}

// GetDomainDNSRecords queries DNS for all records for a given domain
func (d *DNSX) GetDomainDNSRecords(ctx context.Context, domain string) (DNSRecordsReport, error) {
	errors := []string{}

	dnsRecords, err := d.getDNSRecords(domain, questionTypes)
	if err != nil {
		// append the error to the errors slice
		// we don't want to return the error here because we want to return the DNS records even if there are errors
		errors = append(errors, err.Error())
	}

	// The DMARC record is always in the _dmarc subdomain (RFC-7489)
	dmarcRecords, err := d.getDNSRecords("_dmarc."+domain, []uint16{miekgdns.TypeTXT})
	if err != nil {
		errors = append(errors, err.Error())
	}

	// The DKIM record is always in the _domainkey subdomain (RFC-6376)
	// the _domainkey subdomain itself includes a subdomain named after a selector which we
	// don't know in advance, so we need to check common selectors
	dkimRecords := DNSRecords{}

	for _, selector := range selectors {
		dkimRecordForSelector, err := d.getDNSRecords(selector+"._domainkey."+domain, []uint16{miekgdns.TypeTXT})

		if err != nil {
			errors = append(errors, err.Error())
		}

		dkimRecords.Txt = append(dkimRecords.Txt, dkimRecordForSelector.Txt...)
	}

	report := DNSRecordsReport{
		Domain:          domain,
		DNSRecords:      &dnsRecords,
		DMARCDNSRecords: &dmarcRecords,
		DKIMDNSRecords:  &dkimRecords,
		Errors:          errors,
	}

	// Set the domain for the DMARC and DKIM records if they exist
	if len(dmarcRecords.Txt) > 0 {
		report.DMARCDomain = &dmarcRecords.Txt[0].Name
	}

	if len(dkimRecords.Txt) > 0 {
		report.DKIMDomain = &dkimRecords.Txt[0].Name
	}

	return report, nil
}

// LookupCDN checks if the given domain is a CDN and returns the CDN name and type
func (d *DNSX) LookupCDN(domain string) (value, cdnType string, err error) {
	if net.ParseIP(domain) == nil {
		ips, err := d.Lookup(domain)
		if err != nil {
			return "", "", err
		}

		if len(ips) == 0 {
			return "", "", err
		}

		// this isn't ideal but it's functional and fine for now
		for _, ip := range ips {
			if ip.IsPrivate() {
				return "", "", err
			}
		}

		for _, ip := range ips {
			ok, value, cdnType, err := d.CDNCheck.Check(ip)
			if err != nil {
				return "", "", err
			}

			if ok && cdnType != "" {
				return value, cdnType, nil
			}
		}
	} else {
		ok, value, cdnType, err := d.CDNCheck.Check(net.ParseIP(domain))
		if err != nil {
			return "", "", err
		}

		if ok && cdnType != "" {
			return value, cdnType, nil
		}
	}

	return "", "", nil
}

// getDNSRecords queries for DNS records based on input host string and the question types
// instead of always querying for all types and filtering, we can pass the question types as to not generate unnecessary queries / traffic as well as to perform granular lookups
func (d *DNSX) getDNSRecords(domain string, questionTypes []uint16) (DNSRecords, error) {
	dnsRecords := DNSRecords{}
	// Create the new client and set the question types as you cannot change the type post-creation of the client or at least I couldn't figure out how to
	client, err := NewDNSX(WithQuestionTypes(questionTypes))
	if err != nil {
		return dnsRecords, err
	}

	d.Client = client.Client

	// queryMultiple is what allows us to input the record types we want
	results, err := d.Client.QueryMultiple(domain)
	if err != nil {
		return dnsRecords, err
	}

	if results == nil || results.StatusCode == "NXDOMAIN" || results.Host == "" || results.Timestamp.IsZero() {
		return dnsRecords, nil
	}

	// god i hope i don't regret this...
	populateRecords := func(records []string, recordType string) []*DNSRecord {
		dnsRecordsSlice := make([]*DNSRecord, len(records))
		for i, record := range records {
			dnsRecordsSlice[i] = &DNSRecord{
				Name:  domain,
				TTL:   int(results.TTL),
				Type:  recordType,
				Value: record,
			}
		}

		// these records we know are the ones we can get tasty stuff out of
		if recordType == "A" || recordType == "AAAA" || recordType == "CNAME" || recordType == "MX" {
			for i, record := range records {
				val, cdnType, err := d.LookupCDN(record)
				if err != nil {
					log.Error().Msgf("error looking up CDN for %s: %v", record, err)
					continue
				}

				cdn := fmt.Sprintf("%s (%s)", val, cdnType)

				if cdn != "" {
					dnsRecordsSlice[i].CDN = cdn
				}
			}
		}

		return dnsRecordsSlice
	}

	recordTypes := map[uint16][]string{
		miekgdns.TypeA:     results.A,
		miekgdns.TypeAAAA:  results.AAAA,
		miekgdns.TypeCNAME: results.CNAME,
		miekgdns.TypeMX:    results.MX,
		miekgdns.TypeNS:    results.NS,
		miekgdns.TypeTXT:   results.TXT,
		miekgdns.TypePTR:   results.PTR,
	}

	for qType, records := range recordTypes {
		if slices.Contains(questionTypes, qType) {
			switch qType {
			case miekgdns.TypeA:
				dnsRecords.A = populateRecords(records, "A")
			case miekgdns.TypeAAAA:
				dnsRecords.AAAA = populateRecords(records, "AAAA")
			case miekgdns.TypeCNAME:
				dnsRecords.CNAME = populateRecords(records, "CNAME")
			case miekgdns.TypeMX:
				dnsRecords.MX = populateRecords(records, "MX")
			case miekgdns.TypeNS:
				dnsRecords.NS = populateRecords(records, "NS")
			case miekgdns.TypeTXT:
				dnsRecords.Txt = populateRecords(records, "TXT")
			case miekgdns.TypePTR:
				dnsRecords.PTR = populateRecords(records, "PTR")
			}
		}
	}

	return dnsRecords, nil
}
