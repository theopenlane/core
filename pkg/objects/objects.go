package objects

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Storage is the primary interface that must be implemented by any storage backend and for interacting with Objects
type Storage interface {
	// Upload is used to upload a file to the storage backend
	Upload(context.Context, io.Reader, *UploadFileOptions) (*UploadedFileMetadata, error)
	// ManagerUpload is used to upload multiple files to the storage backend
	ManagerUpload(context.Context, [][]byte) error
	// Download is used to download a file from the storage backend
	Download(context.Context, string, *DownloadFileOptions) (*DownloadFileMetadata, io.ReadCloser, error)
	// GetPresignedURL is used to get a presigned URL for a file in the storage backend
	GetPresignedURL(context.Context, string) (string, error)
	// GetScheme returns the scheme of the storage backend
	GetScheme() *string
	io.Closer
}

//go:generate mockery --config .mockery.yaml

// Objects is the definition for handling objects and file uploads
type Objects struct {
	// Storage is the storage backend that will be used to store the uploaded files
	Storage Storage `json:"-" koanf:"-"`
	// MaxSize is the maximum size of file uploads to accept
	MaxSize int64 `json:"maxSize" koanf:"maxSize"`
	// MaxMemory is the maximum memory to use when parsing a multipart form
	MaxMemory int64 `json:"maxMemory" koanf:"maxMemory"`
	// IgnoreNonExistentKeys is a flag that indicates the handler should skip multipart form key values which do not match the configured
	IgnoreNonExistentKeys bool `json:"ignoreNonExistentKeys" koanf:"ignoreNonExistentKeys"`
	// Keys is a list of keys to look for in the multipart form on the REST request
	// if the keys are not found, the request upload will be skipped
	// this is not used by the graphql handler
	Keys []string `json:"keys" koanf:"keys" default:"[uploadFile]"`
	// ValidationFunc is a custom validation function
	ValidationFunc ValidationFunc `json:"-" koanf:"-"`
	// NameFuncGenerator is a function that allows you to rename your uploaded files
	NameFuncGenerator NameGeneratorFunc `json:"-" koanf:"-"`
	// Uploader is the func that handlers the file upload process and returns the files uploaded
	Uploader UploaderFunc `json:"-" koanf:"-"`
	// Skipper defines a function to skip middleware.
	Skipper SkipperFunc `json:"-" koanf:"-"`
	// ErrorResponseHandler is a custom error response handler
	ErrorResponseHandler ErrResponseHandler `json:"-" koanf:"-"`
}

// New creates a new instance of Objects
func New(opts ...Option) (*Objects, error) {
	handler := &Objects{}

	for _, opt := range opts {
		opt(handler)
	}

	if handler.MaxSize <= 0 {
		handler.MaxSize = defaultFileUploadMaxSize
	}

	if handler.MaxMemory <= 0 {
		handler.MaxMemory = defaultMaxMemorySize
	}

	if handler.ValidationFunc == nil {
		handler.ValidationFunc = defaultValidationFunc
	}

	if handler.NameFuncGenerator == nil {
		handler.NameFuncGenerator = defaultNameGeneratorFunc
	}

	if handler.Uploader == nil {
		handler.Uploader = defaultUploader
	}

	if handler.ErrorResponseHandler == nil {
		handler.ErrorResponseHandler = defaultErrorResponseHandler
	}

	if handler.Skipper == nil {
		handler.Skipper = defaultSkipper
	}

	if handler.Storage == nil {
		return nil, ErrMustProvideStorageBackend
	}

	return handler, nil
}

// Files is a map of field names to a slice of files
type Files map[string][]File

// File is a struct that holds information about a file - there is no distinction between a File received in a multipart form request or used in a download
type File struct {
	// ID is the unique identifier for the file
	ID string `json:"id"`
	// Name of the file
	Name string `json:"name"`
	// Path of the file
	Path string `json:"path"`
	// Type of file that was uploaded
	Type string `json:"type"`
	// Thumbnail is a URL to the thumbnail of the file
	Thumbnail *string `json:"thumbnail"`
	// MD5 hash of the file
	MD5 []byte `json:"md5"`
	// CreatedAt is the time the file was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the time the file was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// OwnerID is the ID of the organization or user who created the file
	OwnerID string `json:"owner_id"`
	// FieldName denotes the field from the multipart form
	FieldName string `json:"field_name,omitempty"`

	// OriginalName of the file from the client side
	OriginalName string `json:"original_name,omitempty"`
	// UploadedFileName denotes the name of the file when it was ultimately
	// uploaded to the storage layer. The distinction is important because of
	// potential changes to the file name that may be done
	UploadedFileName string `json:"uploaded_file_name,omitempty"`
	// FolderDestination is the folder that holds the uploaded file
	FolderDestination string `json:"folder_destination,omitempty"`

	// StorageKey can be used to retrieve the file from the storage backend
	StorageKey string `json:"storage_key,omitempty"`

	// MimeType of the uploaded file
	MimeType string `json:"mime_type,omitempty"`

	// ContentType is the detected content type of the file
	ContentType string `json:"content_type,omitempty"`

	// Size in bytes of the uploaded file
	Size              int64 `json:"size,omitempty"`
	Metadata          map[string]string
	Bucket            string
	PresignedURL      string `json:"url"`
	ProvidedExtension string `json:"provided_extension"`
}

// SKipperFunc is a function that defines whether to skip the middleware
type SkipperFunc func(r *http.Request) bool

// UploaderFunc is a function that handles the file upload process and returns the files uploaded
type UploaderFunc func(ctx context.Context, u *Objects, files []FileUpload) ([]File, error)

// NameGeneratorFunc allows you alter the name of the file before it is ultimately uploaded and stored
type NameGeneratorFunc func(s string) string

// UploadedFileMetadata is a struct that holds information about a file that was successfully uploaded
type UploadedFileMetadata struct {
	// FolderDestination is the folder that holds the file
	FolderDestination string `json:"folder_destination,omitempty"`
	// Key is the unique identifier for the file
	Key string `json:"key,omitempty"`
	// Size in bytes of the uploaded file
	Size int64 `json:"size,omitempty"`
	// PresignedURL is the URL that can be used to download the file
	PresignedURL string `json:"presigned_url,omitempty"`
}

// DownloadFileMetadata is a struct that holds information about a file that was successfully downloaded
type DownloadFileMetadata struct {
	// FolderDestination is the folder that holds the file
	FolderDestination string `json:"folder_destination,omitempty"`
	// Key is the unique identifier for the file
	Key string `json:"key,omitempty"`
	// Size in bytes of the downloaded file
	Size int64 `json:"size,omitempty"`
}
