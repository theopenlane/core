// Package trustcenterurl builds public trust center URLs from server-level configuration. It is a
// leaf package so both the ent hooks and the integration runtime can construct trust center links
// (e.g. tokenized unsubscribe links) without an import cycle
package trustcenterurl

import (
	"fmt"
	"net/url"
)

// Config holds the trust center URL-building configuration, set once at server startup
type Config struct {
	// PreviewZoneID is the cloudflare zone id used for preview domains
	PreviewZoneID string
	// CnameTarget is the CNAME target custom domains point at
	CnameTarget string
	// DefaultTrustCenterDomain is the shared domain slugged trust centers are served from
	DefaultTrustCenterDomain string
	// CacheRefreshScheme is the URL scheme (http in tests, https otherwise)
	CacheRefreshScheme string
}


var config Config

// SetConfig sets the package-level trust center configuration
func SetConfig(c Config) {
	config = c
}

// GetConfig returns the current trust center configuration
func GetConfig() Config {
	return config
}

// BuildURL constructs the public trust center URL from a custom domain or slug, returning empty when
// neither resolves
func BuildURL(customDomain, slug string) string {
	scheme := config.CacheRefreshScheme
	if scheme == "" {
		scheme = "https"
	}

	// in test mode (http scheme) use the default domain for all requests
	if scheme == "http" && config.DefaultTrustCenterDomain != "" {
		return fmt.Sprintf("%s://%s", scheme, config.DefaultTrustCenterDomain)
	}

	if customDomain != "" {
		return fmt.Sprintf("%s://%s", scheme, customDomain)
	}

	if slug != "" && config.DefaultTrustCenterDomain != "" {
		return fmt.Sprintf("%s://%s/%s", scheme, config.DefaultTrustCenterDomain, slug)
	}

	return ""
}

// UnsubscribeURL builds the tokenized unsubscribe link for a trust center. The {{ .unsubscribeToken }}
// placeholder is interpolated per recipient at email render time. Returns empty when the trust center
// URL cannot be resolved
func UnsubscribeURL(customDomain, slug string) string {
	base := BuildURL(customDomain, slug)
	if base == "" {
		return ""
	}

	return fmt.Sprintf("%s/unsubscribe?token={{ .unsubscribeToken }}", base)
}

// UnsubscribeURLWithToken builds the unsubscribe link for a trust center with a concrete per-recipient
// token. It is used by direct system sends that do not run template interpolation (unlike campaign
// sends, which carry the {{ .unsubscribeToken }} placeholder). Returns empty when the URL cannot resolve
func UnsubscribeURLWithToken(customDomain, slug, token string) string {
	base := BuildURL(customDomain, slug)
	if base == "" {
		return ""
	}

	return fmt.Sprintf("%s/unsubscribe?token=%s", base, url.QueryEscape(token))
}
