package ssoutils

import (
	"strings"

	"github.com/samber/lo"
)

// Exemption reasons describe why a subject is not subject to the SSO login redirect
const (
	// ExemptReasonOwner is set when the subject is the organization owner
	ExemptReasonOwner = "owner"
	// ExemptReasonUser is set when the subject's membership carries an explicit SSO exemption
	ExemptReasonUser = "user_exempt"
	// ExemptReasonDomain is set when the subject is a member whose email domain is exempt
	ExemptReasonDomain = "domain_exempt"
	// ExemptReasonSupport is set when the request is an active Openlane support session
	ExemptReasonSupport = "support"
)

// EnforcementInput carries the organization and membership facts needed to decide how a
// subject must authenticate against an organization. It is built from the organization
// setting and the subject's membership so this package stays free of ent dependencies
type EnforcementInput struct {
	// SSOEnforced reports whether the organization has SSO login enforcement enabled
	SSOEnforced bool
	// IDPAuthTested reports whether the SSO connection has been verified; SSO is only effective when tested
	IDPAuthTested bool
	// TFAEnforced reports whether the organization enforces multifactor authentication
	TFAEnforced bool
	// ExemptDomains is the set of email domains whose existing members skip the SSO redirect
	ExemptDomains []string
	// IsMember reports whether the subject is already a member of the organization
	IsMember bool
	// IsOwner reports whether the subject is the organization owner
	IsOwner bool
	// MemberExempt reports whether the subject's membership carries an SSO exemption
	MemberExempt bool
	// Email is the subject's email address used for domain based exemption checks
	Email string
	// SupportSession reports whether the request is an active Openlane support session
	SupportSession bool
}

// Decision is the resolved authentication routing decision for a subject and organization
type Decision struct {
	// SSOEnforced reports whether SSO enforcement is effective (enabled and connection tested)
	SSOEnforced bool
	// Exempt reports whether the subject is exempt from the SSO login redirect
	Exempt bool
	// ExemptReason explains why the subject is exempt; empty when not exempt
	ExemptReason string
	// MustSSO reports whether the subject must be redirected through the SSO login flow
	MustSSO bool
	// TFARequired reports whether the subject must satisfy multifactor authentication; this is
	// independent of SSO exemption
	TFARequired bool
}

// Evaluate resolves how a subject must authenticate against an organization. SSO exemption only
// affects whether the subject is routed through the directory; multifactor enforcement applies
// regardless of exemption
func Evaluate(in EnforcementInput) Decision {
	ssoEnforced := in.SSOEnforced && in.IDPAuthTested

	d := Decision{
		SSOEnforced: ssoEnforced,
		TFARequired: in.TFAEnforced,
	}

	switch {
	case in.IsOwner:
		d.Exempt = true
		d.ExemptReason = ExemptReasonOwner
	case in.MemberExempt:
		d.Exempt = true
		d.ExemptReason = ExemptReasonUser
	case in.IsMember && domainMatches(emailDomain(in.Email), in.ExemptDomains):
		d.Exempt = true
		d.ExemptReason = ExemptReasonDomain
	case in.SupportSession:
		d.Exempt = true
		d.ExemptReason = ExemptReasonSupport
	}

	d.MustSSO = ssoEnforced && !d.Exempt

	return d
}

// emailDomain returns the lowercased domain portion of an email, or empty when it cannot be parsed
func emailDomain(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return ""
	}

	return strings.ToLower(email[at+1:])
}

// domainMatches reports whether domain is non-empty and present in candidates, case-insensitively
func domainMatches(domain string, candidates []string) bool {
	if domain == "" {
		return false
	}

	return lo.ContainsBy(candidates, func(c string) bool {
		return strings.EqualFold(strings.TrimSpace(c), domain)
	})
}
