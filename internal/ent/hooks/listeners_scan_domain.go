package hooks

import (
	"encoding/json"

	"entgo.io/ent"
	"github.com/samber/do/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterGalaDomainScanSubmitListeners registers the listener that submits a openlane_domain_scan
// when the domain scan is created in a pending state
func RegisterGalaDomainScanSubmitListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return registerMutationListeners(g, entityops.MutationListener{
		Schema:     generated.TypeScan,
		Label:      "domain_scan",
		Operations: []string{ent.OpCreate.String()},
		Handle:     handleScanDomainCreated,
	})
}

// RegisterGalaDomainScanUpdateListener registers the listener that creates a pending domain scan for
// every current domain whenever an organization's settings domains field changes, this would then be picked
// up by the scan submit listener to run the scan
func RegisterGalaDomainScanUpdateListener(g *gala.Gala) ([]gala.ListenerID, error) {
	return registerMutationListeners(g, entityops.MutationListener{
		Schema:     generated.TypeOrganizationSetting,
		Label:      "domain_scan",
		Operations: []string{ent.OpUpdateOne.String()},
		Fields:     []string{organizationsetting.FieldDomains},
		Handle:     handleOrganizationSettingDomainsUpdated,
	})
}

// handleScanDomainCreated submits a newly created domain-type scan to the domain_scan gathering data via urlScanner, enrichment with browserRendering.JSON, and dns lookups
func handleScanDomainCreated(inv entityops.Invocation, _ entityops.MutationPayload) error {
	scanRecord, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Scan.Get)
	if err != nil || !ok {
		return err
	}

	if !isPendingDomainScan(scanRecord) {
		return nil
	}

	rt, err := do.Invoke[*intruntime.Runtime](inv.Injector)
	if err != nil || rt == nil {
		return nil
	}

	forceRefresh, _ := scanRecord.Metadata["forceRefresh"].(bool)

	config, err := json.Marshal(cloudflare.DomainScanRequest{
		OrganizationID: scanRecord.OwnerID,
		Domain:         scanRecord.Target,
		ForceRefresh:   forceRefresh,
	})
	if err != nil {
		return err
	}

	_, err = rt.Dispatch(inv.Context, types.DispatchRequest{
		DefinitionID: cloudflare.DefinitionID.ID(),
		Operation:    cloudflare.DomainScanRequestOp.Name(),
		Config:       config,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
	})

	return err
}

// isPendingDomainScan reports whether scanRecord is a domain-type Scan still awaiting submission
func isPendingDomainScan(scanRecord *generated.Scan) bool {
	return scanRecord.ScanType == enums.ScanTypeDomain &&
		scanRecord.Status == enums.ScanStatusPending &&
		scanRecord.PerformedBy == cloudflare.DomainScanPerformedBy
}

// handleOrganizationSettingDomainsUpdated requests a scan for every current domain whenever an
// organization's settings domains field changes; DomainScanRequestOp finds-or-creates and runs
// each one, the same operation the REST-replacing customer request and handleScanDomainCreated use
func handleOrganizationSettingDomainsUpdated(inv entityops.Invocation, _ entityops.MutationPayload) error {
	setting, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.OrganizationSetting.Get)
	if err != nil || !ok {
		return err
	}

	rt, err := do.Invoke[*intruntime.Runtime](inv.Injector)
	if err != nil || rt == nil {
		return nil
	}

	groupID := string(inv.Envelope.ID)

	// set internal context to bypass rate limits on scan requests
	dispatchCtx := rule.WithInternalContext(inv.Context)

	for _, domain := range setting.Domains {
		config, err := json.Marshal(cloudflare.DomainScanRequest{
			OrganizationID: setting.OrganizationID,
			Domain:         domain,
			GroupID:        groupID,
		})
		if err != nil {
			return err
		}

		if _, err := rt.Dispatch(dispatchCtx, types.DispatchRequest{
			DefinitionID: cloudflare.DefinitionID.ID(),
			Operation:    cloudflare.DomainScanRequestOp.Name(),
			Config:       config,
			RunType:      enums.IntegrationRunTypeEvent,
			Runtime:      true,
		}); err != nil {
			return err
		}
	}

	return nil
}
