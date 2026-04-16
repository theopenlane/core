package email

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
)

func TestBrandingFromConfig_AllFields(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName:    "Acme Corp",
		ProductURL:     "https://app.acme.com",
		LogoURL:        "https://cdn.acme.com/logo.png",
		Copyright:      "2025 Acme",
		TroubleText:    "Having trouble? Click here.",
		CompanyAddress: "123 Main St",
		RootURL:        "https://acme.com",
		UnsubscribeURL: "https://acme.com/unsub",
	}

	b := BrandingFromConfig(cfg)

	assert.Equal(t, "Acme Corp", b.Name)
	assert.Equal(t, "https://app.acme.com", b.Link)
	assert.Equal(t, "https://cdn.acme.com/logo.png", b.Logo)
	assert.Equal(t, "2025 Acme", b.Copyright)
	assert.Equal(t, "Having trouble? Click here.", b.TroubleText)
	assert.Equal(t, "123 Main St", b.CompanyAddress)
	assert.Equal(t, "https://acme.com", b.RootURL)
	assert.Equal(t, "https://acme.com/unsub", b.UnsubscribeURL)
}

func TestBrandingFromConfig_EmptyConfig(t *testing.T) {
	b := BrandingFromConfig(RuntimeEmailConfig{})

	assert.Empty(t, b.Name)
	assert.Empty(t, b.Link)
	assert.Empty(t, b.Logo)
}

func TestBrandingFromConfig_PartialFields(t *testing.T) {
	cfg := RuntimeEmailConfig{
		CompanyName: "Partial",
		RootURL:     "https://partial.io",
	}

	b := BrandingFromConfig(cfg)

	assert.Equal(t, "Partial", b.Name)
	assert.Equal(t, "https://partial.io", b.RootURL)
	assert.Empty(t, b.Link)
	assert.Empty(t, b.Logo)
}

func TestFontFamilyCSS_Helvetica(t *testing.T) {
	for _, font := range []enums.Font{
		enums.FontHelvetica, enums.FontHelveticaBold,
		enums.FontHelveticaBoldOblique, enums.FontHelveticaOblique,
	} {
		t.Run(string(font), func(t *testing.T) {
			result := fontFamilyCSS(font)
			assert.Equal(t, "'Helvetica Neue', Helvetica, Arial, sans-serif", result)
		})
	}
}

func TestFontFamilyCSS_Courier(t *testing.T) {
	for _, font := range []enums.Font{
		enums.FontCourier, enums.FontCourierBold,
		enums.FontCourierBoldOblique, enums.FontCourierOblique,
	} {
		t.Run(string(font), func(t *testing.T) {
			result := fontFamilyCSS(font)
			assert.Equal(t, "'Courier New', Courier, monospace", result)
		})
	}
}

func TestFontFamilyCSS_TimesRoman(t *testing.T) {
	for _, font := range []enums.Font{
		enums.FontTimesRoman, enums.FontTimesBold,
		enums.FontTimesBoldItalic, enums.FontTimesItalic,
	} {
		t.Run(string(font), func(t *testing.T) {
			result := fontFamilyCSS(font)
			assert.Equal(t, "'Times New Roman', Times, serif", result)
		})
	}
}

func TestFontFamilyCSS_Symbol(t *testing.T) {
	assert.Equal(t, "Symbol, sans-serif", fontFamilyCSS(enums.FontSymbol))
}

func TestFontFamilyCSS_Unknown(t *testing.T) {
	assert.Empty(t, fontFamilyCSS(enums.FontInvalid))
	assert.Empty(t, fontFamilyCSS(enums.Font("NONEXISTENT")))
	assert.Empty(t, fontFamilyCSS(enums.Font("")))
}
