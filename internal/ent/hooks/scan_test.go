package hooks

import (
	"context"
	"strings"
	"testing"

	"entgo.io/ent"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/docextract/soc2"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
	"github.com/theopenlane/core/v2/pkg/docextract"
)

func TestExceedsSingleReportFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		existing int
		removed  int
		attached int
		cleared  bool
		want     bool
	}{
		{name: "first report attached", attached: 1},
		{name: "two reports attached at once", attached: 2, want: true},
		{name: "existing report left alone", existing: 1},
		{name: "second report added to existing", existing: 1, attached: 1, want: true},
		{name: "report swapped by remove then add", existing: 1, removed: 1, attached: 1},
		{name: "report swapped by clear then add", existing: 1, attached: 1, cleared: true},
		{name: "all reports cleared", existing: 2, cleared: true},
		{name: "two reports added after clear", existing: 1, attached: 2, cleared: true, want: true},
		{name: "remove more than exists", existing: 1, removed: 2, attached: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, exceedsSingleReportFile(tt.existing, tt.removed, tt.attached, tt.cleared), tt.want)
		})
	}
}

func TestCheckReportScanFilesOnCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		reportScan  bool
		uploaded    []string
		builderIDs  []string
		expectedErr error
	}{
		{name: "one uploaded report", reportScan: true, uploaded: []string{"file-1"}},
		{name: "two uploaded reports", reportScan: true, uploaded: []string{"file-1", "file-2"}, expectedErr: ErrReportScanSingleFile},
		{name: "upload plus builder file id", reportScan: true, uploaded: []string{"file-1"}, builderIDs: []string{"file-2"}, expectedErr: ErrReportScanSingleFile},
		{name: "same file id in both sources", reportScan: true, uploaded: []string{"file-1"}, builderIDs: []string{"file-1"}},
		{name: "two builder file ids", reportScan: true, builderIDs: []string{"file-1", "file-2"}, expectedErr: ErrReportScanSingleFile},
		{name: "no files attached", reportScan: true},
		{name: "domain scan with many files", uploaded: []string{"file-1", "file-2", "file-3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &generated.ScanMutation{}
			m.SetOp(ent.OpCreate)
			m.AddFileIDs(tt.builderIDs...)

			if tt.reportScan {
				m.SetScanType(enums.ScanTypeReport)
				m.SetPerformedBy(gemini.PerformedBy)
			} else {
				m.SetScanType(enums.ScanTypeDomain)
			}

			err := checkReportScanFiles(context.Background(), m, tt.uploaded)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)

				return
			}

			assert.NilError(t, err)
		})
	}
}

func TestCheckReportScanFilesUpdateWithoutFileChange(t *testing.T) {
	t.Parallel()

	m := &generated.ScanMutation{}
	m.SetOp(ent.OpUpdateOne)
	m.SetStatus(enums.ScanStatusProcessing)

	assert.NilError(t, checkReportScanFiles(context.Background(), m, nil))
}

func TestNormalizedDomains(t *testing.T) {
	assert.DeepEqual(t, normalizedDomains([]string{" Example.io ", "example.io", "", "Docs.Example.io"}), []string{"example.io", "docs.example.io"})
	assert.Check(t, len(normalizedDomains(nil)) == 0)
}

func TestReportScanRowMatchPerformers(t *testing.T) {
	vendorOnly := reportScanRowMatchForKind(soc2.KindName, gemini.VendorReviewPerformedBy)
	both := reportScanRowMatch(gemini.ReportPerformers...)

	assert.Check(t, len(vendorOnly) == 3)
	assert.DeepEqual(t, vendorOnly[1].In, []string{gemini.VendorReviewPerformedBy})
	assert.DeepEqual(t, both[1].In, []string{gemini.PerformedBy, gemini.VendorReviewPerformedBy})
}

func TestVendorReviewTitle(t *testing.T) {
	assert.Check(t, vendorReviewTitle(soc2.ReportTypeOne, "2026-07-28") == "SOC 2 Type 1 report review, period ending 2026-07-28")
	assert.Check(t, vendorReviewTitle(soc2.ReportTypeTwo, "") == "SOC 2 Type 2 report review")
	assert.Check(t, vendorReviewTitle(docextract.ReportTypeUnknown, "") == "SOC 2 report review")
}

func TestVendorReviewDetailsWithoutFindings(t *testing.T) {
	score := 4
	assessment := gemini.ReportAssessment{
		Vendor:                gemini.ReportVendor{Name: "Meow Cloud", Description: "Cloud infrastructure services."},
		Summary:               "The report covers the Meow Platform scheduling service under the Security criteria.",
		AuditFirm:             "Demo Auditors LLP",
		Opinion:               soc2.OpinionUnqualified,
		PositiveObservations:  []string{"Independent assurance over a twelve month period."},
		RecommendationOutcome: "Approved, Low Risk",
		Recommendation:        "No material unresolved concerns.",
		PotentialRisks: []gemini.ReportRisk{
			{Title: "Shared Responsibility Model", Description: "Meow Cloud secures the platform; Openlane secures what it configures on it.", CustomerResponsibilities: []string{"IAM design", "Firewall rules"}},
		},
	}

	inputs := vendorReviewInputs{reportType: soc2.ReportTypeTwo, customer: "Openlane", rating: enums.VendorRiskRatingLow, score: &score}

	assert.Check(t, vendorReviewSummary(assessment, 0, inputs) == "Reviewed the Meow Cloud SOC 2 Type II report and found no exceptions. Demo Auditors LLP reported no deviations in the testing of Meow Cloud controls and issued an unqualified (clean) opinion. Overall risk posture is low, with a vendor risk score of 4 (low). The report covers the Meow Platform scheduling service under the Security criteria.\n\n**Outcome: Approved, Low Risk.**")

	assert.Check(t, vendorReviewDetails(assessment, nil, inputs) == vendorReviewGeneratedNote+"\n\n"+
		"**Service:** Cloud infrastructure services.\n\n"+
		"**Potential Risks**\n\n"+
		"**Shared Responsibility Model.** Meow Cloud secures the platform; Openlane secures what it configures on it.\n\n"+
		"Openlane retains responsibility for:\n\n- IAM design\n- Firewall rules\n\n"+
		"**Positive Observations**\n\n- Independent assurance over a twelve month period.\n\n"+
		"**Overall Recommendation**\n\n**Approved, Low Risk.** No material unresolved concerns.")
}

func TestVendorReviewDetailsWithFindings(t *testing.T) {
	assessment := gemini.ReportAssessment{Vendor: gemini.ReportVendor{Name: "Purr Analytics"}, Opinion: soc2.OpinionQualified}
	findings := []gemini.ReportFinding{
		{Description: "Two of 25 sampled demo hires had no background check.", Severity: "medium", RefCodes: []string{"CC1.1"}, Open: true, ManagementResponse: "Adding background checks to the demo onboarding checklist."},
		{Description: "Late access review"},
	}
	inputs := vendorReviewInputs{customer: "Openlane", rating: enums.VendorRiskRatingMedium}

	assert.Check(t, strings.HasPrefix(vendorReviewSummary(assessment, len(findings), inputs), "Reviewed the Purr Analytics SOC 2 report and found 2 exceptions. The auditor reported 2 deviations in the testing of Purr Analytics controls and issued a qualified opinion. Overall risk posture is medium."))

	details := vendorReviewDetails(assessment, findings, inputs)

	assert.Check(t, strings.Contains(details, "**Findings**\n\n- **CC1.1 · medium · open** — Two of 25 sampled demo hires had no background check. Management: Adding background checks to the demo onboarding checklist.\n- **remediated** — Late access review\n"))
	assert.Check(t, !strings.Contains(details, "Reviewed the Purr Analytics"))
	assert.Check(t, !strings.Contains(details, "**Potential Risks**"))
	assert.Check(t, !strings.HasSuffix(details, "\n"))
}

func TestVendorReviewFlagsScoringMismatch(t *testing.T) {
	score := 38
	assessment := gemini.ReportAssessment{Vendor: gemini.ReportVendor{Name: "Whiskers Catering"}, Opinion: soc2.OpinionUnqualified, RecommendationOutcome: "Approved, Moderate Risk"}
	inputs := vendorReviewInputs{customer: "Openlane", rating: enums.VendorRiskRatingCritical, score: &score, scores: []*generated.VendorRiskScore{
		{QuestionKey: "BCR-01.1", QuestionName: "DR tested", Answer: lo.ToPtr("false"), Impact: enums.VendorRiskImpactHigh, Likelihood: enums.VendorRiskLikelihoodLow, Score: 4, Notes: lo.ToPtr("No demo recovery test, section 4"), Tags: []string{reportScoringTag}},
		{QuestionKey: "IAM-05.1", QuestionName: "Least privilege", Answer: lo.ToPtr("true"), Notes: lo.ToPtr("Checked by hand")},
	}}

	summary := vendorReviewSummary(assessment, 0, inputs)

	assert.Check(t, strings.Contains(summary, "Overall risk posture is moderate, with a vendor risk score of 38 (critical)."))
	assert.Check(t, strings.HasSuffix(summary, "\n\n**Outcome: Approved, Moderate Risk.**\n\nCheck the scoring: vendor scoring rated this vendor critical with a score of 38, which does not match the review's Approved, Moderate Risk outcome."))

	details := vendorReviewDetails(assessment, nil, inputs)

	assert.Check(t, strings.HasSuffix(details, "one vendor scoring question answered from this report; see Vendor Scoring below for the answers and evidence."))
	assert.Check(t, !strings.Contains(details, "DR tested"))

	inputs.rating = enums.VendorRiskRatingMedium
	assert.Check(t, vendorScoringMismatch(assessment, inputs) == "")
}

func TestCountPhrase(t *testing.T) {
	assert.Check(t, countPhrase(0, "exception") == "no exceptions")
	assert.Check(t, countPhrase(1, "exception") == "one exception")
	assert.Check(t, countPhrase(3, "deviation") == "3 deviations")
}

func TestMissingServices(t *testing.T) {
	assert.DeepEqual(t, missingServices([]string{"Managed DNS"}, []string{"managed dns", " Load balancing ", "", "Load Balancing"}), []string{"Load balancing"})
	assert.Check(t, len(missingServices([]string{"A"}, []string{"a"})) == 0)
	assert.Check(t, len(missingServices(nil, nil)) == 0)
}
