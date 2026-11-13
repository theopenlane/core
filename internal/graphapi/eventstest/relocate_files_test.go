//go:build test

package eventstest_test

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/httpserve/serveropts"
	"github.com/theopenlane/core/v2/internal/objects/relocate"
	coreutils "github.com/theopenlane/core/v2/internal/testutils"
	mock_shared "github.com/theopenlane/core/v2/pkg/objects/mocks"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

const (
	relocateTargetBucket = "sippy"
	relocateTargetRegion = "auto"
	relocateSippySource  = "opln"
	relocateOldS3Bucket  = "openlane-prod"
	relocateOldR2Bucket  = "ol-trust-center"
	relocateSize         = int64(14)
	relocateETag         = "5d41402abc4b2a76b9719d911017c592"
	relocateMultipart    = "5d41402abc4b2a76b9719d911017c592-2"
	relocateContentType  = "text/plain"
	relocateKeepMD5      = "recorded-md5"
	relocateKeepContent  = "application/recorded"
	relocateKeepSize     = int64(3)
	relocateKeepFileExt  = ".recorded"
	relocateMissingToken = "missing"
	relocateMultiToken   = "multipart"
	relocateKeyPrefix    = "relocate/"
)

// relocatePlan moves records onto R2, pulling S3 records through Sippy
var relocatePlan = relocate.Plan{Target: storage.R2Provider, PullFrom: storage.S3Provider}

// relocateProvider adds touch, stat and copy to the mock provider and records the buckets used
type relocateProvider struct {
	*mock_shared.MockProvider

	mu          sync.Mutex
	touched     []string
	copied      map[string]string
	transferred []string
}

// Touch records the bucket the object was read through
func (p *relocateProvider) Touch(_ context.Context, f *storagetypes.File) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.touched = append(p.touched, f.Bucket)

	return nil
}

// Stat reports fixed metadata, a multipart etag for keys that ask for one
func (p *relocateProvider) Stat(_ context.Context, f *storagetypes.File) (*storagetypes.ObjectInfo, error) {
	etag := relocateETag
	if strings.Contains(f.Key, relocateMultiToken) {
		etag = relocateMultipart
	}

	return &storagetypes.ObjectInfo{Size: relocateSize, ETag: etag, ContentType: relocateContentType}, nil
}

// Open streams fixed content for transfers
func (p *relocateProvider) Open(_ context.Context, _ *storagetypes.File) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("hello, world!\n")), nil
}

// seededTransfers returns the transferred keys this test seeded; other tests leave referenced rows the backfill moves as well
func (p *relocateProvider) seededTransfers() []string {
	return lo.Filter(p.transferred, func(key string, _ int) bool { return strings.HasPrefix(key, relocateKeyPrefix) })
}

// Copy records the source bucket per destination bucket and key
func (p *relocateProvider) Copy(_ context.Context, f *storagetypes.File, opts *storage.CopyOptions) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.copied[opts.Bucket+"/"+opts.Key] = f.Bucket

	return nil
}

// seedFile creates a live file record and removes it after the test
func seedFile(t *testing.T, ctx context.Context, name, provider, bucket string, modify func(*generated.FileCreate)) *generated.File {
	t.Helper()

	key := relocateKeyPrefix + ulids.New().String() + "-" + name
	// required by the schema, empty on the production rows the backfill fills
	create := suite.Client.DB.File.Create().
		SetProvidedFileName(name).
		SetProvidedFileExtension("").
		SetDetectedContentType("").
		SetStorageProvider(provider).
		SetStoragePath(key).
		SetStorageVolume(bucket).
		SetURI(provider + "://" + bucket + "/" + key)

	if modify != nil {
		modify(create)
	}

	f, err := create.Save(ctx)
	assert.NilError(t, err)

	t.Cleanup(func() {
		if err := suite.Client.DB.File.DeleteOneID(f.ID).Exec(entx.SkipSoftDelete(ctx)); err != nil && !generated.IsNotFound(err) {
			t.Error(err)
		}
	})

	return f
}

// attachContact references the file from a contact removed after the test
func attachContact(t *testing.T, ctx context.Context, fileID string) *generated.Contact {
	t.Helper()

	c, err := suite.Client.DB.Contact.Create().SetFullName("relocate " + fileID).AddFileIDs(fileID).Save(ctx)
	assert.NilError(t, err)

	t.Cleanup(func() {
		if err := suite.Client.DB.Contact.DeleteOneID(c.ID).Exec(entx.SkipSoftDelete(ctx)); err != nil && !generated.IsNotFound(err) {
			t.Error(err)
		}
	})

	return c
}

func TestRelocateFiles(t *testing.T) {
	// rows are seeded by an organization member, the backfill runs as the gala routine does
	user := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	seedCtx := th.SetInternalContext(user.UserCtx, suite.Client.DB)
	ctx := auth.WithCaller(context.Background(), serveropts.BackfillCaller())

	var provider *relocateProvider

	buckets := map[storage.ProviderType]string{storage.R2Provider: relocateTargetBucket, storage.S3Provider: relocateSippySource}

	regions := map[storage.ProviderType]string{storage.R2Provider: relocateTargetRegion}

	svc, mockProvider, err := coreutils.MockStorageServiceWithBuckets(t, buckets, regions, func(m *mock_shared.MockProvider) storage.Provider {
		provider = &relocateProvider{MockProvider: m, copied: map[string]string{}}

		return provider
	})
	assert.NilError(t, err)

	mockProvider.EXPECT().ProviderType().Return(storagetypes.R2Provider).Maybe()
	mockProvider.EXPECT().
		ListObjects(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, prefix string, _ int) ([]string, error) {
			if strings.Contains(prefix, relocateMissingToken) {
				return nil, nil
			}

			return []string{prefix}, nil
		}).
		Maybe()
	mockProvider.EXPECT().
		Upload(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, _ io.Reader, opts *storage.UploadOptions) (*storagetypes.UploadedFileMetadata, error) {
			provider.mu.Lock()
			defer provider.mu.Unlock()

			provider.transferred = append(provider.transferred, opts.FileName)

			return &storagetypes.UploadedFileMetadata{FileMetadata: storagetypes.FileMetadata{Key: opts.FileName}}, nil
		}).
		Maybe()

	s3 := storage.S3Provider.String()
	r2 := storage.R2Provider.String()

	orphan := seedFile(t, seedCtx, "orphan.txt", s3, relocateSippySource, nil)
	orphanOldBucket := seedFile(t, seedCtx, "orphan-old.txt", s3, relocateOldS3Bucket, nil)
	orphanOldR2 := seedFile(t, seedCtx, "orphan-old-r2.txt", r2, relocateOldR2Bucket, nil)

	softParent := seedFile(t, seedCtx, "soft-parent.txt", s3, relocateSippySource, nil)
	parent := attachContact(t, seedCtx, softParent.ID)
	assert.NilError(t, suite.Client.DB.Contact.DeleteOneID(parent.ID).Exec(seedCtx))

	recorded := seedFile(t, seedCtx, "recorded.pdf", s3, relocateSippySource, func(c *generated.FileCreate) {
		c.SetMd5Hash(relocateKeepMD5).
			SetDetectedContentType(relocateKeepContent).
			SetPersistedFileSize(relocateKeepSize).
			SetProvidedFileExtension(relocateKeepFileExt)
	})
	attachContact(t, seedCtx, recorded.ID)

	legacy := seedFile(t, seedCtx, "legacy.txt", s3, relocateSippySource, func(c *generated.FileCreate) {
		c.SetStorageProvider("")
	})
	attachContact(t, seedCtx, legacy.ID)

	multipart := seedFile(t, seedCtx, relocateMultiToken+".bin", s3, relocateSippySource, nil)
	attachContact(t, seedCtx, multipart.ID)

	oldBucket := seedFile(t, seedCtx, "old-bucket.txt", s3, relocateOldS3Bucket, nil)
	attachContact(t, seedCtx, oldBucket.ID)

	notCopied := seedFile(t, seedCtx, relocateMissingToken+".txt", s3, relocateSippySource, nil)
	attachContact(t, seedCtx, notCopied.ID)

	oldR2 := seedFile(t, seedCtx, "old-r2.txt", r2, relocateOldR2Bucket, nil)
	attachContact(t, seedCtx, oldR2.ID)

	landed := seedFile(t, seedCtx, "landed.txt", r2, relocateTargetBucket, nil)
	attachContact(t, seedCtx, landed.ID)

	load := func(id string) *generated.File {
		f, err := suite.Client.DB.File.Query().Where(file.ID(id)).Only(entx.SkipSoftDelete(seedCtx))
		assert.NilError(t, err)

		return f
	}

	// swapped after seeding so the restore runs before the seeded rows are cleaned up
	original := suite.Client.DB.ObjectManager
	suite.Client.DB.ObjectManager = svc

	t.Cleanup(func() {
		suite.Client.DB.ObjectManager = original
	})

	assert.NilError(t, relocate.Files(ctx, suite.Client.DB, relocatePlan))

	t.Run("orphans are left untouched for the orphan sweep", func(t *testing.T) {
		for _, seeded := range []*generated.File{orphan, orphanOldBucket, orphanOldR2} {
			f := load(seeded.ID)

			assert.Equal(t, f.StorageProvider, seeded.StorageProvider)
			assert.Equal(t, f.StorageVolume, seeded.StorageVolume)
			assert.Equal(t, f.URI, seeded.URI)
		}
	})

	t.Run("s3 file with a soft deleted parent is read through the target bucket and relocated", func(t *testing.T) {
		f := load(softParent.ID)

		assert.Equal(t, f.StorageProvider, r2)
		assert.Equal(t, f.StorageScheme, "r2://")
		assert.Equal(t, f.StorageVolume, relocateTargetBucket)
		assert.Equal(t, f.StorageRegion, relocateTargetRegion)
		assert.Equal(t, f.URI, "r2://"+relocateTargetBucket+"/"+softParent.StoragePath)
		assert.Equal(t, f.StoragePath, softParent.StoragePath)
		assert.Equal(t, f.Md5Hash, relocateETag)
		assert.Equal(t, f.PersistedFileSize, relocateSize)
		assert.Equal(t, f.DetectedContentType, relocateContentType)
		assert.Equal(t, f.DetectedMimeType, relocateContentType)
		assert.Equal(t, f.ProvidedFileExtension, ".txt")

		provider.mu.Lock()
		defer provider.mu.Unlock()

		for _, bucket := range provider.touched {
			assert.Equal(t, bucket, relocateTargetBucket)
		}
	})

	t.Run("recorded metadata is not overwritten", func(t *testing.T) {
		f := load(recorded.ID)

		assert.Equal(t, f.StorageProvider, r2)
		assert.Equal(t, f.Md5Hash, relocateKeepMD5)
		assert.Equal(t, f.DetectedContentType, relocateKeepContent)
		assert.Equal(t, f.PersistedFileSize, relocateKeepSize)
		assert.Equal(t, f.ProvidedFileExtension, relocateKeepFileExt)
		assert.Equal(t, f.DetectedMimeType, relocateContentType, "empty mime type is filled")
	})

	t.Run("legacy record with no provider and an s3 uri is relocated", func(t *testing.T) {
		assert.Equal(t, load(legacy.ID).StorageProvider, r2)
	})

	t.Run("multipart etag is not recorded as an md5", func(t *testing.T) {
		f := load(multipart.ID)

		assert.Equal(t, f.StorageProvider, r2)
		assert.Equal(t, f.Md5Hash, "")
		assert.Equal(t, f.PersistedFileSize, relocateSize)
	})

	t.Run("s3 file in a bucket sippy does not read is transferred by download and upload", func(t *testing.T) {
		f := load(oldBucket.ID)

		assert.Equal(t, f.StorageProvider, r2)
		assert.Equal(t, f.StorageVolume, relocateTargetBucket)
		assert.Equal(t, f.StorageRegion, relocateTargetRegion)
		assert.Equal(t, f.URI, "r2://"+relocateTargetBucket+"/"+oldBucket.StoragePath)

		provider.mu.Lock()
		defer provider.mu.Unlock()

		assert.DeepEqual(t, provider.seededTransfers(), []string{oldBucket.StoragePath})
	})

	t.Run("object absent from the target bucket leaves the row untouched", func(t *testing.T) {
		f := load(notCopied.ID)

		assert.Equal(t, f.StorageProvider, s3)
		assert.Equal(t, f.StorageVolume, relocateSippySource)
		assert.Equal(t, f.URI, notCopied.URI)
		assert.Equal(t, f.Md5Hash, "")
	})

	t.Run("r2 file in another bucket is copied into the target bucket and repointed", func(t *testing.T) {
		f := load(oldR2.ID)

		assert.Equal(t, f.StorageProvider, r2)
		assert.Equal(t, f.StorageVolume, relocateTargetBucket)
		assert.Equal(t, f.StorageRegion, relocateTargetRegion)
		assert.Equal(t, f.URI, "r2://"+relocateTargetBucket+"/"+oldR2.StoragePath)
		assert.Equal(t, f.StoragePath, oldR2.StoragePath)
		assert.Equal(t, f.Md5Hash, relocateETag)

		provider.mu.Lock()
		defer provider.mu.Unlock()

		assert.Equal(t, provider.copied[relocateTargetBucket+"/"+oldR2.StoragePath], relocateOldR2Bucket)
		assert.Equal(t, len(provider.copied), 1, "only the r2 row outside the target bucket is copied")
	})

	t.Run("r2 file already in the target bucket is untouched", func(t *testing.T) {
		f := load(landed.ID)

		assert.Equal(t, f.URI, landed.URI)
		assert.Equal(t, f.StorageRegion, "")
		assert.Equal(t, f.Md5Hash, "")
	})

	t.Run("rerun only retries the file that has not landed", func(t *testing.T) {
		provider.mu.Lock()
		touchedBefore, copiedBefore := len(provider.touched), len(provider.copied)
		provider.mu.Unlock()

		assert.NilError(t, relocate.Files(ctx, suite.Client.DB, relocatePlan))

		provider.mu.Lock()
		defer provider.mu.Unlock()

		assert.Equal(t, len(provider.touched)-touchedBefore, 1)
		assert.Equal(t, len(provider.copied), copiedBefore)
		assert.Equal(t, len(provider.seededTransfers()), 1)
		assert.Equal(t, load(notCopied.ID).StorageProvider, s3)
		assert.Equal(t, load(softParent.ID).StorageProvider, r2)
		assert.Equal(t, load(oldR2.ID).StorageVolume, relocateTargetBucket)
	})

	t.Run("every referenced row is moved by one strategy", func(t *testing.T) {
		provider.mu.Lock()
		defer provider.mu.Unlock()

		// two runs: five pulls on the first (soft parent, recorded, legacy, multipart, not copied), one retry on the second
		assert.Equal(t, len(provider.touched), 6)
		assert.Equal(t, len(provider.copied), 1)
		assert.Equal(t, len(provider.seededTransfers()), 1)
	})
}
