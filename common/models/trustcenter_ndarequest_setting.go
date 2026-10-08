package models

import "io"

// TrustCenterNDARequestSetting contains the manu possible approval rules
type TrustCenterNDARequestSetting struct {

	// AllowDisposableEmail enables known disposable emails to be used
	AllowDisposableEmail bool `json:"allowDisposableEmail"`

	// WorkEmailOnly restricts the auto approval to allow for only work emails
	WorkEmailOnly bool `json:"workEmailOnly"`

	// AllowRoleAccount if the domain matches a known role format like
	// support@, sales@
	AllowRoleAccount bool `json:"allowRoleAccount"`

	// ApproveFromExistingRequestDomain enables auto approval if an existing approved
	// request already uses this domain
	ApproveFromExistingRequestDomain bool `json:"approveFromExistingRequestDomain"`

	// ApproveIfContactExists enables auto approval if an active contact in the
	// organization matches the email
	ApproveIfContactExists bool `json:"approveIfContactExists"`

	// ApproveFromContactDomain enables auto approval if the domain exists
	// in the contacts
	ApproveFromContactDomain bool `json:"approveFromContactDomain"`

	// UseDomainBlocklist enables the usage of the domain blocklist to
	// automatically exclude certain domains. This could be
	// for competitors or others
	UseDomainBlocklist bool `json:"useDomainBlocklist"`

	// DomainBlocklist contains the domains that are not allowed to be
	// automatically approved. This is only effective if UseDomainBlocklist is true
	DomainBlocklist []string `json:"domainBlocklist,omitempty"`

	// UseDomainBlocklist allows the usage of a whitelist to approve
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
