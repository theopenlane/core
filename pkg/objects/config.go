package objects

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/theopenlane/utils/ulids"
)

// Config is the configuration for the object store
type Config struct {
	// Enabled indicates if the store is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// Provider is the name of the provider, eg. disk, s3, will default to disk if nothing is set
	Provider string `json:"provider" koanf:"provider" `
	// AccessKey is the access key for the storage provider
	AccessKey string `json:"accessKey" koanf:"accessKey" sensitive:"true"`
	// Region is the region for the storage provider
	Region string `json:"region" koanf:"region"`
	// SecretKey is the secret key for the storage provider
	SecretKey string `json:"secretKey" koanf:"secretKey" sensitive:"true"`
	// CredentialsJSON is the credentials JSON for the storage provider
	CredentialsJSON string `json:"credentialsJSON" koanf:"credentialsJSON" sensitive:"true"`
	// DefaultBucket is the default bucket name for the storage provider, if not set, it will use the default
	// this is the local path for disk storage or the bucket name for S3
	DefaultBucket string `json:"defaultBucket" koanf:"defaultBucket" default:"file_uploads"`
	// LocalURL is the URL to use for the "presigned" URL for the file when using local storage
	// e.g for local development, this can be http://localhost:17608/files/
	LocalURL string `json:"localURL" koanf:"localURL" default:"http://localhost:17608/files/"`
	// Keys is a list of keys to look for in the multipart form on the REST request
	// if the keys are not found, the request upload will be skipped
	// this is not used when uploading files with gqlgen and the graphql handler
	Keys []string `json:"keys" koanf:"keys" default:"[uploadFile]"`
	// MaxUploadSizeMB is the maximum size of file uploads to accept in megabytes
	MaxUploadSizeMB int64 `json:"maxSizeMB" koanf:"maxSizeMB"`
	// MaxUploadMemoryMB is the maximum memory in megabytes to use when parsing a multipart form
	MaxUploadMemoryMB int64 `json:"maxMemoryMB" koanf:"maxMemoryMB"`
	// Endpoint is used for other s3 compatible storage systems e.g minio, digital ocean spaces .
	// they do not use the same s3 endpoint
	Endpoint string `json:"endpoint" koanf:"endpoint"`
	// UsePathStyle is useful for other s3 compatible systems that use path styles not bucket.host path
	// minio is a popular example here
	UsePathStyle bool `json:"usePathStyle" koanf:"usePathStyle"`
}

var (
	// allows all file pass through
	defaultValidationFunc ValidationFunc = func(_ File) error {
		return nil
	}

	// defaultNameGeneratorFunc uses the objects-158888-originalname to
	// upload files
	defaultNameGeneratorFunc NameGeneratorFunc = func(s string) string {
		return fmt.Sprintf("objects-%d-%s", time.Now().Unix(), s)
	}

	// defaultFileUploadMaxSize is the default maximum file upload size
	defaultFileUploadMaxSize int64 = 32 << 20

	// defaultMaxMemorySize is the default maximum memory size for parsing a multipart form
	defaultMaxMemorySize int64 = 32 << 20

	// defaultErrorResponseHandler is the default error response handler
	defaultErrorResponseHandler ErrResponseHandler = func(err error, statusCode int) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			fmt.Fprintf(w, `{"message" : "could not upload file", "error" : "%s"}`, err.Error())
		}
	}

	defaultSkipper SkipperFunc = func(_ *http.Request) bool {
		return false
	}

	// defaultUploader is the default uploader function that uploads files to the storage
	defaultUploader UploaderFunc = func(ctx context.Context, u *Objects, files []FileUpload) ([]File, error) {
		uploadedFiles := make([]File, 0, len(files))

		for _, f := range files {
			fileID := ulids.New().String()
			uploadedFileName := u.NameFuncGenerator(fileID + "_" + f.Filename)

			contentType, err := DetectContentType(f.File)
			if err != nil {
				return nil, err
			}

			fileData := File{
				ID:               fileID,
				FieldName:        f.Key,
				OriginalName:     f.Filename,
				UploadedFileName: uploadedFileName,
				MimeType:         f.ContentType,
				ContentType:      contentType,
			}

			// validate the file
			if err := u.ValidationFunc(fileData); err != nil {
				return nil, err
			}

			metadata, err := u.Storage.Upload(ctx, files[0].File, &UploadFileOptions{
				FileName: uploadedFileName,
			})
			if err != nil {
				return nil, err
			}

			// add metadata to file information
			fileData.Size = metadata.Size
			fileData.FolderDestination = metadata.FolderDestination
			fileData.StorageKey = metadata.Key

			uploadedFiles = append(uploadedFiles, fileData)
		}

		return uploadedFiles, nil
	}
)

// OrganizationNameFunc is a function that generates the organization name
var OrganizationNameFunc NameGeneratorFunc = func(s string) string {
	return s
}
