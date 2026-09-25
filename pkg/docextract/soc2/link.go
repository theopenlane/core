package soc2

const (
	// ReviewExternalIDKey is the field on a finding naming the review whose test produced it
	ReviewExternalIDKey = "reviewExternalID"
	// testDetailsKey is the field on a finding carrying the test procedure printed in its table row
	testDetailsKey = "testDetails"
)

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

// matchReview picks the review a finding belongs to; the exception and the test it was reported
// against are printed in one table row, so the test text carried on the finding identifies the
// review outright, and a control with a single test needs no text at all. Anything else is left
// unlinked and reported, because attributing a finding to the wrong test is worse than not doing it
func matchReview(byControl map[string][]candidateReview, finding map[string]any) (string, bool) {
	candidates := candidatesFor(byControl, refCodesOf(finding))

	if len(candidates) == 0 {
		return "", false
	}

	testDetails, _ := finding[testDetailsKey].(string)
	if externalID, found := exactMatch(candidates, testDetails); found {
		return externalID, true
	}

	if len(candidates) == 1 {
		return candidates[0].externalID, true
	}

	return "", false
}

// exactMatch finds the one review whose test text is the text the finding was reported against,
// ignoring the case and wrapping a second pass over the same table can change
func exactMatch(candidates []candidateReview, testDetails string) (string, bool) {
	wanted := normalizeTestText(testDetails)
	if wanted == "" {
		return "", false
	}

	matches := make([]string, 0, 1)

	for _, candidate := range candidates {
		if normalizeTestText(candidate.details) == wanted {
			matches = append(matches, candidate.externalID)
		}
	}

	if len(matches) != 1 {
		return "", false
	}

	return matches[0], true
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
