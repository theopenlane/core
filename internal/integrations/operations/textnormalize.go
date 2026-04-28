package operations

import (
	"regexp"
	"strings"
)

// multiSpaceRe matches 3 or more consecutive spaces used as paragraph separators
// in descriptions that lack proper newlines (e.g. from certain CVE feeds)
var multiSpaceRe = regexp.MustCompile(` {3,}`)

// normalizeDescription converts descriptions that use runs of spaces as paragraph
// separators into proper markdown paragraphs. Descriptions that already contain
// newlines are returned unchanged.
func normalizeDescription(desc string) string {
	if strings.Contains(desc, "\n") {
		return desc
	}

	return multiSpaceRe.ReplaceAllString(strings.TrimSpace(desc), "\n\n")
}
