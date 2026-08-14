package hooks

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/scan"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// domainScanListenerLabel names the domain scan listeners for resolution logging
const domainScanListenerLabel = "domain_scan"

// DomainScanListeners returns the domain scan listeners: one submits a openlane_domain_scan
// when a domain scan is created in a pending state, the other creates a pending domain scan
// for every current domain whenever an organization's settings domains field changes
func DomainScanListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaScan,
			Label:      domainScanListenerLabel,
			Operations: []string{entityops.OpCreate},
			Match: []entityops.FieldMatch{
				{Field: scan.FieldScanType, In: []string{string(enums.ScanTypeDomain)}},
				{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusPending)}},
				{Field: scan.FieldPerformedBy, In: []string{cloudflare.DomainScanPerformedBy}},
			},
			Handle: handleScanDomainCreated,
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaOrganizationSetting,
			Label:      domainScanListenerLabel,
			Operations: []string{entityops.OpUpdateOne},
			Fields:     []string{organizationsetting.FieldDomains},
			Handle:     handleOrganizationSettingDomainsUpdated,
		},
	}
}

// handleScanDomainCreated submits a newly created domain-type scan to the domain_scan gathering data via urlScanner, enrichment with browserRendering.JSON, and dns lookups
func handleScanDomainCreated(inv entityops.Invocation, _ entityops.MutationPayload) error {
	scanRecord, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Scan.Get)
	if err != nil || !ok {
		return err
	}

	// re-check the loaded row: a scan already processed between commit and delivery is done
	if !isPendingDomainScan(scanRecord) {
		return nil
	}

	rt, ok := gala.Resolve[*intruntime.Runtime](inv.Context, inv.Injector, domainScanListenerLabel)
	if !ok {
		return nil
	}

	forceRefresh, _ := scanRecord.Metadata["forceRefresh"].(bool)

	return dispatchDomainScan(inv.Context, rt, dispatchUniqueKeys.Key(cloudflare.DomainScanRequestOp.Name(), string(inv.Envelope.ID)), cloudflare.DomainScanRequest{
		OrganizationID: scanRecord.OwnerID,
		Domain:         scanRecord.Target,
		ForceRefresh:   forceRefresh,
	})
}

// handleOrganizationSettingDomainsUpdated requests a scan for every current domain whenever an
// organization's settings domains field changes; DomainScanRequestOp finds-or-creates and runs
// each one, the same operation the REST-replacing customer request and handleScanDomainCreated use
func handleOrganizationSettingDomainsUpdated(inv entityops.Invocation, _ entityops.MutationPayload) error {
	setting, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.OrganizationSetting.Get)
	if err != nil || !ok {
		return err
	}

	rt, ok := gala.Resolve[*intruntime.Runtime](inv.Context, inv.Injector, domainScanListenerLabel)
	if !ok {
		return nil
	}

	groupID := string(inv.Envelope.ID)

	// set internal context to bypass rate limits on scan requests
	dispatchCtx := rule.WithInternalContext(inv.Context)

	for _, domain := range setting.Domains {
		if err := dispatchDomainScan(dispatchCtx, rt, dispatchUniqueKeys.Key(cloudflare.DomainScanRequestOp.Name(), string(inv.Envelope.ID), domain), cloudflare.DomainScanRequest{
			OrganizationID: setting.OrganizationID,
			Domain:         domain,
			GroupID:        groupID,
		}); err != nil {
			return err
		}
	}

	return nil
}

// dispatchUniqueKeys is the dedup key namespace for event-triggered dispatch runs
var dispatchUniqueKeys = gala.NewUniqueKeyNamespace("dispatch")

// isPendingDomainScan reports whether scanRecord is a domain-type Scan still awaiting submission
func isPendingDomainScan(scanRecord *entgen.Scan) bool {
	return scanRecord.ScanType == enums.ScanTypeDomain &&
		scanRecord.Status == enums.ScanStatusPending &&
		scanRecord.PerformedBy == cloudflare.DomainScanPerformedBy
}

// dispatchDomainScan marshals the request and dispatches the cloudflare domain scan operation
// as an event-triggered runtime run
func dispatchDomainScan(ctx context.Context, rt *intruntime.Runtime, uniqueKey string, req cloudflare.DomainScanRequest) error {
	config, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = rt.Dispatch(ctx, types.DispatchRequest{
		DefinitionID: cloudflare.DefinitionID.ID(),
		Operation:    cloudflare.DomainScanRequestOp.Name(),
		Config:       config,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
		// a listener retry after a crashed cycle must not enqueue the scan twice
		UniqueKey: uniqueKey,
	})

	return err
}
