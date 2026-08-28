//go:build test

package testharness

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/stretchr/testify/mock"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/objects"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
	mock_shared "github.com/theopenlane/core/v2/pkg/objects/mocks"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// test upload fixtures, resolved from the repo root so they do not depend on the
// working directory of whichever suite is running
var (
	LogoFilePath = filepath.Join(repoRoot(), "internal", "graphapi", "testdata", "uploads", "logo.png")
	PdfFilePath  = filepath.Join(repoRoot(), "internal", "graphapi", "testdata", "uploads", "hello.pdf")
	TxtFilePath  = filepath.Join(repoRoot(), "internal", "graphapi", "testdata", "uploads", "hello.txt")
)

// LogoFileFunc creates a graphql.Upload func for the logo.png test file
func LogoFileFunc(t *testing.T) func() *graphql.Upload {
	return func() *graphql.Upload {
		return UploadFile(t, LogoFilePath)
	}
}

// UploadFileFunc creates a graphql.Upload func for the specified file path
func UploadFileFunc(t *testing.T, path string) func() *graphql.Upload {
	return func() *graphql.Upload {
		return UploadFile(t, path)
	}
}

// UploadFile creates a graphql.Upload for the specified file path
func UploadFile(t *testing.T, path string) *graphql.Upload {
	pdfFile, err := storage.NewUploadFile(path)
	assert.NilError(t, err)
	return &graphql.Upload{
		File:        pdfFile.RawFile,
		Filename:    pdfFile.OriginalName,
		Size:        pdfFile.Size,
		ContentType: pdfFile.ContentType,
	}
}

func GetMD5Hash(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	assert.NilError(t, err)

	sum := md5.Sum(data)

	return hex.EncodeToString(sum[:])
}

func TestPDFBytes() []byte {
	page := map[string]any{
		"paper":  "A4P",
		"origin": "UpperLeft",
		"fonts": map[string]any{
			"f": map[string]any{"name": "Helvetica", "size": 12},
		},
		"pages": map[string]any{
			"1": map[string]any{
				"content": map[string]any{
					"text": []map[string]any{
						{"value": "test", "pos": [2]float64{20, 20}, "font": map[string]any{"name": "$f"}},
					},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(page)

	var buf bytes.Buffer
	_ = api.Create(nil, bytes.NewReader(jsonData), &buf, nil)

	return buf.Bytes()
}

func getMockDiskFileMetadata(upload graphql.Upload) pkgobjects.FileMetadata {
	return pkgobjects.FileMetadata{
		Key:          "test-key",
		Size:         upload.Size,
		Folder:       "test-folder",
		Bucket:       "test-bucket",
		ContentType:  upload.ContentType,
		ProviderType: storage.DiskProvider,
		FullURI:      "file:///tmp/test-file",
	}
}

var mockScheme = "file://"

// ExpectUpload sets up the mock object store to expect an upload and related operations
func ExpectUpload(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []graphql.Upload) {
	assert.Assert(t, mockProvider != nil)

	for _, upload := range expectedUploads {
		mockProvider.On("GetScheme").Return(&mockScheme).Once()
		mockProvider.On("ProviderType").Return(storage.DiskProvider).Maybe()
		mockProvider.On("Upload", mock.Anything, mock.Anything, mock.Anything).Return(&storage.UploadedMetadata{
			FileMetadata: getMockDiskFileMetadata(upload),
		}, nil).Once()

		// Allow document hooks to download the just-uploaded content for parsing
		mockProvider.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(&storage.DownloadedMetadata{
			File: TestPDFBytes(),
			Size: upload.Size,
		}, nil).Maybe()
	}
}

// ExpectAttestedUpload sets up mock expectations for the attested PDF upload triggered by attestNDADocument
func ExpectAttestedUpload(t *testing.T, mockProvider *mock_shared.MockProvider) {
	assert.Assert(t, mockProvider != nil)

	mockProvider.On("GetScheme").Return(&mockScheme).Once()
	mockProvider.On("ProviderType").Return(storage.DiskProvider).Maybe()
	mockProvider.On("Upload", mock.Anything, mock.Anything, mock.Anything).Return(&storage.UploadedMetadata{
		FileMetadata: pkgobjects.FileMetadata{
			Key:          "test-key-attested",
			Folder:       "test-folder",
			Bucket:       "test-bucket",
			ContentType:  "application/pdf",
			ProviderType: storage.DiskProvider,
			FullURI:      "file:///tmp/test-file-attested",
		},
	}, nil).Once()
}

func ExpectUploadWithTemplateKind(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []graphql.Upload, kind enums.TemplateKind) {
	assert.Assert(t, mockProvider != nil)

	for _, upload := range expectedUploads {
		mockProvider.On("GetScheme").Return(&mockScheme).Once()
		mockProvider.On("ProviderType").Return(storage.DiskProvider).Maybe()
		uploadOpts := mock.MatchedBy(func(opts *storage.UploadOptions) bool {
			if opts == nil || opts.ProviderHints == nil || opts.ProviderHints.Metadata == nil {
				return false
			}

			return opts.ProviderHints.Metadata[objects.TemplateKindMetadataKey] == kind.String()
		})
		mockProvider.On("Upload", mock.Anything, mock.Anything, uploadOpts).Return(&storage.UploadedMetadata{
			FileMetadata: getMockDiskFileMetadata(upload),
		}, nil).Once()

		// Allow document hooks to download the just-uploaded content for parsing
		mockProvider.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(&storage.DownloadedMetadata{
			File: TestPDFBytes(),
			Size: upload.Size,
		}, nil).Maybe()
	}
}

// ExpectDelete sets up the mock object store to expect a delete and related operations
func ExpectDelete(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []graphql.Upload) {
	assert.Assert(t, mockProvider != nil)

	for range expectedUploads {
		mockProvider.On("GetScheme").Return(&mockScheme).Once()
		mockProvider.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	}
}

// ExpectUploadNillable sets up the mock object store to expect an upload and related operations
func ExpectUploadNillable(t *testing.T, mockProvider *mock_shared.MockProvider, expectedUploads []*graphql.Upload) {
	assert.Check(t, mockProvider != nil)

	for _, upload := range expectedUploads {
		if upload != nil {
			mockProvider.On("GetScheme").Return(&mockScheme).Once()
			mockProvider.On("ProviderType").Return(storage.DiskProvider).Maybe()
			mockProvider.On("Upload", mock.Anything, mock.Anything, mock.Anything).Return(&storage.UploadedMetadata{
				FileMetadata: getMockDiskFileMetadata(*upload),
			}, nil).Once()

			// Allow document hooks to download the just-uploaded content for parsing
			mockProvider.On("Download", mock.Anything, mock.Anything, mock.Anything).Return(&storage.DownloadedMetadata{
				File: []byte("test content"),
				Size: upload.Size,
			}, nil).Maybe()
		}
	}
}

// ExpectUploadCheckOnly sets up the mock object store to expect an upload check only operation
// but fails before the upload is attempted
func ExpectUploadCheckOnly(t *testing.T, mockProvider *mock_shared.MockProvider) {
	assert.Assert(t, mockProvider != nil)

	mockProvider.On("GetScheme").Return(&mockScheme).Once()
}
