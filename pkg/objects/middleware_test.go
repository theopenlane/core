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
	"github.com/stretchr/testify/require"
	"github.com/test-go/testify/mock"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/pkg/objects"
	mocks "github.com/theopenlane/core/pkg/objects/mocks"
)

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

			handler, err := objects.New(opts...)
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

			upload := &objects.Upload{
				ObjectStorage: handler,
				Uploader:      mockUploader(),
			}

			objects.FileUploadMiddleware(objects.UploadConfig{
				Keys:   []string{"form-field"},
				Upload: upload,
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func mockUploader() func(ctx context.Context, u *objects.Upload, files []objects.FileUpload) ([]objects.File, error) {
	return func(ctx context.Context, u *objects.Upload, files []objects.FileUpload) ([]objects.File, error) {
		uploadedFiles := make([]objects.File, 0, len(files))

		for _, f := range files {
			fileID := ulids.New().String()
			uploadedFileName := u.ObjectStorage.NameFuncGenerator(fileID + "_" + f.Filename)

			mimeType, err := objects.DetectContentType(f.File)
			if err != nil {
				return nil, err
			}

			fileData := objects.File{
				ID:               fileID,
				FieldName:        f.Key,
				OriginalName:     f.Filename,
				UploadedFileName: uploadedFileName,
				MimeType:         mimeType,
			}

			// validate the file
			if err := u.ObjectStorage.ValidationFunc(fileData); err != nil {
				return nil, err
			}

			metadata, err := u.ObjectStorage.Storage.Upload(ctx, files[0].File, &objects.UploadFileOptions{
				FileName: uploadedFileName,
			})
			if err != nil {
				return nil, err
			}

			// add metadata to file information
			fileData.Size = metadata.Size
			fileData.FolderDestination = metadata.FolderDestination
			fileData.StorageKey = metadata.Key

			uploadedFiles = append(uploadedFiles, fileData)
		}

		return uploadedFiles, nil
	}
}
