package objects_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/objects"
	mocks "github.com/theopenlane/core/pkg/objects/mocks"
)

func TestFileUploadMiddleware(t *testing.T) {
	tt := []struct {
		name               string
		maxFileSize        int64
		pathToFile         string
		fn                 func(store *mocks.MockStorage, size int64)
		expectedStatusCode int
		validMimeTypes     []string
		// ignoreFormField instructs the test to not add the
		// multipart form data part to the request
		ignoreFormField bool

		useIgnoreSkipOpt bool
		uploadExpected   bool
	}{
		{
			name:        "uploading succeeds",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(context.Background(), mock.Anything, mock.Anything).
					Return(&objects.UploadedFileMetadata{
						Size: size,
					}, nil).
					Once()
			},
			expectedStatusCode: http.StatusAccepted,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"text/markdown", "text/plain", "text/plain; charset=utf-8"},
			uploadExpected:     true,
		},
		{
			name:        "upload fails because form field does not exist",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				// leave empty because we should not be calling Upload
			},
			expectedStatusCode: http.StatusOK,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"image/png", "application/pdf"},
			ignoreFormField:    true,
			uploadExpected:     false,
		},
		{
			// this test case will use the WithIgnore option
			name:        "upload middleware succeeds even if the form field does not exist",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				// leave empty because we should not be calling Upload
			},
			expectedStatusCode: http.StatusAccepted,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"image/png", "application/pdf"},
			ignoreFormField:    true,
			useIgnoreSkipOpt:   true,
			uploadExpected:     false,
		},
		{
			name:        "upload fails because of mimetype validation constraints",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				// leave empty because we should not be calling Upload
			},
			expectedStatusCode: http.StatusBadRequest,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"image/png", "application/pdf"},
			uploadExpected:     false,
		},
		{
			name:        "upload fails because of storage layer",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(context.Background(), mock.Anything, mock.Anything).
					Return(&objects.UploadedFileMetadata{
						Size: size,
					}, errors.New("could not upload file to storage backend")). // nolint:err113
					Times(1)
			},
			expectedStatusCode: http.StatusBadRequest,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"text/markdown", "text/plain", "text/plain; charset=utf-8"},
			uploadExpected:     false,
		},
		{
			name:        "upload fails because file is too large",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				// leave empty because we should not be calling Upload
			},
			expectedStatusCode: http.StatusBadRequest,
			pathToFile:         "image.jpg",
			validMimeTypes:     []string{"image/jpeg"},
			uploadExpected:     false,
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			storage := mocks.NewMockStorage(t)

			opts := []objects.Option{
				objects.WithMaxFileSize(v.maxFileSize),
				objects.WithStorage(storage),
				objects.WithValidationFunc(objects.MimeTypeValidator(v.validMimeTypes...)),
			}

			if v.useIgnoreSkipOpt {
				opts = append(opts, objects.WithIgnoreNonExistentKey(true))
			}

			objectHandler, err := objects.New(opts...)
			require.NoError(t, err)

			buffer := bytes.NewBuffer(nil)

			multipartWriter := multipart.NewWriter(buffer)

			var formFieldWriter io.Writer = bytes.NewBuffer(nil)

			if !v.ignoreFormField {
				var err error
				formFieldWriter, err = multipartWriter.CreateFormFile("form-field", v.pathToFile)
				require.NoError(t, err)
			}

			fileToUpload, err := os.Open(filepath.Join("testdata", v.pathToFile))
			require.NoError(t, err)

			n, err := io.Copy(formFieldWriter, fileToUpload)
			require.NoError(t, err)

			v.fn(storage, int64(n))

			require.NoError(t, multipartWriter.Close())

			recorder := httptest.NewRecorder()

			r := httptest.NewRequest(http.MethodPatch, "/", buffer)
			r.Header.Set("Content-Type", multipartWriter.FormDataContentType())

			// set the form field key
			objectHandler.Keys = []string{"form-field"}

			objects.FileUploadMiddleware(objectHandler)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if v.useIgnoreSkipOpt {
					w.WriteHeader(http.StatusAccepted)
					fmt.Fprintf(w, "skipping check since we did not upload any file")
					return
				}

				file, err := objects.FilesFromContextWithKey(r.Context(), "form-field")
				if v.uploadExpected {
					require.NoError(t, err)

					require.NoError(t, err)
					require.Len(t, file, 1)

					assert.Equal(t, v.pathToFile, file[0].OriginalName)

					w.WriteHeader(http.StatusAccepted)
					fmt.Fprintf(w, "successfully uploaded the file")
				} else {
					require.Error(t, err)
				}
			})).ServeHTTP(recorder, r)

			result := recorder.Result()

			respBody := result.Body
			defer respBody.Close()

			require.Equal(t, v.expectedStatusCode, result.StatusCode)

			verifyMatch(t, recorder)
		})
	}
}

func TestCreateURI(t *testing.T) {
	tests := []struct {
		scheme      string
		destination string
		key         string
		expectedURI string
	}{
		{
			scheme:      "https://",
			destination: "example.com",
			key:         "file123",
			expectedURI: "https://example.com/file123",
		},
		{
			scheme:      "s3://",
			destination: "bucket",
			key:         "upload/file456",
			expectedURI: "s3://bucket/upload/file456",
		},
		{
			scheme:      "ftp://",
			destination: "ftp.example.com",
			key:         "dir/file789",
			expectedURI: "ftp://ftp.example.com/dir/file789",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s%s/%s", tt.scheme, tt.destination, tt.key), func(t *testing.T) {
			uri := objects.CreateURI(tt.scheme, tt.destination, tt.key)
			assert.Equal(t, tt.expectedURI, uri)
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		size         int64
		expectedSize string
	}{
		{
			size:         1024,
			expectedSize: "1.00 KB",
		},
		{
			size:         1048576,
			expectedSize: "1.00 MB",
		},
		{
			size:         1073741824,
			expectedSize: "1.00 GB",
		},
		{
			size:         512,
			expectedSize: "512 bytes",
		},
		{
			size:         1536,
			expectedSize: "1.50 KB",
		},
		{
			size:         1572864,
			expectedSize: "1.50 MB",
		},
		{
			size:         1610612736,
			expectedSize: "1.50 GB",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d bytes", tt.size), func(t *testing.T) {
			result := objects.FormatFileSize(tt.size)
			assert.Equal(t, tt.expectedSize, result)
		})
	}
}

func verifyMatch(t *testing.T, v interface{}) {
	g := goldie.New(t, goldie.WithFixtureDir("./testdata/golden"))

	b := new(bytes.Buffer)

	var err error

	if d, ok := v.(*httptest.ResponseRecorder); ok {
		_, err = io.Copy(b, d.Body)
	} else {
		err = json.NewEncoder(b).Encode(v)
	}

	require.NoError(t, err)
	g.Assert(t, t.Name(), b.Bytes())
}
