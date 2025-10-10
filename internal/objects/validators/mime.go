package validators

import (
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

var importSchemaMimeTypes = []string{
	"text/plain; charset=utf-8", "text/plain",
	"text/markdown",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}

var sharedMimeTypes = []string{
	"image/jpeg", "image/png",
	"application/pdf",
	"text/plain",
	"text/plain; charset=utf-8",
	"application/zip",
	"application/rtf",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"application/vnd.oasis.opendocument.text",
	"text/markdown",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"application/x-vnd.oasis.opendocument.spreadsheet",
	"text/csv",
	"application/x-yaml", "application/x-yaml; charset=utf-8", "text/yaml",
	"application/json", "application/json; charset=utf-8",
}

var validMimeTypes = map[string][]string{
	"avatarFile":         {"image/jpeg", "image/png"},
	"logoFile":           {"image/jpeg", "image/png", "image/svg+xml"},
	"faviconFile":        {"image/jpeg", "image/png", "image/x-icon"},
	"evidenceFiles":      sharedMimeTypes,
	"noteFiles":          sharedMimeTypes,
	"exportFiles":        {"text/csv", "text/plain; charset=utf-8", "text/plain", "application/json", "application/json; charset=utf-8"},
	"procedureFile":      importSchemaMimeTypes,
	"policyFile":         importSchemaMimeTypes,
	"trustCenterDocFile": {"application/pdf"},
	"templateFiles":      sharedMimeTypes,
}

// MimeTypeValidator returns a storage.ValidationFunc enforcing the configured mime-type set per form field.
var MimeTypeValidator storage.ValidationFunc = func(f storage.File) error {
	if mimes, ok := validMimeTypes[f.FieldName]; ok {
		return pkgobjects.MimeTypeValidator(mimes...)(f)
	}

	return pkgobjects.MimeTypeValidator(sharedMimeTypes...)(f)
}
