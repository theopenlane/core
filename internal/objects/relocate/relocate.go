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
	"github.com/theopenlane/core/v2/internal/ent/generated/actionplan"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/procedure"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterentity"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterwatermarkconfig"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/internal/ent/interceptors"
	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

const (
	// pageSize is the number of file records handled per page
	pageSize = 200
	// listLimit is the number of keys requested when confirming an object exists in R2
	listLimit = 1
	// multipartETagMarker is present in the ETag of multipart uploads, which is not an MD5 digest
	multipartETagMarker = "-"
)

// opaqueContentTypes are the content types object stores report when none was recorded
var opaqueContentTypes = []string{"application/octet-stream", "binary/octet-stream"}

// stats are the counts reported when a run finishes
type stats struct {
	scanned          int
	relocated        int
	metadataFilled   int
	purgedOrphans    int
	purgedOldBucket  int
	skippedNotCopied int
	errors           int
}

// referenceLookup returns the subset of ids referenced by one foreign key that points at files
type referenceLookup func(ctx context.Context, client *ent.Client, ids []string) ([]string, error)

// referenceLookups cover every inbound foreign key to files that has no edge predicate on the
// file schema
var referenceLookups = []referenceLookup{
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.ActionPlan.Query().Where(actionplan.FileIDIn(ids...)).Select(actionplan.FieldFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.DirectoryAccount.Query().Where(directoryaccount.AvatarLocalFileIDIn(ids...)).Select(directoryaccount.FieldAvatarLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Entity.Query().Where(entity.LogoFileIDIn(ids...)).Select(entity.FieldLogoFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.InternalPolicy.Query().Where(internalpolicy.FileIDIn(ids...)).Select(internalpolicy.FieldFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Organization.Query().Where(organization.AvatarLocalFileIDIn(ids...)).Select(organization.FieldAvatarLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Procedure.Query().Where(procedure.FileIDIn(ids...)).Select(procedure.FieldFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Standard.Query().Where(standard.LogoFileIDIn(ids...)).Select(standard.FieldLogoFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Subprocessor.Query().Where(subprocessor.LogoFileIDIn(ids...)).Select(subprocessor.FieldLogoFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterEntity.Query().Where(trustcenterentity.LogoFileIDIn(ids...)).Select(trustcenterentity.FieldLogoFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterNDARequest.Query().Where(trustcenterndarequest.FileIDIn(ids...)).Select(trustcenterndarequest.FieldFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterSetting.Query().Where(trustcentersetting.FaviconLocalFileIDIn(ids...)).Select(trustcentersetting.FieldFaviconLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterSetting.Query().Where(trustcentersetting.HeroImageLocalFileIDIn(ids...)).Select(trustcentersetting.FieldHeroImageLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterSetting.Query().Where(trustcentersetting.LogoLocalFileIDIn(ids...)).Select(trustcentersetting.FieldLogoLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterWatermarkConfig.Query().Where(trustcenterwatermarkconfig.LogoIDIn(ids...)).Select(trustcenterwatermarkconfig.FieldLogoID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.User.Query().Where(user.AvatarLocalFileIDIn(ids...)).Select(user.FieldAvatarLocalFileID).Strings(ctx)
	},
}

// S3Files moves every live file record that points at S3 onto R2. Records referenced by a parent
// are relocated when they live in olpn and purged when they live in any other
// bucket or have no parent
func S3Files(ctx context.Context, client *ent.Client) error {
	ctx = entx.SkipSoftDelete(ctx)

	r := &relocator{client: client, objects: client.ObjectManager}

	var last string

	for {
		page, err := client.File.Query().
			Where(selection(), file.IDGT(last)).
			Order(ent.Asc(file.FieldID)).
			Limit(pageSize).
			All(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("relocate s3 files: failed to read page")

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
	logx.FromContext(ctx).Info().Int("scanned", s.scanned).Int("relocated", s.relocated).Int("metadata_filled", s.metadataFilled).Int("purged_orphans", s.purgedOrphans).Int("purged_old_bucket", s.purgedOldBucket).Int("skipped_not_copied", s.skippedNotCopied).Int("errors", s.errors).Msg("relocate s3 files: complete")

	return nil
}

// selection matches live records stored in S3, including the legacy records that carry no
// provider and are identified by an s3 URI
func selection() predicate.File {
	s3 := storage.S3Provider.String()

	return file.And(
		file.DeletedAtIsNil(),
		file.Or(
			file.StorageProviderEQ(s3),
			file.And(
				file.Or(file.StorageProviderIsNil(), file.StorageProviderEQ("")),
				file.URIHasPrefix(storage.Scheme(storage.S3Provider)),
			),
		),
	)
}

// relocator carries the run state shared across pages
type relocator struct {
	client  *ent.Client
	objects *objects.Service
	stats   stats
}

// page classifies one page of records and handles each independently
func (r *relocator) page(ctx context.Context, page []*ent.File) {
	current, err := currentIDs(ctx, r.client, lo.Map(page, func(f *ent.File, _ int) string { return f.ID }))
	if err != nil {
		// without a trustworthy classification nothing on the page may be purged
		r.stats.errors += len(page)

		logx.FromContext(ctx).Error().Err(err).Msg("relocate s3 files: failed to classify page")

		return
	}

	for _, f := range page {
		if err := r.row(ctx, f, lo.HasKey(current, f.ID)); err != nil {
			r.stats.errors++

			logx.FromContext(ctx).Error().Err(err).Str("file_id", f.ID).Msg("relocate s3 files: failed to handle file")
		}
	}
}

// row handles one record according to its classification
func (r *relocator) row(ctx context.Context, f *ent.File, current bool) error {
	switch {
	case !current:
		logx.FromContext(ctx).Debug().Str("file_id", f.ID).Msg("relocate s3 files: purging orphan")

		if err := r.purge(ctx, f); err != nil {
			return err
		}

		r.stats.purgedOrphans++

		return nil
	case f.StorageVolume != "olpn":
		logx.FromContext(ctx).Info().Str("file_id", f.ID).Str("bucket", f.StorageVolume).Msg("relocate s3 files: purging file in unmigrated bucket")

		if err := r.purge(ctx, f); err != nil {
			return err
		}

		r.stats.purgedOldBucket++

		return nil
	default:
		return r.relocate(ctx, f)
	}
}

// purge hard deletes the record and leaves the object it points at untouched
func (r *relocator) purge(ctx context.Context, f *ent.File) error {
	return r.client.File.DeleteOneID(f.ID).Exec(objects.WithSkipStorageCleanup(ctx))
}

// relocate pulls the object through R2, which makes Sippy copy it from S3, confirms it landed and
// then repoints the record at R2 while filling metadata that was never recorded
func (r *relocator) relocate(ctx context.Context, f *ent.File) error {
	if f.StoragePath == "" {
		return ErrMissingStorageKey
	}

	source := interceptors.StorageFileFromEnt(f)
	source.ProviderType = storage.R2Provider
	source.FileMetadata.ProviderType = storage.R2Provider
	source.ProviderHints = &storage.ProviderHints{KnownProvider: storage.R2Provider}

	provider, bucket, err := r.objects.ResolveProvider(ctx, source)
	if err != nil {
		return err
	}

	toucher, ok := provider.(storage.Toucher)
	if !ok {
		return ErrTouchUnsupported
	}

	if err := toucher.Touch(ctx, source); err != nil {
		return err
	}

	keys, err := provider.ListObjects(ctx, f.StoragePath, listLimit)
	if err != nil {
		return err
	}

	if !lo.Contains(keys, f.StoragePath) {
		r.stats.skippedNotCopied++

		logx.FromContext(ctx).Debug().Str("file_id", f.ID).Str("key", f.StoragePath).Msg("relocate s3 files: object not present in r2 after read, retrying next run")

		return nil
	}

	statter, ok := provider.(storage.Statter)
	if !ok {
		return ErrStatUnsupported
	}

	info, err := statter.Stat(ctx, source)
	if err != nil {
		return err
	}

	scheme := storage.Scheme(storage.R2Provider)

	update := r.client.File.Update().
		Where(file.ID(f.ID), file.Or(file.StorageProviderEQ(storage.S3Provider.String()), file.StorageProviderIsNil(), file.StorageProviderEQ(""))).
		SetStorageProvider(storage.R2Provider.String()).
		SetStorageScheme(scheme).
		SetStorageVolume(bucket).
		SetURI(scheme + bucket + "/" + f.StoragePath)

	filled := fillMetadata(ctx, update, f, info)

	n, err := update.Save(ctx)
	if err != nil {
		return err
	}

	if n == 0 {
		return nil
	}

	r.stats.relocated++

	if filled {
		r.stats.metadataFilled++
	}

	logx.FromContext(ctx).Debug().Str("file_id", f.ID).Str("bucket", bucket).Bool("metadata_filled", filled).Msg("relocate s3 files: relocated")

	return nil
}

// fillMetadata sets on the update every metadata field the record has never had a value for,
// leaving recorded values alone, and reports whether any was set
func fillMetadata(ctx context.Context, update *ent.FileUpdate, f *ent.File, info *storagetypes.ObjectInfo) bool {
	var filled bool

	if f.Md5Hash == "" && info.ETag != "" {
		switch {
		case strings.Contains(info.ETag, multipartETagMarker):
			logx.FromContext(ctx).Debug().Str("file_id", f.ID).Str("etag", info.ETag).Msg("relocate s3 files: multipart etag is not an md5, leaving hash empty")
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

// currentIDs returns the ids of records with at least one inbound reference, soft deleted
// parents included
func currentIDs(ctx context.Context, client *ent.Client, ids []string) (map[string]struct{}, error) {
	edged, err := client.File.Query().
		Where(file.IDIn(ids...), file.Or(
			file.HasGroups(),
			file.HasContact(),
			file.HasEntity(),
			file.HasOrganizationSetting(),
			file.HasTemplate(),
			file.HasDocument(),
			file.HasProgram(),
			file.HasPlatform(),
			file.HasEvidence(),
			file.HasIdentityHolder(),
			file.HasScan(),
			file.HasEvents(),
			file.HasIntegrations(),
			file.HasSecrets(),
			file.HasTrustCenterEntities(),
			file.HasTrustCenterDoc(),
			file.HasOriginalTrustCenterDoc(),
		)).
		IDs(ctx)
	if err != nil {
		return nil, err
	}

	current := lo.Keyify(edged)

	for _, lookup := range referenceLookups {
		referenced, err := lookup(ctx, client, ids)
		if err != nil {
			return nil, err
		}

		for _, id := range referenced {
			current[id] = struct{}{}
		}
	}

	return current, nil
}
