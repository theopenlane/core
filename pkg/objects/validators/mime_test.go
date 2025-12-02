package validators

import (
	"testing"

	"github.com/stretchr/testify/require"

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

		require.NoError(t, mimeTypeValidator(file))
	})

	t.Run("rejects disallowed avatar mime types", func(t *testing.T) {
		file := storage.File{
			FieldName: "avatarFile",
			FileMetadata: storage.FileMetadata{
				ContentType: "text/plain",
			},
		}

		require.Error(t, mimeTypeValidator(file))
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

		require.NoError(t, mimeTypeValidator(file))
	})

	t.Run("shared defaults reject unsupported types", func(t *testing.T) {
		file := storage.File{
			FieldName: "customUpload",
			FileMetadata: storage.FileMetadata{
				ContentType: "application/x-malware",
			},
		}

		require.Error(t, mimeTypeValidator(file))
	})
}
