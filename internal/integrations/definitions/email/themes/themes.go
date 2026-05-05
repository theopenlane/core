package themes

import (
	"embed"

	"github.com/theopenlane/newman/render"
)

//go:embed templates/*
var themeFS embed.FS

// sharedText is the common text template used across all themes
var sharedText = readEmbed("templates/standard.tpl.txt")

// Base is the default email theme with brand header, hero logo, title, body
// paragraphs, optional CTA button, fine-print outros, and a full footer with
// tagline, socials, address, and unsubscribe
var Base = &render.Theme{
	Name: "base",
	HTML: readEmbed("templates/base.tpl.html"),
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
