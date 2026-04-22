package themes

import (
	"embed"

	"github.com/theopenlane/newman/render"
)

//go:embed templates/*
var themeFS embed.FS

// sharedText is the common text template used across all themes
var sharedText = readEmbed("templates/standard.tpl.txt")

// Standard is the primary Openlane branded email theme
var Standard = &render.Theme{
	Name: "openlane-standard",
	HTML: readEmbed("templates/standard.tpl.html"),
	Text: sharedText,
}

// TrustCenter is the trust center email theme with a simpler layout,
// no signature/help section, and a trust-center-specific footer
var TrustCenter = &render.Theme{
	Name: "openlane-trustcenter",
	HTML: readEmbed("templates/trustcenter.tpl.html"),
	Text: sharedText,
}

// Questionnaire is the questionnaire email theme with a simple layout
// and questionnaire-specific footer
var Questionnaire = &render.Theme{
	Name: "openlane-questionnaire",
	HTML: readEmbed("templates/questionnaire.tpl.html"),
	Text: sharedText,
}

// Customer is a neutral, brand-configurable theme for customer-authored
// email templates. Uses per-send Config and Style fields for identity
// and colors without Openlane-specific chrome
var Customer = &render.Theme{
	Name: "customer",
	HTML: readEmbed("templates/customer.tpl.html"),
	Text: sharedText,
}

// ModernMessage is a centered single-panel theme with brand header, hero logo,
// title, body paragraphs, optional CTA button, fine-print outros, and a full
// footer with tagline, socials, address, and unsubscribe. Suitable for short
// transactional or marketing messages that need a single call to action
var ModernMessage = &render.Theme{
	Name: "modern-message",
	HTML: readEmbed("templates/modern-message.tpl.html"),
	Text: sharedText,
}

// ModernHero is a hero-image-forward theme with a full-width header image,
// large 40px headline, body paragraphs, optional CTA, and the shared modern
// footer. Suitable for launches, announcements, and welcome-shaped messages
var ModernHero = &render.Theme{
	Name: "modern-hero",
	HTML: readEmbed("templates/modern-hero.tpl.html"),
	Text: sharedText,
}

// readEmbed reads a file from the embedded theme filesystem
func readEmbed(name string) string {
	data, err := themeFS.ReadFile(name)
	if err != nil {
		panic("themes: " + err.Error())
	}

	return string(data)
}
