package email

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/validator"
)

// BrandOverrides holds the customer-authored logo and color overrides shared by brandable email operations
type BrandOverrides struct {
	// LogoURL overrides the hero logo displayed in the email body for this send
	LogoURL string `json:"logoURL,omitempty" jsonschema:"format=uri,description=Hero logo URL override for this send"`
	// HeaderLogoURL overrides the small logo/icon in the top header bar for this send
	HeaderLogoURL string `json:"headerLogoURL,omitempty" jsonschema:"format=uri,description=Header bar logo/icon URL override for this send"`
	// PrimaryColor overrides the headline/emphasis color for this send
	PrimaryColor string `json:"primaryColor,omitempty" jsonschema:"format=color,description=Primary headline color override (hex)"`
	// BodyBackgroundColor overrides the outer page background color for this send
	BodyBackgroundColor string `json:"bodyBackgroundColor,omitempty" jsonschema:"format=color,description=Outer page background color override (hex)"`
	// CardBackgroundColor overrides the card container background color for this send
	CardBackgroundColor string `json:"cardBackgroundColor,omitempty" jsonschema:"format=color,description=Card container background color override (hex)"`
	// HeroBackgroundColor overrides the hero banner background color for this send
	HeroBackgroundColor string `json:"heroBackgroundColor,omitempty" jsonschema:"format=color,description=Hero banner background color override (hex)"`
	// ButtonColor overrides the call-to-action button background color for this send
	ButtonColor string `json:"buttonColor,omitempty" jsonschema:"format=color,description=CTA button background color override (hex)"`
	// ButtonTextColor overrides the call-to-action button text color for this send
	ButtonTextColor string `json:"buttonTextColor,omitempty" jsonschema:"format=color,description=CTA button text color override (hex)"`
	// TextColor overrides the body paragraph text color for this send
	TextColor string `json:"textColor,omitempty" jsonschema:"format=color,description=Body text color override (hex)"`
	// FooterTextColor overrides the muted text color for this send
	FooterTextColor string `json:"footerTextColor,omitempty" jsonschema:"format=color,description=Muted text color override (hex)"`
}

// apply overlays the set overrides onto the config, skipping invalid hex colors
func (b BrandOverrides) apply(cfg RuntimeEmailConfig) RuntimeEmailConfig {
	cfg.LogoURL = lo.CoalesceOrEmpty(b.LogoURL, cfg.LogoURL)
	cfg.HeaderLogoURL = lo.CoalesceOrEmpty(b.HeaderLogoURL, cfg.HeaderLogoURL)
	cfg.HeadingColor = overrideColor(b.PrimaryColor, cfg.HeadingColor)
	cfg.BodyBackgroundColor = overrideColor(b.BodyBackgroundColor, cfg.BodyBackgroundColor)
	cfg.CardBackgroundColor = overrideColor(b.CardBackgroundColor, cfg.CardBackgroundColor)
	cfg.HeroBackgroundColor = overrideColor(b.HeroBackgroundColor, cfg.HeroBackgroundColor)
	cfg.ButtonColor = overrideColor(b.ButtonColor, cfg.ButtonColor)
	cfg.ButtonTextColor = overrideColor(b.ButtonTextColor, cfg.ButtonTextColor)
	cfg.TextColor = overrideColor(b.TextColor, cfg.TextColor)
	cfg.FooterTextColor = overrideColor(b.FooterTextColor, cfg.FooterTextColor)

	return cfg
}

// overrideColor returns the override when it is a valid hex color, otherwise the fallback
func overrideColor(override, fallback string) string {
	if override == "" || validator.HexColorValidator(override) != nil {
		return fallback
	}

	return override
}
