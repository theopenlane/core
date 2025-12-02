package objects

import (
	"github.com/theopenlane/shared/objects/storage"
)

// File aliases storage.File so callers can reference a single top-level type.
type (
	File            = storage.File
	Files           = storage.Files
	ParentObject    = storage.ParentObject
	FileMetadata    = storage.FileMetadata
	ProviderHints   = storage.ProviderHints
	UploadOptions   = storage.UploadOptions
	DownloadOptions = storage.DownloadOptions
)
