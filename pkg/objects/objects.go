package objects

import (
	"context"
	"io"
	"time"
)

// Storage is the primary interface that must be implemented by any storage backend and for interacting with Objects
type Storage interface {
	Upload(context.Context, io.Reader, *UploadFileOptions) (*UploadedFileMetadata, error)
	ManagerUpload(context.Context, [][]byte) error
	Download(context.Context, string, *DownloadFileOptions) (*DownloadFileMetadata, io.ReadCloser, error)
	GetPresignedURL(context.Context, string) string
	io.Closer
}

//go:generate mockgen -destination=mocks/objects.go -source=objects.go -package mocks

// Objects is the definition for handling objects and file uploads
type Objects struct {
	// Storage is the storage backend that will be used to store the uploaded files
	Storage Storage
	// MaxSize is the maximum size of file uploads to accept
	MaxSize int64
	// ignoreNonExistentKeys is a flag that indicates the handler should skip multipart form key values which do not match the configured
	IgnoreNonExistentKeys bool
	// ValidationFunc is a custom validation function
	ValidationFunc ValidationFunc
	// NameFuncGenerator is a function that allows you to rename your uploaded files
	NameFuncGenerator NameGeneratorFunc
	// ErrorResponseHandler is a custom error response handler
	ErrorResponseHandler ErrResponseHandler
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

	if handler.ValidationFunc == nil {
		handler.ValidationFunc = defaultValidationFunc
	}

	if handler.NameFuncGenerator == nil {
		handler.NameFuncGenerator = defaultNameGeneratorFunc
	}

	if handler.ErrorResponseHandler == nil {
		handler.ErrorResponseHandler = defaultErrorResponseHandler
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
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Type      string    `json:"type"`
	Thumbnail *string   `json:"thumbnail"`
	MD5       []byte    `json:"md5"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	OwnerID   string    `json:"owner_id"`
	// FieldName denotes the field from the multipart form
	FieldName string `json:"field_name,omitempty"`

	// The name of the file from the client side
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

	// Size in bytes of the uploaded file
	Size              int64 `json:"size,omitempty"`
	Metadata          map[string]string
	Bucket            string
	PresignedURL      string `json:"url"`
	ProvidedExtension string `json:"provided_extension"`
}

// NameGeneratorFunc allows you alter the name of the file before it is ultimately uplaoded and stored
type NameGeneratorFunc func(s string) string

// UploadedFileMetadata is a struct that holds information about a file that was successfully uploaded
type UploadedFileMetadata struct {
	FolderDestination string `json:"folder_destination,omitempty"`
	Key               string `json:"key,omitempty"`
	Size              int64  `json:"size,omitempty"`
	PresignedURL      string `json:"presigned_url,omitempty"`
}

// DonwloadFileMetadata is a struct that holds information about a file that was successfully downloaded
type DownloadFileMetadata struct {
	FolderDestination string `json:"folder_destination,omitempty"`
	Key               string `json:"key,omitempty"`
	Size              int64  `json:"size,omitempty"`
}
