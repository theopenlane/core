package domainscan

import "fmt"

// Finding severities for posture checks
const (
	PostureSeverityHigh   = "high"
	PostureSeverityMedium = "medium"
	PostureSeverityLow    = "low"
)

// buildPostureFindings raises one finding per failing email authentication or well-known
// file check. Checks whose input was never probed are skipped rather than reported as failing
func buildPostureFindings(enrichment Enrichment) []PostureFinding {
	var findings []PostureFinding

	if auth := buildEmailAuth(enrichment); auth != nil {
		findings = append(findings, emailAuthFindings(*auth)...)
	}

	if enrichment.WellKnown != nil {
		findings = append(findings, wellKnownFindings(*enrichment.WellKnown)...)
	}

	return findings
}

// emailAuthFindings reports DMARC, SPF and DKIM gaps. DKIM is only assessed for domains
// with mail exchangers, since a domain that receives no mail is not expected to sign any
func emailAuthFindings(auth EmailAuth) []PostureFinding {
	var findings []PostureFinding

	switch {
	case auth.DMARCRecord == "" || auth.DMARCPolicy == "":
		findings = append(findings, PostureFinding{
			Check:       "dmarc",
			Title:       "DMARC: Not published",
			Description: "No DMARC record was found. Receiving mail servers have no instruction about what to do with messages forging this domain, so most will deliver them. Publish a DMARC record at p=none with rua reporting, then step to quarantine and reject.",
			Severity:    PostureSeverityHigh,
		})
	case auth.DMARCPolicy == DMARCPolicyNone:
		findings = append(findings, PostureFinding{
			Check:       "dmarc",
			Title:       "DMARC: Not enforcing",
			Description: "The DMARC policy is p=none, which asks receivers to report spoofed mail but not block it. Review the aggregate reports for legitimate senders, then move to p=quarantine and finally p=reject.",
			Severity:    PostureSeverityHigh,
		})
	case !auth.Enforces():
		findings = append(findings, PostureFinding{
			Check:       "dmarc",
			Title:       "DMARC: Partially enforced",
			Description: fmt.Sprintf("The DMARC policy is p=%s but pct=%d applies it to only that share of failing messages; the rest are delivered as though there were no policy. Finish the rollout by setting pct=100.", auth.DMARCPolicy, auth.DMARCPercentage),
			Severity:    PostureSeverityMedium,
		})
	}

	switch auth.SPFPolicy {
	case "":
		findings = append(findings, PostureFinding{
			Check:       "spf",
			Title:       "SPF: Not published",
			Description: "No SPF record was found, so there is no published list of who may send mail on this domain's behalf and DMARC has one less signal to work with. Publish a v=spf1 record listing the sending services and ending in ~all.",
			Severity:    PostureSeverityHigh,
		})
	case SPFPolicyPassAll, SPFPolicyNeutral:
		findings = append(findings, PostureFinding{
			Check:       "spf",
			Title:       "SPF: Permissive",
			Description: fmt.Sprintf("The SPF record ends in a qualifier that authorizes any host to send mail as this domain (%s). Change the trailing mechanism to ~all or -all.", auth.SPFRecord),
			Severity:    PostureSeverityHigh,
		})
	}

	if len(auth.DKIMSelectors) == 0 && len(auth.MXHosts) > 0 {
		findings = append(findings, PostureFinding{
			Check:       "dkim",
			Title:       "DKIM: Not detected",
			Description: "No DKIM selectors were found at the conventional names. Without DKIM signing, DMARC has less to work with and legitimate forwarded mail is more likely to fail. Enable DKIM on every sending service.",
			Severity:    PostureSeverityMedium,
		})
	}

	return findings
}

// wellKnownFindings reports missing or weak security.txt, HTTPS enforcement, llms.txt and
// AI crawler policy. A nil probe means nothing was checked and raises nothing
func wellKnownFindings(wk WellKnown) []PostureFinding {
	var findings []PostureFinding

	if wk.SecurityTxt != nil {
		switch {
		case !wk.SecurityTxt.Present:
			findings = append(findings, PostureFinding{
				Check:       "security_txt",
				Title:       "security.txt: Not published",
				Description: "No security.txt was found. A researcher who discovers a vulnerability has no obvious place to report it. Publish /.well-known/security.txt with a Contact and Expires field per RFC 9116.",
				Severity:    PostureSeverityMedium,
			})
		case wk.SecurityTxt.Expired:
			findings = append(findings, PostureFinding{
				Check:       "security_txt",
				Title:       "security.txt: Expired",
				Description: "The published security.txt has an Expires date in the past, which tells researchers to treat it as stale. Update the Expires field.",
				Severity:    PostureSeverityLow,
			})
		}
	}

	if t := wk.Transport; t != nil && t.HTTPSReachable {
		switch {
		case t.HSTS:
		case t.RedirectsToHTTPS:
			findings = append(findings, PostureFinding{
				Check:       "https",
				Title:       "HTTPS: No HSTS header",
				Description: "The site redirects to HTTPS but does not send a Strict-Transport-Security header, so a visitor's first request on a new network can still be intercepted. Add an HSTS header with a max-age of at least one year.",
				Severity:    PostureSeverityMedium,
			})
		default:
			findings = append(findings, PostureFinding{
				Check:       "https",
				Title:       "HTTPS: Not enforced",
				Description: "The site does not redirect plain HTTP to HTTPS. Traffic can be read or modified on a shared or hostile network. Redirect all HTTP requests to HTTPS and add an HSTS header.",
				Severity:    PostureSeverityHigh,
			})
		}
	}

	if wk.LLMs != nil && !wk.LLMs.Present {
		findings = append(findings, PostureFinding{
			Check:       "llms_txt",
			Title:       "llms.txt: Not published",
			Description: "No llms.txt was found. AI agents asked about this company fall back to whatever they can scrape. Publish /llms.txt with a title, a one-line summary and links to the pages that matter.",
			Severity:    PostureSeverityLow,
		})
	}

	if wk.Robots != nil && !wk.Robots.StatesAIPosition() {
		title := "AI crawler policy: None stated"
		description := "robots.txt says nothing about AI crawlers, and every major AI crawler treats silence as permission. Add explicit rules for the AI user agents this company wants to allow or block."

		if !wk.Robots.Present {
			title = "AI crawler policy: No robots.txt"
			description = "No robots.txt was found, so there is no stated position on any crawler. Publish one with explicit rules for AI user agents."
		}

		findings = append(findings, PostureFinding{
			Check:       "ai_crawlers",
			Title:       title,
			Description: description,
			Severity:    PostureSeverityLow,
		})
	}

	return findings
}
