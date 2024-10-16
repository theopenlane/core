package objects

import (
	"github.com/theopenlane/core/pkg/objects"
)

// validMimeTypes is a map of valid mime types for the given key
// to be used in the validation function
// add the key and the valid mime types to the map
var validMimeTypes = map[string][]string{
	"avatarFile": {"image/jpeg", "image/png"},
}

// MimeTypeValidator returns a validation function for the given key
var MimeTypeValidator objects.ValidationFunc = func(f objects.File) error {
	if mimes, ok := validMimeTypes[f.FieldName]; ok {
		return objects.MimeTypeValidator(mimes...)(f)
	}

	return nil
}
