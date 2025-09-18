package storage_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestProviderTypeConstants(t *testing.T) {
	// Test that provider type constants are properly defined
	assert.Equal(t, "s3", string(storage.S3Provider))
	assert.Equal(t, "r2", string(storage.R2Provider))
	assert.Equal(t, "gcs", string(storage.GCSProvider))
	assert.Equal(t, "disk", string(storage.DiskProvider))
}

func TestConfigurationConstants(t *testing.T) {
	// Test default configuration values
	assert.Equal(t, int64(32<<20), int64(storage.DefaultMaxFileSize)) // 32MB
	assert.Equal(t, int64(32<<20), int64(storage.DefaultMaxMemory))   // 32MB
	assert.Equal(t, "uploadFile", storage.DefaultUploadFileKey)
}

func TestDefaultValidationFunc(t *testing.T) {
	opts := &storage.UploadOptions{
		FileName:    "test.txt",
		ContentType: "text/plain",
	}

	err := storage.DefaultValidationFunc(nil, opts)
	assert.NoError(t, err, "DefaultValidationFunc should always return nil")
}

func TestDefaultNameGeneratorFunc(t *testing.T) {
	originalName := "test_file.txt"
	generatedName := storage.DefaultNameGeneratorFunc(originalName)
	assert.Equal(t, originalName, generatedName, "DefaultNameGeneratorFunc should return the original name unchanged")
}

func TestDefaultSkipper(t *testing.T) {
	req, err := http.NewRequest("POST", "/upload", nil)
	assert.NoError(t, err)

	result := storage.DefaultSkipper(req)
	assert.False(t, result, "DefaultSkipper should always return false")
}

func TestDefaultErrorResponseHandler(t *testing.T) {
	testError := assert.AnError
	statusCode := http.StatusBadRequest

	handler := storage.DefaultErrorResponseHandler(testError, statusCode)
	assert.NotNil(t, handler, "DefaultErrorResponseHandler should return a non-nil handler")

	// We can't easily test the HTTP response without a full HTTP test setup,
	// but we can verify that the handler function is created successfully
}

func TestIntegrationStruct(t *testing.T) {
	integration := storage.Integration{
		ID:             "integration-123",
		OrganizationID: "org-456",
		ProviderType:   storage.S3Provider,
		HushID:         "hush-789",
		Config: map[string]any{
			"region": "us-east-1",
			"bucket": "my-bucket",
		},
		Enabled:     true,
		Name:        "Production S3",
		Description: "Main production storage",
	}

	assert.Equal(t, "integration-123", integration.ID)
	assert.Equal(t, "org-456", integration.OrganizationID)
	assert.Equal(t, storage.S3Provider, integration.ProviderType)
	assert.Equal(t, "hush-789", integration.HushID)
	assert.True(t, integration.Enabled)
	assert.Equal(t, "Production S3", integration.Name)
	assert.Equal(t, "Main production storage", integration.Description)
	assert.Equal(t, "us-east-1", integration.Config["region"])
	assert.Equal(t, "my-bucket", integration.Config["bucket"])
}

func TestFileStruct(t *testing.T) {
	file := storage.File{
		ID:                "file-123",
		Name:              "test.txt",
		OriginalName:      "original_test.txt",
		MD5:               []byte("md5hash"),
		URI:               "https://example.com/test.txt",
		PresignedURL:      "https://presigned.example.com/test.txt",
		FieldName:         "upload",
		FolderDestination: "/uploads",
		ProvidedExtension: ".txt",
		Parent: storage.ParentObject{
			ID:   "parent-456",
			Type: "Evidence",
		},
		Metadata: map[string]string{
			"uploaded_by": "user-123",
			"department":  "compliance",
		},
	}

	assert.Equal(t, "file-123", file.ID)
	assert.Equal(t, "test.txt", file.Name)
	assert.Equal(t, "original_test.txt", file.OriginalName)
	assert.Equal(t, []byte("md5hash"), file.MD5)
	assert.Equal(t, "https://example.com/test.txt", file.URI)
	assert.Equal(t, "https://presigned.example.com/test.txt", file.PresignedURL)
	assert.Equal(t, "upload", file.FieldName)
	assert.Equal(t, "/uploads", file.FolderDestination)
	assert.Equal(t, ".txt", file.ProvidedExtension)
	assert.Equal(t, "parent-456", file.Parent.ID)
	assert.Equal(t, "Evidence", file.Parent.Type)
	assert.Equal(t, "user-123", file.Metadata["uploaded_by"])
	assert.Equal(t, "compliance", file.Metadata["department"])
}

func TestUploadOptions(t *testing.T) {
	opts := storage.UploadOptions{
		FileName:    "test.txt",
		ContentType: "text/plain",
		Metadata: map[string]string{
			"key": "value",
		},
		ProviderHints: &storage.ProviderHints{
			Feature: "evidence",
		},
	}

	assert.Equal(t, "test.txt", opts.FileName)
	assert.Equal(t, "text/plain", opts.ContentType)
	assert.Equal(t, "value", opts.Metadata["key"])

	// Test provider hints type assertion
	if hints, ok := opts.ProviderHints.(*storage.ProviderHints); ok {
		assert.Equal(t, "evidence", hints.Feature)
	} else {
		t.Error("ProviderHints should be of type *ProviderHints")
	}
}

func TestDownloadResult(t *testing.T) {
	result := storage.DownloadResult{
		File: []byte("file content"),
		Size: 12,
	}

	assert.Equal(t, []byte("file content"), result.File)
	assert.Equal(t, int64(12), result.Size)
}

func TestFileUpload(t *testing.T) {
	upload := storage.FileUpload{
		Filename:             "test.txt",
		Size:                 100,
		ContentType:          "text/plain",
		Key:                  "upload",
		CorrelatedObjectID:   "object-123",
		CorrelatedObjectType: "Evidence",
	}

	assert.Equal(t, "test.txt", upload.Filename)
	assert.Equal(t, int64(100), upload.Size)
	assert.Equal(t, "text/plain", upload.ContentType)
	assert.Equal(t, "upload", upload.Key)
	assert.Equal(t, "object-123", upload.CorrelatedObjectID)
	assert.Equal(t, "Evidence", upload.CorrelatedObjectType)
}

func TestParentObject(t *testing.T) {
	parent := storage.ParentObject{
		ID:   "parent-123",
		Type: "Evidence",
	}

	assert.Equal(t, "parent-123", parent.ID)
	assert.Equal(t, "Evidence", parent.Type)
}

func TestProviderConfig(t *testing.T) {
	config := storage.ProviderConfig{
		Enabled:     true,
		Keys:        []string{"upload1", "upload2"},
		MaxSizeMB:   100,
		MaxMemoryMB: 64,
		Providers: storage.ProviderConfigs{
			S3: storage.ProviderCredentials{
				Enabled:         true,
				AccessKeyID:     "access-key",
				SecretAccessKey: "secret-key",
				Region:          "us-east-1",
				Bucket:          "my-bucket",
			},
			CloudflareR2: storage.ProviderCredentials{
				Enabled:   true,
				AccountID: "account-123",
				APIToken:  "api-token",
				Bucket:    "r2-bucket",
			},
			GCS: storage.ProviderCredentials{
				Enabled:         true,
				ProjectID:       "project-123",
				CredentialsJSON: `{"type": "service_account"}`,
				Bucket:          "gcs-bucket",
			},
			Disk: storage.ProviderCredentials{
				Enabled: true,
			},
		},
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, []string{"upload1", "upload2"}, config.Keys)
	assert.Equal(t, int64(100), config.MaxSizeMB)
	assert.Equal(t, int64(64), config.MaxMemoryMB)

	// Test S3 provider
	assert.True(t, config.Providers.S3.Enabled)
	assert.Equal(t, "access-key", config.Providers.S3.AccessKeyID)
	assert.Equal(t, "secret-key", config.Providers.S3.SecretAccessKey)
	assert.Equal(t, "us-east-1", config.Providers.S3.Region)
	assert.Equal(t, "my-bucket", config.Providers.S3.Bucket)

	// Test CloudflareR2 provider
	assert.True(t, config.Providers.CloudflareR2.Enabled)
	assert.Equal(t, "account-123", config.Providers.CloudflareR2.AccountID)
	assert.Equal(t, "api-token", config.Providers.CloudflareR2.APIToken)
	assert.Equal(t, "r2-bucket", config.Providers.CloudflareR2.Bucket)

	// Test GCS provider
	assert.True(t, config.Providers.GCS.Enabled)
	assert.Equal(t, "project-123", config.Providers.GCS.ProjectID)
	assert.Equal(t, `{"type": "service_account"}`, config.Providers.GCS.CredentialsJSON)
	assert.Equal(t, "gcs-bucket", config.Providers.GCS.Bucket)

	// Test Disk provider
	assert.True(t, config.Providers.Disk.Enabled)
}

func TestProviderCredentials(t *testing.T) {
	creds := storage.ProviderCredentials{
		Enabled:         true,
		AccessKeyID:     "access-key-123",
		SecretAccessKey: "secret-key-456",
		Region:          "us-west-2",
		Bucket:          "test-bucket",
		Endpoint:        "https://custom-endpoint.com",
		ProjectID:       "gcp-project-123",
		CredentialsJSON: `{"type": "service_account", "project_id": "gcp-project-123"}`,
		AccountID:       "cloudflare-account-123",
		APIToken:        "cloudflare-token-456",
	}

	assert.True(t, creds.Enabled)
	assert.Equal(t, "access-key-123", creds.AccessKeyID)
	assert.Equal(t, "secret-key-456", creds.SecretAccessKey)
	assert.Equal(t, "us-west-2", creds.Region)
	assert.Equal(t, "test-bucket", creds.Bucket)
	assert.Equal(t, "https://custom-endpoint.com", creds.Endpoint)
	assert.Equal(t, "gcp-project-123", creds.ProjectID)
	assert.Contains(t, creds.CredentialsJSON, "gcp-project-123")
	assert.Equal(t, "cloudflare-account-123", creds.AccountID)
	assert.Equal(t, "cloudflare-token-456", creds.APIToken)
}

func TestFilesType(t *testing.T) {
	files := storage.Files{
		"upload1": []storage.File{
			{ID: "file1", Name: "test1.txt"},
			{ID: "file2", Name: "test2.txt"},
		},
		"upload2": []storage.File{
			{ID: "file3", Name: "test3.txt"},
		},
	}

	assert.Len(t, files, 2)
	assert.Len(t, files["upload1"], 2)
	assert.Len(t, files["upload2"], 1)
	assert.Equal(t, "file1", files["upload1"][0].ID)
	assert.Equal(t, "test1.txt", files["upload1"][0].Name)
	assert.Equal(t, "file3", files["upload2"][0].ID)
	assert.Equal(t, "test3.txt", files["upload2"][0].Name)
}


func TestFunctionTypes(t *testing.T) {
	// Test that function types are properly defined

	// ValidationFunc
	var validationFunc storage.ValidationFunc = func(ctx context.Context, opts *storage.UploadOptions) error {
		return nil
	}
	assert.NotNil(t, validationFunc)

	// UploaderFunc
	var uploaderFunc storage.UploaderFunc = func(ctx context.Context, service *storage.ObjectService, files []storage.FileUpload) ([]storage.File, error) {
		return nil, nil
	}
	assert.NotNil(t, uploaderFunc)

	// NameGeneratorFunc
	var nameGenFunc storage.NameGeneratorFunc = func(originalName string) string {
		return originalName
	}
	assert.NotNil(t, nameGenFunc)

	// SkipperFunc
	var skipperFunc storage.SkipperFunc = func(r *http.Request) bool {
		return false
	}
	assert.NotNil(t, skipperFunc)

	// ErrResponseHandler
	var errHandler storage.ErrResponseHandler = func(err error, statusCode int) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {}
	}
	assert.NotNil(t, errHandler)
}

func TestTypeAliases(t *testing.T) {
	// Test that type aliases are working correctly by creating instances

	// These should compile without errors, demonstrating that the aliases are properly defined
	var _ storage.ProviderType = storage.S3Provider
	var _ storage.UploadFileOptions = storage.UploadFileOptions{}
	var _ storage.UploadedFileMetadata = storage.UploadedFileMetadata{}
	var _ storage.DownloadFileOptions = storage.DownloadFileOptions{}
	var _ storage.DownloadFileMetadata = storage.DownloadFileMetadata{}
	var _ storage.ProviderHints = storage.ProviderHints{}
	var _ storage.FileStorageMetadata = storage.FileStorageMetadata{}
}

func TestComplexTypeInteraction(t *testing.T) {
	// Test interaction between different types
	upload := storage.FileUpload{
		Filename:    "test.txt",
		Size:        100,
		ContentType: "text/plain",
	}

	opts := storage.UploadOptions{
		FileName:    upload.Filename,
		ContentType: upload.ContentType,
		Metadata: map[string]string{
			"original_size": "100",
		},
	}

	file := storage.File{
		ID:           "file-123",
		Name:         opts.FileName,
		OriginalName: upload.Filename,
		Metadata:     opts.Metadata,
	}

	parent := storage.ParentObject{
		ID:   file.ID,
		Type: "TestFile",
	}

	assert.Equal(t, upload.Filename, opts.FileName)
	assert.Equal(t, opts.FileName, file.Name)
	assert.Equal(t, file.ID, parent.ID)
	assert.Equal(t, "100", file.Metadata["original_size"])
}
