package onedrive

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	msgraphdrives "github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

const folderSyncPageSize = int32(250)

// documentMIMETypes is the set of MIME types treated as policy documents
var documentMIMETypes = map[string]bool{
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"application/pdf": true,
	"text/plain":      true,
	"text/html":       true,
}

// driveItemPayload is the JSON-serializable representation of one OneDrive file
type driveItemPayload struct {
	// ID is the stable OneDrive item identifier
	ID string `json:"id,omitempty"`
	// Name is the file name without extension
	Name string `json:"name,omitempty"`
	// MimeType is the file MIME type
	MimeType string `json:"mimeType,omitempty"`
	// WebURL is the OneDrive browser URL for opening the file
	WebURL string `json:"webUrl,omitempty"`
	// LastModifiedDateTime is when the file was last modified
	LastModifiedDateTime time.Time `json:"lastModifiedDateTime,omitempty"`
	// CreatedDateTime is when the file was created
	CreatedDateTime time.Time `json:"createdDateTime,omitempty"`
}

// FolderSync lists OneDrive documents in a configured folder and emits ingest envelopes for policy creation
type FolderSync struct{}

// IngestHandle adapts folder sync to the ingest operation registration boundary
func (f FolderSync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(oneDriveClient, func(ctx context.Context, request types.OperationRequest, c *DriveClient) ([]types.IngestPayloadSet, error) {
		var input UserInput

		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &input)
		}

		return f.Run(ctx, c, parseFolderID(input.FolderID))
	})
}

// Run lists all document files in the given folder (or the drive root if folderID is empty) and returns ingest payload sets
func (FolderSync) Run(ctx context.Context, c *DriveClient, folderID string) ([]types.IngestPayloadSet, error) {
	items, err := listFolderDocs(ctx, c, folderID)
	if err != nil {
		return nil, err
	}

	envelopes := make([]types.MappingEnvelope, 0, len(items))

	for _, item := range items {
		payload := driveItemToPayload(item)

		envelope, err := providerkit.MarshalEnvelope(payload.ID, payload, ErrPayloadEncode)
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

const (
	meDriveRootChildrenURL = "https://graph.microsoft.com/v1.0/me/drive/root/children"
	meDrivePathChildrenURL = "https://graph.microsoft.com/v1.0/me/drive/root:/%s:/children"
)

// folderChildrenURL returns the Graph API URL for listing children of the given folder.
// An empty path lists the drive root; any other value is treated as a path relative to the root
// (e.g. "Policies"). Raw OneDrive CIDs and item IDs are not supported — use paths.
func folderChildrenURL(folderPath string) string {
	if folderPath == "" {
		return meDriveRootChildrenURL
	}

	return fmt.Sprintf(meDrivePathChildrenURL, folderPath)
}

// listFolderDocs pages through all document files in the specified OneDrive folder
func listFolderDocs(ctx context.Context, c *DriveClient, folderPath string) ([]models.DriveItemable, error) {
	reqCfg := &msgraphdrives.ItemItemsItemChildrenRequestBuilderGetRequestConfiguration{
		QueryParameters: &msgraphdrives.ItemItemsItemChildrenRequestBuilderGetQueryParameters{
			Select: []string{"id", "name", "file", "webUrl", "lastModifiedDateTime", "createdDateTime"},
			Top:    lo.ToPtr(folderSyncPageSize),
		},
	}

	builder := c.Graph.Drives().ByDriveId("me").Items().ByDriveItemId("root").Children().WithUrl(folderChildrenURL(folderPath))

	page, err := builder.Get(ctx, reqCfg)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("folder_path", folderPath).Msg("onedrive: failed to list folder children")
		return nil, ErrFolderListFailed
	}

	var items []models.DriveItemable

	for _, item := range page.GetValue() {
		if isDocumentItem(item) {
			items = append(items, item)
		}
	}

	for page.GetOdataNextLink() != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		page, err = builder.WithUrl(*page.GetOdataNextLink()).Get(ctx, nil)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("onedrive: failed to list folder children page")
			return nil, ErrFolderListFailed
		}

		for _, item := range page.GetValue() {
			if isDocumentItem(item) {
				items = append(items, item)
			}
		}
	}

	return items, nil
}

// isDocumentItem reports whether the item is a document file eligible for policy sync
func isDocumentItem(item models.DriveItemable) bool {
	file := item.GetFile()
	if file == nil {
		return false
	}

	mimeType := lo.FromPtr(file.GetMimeType())
	return documentMIMETypes[mimeType]
}

// driveItemToPayload maps a DriveItemable SDK model to a JSON-serializable payload struct
func driveItemToPayload(item models.DriveItemable) driveItemPayload {
	name := lo.FromPtr(item.GetName())
	p := driveItemPayload{
		ID:   lo.FromPtr(item.GetId()),
		Name: strings.TrimSuffix(name, path.Ext(name)),
	}

	if item.GetFile() != nil {
		p.MimeType = lo.FromPtr(item.GetFile().GetMimeType())
	}

	if item.GetLastModifiedDateTime() != nil {
		p.LastModifiedDateTime = *item.GetLastModifiedDateTime()
	}

	if item.GetCreatedDateTime() != nil {
		p.CreatedDateTime = *item.GetCreatedDateTime()
	}

	if item.GetWebUrl() != nil {
		p.WebURL = *item.GetWebUrl()
	}

	return p
}
