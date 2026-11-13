package relocate

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/common/storagetypes"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/interceptors"
	"github.com/theopenlane/core/v2/internal/objects/filerefs"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

const (
	// pageSize is the number of file records handled per page
	pageSize = 200
	// listLimit is the keys requested when confirming an object landed
	listLimit = 1
	// multipartETagMarker is present in the ETag of multipart uploads, which is not an MD5 digest
	multipartETagMarker = "-"
	// schemeSeparator splits a storage URI into its scheme and the rest
	schemeSeparator = "://"
)

// opaqueContentTypes are the content types object stores report when none was recorded
var opaqueContentTypes = []string{"application/octet-stream", "binary/octet-stream"}

// Plan names the target provider and how objects reach it
type Plan struct {
	// Target is the provider records are moved onto, in its configured bucket
	Target storage.ProviderType
	// PullFrom is a provider Target fetches from on first read, such as S3 behind R2 Sippy; empty when none
	PullFrom storage.ProviderType
}

// stats are the counts reported when a run finishes
type stats struct {
	scanned          int
	copied           int
	pulled           int
	transferred      int
	metadataFilled   int
	skippedOrphans   int
	skippedNotLanded int
	errors           int
}

// Files moves every referenced file record onto the plan's target bucket by copy, pull-through read
// or download and upload under the same key; orphans are left for the file orphan sweep
func Files(ctx context.Context, client *ent.Client, plan Plan) error {
	ctx = entx.SkipSoftDelete(ctx)

	provider, options, err := client.ObjectManager.ResolveProvider(ctx, storedFile(plan.Target, "", ""))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("provider", plan.Target.String()).Msg("relocate files: failed to resolve target provider")

		return err
	}

	bucket := options.Bucket

	r := &relocator{client: client, plan: plan, provider: provider, bucket: bucket, region: options.Region}

	if plan.PullFrom != "" {
		_, pullOptions, err := client.ObjectManager.ResolveProvider(ctx, storedFile(plan.PullFrom, "", ""))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("provider", plan.PullFrom.String()).Msg("relocate files: failed to resolve pull-through provider")

			return err
		}

		r.pullBucket = pullOptions.Bucket
	}

	var last string

	for {
		page, err := client.File.Query().
			Where(selection(plan.Target, bucket), file.IDGT(last)).
			Order(ent.Asc(file.FieldID)).
			Limit(pageSize).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("relocate files: failed to read page")

			return err
		}

		if len(page) == 0 {
			break
		}

		last = page[len(page)-1].ID
		r.stats.scanned += len(page)

		r.page(ctx, page)
	}

	s := r.stats
	logx.FromContext(ctx).Info().Str("provider", plan.Target.String()).Str("bucket", bucket).Str("region", r.region).Int("scanned", s.scanned).Int("copied", s.copied).Int("pulled", s.pulled).Int("transferred", s.transferred).Int("metadata_filled", s.metadataFilled).Int("skipped_orphans", s.skippedOrphans).Int("skipped_not_landed", s.skippedNotLanded).Int("errors", s.errors).Msg("relocate files: complete")

	return nil
}

// selection matches live records not already stored by provider in bucket
func selection(provider storage.ProviderType, bucket string) predicate.File {
	return file.And(
		file.DeletedAtIsNil(),
		file.Or(
			file.StorageProviderNEQ(provider.String()),
			file.StorageProviderIsNil(),
			file.StorageVolumeNEQ(bucket),
		),
	)
}

// storedFile describes key in bucket as an object provider operates on
func storedFile(provider storage.ProviderType, key, bucket string) *storagetypes.File {
	return &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			Key:           key,
			Bucket:        bucket,
			ProviderType:  provider,
			ProviderHints: &storage.ProviderHints{KnownProvider: provider},
		},
	}
}

// recordProvider is the provider a record is stored on, from the URI scheme for legacy records
func recordProvider(f *ent.File) (storage.ProviderType, error) {
	if f.StorageProvider != "" {
		return storage.ProviderType(f.StorageProvider), nil
	}

	scheme, _, ok := strings.Cut(f.URI, schemeSeparator)
	if !ok || scheme == "" {
		return "", ErrUnknownProvider
	}

	return storage.ProviderType(scheme), nil
}

// relocator carries the run state shared across pages
type relocator struct {
	client *ent.Client
	plan   Plan
	// provider is the target provider
	provider storage.Provider
	// bucket is the target bucket
	bucket string
	// region is the target provider's configured region
	region string
	// pullBucket is the pull-through provider's configured bucket
	pullBucket string
	stats      stats
}

// page classifies one page of records and handles each independently
func (r *relocator) page(ctx context.Context, page []*ent.File) {
	current, err := filerefs.Referenced(ctx, r.client, lo.Map(page, func(f *ent.File, _ int) string { return f.ID }))
	if err != nil {
		// nothing moves without a classification
		r.stats.errors += len(page)

		logx.FromContext(ctx).Error().Err(err).Msg("relocate files: failed to classify page")

		return
	}

	for _, f := range page {
		if err := r.row(ctx, f, lo.HasKey(current, f.ID)); err != nil {
			r.stats.errors++

			logx.FromContext(ctx).Error().Err(err).Str("file_id", f.ID).Msg("relocate files: failed to handle file")
		}
	}
}

// row picks the move for one record
func (r *relocator) row(ctx context.Context, f *ent.File, current bool) error {
	if !current {
		r.stats.skippedOrphans++

		logx.FromContext(ctx).Debug().Str("file_id", f.ID).Msg("relocate files: nothing references the file, leaving it for the orphan sweep")

		return nil
	}

	if f.StoragePath == "" {
		return ErrMissingStorageKey
	}

	source, err := recordProvider(f)
	if err != nil {
		return err
	}

	switch {
	case source == r.plan.Target:
		return r.copy(ctx, f)
	case source == r.plan.PullFrom && f.StorageVolume == r.pullBucket:
		return r.pull(ctx, f)
	default:
		return r.transfer(ctx, f, source)
	}
}

// copy has the target copy the object into the target bucket under the same key
func (r *relocator) copy(ctx context.Context, f *ent.File) error {
	copier, ok := r.provider.(storage.Copier)
	if !ok {
		return ErrCopyUnsupported
	}

	if err := copier.Copy(ctx, interceptors.StorageFileFromEnt(f), &storage.CopyOptions{Bucket: r.bucket, Key: f.StoragePath}); err != nil {
		return err
	}

	return r.land(ctx, f, &r.stats.copied)
}

// pull reads the object through the target so it fetches from the pull-through provider
func (r *relocator) pull(ctx context.Context, f *ent.File) error {
	toucher, ok := r.provider.(storage.Toucher)
	if !ok {
		return ErrTouchUnsupported
	}

	if err := toucher.Touch(ctx, storedFile(r.plan.Target, f.StoragePath, r.bucket)); err != nil {
		return err
	}

	return r.land(ctx, f, &r.stats.pulled)
}

// transfer streams the object from its provider into the target under the same key
func (r *relocator) transfer(ctx context.Context, f *ent.File, source storage.ProviderType) error {
	from := interceptors.StorageFileFromEnt(f)
	from.ProviderType = source
	from.ProviderHints.KnownProvider = source

	sourceProvider, _, err := r.client.ObjectManager.ResolveProvider(ctx, from)
	if err != nil {
		return err
	}

	opener, ok := sourceProvider.(storage.Opener)
	if !ok {
		return ErrOpenUnsupported
	}

	// the body streams into the upload, which spools anything over objects.MaxInMemorySize to disk
	body, err := opener.Open(ctx, from)
	if err != nil {
		return err
	}

	defer body.Close()

	uploaded, err := r.provider.Upload(ctx, body, &storage.UploadOptions{FileName: f.StoragePath, ContentType: f.DetectedContentType})
	if err != nil {
		return err
	}

	if uploaded.Key != "" && uploaded.Key != f.StoragePath {
		return ErrKeyChanged
	}

	return r.land(ctx, f, &r.stats.transferred)
}

// land confirms the object is in the target bucket, repoints the record, fills missing metadata and bumps counter
func (r *relocator) land(ctx context.Context, f *ent.File, counter *int) error {
	keys, err := r.provider.ListObjects(ctx, f.StoragePath, listLimit)
	if err != nil {
		return err
	}

	if !lo.Contains(keys, f.StoragePath) {
		r.stats.skippedNotLanded++

		logx.FromContext(ctx).Debug().Str("file_id", f.ID).Str("key", f.StoragePath).Msg("relocate files: object not present in target bucket, retrying next run")

		return nil
	}

	target := r.plan.Target.String()
	scheme := storage.Scheme(r.plan.Target)

	// a record already pointing at the bucket is not counted twice
	update := r.client.File.Update().
		Where(file.ID(f.ID), file.Or(file.StorageProviderNEQ(target), file.StorageProviderIsNil(), file.StorageVolumeNEQ(r.bucket))).
		SetStorageProvider(target).
		SetStorageScheme(scheme).
		SetStorageVolume(r.bucket).
		SetStorageRegion(r.region).
		SetURI(scheme + r.bucket + "/" + f.StoragePath)

	var filled bool

	if statter, ok := r.provider.(storage.Statter); ok {
		info, err := statter.Stat(ctx, storedFile(r.plan.Target, f.StoragePath, r.bucket))
		if err != nil {
			return err
		}

		filled = fillMetadata(ctx, update, f, info)
	}

	n, err := update.Save(ctx)
	if err != nil || n == 0 {
		return err
	}

	*counter++

	if filled {
		r.stats.metadataFilled++
	}

	logx.FromContext(ctx).Debug().Str("file_id", f.ID).Str("from", f.StorageVolume).Str("bucket", r.bucket).Str("region", r.region).Bool("metadata_filled", filled).Msg("relocate files: landed")

	return nil
}

// fillMetadata sets the metadata fields the record never had and reports whether any was set
func fillMetadata(ctx context.Context, update *ent.FileUpdate, f *ent.File, info *storagetypes.ObjectInfo) bool {
	var filled bool

	if f.Md5Hash == "" && info.ETag != "" {
		switch {
		case strings.Contains(info.ETag, multipartETagMarker):
			logx.FromContext(ctx).Debug().Str("file_id", f.ID).Str("etag", info.ETag).Msg("relocate files: multipart etag is not an md5, leaving hash empty")
		default:
			update.SetMd5Hash(info.ETag)

			filled = true
		}
	}

	if info.ContentType != "" && !lo.Contains(opaqueContentTypes, info.ContentType) {
		if f.DetectedContentType == "" {
			update.SetDetectedContentType(info.ContentType)

			filled = true
		}

		if f.DetectedMimeType == "" {
			update.SetDetectedMimeType(info.ContentType)

			filled = true
		}
	}

	if ext := filepath.Ext(f.ProvidedFileName); f.ProvidedFileExtension == "" && ext != "" {
		update.SetProvidedFileExtension(ext)

		filled = true
	}

	if f.PersistedFileSize == 0 && info.Size > 0 {
		update.SetPersistedFileSize(info.Size)

		filled = true
	}

	return filled
}
