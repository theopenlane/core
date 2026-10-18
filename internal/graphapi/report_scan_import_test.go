package graphapi

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/graphapi/model"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
)

func TestBuildReportScanImportEnvelope(t *testing.T) {
	input := model.ImportReportScanInput{
		ScanID: "scan-1",
		Sections: []*model.ImportReportScanSectionInput{
			{Section: "programs", Keys: []string{"Acme SOC 2"}},
			{Section: "controls", Keys: []string{"CC1.1", "CC1.2"}, Edited: []map[string]any{{"ref_code": "CC1.2", "title": "edited"}}},
			{Section: "controls", Keys: []string{"CC2.1"}},
			{Section: "findings", Keys: []string{}},
		},
	}

	got := buildReportScanImportEnvelope("org-1", input)

	assert.DeepEqual(t, got, gemini.ReportScanImport{
		OrganizationID: "org-1",
		ScanID:         "scan-1",
		Accepted: map[string][]string{
			"programs": {"Acme SOC 2"},
			"controls": {"CC1.1", "CC1.2", "CC2.1"},
			"findings": nil,
		},
		Edited: map[string][]map[string]any{
			"controls": {{"ref_code": "CC1.2", "title": "edited"}},
		},
	})
}
