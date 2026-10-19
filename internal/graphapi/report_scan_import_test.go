package graphapi

import (
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/graphapi/model"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/gemini"
)

func TestBuildReportScanImportEnvelope(t *testing.T) {
	input := model.ImportReportScanInput{
		ScanID:   "scan-1",
		Programs: &model.ImportReportScanSectionInput{Keys: []string{"Acme SOC 2"}},
		Controls: &model.ImportReportScanSectionInput{Keys: []string{"CC1.1", "CC1.2"}},
		Findings: &model.ImportReportScanSectionInput{Keys: []string{}},
		Platforms: &model.ImportReportScanPlatformsInput{
			Keys:   []string{"Meow Platform"},
			Edited: []*generated.CreatePlatformInput{{Name: "Meow Platform", BusinessPurpose: lo.ToPtr("Scheduling")}},
		},
		SystemDetails: &model.ImportReportScanSystemDetailsInput{
			Keys: []string{"Scheduler"},
			Edited: []*model.ReportScanSystemDetailEdit{
				{SystemDetail: &generated.CreateSystemDetailInput{SystemName: "Scheduler", Description: lo.ToPtr("edited")}, PlatformName: lo.ToPtr("Meow Platform")},
			},
		},
	}

	got, err := buildReportScanImportEnvelope("org-1", input)

	assert.NilError(t, err)
	assert.DeepEqual(t, got, gemini.ReportScanImport{
		OrganizationID: "org-1",
		ScanID:         "scan-1",
		Accepted: map[string][]string{
			"programs":      {"Acme SOC 2"},
			"controls":      {"CC1.1", "CC1.2"},
			"findings":      {},
			"platforms":     {"Meow Platform"},
			"systemdetails": {"Scheduler"},
		},
		Edited: map[string][]map[string]any{
			"platforms":     {{"name": "Meow Platform", "business_purpose": "Scheduling"}},
			"systemdetails": {{"system_name": "Scheduler", "description": "edited", "platform_name": "Meow Platform"}},
		},
	})
}
