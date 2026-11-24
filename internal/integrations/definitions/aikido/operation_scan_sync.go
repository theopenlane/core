package aikido

import (
	"context"
	"fmt"

	"github.com/theopenlane/utils/mapx"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

const (
	// scanKindKey is the scan payload key naming which assessment listing the record came from
	scanKindKey = "kind"
	// scanKindPentest is the scan kind for AI pentest assessments
	scanKindPentest = "ai_pentest"
	// scanKindCodeAudit is the scan kind for AI code analysis assessments
	scanKindCodeAudit = "code_audit"
)

// scanKinds maps each assessment listing to the scan kind its records are tagged with
var scanKinds = []struct {
	// kind is the scan kind tagged on the records
	kind string
	// path is the list endpoint
	path string
}{
	{kind: scanKindPentest, path: pathPentests},
	{kind: scanKindCodeAudit, path: pathCodeAudits},
}

// runScanSync collects Aikido AI pentest and code analysis assessments for scan ingestion
func runScanSync(ctx context.Context, _ types.OperationRequest, client *APIClient, _ ScanSync) ([]types.IngestPayloadSet, error) {
	return CollectScans(ctx, client)
}

// CollectScans lists the AI pentest and code analysis assessments, whose issues link back to them by assessment id
func CollectScans(ctx context.Context, c *APIClient) ([]types.IngestPayloadSet, error) {
	envelopes := make([]types.MappingEnvelope, 0)

	for _, scanKind := range scanKinds {
		records, err := c.listAll(ctx, scanKind.path, nil, assessmentPageSize)
		if err != nil {
			return nil, fmt.Errorf("%w: %s: %w", ErrScansFetchFailed, scanKind.kind, err)
		}

		for _, record := range records {
			record[scanKindKey] = scanKind.kind

			envelope, err := recordEnvelope(ctx, mapx.Value[string](record, "id"), record)
			if err != nil {
				return nil, err
			}

			envelopes = append(envelopes, envelope)
		}
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaScan.Name,
			Envelopes: envelopes,
		},
	}, nil
}
