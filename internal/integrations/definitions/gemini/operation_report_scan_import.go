package gemini

import (
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// ReportScanImport imports the items a reviewer accepted from a parsed SOC 2 report into the organization
type ReportScanImport struct {
	// OrganizationID is the organization the records are imported into
	OrganizationID string `json:"organizationId"`
	// ScanID is the report scan whose stored sections are imported
	ScanID string `json:"scanId"`
	// Accepted lists, per section, the keys of the items the reviewer accepted
	Accepted map[string][]string `json:"accepted"`
	// Edited holds, per section, reviewer-edited items that replace the stored item with the same key
	Edited map[string][]map[string]any `json:"edited,omitempty"`
}

// ReportScanImportOp is the operation ref for importing the accepted items of a parsed report
var ReportScanImportOp = types.OperationPayloadOf[ReportScanImport]().
	IngestsRequest(ingestReportScanImport).
	AfterIngest(afterReportScanImport).
	Ingest(reportScanImportContracts...).
	Policy(types.ExecutionPolicy{SkipRunRecord: true}).
	CustomerSelectable(false).
	Internal()
