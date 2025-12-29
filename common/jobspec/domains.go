package jobspec

// CreateCustomDomainArgs for the worker to process the custom domain
type CreateCustomDomainArgs struct {
	// ID of the custom domain in our system
	CustomDomainID string `json:"custom_domain_id"`
}

// Kind satisfies the river.Job interface
func (CreateCustomDomainArgs) Kind() string { return "create_custom_domain" }

// CreatePreviewDomainArgs for the worker to process the preview domain creation
type CreatePreviewDomainArgs struct {
	// TrustCenterID is the ID of the trust center to create a preview domain for
	TrustCenterID string `json:"trust_center_id"`
	// TrustCenterPreviewZoneID is the cloudflare zone id for the trust center preview domain
	TrustCenterPreviewZoneID string `json:"trust_center_preview_zone_id"`
	// TrustCenterCnameTarget is the cname target for the trust center
	TrustCenterCnameTarget string `json:"trust_center_cname_target"`
}

// Kind satisfies the river.Job interface
func (CreatePreviewDomainArgs) Kind() string { return "create_preview_domain" }

// DeleteCustomDomainArgs for the worker to process the custom domain
type DeleteCustomDomainArgs struct {
	// CustomDomainID of the id of the custom domain in the system
	CustomDomainID string `json:"custom_domain_id"`
	// DNSVerificationID of the dns verification id in the system
	DNSVerificationID string `json:"dns_verification_id"`
	// CloudflareCustomHostnameID of the cloudflare custom hostname id to delete
	CloudflareCustomHostnameID string `json:"cloudflare_custom_hostname_id"`
	// CloudflareZoneID of the cloudflare zone id where the custom domain is located
	CloudflareZoneID string `json:"cloudflare_zone_id"`
}

// Kind satisfies the river.Job interface
func (DeleteCustomDomainArgs) Kind() string { return "delete_custom_domain" }

// DeletePreviewDomainArgs for the worker to process the preview domain deletion
type DeletePreviewDomainArgs struct {
	// CustomDomainID is the ID of the custom domain to delete
	CustomDomainID string `json:"custom_domain_id"`
	// TrustCenterPreviewZoneID is the cloudflare zone id for the trust center preview domain
	TrustCenterPreviewZoneID string `json:"trust_center_preview_zone_id"`
}

// Kind satisfies the river.Job interface
func (DeletePreviewDomainArgs) Kind() string { return "delete_preview_domain" }

// ValidateCustomDomainArgs for the worker to process the custom domain
type ValidateCustomDomainArgs struct {
	CustomDomainID string `json:"custom_domain_id"`
}

// Kind satisfies the river.Job interface
func (ValidateCustomDomainArgs) Kind() string { return "validate_custom_domain" }

// ValidatePreviewDomainArgs for the worker to process the preview domain creation
type ValidatePreviewDomainArgs struct {
	// TrustCenterID is the ID of the trust center to validate the preview domain for
	TrustCenterID string `json:"trust_center_id"`
	// TrustCenterPreviewZoneID is the cloudflare zone id for the trust center preview domain
	TrustCenterPreviewZoneID string `json:"trust_center_preview_zone_id"`
}

// Kind satisfies the river.Job interface
func (ValidatePreviewDomainArgs) Kind() string { return "validate_preview_domain" }

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

// SyncTrustCenterCacheArgs for the worker to refresh trust center cache entries
type SyncTrustCenterCacheArgs struct {
	// TrustCenterID is the ID of the trust center to refresh cache for
	TrustCenterID string `json:"trust_center_id"`
}

// Kind satisfies the river.Job interface
func (SyncTrustCenterCacheArgs) Kind() string { return "sync_trust_center_cache" }
