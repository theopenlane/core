package soc2

import (
	"regexp"
	"strings"
)

// whitespacePattern collapses runs of whitespace when comparing test text
var whitespacePattern = regexp.MustCompile(`\s+`)

// testKey identifies the test a table row records, the control it was printed under paired with
// its verbatim text; it is what the review merge collapses restatements by and what a finding is
// attributed by, so both agree on when two rows describe the same test. It reports false when
// there is no text to identify the row with
func testKey(controlCode, details string) (string, bool) {
	normalized := normalizeTestText(details)
	if normalized == "" || controlCode == "" {
		return "", false
	}

	return controlCode + "\x00" + normalized, true
}

// normalizeTestText lowers and collapses whitespace so the same text reprinted with different
// wrapping still matches
func normalizeTestText(details string) string {
	return whitespacePattern.ReplaceAllString(strings.ToLower(strings.TrimSpace(details)), " ")
}
