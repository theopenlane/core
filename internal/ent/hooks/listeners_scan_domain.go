package hooks

import (
	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/operations"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterGalaDomainScanSubmitListeners registers the listener that submits a openlane_domain_scan to the
// domain scan, which includes cloudflare url scanner, browserrendering.json, and more when the domain
// scan is created in a pending state. This is skipped when coming from organizations settings updates
// because those are submitted via the hook in order to determine the new vs. old domains
func RegisterGalaDomainScanSubmitListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeScan),
		Name:       "scan.domain_submit",
		Operations: []string{ent.OpCreate.String()},
		Handle:     handleScanDomainCreated,
	})
}

// handleScanDomainCreated submits a newly created domain-type scan to the domain_scan gathering data via urlScanner, enrichment with browserRendering.JSON, and dns lookups
func handleScanDomainCreated(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	scanID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || scanID == "" {
		return nil
	}

	scanRecord, err := client.Scan.Get(ctx.Context, scanID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	if !isPendingDomainScan(scanRecord) {
		return nil
	}

	rt := intruntime.FromClient(ctx.Context, client)
	if rt == nil {
		return nil
	}

	forceRefresh, _ := scanRecord.Metadata["forceRefresh"].(bool)

	receipt := rt.Gala().EmitWithHeaders(ctx.Context, operations.DomainScanSubmitTopic, operations.DomainScanSubmitEnvelope{
		OrganizationID: scanRecord.OwnerID,
		ScanID:         scanRecord.ID,
		Domain:         scanRecord.Target,
		ForceRefresh:   forceRefresh,
	}, gala.Headers{})

	return receipt.Err
}

// isPendingDomainScan reports whether scanRecord is a domain-type Scan still awaiting submission
func isPendingDomainScan(scanRecord *generated.Scan) bool {
	return scanRecord.ScanType == enums.ScanTypeDomain &&
		scanRecord.Status == enums.ScanStatusPending &&
		scanRecord.PerformedBy == operations.DomainScanPerformedBy
}
