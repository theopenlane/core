package email

import (
	"fmt"
	"strings"
	"time"

	"github.com/theopenlane/newman/render"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
)

// BrandingFromConfig converts a RuntimeEmailConfig into a render.Branding for use with the
// newman rendering engine
func BrandingFromConfig(config RuntimeEmailConfig) render.Branding {
	year := time.Now().Year()

	corp := config.Corporation
	if corp == "" {
		corp = config.CompanyName
	}

	copyright := fmt.Sprintf("Copyright (c) %d %s. All rights reserved", year, corp)

	return render.Branding{
		Name:        config.CompanyName,
		Link:        config.ProductURL,
		Logo:        config.LogoURL,
		Copyright:   copyright,
		TroubleText: "If you're having trouble with the button '{ACTION}', copy and paste the URL below into your web browser",
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

// brandingCSS generates CSS rules from EmailBranding color and font fields for injection
// into rendered HTML before CSS inlining. Uses common email-safe element selectors for
// broad applicability across customer-authored templates
func brandingCSS(eb *generated.EmailBranding) string {
	var b strings.Builder

	if eb.BackgroundColor != "" {
		fmt.Fprintf(&b, "body { background-color: %s; }\n", eb.BackgroundColor)
	}

	if eb.TextColor != "" {
		fmt.Fprintf(&b, "body, p, td, li { color: %s; }\n", eb.TextColor)
	}

	if eb.FontFamily != "" {
		if fontCSS := fontFamilyCSS(eb.FontFamily); fontCSS != "" {
			fmt.Fprintf(&b, "body, p, td { font-family: %s; }\n", fontCSS)
		}
	}

	if eb.LinkColor != "" {
		fmt.Fprintf(&b, "a { color: %s; }\n", eb.LinkColor)
	}

	if eb.ButtonColor != "" || eb.ButtonTextColor != "" {
		b.WriteString(".button, .btn {")

		if eb.ButtonColor != "" {
			fmt.Fprintf(&b, " background-color: %s;", eb.ButtonColor)
		}

		if eb.ButtonTextColor != "" {
			fmt.Fprintf(&b, " color: %s;", eb.ButtonTextColor)
		}

		b.WriteString(" }\n")
	}

	return b.String()
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
