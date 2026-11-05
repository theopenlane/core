package handlers

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/storagetypes"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestDownload_SetDownloadTokens_Success(t *testing.T) {
	token := "test-token-value"
	secret := []byte("test-secret-bytes")
	ttl := time.Now().UTC().Add(1 * time.Hour)

	tokenRecord := &ent.FileDownloadToken{
		Token:  token,
		Secret: &secret,
		TTL:    &ttl,
	}

	d := &Download{
		Token: token,
	}

	err := d.setDownloadTokens(tokenRecord, token)
	assert.NilError(t, err)

	assert.Check(t, is.DeepEqual(token, d.DownloadToken.Token.String))
	assert.Check(t, d.DownloadToken.Token.Valid)
	assert.Check(t, is.DeepEqual(secret, d.Secret))
	assert.Check(t, d.Expires.Valid)
	assert.Check(t, is.DeepEqual(ttl.Format(time.RFC3339Nano), d.Expires.String))
}

func TestDownload_SetDownloadTokens_TokenMismatch(t *testing.T) {
	storedToken := "stored-token"
	requestToken := "different-token"

	secret := []byte("test-secret")

	tokenRecord := &ent.FileDownloadToken{
		Token:  storedToken,
		Secret: &secret,
	}

	d := &Download{
		Token: requestToken,
	}

	err := d.setDownloadTokens(tokenRecord, requestToken)
	assert.Assert(t, err != nil)
	assert.Check(t, is.ErrorIs(err, ErrNotFound))
}

func TestDownload_SetDownloadTokens_NilToken(t *testing.T) {
	secret := []byte("secret")

	tokenRecord := &ent.FileDownloadToken{
		Token:  "",
		Secret: &secret,
	}

	d := &Download{
		Token: "some-token",
	}

	err := d.setDownloadTokens(tokenRecord, "some-token")
	assert.Assert(t, err != nil)
	assert.Check(t, is.ErrorIs(err, ErrNotFound))
}

func TestDownload_SetDownloadTokens_NoExpiration(t *testing.T) {
	token := "test-token"
	secret := []byte("secret")

	tokenRecord := &ent.FileDownloadToken{
		Token:  token,
		Secret: &secret,
		TTL:    nil,
	}

	d := &Download{
		Token: token,
	}

	err := d.setDownloadTokens(tokenRecord, token)
	assert.NilError(t, err)

	assert.Check(t, is.DeepEqual(token, d.DownloadToken.Token.String))
	assert.Check(t, d.DownloadToken.Token.Valid)
	assert.Check(t, !d.Expires.Valid)
}

func TestDownload_SetDownloadTokens_EmptySecret(t *testing.T) {
	token := "test-token"

	tokenRecord := &ent.FileDownloadToken{
		Token:  token,
		Secret: &[]byte{},
	}

	d := &Download{
		Token: token,
	}

	err := d.setDownloadTokens(tokenRecord, token)
	assert.Assert(t, err != nil)
	assert.Check(t, is.ErrorIs(err, ErrNotFound))
}

func TestDownload_GetDownloadToken(t *testing.T) {
	tests := []struct {
		name     string
		download *Download
		expected string
	}{
		{
			name: "valid token",
			download: &Download{
				DownloadToken: DownloadToken{
					Token: sql.NullString{String: "test-token", Valid: true},
				},
			},
			expected: "test-token",
		},
		{
			name: "invalid token",
			download: &Download{
				DownloadToken: DownloadToken{
					Token: sql.NullString{Valid: false},
				},
			},
			expected: "",
		},
		{
			name: "empty token",
			download: &Download{
				DownloadToken: DownloadToken{
					Token: sql.NullString{String: "", Valid: true},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.download.GetDownloadToken()
			assert.Check(t, is.DeepEqual(tt.expected, result))
		})
	}
}

func TestDownload_GetDownloadExpires(t *testing.T) {
	tests := []struct {
		name        string
		download    *Download
		expectError bool
		checkZero   bool
	}{
		{
			name: "valid expiration",
			download: &Download{
				Expires: sql.NullString{
					String: time.Now().Add(1 * time.Hour).Format(time.RFC3339Nano),
					Valid:  true,
				},
			},
			expectError: false,
			checkZero:   false,
		},
		{
			name: "invalid expiration",
			download: &Download{
				Expires: sql.NullString{Valid: false},
			},
			expectError: false,
			checkZero:   true,
		},
		{
			name: "malformed expiration",
			download: &Download{
				Expires: sql.NullString{
					String: "not-a-valid-time",
					Valid:  true,
				},
			},
			expectError: true,
			checkZero:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.download.GetDownloadExpires()

			if tt.expectError {
				assert.Check(t, err != nil)
			} else {
				assert.Check(t, err)
				if tt.checkZero {
					assert.Check(t, result.IsZero())
				} else {
					assert.Check(t, !result.IsZero())
				}
			}
		})
	}
}

func TestValidateTokenAuthorization_NoUserID(t *testing.T) {
	ctx := context.Background()
	downloadToken := &tokens.DownloadToken{
		UserID: ulid.ULID{},
	}

	err := validateTokenAuthorization(ctx, downloadToken)
	assert.Check(t, err)
}

func TestValidateTokenAuthorization_WithMatchingUser(t *testing.T) {
	userULID := ulids.New()

	authUser := &auth.Caller{
		SubjectID: userULID.String(),
	}

	ctx := auth.WithCaller(context.Background(), authUser)

	downloadToken := &tokens.DownloadToken{
		UserID: userULID,
	}

	err := validateTokenAuthorization(ctx, downloadToken)
	assert.Check(t, err)
}

func TestValidateTokenAuthorization_WithMismatchedUser(t *testing.T) {
	userULID1 := ulids.New()
	userULID2 := ulids.New()

	authUser := &auth.Caller{
		SubjectID: userULID1.String(),
	}

	ctx := auth.WithCaller(context.Background(), authUser)

	downloadToken := &tokens.DownloadToken{
		UserID: userULID2,
	}

	err := validateTokenAuthorization(ctx, downloadToken)
	assert.Check(t, err != nil)
	assert.Check(t, is.ErrorIs(err, ErrUnauthorized))
}

func TestValidateTokenAuthorization_NoAuthUser(t *testing.T) {
	userULID := ulids.New()

	ctx := context.Background()

	downloadToken := &tokens.DownloadToken{
		UserID: userULID,
	}

	err := validateTokenAuthorization(ctx, downloadToken)
	assert.Check(t, err != nil)
	assert.Check(t, is.ErrorIs(err, ErrUnauthorized))
}

func TestBuildStorageFile(t *testing.T) {
	fileID := ulids.New().String()
	fileName := "test-file.txt"
	storagePath := "/path/to/file"
	storageVolume := "test-bucket"
	contentType := "text/plain"
	provider := "database"

	entFile := &ent.File{
		ID:                  fileID,
		ProvidedFileName:    fileName,
		StoragePath:         storagePath,
		StorageVolume:       storageVolume,
		DetectedContentType: contentType,
		StorageProvider:     provider,
	}

	result := buildStorageFile(entFile)

	assert.Check(t, result != nil)
	assert.Check(t, is.DeepEqual(fileID, result.ID))
	assert.Check(t, is.DeepEqual(fileName, result.OriginalName))
	assert.Check(t, is.DeepEqual(storagetypes.ProviderType(provider), result.ProviderType))
	assert.Check(t, is.DeepEqual(storagePath, result.FileMetadata.Key))
	assert.Check(t, is.DeepEqual(storageVolume, result.FileMetadata.Bucket))
	assert.Check(t, is.DeepEqual(contentType, result.FileMetadata.ContentType))
	assert.Check(t, result.FileMetadata.ProviderHints != nil)
	assert.Check(t, is.DeepEqual(storagetypes.ProviderType(provider), result.FileMetadata.ProviderHints.KnownProvider))
}

func TestBuildObjectURI_FromEntFile(t *testing.T) {
	uri := "s3://bucket/path/to/file"
	entFile := &ent.File{
		URI: uri,
	}

	result := buildObjectURI(nil, entFile)
	assert.Check(t, is.DeepEqual(uri, result))
}

func TestBuildObjectURI_FromStorageFile(t *testing.T) {
	fullURI := "database://volume/path"
	storageFile := &storagetypes.File{
		FullURI: fullURI,
	}

	result := buildObjectURI(storageFile, nil)
	assert.Check(t, is.DeepEqual(fullURI, result))
}

func TestBuildObjectURI_PreferEntFile(t *testing.T) {
	entURI := "s3://bucket/file1"
	storageURI := "s3://bucket/file2"

	entFile := &ent.File{
		URI: entURI,
	}

	storageFile := &storagetypes.File{
		FullURI: storageURI,
	}

	result := buildObjectURI(storageFile, entFile)
	assert.Check(t, is.DeepEqual(entURI, result))
}

func TestBuildObjectURI_EmptyEntFileURI(t *testing.T) {
	storageURI := "database://volume/path"

	entFile := &ent.File{
		URI: "",
	}

	storageFile := &storagetypes.File{
		FullURI: storageURI,
	}

	result := buildObjectURI(storageFile, entFile)
	assert.Check(t, is.DeepEqual(storageURI, result))
}

func TestBuildObjectURI_BothNil(t *testing.T) {
	result := buildObjectURI(nil, nil)
	assert.Check(t, is.Equal("", result))
}

func TestBuildObjectURI_EmptyBoth(t *testing.T) {
	entFile := &ent.File{
		URI: "",
	}

	storageFile := &storagetypes.File{
		FullURI: "",
	}

	result := buildObjectURI(storageFile, entFile)
	assert.Check(t, is.Equal("", result))
}
