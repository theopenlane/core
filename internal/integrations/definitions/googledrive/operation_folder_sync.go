package googledrive

import (
	"context"

	"google.golang.org/api/drive/v3"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	// folderSyncPageSize is the number of files to request per page
	folderSyncPageSize = int64(100)
	// googleDocMIMEType is the MIME type for Google Docs
	googleDocMIMEType = "application/vnd.google-apps.document"
)

// FolderSync lists Google Docs in a configured folder and emits ingest envelopes for policy creation
type FolderSync struct{}

// IngestHandle adapts folder sync to the ingest operation registration boundary
func (f FolderSync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(driveClient, func(ctx context.Context, request types.OperationRequest, svc *drive.Service) ([]types.IngestPayloadSet, error) {
		var input UserInput

		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &input)
		}

		folderID := parseFolderID(input.FolderID)
		if folderID == "" {
			return nil, ErrFolderIDMissing
		}

		return f.Run(ctx, svc, folderID)
	})
}

// Run lists all Google Docs in the given folder and returns ingest payload sets
func (FolderSync) Run(ctx context.Context, svc *drive.Service, folderID string) ([]types.IngestPayloadSet, error) {
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
			Schema:    integrationgenerated.IntegrationMappingSchemaInternalPolicy,
			Envelopes: envelopes,
		},
	}, nil
}

// listFolderDocs pages through all Google Docs in the specified folder
func listFolderDocs(ctx context.Context, svc *drive.Service, folderID string) ([]*drive.File, error) {
	var files []*drive.File

	pageToken := ""
	query := "'" + folderID + "' in parents and mimeType = '" + googleDocMIMEType + "' and trashed = false"

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		call := svc.Files.List().
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
