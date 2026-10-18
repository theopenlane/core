package objectstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

var (
	// importPolicy is the execution policy of the record import
	importPolicy = types.ExecutionPolicy{Reconcile: true}
	// importContracts lists the schemas the record import ingests
	importContracts = []types.IngestContract{{Schema: entityMapping.Schema}}
)

// runImportRecords ingests the JSON objects under the configured prefix one object at a time and moves each object whose
// records imported without failure into the processed directory when the spec asks for it; the ownerless runtime path
// fills an empty spec from the runtime config
func runImportRecords(ctx context.Context, request types.OperationRequest, client *Client, cfg ImportRecords) (json.RawMessage, error) {
	spec := cfg.ImportSpec

	if request.Integration == nil && spec.Prefix == "" {
		spec = client.Import
	}

	if spec.Prefix == "" {
		return nil, ErrPrefixRequired
	}

	if err := ensureLinkTargets(ctx, request); err != nil {
		return nil, err
	}

	keys, err := listRecordObjects(ctx, client.Provider, spec.Prefix)
	if err != nil {
		return nil, err
	}

	oc, _ := gala.OperationContextFromContext(ctx)

	ingest := operations.IngestContext{
		Registry:     request.Services.Definitions(),
		DB:           request.DB,
		Runtime:      request.Services.Gala(),
		Integration:  request.Integration,
		DefinitionID: DefinitionID.ID(),
	}
	options := operations.IngestOptionsFromOperationContext(oc)
	result := ImportRecordsResult{Objects: len(keys)}

	for _, key := range keys {
		payloadSet, err := loadPayloadSet(ctx, client.Provider, key)
		if err != nil {
			return nil, err
		}

		ingested, err := operations.ProcessPayloadSets(ctx, ingest, ImportRecordsOp.Name(), importContracts, importPolicy, []types.IngestPayloadSet{payloadSet}, options)
		result.Merge(ingested)

		if err != nil {
			return nil, err
		}

		if !spec.MoveProcessed || ingested.Failed > 0 {
			continue
		}

		if err := moveProcessed(ctx, client.Provider, key, processedKey(spec.Prefix, key)); err != nil {
			return nil, err
		}

		result.Moved++
	}

	logx.FromContext(ctx).Info().Str("prefix", spec.Prefix).Int("objects", result.Objects).Int("attempted", result.Attempted).Int("failed", result.Failed).Int("moved", result.Moved).Msg("objectstore: record import completed")

	return providerkit.EncodeResult(result, ErrResultEncode)
}

// validateImportRecords rejects a stored import without a prefix to read
func validateImportRecords(_ context.Context, _ types.InstallationRequest, cfg *ImportRecords) error {
	if cfg.Prefix == "" {
		return ErrPrefixRequired
	}

	return nil
}

// ensureLinkTargets creates the entity type the mapping links records to when it does not exist in the scope the link
// resolver searches: system-owned for the ownerless runtime import, the installation's organization otherwise
func ensureLinkTargets(ctx context.Context, request types.OperationRequest) error {
	ownerScope := entitytype.SystemOwned(true)
	create := request.DB.EntityType.Create().SetName(vendorEntityTypeName)

	if request.Integration != nil {
		ownerScope = entitytype.OwnerID(request.Integration.OwnerID)
		create = create.SetOwnerID(request.Integration.OwnerID)
	}

	exists, err := request.DB.EntityType.Query().Where(entitytype.NameEqualFold(vendorEntityTypeName), ownerScope).Exist(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrEntityTypeSeed, err)
	}

	if exists {
		return nil
	}

	created, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrEntityTypeSeed, err)
	}

	logx.FromContext(ctx).Info().Str("entity_type_id", created.ID).Bool("system_owned", created.SystemOwned).Msg("objectstore: created entity type for import link")

	return nil
}

// listRecordObjects lists the JSON objects under the prefix, leaving out the processed directory
func listRecordObjects(ctx context.Context, client storage.Provider, prefix string) ([]string, error) {
	keys, err := client.ListObjects(ctx, prefix, 0)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("prefix", prefix).Msg("objectstore: list objects failed")

		return nil, fmt.Errorf("%w: %w", ErrListObjectsFailed, err)
	}

	processed := path.Join(prefix, processedDir) + "/"

	keys = lo.Filter(keys, func(key string, _ int) bool {
		return strings.HasSuffix(key, jsonObjectSuffix) && !strings.HasPrefix(key, processed)
	})

	if len(keys) == 0 {
		logx.FromContext(ctx).Info().Str("prefix", prefix).Msg("objectstore: no json objects under prefix")
	}

	return keys, nil
}

// processedKey returns the key an imported object moves to: its path under the prefix, re-rooted under the processed directory
func processedKey(prefix, key string) string {
	return path.Join(prefix, processedDir, strings.TrimPrefix(key, prefix))
}

// moveProcessed copies the object to the destination key and deletes the original
func moveProcessed(ctx context.Context, client storage.Provider, key, destination string) error {
	copier, ok := client.(storage.Copier)
	if !ok {
		return ErrProviderCannotCopy
	}

	file := &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: key}}

	if err := copier.Copy(ctx, file, &storage.CopyOptions{Key: destination}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", key).Str("destination", destination).Msg("objectstore: processed object copy failed")

		return fmt.Errorf("%w: %s: %w", ErrObjectMoveFailed, key, err)
	}

	if err := client.Delete(ctx, file, &storagetypes.DeleteFileOptions{}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", key).Msg("objectstore: processed object delete failed")

		return fmt.Errorf("%w: %s: %w", ErrObjectMoveFailed, key, err)
	}

	return nil
}

// loadPayloadSet reads one object and wraps its records as the ingest payload set of the import schema
func loadPayloadSet(ctx context.Context, client storage.Provider, key string) (types.IngestPayloadSet, error) {
	records, err := loadRecords(ctx, client, key)
	if err != nil {
		return types.IngestPayloadSet{}, err
	}

	envelopes := lo.Map(records, func(record json.RawMessage, idx int) types.MappingEnvelope {
		return providerkit.RawEnvelope(fmt.Sprintf("%s#%d", key, idx), record)
	})

	return types.IngestPayloadSet{Schema: entityMapping.Schema, Envelopes: envelopes}, nil
}

// loadRecords reads one object and decodes its records
func loadRecords(ctx context.Context, client storage.Provider, key string) ([]json.RawMessage, error) {
	downloaded, err := client.Download(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: key}}, &storagetypes.DownloadFileOptions{})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", key).Msg("objectstore: object download failed")

		return nil, fmt.Errorf("%w: %s: %w", ErrObjectDownloadFailed, key, err)
	}

	records, err := decodeRecords(downloaded.File)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", key).Int("bytes", len(downloaded.File)).Msg("objectstore: object content is not json records")

		return nil, fmt.Errorf("%w: %s: %w", ErrObjectContentInvalid, key, err)
	}

	return records, nil
}

// decodeRecords parses a document as a JSON array of records, wrapping a single JSON object as one record
func decodeRecords(document []byte) ([]json.RawMessage, error) {
	trimmed := bytes.TrimSpace(document)

	switch {
	case bytes.HasPrefix(trimmed, []byte("{")):
		var record json.RawMessage
		if err := json.Unmarshal(trimmed, &record); err != nil {
			return nil, err
		}

		return []json.RawMessage{record}, nil
	default:
		var records []json.RawMessage
		if err := json.Unmarshal(trimmed, &records); err != nil {
			return nil, err
		}

		return records, nil
	}
}
