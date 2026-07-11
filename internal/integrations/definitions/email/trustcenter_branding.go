package email

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/generated"
)

// TrustCenterBranding carries the trust center identity applied to subscriber-facing emails. Trust
// center palettes are authored for the web, where dark themes are common, so emails never hand the
// configured colors control of the containers, text, buttons, or backgrounds — the accent color is
// applied only to decorative details (borders), where any value is safe
type TrustCenterBranding struct {
	// CompanyName overrides the footer company name with the trust center's branding when present
	CompanyName string `json:"companyName,omitempty" jsonschema:"description=Company display name shown in the footer"`
	// LogoURL overrides the hero and header logo with the trust center's branding when present
	LogoURL string `json:"logoURL,omitempty" jsonschema:"format=uri,description=Hero logo URL override"`
	// AccentColor is the trust center's accent, applied to decorative borders only
	AccentColor string `json:"accentColor,omitempty" jsonschema:"format=color,description=Accent color applied to decorative borders (hex)"`
}

// TrustCenterBrandingFromSetting maps a trust center setting's identity onto the email branding
// overlay; a nil setting yields the zero overlay so every value falls back to the system branding.
// The setting's background, foreground, and button-driving colors are web-only and never map into
// emails; the accent (falling back to the primary color) surfaces only as a decorative border
func TrustCenterBrandingFromSetting(setting *generated.TrustCenterSetting) TrustCenterBranding {
	if setting == nil {
		return TrustCenterBranding{}
	}

	return TrustCenterBranding{
		CompanyName: setting.CompanyName,
		LogoURL:     lo.FromPtr(setting.LogoRemoteURL),
		AccentColor: lo.CoalesceOrEmpty(setting.AccentColor, setting.PrimaryColor),
	}
}

// apply overlays the trust center identity onto the runtime email config: company name and logo are
// honored, and the accent color is exposed for decorative borders. Headings, text, backgrounds, and
// buttons keep the config's system values. The logo renders once, in the upper-left header slot;
// the in-body hero logo is suppressed so trust center emails never show a double logo
func (b TrustCenterBranding) apply(cfg RuntimeEmailConfig) RuntimeEmailConfig {
	cfg.CompanyName = lo.CoalesceOrEmpty(b.CompanyName, cfg.CompanyName)
	cfg.HeaderLogoURL = lo.CoalesceOrEmpty(b.LogoURL, cfg.HeaderLogoURL, cfg.LogoURL)
	cfg.LogoURL = ""
	cfg.AccentBorderColor = lo.CoalesceOrEmpty(b.AccentColor, cfg.AccentBorderColor)

	return cfg
}

// trustCenterEmailConfig applies the shared trust center branding overlay and the per-recipient
// unsubscribe link to the runtime email config, the common Config hook body for every trust center
// subscriber email. Openlane's platform marketing chrome (tagline, social links, header text) is
// stripped: these messages speak for the trust center's company, not the platform
func trustCenterEmailConfig(cfg RuntimeEmailConfig, b TrustCenterBranding, unsubscribeURL string) RuntimeEmailConfig {
	cfg = b.apply(cfg)

	cfg.Tagline = ""
	cfg.Social = nil
	cfg.HeaderText = ""

	if unsubscribeURL != "" {
		cfg.UnsubscribeURL = unsubscribeURL
	}

	return cfg
}
