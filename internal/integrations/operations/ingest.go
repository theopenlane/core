package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// IngestOptions carries the minimal ingest-time metadata needed by persistence.
type IngestOptions struct {
	DirectorySyncRunID               string
	SkipDirectorySyncRunFinalization bool
}

type installationFilterConfig struct {
	FilterExpr string `json:"filterExpr,omitempty"`
}

// ProcessIngest decodes one operation response into payload sets and persists them.
func ProcessIngest(ctx context.Context, reg *registry.Registry, db *ent.Client, installation *ent.Integration, operation types.OperationRegistration, response json.RawMessage) error {
	var payloadSets []types.IngestPayloadSet
	if err := json.Unmarshal(response, &payloadSets); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestPayloadsInvalid, err)
	}

	return ProcessPayloadSets(ctx, reg, db, installation, operation.Ingest, payloadSets)
}

// ProcessPayloadSets persists one batch of mapped payload sets with default options.
func ProcessPayloadSets(ctx context.Context, reg *registry.Registry, db *ent.Client, installation *ent.Integration, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet) error {
	return ProcessPayloadSetsWithOptions(ctx, reg, db, installation, contracts, payloadSets, IngestOptions{})
}

// ProcessPayloadSetsWithOptions persists one batch of mapped payload sets.
func ProcessPayloadSetsWithOptions(ctx context.Context, reg *registry.Registry, db *ent.Client, installation *ent.Integration, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) (err error) {
	if reg == nil || db == nil || installation == nil {
		return ErrIngestPersistFailed
	}

	definition, ok := reg.Definition(installation.DefinitionID)
	if !ok {
		return ErrIngestDefinitionNotFound
	}

	installationFilterExpr, err := resolveInstallationFilterExpr(installation)
	if err != nil {
		return ErrIngestInstallationFilterConfigInvalid
	}

	directorySyncRunID := options.DirectorySyncRunID
	if directorySyncRunID == "" && needsDirectorySyncRun(contracts) {
		directorySyncRunID, err = createDirectorySyncRun(ctx, db, installation)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}
	}

	shouldFinalizeDirectorySyncRun := directorySyncRunID != "" && needsDirectorySyncRun(contracts) && !options.SkipDirectorySyncRunFinalization
	if shouldFinalizeDirectorySyncRun {
		defer func() {
			if finalizeErr := finalizeDirectorySyncRun(ctx, db, directorySyncRunID, err); finalizeErr != nil && err == nil {
				err = finalizeErr
			}
		}()
	}

	for _, payloadSet := range payloadSets {
		if !contractIncludesSchema(contracts, payloadSet.Schema) {
			return ErrIngestSchemaNotFound
		}
		if _, ok := integrationgenerated.IntegrationMappingSchemas[payloadSet.Schema]; !ok {
			return ErrIngestSchemaNotFound
		}

		for _, envelope := range payloadSet.Envelopes {
			if err := ingestRecord(ctx, db, installation, definition, payloadSet.Schema, envelope, installationFilterExpr, directorySyncRunID); err != nil {
				return err
			}
		}
	}

	return nil
}

func ingestRecord(ctx context.Context, db *ent.Client, installation *ent.Integration, definition types.Definition, schema string, envelope types.MappingEnvelope, installationFilterExpr string, directorySyncRunID string) error {
	mapping, found := findMapping(definition.Mappings, schema, envelope.Variant)
	if !found {
		return ErrIngestMappingNotFound
	}

	matched, err := envelopeIncludedByFilters(ctx, installationFilterExpr, mapping.FilterExpr, envelope)
	if err != nil {
		return ErrIngestFilterFailed
	}
	if !matched {
		return nil
	}

	mapped, err := providerkit.EvalMap(ctx, mapping.MapExpr, envelope)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIngestTransformFailed, err)
	}

	switch schema {
	case integrationgenerated.IntegrationMappingSchemaDirectoryAccount:
		return upsertDirectoryAccount(ctx, db, installation, directorySyncRunID, mapped)
	case integrationgenerated.IntegrationMappingSchemaDirectoryGroup:
		return upsertDirectoryGroup(ctx, db, installation, directorySyncRunID, mapped)
	case integrationgenerated.IntegrationMappingSchemaDirectoryMembership:
		return upsertDirectoryMembership(ctx, db, installation, directorySyncRunID, mapped)
	case integrationgenerated.IntegrationMappingSchemaVulnerability:
		return upsertVulnerability(ctx, db, installation, mapped)
	default:
		return ErrIngestUnsupportedSchema
	}
}

func upsertDirectoryAccount(ctx context.Context, db *ent.Client, installation *ent.Integration, directorySyncRunID string, mapped json.RawMessage) error {
	var createInput ent.CreateDirectoryAccountInput
	if err := json.Unmarshal(mapped, &createInput); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	if createInput.OwnerID == nil && installation.OwnerID != "" {
		createInput.OwnerID = &installation.OwnerID
	}
	if createInput.IntegrationID == nil {
		createInput.IntegrationID = &installation.ID
	}
	if createInput.PlatformID == nil && installation.PlatformID != "" {
		createInput.PlatformID = &installation.PlatformID
	}
	if createInput.DirectorySyncRunID == nil && directorySyncRunID != "" {
		createInput.DirectorySyncRunID = &directorySyncRunID
	}

	if createInput.IntegrationID == nil || *createInput.IntegrationID == "" || createInput.DirectorySyncRunID == nil || *createInput.DirectorySyncRunID == "" || createInput.ExternalID == "" {
		return ErrIngestUpsertKeyMissing
	}

	existing, err := db.DirectoryAccount.Query().
		Where(
			directoryaccount.IntegrationID(*createInput.IntegrationID),
			directoryaccount.DirectorySyncRunID(*createInput.DirectorySyncRunID),
			directoryaccount.ExternalID(createInput.ExternalID),
		).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			if ent.IsNotSingular(err) {
				return ErrIngestUpsertConflict
			}

			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}

		if _, err := db.DirectoryAccount.Create().SetInput(createInput).Save(ctx); err != nil {
			return wrapIngestPersistError(err)
		}

		return nil
	}

	var updateInput ent.UpdateDirectoryAccountInput
	if err := json.Unmarshal(mapped, &updateInput); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	if _, err := db.DirectoryAccount.UpdateOneID(existing.ID).SetInput(updateInput).Save(ctx); err != nil {
		return wrapIngestPersistError(err)
	}

	return nil
}

func upsertDirectoryGroup(ctx context.Context, db *ent.Client, installation *ent.Integration, directorySyncRunID string, mapped json.RawMessage) error {
	var createInput ent.CreateDirectoryGroupInput
	if err := json.Unmarshal(mapped, &createInput); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	if createInput.OwnerID == nil && installation.OwnerID != "" {
		createInput.OwnerID = &installation.OwnerID
	}
	if createInput.IntegrationID == "" {
		createInput.IntegrationID = installation.ID
	}
	if createInput.PlatformID == nil && installation.PlatformID != "" {
		createInput.PlatformID = &installation.PlatformID
	}
	if createInput.DirectorySyncRunID == "" {
		createInput.DirectorySyncRunID = directorySyncRunID
	}

	if createInput.IntegrationID == "" || createInput.DirectorySyncRunID == "" || createInput.ExternalID == "" {
		return ErrIngestUpsertKeyMissing
	}

	existing, err := db.DirectoryGroup.Query().
		Where(
			directorygroup.IntegrationID(createInput.IntegrationID),
			directorygroup.DirectorySyncRunID(createInput.DirectorySyncRunID),
			directorygroup.ExternalID(createInput.ExternalID),
		).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			if ent.IsNotSingular(err) {
				return ErrIngestUpsertConflict
			}

			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}

		if _, err := db.DirectoryGroup.Create().SetInput(createInput).Save(ctx); err != nil {
			return wrapIngestPersistError(err)
		}

		return nil
	}

	var updateInput ent.UpdateDirectoryGroupInput
	if err := json.Unmarshal(mapped, &updateInput); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	if _, err := db.DirectoryGroup.UpdateOneID(existing.ID).SetInput(updateInput).Save(ctx); err != nil {
		return wrapIngestPersistError(err)
	}

	return nil
}

func upsertDirectoryMembership(ctx context.Context, db *ent.Client, installation *ent.Integration, directorySyncRunID string, mapped json.RawMessage) error {
	var createInput ent.CreateDirectoryMembershipInput
	if err := json.Unmarshal(mapped, &createInput); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	if createInput.OwnerID == nil && installation.OwnerID != "" {
		createInput.OwnerID = &installation.OwnerID
	}
	if createInput.IntegrationID == "" {
		createInput.IntegrationID = installation.ID
	}
	if createInput.PlatformID == nil && installation.PlatformID != "" {
		createInput.PlatformID = &installation.PlatformID
	}
	if createInput.DirectorySyncRunID == "" {
		createInput.DirectorySyncRunID = directorySyncRunID
	}

	if createInput.IntegrationID == "" || createInput.DirectorySyncRunID == "" || createInput.DirectoryAccountID == "" || createInput.DirectoryGroupID == "" {
		return ErrIngestUpsertKeyMissing
	}

	accountSnapshot, err := db.DirectoryAccount.Query().
		Where(
			directoryaccount.IntegrationID(createInput.IntegrationID),
			directoryaccount.DirectorySyncRunID(createInput.DirectorySyncRunID),
			directoryaccount.ExternalID(createInput.DirectoryAccountID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("%w: directory account snapshot not found", ErrIngestPersistFailed)
		}
		if ent.IsNotSingular(err) {
			return ErrIngestUpsertConflict
		}

		return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
	}

	groupSnapshot, err := db.DirectoryGroup.Query().
		Where(
			directorygroup.IntegrationID(createInput.IntegrationID),
			directorygroup.DirectorySyncRunID(createInput.DirectorySyncRunID),
			directorygroup.ExternalID(createInput.DirectoryGroupID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("%w: directory group snapshot not found", ErrIngestPersistFailed)
		}
		if ent.IsNotSingular(err) {
			return ErrIngestUpsertConflict
		}

		return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
	}

	createInput.DirectoryAccountID = accountSnapshot.ID
	createInput.DirectoryGroupID = groupSnapshot.ID

	existing, err := db.DirectoryMembership.Query().
		Where(
			directorymembership.DirectoryAccountID(createInput.DirectoryAccountID),
			directorymembership.DirectoryGroupID(createInput.DirectoryGroupID),
			directorymembership.DirectorySyncRunID(createInput.DirectorySyncRunID),
		).
		Only(ctx)
	if err != nil {
		if !ent.IsNotFound(err) {
			if ent.IsNotSingular(err) {
				return ErrIngestUpsertConflict
			}

			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}

		if _, err := db.DirectoryMembership.Create().SetInput(createInput).Save(ctx); err != nil {
			return wrapIngestPersistError(err)
		}

		return nil
	}

	var updateInput ent.UpdateDirectoryMembershipInput
	if err := json.Unmarshal(mapped, &updateInput); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	if _, err := db.DirectoryMembership.UpdateOneID(existing.ID).SetInput(updateInput).Save(ctx); err != nil {
		return wrapIngestPersistError(err)
	}

	return nil
}

func upsertVulnerability(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped json.RawMessage) error {
	var createInput ent.CreateVulnerabilityInput
	if err := json.Unmarshal(mapped, &createInput); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	if createInput.OwnerID == nil && installation.OwnerID != "" {
		createInput.OwnerID = &installation.OwnerID
	}
	if len(createInput.IntegrationIDs) == 0 {
		createInput.IntegrationIDs = []string{installation.ID}
	}

	if createInput.ExternalID == "" && (createInput.CveID == nil || *createInput.CveID == "") {
		return ErrIngestUpsertKeyMissing
	}

	var (
		ownerID    string
		byExternal *ent.Vulnerability
		byCVE      *ent.Vulnerability
		err        error
	)
	if createInput.OwnerID != nil {
		ownerID = *createInput.OwnerID
	}

	if createInput.ExternalID != "" {
		byExternal, err = db.Vulnerability.Query().
			Where(
				vulnerability.OwnerID(ownerID),
				vulnerability.ExternalID(createInput.ExternalID),
			).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			if ent.IsNotSingular(err) {
				return ErrIngestUpsertConflict
			}

			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}
		if ent.IsNotFound(err) {
			byExternal = nil
		}
	}

	if createInput.CveID != nil && *createInput.CveID != "" {
		byCVE, err = db.Vulnerability.Query().
			Where(
				vulnerability.OwnerID(ownerID),
				vulnerability.CveID(*createInput.CveID),
			).
			Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			if ent.IsNotSingular(err) {
				return ErrIngestUpsertConflict
			}

			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}
		if ent.IsNotFound(err) {
			byCVE = nil
		}
	}

	if byExternal != nil && byCVE != nil && byExternal.ID != byCVE.ID {
		return ErrIngestUpsertConflict
	}

	existing := byExternal
	if existing == nil {
		existing = byCVE
	}

	if existing == nil {
		if _, err := db.Vulnerability.Create().SetInput(createInput).Save(ctx); err != nil {
			return wrapIngestPersistError(err)
		}

		return nil
	}

	var updateInput ent.UpdateVulnerabilityInput
	if err := json.Unmarshal(mapped, &updateInput); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	if _, err := db.Vulnerability.UpdateOneID(existing.ID).SetInput(updateInput).Save(ctx); err != nil {
		return wrapIngestPersistError(err)
	}

	return nil
}

func resolveInstallationFilterExpr(installation *ent.Integration) (string, error) {
	if installation == nil {
		return "", nil
	}

	var cfg installationFilterConfig
	if err := jsonx.UnmarshalIfPresent(installation.Config.ClientConfig, &cfg); err != nil {
		return "", err
	}

	return cfg.FilterExpr, nil
}

func envelopeIncludedByFilters(ctx context.Context, installationFilterExpr string, mappingFilterExpr string, envelope types.MappingEnvelope) (bool, error) {
	matched, err := providerkit.EvalFilter(ctx, installationFilterExpr, envelope)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, nil
	}

	return providerkit.EvalFilter(ctx, mappingFilterExpr, envelope)
}

func findMapping(mappings []types.MappingRegistration, schema string, variant string) (types.MappingOverride, bool) {
	for _, mapping := range mappings {
		if mapping.Schema == schema && mapping.Variant == variant {
			return mapping.Spec, true
		}
	}

	return types.MappingOverride{}, false
}

func contractIncludesSchema(contracts []types.IngestContract, schema string) bool {
	for _, contract := range contracts {
		if contract.Schema == schema {
			return true
		}
	}

	return false
}

func needsDirectorySyncRun(contracts []types.IngestContract) bool {
	for _, contract := range contracts {
		switch contract.Schema {
		case integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
			integrationgenerated.IntegrationMappingSchemaDirectoryMembership:
			return true
		}
	}

	return false
}

func createDirectorySyncRun(ctx context.Context, db *ent.Client, installation *ent.Integration) (string, error) {
	create := db.DirectorySyncRun.Create().
		SetIntegrationID(installation.ID).
		SetStatus(enums.DirectorySyncRunStatusRunning)

	if installation.OwnerID != "" {
		create.SetOwnerID(installation.OwnerID)
	}
	if installation.PlatformID != "" {
		create.SetPlatformID(installation.PlatformID)
	}

	run, err := create.Save(ctx)
	if err != nil {
		return "", err
	}

	return run.ID, nil
}

func finalizeDirectorySyncRun(ctx context.Context, db *ent.Client, directorySyncRunID string, ingestErr error) error {
	update := db.DirectorySyncRun.UpdateOneID(directorySyncRunID).
		SetCompletedAt(time.Now())

	if ingestErr != nil {
		update.SetStatus(enums.DirectorySyncRunStatusFailed)
		update.SetError(ingestErr.Error())
	} else {
		update.SetStatus(enums.DirectorySyncRunStatusCompleted)
		update.ClearError()
	}

	return update.Exec(ctx)
}

func wrapIngestPersistError(err error) error {
	if err == nil {
		return nil
	}

	var validationErr *ent.ValidationError
	if errors.As(err, &validationErr) {
		if strings.Contains(validationErr.Error(), "missing required field") || strings.Contains(validationErr.Error(), "missing required edge") {
			return fmt.Errorf("%w: %w", ErrIngestRequiredKeyMissing, err)
		}

		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	}

	if ent.IsConstraintError(err) {
		return fmt.Errorf("%w: %w", ErrIngestUpsertConflict, err)
	}

	return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
}
