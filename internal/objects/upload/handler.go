package upload

import (
	"context"
	"io"
	"path"
	"strings"
	"time"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/objects/store"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
)

// HandleUploads persists metadata, uploads files to storage, and enriches the request context with uploaded file details.
func HandleUploads(ctx context.Context, svc *objects.Service, files []pkgobjects.File) (context.Context, []pkgobjects.File, error) {
	if len(files) == 0 {
		return ctx, nil, nil
	}

	var uploadedFiles []pkgobjects.File

	for _, file := range files {
		pkgobjects.AddUpload()
		metrics.StartFileUpload()
		startTime := time.Now()

		finish := func(status string) {
			metrics.FinishFileUpload(status, time.Since(startTime).Seconds())
			pkgobjects.DoneUpload()
		}

		// we are intentionally swallowing this error because if we can't get the org ID
		// we just won't populate the provider hints with it. The upload can still proceed
		// without it - failing at this stage prevents the upload from ever progressing
		orgID, _ := auth.GetOrganizationIDFromContext(ctx)
		if orgID != "" && file.Parent.ID == "" && file.CorrelatedObjectID == "" && file.CorrelatedObjectType == "" {
			file.CorrelatedObjectID = orgID
			file.CorrelatedObjectType = "organization"
		}

		// Normalize metadata (content type, hints) before we persist the file record so
		// downstream storage providers see consistent values.
		uploadOpts := BuildUploadOptions(ctx, &file)

		entFile, err := store.CreateFileRecord(ctx, file)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("file", file.OriginalName).Msg("failed to create file record")
			finish("error")

			return ctx, nil, err
		}

		folderPath := path.Join(orgID, entFile.ID)
		uploadOpts.FolderDestination = folderPath
		file.Folder = folderPath

		if uploadOpts.ProviderHints == nil {
			uploadOpts.ProviderHints = &pkgobjects.ProviderHints{}
		}

		if uploadOpts.ProviderHints.Metadata == nil {
			uploadOpts.ProviderHints.Metadata = map[string]string{}
		}

		uploadOpts.ProviderHints.Metadata["file_id"] = entFile.ID

		uploadedFile, err := svc.Upload(ctx, file.RawFile, uploadOpts)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("file", file.OriginalName).Msg("failed to upload file")
			finish("error")

			return ctx, nil, err
		}

		if closer, ok := file.RawFile.(io.Closer); ok {
			_ = closer.Close()
		}

		mergeUploadedFileMetadata(uploadedFile, entFile.ID, file)
		if err := store.UpdateFileWithStorageMetadata(ctx, entFile, *uploadedFile); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to update file metadata")
			finish("error")

			return ctx, nil, err
		}

		uploadedFiles = append(uploadedFiles, *uploadedFile)
		finish("success")
	}

	if len(uploadedFiles) == 0 {
		return ctx, nil, nil
	}

	contextFilesMap := make(pkgobjects.Files)
	for _, file := range uploadedFiles {
		fieldName := file.FieldName
		if fieldName == "" {
			fieldName = "uploads"
		}

		contextFilesMap[fieldName] = append(contextFilesMap[fieldName], file)
	}

	ctx = pkgobjects.WriteFilesToContext(ctx, contextFilesMap)
	return ctx, uploadedFiles, nil
}

// HandleRollback removes uploaded files from storage in case of an error during processing the rest of the request.
func HandleRollback(ctx context.Context, svc *objects.Service, files []pkgobjects.File) {
	if len(files) == 0 {
		return
	}

	// rollback the file uploads in case of an error
	for _, file := range files {
		if err := svc.Delete(ctx, &storagetypes.File{
			ID:                file.ID,
			ProviderType:      file.ProviderType,
			OriginalName:      file.OriginalName,
			MD5:               file.MD5,
			ProvidedExtension: file.ProvidedExtension,
			Parent:            file.Parent,
			FileMetadata: storagetypes.FileMetadata{
				Bucket:       file.Bucket,
				Key:          file.Key,
				Folder:       file.Folder,
				ProviderType: file.ProviderType,
			},
		}, &storagetypes.DeleteFileOptions{
			Reason: "rolling back file upload due to mutation error",
		}); err != nil {
			// intentionally continue deleting other files even if one fails because this
			// is a best-effort cleanup process for a failed request
			logx.FromContext(ctx).Error().Err(err).Msg("failed to delete uploaded file during rollback")
		}
	}
}

// BuildUploadOptions prepares upload options enriched with provider hints and ensures
// the file has a detected content type when one was not provided by the client.
func BuildUploadOptions(ctx context.Context, f *pkgobjects.File) *pkgobjects.UploadOptions {
	if f == nil {
		return &pkgobjects.UploadOptions{}
	}

	if f.ProviderHints == nil {
		f.ProviderHints = &pkgobjects.ProviderHints{}
	}

	orgID, _ := auth.GetOrganizationIDFromContext(ctx)
	objects.PopulateProviderHints(f, orgID)

	contentType := f.ContentType
	if contentType == "" || strings.EqualFold(contentType, "application/octet-stream") {
		// When we buffer the upload we lose any stream-specific metadata, so detect the MIME now
		// and swap in the buffered reader so the downstream provider still has access to the data.
		if f.RawFile != nil {
			if detected, err := storage.DetectContentType(f.RawFile); err == nil && detected != "" {
				contentType = detected
			} else if buffered, err := pkgobjects.NewBufferedReaderFromReader(f.RawFile); err == nil {
				if detected, err := storage.DetectContentType(buffered); err == nil && detected != "" {
					contentType = detected
				}
				// Replace the original reader so the upload pipeline can still stream the contents.
				f.RawFile = buffered
			}
		}

		if contentType != "" {
			f.ContentType = contentType
		}
	}

	return &pkgobjects.UploadOptions{
		FileName:          f.OriginalName,
		ContentType:       contentType,
		Bucket:            f.Bucket,
		FolderDestination: f.Folder,
		FileMetadata: pkgobjects.FileMetadata{
			Key:           f.FieldName,
			ProviderHints: f.ProviderHints,
		},
	}
}

func mergeUploadedFileMetadata(dest *pkgobjects.File, entFileID string, src pkgobjects.File) {
	dest.ID = entFileID
	dest.FieldName = src.FieldName
	dest.Parent = src.Parent
	dest.CorrelatedObjectID = src.CorrelatedObjectID
	dest.CorrelatedObjectType = src.CorrelatedObjectType
	if len(dest.Metadata) == 0 && len(src.Metadata) > 0 {
		dest.Metadata = src.Metadata
	}
}
