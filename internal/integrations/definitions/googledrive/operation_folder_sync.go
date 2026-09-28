package googledrive

import (
	"context"

	"google.golang.org/api/drive/v3"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// folderSyncPageSize is the number of files to request per page
	folderSyncPageSize = int64(100)
	// googleDocMIMEType is the MIME type for Google Docs
	googleDocMIMEType = "application/vnd.google-apps.document"
)

// folderSyncOperation is the operation ref for the folder sync operation
var folderSyncOperation = types.OperationRefOf[FolderSync]().Ingests(driveClient, runFolderSync)

// FolderSync lists Google Docs in a configured folder and emits ingest envelopes for policy creation
type FolderSync struct {
	// Switch toggles the folder sync off for the installation
	types.Switch
	// FolderID is the Google Drive folder ID (or full URL) containing policy documents; required so an install is scoped to one folder rather than the caller's entire Drive
	FolderID string `json:"folderId,omitempty" jsonschema:"title=Folder ID,description=Google Drive folder ID or URL containing policy documents,required"`
	// FilterExpr is an optional CEL expression to filter which documents in the folder are eligible
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to filter documents before creating policies"`
}

// runFolderSync lists all Google Docs in the configured folder and returns ingest payload sets
func runFolderSync(ctx context.Context, _ types.OperationRequest, svc DriveClient, cfg FolderSync) ([]types.IngestPayloadSet, error) {
	folderID := parseFolderID(cfg.FolderID)
	if folderID == "" {
		logx.FromContext(ctx).Error().Bool("folder_id_present", cfg.FolderID != "").Msg("googledrive: no folder id in the installation user input, the integration needs to be reconfigured with a folder selected")

		return nil, ErrFolderIDMissing
	}

	files, err := listFolderDocs(ctx, svc, folderID)
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
