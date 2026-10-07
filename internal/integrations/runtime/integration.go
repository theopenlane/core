package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/riverqueue/river"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/mapx"
	"github.com/theopenlane/core/v2/pkg/metrics"
)

const (
	// PendingInstallationTTL is how long a pending installation may wait for its auth flow
	PendingInstallationTTL = 168 * time.Hour
	// userInputNameKey is the user input key mirrored onto the installation name
	userInputNameKey = "name"
	// userInputPrimaryDirectoryKey is the user input key mirrored onto the installation's primary directory flag
	userInputPrimaryDirectoryKey = "primaryDirectory"
	// installationVersionAheadSnooze is how long a job snoozes when a newer binary stamped its installation, leaving it to that binary's workers
	installationVersionAheadSnooze = time.Minute
)

// IntegrationLookup holds the query constraints for resolving an integration
type IntegrationLookup struct {
	// IntegrationID is the unique identifier of the integration installation and required
	IntegrationID string
	// OwnerID scopes the integration to a specific owner, if provided
	OwnerID string
	// DefinitionID validates the integration belongs to a specific definition, if provided
	DefinitionID string
}

// ResolveIntegration resolves an integration by ID with optional owner and definition checks
func (r *Runtime) ResolveIntegration(ctx context.Context, lookup IntegrationLookup) (*ent.Integration, error) {
	if lookup.IntegrationID == "" {
		return nil, ErrIntegrationIDRequired
	}

	query := r.DB().Integration.Query().Where(integration.IDEQ(lookup.IntegrationID))
	if lookup.OwnerID != "" {
		query = query.Where(integration.OwnerIDEQ(lookup.OwnerID))
	}

	record, err := query.Only(ctx)
	if err != nil {
		return nil, err
	}

	if lookup.DefinitionID != "" && record.DefinitionID != lookup.DefinitionID {
		return nil, ErrInstallationDefinitionMismatch
	}

	return record, nil
}

// resolveCurrentIntegration resolves the integration an in-flight job processes, snoozing the job for a binary that has the version when a
// newer binary already stamped the installation, else upgrading it to the current definition version before anything reads it
func (r *Runtime) resolveCurrentIntegration(ctx context.Context, lookup IntegrationLookup) (*ent.Integration, error) {
	installation, err := r.ResolveIntegration(ctx, lookup)
	if err != nil {
		return nil, err
	}

	if current := r.Registry().Version(installation.DefinitionID); current != "" && outdated(current, installation.DefinitionVersion) {
		return nil, fmt.Errorf("%w: %w", ErrInstallationVersionAhead, river.JobSnooze(installationVersionAheadSnooze))
	}

	if err := r.ensureCurrentVersion(auth.EnsureIntegrationCaller(ctx, installation.OwnerID), installation); err != nil {
		return nil, err
	}

	return installation, nil
}

// ResolveOwnerIntegration returns the operational installation id for the definition and owner, or empty when none is selectable
func (r *Runtime) ResolveOwnerIntegration(ctx context.Context, definitionID, ownerID string, prefer ...func(*ent.Integration) bool) (string, error) {
	integrations, err := r.DB().Integration.Query().
		Where(
			integration.OwnerIDEQ(ownerID),
			integration.DefinitionIDEQ(definitionID),
			integration.StatusIn(enums.IntegrationOperationalStatuses...),
		).All(ctx)
	if err != nil {
		return "", err
	}

	switch {
	case len(integrations) == 1:
		return integrations[0].ID, nil
	case len(prefer) == 0:
		return "", nil
	}

	preferred, found := lo.Find(integrations, prefer[0])
	if !found {
		return "", nil
	}

	return preferred.ID, nil
}

// EnsureInstallation creates the installation, or resolves the one integrationID names, writing the submitted user input and operation config in one statement;
// a new installation of a definition without connections is identified by its own id, and one without credentials is created connected
func (r *Runtime) EnsureInstallation(ctx context.Context, ownerID, integrationID string, def types.Definition, userInput json.RawMessage, operationConfig map[string]json.RawMessage) (*ent.Integration, bool, error) {
	var current *ent.Integration

	if integrationID != "" {
		record, err := r.ResolveIntegration(ctx, IntegrationLookup{
			IntegrationID: integrationID,
			OwnerID:       ownerID,
			DefinitionID:  def.ID,
		})
		if err != nil {
			return nil, false, err
		}

		current = record
	}

	nextInput, nextConfig, err := installationInput(ctx, def, current, userInput, operationConfig)
	if err != nil {
		return nil, false, err
	}

	if current != nil {
		return r.updateInstallationInput(ctx, current, def, nextInput, nextConfig, userInput, operationConfig)
	}

	id := integration.DefaultID()

	create := r.DB().Integration.Create().
		SetID(id).
		SetOwnerID(ownerID).
		SetName(def.DisplayName).
		SetDescription(def.Description).
		SetKind(def.Family).
		SetIntegrationType(def.Category).
		SetDefinitionID(def.ID).
		SetDefinitionVersion(r.Registry().Version(def.ID)).
		SetDefinitionSlug(def.ID).
		SetFamily(def.Family).
		SetUserInput(nextInput).
		SetOperationConfig(nextConfig).
		SetStatus(enums.IntegrationStatusConnected)

	mirrorUserInput(create.Mutation(), userInput)

	if len(def.ConnectionList()) > 0 {
		create.SetStatus(enums.IntegrationStatusPending).SetExpiresAt(time.Now().Add(PendingInstallationTTL))
	}

	if len(def.ConnectionList()) == 0 {
		metadata := selfInstanceMetadata(id)
		display, _ := jsonx.ToMap(metadata.Display)

		create.SetInstallationMetadata(metadata).SetMetadata(mapx.PruneMapZeroAny(display))
	}

	record, err := create.Save(ctx)
	if err != nil {
		return nil, false, err
	}

	metrics.RecordIntegrationInstalled(def.ID)

	r.createVendor(ctx, ownerID, def, record.ID)

	return record, true, nil
}

// updateInstallationInput upgrades an existing installation to the current definition version with the submitted documents replacing what they replace,
// writes the submitted documents in one statement, and recovers an errored installation whose connection passes its health check once they are stored
func (r *Runtime) updateInstallationInput(ctx context.Context, installation *ent.Integration, def types.Definition, nextInput types.IntegrationUserInput, nextConfig types.IntegrationOperationConfig, userInput json.RawMessage, operationConfig map[string]json.RawMessage) (*ent.Integration, bool, error) {
	wasErrored := installation.Status == enums.IntegrationStatusErrored

	installation.UserInput, installation.OperationConfig = nextInput, nextConfig

	if err := r.ensureCurrentVersion(ctx, installation); err != nil {
		return nil, false, err
	}

	if jsonx.IsEmptyRawMessage(userInput) && len(operationConfig) == 0 {
		return installation, false, nil
	}

	nextInput, nextConfig = mergeInput(def, installation, userInput, operationConfig)

	update := r.DB().Integration.UpdateOneID(installation.ID).SetUserInput(nextInput).SetOperationConfig(nextConfig)
	mirrorUserInput(update.Mutation(), userInput)

	updated, err := update.Save(ctx)
	if err != nil {
		return nil, false, err
	}

	r.keystore().InvalidateClients(updated.ID)

	if !wasErrored {
		return updated, false, nil
	}

	verified, checkErr, err := r.verifyConnection(ctx, updated, def)
	if err != nil {
		return nil, false, err
	}

	if checkErr != nil {
		return nil, false, checkErr
	}

	if len(def.ConnectionList()) > 0 {
		if err := r.saveInstallationMetadata(privacy.DecisionContext(ctx, privacy.Allow), updated, def, verified.Connection, verified.Metadata); err != nil {
			return nil, false, err
		}
	}

	if err := r.ClearIntegrationUnhealthy(ctx, updated); err != nil {
		return nil, false, err
	}

	r.assessOperationHealth(ctx, updated, def)

	return updated, false, nil
}

// installationInput validates the submitted user input and operation config and returns them merged over what current stores, nil current for a new installation
func installationInput(ctx context.Context, def types.Definition, current *ent.Integration, userInput json.RawMessage, operationConfig map[string]json.RawMessage) (types.IntegrationUserInput, types.IntegrationOperationConfig, error) {
	nextInput, nextConfig := mergeInput(def, current, userInput, operationConfig)

	req := types.InstallationRequest{Integration: current}

	if current != nil {
		req.UserInput = current.UserInput.Data
	}

	if !jsonx.IsEmptyRawMessage(userInput) && def.UserInput != nil {
		if err := operations.ValidateInput(ctx, req, def.UserInput.Schema, def.UserInput.Validate, userInput, ErrUserInputInvalid); err != nil {
			return nextInput, nextConfig, err
		}
	}

	req.UserInput = nextInput.Data

	for _, name := range slices.Sorted(maps.Keys(operationConfig)) {
		operation, found := def.Operation(name)
		if !found || !operation.Stored {
			return nextInput, nextConfig, fmt.Errorf("%w: %s", ErrOperationNotFound, name)
		}

		if err := operations.ValidateInput(ctx, req, operation.Input.Schema, operation.Input.Validate, operationConfig[name], types.ErrOperationConfigInvalid); err != nil {
			return nextInput, nextConfig, fmt.Errorf("%w: operation %s", err, name)
		}
	}

	return nextInput, nextConfig, nil
}

// mergeInput returns the submitted user input and operation config merged over what current stores, nil current for a new installation
func mergeInput(def types.Definition, current *ent.Integration, userInput json.RawMessage, operationConfig map[string]json.RawMessage) (types.IntegrationUserInput, types.IntegrationOperationConfig) {
	var (
		nextInput  types.IntegrationUserInput
		nextConfig types.IntegrationOperationConfig
	)

	if current != nil {
		nextInput, nextConfig = current.UserInput, current.OperationConfig
	}

	if !jsonx.IsEmptyRawMessage(userInput) {
		nextInput = types.IntegrationUserInput{Data: jsonx.CloneRawMessage(userInput)}

		if def.UserInput != nil {
			nextInput.Layout = def.UserInput.Name
		}
	}

	for name, document := range operationConfig {
		nextConfig = nextConfig.With(name, jsonx.CloneRawMessage(document))
	}

	return nextInput, nextConfig
}

// mirrorUserInput sets the record fields the submitted user input carries: the installation name and its primary directory flag
func mirrorUserInput(mutation *ent.IntegrationMutation, userInput json.RawMessage) {
	if name, ok := jsonx.DecodeObjectKey[string](userInput, userInputNameKey); ok && name != "" {
		mutation.SetName(name)
	}

	if primary, ok := jsonx.DecodeObjectKey[bool](userInput, userInputPrimaryDirectoryKey); ok {
		mutation.SetPrimaryDirectory(primary)
	}
}

// createVendor best-effort links or creates the integration family as a vendor in the org
func (r *Runtime) createVendor(ctx context.Context, ownerID string, def types.Definition, integrationID string) {
	ctx = logx.WithFields(ctx, map[string]any{"vendor": def.Family, "org_id": ownerID})

	vendorIDs, err := r.DB().Entity.Query().Where(
		entity.Or(
			entity.NameEqualFold(def.Family),
			entity.DisplayNameEqualFold(def.Family),
		),
		entity.OwnerID(ownerID),
	).IDs(ctx)
	if err != nil {
		logx.FromContext(ctx).Info().Err(err).Msg("error looking for existing vendor, skipping creation")
		return
	}

	if len(vendorIDs) > 0 {
		// update the integration edges
		if err := r.DB().Entity.Update().Where(entity.IDIn(vendorIDs...)).AddIntegrationIDs(
			integrationID).Exec(ctx); err != nil {
			logx.FromContext(ctx).Info().Err(err).Msg("error update vendor edges to integration")
		}

		logx.FromContext(ctx).Debug().Msg("successfully updated vendor from integration setup")

		return
	}

	vendorInput := ent.CreateEntityInput{
		Name:           &def.Family,
		Tags:           []string{"integration"},
		ApprovedForUse: lo.ToPtr(true),
		IntegrationIDs: []string{integrationID},
	}

	subprocessors, err := r.DB().Subprocessor.Query().Where(
		subprocessor.NameEqualFold(def.Family),
	).All(ctx)
	if err == nil && len(subprocessors) > 0 {
		vendorInput.Description = &subprocessors[0].Description
	}

	existingEntityType, err := r.DB().EntityType.Query().
		Where(
			entitytype.NameEqualFold("vendor"),
			entitytype.OwnerID(ownerID),
		).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Info().Err(err).Msg("error looking up vendor entity type, skipping creation")
		return
	}

	if err := r.DB().Entity.Create().SetInput(vendorInput).SetEntityTypeID(existingEntityType.ID).Exec(ctx); err != nil {
		logx.FromContext(ctx).Info().Err(err).Msg("error creating vendor")
		return
	}

	logx.FromContext(ctx).Debug().Msg("successfully created vendor from integration setup")
}
