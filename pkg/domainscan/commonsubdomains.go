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

// vendorAlias is the source of truth for a known vendor's name and every host
// or domain it's known to operate from
// TODO: use system owned vendors to do the lookup instead
type vendorAlias struct {
	name    string
	hosts   []string
	domains []string
}

var knownVendorAliases = []vendorAlias{
	{name: "Google Workspace", hosts: []string{"admin.google.com"}},
	{name: "Google Cloud", hosts: []string{"cloud.google.com"}},
	{name: "Amazon SES", domains: []string{"amazonses.com"}},
	{name: "Openlane", domains: []string{"theopenlane.io"}},
	{name: "The Open Lane", domains: []string{"theopenlane.io"}},
	{name: "Hubspot", domains: []string{"hubspotemail.net"}},
	{name: "Help Scout", domains: []string{"helpscoutdocs.com"}},
	{name: "Atlassian Statuspage", domains: []string{"stspg-customer.com"}},
	{name: "Vercel", domains: []string{"vercel-dns.com"}},
	{name: "Stripe", domains: []string{"stripecdn.com"}},
}

// vendorHostNames overrides the display name derived from an exact hostname, derived from knownVendorAliases
var vendorHostNames = map[string]string{}

// vendorDomainNames overrides the display name derived from a registrable domain
var vendorDomainNames = map[string]string{}

// vendorNameDomains looks up a known vendor's domain given only its name (e.g. a technology
// or subprocessor name with no URL attached), the inverse of vendorDomainNames, keyed lowercase
var vendorNameDomains = map[string]string{}

func init() {
	for _, v := range knownVendorAliases {
		for _, host := range v.hosts {
			vendorHostNames[host] = v.name
		}

		for _, domain := range v.domains {
			vendorDomainNames[domain] = v.name

			if _, ok := vendorNameDomains[strings.ToLower(v.name)]; !ok {
				vendorNameDomains[strings.ToLower(v.name)] = domain
			}
		}
	}
}
