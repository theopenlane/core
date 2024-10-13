package objects_test

import (
	"bytes"
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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/mocks"
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

func TestObjects(t *testing.T) {
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
	}{
		{
			name:        "uploading succeeds",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&objects.UploadedFileMetadata{
						Size: size,
					}, nil).
					Times(1)
			},
			expectedStatusCode: http.StatusAccepted,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"text/markdown", "text/plain", "text/plain; charset=utf-8"},
		},
		{
			name:        "upload fails because form field does not exist",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&objects.UploadedFileMetadata{
						Size: size,
					}, errors.New("could not upload file")). // nolint:err113
					Times(0)                                 // make sure this is never called
			},
			expectedStatusCode: http.StatusInternalServerError,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"image/png", "application/pdf"},
			ignoreFormField:    true,
		},
		{
			// this test case will use the WithIgnore option
			name:        "upload middleware succeeds even if the form field does not exist",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&objects.UploadedFileMetadata{
						Size: size,
					}, errors.New("could not upload file")). // nolint:err113
					Times(0)                                 // make sure this is never called
			},
			expectedStatusCode: http.StatusAccepted,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"image/png", "application/pdf"},
			ignoreFormField:    true,
			useIgnoreSkipOpt:   true,
		},
		{
			name:        "upload fails because of mimetype validation constraints",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&objects.UploadedFileMetadata{
						Size: size,
					}, errors.New("could not upload file")). // nolint:err113
					Times(0)                                 // make sure this is never called
			},
			expectedStatusCode: http.StatusInternalServerError,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"image/png", "application/pdf"},
		},
		{
			name:        "upload fails because of storage layer",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&objects.UploadedFileMetadata{
						Size: size,
					}, errors.New("could not upload file")). // nolint:err113
					Times(1)
			},
			expectedStatusCode: http.StatusInternalServerError,
			pathToFile:         "objects.md",
			validMimeTypes:     []string{"text/markdown", "text/plain", "text/plain; charset=utf-8"},
		},
		{
			name:        "upload fails because file is too large",
			maxFileSize: 1024,
			fn: func(store *mocks.MockStorage, size int64) {
				store.EXPECT().
					Upload(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&objects.UploadedFileMetadata{
						Size: size,
					}, errors.New("could not upload file")). // nolint:err113
					Times(0)                                 // never call this
			},
			expectedStatusCode: http.StatusInternalServerError,
			pathToFile:         "image.jpg",
			validMimeTypes:     []string{"image/jpeg"},
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := mocks.NewMockStorage(ctrl)

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

			handler.UploadHandler("form-field")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if v.useIgnoreSkipOpt {
					w.WriteHeader(http.StatusAccepted)
					fmt.Fprintf(w, "skipping check since we did not upload any file")
					return
				}

				file, err := objects.FilesFromContextWithKey(r.Context(), "form-field")

				require.NoError(t, err)

				require.Equal(t, v.pathToFile, file[0].OriginalName)

				w.WriteHeader(http.StatusAccepted)
				fmt.Fprintf(w, "successfully uploaded the file")
			})).ServeHTTP(recorder, r)

			result := recorder.Result()

			respBody := result.Body
			defer respBody.Close()

			require.Equal(t, v.expectedStatusCode, result.StatusCode)

			verifyMatch(t, recorder)
		})
	}
}
