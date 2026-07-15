package domainscan

import "strings"

// commonSubdomains lists subdomain labels ("<label>.<apex>") that vendors conventionally use
// instead of the apex (e.g. Resend's "send.<domain>", Stripe's "checkout.<domain>"). Naming
// conventions to probe, not a vendor lookup table — vendor names still come from whatever's found
var commonSubdomains = []string{
	"mail",
	"send",
	"email",
	"smtp",
	"mg",
	"bounce",
	"notifications",
	"updates",
	"news",
	"newsletter",
	"marketing",
	"transactional",
	"checkout",
	"status",
	"app",
	"portal",
	"api",
	"docs",
	"support",
	"billing",
	"account",
	"shop",
	"store",
	"blog",
	"help",
}

// systemSubdomainCandidates are subdomain labels commonly used for a company's own product
// surfaces (a web console, public API, docs, etc.), probed for HTTP reachability so the company
// profile prompt can be handed verified evidence instead of relying on whatever the model happens
// to notice linked from the single rendered homepage — this works for any company, not just one
var systemSubdomainCandidates = []string{
	"console",
	"app",
	"portal",
	"api",
	"graphql",
	"docs",
	"cli",
}

// dkimSelector is a DKIM selector label ("<selector>._domainkey.<host>") to probe for, paired with
// the vendor it's attributable to
var commonDKIMSelectors = []dkimSelector{
	{"google", "Google"},
	{"resend", "Resend"},
	{"mandrill", "Mandrill"},
	{"mailgun", "Mailgun"},
	{"mailo", "Mailo"},
	{"zoho", "Zoho"},
	{"amazonses", "Amazon SES"},
	{"pm", "Postmark"},
	{"selector1", "Microsoft 365"},
	{"selector2", "Microsoft 365"},
	{"s1", "SendGrid"},
	{"s2", "SendGrid"},
	{"k1", ""},
	{"k2", ""},
	{"k3", ""},
	{"default", ""},
	{"dkim", ""},
}

// dkimSelector pairs a DKIM selector label with its vendor, see commonDKIMSelectors
type dkimSelector struct {
	label  string
	vendor string
}

// vendorAlias is the source of truth for a known vendor: its canonical display
// name, any alternate names it's also known by (e.g. "The Open Lane" for
// "Openlane"), and every host or domain it's known to operate from
// TODO: use system owned vendors to do the lookup instead
type vendorAlias struct {
	name        string
	alsoKnownAs []string
	hosts       []string
	domains     []string
}

var knownVendorAliases = []vendorAlias{
	{name: "Google Workspace", hosts: []string{"admin.google.com"}},
	{name: "Google Cloud", alsoKnownAs: []string{"Googlecloud"}, hosts: []string{"cloud.google.com"}},
	{name: "Google Drive", alsoKnownAs: []string{"Googlecloud"}, hosts: []string{"drive.google.com"}},
	{name: "AWS", alsoKnownAs: []string{"Amazonses", "Amazon Web Services"}, domains: []string{"aws.amazon.com"}},
	{name: "Openlane", alsoKnownAs: []string{"The Open Lane"}, domains: []string{"theopenlane.io"}},
	{name: "Hubspot", alsoKnownAs: []string{"hubspotemail"}, domains: []string{"hubspot.com"}},
	{name: "Help Scout", alsoKnownAs: []string{"Helpscoutdocs"}, domains: []string{"helpscoutdocs.com"}},
	{name: "Atlassian Statuspage", alsoKnownAs: []string{"stspg-customer"}, domains: []string{"stspg-customer.com"}},
	{name: "Vercel", alsoKnownAs: []string{"vercel-dns"}, domains: []string{"vercel-dns.com"}},
	{name: "Stripe", alsoKnownAs: []string{"Stripecdn"}, domains: []string{"stripe.com"}},
}

// vendorHostNames overrides the display name derived from an exact hostname, derived from knownVendorAliases
var vendorHostNames = map[string]string{}

// vendorDomainNames overrides the display name derived from a registrable domain
var vendorDomainNames = map[string]string{}

// vendorNameDomains looks up a known vendor's domain given only its name (e.g. a technology
// or subprocessor name with no URL attached), the inverse of vendorDomainNames, keyed lowercase
var vendorNameDomains = map[string]string{}

// vendorCanonicalNames maps a known vendor's name and every alsoKnownAs alternate down to
// its single canonical display name, so alternate spellings collapse into one vendor group
var vendorCanonicalNames = map[string]string{}

func init() {
	for _, v := range knownVendorAliases {
		for _, host := range v.hosts {
			vendorHostNames[host] = v.name
		}

		for _, domain := range v.domains {
			vendorDomainNames[domain] = v.name
		}

		// prefer a domain for the name -> URL lookup, but fall back to a known host
		// (e.g. Google Cloud has no apex domain of its own, only cloud.google.com) so we
		// never leave a bare name with no known domain for a caller to guess at
		if _, ok := vendorNameDomains[strings.ToLower(v.name)]; !ok {
			switch {
			case len(v.domains) > 0:
				vendorNameDomains[strings.ToLower(v.name)] = v.domains[0]
			case len(v.hosts) > 0:
				vendorNameDomains[strings.ToLower(v.name)] = v.hosts[0]
			}
		}

		vendorCanonicalNames[strings.ToLower(v.name)] = v.name

		for _, alias := range v.alsoKnownAs {
			vendorCanonicalNames[strings.ToLower(alias)] = v.name
		}
	}
}
