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
	Name:   "openlane-standard",
	HTML:   readEmbed("templates/standard.tpl.html"),
	Text:   sharedText,
	Styles: render.ParseStyleMap(readEmbed("templates/standard.css")),
}

// TrustCenter is the trust center email theme with a simpler layout,
// no signature/help section, and a trust-center-specific footer
var TrustCenter = &render.Theme{
	Name:   "openlane-trustcenter",
	HTML:   readEmbed("templates/trustcenter.tpl.html"),
	Text:   sharedText,
	Styles: render.ParseStyleMap(readEmbed("templates/trustcenter.css")),
}

// Questionnaire is the questionnaire email theme with a simple layout
// and questionnaire-specific footer
var Questionnaire = &render.Theme{
	Name:   "openlane-questionnaire",
	HTML:   readEmbed("templates/questionnaire.tpl.html"),
	Text:   sharedText,
	Styles: render.ParseStyleMap(readEmbed("templates/questionnaire.css")),
}

// Customer is a neutral, brand-configurable theme for customer-authored
// email templates. Uses EmailBranding fields for identity and colors
// without Openlane-specific chrome
var Customer = &render.Theme{
	Name:   "customer",
	HTML:   readEmbed("templates/customer.tpl.html"),
	Text:   sharedText,
	Styles: render.ParseStyleMap(readEmbed("templates/customer.css")),
}

// readEmbed reads a file from the embedded theme filesystem
func readEmbed(name string) string {
	data, err := themeFS.ReadFile(name)
	if err != nil {
		panic("themes: " + err.Error())
	}

	return string(data)
}
