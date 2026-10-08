package soc2

const (
	// ReviewExternalIDKey is the field on a finding naming the review whose test produced it
	ReviewExternalIDKey = "reviewExternalID"
	// testDetailsKey is the field on a finding carrying the test procedure printed in its table row
	testDetailsKey = "testDetails"
)

// reviewIndex is what a finding is matched against: every review addressed by the test it records,
// and every review grouped under its control for when the control has only one
type reviewIndex struct {
	byTest    map[string]string
	byControl map[string][]string
}

// LinkFindings points each finding at the review whose test produced it, reporting how many were
// attributed and how many were not; a report states its exceptions against a specific test, so an
// unlinked finding means the attribution failed rather than that none exists
func LinkFindings(reviews, findings []any) (linked int, unlinked int) {
	index := indexReviews(reviews)

	for _, item := range findings {
		finding, ok := item.(map[string]any)
		if !ok {
			continue
		}

		externalID, found := matchReview(index, finding)
		if !found {
			unlinked++

			continue
		}

		finding[ReviewExternalIDKey] = externalID
		linked++
	}

	return linked, unlinked
}

// indexReviews builds the lookups a finding is matched against; a test key shared by two reviews
// is recorded as ambiguous so no finding is attributed to an arbitrary one of them
func indexReviews(reviews []any) reviewIndex {
	index := reviewIndex{byTest: map[string]string{}, byControl: map[string][]string{}}

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

		index.byControl[controlCode] = append(index.byControl[controlCode], externalID)

		details, _ := review["details"].(string)

		key, ok := testKey(controlCode, details)
		if !ok {
			continue
		}

		if _, clash := index.byTest[key]; clash {
			index.byTest[key] = ""

			continue
		}

		index.byTest[key] = externalID
	}

	return index
}

// matchReview picks the review a finding belongs to; the exception and the test it was reported
// against are printed in one table row, so the test text carried on the finding identifies the
// review outright, and a control with a single test needs no text at all. Anything else is left
// unlinked and reported, because attributing a finding to the wrong test is worse than not doing it
func matchReview(index reviewIndex, finding map[string]any) (string, bool) {
	refCodes := refCodesOf(finding)

	testDetails, _ := finding[testDetailsKey].(string)

	// the finding names its control among its ref codes, in no guaranteed order
	for _, refCode := range refCodes {
		key, ok := testKey(refCode, testDetails)
		if !ok {
			continue
		}

		if externalID := index.byTest[key]; externalID != "" {
			return externalID, true
		}
	}

	candidates := candidatesFor(index.byControl, refCodes)
	if len(candidates) == 1 {
		return candidates[0], true
	}

	return "", false
}

// candidatesFor collects the reviews of every control the finding names
func candidatesFor(byControl map[string][]string, refCodes []string) []string {
	candidates := make([]string, 0, len(refCodes))
	seen := map[string]bool{}

	for _, refCode := range refCodes {
		for _, externalID := range byControl[refCode] {
			if seen[externalID] {
				continue
			}

			seen[externalID] = true

			candidates = append(candidates, externalID)
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
