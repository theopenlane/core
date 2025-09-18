package storage

import (
	"context"
	"io"
	"time"

	"github.com/theopenlane/utils/ulids"
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
	// Apply validation
	if err := s.validationFunc(ctx, opts); err != nil {
		return nil, err
	}

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

	var providerHints *ProviderHints
	if opts.ProviderHints != nil {
		if hints, ok := opts.ProviderHints.(*ProviderHints); ok {
			providerHints = hints
		}
	}

	storageOpts := &UploadFileOptions{
		FileName:      fileName,
		ContentType:   contentType,
		Metadata:      opts.Metadata,
		ProviderHints: providerHints,
	}

	// Upload using provided storage provider
	metadata, err := provider.Upload(ctx, reader, storageOpts)
	if err != nil {
		return nil, err
	}

	// Create file object with complete metadata
	file := &File{
		ID:                  ulids.New().String(),
		Name:                fileName,
		OriginalName:        fileName,
		FileStorageMetadata: metadata.FileStorageMetadata, // Use the full metadata from provider
		FolderDestination:   metadata.FolderDestination,
		Metadata:            opts.Metadata,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Set FieldName from metadata if available
	if opts.Metadata != nil {
		if fieldName, ok := opts.Metadata["key"]; ok {
			file.FieldName = fieldName
		}
	}

	return file, nil
}


// Download downloads a file using a specific storage provider client
func (s *ObjectService) Download(ctx context.Context, provider Provider, file *File) (*DownloadResult, error) {
	storageOpts := &DownloadFileOptions{
		FileName:       file.Key,
		Metadata:       file.Metadata,
		IntegrationID:  file.IntegrationID,
		HushID:         file.HushID,
		OrganizationID: file.OrganizationID,
	}

	metadata, err := provider.Download(ctx, storageOpts)
	if err != nil {
		return nil, err
	}

	return &DownloadResult{
		File: metadata.File,
		Size: metadata.Size,
	}, nil
}

// GetPresignedURL gets a presigned URL for a file using a specific storage provider client
func (s *ObjectService) GetPresignedURL(_ context.Context, provider Provider, file *File, duration time.Duration) (string, error) {
	return provider.GetPresignedURL(file.Key, duration)
}

// Delete deletes a file using a specific storage provider client
func (s *ObjectService) Delete(ctx context.Context, provider Provider, file *File) error {
	return provider.Delete(ctx, file.Key)
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
func (s *ObjectService) defaultUploader(_ context.Context, _ *ObjectService, _ []FileUpload) ([]File, error) {
	return nil, ErrProviderResolutionRequired
}
