package email

import (
	"fmt"
	"net/url"
	"time"

	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

// BrandingFromConfig converts RuntimeEmailConfig into render.Branding for use with the newman rendering engine
func BrandingFromConfig(config RuntimeEmailConfig, recipientEmail string) render.Branding {
	year := time.Now().Year()

	corp := config.Corporation
	if corp == "" {
		corp = config.CompanyName
	}

	copyright := fmt.Sprintf("\u00A9 %d %s All rights reserved.", year, corp)

	unsubscribeURL := fmt.Sprintf("%s/unsubscribe?email=%s", config.ProductURL, url.QueryEscape(recipientEmail))

	return render.Branding{
		Name:           config.CompanyName,
		Link:           config.ProductURL,
		Logo:           config.LogoURL,
		Copyright:      copyright,
		TroubleText:    "If you're having trouble with the button '{ACTION}', copy and paste the URL below into your web browser",
		CompanyAddress: config.CompanyAddress,
		RootURL:        config.RootURL,
		UnsubscribeURL: unsubscribeURL,
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

// selectDefaultBranding picks the default EmailBranding from a slice, falling back to the
// first element when no default is flagged. Returns nil for empty slices
func selectDefaultBranding(brandings []*generated.EmailBranding) *generated.EmailBranding {
	if len(brandings) == 0 {
		return nil
	}

	for _, eb := range brandings {
		if eb.IsDefault {
			return eb
		}
	}

	return brandings[0]
}
