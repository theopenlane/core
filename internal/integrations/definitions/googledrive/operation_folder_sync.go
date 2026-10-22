package googledrive

import (
	"context"

	"google.golang.org/api/drive/v3"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

const (
	// folderSyncPageSize is the number of files to request per page
	folderSyncPageSize = int64(100)
	// googleDocMIMEType is the MIME type for Google Docs
	googleDocMIMEType = "application/vnd.google-apps.document"
)

// folderSyncOperation is the operation ref for the folder sync ingest operation
var folderSyncOperation = types.OperationRefOf[FolderSync]()

// FolderSync lists Google Docs in a folder and emits ingest envelopes
type FolderSync struct {
	types.OperationSettings
	// FolderID is the Google Drive folder ID or URL containing policy documents
	FolderID string `json:"folderId,omitempty" jsonschema:"title=Folder ID,description=Google Drive folder ID or URL containing policy documents"`
}

// runFolderSync collects the Google Docs selected in the Picker, or every Google Doc in the configured folder, and returns ingest payload sets
func runFolderSync(ctx context.Context, req types.OperationRequest, svc DriveClient, cfg FolderSync) ([]types.IngestPayloadSet, error) {
	metadata, err := installation.Decode(req.Integration)
	if err != nil {
		return nil, ErrInstallationMetadataDecode
	}

	var files []*drive.File

	if len(metadata.SelectedDocuments) > 0 {
		files, err = getSelectedDocs(ctx, svc, metadata.SelectedDocuments)
	} else {
		folderID := parseFolderID(cfg.FolderID)
		if folderID == "" {
			logx.FromContext(ctx).Error().Bool("folder_id_present", cfg.FolderID != "").Msg("googledrive: no documents selected and no folder configured, documents need to be selected in the picker")

			return nil, ErrNoDocumentsSelected
		}

		files, err = listFolderDocs(ctx, svc, folderID)
	}

	if err != nil {
		return nil, err
	}

	envelopes := make([]types.MappingEnvelope, 0, len(files))

	for _, file := range files {
		envelope, err := providerkit.MarshalEnvelope(file.Id, file, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaInternalPolicy.Name,
			Envelopes: envelopes,
		},
	}, nil
}

// getSelectedDocs fetches metadata for each selected file, skipping files that are inaccessible or not Google Docs
func getSelectedDocs(ctx context.Context, c DriveClient, selected []types.SelectedResource) ([]*drive.File, error) {
	files := make([]*drive.File, 0, len(selected))

	for _, resource := range selected {
		// check for error context, e.g. canceled or deadline exceeded before continuing
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		fileID := resource.ID

		file, err := c.Svc.Files.Get(fileID).
			Fields("id,name,mimeType,modifiedTime,createdTime").
			SupportsAllDrives(true).
			Context(ctx).
			Do()
		if err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("file_id", fileID).Msg("googledrive: skipping selected file that could not be fetched")

			continue
		}

		if file.MimeType != googleDocMIMEType {
			logx.FromContext(ctx).Warn().Str("file_id", fileID).Str("mime_type", file.MimeType).Msg("googledrive: skipping selected file that is not a google doc")

			continue
		}

		files = append(files, file)
	}

	return files, nil
}

// listFolderDocs pages through all Google Docs in the specified folder
func listFolderDocs(ctx context.Context, c DriveClient, folderID string) ([]*drive.File, error) {
	var files []*drive.File

	pageToken := ""
	query := "'" + folderID + "' in parents and mimeType = '" + googleDocMIMEType + "' and trashed = false"

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		call := c.Svc.Files.List().
			Q(query).
			PageSize(folderSyncPageSize).
			Fields("nextPageToken,files(id,name,modifiedTime,createdTime)").
			IncludeItemsFromAllDrives(true).
			SupportsAllDrives(true).
			OrderBy("name").
			Context(ctx)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("Failed to list drive files inside of folder")

			return nil, ErrFolderListFailed
		}

		files = append(files, resp.Files...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return files, nil
}
