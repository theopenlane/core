package hooks

import (
	"context"
	"strings"
	"testing"

	"entgo.io/ent"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/docextract/soc2"
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
	assert.DeepEqual(t, normalizedDomains([]string{" Example.io ", "example.io", "", "Docs.Example.io"}), []string{"cycle.io", "example.example.io"})
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
	details := vendorReviewDetails(gemini.ReportAssessment{Opinion: soc2.OpinionUnqualified, RiskRating: "low", RiskRationale: "Clean opinion.", PeriodEnd: "2026-07-28"}, nil)

	assert.Check(t, details == "Clean opinion.\n\n**Auditor opinion:** Unqualified (clean)  \n**Report period end:** 2026-07-28  \n**Risk rating:** LOW\n\n**Findings:** The auditor noted no exceptions.")
}

func TestVendorReviewDetailsWithFindings(t *testing.T) {
	details := vendorReviewDetails(gemini.ReportAssessment{Opinion: soc2.OpinionQualified, RiskRating: "medium", RiskRationale: "One access exception."}, []gemini.ReportFinding{
		{Description: "One of 25 accounts lacked MFA", Severity: "high", RefCodes: []string{"CC6.1"}, Open: true},
		{Description: "Late access review", Severity: "low"},
	})

	assert.Check(t, strings.Contains(details, "**Findings:**\n\n| Summary | Criticality | Related Controls |\n| --- | --- | --- |\n"))
	assert.Check(t, strings.Contains(details, "| One of 25 accounts lacked MFA | high | CC6.1 |"))
	assert.Check(t, strings.Contains(details, "| Late access review | low |  |"))
	assert.Check(t, !strings.HasSuffix(details, "\n"))
}

func TestMarkdownCell(t *testing.T) {
	assert.Check(t, markdownCell("a | b\nc  d") == `a \| b c d`)
}

func TestMissingServices(t *testing.T) {
	assert.DeepEqual(t, missingServices([]string{"Managed DNS"}, []string{"managed dns", " Load balancing ", "", "Load Balancing"}), []string{"Load balancing"})
	assert.Check(t, len(missingServices([]string{"A"}, []string{"a"})) == 0)
	assert.Check(t, len(missingServices(nil, nil)) == 0)
}
