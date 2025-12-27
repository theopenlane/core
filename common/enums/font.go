package enums

import (
	"fmt"
	"io"
	"strings"
)

// Font is a custom type for font names
type Font string

var (
	// FontCourier represents the Courier font
	FontCourier Font = "COURIER"
	// FontCourierBold represents the Courier-Bold font
	FontCourierBold Font = "COURIER_BOLD"
	// FontCourierBoldOblique represents the Courier-BoldOblique font
	FontCourierBoldOblique Font = "COURIER_BOLDOBLIQUE"
	// FontCourierOblique represents the Courier-Oblique font
	FontCourierOblique Font = "COURIER_OBLIQUE"
	// FontHelvetica represents the Helvetica font
	FontHelvetica Font = "HELVETICA"
	// FontHelveticaBold represents the Helvetica-Bold font
	FontHelveticaBold Font = "HELVETICA_BOLD"
	// FontHelveticaBoldOblique represents the Helvetica-BoldOblique font
	FontHelveticaBoldOblique Font = "HELVETICA_BOLDOBLIQUE"
	// FontHelveticaOblique represents the Helvetica-Oblique font
	FontHelveticaOblique Font = "HELVETICA_OBLIQUE"
	// FontSymbol represents the Symbol font
	FontSymbol Font = "SYMBOL"
	// FontTimesBold represents the Times-Bold font
	FontTimesBold Font = "TIMES_BOLD"
	// FontTimesBoldItalic represents the Times-BoldItalic font
	FontTimesBoldItalic Font = "TIMES_BOLDITALIC"
	// FontTimesItalic represents the Times-Italic font
	FontTimesItalic Font = "TIMES_ITALIC"
	// FontTimesRoman represents the Times-Roman font
	FontTimesRoman Font = "TIMES_ROMAN"
	// FontInvalid indicates that the font is invalid
	FontInvalid Font = "FONT_INVALID"
)

// Values returns a slice of strings that represents all the possible values of the Font enum.
func (Font) Values() (kinds []string) {
	for _, s := range []Font{
		FontCourier, FontCourierBold, FontCourierBoldOblique, FontCourierOblique,
		FontHelvetica, FontHelveticaBold, FontHelveticaBoldOblique, FontHelveticaOblique,
		FontSymbol, FontTimesBold, FontTimesBoldItalic, FontTimesItalic, FontTimesRoman,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the font as a string
func (r Font) String() string {
	return string(r)
}

// ToFontStr converts the enum to the supported font string format
func (r Font) ToFontStr() string {
	switch r {
	case FontCourier:
		return "Courier"
	case FontCourierBold:
		return "Courier-Bold"
	case FontCourierBoldOblique:
		return "Courier-BoldOblique"
	case FontCourierOblique:
		return "Courier-Oblique"
	case FontHelvetica:
		return "Helvetica"
	case FontHelveticaBold:
		return "Helvetica-Bold"
	case FontHelveticaBoldOblique:
		return "Helvetica-BoldOblique"
	case FontHelveticaOblique:
		return "Helvetica-Oblique"
	case FontSymbol:
		return "Symbol"
	case FontTimesBold:
		return "Times-Bold"
	case FontTimesBoldItalic:
		return "Times-BoldItalic"
	case FontTimesItalic:
		return "Times-Italic"
	case FontTimesRoman:
		return "Times-Roman"
	default:
		return ""
	}
}

// ToFont returns the font enum based on string input
func ToFont(r string) *Font {
	switch r := strings.ToUpper(r); r {
	case FontCourier.String():
		return &FontCourier
	case FontCourierBold.String():
		return &FontCourierBold
	case FontCourierBoldOblique.String():
		return &FontCourierBoldOblique
	case FontCourierOblique.String():
		return &FontCourierOblique
	case FontHelvetica.String():
		return &FontHelvetica
	case FontHelveticaBold.String():
		return &FontHelveticaBold
	case FontHelveticaBoldOblique.String():
		return &FontHelveticaBoldOblique
	case FontHelveticaOblique.String():
		return &FontHelveticaOblique
	case FontSymbol.String():
		return &FontSymbol
	case FontTimesBold.String():
		return &FontTimesBold
	case FontTimesBoldItalic.String():
		return &FontTimesBoldItalic
	case FontTimesItalic.String():
		return &FontTimesItalic
	case FontTimesRoman.String():
		return &FontTimesRoman
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r Font) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *Font) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for Font, got: %T", v) //nolint:err113
	}

	*r = Font(str)

	return nil
}
