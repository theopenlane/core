package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestMimeTypeValidatorFieldSpecific(t *testing.T) {
	t.Run("allows configured avatar file types", func(t *testing.T) {
		file := storage.File{
			FieldName: "avatarFile",
			FileMetadata: storage.FileMetadata{
				ContentType: "image/png",
			},
		}

		assert.NoError(t, mimeTypeValidator(file))
	})

	t.Run("rejects disallowed avatar mime types", func(t *testing.T) {
		file := storage.File{
			FieldName: "avatarFile",
			FileMetadata: storage.FileMetadata{
				ContentType: "text/plain",
			},
		}

		assert.Error(t, mimeTypeValidator(file))
	})
}

func TestMimeTypeValidatorSharedFallback(t *testing.T) {
	t.Run("uses shared defaults for unknown fields", func(t *testing.T) {
		file := storage.File{
			FieldName: "customUpload",
			FileMetadata: storage.FileMetadata{
				ContentType: "application/pdf",
			},
		}

		assert.NoError(t, mimeTypeValidator(file))
	})

	t.Run("shared defaults reject unsupported types", func(t *testing.T) {
		file := storage.File{
			FieldName: "customUpload",
			FileMetadata: storage.FileMetadata{
				ContentType: "application/x-malware",
			},
		}

		assert.Error(t, mimeTypeValidator(file))
	})
}

// TestMimeTypeValidatorImportSchemaFields locks down the allowlist for the
// fields that share importSchemaMimeTypes (internalPolicyFile, procedureFile,
// actionPlanFile). This is a security boundary — accidentally widening it
// would let users upload formats the import pipeline cannot safely process.
func TestMimeTypeValidatorImportSchemaFields(t *testing.T) {
	importSchemaFields := []string{
		"internalPolicyFile",
		"procedureFile",
		"actionPlanFile",
	}

	allowedTypes := []string{
		"text/plain",
		"text/plain; charset=utf-8",
		"text/markdown",
		"text/x-markdown",
		"text/html",
		"application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	}

	rejectedTypes := []string{
		"application/javascript",
		"application/x-msdownload",
		"application/x-sh",
		"image/png",
		"application/zip",
	}

	for _, field := range importSchemaFields {
		for _, mime := range allowedTypes {
			t.Run(field+" allows "+mime, func(t *testing.T) {
				file := storage.File{
					FieldName: field,
					FileMetadata: storage.FileMetadata{
						ContentType: mime,
					},
				}
				assert.NoError(t, mimeTypeValidator(file))
			})
		}

		for _, mime := range rejectedTypes {
			t.Run(field+" rejects "+mime, func(t *testing.T) {
				file := storage.File{
					FieldName: field,
					FileMetadata: storage.FileMetadata{
						ContentType: mime,
					},
				}
				assert.Error(t, mimeTypeValidator(file))
			})
		}
	}
}
