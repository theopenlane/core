package hooks

import (
	"context"
	"encoding/json"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/internal/integrations/operations"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/internal/workflows"
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

	allowCtx := workflows.AllowContext(ctx.Context)

	scanRecord, err := client.Scan.Get(allowCtx, scanID)
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

	// forceRefresh, _ := scanRecord.Metadata["forceRefresh"].(bool)

	// return rt.HandleDomainScanSubmit(ctx.Context, scanRecord.OwnerID, scanRecord.ID, scanRecord.Target, forceRefresh)
	return nil
}

// sendSystemEmail marshals the input and executes a system email operation via
// the integration runtime on the ent client
func createDomainScan(ctx context.Context, client *generated.Client, input any) error {
	rt := intruntime.FromClient(ctx, client)
	if rt == nil {
		return nil
	}

	config, err := json.Marshal(input)
	if err != nil {
		return err
	}

	_, err = rt.Dispatch(ctx, types.DispatchRequest{
		DefinitionID: cloudflare.DefinitionID.ID(),
		Operation:    cloudflare.DomainScanSubmitOp.Name(),
		Config:       config,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
	})

	return err
}

// isPendingDomainScan reports whether scanRecord is a domain-type Scan still awaiting submission.
// Status is "pending" for scans queued via the REST domain-scan endpoint (see
// internal/httpserve/handlers/domainscan.go) and "processing" for the ent default used elsewhere
// (e.g. the generated createScan mutation) - either means "please go submit this"
func isPendingDomainScan(scanRecord *generated.Scan) bool {
	return scanRecord.ScanType == enums.ScanTypeDomain &&
		(scanRecord.Status == enums.ScanStatusPending || scanRecord.Status == enums.ScanStatusProcessing) &&
		scanRecord.PerformedBy == operations.DomainScanPerformedBy
}
