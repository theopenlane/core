package docextract

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/theopenlane/core/v2/pkg/pdftext"
)

var (
	// whitespacePattern collapses runs of whitespace so phrases split across lines still match
	whitespacePattern = regexp.MustCompile(`\s+`)
	// curlyApostrophes are normalized to a plain apostrophe before matching possessives
	curlyApostrophes = strings.NewReplacer("’", "'", "‘", "'", "ʼ", "'")
)

// profiles maps each document kind to the profile that decides whether an upload is that kind
var profiles = map[string]Profile{}

// RegisterProfile adds or replaces the profile for a document kind
func RegisterProfile(kind string, profile Profile) {
	profiles[kind] = profile
}

// Validate checks that a pdf is the expected kind of document, so a caller can fail fast with a
// readable reason before any extraction is scheduled
func Validate(pdf []byte, kind string) (Validation, error) {
	profile, ok := profiles[kind]
	if !ok {
		return Validation{Kind: kind}, fmt.Errorf("%w: %s", ErrNoProfile, kind)
	}

	return profile.Validate(pdf, kind)
}

// Validate applies the profile to the pdf
func (p Profile) Validate(pdf []byte, kind string) (Validation, error) {
	result := Validation{Kind: kind, ReportType: ReportTypeUnknown}

	document, err := pdftext.Open(pdf)
	if err != nil {
		return result, &ValidationError{Err: fmt.Errorf("%w: %w", ErrUnreadable, err), Reason: ErrUnreadable.Error()}
	}

	result.Pages = document.PageCount()

	if result.Pages < p.MinPages {
		return result, &ValidationError{Err: ErrTooShort, Reason: fmt.Sprintf("the uploaded file has fewer than %d pages and does not look like a %s", p.MinPages, p.Kind)}
	}

	text := normalizeText(document.Text())
	result.TextLength = len(bytes.TrimSpace(text))

	if result.TextLength < MinExtractedText {
		return result, &ValidationError{Err: ErrNoText, Reason: ErrNoText.Error()}
	}

	found := p.matchMarkers(text)

	for _, marker := range p.Markers {
		switch {
		case marker.Weight == MarkerRequired && !found[marker.Name]:
			return result, &ValidationError{Err: ErrNotMatched, Reason: fmt.Sprintf("the uploaded file does not appear to be a %s, it never references %s", p.Kind, marker.Name)}
		case !found[marker.Name]:
			continue
		case marker.Weight == MarkerStrong:
			result.Strong = append(result.Strong, marker.Name)
		case marker.Weight == MarkerSupporting:
			result.Supporting = append(result.Supporting, marker.Name)
		}
	}

	if len(result.Strong) < p.MinStrong {
		reason := fmt.Sprintf("the uploaded file is missing the sections expected in a %s", p.Kind)
		if p.MissingSectionsHint != "" {
			reason += ", such as " + p.MissingSectionsHint
		}

		return result, &ValidationError{Err: ErrMissingSections, Reason: reason}
	}

	result.Confidence = ConfidenceAccepted
	if len(result.Strong) >= p.HighConfidence {
		result.Confidence = ConfidenceHigh
	}

	if p.Subtype != nil {
		result.ReportType = p.Subtype(found)
	}

	return result, nil
}

// matchMarkers reports which markers appear in the extracted text
func (p Profile) matchMarkers(text []byte) map[string]bool {
	found := make(map[string]bool, len(p.Markers))

	for _, marker := range p.Markers {
		found[marker.Name] = marker.Pattern.Match(text)
	}

	return found
}

// normalizeText collapses whitespace and apostrophe variants so phrases match regardless of layout
func normalizeText(text []byte) []byte {
	return []byte(curlyApostrophes.Replace(string(whitespacePattern.ReplaceAll(text, []byte(" ")))))
}
