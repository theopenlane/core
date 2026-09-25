package objectstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// IngestHandle adapts the installation record import to the ingest operation registration boundary
func (ImportRecords) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(storageClient, func(ctx context.Context, request types.OperationRequest, client *Client) ([]types.IngestPayloadSet, error) {
		cfg, err := resolveImportConfig(request.Config, client.Import)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Int("config_bytes", len(request.Config)).Msg("objectstore: import records config could not be resolved")

			return nil, err
		}

		if err := cfg.ensureLinkTargets(ctx, request); err != nil {
			return nil, err
		}

		return cfg.Run(ctx, client.Provider)
	})
}

// IngestHandle adapts the triggered system import to the ingest operation registration boundary
func (SystemImport) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(storageClient, SystemImportOp, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *Client, cfg SystemImport) ([]types.IngestPayloadSet, error) {
		spec := cfg.spec(client.Import)

		if err := spec.ensureLinkTargets(ctx, request); err != nil {
			return nil, err
		}

		return spec.Run(ctx, client.Provider)
	})
}

// ensureLinkTargets creates the entity type the vendor mapping variant links records to when it does not exist in the
// scope the link resolver searches: system-owned for the ownerless runtime import, the installation's organization otherwise
func (i ImportRecords) ensureLinkTargets(ctx context.Context, request types.OperationRequest) error {
	if i.Variant != variantVendor {
		return nil
	}

	ownerScope := entitytype.SystemOwned(true)
	create := request.DB.EntityType.Create().SetName(variantVendor)

	if request.Integration != nil {
		ownerScope = entitytype.OwnerID(request.Integration.OwnerID)
		create = create.SetOwnerID(request.Integration.OwnerID)
	}

	exists, err := request.DB.EntityType.Query().Where(entitytype.NameEqualFold(variantVendor), ownerScope).Exist(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrVendorEntityTypeSeed, err)
	}

	if exists {
		return nil
	}

	created, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrVendorEntityTypeSeed, err)
	}

	logx.FromContext(ctx).Info().Str("entity_type_id", created.ID).Bool("system_owned", created.SystemOwned).Msg("objectstore: created vendor entity type for import link")

	return nil
}

// resolveImportConfig decodes the operation config, falling back to the import spec carried on the client
func resolveImportConfig(config json.RawMessage, configured ImportRecords) (ImportRecords, error) {
	cfg, err := importRecordsOperation.UnmarshalConfig(config)
	if err != nil {
		return ImportRecords{}, fmt.Errorf("%w: %w", ErrOperationConfigInvalid, err)
	}

	if jsonx.IsEmptyRawMessage(config) {
		cfg = configured
	}

	if cfg.Prefix == "" {
		return ImportRecords{}, ErrPrefixRequired
	}

	return cfg, nil
}

// Run lists the JSON objects under the prefix and returns their records as one ingest payload set
func (i ImportRecords) Run(ctx context.Context, client storage.Provider) ([]types.IngestPayloadSet, error) {
	if !lo.Contains(importSchemas, i.Schema) {
		logx.FromContext(ctx).Error().Str("schema", i.Schema).Strs("supported", importSchemas).Msg("objectstore: import schema unsupported")

		return nil, ErrSchemaUnsupported
	}

	keys, err := client.ListObjects(ctx, i.Prefix, 0)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("prefix", i.Prefix).Msg("objectstore: list objects failed")

		return nil, fmt.Errorf("%w: %w", ErrListObjectsFailed, err)
	}

	keys = lo.Filter(keys, func(key string, _ int) bool {
		return strings.HasSuffix(key, jsonObjectSuffix)
	})

	if len(keys) == 0 {
		logx.FromContext(ctx).Info().Str("prefix", i.Prefix).Msg("objectstore: no json objects under prefix")

		return nil, nil
	}

	var envelopes []types.MappingEnvelope

	for _, key := range keys {
		records, err := loadRecords(ctx, client, key)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, lo.Map(records, func(record json.RawMessage, idx int) types.MappingEnvelope {
			return providerkit.RawEnvelopeVariant(i.Variant, fmt.Sprintf("%s#%d", key, idx), record)
		})...)
	}

	logx.FromContext(ctx).Info().Str("prefix", i.Prefix).Str("schema", i.Schema).Str("variant", i.Variant).Int("objects", len(keys)).Int("records", len(envelopes)).Msg("objectstore: bucket records loaded for ingest")

	return []types.IngestPayloadSet{
		{
			Schema:    i.Schema,
			Envelopes: envelopes,
		},
	}, nil
}

// loadRecords downloads one object and decodes its records
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
