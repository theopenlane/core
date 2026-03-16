package operations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

type ingestState struct {
	directorySyncRunID               string
	directoryRecords                 int
	skipDirectorySyncRunFinalization bool
}

// IngestOptions controls ingest-time behavior that depends on the caller context
type IngestOptions struct {
	// DirectorySyncRunID forces directory ingest to reuse an existing sync run instead of creating a new one
	DirectorySyncRunID string
	// SkipDirectorySyncRunFinalization leaves the referenced directory sync run open after ingest completes
	SkipDirectorySyncRunFinalization bool
}

// ProcessIngest resolves declared ingest contracts, applies CEL mappings, and upserts normalized records
func ProcessIngest(ctx context.Context, reg *registry.Registry, db *ent.Client, installation *ent.Integration, operation types.OperationRegistration, response json.RawMessage) (retErr error) {
	if len(operation.Ingest) == 0 {
		return nil
	}

	var payloadSets []types.IngestPayloadSet
	if err := jsonx.RoundTrip(response, &payloadSets); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("operation", operation.Name).Msg("failed to decode ingest payloads")
		return ErrIngestPayloadsInvalid
	}

	return ProcessPayloadSets(ctx, reg, db, installation, operation.Ingest, payloadSets)
}

// ProcessPayloadSets applies declared ingest contracts, CEL mappings, and upserts normalized records.
func ProcessPayloadSets(ctx context.Context, reg *registry.Registry, db *ent.Client, installation *ent.Integration, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet) (retErr error) {
	return ProcessPayloadSetsWithOptions(ctx, reg, db, installation, contracts, payloadSets, IngestOptions{})
}

// ProcessPayloadSetsWithOptions applies declared ingest contracts, CEL mappings, and upserts normalized records with caller-supplied ingest options
func ProcessPayloadSetsWithOptions(ctx context.Context, reg *registry.Registry, db *ent.Client, installation *ent.Integration, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) (retErr error) {
	if len(contracts) == 0 {
		return nil
	}

	definition, ok := reg.Definition(installation.DefinitionID)
	if !ok {
		return ErrIngestDefinitionNotFound
	}

	state := ingestState{
		directorySyncRunID:               options.DirectorySyncRunID,
		skipDirectorySyncRunFinalization: options.SkipDirectorySyncRunFinalization,
	}
	defer finalizeIngestState(ctx, db, installation, &state, &retErr)
	logger := logx.FromContext(ctx).With().Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Logger()

	for _, contract := range contracts {
		ingestSchema, ok := integrationgenerated.IntegrationIngestSchemas[contract.Schema]
		if !ok {
			return ErrIngestSchemaNotFound
		}

		payloadSet, found := lo.Find(payloadSets, func(item types.IngestPayloadSet) bool {
			return item.Schema == contract.Schema
		})
		if !found {
			continue
		}

		mappingSchema, ok := integrationgenerated.IntegrationMappingSchemas[ingestSchema.Name]
		if !ok {
			return ErrIngestSchemaNotFound
		}

		for _, envelope := range payloadSet.Envelopes {
			envelopeLogger := logger.With().Str("schema", ingestSchema.Name).Str("variant", envelope.Variant).Logger()
			mapping, found := lo.Find(definition.Mappings, func(item types.MappingRegistration) bool {
				return item.Schema == contract.Schema && item.Variant == envelope.Variant
			})
			if !found {
				mapping, found = lo.Find(definition.Mappings, func(item types.MappingRegistration) bool {
					return item.Schema == contract.Schema && item.Variant == ""
				})
			}
			if !found {
				envelopeLogger.Error().Msg("ingest mapping not found")
				return ErrIngestMappingNotFound
			}

			passed, err := providerkit.EvalFilter(ctx, mapping.Spec.FilterExpr, envelope)
			if err != nil {
				envelopeLogger.Error().Err(err).Msg("ingest filter evaluation failed")
				return ErrIngestFilterFailed
			}
			if !passed {
				continue
			}

			mappedRaw, err := providerkit.EvalMap(ctx, mapping.Spec.MapExpr, envelope)
			if err != nil {
				envelopeLogger.Error().Err(err).Msg("ingest map evaluation failed")
				return ErrIngestTransformFailed
			}

			mappedDocument, err := jsonx.ToMap(mappedRaw)
			if err != nil {
				envelopeLogger.Error().Err(err).Msg("ingest mapped document decode failed")
				return ErrIngestMappedDocumentInvalid
			}

			if contract.EnsurePayloads {
				if _, ok := mappingSchema.AllowedKeys["rawPayload"]; ok {
					if _, exists := mappedDocument["rawPayload"]; !exists {
						mappedDocument["rawPayload"] = jsonx.DecodeAnyOrNil(envelope.Payload)
					}
				}
			}

			switch ingestSchema.Name {
			case integrationgenerated.IntegrationMappingSchemaDirectoryAccount:
				err = prepareDirectoryAccountMappedDocument(ctx, db, installation, mappedDocument, envelope, &state)
			case integrationgenerated.IntegrationMappingSchemaDirectoryGroup:
				err = prepareDirectoryGroupMappedDocument(ctx, db, installation, mappedDocument, envelope, &state)
			case integrationgenerated.IntegrationMappingSchemaDirectoryMembership:
				err = prepareDirectoryMembershipMappedDocument(ctx, db, installation, mappedDocument, envelope, &state)
			default:
				err = nil
			}
			if err != nil {
				envelopeLogger.Error().Err(err).Msg("ingest mapped document preparation failed")
				return err
			}

			if err := validateMappedDocument(mappingSchema, mappedDocument); err != nil {
				envelopeLogger.Error().Err(err).Msg("ingest mapped document validation failed")
				return err
			}

			switch ingestSchema.Name {
			case integrationgenerated.IntegrationMappingSchemaDirectoryAccount:
				err = upsertDirectoryAccount(ctx, db, installation, mappedDocument)
			case integrationgenerated.IntegrationMappingSchemaDirectoryGroup:
				err = upsertDirectoryGroup(ctx, db, installation, mappedDocument)
			case integrationgenerated.IntegrationMappingSchemaDirectoryMembership:
				err = upsertDirectoryMembership(ctx, db, installation, mappedDocument)
			case integrationgenerated.IntegrationMappingSchemaVulnerability:
				err = upsertVulnerability(ctx, db, installation, mappedDocument)
			default:
				err = ErrIngestUnsupportedSchema
			}
			if err != nil {
				return err
			}

			switch ingestSchema.Name {
			case integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
				integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
				integrationgenerated.IntegrationMappingSchemaDirectoryMembership:
				state.directoryRecords++
			}
		}
	}

	return nil
}

func finalizeIngestState(ctx context.Context, db *ent.Client, installation *ent.Integration, state *ingestState, retErr *error) {
	if state == nil || state.directorySyncRunID == "" || state.skipDirectorySyncRunFinalization {
		return
	}

	update := db.DirectorySyncRun.UpdateOneID(state.directorySyncRunID).SetCompletedAt(time.Now().UTC())
	if retErr != nil && *retErr != nil {
		if err := update.SetStatus(enums.DirectorySyncRunStatusFailed).SetError((*retErr).Error()).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("directory_sync_run_id", state.directorySyncRunID).Msg("failed to finalize directory sync run after ingest failure")
		}

		return
	}

	if err := update.SetStatus(enums.DirectorySyncRunStatusCompleted).SetFullCount(state.directoryRecords).SetDeltaCount(state.directoryRecords).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("directory_sync_run_id", state.directorySyncRunID).Msg("failed to finalize directory sync run after ingest success")
	}
}

func prepareDirectoryAccountMappedDocument(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any, envelope types.MappingEnvelope, state *ingestState) error {
	syncRunID, now, err := prepareDirectoryBase(ctx, db, installation, mapped, state, integrationgenerated.IntegrationMappingDirectoryAccountIntegrationID)
	if err != nil {
		return err
	}

	setDefaultMappedStrings(mapped,
		integrationgenerated.IntegrationMappingDirectoryAccountDirectorySyncRunID, syncRunID,
		integrationgenerated.IntegrationMappingDirectoryAccountStatus, enums.DirectoryAccountStatusActive.String(),
		integrationgenerated.IntegrationMappingDirectoryAccountMfaState, enums.DirectoryAccountMFAStateUnknown.String(),
		integrationgenerated.IntegrationMappingDirectoryAccountAccountType, enums.DirectoryAccountTypeUser.String(),
		integrationgenerated.IntegrationMappingDirectoryAccountProfileHash, payloadHash(envelope.Payload),
	)
	setDefaultMappedTime(mapped, integrationgenerated.IntegrationMappingDirectoryAccountObservedAt, now)
	setDefaultMappedProfile(mapped, integrationgenerated.IntegrationMappingDirectoryAccountProfile, envelope.Payload)
	normalizeMappedString(mapped, integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail, strings.ToLower)

	return nil
}

func prepareDirectoryGroupMappedDocument(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any, envelope types.MappingEnvelope, state *ingestState) error {
	syncRunID, now, err := prepareDirectoryBase(ctx, db, installation, mapped, state, integrationgenerated.IntegrationMappingDirectoryGroupIntegrationID)
	if err != nil {
		return err
	}

	setDefaultMappedStrings(mapped,
		integrationgenerated.IntegrationMappingDirectoryGroupDirectorySyncRunID, syncRunID,
		integrationgenerated.IntegrationMappingDirectoryGroupClassification, enums.DirectoryGroupClassificationTeam.String(),
		integrationgenerated.IntegrationMappingDirectoryGroupStatus, enums.DirectoryGroupStatusActive.String(),
		integrationgenerated.IntegrationMappingDirectoryGroupProfileHash, payloadHash(envelope.Payload),
	)
	setDefaultMappedTime(mapped, integrationgenerated.IntegrationMappingDirectoryGroupObservedAt, now)
	setDefaultMappedProfile(mapped, integrationgenerated.IntegrationMappingDirectoryGroupProfile, envelope.Payload)
	normalizeMappedString(mapped, integrationgenerated.IntegrationMappingDirectoryGroupEmail, strings.ToLower)

	return nil
}

func prepareDirectoryMembershipMappedDocument(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any, envelope types.MappingEnvelope, state *ingestState) error {
	syncRunID, now, err := prepareDirectoryBase(ctx, db, installation, mapped, state, integrationgenerated.IntegrationMappingDirectoryMembershipIntegrationID)
	if err != nil {
		return err
	}

	setDefaultMappedStrings(mapped,
		integrationgenerated.IntegrationMappingDirectoryMembershipDirectorySyncRunID, syncRunID,
		integrationgenerated.IntegrationMappingDirectoryMembershipLastConfirmedRunID, syncRunID,
		integrationgenerated.IntegrationMappingDirectoryMembershipRole, enums.DirectoryMembershipRoleMember.String(),
		integrationgenerated.IntegrationMappingDirectoryMembershipSource, installation.DefinitionSlug,
	)
	setDefaultMappedTime(mapped, integrationgenerated.IntegrationMappingDirectoryMembershipObservedAt, now)
	setDefaultMappedProfile(mapped, integrationgenerated.IntegrationMappingDirectoryMembershipMetadata, envelope.Payload)

	return resolveMembershipReferences(ctx, db, installation, mapped)
}

func prepareDirectoryBase(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any, state *ingestState, integrationKey string) (string, time.Time, error) {
	injectDirectoryContext(mapped, installation)
	setDefaultMappedString(mapped, integrationKey, installation.ID)

	syncRunID, err := ensureDirectorySyncRunID(ctx, db, installation, state)
	if err != nil {
		return "", time.Time{}, err
	}

	return syncRunID, time.Now().UTC(), nil
}

func injectDirectoryContext(mapped map[string]any, installation *ent.Integration) {
	setDefaultMappedStrings(mapped,
		"environmentID", installation.EnvironmentID,
		"environmentName", installation.EnvironmentName,
		"scopeID", installation.ScopeID,
		"scopeName", installation.ScopeName,
		"platformID", installation.PlatformID,
	)
}

func setDefaultMappedStrings(mapped map[string]any, values ...string) {
	for i := 0; i+1 < len(values); i += 2 {
		setDefaultMappedString(mapped, values[i], values[i+1])
	}
}

func setDefaultMappedString(mapped map[string]any, key string, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}

	current, ok := mapped[key]
	if ok {
		if stringValue, ok := current.(string); ok && strings.TrimSpace(stringValue) != "" {
			return
		}
		if current != nil {
			return
		}
	}

	mapped[key] = value
}

func setDefaultMappedTime(mapped map[string]any, key string, value time.Time) {
	if _, ok := mapped[key]; ok && mapped[key] != nil {
		return
	}

	mapped[key] = value
}

func setDefaultMappedProfile(mapped map[string]any, key string, rawPayload json.RawMessage) {
	if _, ok := mapped[key]; ok && mapped[key] != nil {
		return
	}

	mapped[key] = jsonx.DecodeAnyOrNil(rawPayload)
}

func normalizeMappedString(mapped map[string]any, key string, normalize func(string) string) {
	value, ok := mapped[key].(string)
	if !ok || normalize == nil {
		return
	}

	mapped[key] = normalize(strings.TrimSpace(value))
}

func payloadHash(rawPayload json.RawMessage) string {
	sum := sha256.Sum256(rawPayload)
	return hex.EncodeToString(sum[:])
}

func ensureDirectorySyncRunID(ctx context.Context, db *ent.Client, installation *ent.Integration, state *ingestState) (string, error) {
	if state != nil && state.directorySyncRunID != "" {
		return state.directorySyncRunID, nil
	}

	run, err := db.DirectorySyncRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		SetStatus(enums.DirectorySyncRunStatusRunning).
		Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Msg("failed to create directory sync run for ingest")
		return "", ErrIngestPersistFailed
	}

	if state != nil {
		state.directorySyncRunID = run.ID
	}

	return run.ID, nil
}

func resolveMembershipReferences(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any) error {
	accountRef, _ := mapped[integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID].(string)
	accountID, err := resolveDirectoryAccountReference(ctx, db, installation, accountRef)
	if err != nil {
		return err
	}

	groupRef, _ := mapped[integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID].(string)
	groupID, err := resolveDirectoryGroupReference(ctx, db, installation, groupRef)
	if err != nil {
		return err
	}

	mapped[integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID] = accountID
	mapped[integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID] = groupID

	return nil
}

func resolveDirectoryAccountReference(ctx context.Context, db *ent.Client, installation *ent.Integration, value string) (string, error) {
	return resolveDirectoryReference(ctx, installation, value, "directory_account_ref", "failed to resolve directory account reference for ingest", func(ctx context.Context, value string) (string, error) {
		account, err := db.DirectoryAccount.Query().Where(
			directoryaccount.OwnerIDEQ(installation.OwnerID),
			directoryaccount.Or(
				directoryaccount.IDEQ(value),
				directoryaccount.And(directoryaccount.IntegrationIDEQ(installation.ID), directoryaccount.ExternalIDEQ(value)),
				directoryaccount.CanonicalEmailEQ(strings.ToLower(value)),
			),
		).Only(ctx)
		if err != nil {
			return "", err
		}

		return account.ID, nil
	})
}

func resolveDirectoryGroupReference(ctx context.Context, db *ent.Client, installation *ent.Integration, value string) (string, error) {
	return resolveDirectoryReference(ctx, installation, value, "directory_group_ref", "failed to resolve directory group reference for ingest", func(ctx context.Context, value string) (string, error) {
		group, err := db.DirectoryGroup.Query().Where(
			directorygroup.OwnerIDEQ(installation.OwnerID),
			directorygroup.Or(
				directorygroup.IDEQ(value),
				directorygroup.And(directorygroup.IntegrationIDEQ(installation.ID), directorygroup.ExternalIDEQ(value)),
				directorygroup.EmailEQ(strings.ToLower(value)),
			),
		).Only(ctx)
		if err != nil {
			return "", err
		}

		return group.ID, nil
	})
}

func resolveDirectoryReference(ctx context.Context, installation *ent.Integration, value string, logKey string, message string, resolve func(context.Context, string) (string, error)) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrIngestRequiredKeyMissing
	}

	id, err := resolve(ctx, value)
	if err == nil {
		return id, nil
	}
	if ent.IsNotFound(err) {
		return "", ErrIngestMappedDocumentInvalid
	}
	if ent.IsNotSingular(err) {
		return "", ErrIngestUpsertConflict
	}

	logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str(logKey, value).Msg(message)
	return "", ErrIngestPersistFailed
}

func validateMappedDocument(schema integrationgenerated.IntegrationMappingSchema, mapped map[string]any) error {
	for key := range mapped {
		if _, ok := schema.AllowedKeys[key]; !ok {
			return ErrIngestMappedDocumentInvalid
		}
	}

	for _, key := range schema.RequiredKeys {
		value, ok := mapped[key]
		if !ok || value == nil {
			return ErrIngestRequiredKeyMissing
		}
		if stringValue, ok := value.(string); ok && stringValue == "" {
			return ErrIngestRequiredKeyMissing
		}
	}

	return nil
}

func executeMappedUpsert[CreateInput any, UpdateInput any](
	ctx context.Context,
	installation *ent.Integration,
	mapped map[string]any,
	entity string,
	recordIDKey string,
	prepareInputs func(*CreateInput, *UpdateInput),
	queryExisting func(context.Context) (string, error),
	createRecord func(context.Context, CreateInput) error,
	beforeUpdate func(context.Context, string, *UpdateInput) error,
	updateRecord func(context.Context, string, UpdateInput) error,
) error {
	logger := logx.FromContext(ctx).With().Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("entity", entity).Logger()
	createInput, err := integrationgenerated.DecodeInput[CreateInput](mapped)
	if err != nil {
		logger.Error().Err(err).Msg("failed to decode create input")
		return ErrIngestMappedDocumentInvalid
	}

	updateInput, err := integrationgenerated.DecodeInput[UpdateInput](mapped)
	if err != nil {
		logger.Error().Err(err).Msg("failed to decode update input")
		return ErrIngestMappedDocumentInvalid
	}

	if prepareInputs != nil {
		prepareInputs(&createInput, &updateInput)
	}

	recordID, err := queryExisting(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			if err := createRecord(ctx, createInput); err != nil {
				logger.Error().Err(err).Msg("failed to create from ingest")
				return ErrIngestPersistFailed
			}
			return nil
		}
		if ent.IsNotSingular(err) {
			logger.Error().Err(err).Msg("ingest upsert keys matched multiple records")
			return ErrIngestUpsertConflict
		}

		logger.Error().Err(err).Msg("failed to load for ingest upsert")
		return ErrIngestPersistFailed
	}

	if beforeUpdate != nil {
		if err := beforeUpdate(ctx, recordID, &updateInput); err != nil {
			return err
		}
	}

	if err := updateRecord(ctx, recordID, updateInput); err != nil {
		logger.Error().Err(err).Str(recordIDKey, recordID).Msg("failed to update from ingest")
		return ErrIngestPersistFailed
	}

	return nil
}

func upsertDirectoryAccount(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any) error {
	predicates := directoryAccountUpsertPredicates(installation, mapped)
	if len(predicates) == 0 {
		return ErrIngestUpsertKeyMissing
	}

	return executeMappedUpsert(
		ctx, installation, mapped, "directory account", "directory_account_id",
		func(createInput *ent.CreateDirectoryAccountInput, updateInput *ent.UpdateDirectoryAccountInput) {
			ownerID := installation.OwnerID
			createInput.OwnerID = &ownerID
			updateInput.OwnerID = &ownerID
			if createInput.IntegrationID == nil || *createInput.IntegrationID == "" {
				createInput.IntegrationID = &installation.ID
			}
		},
		func(ctx context.Context) (string, error) {
			record, err := db.DirectoryAccount.Query().Where(directoryaccount.OwnerIDEQ(installation.OwnerID), directoryaccount.Or(predicates...)).Only(ctx)
			if err != nil {
				return "", err
			}
			return record.ID, nil
		},
		func(ctx context.Context, input ent.CreateDirectoryAccountInput) error {
			_, err := db.DirectoryAccount.Create().SetInput(input).Save(ctx)
			return err
		},
		nil,
		func(ctx context.Context, recordID string, input ent.UpdateDirectoryAccountInput) error {
			return db.DirectoryAccount.UpdateOneID(recordID).SetInput(input).Exec(ctx)
		},
	)
}

func upsertDirectoryGroup(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any) error {
	predicates := directoryGroupUpsertPredicates(installation, mapped)
	if len(predicates) == 0 {
		return ErrIngestUpsertKeyMissing
	}

	return executeMappedUpsert(
		ctx, installation, mapped, "directory group", "directory_group_id",
		func(createInput *ent.CreateDirectoryGroupInput, updateInput *ent.UpdateDirectoryGroupInput) {
			ownerID := installation.OwnerID
			createInput.OwnerID = &ownerID
			updateInput.OwnerID = &ownerID
			if createInput.IntegrationID == "" {
				createInput.IntegrationID = installation.ID
			}
		},
		func(ctx context.Context) (string, error) {
			record, err := db.DirectoryGroup.Query().Where(directorygroup.OwnerIDEQ(installation.OwnerID), directorygroup.Or(predicates...)).Only(ctx)
			if err != nil {
				return "", err
			}
			return record.ID, nil
		},
		func(ctx context.Context, input ent.CreateDirectoryGroupInput) error {
			_, err := db.DirectoryGroup.Create().SetInput(input).Save(ctx)
			return err
		},
		nil,
		func(ctx context.Context, recordID string, input ent.UpdateDirectoryGroupInput) error {
			return db.DirectoryGroup.UpdateOneID(recordID).SetInput(input).Exec(ctx)
		},
	)
}

func upsertDirectoryMembership(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any) error {
	predicates := directoryMembershipUpsertPredicates(installation, mapped)
	if len(predicates) == 0 {
		return ErrIngestUpsertKeyMissing
	}

	return executeMappedUpsert(
		ctx, installation, mapped, "directory membership", "directory_membership_id",
		func(createInput *ent.CreateDirectoryMembershipInput, updateInput *ent.UpdateDirectoryMembershipInput) {
			ownerID := installation.OwnerID
			createInput.OwnerID = &ownerID
			updateInput.OwnerID = &ownerID
			if createInput.IntegrationID == "" {
				createInput.IntegrationID = installation.ID
			}
		},
		func(ctx context.Context) (string, error) {
			record, err := db.DirectoryMembership.Query().Where(directorymembership.OwnerIDEQ(installation.OwnerID), directorymembership.Or(predicates...)).Only(ctx)
			if err != nil {
				return "", err
			}
			return record.ID, nil
		},
		func(ctx context.Context, input ent.CreateDirectoryMembershipInput) error {
			_, err := db.DirectoryMembership.Create().SetInput(input).Save(ctx)
			return err
		},
		nil,
		func(ctx context.Context, recordID string, input ent.UpdateDirectoryMembershipInput) error {
			return db.DirectoryMembership.UpdateOneID(recordID).SetInput(input).Exec(ctx)
		},
	)
}

func upsertVulnerability(ctx context.Context, db *ent.Client, installation *ent.Integration, mapped map[string]any) error {
	predicates := vulnerabilityUpsertPredicates(mapped)
	if len(predicates) == 0 {
		return ErrIngestUpsertKeyMissing
	}

	return executeMappedUpsert(
		ctx, installation, mapped, "vulnerability", "vulnerability_id",
		func(createInput *ent.CreateVulnerabilityInput, _ *ent.UpdateVulnerabilityInput) {
			ownerID := installation.OwnerID
			createInput.OwnerID = &ownerID
			if !lo.Contains(createInput.IntegrationIDs, installation.ID) {
				createInput.IntegrationIDs = append(createInput.IntegrationIDs, installation.ID)
			}
		},
		func(ctx context.Context) (string, error) {
			record, err := db.Vulnerability.Query().Where(vulnerability.OwnerIDEQ(installation.OwnerID), vulnerability.Or(predicates...)).Only(ctx)
			if err != nil {
				return "", err
			}
			return record.ID, nil
		},
		func(ctx context.Context, input ent.CreateVulnerabilityInput) error {
			_, err := db.Vulnerability.Create().SetInput(input).Save(ctx)
			return err
		},
		func(ctx context.Context, recordID string, updateInput *ent.UpdateVulnerabilityInput) error {
			linked, err := db.Vulnerability.Query().Where(vulnerability.IDEQ(recordID), vulnerability.HasIntegrationsWith(integration.IDEQ(installation.ID))).Exist(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("integration_id", installation.ID).Str("definition_id", installation.DefinitionID).Str("vulnerability_id", recordID).Msg("failed to load vulnerability integration linkage")
				return ErrIngestPersistFailed
			}

			if !linked && !lo.Contains(updateInput.AddIntegrationIDs, installation.ID) {
				updateInput.AddIntegrationIDs = append(updateInput.AddIntegrationIDs, installation.ID)
			}
			return nil
		},
		func(ctx context.Context, recordID string, input ent.UpdateVulnerabilityInput) error {
			return db.Vulnerability.UpdateOneID(recordID).SetInput(input).Exec(ctx)
		},
	)
}

func directoryAccountUpsertPredicates(installation *ent.Integration, mapped map[string]any) []predicate.DirectoryAccount {
	predicates := make([]predicate.DirectoryAccount, 0, 2)

	externalID, _ := mapped[integrationgenerated.IntegrationMappingDirectoryAccountExternalID].(string)
	if externalID != "" {
		predicates = append(predicates, directoryaccount.And(
			directoryaccount.IntegrationIDEQ(installation.ID),
			directoryaccount.ExternalIDEQ(externalID),
		))
	}

	canonicalEmail, _ := mapped[integrationgenerated.IntegrationMappingDirectoryAccountCanonicalEmail].(string)
	if canonicalEmail != "" {
		predicates = append(predicates, directoryaccount.CanonicalEmailEQ(strings.ToLower(canonicalEmail)))
	}

	return predicates
}

func directoryGroupUpsertPredicates(installation *ent.Integration, mapped map[string]any) []predicate.DirectoryGroup {
	externalID, _ := mapped[integrationgenerated.IntegrationMappingDirectoryGroupExternalID].(string)
	if externalID == "" {
		return nil
	}

	return []predicate.DirectoryGroup{
		directorygroup.And(
			directorygroup.IntegrationIDEQ(installation.ID),
			directorygroup.ExternalIDEQ(externalID),
		),
	}
}

func directoryMembershipUpsertPredicates(installation *ent.Integration, mapped map[string]any) []predicate.DirectoryMembership {
	accountID, _ := mapped[integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryAccountID].(string)
	groupID, _ := mapped[integrationgenerated.IntegrationMappingDirectoryMembershipDirectoryGroupID].(string)
	syncRunID, _ := mapped[integrationgenerated.IntegrationMappingDirectoryMembershipDirectorySyncRunID].(string)
	if accountID == "" || groupID == "" || syncRunID == "" {
		return nil
	}

	return []predicate.DirectoryMembership{
		directorymembership.And(
			directorymembership.IntegrationIDEQ(installation.ID),
			directorymembership.DirectoryAccountIDEQ(accountID),
			directorymembership.DirectoryGroupIDEQ(groupID),
			directorymembership.DirectorySyncRunIDEQ(syncRunID),
		),
	}
}

func vulnerabilityUpsertPredicates(mapped map[string]any) []predicate.Vulnerability {
	predicates := make([]predicate.Vulnerability, 0, 2)

	externalID, _ := mapped[integrationgenerated.IntegrationMappingVulnerabilityExternalID].(string)
	if externalID != "" {
		predicates = append(predicates, vulnerability.ExternalIDEQ(externalID))
	}

	cveID, _ := mapped[integrationgenerated.IntegrationMappingVulnerabilityCveID].(string)
	if cveID != "" {
		predicates = append(predicates, vulnerability.CveIDEQ(cveID))
	}

	return predicates
}
