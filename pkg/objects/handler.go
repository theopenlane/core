package objects

import (
	"fmt"
	"net/http"

	"golang.org/x/sync/errgroup"
)

// Upload is a HTTP middleware that takes in a list of form fields and the next
// HTTP handler to run after the upload process is completed
func (h *Objects) UploadHandler(keys ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, h.MaxSize)

			if err := r.ParseMultipartForm(h.MaxSize); err != nil {
				h.ErrorResponseHandler(err).ServeHTTP(w, r)

				return
			}

			var wg errgroup.Group

			// pre-make the number of File structs based on the number of keys found after ranging over the http headers
			uploadedFiles := make(Files, len(keys))

			for _, key := range keys {
				wg.Go(func() error {
					return upload(r, h, key, uploadedFiles)
				})
			}

			if err := wg.Wait(); err != nil {
				h.ErrorResponseHandler(err).ServeHTTP(w, r)

				return
			}

			// write the uploaded files to the context
			r = r.WithContext(WriteFilesToContext(r.Context(), uploadedFiles))

			next.ServeHTTP(w, r)
		})
	}
}

func upload(r *http.Request, h *Objects, key string, uploadedFiles Files) error {
	fileHeaders, ok := r.MultipartForm.File[key]
	if !ok {
		if h.IgnoreNonExistentKeys {
			return nil
		}

		return fmt.Errorf("%w: %s", ErrFilesNotFound, key)
	}

	uploadedFiles[key] = make([]File, 0, len(fileHeaders))

	for _, header := range fileHeaders {
		f, err := header.Open()
		if err != nil {
			return fmt.Errorf("%w (%s): %v", ErrFileOpenFailed, key, err)
		}

		defer f.Close()

		uploadedFileName := h.NameFuncGenerator(header.Filename)

		mimeType, err := DetectContentType(f)
		if err != nil {
			return fmt.Errorf("%w (%s): %v", ErrInvalidMimeType, key, err)
		}

		fileData := File{
			FieldName:        key,
			OriginalName:     header.Filename,
			UploadedFileName: uploadedFileName,
			MimeType:         mimeType,
		}

		if err := h.ValidationFunc(fileData); err != nil {
			return fmt.Errorf("%w (%s): %v", ErrValidationFailed, key, err)
		}

		metadata, err := h.Storage.Upload(r.Context(), f, &UploadFileOptions{
			FileName: uploadedFileName,
		})
		if err != nil {
			return fmt.Errorf("%w: upload failed for (%s)", err, key)
		}

		fileData.Size = metadata.Size
		fileData.FolderDestination = metadata.FolderDestination
		fileData.StorageKey = metadata.Key

		uploadedFiles[key] = append(uploadedFiles[key], fileData)
	}

	return nil
}
