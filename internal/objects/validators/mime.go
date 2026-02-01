package validators

import (
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

var importSchemaMimeTypes = []string{
	"text/plain; charset=utf-8", "text/plain",
	"text/markdown",
	"text/x-markdown",
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
	"avatarFile":         {"image/jpeg", "image/png", "image/webp"},
	"logoFile":           {"image/jpeg", "image/png", "image/svg+xml", "image/webp"},
	"faviconFile":        {"image/jpeg", "image/png", "image/x-icon"},
	"evidenceFiles":      sharedMimeTypes,
	"noteFiles":          sharedMimeTypes,
	"exportFiles":        {"text/csv", "text/plain; charset=utf-8", "text/plain", "application/json", "application/json; charset=utf-8"},
	"procedureFile":      importSchemaMimeTypes,
	"internalPolicyFile": importSchemaMimeTypes,
	"actionPlanFile":     importSchemaMimeTypes,
	"trustCenterDocFile": {"application/pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	"documentDataFile":   {"application/pdf"},
	"templateFiles":      sharedMimeTypes,
	"watermarkFile":      {"image/jpeg", "image/png"},
}

// MimeTypeValidator returns a storage.ValidationFunc enforcing the configured mime-type set per form field.
func mimeTypeValidator(f storage.File) error {
	if mimes, ok := validMimeTypes[f.FieldName]; ok {
		return pkgobjects.MimeTypeValidator(mimes...)(f)
	}

	return pkgobjects.MimeTypeValidator(sharedMimeTypes...)(f)
}

var MimeTypeValidator storage.ValidationFunc = mimeTypeValidator
