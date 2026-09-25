package soc2

import (
	"fmt"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/pdftext/pdftest"
)

func TestValidateTypeTwo(t *testing.T) {
	result, err := docextract.Validate(pdftest.Build(demoTypeTwoPages(), false), KindName)

	assert.NilError(t, err)
	assert.Check(t, result.Kind == KindName)
	assert.Check(t, result.Pages == 6)
	assert.Check(t, result.Confidence == docextract.ConfidenceHigh)
	assert.Check(t, result.ReportType == ReportTypeTwo)
	assert.Check(t, len(result.Strong) >= HighConfidenceMarkers)
	assert.Check(t, contains(result.Strong, "auditor_report"))
	assert.Check(t, contains(result.Strong, "management_assertion"))
	assert.Check(t, contains(result.Strong, "criteria_ids"))
	assert.Check(t, contains(result.Strong, "type2_testing"))
	assert.Check(t, contains(result.Supporting, "cuecs"))
}

func TestValidateTypeOne(t *testing.T) {
	result, err := docextract.Validate(pdftest.Build(demoTypeOnePages(), false), KindName)

	assert.NilError(t, err)
	assert.Check(t, result.Confidence == docextract.ConfidenceHigh)
	assert.Check(t, result.ReportType == ReportTypeOne)
	assert.Check(t, !contains(result.Strong, "type2_testing"))
	assert.Check(t, contains(result.Supporting, "as_of_date"))
}

func TestValidateTooShort(t *testing.T) {
	result, err := docextract.Validate(pdftest.Build(demoTypeTwoPages()[:2], false), KindName)

	assert.ErrorIs(t, err, docextract.ErrTooShort)
	assert.Check(t, result.Pages == 2)
	assert.Check(t, strings.Contains(docextract.ValidationReason(err), "fewer than 5 pages"))
}

func TestValidateInvoiceIsNotSOC2(t *testing.T) {
	_, err := docextract.Validate(demoInvoice(), KindName)

	assert.ErrorIs(t, err, docextract.ErrNotMatched)
	assert.Check(t, strings.Contains(docextract.ValidationReason(err), "never references SOC 2"))
}

func TestValidateBrochureMissingSections(t *testing.T) {
	result, err := docextract.Validate(demoMarketingBrochure(), KindName)

	assert.ErrorIs(t, err, docextract.ErrMissingSections)
	assert.Check(t, len(result.Strong) < MinStrongMarkers)
}

func TestValidateCompositeFont(t *testing.T) {
	result, err := docextract.Validate(pdftest.Build(demoTypeTwoPages(), true), KindName)

	assert.NilError(t, err)
	assert.Check(t, result.Confidence == docextract.ConfidenceHigh)
	assert.Check(t, result.ReportType == ReportTypeTwo)
	assert.Check(t, result.TextLength > docextract.MinExtractedText)
}

func TestSubtypeUnknownWithoutDates(t *testing.T) {
	assert.Check(t, subtype(map[string]bool{"type2_testing": true}) == docextract.ReportTypeUnknown)
	assert.Check(t, subtype(map[string]bool{"as_of_date": true, "type2_testing": true}) == docextract.ReportTypeUnknown)
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}

	return false
}

func demoTypeTwoPages() [][]string {
	return [][]string{
		{
			"DEMO ONLY - NOT A REAL REPORT",
			"Meow Cloud Services Inc.",
			"SOC 2 Type II Report",
			"Report on the Description of the Meow Platform System",
			"Relevant to Security, Availability, and Confidentiality",
			"For the period January 1, 2026 through June 30, 2026",
			"Prepared by Demo Auditors LLP, Certified Public Accountants",
		},
		{
			"Section 1 - Independent Service Auditor's Report",
			"To the management of Meow Cloud Services Inc.",
			"We have examined Meow Cloud Services Inc.'s description of its Meow Platform system",
			"throughout the period January 1, 2026 to June 30, 2026 based on the criteria for a description",
			"of a service organization's system set forth in the AICPA Description Criteria and the",
			"suitability of the design and operating effectiveness of controls stated in the description",
			"to meet the applicable Trust Services Criteria for Security, Availability, and Confidentiality.",
		},
		{
			"Section 2 - Management's Assertion",
			"We have prepared the description of the Meow Platform system based on the AICPA criteria.",
			"The controls stated in the description were suitably designed and operated effectively",
			"throughout the period January 1, 2026 to June 30, 2026.",
			"Demo Signatory, Chief Demo Officer",
		},
		{
			"Section 3 - Description of the System",
			"Meow Cloud Services Inc. provides a fictional cat scheduling platform to demo customers.",
			"Complementary User Entity Controls",
			"User entities are responsible for managing their own user access to the platform.",
			"Subservice Organizations",
			"The company uses Demo Hosting Co as a subservice organization for data center services.",
		},
		{
			"Section 4 - Trust Services Criteria, Controls, and Tests of Controls",
			"CC1.1 The entity demonstrates a commitment to integrity and ethical values.",
			"ACF-01 A code of conduct is acknowledged annually by all demo employees.",
			"Tests of Controls: Inspected acknowledgement records for a sample of demo employees.",
			"Results of Tests: No exceptions noted.",
			"CC6.1 The entity implements logical access security software and architectures.",
			"ACF-12 Access to the demo production environment requires multi-factor authentication.",
			"Tests of Controls: Inspected the MFA configuration for the demo environment.",
			"Results of Tests: Exceptions noted. One of 25 sampled demo accounts lacked MFA.",
		},
		{
			"Section 4 continued",
			"A1.1 The entity maintains capacity demand and monitors current usage.",
			"ACF-20 Demo capacity dashboards are reviewed weekly.",
			"Tests of Controls: Inspected weekly review evidence for a sample of weeks.",
			"Results of Tests: No exceptions noted.",
			"Section 5 - Other Information Provided by Meow Cloud Services Inc.",
			"Management's response to the exception noted above is included for demo purposes only.",
		},
	}
}

func demoTypeOnePages() [][]string {
	return [][]string{
		{
			"DEMO ONLY - NOT A REAL REPORT",
			"Purr Analytics Ltd.",
			"SOC 2 Type I Report",
			"Report on the Description of the Purr Analytics System and the Suitability of the Design of Controls",
			"Relevant to Security",
			"As of March 31, 2026",
			"Prepared by Demo Auditors LLP, CPAs",
		},
		{
			"Section 1 - Independent Service Auditor's Report",
			"We have examined Purr Analytics Ltd.'s description of its system as of March 31, 2026",
			"and the suitability of the design of controls to meet the applicable Trust Services Criteria",
			"established by the AICPA for Security.",
		},
		{
			"Section 2 - Assertion of Management",
			"The description fairly presents the Purr Analytics system as of March 31, 2026.",
			"The controls stated in the description were suitably designed as of that date.",
		},
		{
			"Section 3 - System Description",
			"Purr Analytics Ltd. provides a fictional feline sentiment analysis service to demo customers.",
			"Complementary User Entity Controls are described for demo user entities.",
		},
		{
			"Section 4 - Applicable Trust Services Criteria and Related Controls",
			"CC1.1 The entity demonstrates a commitment to integrity and ethical values.",
			"CC2.1 The entity obtains or generates relevant quality information.",
			"CC6.1 The entity implements logical access security software.",
			"CC7.2 The entity monitors system components for anomalies.",
		},
	}
}

func demoInvoice() []byte {
	return pdftest.Repeat([]string{
		"DEMO ONLY - NOT A REAL INVOICE",
		"Whiskers Catering Co.",
		fmt.Sprintf("Invoice DEMO-%04d", 1),
		"Bill to: Demo Customer, 123 Example Street",
		"Line item: Tuna platter, quantity 12, unit price 4.50",
		"Line item: Catnip garnish, quantity 3, unit price 1.25",
		"Subtotal 57.75, tax 4.62, total 62.37",
		"Payment due within 30 days. Thank you for your demo business.",
	}, 6)
}

func demoMarketingBrochure() []byte {
	return pdftest.Repeat([]string{
		"DEMO ONLY - NOT A REAL BROCHURE",
		"Meow Cloud Services Inc. - Trust and Security Overview",
		"We are SOC 2 compliant and take the security of your demo data seriously.",
		"Our platform offers availability guarantees and confidentiality for every demo customer.",
		"Request a copy of our latest audit report from your demo account manager.",
	}, 6)
}

func TestSectionTooThinIgnoresSectionsWithoutAFloor(t *testing.T) {
	_, thin := SectionTooThin("contacts", 0, 0)

	assert.Check(t, !thin)
}

func TestSectionTooThinFlagsAnUnscopedSectionBelowTheFloor(t *testing.T) {
	minimum, thin := SectionTooThin(ControlsPart, 2, 0)

	assert.Check(t, thin)
	assert.Check(t, minimum == MinSectionItems)
}

func TestSectionTooThinAcceptsTheFloor(t *testing.T) {
	_, thin := SectionTooThin(ControlsPart, MinSectionItems, 0)

	assert.Check(t, !thin)
}

func TestSectionTooThinCapsTheFloorAtTheScopeSize(t *testing.T) {
	minimum, thin := SectionTooThin(ReviewsPart, 2, 2)

	assert.Check(t, !thin)
	assert.Check(t, minimum == 2)
}

func TestSectionTooThinFlagsAScopedSectionBelowItsScope(t *testing.T) {
	minimum, thin := SectionTooThin(ReviewsPart, 1, 3)

	assert.Check(t, thin)
	assert.Check(t, minimum == 3)
}
