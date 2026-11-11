package urlx

import (
	"regexp"
	"strings"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/utils/keygen"
)

const (
	// slugSeparator joins slug words and the collision suffix
	slugSeparator = "-"
	// slugSuffixLength is the number of random characters appended to a slug that is already taken
	slugSuffixLength = 4
	// slugAttempts bounds the candidates UniqueSlug tries, the bare base first and then suffixed ones
	slugAttempts = 6
)

// slugDisallowed matches every run of characters a slug may not contain, including existing
// hyphens so that runs collapse to one separator
var slugDisallowed = regexp.MustCompile(`[^a-z0-9]+`)

// Slug derives a URL path segment from a human-readable name: strcase.KebabCase output reduced
// to ASCII a-z, 0-9 and single hyphens with no leading or trailing hyphen
func Slug(name string) string {
	return strings.Trim(slugDisallowed.ReplaceAllString(strcase.KebabCase(name), slugSeparator), slugSeparator)
}

// UniqueSlug returns base when taken reports it free, otherwise base joined with a short random
// lowercase suffix, trying a bounded number of fresh suffixes before returning ErrSlugExhausted
func UniqueSlug(base string, taken func(string) (bool, error)) (string, error) {
	for attempt := range slugAttempts {
		candidate := base
		if attempt > 0 || base == "" {
			candidate = strings.TrimPrefix(base+slugSeparator+slugSuffix(), slugSeparator)
		}

		exists, err := taken(candidate)
		if err != nil {
			return "", err
		}

		if !exists {
			return candidate, nil
		}
	}

	return "", ErrSlugExhausted
}

// slugSuffix returns the random lowercase alphanumeric collision suffix
func slugSuffix() string {
	return strings.ToLower(keygen.AlphaNumeric(slugSuffixLength))
}
