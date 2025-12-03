package corejobs

// ClearTrustCenterCacheArgs for the worker to clear trust center cache
type ClearTrustCenterCacheArgs struct {
	// CustomDomain is the custom domain for the trust center
	// If provided, will clear cache for this custom domain
	CustomDomain string `json:"custom_domain,omitempty"`

	// TrustCenterSlug is the slug for the trust center
	// Used with default domain: trust.theopenlane.net/<trust center slug>
	// If CustomDomain is not provided, this will be used
	TrustCenterSlug string `json:"trust_center_slug,omitempty"`
}

// Kind satisfies the river.Job interface
func (ClearTrustCenterCacheArgs) Kind() string { return "clear_trust_center_cache" }
