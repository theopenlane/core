package runtime

import (
	"context"
	"fmt"

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
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/risk"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// ProvenanceSchemeVersion identifies the provenance scheme installations are converted to; an
// installation is converted once its DefinitionProviderState.ProvenanceVersion equals this value
const ProvenanceSchemeVersion = "v1"

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

// EnsureInstallationConverted converts every not-yet-converted installation in the given
// installation's organization to the current provenance scheme before any of them may ingest.
// Every sibling installation's rows must be attributed before any one of them ingests, because the
// first sync to run claims any row it matches that still lacks provenance. Sibling failures never
// fail the caller; the caller only fails when its own conversion could not complete
func (r *Runtime) EnsureInstallationConverted(ctx context.Context, installation *ent.Integration) error {
	orgCtx := entityops.WithEmissionVetoed(auth.EnsureIntegrationCaller(privacy.DecisionContext(ctx, privacy.Allow), installation.OwnerID))

	siblings, err := r.DB().Integration.Query().
		Where(integration.OwnerIDEQ(installation.OwnerID)).
		Order(integration.ByCreatedAt(sql.OrderAsc()), integration.ByID(sql.OrderAsc())).
		All(orgCtx)
	if err != nil {
		return err
	}

	var callerErr error

	for _, sibling := range siblings {
		target := sibling
		if target.ID == installation.ID {
			target = installation
		}

		def, converted, stateErr := r.provenanceConversionState(target)
		if stateErr != nil && target.ID == installation.ID {
			callerErr = stateErr
		}

		if stateErr != nil || converted {
			continue
		}

		if err := r.convertInstallationProvenance(orgCtx, target, def); err != nil && target.ID == installation.ID {
			callerErr = err
		}
	}

	if _, converted, stateErr := r.provenanceConversionState(installation); stateErr == nil && converted {
		return nil
	}

	if installation.InstallationMetadata.Display.ExternalID == "" {
		return ErrInstallationInstanceIDRequired
	}

	return fmt.Errorf("%w: %w", ErrInstallationConversionFailed, callerErr)
}

// provenanceConversionState resolves one installation's definition and reports whether the
// installation has already been converted to the current provenance scheme; an installation whose
// definition is not registered or whose definition id is empty cannot be evaluated and is reported
// as not converted alongside the resolution error
func (r *Runtime) provenanceConversionState(installation *ent.Integration) (types.Definition, bool, error) {
	def, err := r.resolveDefinitionForInstallation(installation)
	if err != nil {
		return types.Definition{}, false, err
	}

	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return def, false, err
	}

	return def, state.ProvenanceVersion == ProvenanceSchemeVersion, nil
}

// convertInstallationProvenance refreshes one installation's instance id, stamps every ingest
// schema's fill-only provenance columns for that installation, and persists the conversion marker
// once every stamp succeeds against a resolved instance id; otherwise the installation is left
// unconverted so the next ingest retries
func (r *Runtime) convertInstallationProvenance(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	instCtx := intobvs.WithInstallation(ctx, installation)

	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return err
	}

	if state.CredentialRef == (types.CredentialSlotID{}) && len(def.Connections) > 0 {
		return r.persistProvenanceConversion(instCtx, installation, def)
	}

	if err := r.RefreshInstallationMetadata(instCtx, installation); err != nil {
		logx.FromContext(instCtx).Error().Err(err).Msg("provenance conversion: instance metadata refresh failed; falling back to the installation's stored instance id")
	}

	instanceID := installation.InstallationMetadata.Display.ExternalID

	stamped, stampErr := stampInstallationProvenance(instCtx, r.DB(), installation, instanceID)
	if stampErr != nil {
		return stampErr
	}

	if instanceID == "" {
		return ErrInstallationInstanceIDRequired
	}

	if err := r.persistProvenanceConversion(instCtx, installation, def); err != nil {
		return err
	}

	logx.FromContext(instCtx).Info().Int("stamped_rows", stamped).Msg("provenance conversion completed for installation")

	return nil
}

// stampInstallationProvenance runs every provenance stamp against one installation's records,
// running every stamp regardless of earlier failures, and reports the total rows stamped and the
// first stamp error encountered, if any
func stampInstallationProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	var stamped int

	var firstErr error

	for _, stamp := range provenanceStamps {
		affected, err := stamp.Apply(ctx, db, installation, instanceID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("schema", stamp.Schema).Msg("provenance conversion: stamp failed")

			if firstErr == nil {
				firstErr = err
			}

			continue
		}

		stamped += affected
	}

	return stamped, firstErr
}

// persistProvenanceConversion marks one installation as converted to the current provenance
// scheme, preserving its existing credential reference, and updates the in-memory installation so
// the caller observes the converted state immediately
func (r *Runtime) persistProvenanceConversion(ctx context.Context, installation *ent.Integration, def types.Definition) error {
	state, err := def.ProviderState(installation.ProviderState)
	if err != nil {
		return err
	}

	state.ProvenanceVersion = ProvenanceSchemeVersion

	next, err := def.WithProviderState(installation.ProviderState, state)
	if err != nil {
		return err
	}

	if err := r.DB().Integration.UpdateOneID(installation.ID).SetProviderState(next).Exec(ctx); err != nil {
		return err
	}

	installation.ProviderState = next

	return nil
}

// stampActionPlanProvenance fills provenance on action plans linked to the installation
func stampActionPlanProvenance(ctx context.Context, db *ent.Client, installation *ent.Integration, instanceID string) (int, error) {
	predicates := []predicate.ActionPlan{
		actionplan.And(
			actionplan.HasIntegrationsWith(integration.ID(installation.ID)),
			actionplan.Not(actionplan.HasIntegrationsWith(integration.IDNEQ(installation.ID))),
			actionplan.Or(actionplan.SourceDefinitionIDIsNil(), actionplan.SourceDefinitionID(""), actionplan.SourceInstanceIDIsNil(), actionplan.SourceInstanceID("")),
			actionplan.Or(actionplan.ManagedByIsNil(), actionplan.ManagedBy(""), actionplan.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, actionplan.And(actionplan.ManagedBy(installation.ID), actionplan.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.ActionPlan.Update().
		Where(actionplan.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.Asset{
		asset.And(
			asset.IntegrationID(installation.ID),
			asset.Or(asset.SourceDefinitionIDIsNil(), asset.SourceDefinitionID(""), asset.SourceInstanceIDIsNil(), asset.SourceInstanceID("")),
			asset.Or(asset.ManagedByIsNil(), asset.ManagedBy(""), asset.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, asset.And(asset.ManagedBy(installation.ID), asset.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.Asset.Update().
		Where(asset.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.CheckResult{
		checkresult.And(
			checkresult.IntegrationID(installation.ID),
			checkresult.Or(checkresult.SourceDefinitionIDIsNil(), checkresult.SourceDefinitionID(""), checkresult.SourceInstanceIDIsNil(), checkresult.SourceInstanceID("")),
			checkresult.Or(checkresult.ManagedByIsNil(), checkresult.ManagedBy(""), checkresult.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, checkresult.And(checkresult.ManagedBy(installation.ID), checkresult.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.CheckResult.Update().
		Where(checkresult.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.Contact{
		contact.And(
			contact.IntegrationID(installation.ID),
			contact.Or(contact.SourceDefinitionIDIsNil(), contact.SourceDefinitionID(""), contact.SourceInstanceIDIsNil(), contact.SourceInstanceID("")),
			contact.Or(contact.ManagedByIsNil(), contact.ManagedBy(""), contact.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, contact.And(contact.ManagedBy(installation.ID), contact.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.Contact.Update().
		Where(contact.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.DirectoryAccount{
		directoryaccount.And(
			directoryaccount.IntegrationID(installation.ID),
			directoryaccount.Or(directoryaccount.SourceDefinitionIDIsNil(), directoryaccount.SourceDefinitionID(""), directoryaccount.SourceInstanceIDIsNil(), directoryaccount.SourceInstanceID("")),
			directoryaccount.Or(directoryaccount.ManagedByIsNil(), directoryaccount.ManagedBy(""), directoryaccount.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, directoryaccount.And(directoryaccount.ManagedBy(installation.ID), directoryaccount.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.DirectoryAccount.Update().
		Where(directoryaccount.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.DirectoryGroup{
		directorygroup.And(
			directorygroup.IntegrationID(installation.ID),
			directorygroup.Or(directorygroup.SourceDefinitionIDIsNil(), directorygroup.SourceDefinitionID(""), directorygroup.SourceInstanceIDIsNil(), directorygroup.SourceInstanceID("")),
			directorygroup.Or(directorygroup.ManagedByIsNil(), directorygroup.ManagedBy(""), directorygroup.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, directorygroup.And(directorygroup.ManagedBy(installation.ID), directorygroup.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.DirectoryGroup.Update().
		Where(directorygroup.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.DirectoryMembership{
		directorymembership.And(
			directorymembership.IntegrationID(installation.ID),
			directorymembership.Or(directorymembership.SourceDefinitionIDIsNil(), directorymembership.SourceDefinitionID(""), directorymembership.SourceInstanceIDIsNil(), directorymembership.SourceInstanceID("")),
			directorymembership.Or(directorymembership.ManagedByIsNil(), directorymembership.ManagedBy(""), directorymembership.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, directorymembership.And(directorymembership.ManagedBy(installation.ID), directorymembership.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.DirectoryMembership.Update().
		Where(directorymembership.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.Entity{
		entity.And(
			entity.HasIntegrationsWith(integration.ID(installation.ID)),
			entity.Not(entity.HasIntegrationsWith(integration.IDNEQ(installation.ID))),
			entity.Or(entity.SourceDefinitionIDIsNil(), entity.SourceDefinitionID(""), entity.SourceInstanceIDIsNil(), entity.SourceInstanceID("")),
			entity.Or(entity.ManagedByIsNil(), entity.ManagedBy(""), entity.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, entity.And(entity.ManagedBy(installation.ID), entity.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.Entity.Update().
		Where(entity.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.Finding{
		finding.And(
			finding.HasIntegrationsWith(integration.ID(installation.ID)),
			finding.Not(finding.HasIntegrationsWith(integration.IDNEQ(installation.ID))),
			finding.Or(finding.SourceDefinitionIDIsNil(), finding.SourceDefinitionID(""), finding.SourceInstanceIDIsNil(), finding.SourceInstanceID("")),
			finding.Or(finding.ManagedByIsNil(), finding.ManagedBy(""), finding.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, finding.And(finding.ManagedBy(installation.ID), finding.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.Finding.Update().
		Where(finding.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.InternalPolicy{
		internalpolicy.And(
			internalpolicy.HasIntegrationsWith(integration.ID(installation.ID)),
			internalpolicy.Not(internalpolicy.HasIntegrationsWith(integration.IDNEQ(installation.ID))),
			internalpolicy.Or(internalpolicy.SourceDefinitionIDIsNil(), internalpolicy.SourceDefinitionID(""), internalpolicy.SourceInstanceIDIsNil(), internalpolicy.SourceInstanceID("")),
			internalpolicy.Or(internalpolicy.ManagedByIsNil(), internalpolicy.ManagedBy(""), internalpolicy.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, internalpolicy.And(internalpolicy.ManagedBy(installation.ID), internalpolicy.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.InternalPolicy.Update().
		Where(internalpolicy.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.Risk{
		risk.And(
			risk.IntegrationID(installation.ID),
			risk.Or(risk.SourceDefinitionIDIsNil(), risk.SourceDefinitionID(""), risk.SourceInstanceIDIsNil(), risk.SourceInstanceID("")),
			risk.Or(risk.ManagedByIsNil(), risk.ManagedBy(""), risk.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, risk.And(risk.ManagedBy(installation.ID), risk.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.Risk.Update().
		Where(risk.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

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
	predicates := []predicate.Vulnerability{
		vulnerability.And(
			vulnerability.HasIntegrationsWith(integration.ID(installation.ID)),
			vulnerability.Not(vulnerability.HasIntegrationsWith(integration.IDNEQ(installation.ID))),
			vulnerability.Or(vulnerability.SourceDefinitionIDIsNil(), vulnerability.SourceDefinitionID(""), vulnerability.SourceInstanceIDIsNil(), vulnerability.SourceInstanceID("")),
			vulnerability.Or(vulnerability.ManagedByIsNil(), vulnerability.ManagedBy(""), vulnerability.ManagedBy(installation.ID)),
		),
	}

	if instanceID != "" {
		predicates = append(predicates, vulnerability.And(vulnerability.ManagedBy(installation.ID), vulnerability.SourceInstanceIDNEQ(instanceID)))
	}

	update := db.Vulnerability.Update().
		Where(vulnerability.Or(predicates...)).
		SetSourceDefinitionID(installation.DefinitionID).
		SetManagedBy(installation.ID)

	if installation.DefinitionVersion != "" {
		update.SetSourceDefinitionVersion(installation.DefinitionVersion)
	}

	if instanceID != "" {
		update.SetSourceInstanceID(instanceID)
	}

	return update.Save(ctx)
}
