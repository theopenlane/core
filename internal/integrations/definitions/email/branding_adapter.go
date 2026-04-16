package email

import (
	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/common/enums"
)

// BrandingFromConfig converts RuntimeEmailConfig into render.Branding for use with the newman rendering engine
func BrandingFromConfig(config RuntimeEmailConfig) render.Branding {
	return render.Branding{
		Name:           config.CompanyName,
		Link:           config.ProductURL,
		Logo:           config.LogoURL,
		Copyright:      config.Copyright,
		TroubleText:    config.TroubleText,
		CompanyAddress: config.CompanyAddress,
		RootURL:        config.RootURL,
		UnsubscribeURL: config.UnsubscribeURL,
	}
}

// fontFamilyCSS converts a Font enum value to a CSS font-family string with fallbacks
func fontFamilyCSS(font enums.Font) string {
	switch font {
	case enums.FontHelvetica, enums.FontHelveticaBold, enums.FontHelveticaBoldOblique, enums.FontHelveticaOblique:
		return "'Helvetica Neue', Helvetica, Arial, sans-serif"
	case enums.FontCourier, enums.FontCourierBold, enums.FontCourierBoldOblique, enums.FontCourierOblique:
		return "'Courier New', Courier, monospace"
	case enums.FontTimesRoman, enums.FontTimesBold, enums.FontTimesBoldItalic, enums.FontTimesItalic:
		return "'Times New Roman', Times, serif"
	case enums.FontSymbol:
		return "Symbol, sans-serif"
	default:
		return ""
	}
}
