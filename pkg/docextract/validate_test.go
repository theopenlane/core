package docextract

import (
	"regexp"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/pdftext/pdftest"
)

const testKind = "test_memo"

func init() {
	RegisterProfile(testKind, Profile{
		Kind:                "test memo",
		MinPages:            2,
		MinStrong:           1,
		HighConfidence:      2,
		MissingSectionsHint: "a subject or a signature",
		Markers: []Marker{
			{Name: "memo", Weight: MarkerRequired, Pattern: regexp.MustCompile(`(?i)\bmemo\b`)},
			{Name: "subject", Weight: MarkerStrong, Pattern: regexp.MustCompile(`(?i)subject:`)},
			{Name: "signature", Weight: MarkerStrong, Pattern: regexp.MustCompile(`(?i)signed,`)},
			{Name: "date", Weight: MarkerSupporting, Pattern: regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)},
		},
		Subtype: func(found map[string]bool) ReportType {
			if found["signature"] {
				return "signed"
			}

			return ReportTypeUnknown
		},
	})
}

func memoLines() []string {
	return []string{
		"MEMO to all staff regarding the office cat policy and related feeding schedules",
		"Subject: Office cat feeding schedule and treat allocation for the coming quarter",
		"Dated 2026-09-21 and effective immediately for all floors and remote offices",
		"Signed, the Facilities Team on behalf of the cat welfare committee and management",
	}
}

func TestValidateAccepted(t *testing.T) {
	result, err := Validate(pdftest.Repeat(memoLines(), 3), testKind)

	assert.NilError(t, err)
	assert.Check(t, result.Kind == testKind)
	assert.Check(t, result.Pages == 3)
	assert.Check(t, result.Confidence == ConfidenceHigh)
	assert.Check(t, result.ReportType == ReportType("signed"))
	assert.DeepEqual(t, result.Strong, []string{"subject", "signature"})
	assert.DeepEqual(t, result.Supporting, []string{"date"})
	assert.Check(t, result.TextLength > MinExtractedText)
}

func TestValidateAcceptedLowConfidence(t *testing.T) {
	lines := []string{memoLines()[0], memoLines()[1], memoLines()[2]}

	result, err := Validate(pdftest.Repeat(lines, 3), testKind)

	assert.NilError(t, err)
	assert.Check(t, result.Confidence == ConfidenceAccepted)
	assert.Check(t, result.ReportType == ReportTypeUnknown)
}

func TestValidateTooShort(t *testing.T) {
	result, err := Validate(pdftest.Repeat(memoLines(), 1), testKind)

	assert.ErrorIs(t, err, ErrTooShort)
	assert.Check(t, result.Pages == 1)
	assert.Check(t, strings.Contains(ValidationReason(err), "fewer than 2 pages"))
	assert.Check(t, strings.Contains(ValidationReason(err), "test memo"))
}

func TestValidateNotMatched(t *testing.T) {
	lines := []string{
		"A newsletter about the quarterly picnic, the raffle winners, and the new parking rules",
		"Subject: Picnic logistics including food, games, and the shuttle timetable for attendees",
		"Signed, the Social Committee with thanks to everyone who volunteered their weekend",
	}

	_, err := Validate(pdftest.Repeat(lines, 3), testKind)

	assert.ErrorIs(t, err, ErrNotMatched)
	assert.Check(t, strings.Contains(ValidationReason(err), "never references memo"))
}

func TestValidateMissingSections(t *testing.T) {
	lines := []string{
		"MEMO about nothing in particular, circulated widely to keep everyone informed of little",
		"There is no subject line and nobody signed it, which is exactly the point of this test case",
	}

	result, err := Validate(pdftest.Repeat(lines, 3), testKind)

	assert.ErrorIs(t, err, ErrMissingSections)
	assert.Check(t, len(result.Strong) == 0)
	assert.Check(t, strings.Contains(ValidationReason(err), "such as a subject or a signature"))
}

func TestValidateNoText(t *testing.T) {
	result, err := Validate(pdftest.Repeat([]string{"x"}, 3), testKind)

	assert.ErrorIs(t, err, ErrNoText)
	assert.Check(t, result.TextLength < MinExtractedText)
}

func TestValidateUnreadable(t *testing.T) {
	_, err := Validate([]byte("not a pdf"), testKind)

	assert.ErrorIs(t, err, ErrUnreadable)
	assert.Check(t, ValidationReason(err) == ErrUnreadable.Error())
}

func TestValidateNoProfile(t *testing.T) {
	_, err := Validate(pdftest.Repeat(memoLines(), 3), "unknown")

	assert.ErrorIs(t, err, ErrNoProfile)
}

func TestValidationReasonPassesThroughPlainErrors(t *testing.T) {
	assert.Check(t, ValidationReason(ErrNoProfile) == ErrNoProfile.Error())
}

func TestNormalizeText(t *testing.T) {
	normalized := normalizeText([]byte("Management’s   assertion\n\tsplit  across\nlines"))

	assert.Check(t, string(normalized) == "Management's assertion split across lines")
}
