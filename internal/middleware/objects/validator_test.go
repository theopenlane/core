package objects

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/objects"
)

func TestMimeTypeValidator(t *testing.T) {
	tests := []struct {
		name    string
		file    objects.File
		wantErr bool
	}{
		{
			name: "Valid mime type for avatarFile",
			file: objects.File{
				FieldName:   "avatarFile",
				ContentType: "image/jpeg",
			},
			wantErr: false,
		},
		{
			name: "Invalid mime type for avatarFile",
			file: objects.File{
				FieldName:   "avatarFile",
				ContentType: "application/pdf",
			},
			wantErr: true,
		},
		{
			name: "No mime type validation for unknown key",
			file: objects.File{
				FieldName:   "unknownKey",
				ContentType: "application/pdf",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MimeTypeValidator(tt.file)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}
