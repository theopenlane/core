package objects

import (
	"context"

	"github.com/theopenlane/core/pkg/objects/storage"
)

var importSchemaMimeTypes = []string{
	"text/plain; charset=utf-8", "text/plain",
	"text/markdown",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}

// sharedMimeTypes contains mime types that are shared between different file types
var sharedMimeTypes = []string{
	"image/jpeg", "image/png",
	"application/pdf",
	"text/plain",
	"text/plain; charset=utf-8",
	"application/zip",
	"application/rtf", // rich text format
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document", // docx
	"application/vnd.oasis.opendocument.text",                                 // open document text
	"text/markdown",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // xlsx
	"application/x-vnd.oasis.opendocument.spreadsheet",                  // open document spreadsheet
	"text/csv",
	"application/x-yaml", "application/x-yaml; charset=utf-8", "text/yaml",
	"application/json", "application/json; charset=utf-8",
}

// validMimeTypes is a map of valid mime types for the given key
// to be used in the validation function
// add the key and the valid mime types to the map
var validMimeTypes = map[string][]string{
	"avatarFile":    {"image/jpeg", "image/png"},
	"logoFile":      {"image/jpeg", "image/png", "image/svg+xml"},
	"faviconFile":   {"image/jpeg", "image/png", "image/x-icon"},
	"evidenceFiles": sharedMimeTypes,
	"noteFiles":     sharedMimeTypes,
	"exportFiles":   {"text/csv", "text/plain; charset=utf-8", "text/plain", "application/json", "application/json; charset=utf-8"},
	"procedureFile": importSchemaMimeTypes,
	"policyFile":    importSchemaMimeTypes,
}

// ApplicationMimeTypeValidator returns a validation function for the given key
var ApplicationMimeTypeValidator storage.ValidationFunc = func(ctx context.Context, opts *storage.UploadOptions) error {
	if mimes, ok := validMimeTypes[opts.FileName]; ok {
		return storage.MimeTypeValidator(mimes...)(ctx, opts)
	}

	// Default to sharedMimeTypes if the type isn't in the map
	return storage.MimeTypeValidator(sharedMimeTypes...)(ctx, opts)
}