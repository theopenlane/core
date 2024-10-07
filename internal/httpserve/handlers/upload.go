package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"

	echo "github.com/theopenlane/echox"
	"golang.org/x/sync/errgroup"

	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/objects"
)

func bindRequest(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return err
	}
	return c.Validate(req)
}

type uploadFilesRequest struct {
	Files []*multipart.FileHeader `form:"files"`
}

func readFile(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, src); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readFiles(files []*multipart.FileHeader) (filesBytes [][]byte, err error) {
	var wg sync.WaitGroup
	ch := make(chan []byte, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()
			f, err := readFile(file)
			if err != nil {
				return
			}
			ch <- f
		}(file)
	}
	wg.Wait()
	close(ch)

	var fb [][]byte
	for file := range ch {
		fb = append(fb, file)
	}
	return fb, err
}

func toJson[T any](obj T) string {
	b, _ := json.Marshal(obj)
	return string(b)
}

func (h *Handler) UploadFiles(c echo.Context) error {
	var req uploadFilesRequest

	if err := bindRequest(c, &req); err != nil {
		log.Error().Err(err).Msg("failed to bind request")
		return h.InvalidInput(c, err)
	}

	files, err := readFiles(req.Files)
	if err != nil {
		log.Error().Err(err).Msg("failed to read files")
		return h.InternalServerError(c, err)
	}

	if err := h.ObjectStorage.Storage.ManagerUpload(c.Request().Context(), files); err != nil {
		log.Error().Err(err).Msg("failed to upload files")
		return h.InternalServerError(c, err)
	}

	out := models.UploadFilesReply{
		Message: "Files uploaded successfully",
	}

	return h.Success(c, out)
}

// FileUploadHandler is responsible for uploading files
func (h *Handler) FileUploadHandler(ctx echo.Context, keys ...string) error {
	var in models.UploadFilesRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	r := ctx.Request()
	w := ctx.Response()

	r.Body = http.MaxBytesReader(w, r.Body, h.ObjectStorage.MaxSize)

	err := r.ParseMultipartForm(h.ObjectStorage.MaxSize)
	if err != nil {
		h.ObjectStorage.ErrorResponseHandler(err).ServeHTTP(w, r)
		return err
	}

	var wg errgroup.Group

	out := models.UploadFilesReply{}

	uploadedFiles := make(objects.Files, len(keys))

	for _, key := range keys {
		key := key

		wg.Go(func() error {
			fileHeaders, ok := r.MultipartForm.File[key]
			if !ok {
				if h.ObjectStorage.IgnoreNonExistentKeys {
					return nil
				}

				log.Error().Str("key", key).Msg("files not found")

				return err
			}

			uploadedFiles[key] = make([]objects.File, 0, len(fileHeaders))

			for _, header := range fileHeaders {
				f, err := header.Open()
				if err != nil {
					log.Error().Err(err).Str("key", key).Msg("failed to open file")
					return err
				}

				defer f.Close()

				mimeType, err := objects.DetectContentType(f)
				if err != nil {
					log.Error().Err(err).Str("key", key).Msg("failed to fetch content type")
					return err
				}

				set := ent.CreateFileInput{
					ProvidedFileName:      header.Filename,
					ProvidedFileExtension: mimeType,
					DetectedContentType:   formatFileSize(header.Size),
				}

				entfile, err := transaction.FromContext(reqCtx).File.Create().SetInput(set).Save(reqCtx)
				if err != nil {
					log.Error().Err(err).Msg("failed to create file")
					return err
				}

				uploadedFileName := h.ObjectStorage.NameFuncGenerator(entfile.ID + "_" + header.Filename)
				fileData := objects.File{
					FieldName:        key,
					OriginalName:     header.Filename,
					UploadedFileName: uploadedFileName,
					MimeType:         mimeType,
				}

				if err := h.ObjectStorage.ValidationFunc(fileData); err != nil {
					log.Error().Err(err).Str("key", key).Msg("failed to validate file")
					return err
				}

				metadata, err := h.Storage.Upload(r.Context(), f, &objects.UploadFileOptions{
					FileName:    uploadedFileName,
					ContentType: mimeType,
					Metadata: map[string]string{
						"file_id": entfile.ID,
					},
				})
				if err != nil {
					log.Error().Err(err).Str("key", key).Msg("failed to upload file")
					return err
				}

				presignedurl := h.Storage.GetPresignedURL(context.TODO(), uploadedFileName)

				newfile := objects.File{
					ID:                entfile.ID,
					Name:              header.Filename,
					MimeType:          mimeType,
					ProvidedExtension: filepath.Ext(header.Filename),
					Size:              header.Size,
					PresignedURL:      presignedurl,
				}

				out.FileIdentifiers = append(out.FileIdentifiers, newfile.ID)
				out.PresignedURL = presignedurl
				out.Message = "File uploaded successfully, god damn matt you're a beautiful bastard"
				out.FileCount++

				fileData.PresignedURL = presignedurl
				fileData.Size = metadata.Size
				fileData.FolderDestination = metadata.FolderDestination
				fileData.StorageKey = metadata.Key

				log.Info().Str("file", fileData.UploadedFileName).Msg("file uploaded")
				log.Info().Str("file", fileData.UploadedFileName).Str("id", fileData.FolderDestination).Msg("ent file ID")
				log.Info().Str("file", fileData.UploadedFileName).Str("mime_type", fileData.MimeType).Msg("detected mime type")
				log.Info().Str("file", fileData.UploadedFileName).Str("size", formatFileSize(fileData.Size)).Msg("calculated file size")
				log.Info().Str("file", fileData.UploadedFileName).Str("presigned_url", fileData.PresignedURL).Msg("presigned URL")

				uploadedFiles[key] = append(uploadedFiles[key], fileData)
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		h.ObjectStorage.ErrorResponseHandler(err).ServeHTTP(w, r)
		return err
	}

	r = r.WithContext(objects.WriteFilesToContext(r.Context(), uploadedFiles))

	return h.SuccessBlob(ctx, out)
}

const MaxUploadSize = 32 * 1024 * 1024 // 32MB

func (h *Handler) CreateFile(ctx context.Context, input ent.CreateFileInput) (*ent.File, error) {
	file, err := transaction.FromContext(ctx).File.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, err
	}

	return file, err
}

// Progress is used to track the progress of a file upload
// It implements the io.Writer interface so it can be passed to an io.TeeReader()
type Progress struct {
	TotalSize int64
	BytesRead int64
}

// Write is used to satisfy the io.Writer interface Instead of writing somewhere, it simply aggregates the total bytes on each read
func (pr *Progress) Write(p []byte) (n int, err error) {
	n, err = len(p), nil

	pr.BytesRead += int64(n)

	pr.Print()

	return
}

// Print displays the current progress of the file upload
func (pr *Progress) Print() {
	if pr.BytesRead == pr.TotalSize {
		fmt.Println("DONE!")

		return
	}

	fmt.Printf("File upload in progress: %d\n", pr.BytesRead)
}

// formatFileSize converts a file size in bytes to a human-readable string in MB/GB notation.
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}
