package enums

import (
	"fmt"
	"io"
	"strings"
)

// Font is a custom type for font names
type Font string

var (
	// FontArial represents the Arial font
	FontArial Font = "arial"
	// FontHelvetica represents the Helvetica font
	FontHelvetica Font = "helvetica"
	// FontTimes represents the Times font
	FontTimes Font = "times"
	// FontTimesNewRoman represents the Times New Roman font
	FontTimesNewRoman Font = "times new roman"
	// FontGeorgia represents the Georgia font
	FontGeorgia Font = "georgia"
	// FontVerdana represents the Verdana font
	FontVerdana Font = "verdana"
	// FontCourier represents the Courier font
	FontCourier Font = "courier"
	// FontCourierNew represents the Courier New font
	FontCourierNew Font = "courier new"
	// FontTrebuchetMS represents the Trebuchet MS font
	FontTrebuchetMS Font = "trebuchet ms"
	// FontComicSansMS represents the Comic Sans MS font
	FontComicSansMS Font = "comic sans ms"
	// FontImpact represents the Impact font
	FontImpact Font = "impact"
	// FontPalatino represents the Palatino font
	FontPalatino Font = "palatino"
	// FontGaramond represents the Garamond font
	FontGaramond Font = "garamond"
	// FontBookman represents the Bookman font
	FontBookman Font = "bookman"
	// FontAvantGarde represents the Avant Garde font
	FontAvantGarde Font = "avant garde"
	// FontInvalid indicates that the font is invalid
	FontInvalid Font = "FONT_INVALID"
)

// Values returns a slice of strings that represents all the possible values of the Font enum.
func (Font) Values() (kinds []string) {
	for _, s := range []Font{
		FontArial, FontHelvetica, FontTimes, FontTimesNewRoman, FontGeorgia,
		FontVerdana, FontCourier, FontCourierNew, FontTrebuchetMS, FontComicSansMS,
		FontImpact, FontPalatino, FontGaramond, FontBookman, FontAvantGarde,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the font as a string
func (r Font) String() string {
	return string(r)
}

// ToFont returns the font enum based on string input
func ToFont(r string) *Font {
	switch r := strings.ToLower(r); r {
	case FontArial.String():
		return &FontArial
	case FontHelvetica.String():
		return &FontHelvetica
	case FontTimes.String():
		return &FontTimes
	case FontTimesNewRoman.String():
		return &FontTimesNewRoman
	case FontGeorgia.String():
		return &FontGeorgia
	case FontVerdana.String():
		return &FontVerdana
	case FontCourier.String():
		return &FontCourier
	case FontCourierNew.String():
		return &FontCourierNew
	case FontTrebuchetMS.String():
		return &FontTrebuchetMS
	case FontComicSansMS.String():
		return &FontComicSansMS
	case FontImpact.String():
		return &FontImpact
	case FontPalatino.String():
		return &FontPalatino
	case FontGaramond.String():
		return &FontGaramond
	case FontBookman.String():
		return &FontBookman
	case FontAvantGarde.String():
		return &FontAvantGarde
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
