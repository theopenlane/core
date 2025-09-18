package objects

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/objects/storage"
)

func TestApplicationMimeTypeValidator(t *testing.T) {
	tests := []struct {
		name    string
		opts    *storage.UploadOptions
		wantErr bool
	}{
		{
			name: "Valid mime type for avatarFile",
			opts: &storage.UploadOptions{
				FileName:    "avatarFile",
				ContentType: "image/jpeg",
			},
			wantErr: false,
		},
		{
			name: "Invalid mime type for avatarFile",
			opts: &storage.UploadOptions{
				FileName:    "avatarFile",
				ContentType: "application/pdf",
			},
			wantErr: true,
		},
		{
			name: "Valid mime type for logoFile with SVG",
			opts: &storage.UploadOptions{
				FileName:    "logoFile",
				ContentType: "image/svg+xml",
			},
			wantErr: false,
		},
		{
			name: "Valid mime type for evidenceFiles with PDF",
			opts: &storage.UploadOptions{
				FileName:    "evidenceFiles",
				ContentType: "application/pdf",
			},
			wantErr: false,
		},
		{
			name: "No mime type validation for unknown key uses sharedMimeTypes",
			opts: &storage.UploadOptions{
				FileName:    "unknownKey",
				ContentType: "application/pdf",
			},
			wantErr: false,
		},
		{
			name: "Invalid mime type for unknown key",
			opts: &storage.UploadOptions{
				FileName:    "unknownKey",
				ContentType: "application/x-malicious",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplicationMimeTypeValidator(context.Background(), tt.opts)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}