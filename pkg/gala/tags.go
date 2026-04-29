package gala

import "regexp"

var invalidTagChars = regexp.MustCompile(`[^\w\-]`)

const maxTagLen = 255

// SanitizeTag returns a River-compatible tag: non-word/hyphen chars replaced with `_`,
// leading and trailing hyphens replaced with `_`, truncated to 255 characters
func SanitizeTag(s string) string {
	s = invalidTagChars.ReplaceAllString(s, "_")

	if len(s) > maxTagLen {
		s = s[:maxTagLen]
	}

	// river requires first and last chars to be word chars, not `-`
	if len(s) > 0 && s[0] == '-' {
		s = "_" + s[1:]
	}

	if len(s) > 0 && s[len(s)-1] == '-' {
		s = s[:len(s)-1] + "_"
	}

	return s
}
