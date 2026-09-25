package soc2

import (
	"slices"
	"strings"
)

const (
	// ReviewExternalIDKey is the field on a finding naming the review whose test produced it
	ReviewExternalIDKey = "reviewExternalID"
	// minMatchWords is how many distinctive words a finding and a test must share before the
	// finding is attributed to that test, when its control has more than one
	minMatchWords = 3
	// minWordLength is the shortest word considered distinctive enough to match on
	minWordLength = 4
)

// commonWords appear in nearly every test description, so they carry no signal when choosing
// between the tests of one control
var commonWords = []string{
	"determine", "further", "inquired", "inspected", "observed", "period", "regarding",
	"related", "sample", "selected", "specified", "that", "the", "these", "this", "were", "with",
}

// candidateReview is the part of a review needed to attribute a finding to it
type candidateReview struct {
	externalID string
	details    string
}

// LinkFindings points each finding at the review whose test produced it, reporting how many were
// attributed and how many were not; a report states its exceptions against a specific test, so an
// unlinked finding means the attribution failed rather than that none exists
func LinkFindings(reviews, findings []any) (linked int, unlinked int) {
	byControl := reviewsByControl(reviews)

	for _, item := range findings {
		finding, ok := item.(map[string]any)
		if !ok {
			continue
		}

		externalID, found := matchReview(byControl, finding)
		if !found {
			unlinked++

			continue
		}

		finding[ReviewExternalIDKey] = externalID
		linked++
	}

	return linked, unlinked
}

// reviewsByControl groups reviews under the control the report printed them beneath
func reviewsByControl(reviews []any) map[string][]candidateReview {
	grouped := map[string][]candidateReview{}

	for _, item := range reviews {
		review, ok := item.(map[string]any)
		if !ok {
			continue
		}

		externalID, _ := review["externalID"].(string)
		if externalID == "" {
			continue
		}

		controlCode, _ := splitRefCodes(refCodesOf(review))
		if controlCode == "" {
			continue
		}

		details, _ := review["details"].(string)

		grouped[controlCode] = append(grouped[controlCode], candidateReview{externalID: externalID, details: details})
	}

	return grouped
}

// matchReview picks the review a finding belongs to; a control usually has a single test, so the
// control alone decides it, and only a control with several tests needs the finding's text
func matchReview(byControl map[string][]candidateReview, finding map[string]any) (string, bool) {
	candidates := candidatesFor(byControl, refCodesOf(finding))

	switch len(candidates) {
	case 0:
		return "", false
	case 1:
		return candidates[0].externalID, true
	}

	description, _ := finding["description"].(string)

	return bestMatch(candidates, description)
}

// candidatesFor collects the reviews of every control the finding names, so a finding is matched
// whichever order its ref codes are listed in
func candidatesFor(byControl map[string][]candidateReview, refCodes []string) []candidateReview {
	candidates := make([]candidateReview, 0, len(refCodes))
	seen := map[string]bool{}

	for _, refCode := range refCodes {
		for _, candidate := range byControl[refCode] {
			if seen[candidate.externalID] {
				continue
			}

			seen[candidate.externalID] = true

			candidates = append(candidates, candidate)
		}
	}

	return candidates
}

// bestMatch picks the test sharing the most distinctive words with the finding, reporting false
// when nothing shares enough to be trusted
func bestMatch(candidates []candidateReview, description string) (string, bool) {
	words := distinctiveWords(description)
	if len(words) == 0 {
		return "", false
	}

	best := ""
	bestScore := 0

	for _, candidate := range candidates {
		score := sharedWordCount(words, distinctiveWords(candidate.details))
		if score > bestScore {
			best, bestScore = candidate.externalID, score
		}
	}

	if bestScore < minMatchWords {
		return "", false
	}

	return best, true
}

// distinctiveWords reduces text to the set of words worth comparing
func distinctiveWords(text string) map[string]bool {
	words := map[string]bool{}

	for _, word := range strings.FieldsFunc(strings.ToLower(text), isWordBreak) {
		if len(word) < minWordLength || slices.Contains(commonWords, word) {
			continue
		}

		words[word] = true
	}

	return words
}

// sharedWordCount counts the words the two sets have in common
func sharedWordCount(first, second map[string]bool) int {
	shared := 0

	for word := range first {
		if second[word] {
			shared++
		}
	}

	return shared
}

// isWordBreak reports whether a rune separates words
func isWordBreak(r rune) bool {
	return (r < 'a' || r > 'z') && (r < '0' || r > '9')
}

// refCodesOf reads an item's ref codes out of a decoded metadata map
func refCodesOf(item map[string]any) []string {
	raw, _ := item["refCodes"].([]any)

	refCodes := make([]string, 0, len(raw))

	for _, value := range raw {
		if refCode, ok := value.(string); ok && refCode != "" {
			refCodes = append(refCodes, refCode)
		}
	}

	return refCodes
}
