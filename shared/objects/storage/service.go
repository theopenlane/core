package storage

import (
	"context"
	"io"
	"time"

	storagetypes "github.com/theopenlane/shared/objects/storage/types"
)

// ObjectService provides pure object management functionality without provider resolution
type ObjectService struct {
	maxSize               int64
	maxMemory             int64
	ignoreNonExistentKeys bool
	keys                  []string
	validationFunc        ValidationFunc
	nameFuncGenerator     NameGeneratorFunc
	uploader              UploaderFunc
	skipper               SkipperFunc
	errorResponseHandler  ErrResponseHandler
}

// NewObjectService creates a new object service instance with default configuration
func NewObjectService() *ObjectService {
	service := &ObjectService{
		maxSize:              DefaultMaxFileSize,
		maxMemory:            DefaultMaxMemory,
		keys:                 []string{DefaultUploadFileKey},
		validationFunc:       DefaultValidationFunc,
		nameFuncGenerator:    DefaultNameGeneratorFunc,
		skipper:              DefaultSkipper,
		errorResponseHandler: DefaultErrorResponseHandler,
	}

	// Set the uploader after service is initialized so defaultUploader can access the service
	service.uploader = service.defaultUploader

	return service
}

// Upload uploads a file using a specific storage provider client
func (s *ObjectService) Upload(ctx context.Context, provider Provider, reader io.Reader, opts *UploadOptions) (*File, error) {
	// Generate file name
	fileName := s.nameFuncGenerator(opts.FileName)

	// Detect content type if not provided
	contentType := opts.ContentType
	if contentType == "" {
		if seeker, ok := reader.(io.ReadSeeker); ok {
			if detectedType, err := DetectContentType(seeker); err == nil {
				contentType = detectedType
			}
		}
	}

	// Validate the file before uploading
	tempFile := File{
		FieldName:    opts.Key,
		OriginalName: fileName,
		FileMetadata: FileMetadata{
			ContentType: contentType,
		},
	}

	if err := s.validationFunc(tempFile); err != nil {
		return nil, err
	}

	storageOpts := &UploadOptions{
		FileName:          fileName,
		ContentType:       contentType,
		Bucket:            opts.Bucket,
		FolderDestination: opts.FolderDestination,
		FileMetadata: FileMetadata{
			Key:           opts.Key,
			Bucket:        opts.Bucket,
			Region:        opts.Region,
			ProviderHints: opts.ProviderHints,
		},
	}

	if opts.ProviderHints != nil {
		storageOpts.ProviderHints = opts.ProviderHints
	}

	// Upload using provided storage provider
	metadata, err := provider.Upload(ctx, reader, storageOpts)
	if err != nil {
		return nil, err
	}

	fileMetadata := FileMetadata{
		Key:           metadata.Key,
		Size:          metadata.Size,
		ContentType:   metadata.ContentType,
		Folder:        metadata.Folder,
		Bucket:        metadata.Bucket,
		Region:        metadata.Region,
		FullURI:       metadata.FullURI,
		ProviderType:  metadata.ProviderType,
		PresignedURL:  metadata.PresignedURL,
		Name:          metadata.Name,
		ProviderHints: metadata.ProviderHints,
	}

	if fileMetadata.Key == "" {
		fileMetadata.Key = storageOpts.Key
	}
	if fileMetadata.ContentType == "" {
		fileMetadata.ContentType = contentType
	}
	if fileMetadata.Folder == "" {
		fileMetadata.Folder = storageOpts.FolderDestination
	}
	if fileMetadata.Bucket == "" {
		fileMetadata.Bucket = storageOpts.Bucket
	}

	if fileMetadata.Region == "" {
		fileMetadata.Region = storageOpts.Region
	}

	if fileMetadata.ProviderType == "" {
		fileMetadata.ProviderType = provider.ProviderType()
	}
	if fileMetadata.Name == "" {
		fileMetadata.Name = fileName
	}
	if fileMetadata.ProviderHints == nil {
		fileMetadata.ProviderHints = storageOpts.ProviderHints
	}

	// Create file object with complete metadata
	file := &File{
		OriginalName: fileName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ProviderType: fileMetadata.ProviderType,
		FieldName:    opts.Key,
		FileMetadata: fileMetadata,
	}

	return file, nil
}

// Download downloads a file using a specific storage provider client
func (s *ObjectService) Download(ctx context.Context, provider Provider, file *storagetypes.File, opts *DownloadOptions) (*DownloadedMetadata, error) {
	metadata, err := provider.Download(ctx, file, opts)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

// GetPresignedURL gets a presigned URL for a file using a specific storage provider client
func (s *ObjectService) GetPresignedURL(ctx context.Context, provider Provider, file *storagetypes.File, opts *storagetypes.PresignedURLOptions) (string, error) {
	return provider.GetPresignedURL(ctx, file, opts)
}

// Delete deletes a file using a specific storage provider client
func (s *ObjectService) Delete(ctx context.Context, provider Provider, file *storagetypes.File, opts *storagetypes.DeleteFileOptions) error {
	return provider.Delete(ctx, file, opts)
}

// Skipper returns the configured skipper function
func (s *ObjectService) Skipper() SkipperFunc {
	return s.skipper
}

// ErrorResponseHandler returns the configured error response handler
func (s *ObjectService) ErrorResponseHandler() ErrResponseHandler {
	return s.errorResponseHandler
}

// MaxSize returns the configured maximum file size
func (s *ObjectService) MaxSize() int64 {
	return s.maxSize
}

// MaxMemory returns the configured maximum memory for multipart forms
func (s *ObjectService) MaxMemory() int64 {
	return s.maxMemory
}

// Keys returns the configured form keys
func (s *ObjectService) Keys() []string {
	return s.keys
}

// IgnoreNonExistentKeys returns whether to ignore non-existent form keys
func (s *ObjectService) IgnoreNonExistentKeys() bool {
	return s.ignoreNonExistentKeys
}

// WithUploader returns a new ObjectService with the specified uploader function
func (s *ObjectService) WithUploader(uploader UploaderFunc) *ObjectService {
	newService := *s // Copy the service
	newService.uploader = uploader

	return &newService
}

// WithValidation returns a new ObjectService with the specified validation function
func (s *ObjectService) WithValidation(validationFunc ValidationFunc) *ObjectService {
	newService := *s // Copy the service
	newService.validationFunc = validationFunc

	return &newService
}

// defaultUploader is the default file upload implementation that requires external provider resolution
func (s *ObjectService) defaultUploader(_ context.Context, _ *ObjectService, _ []File) ([]File, error) {
	return nil, ErrProviderResolutionRequired
}
