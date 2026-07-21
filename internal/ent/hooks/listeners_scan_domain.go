package hooks

import (
	"encoding/json"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/integrations/definitions/cloudflare"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

// RegisterGalaDomainScanSubmitListeners registers the listener that submits a openlane_domain_scan
// when the domain scan is created in a pending state
func RegisterGalaDomainScanSubmitListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeScan),
		Name:       "scan.domain_submit",
		Operations: []string{ent.OpCreate.String()},
		Handle:     handleScanDomainCreated,
	})
}

// RegisterGalaDomainScanUpdateListener registers the listener that creates a pending domain scan for
// every current domain whenever an organization's settings domains field changes, this would then be picked
// up by the scan submit listener to run the scan
func RegisterGalaDomainScanUpdateListener(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeOrganizationSetting),
		Name:       "domainscan.organization_setting_update",
		Operations: []string{ent.OpUpdate.String(), ent.OpUpdateOne.String()},
		Handle:     handleOrganizationSettingDomainsUpdated,
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

	config, err := json.Marshal(cloudflare.DomainScanRequest{
		OrganizationID: scanRecord.OwnerID,
		Domain:         scanRecord.Target,
		ForceRefresh:   forceRefresh,
	})
	if err != nil {
		return err
	}

	_, err = rt.Dispatch(ctx.Context, types.DispatchRequest{
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
func handleOrganizationSettingDomainsUpdated(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if !eventqueue.MutationFieldChanged(payload, organizationsetting.FieldDomains) {
		return nil
	}

	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	settingID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || settingID == "" {
		return nil
	}

	setting, err := client.OrganizationSetting.Get(ctx.Context, settingID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	rt := intruntime.FromClient(ctx.Context, client)
	if rt == nil {
		return nil
	}

	groupID := string(ctx.Envelope.ID)

	for _, domain := range setting.Domains {
		config, err := json.Marshal(cloudflare.DomainScanRequest{
			OrganizationID: setting.OrganizationID,
			Domain:         domain,
			GroupID:        groupID,
		})
		if err != nil {
			return err
		}

		if _, err := rt.Dispatch(ctx.Context, types.DispatchRequest{
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
