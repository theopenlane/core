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
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/domainscan"
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

	fmt.Println("\n------- Full Report --------")
	printReportTables(report)
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

// stringSlice type-asserts v to []string, returning nil if it isn't one
func stringSlice(v any) []string {
	s, _ := v.([]string)
	return s
}

// printReportTables renders every section of a BuildScanReport result as its own table,
// skipping sections the report doesn't contain
func printReportTables(report map[string]any) {
	printSummaryTable(report)
	printVendorsTable(report)
	printTechnologiesTable(report)
	printAssetsTables(report)
	printFindingsTables(report)
	printMetaTable(report)
	printPlatformTable(report)
	printSystemsTable(report)
	printComplianceTable(report)
}

func printSummaryTable(report map[string]any) {
	printSectionHeading("Summary")

	t := newTable("Field", "Value")
	t.Append([]string{"Scan ID", fmt.Sprint(report["external_scan_id"])})
	t.Append([]string{"URL", fmt.Sprint(report["url"])})
	t.Render()
}

func printVendorsTable(report map[string]any) {
	vendors, ok := report["vendors"].([]map[string]any)
	if !ok || len(vendors) == 0 {
		return
	}

	printSectionHeading("Vendors")

	t := newTable("Name", "URL", "Categories")
	for _, v := range vendors {
		t.Append([]string{fmt.Sprint(v["name"]), fmt.Sprint(v["url"]), joinOrDash(stringSlice(v["categories"]))})
	}

	t.Render()
}

func printTechnologiesTable(report map[string]any) {
	technologies, ok := report["technologies"].([]map[string]any)
	if !ok || len(technologies) == 0 {
		return
	}

	printSectionHeading("Technologies")

	t := newTable("Name", "URL", "Categories")
	for _, tech := range technologies {
		t.Append([]string{fmt.Sprint(tech["name"]), fmt.Sprint(tech["url"]), joinOrDash(stringSlice(tech["categories"]))})
	}

	t.Render()
}

func printAssetsTables(report map[string]any) {
	assets, ok := report["assets"].(map[string]any)
	if !ok {
		return
	}

	printSectionHeading("Assets")

	if dnsRecords, ok := assets["dns_records"].([]map[string]string); ok && len(dnsRecords) > 0 {
		t := newTable("Domain", "Type")
		for _, r := range dnsRecords {
			t.Append([]string{r["domain"], r["type"]})
		}

		t.Render()
	}

	if ipAddresses, ok := assets["ip_addresses"].([]map[string]string); ok && len(ipAddresses) > 0 {
		t := newTable("Address", "ASN", "Org")
		for _, ip := range ipAddresses {
			t.Append([]string{ip["address"], ip["asn"], ip["org"]})
		}

		t.Render()
	}

	if internalDomains := stringSlice(assets["internal_domains"]); len(internalDomains) > 0 {
		t := newTable("Internal Domain")
		for _, d := range internalDomains {
			t.Append([]string{d})
		}

		t.Render()
	}
}

func printFindingsTables(report map[string]any) {
	findings, ok := report["findings"].(map[string]any)
	if !ok {
		return
	}

	printSectionHeading("Findings")

	malicious, _ := findings["is_malicious"].(bool)

	t := newTable("Field", "Value")
	t.Append([]string{"Security Violations", joinOrDash(stringSlice(findings["security_violations"]))})
	t.Append([]string{"Risks", joinOrDash(stringSlice(findings["risks"]))})
	t.Append([]string{"Malicious", fmt.Sprint(malicious)})
	t.Append([]string{"Missing Compliance Links", joinOrDash(stringSlice(findings["missing_compliance_links"]))})
	t.Render()

	agentReadiness, ok := findings["agent_readiness"].(map[string]any)
	if !ok {
		return
	}

	fmt.Printf("\nAgent Readiness: level %v (%v)\n", agentReadiness["level"], agentReadiness["level_name"])

	if checklist, ok := agentReadiness["checklist"].(string); ok && checklist != "" {
		fmt.Println(checklist)
	}

	if reference, ok := agentReadiness["reference"].(string); ok && reference != "" {
		fmt.Printf("Reference: %s\n", reference)
	}
}

func printMetaTable(report map[string]any) {
	meta, ok := report["meta"].(map[string]any)
	if !ok {
		return
	}

	printSectionHeading("Meta")

	t := newTable("Field", "Value")

	if rank, ok := meta["rank"]; ok {
		t.Append([]string{"Rank", fmt.Sprint(rank)})
	}

	if urlCategories := stringSlice(meta["url_categories"]); len(urlCategories) > 0 {
		t.Append([]string{"URL Categories", joinOrDash(urlCategories)})
	}

	if domainCategories := stringSlice(meta["domain_categories"]); len(domainCategories) > 0 {
		t.Append([]string{"Domain Categories", joinOrDash(domainCategories)})
	}

	if geo, ok := meta["geolocation"].(map[string]any); ok {
		t.Append([]string{"City", fmt.Sprint(geo["city"])})
		t.Append([]string{"Country", fmt.Sprintf("%v (%v)", geo["country_name"], geo["country"])})
		t.Append([]string{"Region", fmt.Sprint(geo["region"])})
		t.Append([]string{"Coordinates", fmt.Sprintf("%v, %v", geo["latitude"], geo["longitude"])})
	}

	t.Render()
}

func printPlatformTable(report map[string]any) {
	platform, ok := report["platform"].(map[string]any)
	if !ok {
		return
	}

	printSectionHeading("Platform")

	t := newTable("Field", "Value")
	t.Append([]string{"Name", fmt.Sprint(platform["name"])})
	t.Append([]string{"Description", fmt.Sprint(platform["description"])})
	t.Append([]string{"Industry", fmt.Sprint(platform["industry"])})
	t.Append([]string{"Location", fmt.Sprint(platform["location"])})
	t.Append([]string{"Employees", fmt.Sprint(platform["employee_range"])})
	t.Append([]string{"Founded", fmt.Sprint(platform["founded_year"])})
	t.Append([]string{"Est. Revenue", fmt.Sprint(platform["estimated_revenue"])})
	t.Append([]string{"SSO Supported", fmt.Sprint(platform["sso_supported"])})
	t.Append([]string{"MFA Supported", fmt.Sprint(platform["mfa_supported"])})

	if statusPage, ok := platform["status_page_url"].(string); ok && statusPage != "" {
		t.Append([]string{"Status Page", statusPage})
	}

	if customers := stringSlice(platform["customers"]); len(customers) > 0 {
		t.Append([]string{"Customers", joinOrDash(customers)})
	}

	t.Render()

	social, ok := platform["social_links"].(domainscan.SocialLinks)
	if !ok {
		return
	}

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

func printSystemsTable(report map[string]any) {
	systems, ok := report["systems"].([]map[string]any)
	if !ok || len(systems) == 0 {
		return
	}

	printSectionHeading("Systems")

	t := newTable("System", "Description")
	for _, s := range systems {
		t.Append([]string{fmt.Sprint(s["system_name"]), fmt.Sprint(s["description"])})
	}

	t.Render()
}

func printComplianceTable(report map[string]any) {
	compliance, ok := report["compliance"].(map[string]any)
	if !ok {
		return
	}

	printSectionHeading("Compliance")

	t := newTable("Field", "Value")
	t.Append([]string{"Frameworks", joinOrDash(stringSlice(compliance["frameworks"]))})
	t.Append([]string{"SOC 2 Certified", fmt.Sprint(compliance["is_soc2"])})
	t.Append([]string{"Controls", joinOrDash(stringSlice(compliance["controls"]))})

	if hostedBy, ok := compliance["trust_center_hosted_by"].(string); ok && hostedBy != "" {
		t.Append([]string{"Trust Center Hosted By", hostedBy})
	}

	t.Render()

	documents, ok := compliance["documents"].([]domainscan.TrustDocument)
	if !ok || len(documents) == 0 {
		return
	}

	dt := newTable("Document", "Public", "URL")
	for _, d := range documents {
		dt.Append([]string{d.Name, fmt.Sprint(d.Public), d.URL})
	}

	dt.Render()
}

// domainScanHTTPTimeout bounds the Cloudflare API client used for domain scan calls
const domainScanHTTPTimeout = time.Minute

// newCloudflareClient builds a Cloudflare API client from an API token, the same way
// internal/integrations/runtime builds one from the operator-owned runtime config
func newCloudflareClient(apiToken string) *cf.Client {
	return cf.NewClient(
		option.WithAPIToken(apiToken),
		option.WithHTTPClient(&http.Client{Timeout: domainScanHTTPTimeout}),
	)
}

// submitAndPollScan submits domain to Cloudflare's URL Scanner and polls
// until the scan completes, returning the final result. This calls the same
// cloudflare.DomainScanSubmit / cloudflare.DomainScanPoll operations that
// internal/integrations/runtime/domain_scan.go runs from Gala listeners, but
// polls in a blocking loop here instead of scheduling itself as a job
func submitAndPollScan(cfg *domainscan.Config, domain string) *url_scanner.ScanGetResponse {
	ctx := context.Background()

	client := newCloudflareClient(cfg.APIToken)

	submitResult, err := cloudflare.DomainScanSubmit{}.Run(ctx, client, cloudflare.DomainScanSubmit{
		AccountID: cfg.AccountID,
		Domains:   []string{domain},
	})
	if err != nil {
		log.Fatalf("failed to submit scan: %v", err)
	}

	if len(submitResult.Scans) == 0 {
		log.Fatal("no scan was created")
	}

	scanID := submitResult.Scans[0].UUID

	fmt.Printf("submitted scan %s for %s, polling for results...\n", scanID, domain)

	time.Sleep(operations.DomainScanInitialWait)

	for attempt := 0; attempt < operations.DomainScanMaxAttempts; attempt++ {
		pollResult, err := cloudflare.DomainScanPoll{}.Run(ctx, client, cloudflare.DomainScanPoll{
			AccountID:    cfg.AccountID,
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

		time.Sleep(operations.DomainScanPollBackoff(attempt))
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
