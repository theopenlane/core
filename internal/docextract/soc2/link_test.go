package soc2

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func review(externalID, details string, refCodes ...string) map[string]any {
	codes := make([]any, 0, len(refCodes))
	for _, refCode := range refCodes {
		codes = append(codes, refCode)
	}

	return map[string]any{"external_id": externalID, "details": details, "ref_codes": codes}
}

func findingWithTest(description, testDetails string, refCodes ...string) map[string]any {
	item := finding(description, refCodes...)
	item["test_details"] = testDetails

	return item
}

func finding(description string, refCodes ...string) map[string]any {
	codes := make([]any, 0, len(refCodes))
	for _, refCode := range refCodes {
		codes = append(codes, refCode)
	}

	return map[string]any{"description": description, "ref_codes": codes}
}

func TestLinkFindingsUsesTheOnlyReviewForAControl(t *testing.T) {
	reviews := []any{review("soc2-CC 5.2-14-01", "Inspected the pull requests related to a sample of emergency changes", "CC 5.2-14", "CC 5.2")}
	target := finding("Two of 25 emergency changes were not approved within two business days", "CC 5.2-14")

	linked, unlinked := LinkFindings(reviews, []any{target})

	assert.Check(t, linked == 1)
	assert.Check(t, unlinked == 0)
	assert.Check(t, target[ReviewExternalIDKey] == "soc2-CC 5.2-14-01")
}

func TestLinkFindingsLeavesSeveralTestsUnlinkedWithoutRowText(t *testing.T) {
	reviews := []any{
		review("soc2-CC 6.1-04-01", "Inspected the termination tickets related to a sample of terminated employees", "CC 6.1-04"),
		review("soc2-CC 6.1-04-02", "Observed the password parameters configured for minimum length and complexity", "CC 6.1-04"),
	}
	target := finding("One terminated employee retained access because the termination ticket was not raised", "CC 6.1-04")

	linked, unlinked := LinkFindings(reviews, []any{target})

	assert.Check(t, linked == 0)
	assert.Check(t, unlinked == 1)

	_, set := target[ReviewExternalIDKey]
	assert.Check(t, !set)
}

func TestLinkFindingsIgnoresRefCodeOrder(t *testing.T) {
	reviews := []any{review("soc2-CC 1.1-01-01", "Inspected the written job descriptions", "CC 1.1-01", "CC1.1")}
	target := finding("Job descriptions were not documented for two roles", "CC1.1", "CC 1.1-01")

	linked, _ := LinkFindings(reviews, []any{target})

	assert.Check(t, linked == 1)
	assert.Check(t, target[ReviewExternalIDKey] == "soc2-CC 1.1-01-01")
}

func TestLinkFindingsReportsUnmatchedControl(t *testing.T) {
	reviews := []any{review("soc2-CC 5.2-14-01", "Inspected the pull requests", "CC 5.2-14")}
	target := finding("An unrelated exception", "CC 9.9-99")

	linked, unlinked := LinkFindings(reviews, []any{target})

	assert.Check(t, linked == 0)
	assert.Check(t, unlinked == 1)
	_, set := target[ReviewExternalIDKey]
	assert.Check(t, !set)
}

func TestLinkFindingsPrefersTheTestFromTheSameRow(t *testing.T) {
	reviews := []any{
		review("soc2-CC 6.1-04-01", "Inspected the termination tickets related to a sample of terminated employees", "CC 6.1-04"),
		review("soc2-CC 6.1-04-02", "Observed the password parameters configured for minimum length and complexity", "CC 6.1-04"),
	}
	target := findingWithTest(
		"One terminated employee retained access because the termination ticket was not raised",
		"Observed the  password parameters\nconfigured for MINIMUM length and complexity",
		"CC 6.1-04",
	)

	linked, unlinked := LinkFindings(reviews, []any{target})

	assert.Check(t, linked == 1)
	assert.Check(t, unlinked == 0)
	assert.Check(t, target[ReviewExternalIDKey] == "soc2-CC 6.1-04-02")
}

func TestLinkFindingsLeavesUnlinkedWhenRowTextDoesNotMatch(t *testing.T) {
	reviews := []any{
		review("soc2-CC 6.1-04-01", "Inspected the termination tickets related to a sample of terminated employees", "CC 6.1-04"),
		review("soc2-CC 6.1-04-02", "Observed the password parameters configured for minimum length", "CC 6.1-04"),
	}
	target := findingWithTest(
		"One terminated employee retained access because the termination ticket was not raised",
		"a paraphrase the model invented",
		"CC 6.1-04",
	)

	linked, unlinked := LinkFindings(reviews, []any{target})

	assert.Check(t, linked == 0)
	assert.Check(t, unlinked == 1)
}

func TestAssignFindingExternalIDs(t *testing.T) {
	linked := finding("Two of 25 emergency changes were not approved", "CC 5.2-14")
	linked[ReviewExternalIDKey] = "soc2-report-2025-CC 5.2-14-01"
	unlinked := finding("One terminated employee retained access", "CC 6.1-04", "CC6.1")
	reordered := finding("Two of 25 emergency changes were not approved", "CC 5.2-14")
	reordered[ReviewExternalIDKey] = "soc2-report-2025-CC 5.2-14-01"

	AssignFindingExternalIDs([]any{linked, unlinked}, "file1")
	AssignFindingExternalIDs([]any{reordered}, "file1")

	linkedID, _ := linked["external_id"].(string)
	unlinkedID, _ := unlinked["external_id"].(string)

	assert.Check(t, strings.HasPrefix(linkedID, "soc2-report-2025-CC 5.2-14-01-finding-"))
	assert.Check(t, strings.HasPrefix(unlinkedID, "soc2-report-file1-CC 6.1-04-finding-"))
	assert.Check(t, linkedID == reordered["external_id"])
	assert.Check(t, linkedID != unlinkedID)
}
