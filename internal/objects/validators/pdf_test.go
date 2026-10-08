package validators

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/v2/internal/testutils/pdftest"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

func TestIsPasswordProtected(t *testing.T) {
	plain := pdftest.MinimalPDF(t)

	tests := []struct {
		name string
		file storage.File
		want bool
	}{
		{name: "plain pdf", file: storage.File{RawFile: bytes.NewReader(plain)}},
		{name: "owner locked pdf", file: storage.File{RawFile: bytes.NewReader(pdftest.EncryptPDF(t, plain, ""))}},
		{name: "user password pdf", file: storage.File{RawFile: bytes.NewReader(pdftest.EncryptPDF(t, plain, "secret"))}, want: true},
		{name: "not a pdf", file: storage.File{RawFile: bytes.NewReader([]byte("not a pdf"))}},
		{name: "no reader", file: storage.File{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isPasswordProtected(tc.file))
		})
	}
}

func TestPasswordProtectedValidator(t *testing.T) {
	locked := pdftest.EncryptPDF(t, pdftest.MinimalPDF(t), "secret")

	tests := []struct {
		name    string
		file    storage.File
		wantErr error
	}{
		{name: "readable pdf on a guarded field", file: upload(scanFilesField, pdftest.MinimalPDF(t))},
		{name: "password protected pdf on a guarded field", file: upload(scanFilesField, locked), wantErr: ErrPasswordProtectedPDF},
		{name: "password protected pdf on an unguarded field", file: upload("evidenceFiles", locked)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := PasswordProtectedValidator(tc.file)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func upload(field string, content []byte) storage.File {
	return storage.File{
		OriginalName: "report.pdf",
		FieldName:    field,
		RawFile:      bytes.NewReader(content),
		FileMetadata: storage.FileMetadata{ContentType: pdfMimeType},
	}
}
