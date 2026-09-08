package serveropts

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/actionplan"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
	"github.com/theopenlane/core/v2/internal/ent/generated/checkresult"
	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/v2/internal/ent/generated/risk"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/runtime"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// provenanceStamp is one schema's fill-only provenance update for records tied to an installation
type provenanceStamp struct {
	// Schema names the record type for logging
	Schema string
	// Apply runs the bulk fill-only update and reports how many rows were stamped
	Apply func(context.Context, *ent.Client, *ent.Integration, string) (int, error)
}

// provenanceStamps covers every ingest schema whose records carry an installation linkage; Procedure
// has no integration edge, so its records gain provenance on their next changed sync write instead
var provenanceStamps = []provenanceStamp{
	{Schema: "action_plan", Apply: stampActionPlanProvenance},
	{Schema: "asset", Apply: stampAssetProvenance},
	{Schema: "check_result", Apply: stampCheckResultProvenance},
	{Schema: "contact", Apply: stampContactProvenance},
	{Schema: "directory_account", Apply: stampDirectoryAccountProvenance},
	{Schema: "directory_group", Apply: stampDirectoryGroupProvenance},
	{Schema: "directory_membership", Apply: stampDirectoryMembershipProvenance},
	{Schema: "entity", Apply: stampEntityProvenance},
	{Schema: "finding", Apply: stampFindingProvenance},
	{Schema: "internal_policy", Apply: stampInternalPolicyProvenance},
	{Schema: "risk", Apply: stampRiskProvenance},
	{Schema: "vulnerability", Apply: stampVulnerabilityProvenance},
}

// backfillIntegrationProvenance stamps source_definition_id, source_definition_version, and
// source_instance_id on existing integration-sourced records that predate the provenance fields.
// Only rows with no provenance are written, the instance identity comes from the installation's
// freshly re-resolved metadata, and the writes run as the integration virtual actor with
// audit-log bypass
func backfillIntegrationProvenance(ctx context.Context, dbClient *ent.Client, rt *runtime.Runtime) {
	ctx = entityops.WithEmissionVetoed(auth.EnsureIntegrationCaller(ctx, ""))

	installations, err := dbClient.Integration.Query().
		Order(integration.ByCreatedAt(sql.OrderAsc()), integration.ByID(sql.OrderAsc())).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("backfill: failed to query integrations for provenance stamping")

		return
	}

	var stamped, failed int

	for _, installation := range installations {
		if installation.DefinitionID == "" {
			continue
		}

		installCtx := intobvs.WithInstallation(ctx, installation)

		if err := rt.RefreshInstallationMetadata(installCtx, installation); err != nil {
			failed++

			logx.FromContext(installCtx).Error().Err(err).Msg("backfill: metadata refresh failed; skipping provenance stamping for installation")

			continue
		}

		instanceID := installation.InstallationMetadata.Display.ExternalID

		for _, stamp := range provenanceStamps {
			affected, err := stamp.Apply(installCtx, dbClient, installation, instanceID)
			if err != nil {
				failed++

				logx.FromContext(installCtx).Error().Err(err).Str("schema", stamp.Schema).Msg("backfill: provenance stamp failed")

				continue
			}

			stamped += affected
		}
	}

	logx.FromContext(ctx).Info().Int("stamped_rows", stamped).Int("failed_stamps", failed).Int("installations", len(installations)).Msg("backfill: integration provenance completed")
}

// stampActionPlanProvenance fills provenance on action plans linked to the installation
func stampActionPlanProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.ActionPlan.Update().
		Where(actionplan.HasIntegrationsWith(integration.ID(installation.ID)), actionplan.Or(actionplan.SourceDefinitionIDIsNil(), actionplan.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampAssetProvenance fills provenance on assets discovered by the installation
func stampAssetProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.Asset.Update().
		Where(asset.IntegrationID(installation.ID), asset.Or(asset.SourceDefinitionIDIsNil(), asset.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampCheckResultProvenance fills provenance on check results owned by the installation
func stampCheckResultProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.CheckResult.Update().
		Where(checkresult.IntegrationID(installation.ID), checkresult.Or(checkresult.SourceDefinitionIDIsNil(), checkresult.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampContactProvenance fills provenance on contacts sourced by the installation
func stampContactProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.Contact.Update().
		Where(contact.IntegrationID(installation.ID), contact.Or(contact.SourceDefinitionIDIsNil(), contact.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampDirectoryAccountProvenance fills provenance on directory accounts owned by the installation
func stampDirectoryAccountProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.DirectoryAccount.Update().
		Where(directoryaccount.IntegrationID(installation.ID), directoryaccount.Or(directoryaccount.SourceDefinitionIDIsNil(), directoryaccount.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampDirectoryGroupProvenance fills provenance on directory groups owned by the installation
func stampDirectoryGroupProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.DirectoryGroup.Update().
		Where(directorygroup.IntegrationID(installation.ID), directorygroup.Or(directorygroup.SourceDefinitionIDIsNil(), directorygroup.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampDirectoryMembershipProvenance fills provenance on directory memberships owned by the installation
func stampDirectoryMembershipProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.DirectoryMembership.Update().
		Where(directorymembership.IntegrationID(installation.ID), directorymembership.Or(directorymembership.SourceDefinitionIDIsNil(), directorymembership.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampEntityProvenance fills provenance on entities linked to the installation
func stampEntityProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.Entity.Update().
		Where(entity.HasIntegrationsWith(integration.ID(installation.ID)), entity.Or(entity.SourceDefinitionIDIsNil(), entity.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampFindingProvenance fills provenance on findings linked to the installation
func stampFindingProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.Finding.Update().
		Where(finding.HasIntegrationsWith(integration.ID(installation.ID)), finding.Or(finding.SourceDefinitionIDIsNil(), finding.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampInternalPolicyProvenance fills provenance on internal policies linked to the installation
func stampInternalPolicyProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.InternalPolicy.Update().
		Where(internalpolicy.HasIntegrationsWith(integration.ID(installation.ID)), internalpolicy.Or(internalpolicy.SourceDefinitionIDIsNil(), internalpolicy.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampRiskProvenance fills provenance on risks surfaced by the installation
func stampRiskProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.Risk.Update().
		Where(risk.IntegrationID(installation.ID), risk.Or(risk.SourceDefinitionIDIsNil(), risk.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}

// stampVulnerabilityProvenance fills provenance on vulnerabilities linked to the installation
func stampVulnerabilityProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	update := db.Vulnerability.Update().
		Where(vulnerability.HasIntegrationsWith(integration.ID(installation.ID)), vulnerability.Or(vulnerability.SourceDefinitionIDIsNil(), vulnerability.SourceDefinitionID(""))).
		SetSourceDefinitionID(installation.DefinitionID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}
