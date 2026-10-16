package models

import "io"

// TrustCenterNDARequestSetting contains the possible approval rules
type TrustCenterNDARequestSetting struct {
	// AllowDisposableEmail enables known disposable emails to be used
	AllowDisposableEmail bool `json:"allowDisposableEmail"`

	// WorkEmailOnly when true will not auto-approve any free or disposable domains
	WorkEmailOnly bool `json:"workEmailOnly"`

	// AllowRoleAccount when false will not auto-approve any known role format like
	// support@, sales@ even if they pass other checks
	AllowRoleAccount bool `json:"allowRoleAccount"`

	// ApproveFromExistingRequestDomain enables auto approval if an existing approved
	// request already uses this domain, this is used in conjunction with the email rules above
	// so if you have an existing request from a @gmail.com email but you have WorkEmailOnly=true
	// this will not auto approve
	ApproveFromExistingRequestDomain bool `json:"approveFromExistingRequestDomain"`

	// ApproveIfContactExists enables auto approval if an active contact in the
	// organization matches the email
	ApproveIfContactExists bool `json:"approveIfContactExists"`

	// ApproveFromContactDomain enables auto approval if the domain exists
	// in the contacts, this is used in conjunction with the email rules above
	// so if you have a contact with a @gmail.com email but you have WorkEmailOnly=true
	// this will not auto approve
	ApproveFromContactDomain bool `json:"approveFromContactDomain"`

	// UseDomainBlocklist enables the usage of the domain blocklist to
	// automatically exclude certain domains, for example to not auto-approve your
	// competitor email domains
	UseDomainBlocklist bool `json:"useDomainBlocklist"`

	// DomainBlocklist contains the domains that are not allowed to be
	// automatically approved. This is only effective if UseDomainBlocklist is true
	DomainBlocklist []string `json:"domainBlocklist,omitempty"`

	// UseDomainAllowlist allows the usage of a whitelist to approve
	// nda requests from certain domains
	UseDomainAllowlist bool `json:"useDomainAllowlist"`

	// DomainAllowlist is the list of domains to automatically approve.
	// This is only enabled if UseDomainAllowlist is set to true
	DomainAllowlist []string `json:"domainAllowlist,omitempty"`

	// ManualApprovalOnFailure is used to set requests that do not
	// match the rules to a review
	ManualApprovalOnFailure bool `json:"manualApprovalOnFailure"`
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (s TrustCenterNDARequestSetting) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, s)
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (s *TrustCenterNDARequestSetting) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, s)
}
