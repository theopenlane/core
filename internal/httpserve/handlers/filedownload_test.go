package handlers

import (
	"context"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ent "github.com/theopenlane/ent/generated"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	storagetypes "github.com/theopenlane/shared/objects/storage/types"
	"github.com/theopenlane/utils/ulids"
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
	require.NoError(t, err)

	assert.Equal(t, token, d.DownloadToken.Token.String)
	assert.True(t, d.DownloadToken.Token.Valid)
	assert.Equal(t, secret, d.Secret)
	assert.True(t, d.Expires.Valid)
	assert.Equal(t, ttl.Format(time.RFC3339Nano), d.Expires.String)
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
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
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
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
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
	require.NoError(t, err)

	assert.Equal(t, token, d.DownloadToken.Token.String)
	assert.True(t, d.DownloadToken.Token.Valid)
	assert.False(t, d.Expires.Valid)
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
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFound)
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
			assert.Equal(t, tt.expected, result)
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
				DownloadToken: DownloadToken{
					Expires: sql.NullString{
						String: time.Now().Add(1 * time.Hour).Format(time.RFC3339Nano),
						Valid:  true,
					},
				},
			},
			expectError: false,
			checkZero:   false,
		},
		{
			name: "invalid expiration",
			download: &Download{
				DownloadToken: DownloadToken{
					Expires: sql.NullString{Valid: false},
				},
			},
			expectError: false,
			checkZero:   true,
		},
		{
			name: "malformed expiration",
			download: &Download{
				DownloadToken: DownloadToken{
					Expires: sql.NullString{
						String: "not-a-valid-time",
						Valid:  true,
					},
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkZero {
					assert.True(t, result.IsZero())
				} else {
					assert.False(t, result.IsZero())
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
	assert.NoError(t, err)
}

func TestValidateTokenAuthorization_WithMatchingUser(t *testing.T) {
	userULID := ulids.New()

	authUser := &auth.AuthenticatedUser{
		SubjectID: userULID.String(),
	}

	ctx := auth.WithAuthenticatedUser(context.Background(), authUser)

	downloadToken := &tokens.DownloadToken{
		UserID: userULID,
	}

	err := validateTokenAuthorization(ctx, downloadToken)
	assert.NoError(t, err)
}

func TestValidateTokenAuthorization_WithMismatchedUser(t *testing.T) {
	userULID1 := ulids.New()
	userULID2 := ulids.New()

	authUser := &auth.AuthenticatedUser{
		SubjectID: userULID1.String(),
	}

	ctx := auth.WithAuthenticatedUser(context.Background(), authUser)

	downloadToken := &tokens.DownloadToken{
		UserID: userULID2,
	}

	err := validateTokenAuthorization(ctx, downloadToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestValidateTokenAuthorization_NoAuthUser(t *testing.T) {
	userULID := ulids.New()

	ctx := context.Background()

	downloadToken := &tokens.DownloadToken{
		UserID: userULID,
	}

	err := validateTokenAuthorization(ctx, downloadToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
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

	assert.NotNil(t, result)
	assert.Equal(t, fileID, result.ID)
	assert.Equal(t, fileName, result.OriginalName)
	assert.Equal(t, storagetypes.ProviderType(provider), result.ProviderType)
	assert.Equal(t, storagePath, result.FileMetadata.Key)
	assert.Equal(t, storageVolume, result.FileMetadata.Bucket)
	assert.Equal(t, contentType, result.FileMetadata.ContentType)
	assert.NotNil(t, result.FileMetadata.ProviderHints)
	assert.Equal(t, storagetypes.ProviderType(provider), result.FileMetadata.ProviderHints.KnownProvider)
}

func TestBuildObjectURI_FromEntFile(t *testing.T) {
	uri := "s3://bucket/path/to/file"
	entFile := &ent.File{
		URI: uri,
	}

	result := buildObjectURI(nil, entFile)
	assert.Equal(t, uri, result)
}

func TestBuildObjectURI_FromStorageFile(t *testing.T) {
	fullURI := "database://volume/path"
	storageFile := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			FullURI: fullURI,
		},
	}

	result := buildObjectURI(storageFile, nil)
	assert.Equal(t, fullURI, result)
}

func TestBuildObjectURI_PreferEntFile(t *testing.T) {
	entURI := "s3://bucket/file1"
	storageURI := "s3://bucket/file2"

	entFile := &ent.File{
		URI: entURI,
	}

	storageFile := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			FullURI: storageURI,
		},
	}

	result := buildObjectURI(storageFile, entFile)
	assert.Equal(t, entURI, result)
}

func TestBuildObjectURI_EmptyEntFileURI(t *testing.T) {
	storageURI := "database://volume/path"

	entFile := &ent.File{
		URI: "",
	}

	storageFile := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			FullURI: storageURI,
		},
	}

	result := buildObjectURI(storageFile, entFile)
	assert.Equal(t, storageURI, result)
}

func TestBuildObjectURI_BothNil(t *testing.T) {
	result := buildObjectURI(nil, nil)
	assert.Equal(t, "", result)
}

func TestBuildObjectURI_EmptyBoth(t *testing.T) {
	entFile := &ent.File{
		URI: "",
	}

	storageFile := &storagetypes.File{
		FileMetadata: storagetypes.FileMetadata{
			FullURI: "",
		},
	}

	result := buildObjectURI(storageFile, entFile)
	assert.Equal(t, "", result)
}
