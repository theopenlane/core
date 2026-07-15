package domainscan

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

// vendorHostNames overrides the display name derived from an exact hostname
// TODO: use system owned vendors to do the lookup instead
var vendorHostNames = map[string]string{
	"admin.google.com": "Google Workspace",
	"cloud.google.com": "Google Cloud",
}

// vendorDomainNames overrides the display name derived from a registrable domain
// when naively title-casing its first label (see domainVendorName and
// vendorNameFromHostname) produces something other than the vendor's actual name
// TODO: use system owned vendors to do the lookup instead
var vendorDomainNames = map[string]string{
	"amazonses.com":      "Amazon SES",
	"theopenlane.io":     "Openlane",
	"hubspotemail.net":   "Hubspot",
	"helpscoutdocs.com":  "Help Scout",
	"stspg-customer.com": "Atlassian Statuspage",
	"vercel-dns.com":     "Vercel",
}
