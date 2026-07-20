package domainscan

type PromptType string

func getPrompt(t PromptType) string {
	switch t {
	case promptCompany:
		return companyProfilePrompt
	case promptCompliance:
		return compliancePagePrompt
	case promptTrustCenter:
		return trustCenterPrompt
	}

	return ""
}

const (
	promptCompany     PromptType = "COMPANY"
	promptCompliance  PromptType = "COMPLIANCE"
	promptTrustCenter PromptType = "TRUST_CENTER"
)

// companyProfilePrompt guides the AI to extract company information from a website
const companyProfilePrompt = `Extract company profile information from this website.

Basic info: company name, description, industry, headquarters location, approximate employee count range, founding year, and estimated revenue range.

Systems: the small number of distinct technical surfaces that make up the company's own product infrastructure — typically 2-5 of: a web console/dashboard application, a public API, a mobile app, a CLI tool, and the underlying data storage/database backend such as Redis, Postgres, RabbitMQ, etc. Use any subdomains found on the page (e.g. console.<domain>, app.<domain>, api.<domain>, docs.<domain>) to help identify which of these actually exist. Do NOT list the company's individual product modules, capabilities, or named marketing features as separate systems — for example, if the marketing page lists offerings like "Compliance Automation", "Policy Management", "Frameworks", "Trust Center", "Registry", or "Reporting", those are all features within the single web console, not separate systems, and must NOT each become their own entry. A company with one product typically has one system (the console), not one system per feature it markets. For each real system, give a brief 1-2 sentence summary and a fuller description of what it does and what data it handles, drawn from documentation or architecture pages when available.

Customers: named customers, clients, or case study companies referenced in logos, a "trusted by" section, testimonials, or case studies.

Technologies: third-party SaaS technologies or vendors the company itself relies on (e.g. analytics, CRM, hosting, payments).

Social links: LinkedIn, Twitter/X, GitHub, Discord, Instagram, YouTube, and Facebook profile links found in the header, footer, or about page.

Status page: the URL of a public status/uptime page, if one is linked anywhere on the site.

SSO/MFA support: whether the product advertises single sign-on (SSO) support or multi-factor authentication (MFA, 2FA) support, as true/false values based on explicit mentions in marketing, pricing, or documentation pages.`

// subprocessorsPromptText is the shared instruction for extracting subprocessor/vendor names,
// used by both the compliance page and trust center prompts so they ask for it identically
const subprocessorsPromptText = `Subprocessors: a list of only the short company or product names of subprocessors or third-party vendors explicitly listed on the page (e.g. "Amazon Web Services", "Stripe", "Google Cloud Platform"). Each entry must be a name only, never a sentence, explanation, or caveat about subprocessors. If the page does not explicitly list any subprocessors or vendors, return an empty list.`

// controlsPromptText is the shared instruction for extracting concrete security/operational controls,
// used by both the compliance page and trust center prompts so they ask for it identically
const controlsPromptText = `Controls: individual, concrete security or operational practices that this specific page explicitly describes — statements about access control, encryption, monitoring, testing, or training. Do not repeat framework or certification names here (e.g. do not list "SOC 2" or "ISO 27001 Certified" as a control; those belong only in the frameworks list and soc2_certified). Do not invent typical-sounding practices that aren't actually stated on the page; return an empty list if none are described.`

// compliancePagePrompt guides the AI to extract structured compliance information from a page
const compliancePagePrompt = `Analyze this compliance or legal page.

Page type: identify it as one of privacy_policy, terms_of_service, trust_center, dpa, soc2_report, security, subprocessors, gdpr, cookie_policy, or other.

Basic info: the page title and a brief summary.

Frameworks: any compliance frameworks or certifications the company claims to have achieved or comply with, for example SOC 2 Type I, SOC 2 Type II, SOC 1, ISO 27001, ISO 27017, ISO 27018, ISO 27701, ISO 42001, ISO 9001, ISO 22301, GDPR, CCPA, CPRA, PIPEDA, LGPD, POPIA, HIPAA, HITECH, HITRUST, PCI DSS, FedRAMP, StateRAMP, NIST 800-53, NIST CSF, NIST 800-171, CMMC, FISMA, CSA STAR, C5, TISAX, SOX, GLBA, FERPA, COPPA, or APEC CBPR. Specifically note whether the company claims SOC 2 (Type I or Type II) certification or compliance as a true/false value, the last updated or effective date, and URLs to downloadable documents or reports.

Other compliance links: find every other compliance-related link on the page (in the footer, nav, or body) such as privacy policies, terms of service, trust center pages, data processing agreements, subprocessor lists, cookie policies, or security pages, and classify each one using the same page type categories listed above. Actively look for a link to a trust center or trust portal (e.g. pages hosted on Openlane, Vanta, Drata, SafeBase, Whistic, or Conveyor, or a company's own /trust or /security page) even if it isn't the current page.

` + subprocessorsPromptText + `

` + controlsPromptText

// trustCenterPrompt guides the AI to extract structured information from a dedicated trust center or trust portal page
const trustCenterPrompt = `Analyze this trust center or trust portal page.

Hosted by: identify which platform or vendor is hosting this trust center. Look for any "Powered by", footer, or page-title attribution naming it (e.g. Openlane, Vanta, Drata, SafeBase, Whistic, Conveyor, TrustArc, OneTrust, SecurityPal, or any other vendor not listed here), not only the examples given. Respond "self-hosted" only if no such attribution appears anywhere on the page. This hosting platform is itself a vendor the company relies on.

Frameworks: extract all compliance frameworks or certifications listed, for example SOC 2 Type I, SOC 2 Type II, ISO 27001, ISO 27017, ISO 27018, ISO 27701, ISO 42001, ISO 9001, GDPR, CCPA, HIPAA, HITRUST, PCI DSS, FedRAMP, NIST 800-53, NIST CSF, or CSA STAR. Specifically note whether SOC 2 (Type I or Type II) certification or compliance is claimed as a true/false value.

` + controlsPromptText + `

` + subprocessorsPromptText
