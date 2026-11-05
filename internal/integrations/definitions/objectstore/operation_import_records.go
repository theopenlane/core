package objectstore

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"
)

var (
	// importPolicy runs the record import per installation and as a recurring system cycle against the runtime bucket
	importPolicy = types.ExecutionPolicy{Reconcile: true, Scheduled: true}
	// importContracts lists every schema the definition has a mapping for
	importContracts = lo.Map(recordSchemas, func(schema *entityops.Schema, _ int) types.IngestContract {
		return types.IngestContract{Schema: schema.Name}
	})
	// schemaDirectories maps each import schema's directory name, its plural snake_case name, to the schema
	schemaDirectories = lo.SliceToMap(recordSchemas, func(schema *entityops.Schema) (string, *entityops.Schema) {
		return pluralize.NewClient().Plural(schema.Snake), schema
	})
)

// runImportRecords ingests the JSON objects under the configured prefix one object at a time and moves each object whose
// records imported without failure into the processed directory when the spec asks for it; the ownerless runtime path
// fills an empty spec from the runtime config, and a dataset imports its manifest instead
func runImportRecords(ctx context.Context, request types.OperationRequest, client *Client, cfg ImportRecords) (json.RawMessage, error) {
	if cfg.Manifest {
		return runImportDataset(ctx, request, client, cfg)
	}

	spec := cfg.ImportSpec

	if request.Integration == nil && spec.Prefix == "" {
		spec = client.Import
	}

	if spec.Prefix == "" {
		return nil, ErrPrefixRequired
	}

	// a system run has no organization, so its entity types are system-owned
	_, noOrg := auth.GetOrganizationIDFromContext(ctx)

	keys, err := listRecordObjects(ctx, client.Provider, spec.Prefix)
	if err != nil {
		return nil, err
	}

	objects, err := recordObjects(spec.Prefix, keys)
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

	for _, object := range objects {
		payloadSet, err := loadPayloadSet(ctx, client.Provider, object)
		if err != nil {
			return nil, err
		}

		if err := ensureEntityTypes(ctx, request.DB, noOrg != nil, recordEntityTypeNames([]types.IngestPayloadSet{payloadSet})); err != nil {
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

		if err := moveProcessed(ctx, client.Provider, object.key, processedKey(spec.Prefix, object.key)); err != nil {
			return nil, err
		}

		result.Moved++
	}

	logx.FromContext(ctx).Debug().Str("prefix", spec.Prefix).Int("objects", result.Objects).Int("attempted", result.Attempted).Int("failed", result.Failed).Int("moved", result.Moved).Msg("objectstore: record import completed")

	return providerkit.EncodeResult(result, ErrResultEncode)
}

// validateImportRecords rejects a stored import without a prefix to read
func validateImportRecords(_ context.Context, _ types.InstallationRequest, cfg *ImportRecords) error {
	if cfg.Prefix == "" {
		return ErrPrefixRequired
	}

	return nil
}

// ensureEntityTypes creates the named entity types the imported records link to when they do not exist in the scope the
// link resolver searches: system-owned rows for a system run, otherwise the caller's organization's own rows, which the
// organization interceptor scopes to and the create hooks own
func ensureEntityTypes(ctx context.Context, db *generated.Client, systemOwned bool, names []string) error {
	for _, name := range lo.Uniq(names) {
		exists, err := db.EntityType.Query().Where(entitytype.NameEqualFold(name), entitytype.SystemOwned(systemOwned)).Exist(ctx)
		if err != nil {
			return fmt.Errorf("%w: %s: %w", ErrEntityTypeSeed, name, err)
		}

		if exists {
			continue
		}

		created, err := db.EntityType.Create().SetName(name).Save(ctx)
		if err != nil {
			return fmt.Errorf("%w: %s: %w", ErrEntityTypeSeed, name, err)
		}

		logx.FromContext(ctx).Debug().Str("entity_type_id", created.ID).Bool("system_owned", created.SystemOwned).Msg("objectstore: created entity type for import link")
	}

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

// recordObject is one JSON object under the import prefix with the schema its records import as
type recordObject struct {
	// key is the object key in the bucket
	key string
	// schema is the schema the object's records import as
	schema *entityops.Schema
}

// recordObjects resolves each key's schema, failing before anything imports when a directory names none, and orders the
// objects so the targets of a schema's foreign keys import before it
func recordObjects(prefix string, keys []string) ([]recordObject, error) {
	objects := make([]recordObject, 0, len(keys))

	for _, key := range keys {
		schema, err := objectSchema(prefix, key)
		if err != nil {
			return nil, err
		}

		objects = append(objects, recordObject{key: key, schema: schema})
	}

	order := operations.DependencyOrder(recordSchemas)

	slices.SortStableFunc(objects, func(a, b recordObject) int {
		return cmp.Compare(lo.IndexOf(order, a.schema), lo.IndexOf(order, b.schema))
	})

	return objects, nil
}

// objectSchema returns the schema a key's parent directory names, an object directly under the prefix importing as the
// schema the prefix names, or Entity when the prefix names none
func objectSchema(prefix, key string) (*entityops.Schema, error) {
	root := strings.TrimSuffix(prefix, "/")
	directory := path.Dir(strings.TrimPrefix(strings.TrimPrefix(key, root), "/"))

	if directory == "." {
		schema, ok := directorySchema(path.Base(root))

		return lo.Ternary(ok, schema, entityops.SchemaEntity), nil
	}

	schema, ok := directorySchema(path.Base(directory))
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrRecordSchemaUnknown, key)
	}

	return schema, nil
}

// directorySchema returns the import schema a directory names, by its plural snake_case name
func directorySchema(name string) (*entityops.Schema, bool) {
	if name == vendorsDirectory {
		return entityops.SchemaEntity, true
	}

	schema, ok := schemaDirectories[name]

	return schema, ok
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

// loadPayloadSet reads one object and wraps its records as the ingest payload set of its schema
func loadPayloadSet(ctx context.Context, client storage.Provider, object recordObject) (types.IngestPayloadSet, error) {
	records, err := loadRecords(ctx, client, object.key)
	if err != nil {
		return types.IngestPayloadSet{}, err
	}

	envelopes := lo.Map(records, func(record json.RawMessage, idx int) types.MappingEnvelope {
		return providerkit.RawEnvelope(fmt.Sprintf("%s#%d", object.key, idx), record)
	})

	return types.IngestPayloadSet{Schema: object.schema.Name, Envelopes: envelopes}, nil
}

// loadRecords reads one object and decodes its records
func loadRecords(ctx context.Context, client storage.Provider, key string) ([]json.RawMessage, error) {
	document, err := downloadObject(ctx, client, key)
	if err != nil {
		return nil, err
	}

	records, err := decodeRecords(document)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", key).Int("bytes", len(document)).Msg("objectstore: object content is not json records")

		return nil, fmt.Errorf("%w: %s: %w", ErrObjectContentInvalid, key, err)
	}

	return records, nil
}

// downloadObject reads the bytes of one object
func downloadObject(ctx context.Context, client storage.Provider, key string) ([]byte, error) {
	downloaded, err := client.Download(ctx, &storagetypes.File{FileMetadata: storagetypes.FileMetadata{Key: key}}, &storagetypes.DownloadFileOptions{})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("key", key).Msg("objectstore: object download failed")

		return nil, fmt.Errorf("%w: %s: %w", ErrObjectDownloadFailed, key, err)
	}

	return downloaded.File, nil
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
