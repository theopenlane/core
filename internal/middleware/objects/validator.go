package objects

import (
	"github.com/theopenlane/core/pkg/objects"
)

// validMimeTypes is a map of valid mime types for the given key
// to be used in the validation function
// add the key and the valid mime types to the map
var validMimeTypes = map[string][]string{
	"avatarFile": {"image/jpeg", "image/png"},
	"evidenceFiles": {
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
	},
}

// MimeTypeValidator returns a validation function for the given key
var MimeTypeValidator objects.ValidationFunc = func(f objects.File) error {
	if mimes, ok := validMimeTypes[f.FieldName]; ok {
		return objects.MimeTypeValidator(mimes...)(f)
	}

	return nil
}
