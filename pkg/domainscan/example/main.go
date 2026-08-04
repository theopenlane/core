package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	cf "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/cloudflare/cloudflare-go/v7/url_scanner"
	"github.com/olekukonko/tablewriter"

	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/pkg/domainscan"
	"github.com/theopenlane/core/pkg/jsonx"
)

func main() {
	domain := flag.String("domain", "", "domain URL to scan, e.g. https://example.com")
	kind := flag.String("kind", "", `which data to fetch: "company", "compliance", "dns", "urlscan", or empty for the full combined report`)
	format := flag.String("format", "table", `output format for the full combined report: "table" or "json" (other -kind values always print json)`)
	flag.Parse()

	if *domain == "" {
		log.Fatal("usage: go run . -domain https://example.com [-kind company|compliance|dns|urlscan] [-format table|json] [-debug]")
	}

	switch strings.ToLower(*format) {
	case "table", "json":
	default:
		log.Fatalf(`unknown -format %q: must be "table" or "json"`, *format)
	}

	// dns doesn't call the Cloudflare API, so it doesn't need credentials
	if strings.ToLower(*kind) == "dns" {
		printDNSVendorInfo(*domain)
		return
	}

	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	if apiToken == "" || accountID == "" {
		log.Fatal("CLOUDFLARE_API_TOKEN and CLOUDFLARE_ACCOUNT_ID must be set")
	}

	cfg := &domainscan.Config{
		APIToken:  apiToken,
		AccountID: accountID,
	}

	switch strings.ToLower(*kind) {
	case "company":
		printCompanyProfile(cfg, *domain)
	case "compliance":
		printComplianceData(cfg, *domain)
	case "urlscan":
		printURLScan(cfg, *domain)
	case "", "report":
		printFullReport(cfg, *domain, strings.ToLower(*format))
	default:
		log.Fatalf(`unknown -kind %q: must be "company", "compliance", "dns", "urlscan", or empty for the full combined report`, *kind)
	}
}

func printCompanyProfile(cfg *domainscan.Config, domain string) {
	profile, err := cfg.GetCompanyData(context.Background(), domain)
	if err != nil {
		log.Fatalf("browser rendering failed: %v", err)
	}

	out, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal profile: %v", err)
	}

	fmt.Println("------- Company Profile --------")
	fmt.Println(string(out))
}

func printComplianceData(cfg *domainscan.Config, domain string) {
	comp, err := cfg.GetComplianceData(context.Background(), domain)
	if err != nil {
		log.Fatalf("browser rendering failed: %v", err)
	}

	out, err := json.MarshalIndent(comp, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal profile: %v", err)
	}

	fmt.Println("------- Compliance Data --------")
	fmt.Println(string(out))
}

// printURLScan submits domain to Cloudflare's URL Scanner and polls for the
// result, printing it once the scan completes
func printURLScan(cfg *domainscan.Config, domain string) {
	result := submitAndPollScan(cfg, domain)

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal scan result: %v", err)
	}

	fmt.Println("------- URL Scan Result --------")
	fmt.Println(string(out))
}

// printFullReport submits domain for a URL scan, gathers the company profile,
// compliance, and dns enrichment concurrently, and prints the same combined
// report the retrieve_domain_scan worker sends to Openlane as a set of tables,
// one per section, rather than as raw JSON
func printFullReport(cfg *domainscan.Config, domain, format string) {
	var (
		result     *url_scanner.ScanGetResponse
		enrichment domainscan.Enrichment
		wg         sync.WaitGroup
	)

	wg.Add(2)

	// kick off the URL scan (submit + poll) and the enrichment lookups at the
	// same time, since both can take minutes on their own - by the time the
	// slower of the two finishes, the other has likely already completed
	go func() {
		defer wg.Done()

		result = submitAndPollScan(cfg, domain)
	}()

	go func() {
		defer wg.Done()

		fmt.Println("gathering company profile, compliance, and dns enrichment...")

		enrichment, _ = cfg.GatherEnrichment(context.Background(), domain, 3*time.Minute)
	}()

	wg.Wait()

	report := domainscan.BuildScanReport(result, enrichment, nil, nil)

	if format == "json" {
		out, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			log.Fatalf("failed to marshal report: %v", err)
		}

		fmt.Println("\n------- Full Report --------")
		fmt.Println(string(out))

		return
	}

	var typed domainscan.ScanReport
	if err := jsonx.RoundTrip(report, &typed); err != nil {
		log.Fatalf("failed to decode report: %v", err)
	}

	fmt.Println("\n------- Full Report --------")
	printReportTables(typed)
}

// reportTableMaxWidth bounds each table's total width so long cells (e.g. a
// product description, a joined list of controls) wrap instead of stretching
// the table past a normal terminal width
const reportTableMaxWidth = 120

// newTable creates a tablewriter table writing to stdout with the given column headers
func newTable(headers ...string) *tablewriter.Table {
	t := tablewriter.NewTable(os.Stdout, tablewriter.WithMaxWidth(reportTableMaxWidth))
	t.Header(headers)

	return t
}

// printSectionHeading prints a heading line above a report section's table(s)
func printSectionHeading(title string) {
	fmt.Printf("\n-- %s --\n", title)
}

// joinOrDash joins items with ", ", or returns "-" when there's nothing to show
func joinOrDash(items []string) string {
	if len(items) == 0 {
		return "-"
	}

	return strings.Join(items, ", ")
}

// printReportTables renders every section of a BuildScanReport result as its own table,
// skipping sections the report doesn't contain
func printReportTables(report domainscan.ScanReport) {
	printSummaryTable(report)
	printVendorsTable(report.Vendors)
	printTechnologiesTable(report.Technologies)
	printAssetsTables(report.Assets)
	printFindingsTables(report.Findings)
	printMetaTable(report.Meta)
	printPlatformTable(report.Platform)
	printSystemsTable(report.Systems)
	printComplianceTable(report.Compliance)
}

func printSummaryTable(report domainscan.ScanReport) {
	printSectionHeading("Summary")

	t := newTable("Field", "Value")
	t.Append([]string{"Scan ID", report.ExternalScanID})
	t.Append([]string{"URL", report.URL})
	t.Render()
}

func printVendorsTable(vendors []domainscan.Vendor) {
	if len(vendors) == 0 {
		return
	}

	printSectionHeading("Vendors")

	t := newTable("Name", "URL", "Categories")
	for _, v := range vendors {
		t.Append([]string{v.Name, v.URL, joinOrDash(v.Categories)})
	}

	t.Render()
}

func printTechnologiesTable(technologies []domainscan.Technology) {
	if len(technologies) == 0 {
		return
	}

	printSectionHeading("Technologies")

	t := newTable("Name", "URL", "Categories")
	for _, tech := range technologies {
		t.Append([]string{tech.Name, tech.URL, joinOrDash(tech.Categories)})
	}

	t.Render()
}

func printAssetsTables(assets *domainscan.Assets) {
	if assets == nil {
		return
	}

	printSectionHeading("Assets")

	if len(assets.DNSRecords) > 0 {
		t := newTable("Domain", "Type", "Vendor")
		for _, r := range assets.DNSRecords {
			t.Append([]string{r.Domain, r.Type, r.Vendor})
		}

		t.Render()
	}

	if len(assets.IPAddresses) > 0 {
		t := newTable("Address", "ASN", "Org")
		for _, ip := range assets.IPAddresses {
			t.Append([]string{ip.Address, ip.ASN, ip.Org})
		}

		t.Render()
	}
}

func printFindingsTables(findings domainscan.Findings) {
	printSectionHeading("Findings")

	t := newTable("Field", "Value")
	t.Append([]string{"Security Violations", joinOrDash(findings.SecurityViolations)})
	t.Append([]string{"Risks", joinOrDash(findings.Risks)})
	t.Append([]string{"Malicious", fmt.Sprint(findings.IsMalicious)})
	t.Render()

	if findings.MissingComplianceLinks != "" {
		fmt.Println("\nMissing Compliance Links:")
		fmt.Println(findings.MissingComplianceLinks)
	}

	for _, agentReadiness := range findings.AgentReadiness {
		fmt.Printf("\nAgent Readiness: level %v (%v)\n", agentReadiness.Level, agentReadiness.LevelName)

		if agentReadiness.Checklist != "" {
			fmt.Println(agentReadiness.Checklist)
		}

		if agentReadiness.Reference != "" {
			fmt.Printf("Reference: %s\n", agentReadiness.Reference)
		}
	}
}

func printMetaTable(meta *domainscan.Meta) {
	if meta == nil {
		return
	}

	printSectionHeading("Meta")

	t := newTable("Field", "Value")

	if meta.Rank != 0 {
		t.Append([]string{"Rank", fmt.Sprint(meta.Rank)})
	}

	if len(meta.URLCategories) > 0 {
		t.Append([]string{"URL Categories", joinOrDash(meta.URLCategories)})
	}

	if len(meta.DomainCategories) > 0 {
		t.Append([]string{"Domain Categories", joinOrDash(meta.DomainCategories)})
	}

	if geo := meta.Geolocation; geo != nil {
		t.Append([]string{"City", geo.City})
		t.Append([]string{"Country", fmt.Sprintf("%v (%v)", geo.CountryName, geo.Country)})
		t.Append([]string{"Region", geo.Region})
		t.Append([]string{"Coordinates", fmt.Sprintf("%v, %v", geo.Latitude, geo.Longitude)})
	}

	t.Render()
}

func printPlatformTable(platform *domainscan.Platform) {
	if platform == nil {
		return
	}

	printSectionHeading("Platform")

	t := newTable("Field", "Value")
	t.Append([]string{"Name", platform.Name})
	t.Append([]string{"Description", platform.Description})
	t.Append([]string{"Industry", platform.Industry})
	t.Append([]string{"Location", platform.Location})
	t.Append([]string{"Employees", platform.EmployeeRange})
	t.Append([]string{"Founded", platform.FoundedYear})
	t.Append([]string{"Est. Revenue", platform.EstimatedRevenue})
	t.Append([]string{"SSO Supported", fmt.Sprint(platform.SSOSupported)})
	t.Append([]string{"MFA Supported", fmt.Sprint(platform.MFASupported)})

	if platform.StatusPageURL != "" {
		t.Append([]string{"Status Page", platform.StatusPageURL})
	}

	if len(platform.Customers) > 0 {
		t.Append([]string{"Customers", joinOrDash(platform.Customers)})
	}

	t.Render()

	social := platform.SocialLinks

	type socialLink struct {
		network string
		url     string
	}

	links := []socialLink{
		{"LinkedIn", social.LinkedIn},
		{"Twitter", social.Twitter},
		{"GitHub", social.GitHub},
		{"Discord", social.Discord},
		{"Instagram", social.Instagram},
		{"YouTube", social.YouTube},
		{"Facebook", social.Facebook},
	}

	st := newTable("Network", "URL")

	hasAny := false

	for _, link := range links {
		if link.url == "" {
			continue
		}

		hasAny = true

		st.Append([]string{link.network, link.url})
	}

	if hasAny {
		st.Render()
	}
}

func printSystemsTable(systems []domainscan.SystemEntry) {
	if len(systems) == 0 {
		return
	}

	printSectionHeading("Systems")

	t := newTable("System", "Description")
	for _, s := range systems {
		t.Append([]string{s.SystemName, s.Description})
	}

	t.Render()
}

func printComplianceTable(compliance *domainscan.Compliance) {
	if compliance == nil {
		return
	}

	printSectionHeading("Compliance")

	t := newTable("Field", "Value")
	t.Append([]string{"Frameworks", joinOrDash(compliance.Frameworks)})
	t.Append([]string{"SOC 2 Certified", fmt.Sprint(compliance.IsSOC2)})
	t.Append([]string{"Controls", joinOrDash(compliance.Controls)})

	if compliance.TrustCenterHostedBy != "" {
		t.Append([]string{"Trust Center Hosted By", compliance.TrustCenterHostedBy})
	}

	t.Render()

	if len(compliance.Documents) == 0 {
		return
	}

	dt := newTable("Document", "Public", "URL")
	for _, d := range compliance.Documents {
		dt.Append([]string{d.Name, fmt.Sprint(d.Public), d.URL})
	}

	dt.Render()
}

// domainScanHTTPTimeout bounds the Cloudflare API client used for domain scan calls
const domainScanHTTPTimeout = time.Minute

// newCloudflareClient builds a Cloudflare API client from an API token and account ID, the same
// way internal/integrations/definitions/cloudflare builds one from the operator-owned runtime config
func newCloudflareClient(apiToken, accountID string) *cloudflare.CloudflareClient {
	return &cloudflare.CloudflareClient{
		Client: cf.NewClient(
			option.WithAPIToken(apiToken),
			option.WithHTTPClient(&http.Client{Timeout: domainScanHTTPTimeout}),
		),
		Config: cloudflare.ClientConfig{
			AccountID: accountID,
			APIToken:  apiToken,
		},
	}
}

// submitAndPollScan submits domain to Cloudflare's URL Scanner and polls
// until the scan completes, returning the final result. This calls the same
// cloudflare.DomainScanSubmit / cloudflare.DomainScanPoll operations that
// internal/integrations/runtime/domain_scan.go runs from Gala listeners, but
// polls in a blocking loop here instead of scheduling itself as a job
func submitAndPollScan(cfg *domainscan.Config, domain string) *url_scanner.ScanGetResponse {
	ctx := context.Background()

	client := newCloudflareClient(cfg.APIToken, cfg.AccountID)

	submitResult, err := cloudflare.DomainScanSubmit{}.Run(ctx, client, cloudflare.DomainScanSubmit{
		Domains: []string{domain},
	})
	if err != nil {
		log.Fatalf("failed to submit scan: %v", err)
	}

	if len(submitResult.Scans) == 0 {
		log.Fatal("no scan was created")
	}

	scanID := submitResult.Scans[0].UUID

	fmt.Printf("submitted scan %s for %s, polling for results...\n", scanID, domain)

	for attempt := 0; attempt < cloudflare.DomainScanMaxAttempts; attempt++ {
		pollResult, err := cloudflare.DomainScanPoll{}.Run(ctx, client, cloudflare.DomainScanPoll{
			ScanResultID: scanID,
		})
		if err != nil {
			log.Fatalf("failed to get scan result: %v", err)
		}

		if len(pollResult.TaskErrors) > 0 {
			log.Fatalf("scan failed: %s", pollResult.TaskErrors.Error())
		}

		if pollResult.Result.Task.Success {
			return pollResult.Result
		}

		time.Sleep(cloudflare.DomainScanPollBackoff(attempt))
	}

	log.Fatal("timed out waiting for scan to complete")

	return nil
}

func printDNSVendorInfo(domain string) {
	info, err := domainscan.GetDNSVendorInfo(context.Background(), domain)
	if err != nil {
		log.Fatalf("dns lookup failed: %v", err)
	}

	out, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal dns vendor info: %v", err)
	}

	fmt.Println("------- DNS Vendor Info --------")
	fmt.Println(string(out))
}
