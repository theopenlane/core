package soc2

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/pkg/mapx"
)

// findingDigestLength is how many hex characters of the description hash end a finding's external id
const findingDigestLength = 8

// AssignFindingExternalIDs gives each finding a stable external id it is accepted and imported by: a finding
// attributed to a review is keyed under that review, any other under the report and its first ref code, both
// ending in a hash of the description so the id holds when the model returns findings in a different order
func AssignFindingExternalIDs(findings []map[string]any, reportFileID string) {
	for _, finding := range findings {
		finding[entityops.FindingFields.ExternalID.Name] = findingExternalID(finding, reportFileID)
	}
}

// findingExternalID builds one finding's external id
func findingExternalID(finding map[string]any, reportFileID string) string {
	description := mapx.Value[string](finding, entityops.FindingFields.Description.Name)
	sum := sha256.Sum256([]byte(description))
	digest := hex.EncodeToString(sum[:])[:findingDigestLength]

	if reviewID := mapx.Value[string](finding, ReviewExternalIDKey); reviewID != "" {
		return reviewID + "-finding-" + digest
	}

	refCode := ""
	if refCodes := refCodesOf(finding); len(refCodes) > 0 {
		refCode = refCodes[0]
	}

	return "soc2-report-" + reportFileID + "-" + refCode + "-finding-" + digest
}

// reviewIndex is what a finding is matched against: every review addressed by the test it records,
// and every review grouped under its control for when the control has only one
type reviewIndex struct {
	byTest    map[string]string
	byControl map[string][]string
}

// LinkFindings points each finding at the review whose test produced it, reporting how many were
// attributed and how many were not; a report states its exceptions against a specific test, so an
// unlinked finding means the attribution failed rather than that none exists
func LinkFindings(reviews, findings []map[string]any) (linked int, unlinked int) {
	index := indexReviews(reviews)

	for _, finding := range findings {
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
func indexReviews(reviews []map[string]any) reviewIndex {
	index := reviewIndex{byTest: map[string]string{}, byControl: map[string][]string{}}

	for _, review := range reviews {
		externalID := mapx.Value[string](review, entityops.ReviewFields.ExternalID.Name)
		if externalID == "" {
			continue
		}

		controlCode, _ := splitRefCodes(refCodesOf(review))
		if controlCode == "" {
			continue
		}

		index.byControl[controlCode] = append(index.byControl[controlCode], externalID)

		details := mapx.Value[string](review, entityops.ReviewFields.Details.Name)

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

	testDetails := mapx.Value[string](finding, testDetailsKey)

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
	return mapx.Strings(item, refCodesKey)
}
