package hooks

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	entgen "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/cloudflare"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// init registers the domain scan listeners so gala setup picks them up automatically
func init() { registerListeners(DomainScanListeners) }

// DomainScanListeners submits pending domain scans and requests scans for changed
// organization domains
func DomainScanListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaScan,
			Operations: []string{entityops.OpCreate},
			Match: []entityops.FieldMatch{
				{Field: scan.FieldScanType, In: []string{string(enums.ScanTypeDomain)}},
				{Field: scan.FieldStatus, In: []string{string(enums.ScanStatusPending)}},
				{Field: scan.FieldPerformedBy, In: []string{cloudflare.DomainScanPerformedBy}},
			},
			Handle: entityops.RequireDep(handleScanDomainCreated),
		},
		entityops.MutationListener{
			Schema:      entityops.SchemaOrganizationSetting,
			Operations:  []string{entityops.OpUpdateOne},
			Fields:      []string{organizationsetting.FieldDomains},
			ContextKeys: []func(context.Context) context.Context{rule.WithInternalContext},
			Handle:      entityops.RequireDep(handleOrganizationSettingDomainsUpdated),
		},
	}
}

// handleScanDomainCreated submits a newly created domain-type scan to the domain_scan gathering data via urlScanner, enrichment with browserRendering.JSON, and dns lookups
func handleScanDomainCreated(inv entityops.Invocation, _ entityops.MutationPayload, rt *intruntime.Runtime) error {
	scanRecord, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Scan.Get)
	if err != nil || !ok {
		return err
	}

	// re-check the loaded row: a scan already processed between commit and delivery is done
	if !isPendingDomainScan(scanRecord) {
		return nil
	}

	forceRefresh, _ := scanRecord.Metadata["forceRefresh"].(bool)
	isBrandDesignOnly, _ := scanRecord.Metadata[cloudflare.DomainScanBrandDesignOnlyMetadataKey].(bool)
	applyBrandDesign, _ := scanRecord.Metadata[cloudflare.DomainScanApplyBrandDesignMetadataKey].(bool)

	return dispatchDomainScan(inv.Context, rt, cloudflare.DefinitionID.OperationTopics().Key(cloudflare.DomainScanRequestOp.Name(), string(inv.Envelope.ID)), cloudflare.DomainScanRequest{
		ScanID:           scanRecord.ID,
		OrganizationID:   scanRecord.OwnerID,
		Domain:           scanRecord.Target,
		ForceRefresh:     forceRefresh,
		BrandDesignOnly:  isBrandDesignOnly,
		ApplyBrandDesign: applyBrandDesign,
	})
}

// handleOrganizationSettingDomainsUpdated requests a scan for every current domain whenever
// an organization's settings domains field changes
func handleOrganizationSettingDomainsUpdated(inv entityops.Invocation, _ entityops.MutationPayload, rt *intruntime.Runtime) error {
	setting, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.OrganizationSetting.Get)
	if err != nil || !ok {
		return err
	}

	for idx, domain := range setting.Domains {
		if err := dispatchDomainScan(inv.Context, rt, cloudflare.DefinitionID.OperationTopics().Key(cloudflare.DomainScanRequestOp.Name(), string(inv.Envelope.ID), domain), cloudflare.DomainScanRequest{
			OrganizationID: setting.OrganizationID,
			Domain:         domain,
			GroupID:        string(inv.Envelope.ID),
			// just pick out the first one to apply it's brand design to the trustcenters
			ApplyBrandDesign: idx == 0,
		}); err != nil {
			return err
		}
	}

	return nil
}

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
