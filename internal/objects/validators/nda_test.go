package validators

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/internal/testutils/pdftest"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

func TestNDAValidator(t *testing.T) {
	plain := pdftest.MinimalPDF(t)

	tests := []struct {
		name    string
		file    storage.File
		wantErr error
	}{
		{name: "plain pdf", file: ndaUpload(plain, pdfMimeType)},
		{name: "owner locked pdf", file: ndaUpload(pdftest.EncryptPDF(t, plain, ""), pdfMimeType)},
		{name: "password protected pdf", file: ndaUpload(pdftest.EncryptPDF(t, plain, "secret"), pdfMimeType), wantErr: ErrPasswordProtectedNDA},
		{name: "password protected pdf without nda hint", file: storage.File{RawFile: bytes.NewReader(pdftest.EncryptPDF(t, plain, "secret")), FileMetadata: storage.FileMetadata{ContentType: pdfMimeType}}},
		{name: "non pdf nda", file: ndaUpload([]byte("not a pdf"), "text/plain"), wantErr: pkgobjects.ErrUnsupportedMimeType},
		{name: "no reader", file: func() storage.File { f := ndaUpload(nil, pdfMimeType); f.RawFile = nil; return f }()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := NDAValidator(tc.file)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func ndaUpload(content []byte, contentType string) storage.File {
	file := storage.File{
		OriginalName: "nda.pdf",
		FieldName:    "templateFiles",
		RawFile:      bytes.NewReader(content),
		FileMetadata: storage.FileMetadata{ContentType: contentType},
	}

	objects.SetTemplateKindHint(&file, enums.TemplateKindTrustCenterNda)

	return file
}
