package handlers_test

import (
	"bytes"
	"context"
	"net/url"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/storagetypes"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/filedownloadtoken"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	handlerpkg "github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/objects/resolver"
	"github.com/theopenlane/core/internal/objects/upload"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
)

const testTokenIssuer = "http://localhost:17608"

func (suite *HandlerTestSuite) TestDatabaseFileDownloadHandler_Success() {
	t := suite.T()

	restore := suite.swapObjectStoreToDatabase()
	t.Cleanup(restore)

	user := suite.userBuilder(context.Background())

	uploadCtx := auth.NewTestContextWithOrgID(user.ID, user.PersonalOrgID)
	uploadCtx = privacy.DecisionContext(uploadCtx, privacy.Allow)
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
	require.NoError(t, err)
	require.Len(t, uploadedFiles, 1)

	uploaded := uploadedFiles[0]
	require.NotEmpty(t, uploaded.ID)

	fileRecord, err := suite.db.File.Get(uploadCtx, uploaded.ID)
	require.NoError(t, err)

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
	require.NoError(t, err)
	require.NotEmpty(t, presignedURL)

	assert.Contains(t, presignedURL, uploaded.ID)
	assert.Contains(t, presignedURL, "token=")

	parsedURL, err := url.Parse(presignedURL)
	require.NoError(t, err)

	tokenValue := parsedURL.Query().Get("token")
	require.NotEmpty(t, tokenValue)

	tokenRecord, err := suite.db.FileDownloadToken.Query().
		Where(filedownloadtoken.Token(tokenValue)).
		Only(uploadCtx)
	require.NoError(t, err)
	require.NotNil(t, tokenRecord.FileID)
	require.Equal(t, uploaded.ID, *tokenRecord.FileID)
}

func (suite *HandlerTestSuite) TestDatabaseFileDownloadHandler_WithUserID() {
	t := suite.T()

	restore := suite.swapObjectStoreToDatabase()
	t.Cleanup(restore)

	user := suite.userBuilder(context.Background())

	uploadCtx := auth.NewTestContextWithOrgID(user.ID, user.PersonalOrgID)
	uploadCtx = privacy.DecisionContext(uploadCtx, privacy.Allow)
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
	require.NoError(t, err)
	require.Len(t, uploadedFiles, 1)

	uploaded := uploadedFiles[0]
	require.NotEmpty(t, uploaded.ID)

	fileRecord, err := suite.db.File.Get(uploadCtx, uploaded.ID)
	require.NoError(t, err)

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
	require.NoError(t, err)
	require.NotEmpty(t, presignedURL)

	assert.Contains(t, presignedURL, uploaded.ID)
	assert.Contains(t, presignedURL, "token=")

	parsedURL, err := url.Parse(presignedURL)
	require.NoError(t, err)

	tokenValue := parsedURL.Query().Get("token")
	require.NotEmpty(t, tokenValue)

	tokenRecord, err := suite.db.FileDownloadToken.Query().
		Where(filedownloadtoken.Token(tokenValue)).
		Only(uploadCtx)
	require.NoError(t, err)
	require.NotNil(t, tokenRecord.Secret)
	require.NotNil(t, tokenRecord.TTL)
	require.NotNil(t, tokenRecord.FileID)
	require.Equal(t, uploaded.ID, *tokenRecord.FileID)
	require.NotNil(t, tokenRecord.UserID)
	require.Equal(t, user.ID, *tokenRecord.UserID)
}

func (suite *HandlerTestSuite) TestDatabaseFileDownloadHandler_PresignedURLGeneration() {
	t := suite.T()

	baseURL := "https://api.theopenlane.io"

	cfg := storage.ProviderConfig{
		Enabled: true,
		Providers: storage.Providers{
			Database: storage.ProviderConfigs{Enabled: true},
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
	uploadCtx = privacy.DecisionContext(uploadCtx, privacy.Allow)
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
	require.NoError(t, err)
	require.Len(t, uploadedFiles, 1)

	uploaded := uploadedFiles[0]
	require.NotEmpty(t, uploaded.ID)

	fileRecord, err := suite.db.File.Get(uploadCtx, uploaded.ID)
	require.NoError(t, err)

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
	require.NoError(t, err)

	assert.Contains(t, presignedURL, baseURL)
	assert.Contains(t, presignedURL, uploaded.ID)
	assert.Contains(t, presignedURL, "token=")
}

func (suite *HandlerTestSuite) swapObjectStoreToDatabase() func() {
	cfg := storage.ProviderConfig{
		Enabled: true,
		Providers: storage.Providers{
			Database: storage.ProviderConfigs{Enabled: true},
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
	return resolver.NewServiceFromConfig(cfg,
		resolver.WithPresignConfig(func() *tokens.TokenManager {
			return suite.sharedTokenManager
		}, testTokenIssuer, testTokenIssuer),
		resolver.WithPresignBaseURL(baseURL),
	)
}
