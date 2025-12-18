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
