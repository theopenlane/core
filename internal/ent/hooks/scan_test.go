package hooks

import (
	"context"
	"testing"

	"entgo.io/ent"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
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
