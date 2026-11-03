package handlers_test

import (
	"bytes"
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/theopenlane/core/common/storagetypes"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/filedownloadtoken"
	handlerpkg "github.com/theopenlane/core/v2/internal/httpserve/handlers"
	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/internal/objects/resolver"
	"github.com/theopenlane/core/v2/internal/objects/upload"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

const testTokenIssuer = "http://localhost:17608"

func TestDatabaseFileDownloadHandler_Success(t *testing.T) {
	suite.setupTest(t)

	restore := suite.swapObjectStoreToDatabase()
	t.Cleanup(restore)

	user := suite.userBuilder(context.Background())

	uploadCtx := auth.NewTestContextWithOrgID(user.ID, user.PersonalOrgID)
	uploadCtx = auth.WithInternalOperationContext(uploadCtx)
	uploadCtx = ent.NewContext(uploadCtx, suite.db)

	fileContent := []byte("test file content for download")

	uploadCtx, uploadedFiles, err := upload.HandleUploads(uploadCtx, suite.objectStore, []pkgobjects.File{
		{
			RawFile:              bytes.NewReader(fileContent),
			OriginalName:         "testfile.txt",
			FieldName:            "uploadFile",
			CorrelatedObjectID:   user.PersonalOrgID,
			CorrelatedObjectType: "organization",
			FileMetadata: pkgobjects.FileMetadata{
				ContentType: "text/plain",
			},
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, is.Len(uploadedFiles, 1))

	uploaded := uploadedFiles[0]
	assert.Assert(t, len(uploaded.ID) > 0)

	fileRecord, err := suite.db.File.Get(uploadCtx, uploaded.ID)
	assert.NilError(t, err)

	presignedURL, err := suite.objectStore.GetPresignedURL(uploadCtx, &storagetypes.File{
		ID:           uploaded.ID,
		OriginalName: fileRecord.ProvidedFileName,
		ProviderType: storagetypes.DatabaseProvider,
		FileMetadata: storagetypes.FileMetadata{
			Key:          fileRecord.StoragePath,
			Bucket:       fileRecord.StorageVolume,
			ContentType:  fileRecord.DetectedContentType,
			ProviderType: storagetypes.DatabaseProvider,
			FullURI:      fileRecord.URI,
		},
	}, time.Minute)
	assert.NilError(t, err)
	assert.Assert(t, len(presignedURL) > 0)

	assert.Check(t, is.Contains(presignedURL, uploaded.ID))
	assert.Check(t, is.Contains(presignedURL, "token="))

	parsedURL, err := url.Parse(presignedURL)
	assert.NilError(t, err)

	tokenValue := parsedURL.Query().Get("token")
	assert.Assert(t, len(tokenValue) > 0)

	tokenRecord, err := suite.db.FileDownloadToken.Query().
		Where(filedownloadtoken.Token(tokenValue)).
		Only(uploadCtx)
	assert.NilError(t, err)
	assert.Assert(t, tokenRecord.FileID != nil)
	assert.Assert(t, is.DeepEqual(uploaded.ID, *tokenRecord.FileID))
}

func TestDatabaseFileDownloadHandler_WithUserID(t *testing.T) {
	suite.setupTest(t)

	restore := suite.swapObjectStoreToDatabase()
	t.Cleanup(restore)

	user := suite.userBuilder(context.Background())

	uploadCtx := auth.NewTestContextWithOrgID(user.ID, user.PersonalOrgID)
	uploadCtx = auth.WithInternalOperationContext(uploadCtx)
	uploadCtx = ent.NewContext(uploadCtx, suite.db)

	fileContent := []byte("user specific download test")

	uploadCtx, uploadedFiles, err := upload.HandleUploads(uploadCtx, suite.objectStore, []pkgobjects.File{
		{
			RawFile:              bytes.NewReader(fileContent),
			OriginalName:         "userfile.txt",
			FieldName:            "uploadFile",
			CorrelatedObjectID:   user.PersonalOrgID,
			CorrelatedObjectType: "organization",
			FileMetadata: pkgobjects.FileMetadata{
				ContentType: "text/plain",
			},
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, is.Len(uploadedFiles, 1))

	uploaded := uploadedFiles[0]
	assert.Assert(t, len(uploaded.ID) > 0)

	fileRecord, err := suite.db.File.Get(uploadCtx, uploaded.ID)
	assert.NilError(t, err)

	presignedURL, err := suite.objectStore.GetPresignedURL(uploadCtx, &storagetypes.File{
		ID:           uploaded.ID,
		OriginalName: fileRecord.ProvidedFileName,
		ProviderType: storagetypes.DatabaseProvider,
		FileMetadata: storagetypes.FileMetadata{
			Key:          fileRecord.StoragePath,
			Bucket:       fileRecord.StorageVolume,
			ContentType:  fileRecord.DetectedContentType,
			ProviderType: storagetypes.DatabaseProvider,
			FullURI:      fileRecord.URI,
		},
	}, time.Minute)
	assert.NilError(t, err)
	assert.Assert(t, len(presignedURL) > 0)

	assert.Check(t, is.Contains(presignedURL, uploaded.ID))
	assert.Check(t, is.Contains(presignedURL, "token="))

	parsedURL, err := url.Parse(presignedURL)
	assert.NilError(t, err)

	tokenValue := parsedURL.Query().Get("token")
	assert.Assert(t, len(tokenValue) > 0)

	tokenRecord, err := suite.db.FileDownloadToken.Query().
		Where(filedownloadtoken.Token(tokenValue)).
		Only(uploadCtx)
	assert.NilError(t, err)
	assert.Assert(t, tokenRecord.Secret != nil)
	assert.Assert(t, tokenRecord.TTL != nil)
	assert.Assert(t, tokenRecord.FileID != nil)
	assert.Assert(t, is.DeepEqual(uploaded.ID, *tokenRecord.FileID))
	assert.Assert(t, tokenRecord.UserID != nil)
	assert.Assert(t, is.DeepEqual(user.ID, *tokenRecord.UserID))
}

func TestDatabaseFileDownloadHandler_PresignedURLGeneration(t *testing.T) {
	suite.setupTest(t)

	baseURL := "https://api.theopenlane.io"

	cfg := storage.ProviderConfig{
		Enabled: true,
		Providers: storage.Providers{
			Database: storage.DatabaseConfig{ProviderCommon: storage.ProviderCommon{Enabled: true}},
		},
	}

	dbStore := suite.createDatabaseStoreWithPresignConfig(cfg, baseURL)

	originalStore := suite.objectStore
	originalHandlerStore := suite.h.ObjectStore

	suite.objectStore = dbStore
	suite.h.ObjectStore = dbStore

	t.Cleanup(func() {
		suite.objectStore = originalStore
		suite.h.ObjectStore = originalHandlerStore
	})

	user := suite.userBuilder(context.Background())

	uploadCtx := auth.NewTestContextWithOrgID(user.ID, user.PersonalOrgID)
	uploadCtx = auth.WithInternalOperationContext(uploadCtx)
	uploadCtx = ent.NewContext(uploadCtx, suite.db)

	fileContent := []byte("presigned url test")

	uploadCtx, uploadedFiles, err := upload.HandleUploads(uploadCtx, suite.objectStore, []pkgobjects.File{
		{
			RawFile:              bytes.NewReader(fileContent),
			OriginalName:         "presigned.txt",
			FieldName:            "uploadFile",
			CorrelatedObjectID:   user.PersonalOrgID,
			CorrelatedObjectType: "organization",
			FileMetadata: pkgobjects.FileMetadata{
				ContentType: "text/plain",
			},
		},
	})
	assert.NilError(t, err)
	assert.Assert(t, is.Len(uploadedFiles, 1))

	uploaded := uploadedFiles[0]
	assert.Assert(t, len(uploaded.ID) > 0)

	fileRecord, err := suite.db.File.Get(uploadCtx, uploaded.ID)
	assert.NilError(t, err)

	presignedURL, err := suite.objectStore.GetPresignedURL(uploadCtx, &storagetypes.File{
		ID:           uploaded.ID,
		OriginalName: fileRecord.ProvidedFileName,
		ProviderType: storagetypes.DatabaseProvider,
		FileMetadata: storagetypes.FileMetadata{
			Key:          fileRecord.StoragePath,
			Bucket:       fileRecord.StorageVolume,
			ContentType:  fileRecord.DetectedContentType,
			ProviderType: storagetypes.DatabaseProvider,
			FullURI:      fileRecord.URI,
		},
	}, time.Minute)
	assert.NilError(t, err)

	assert.Check(t, is.Contains(presignedURL, baseURL))
	assert.Check(t, is.Contains(presignedURL, uploaded.ID))
	assert.Check(t, is.Contains(presignedURL, "token="))
}

func (suite *HandlerTestSuite) swapObjectStoreToDatabase() func() {
	cfg := storage.ProviderConfig{
		Enabled: true,
		Providers: storage.Providers{
			Database: storage.DatabaseConfig{ProviderCommon: storage.ProviderCommon{Enabled: true}},
		},
	}

	dbStore := suite.createDatabaseStoreWithPresignConfig(cfg, testTokenIssuer)

	originalStore := suite.objectStore
	originalHandlerStore := suite.h.ObjectStore

	suite.objectStore = dbStore
	suite.h.ObjectStore = dbStore

	var originalRouterHandler *handlerpkg.Handler
	if suite.router != nil {
		originalRouterHandler = suite.router.Handler
		if suite.router.Handler == nil {
			suite.router.Handler = suite.h
		}
		suite.router.Handler.ObjectStore = dbStore
	}

	return func() {
		suite.objectStore = originalStore
		suite.h.ObjectStore = originalHandlerStore
		if suite.router != nil {
			if suite.router.Handler != nil {
				suite.router.Handler.ObjectStore = originalHandlerStore
			}
			suite.router.Handler = originalRouterHandler
		}
	}
}

func (suite *HandlerTestSuite) createDatabaseStoreWithPresignConfig(cfg storage.ProviderConfig, baseURL string) *objects.Service {
	service, err := resolver.NewServiceFromConfig(cfg,
		resolver.WithPresignConfig(func() *tokens.TokenManager {
			return suite.sharedTokenManager
		}, testTokenIssuer, testTokenIssuer),
		resolver.WithPresignBaseURL(baseURL),
	)
	assert.NilError(suite.T(), err)

	return service
}
