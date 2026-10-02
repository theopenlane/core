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

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
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

// FolderSync lists OneDrive documents in a folder and emits ingest envelopes
type FolderSync struct {
	types.OperationSettings
	// FolderID is the folder path relative to the drive root; empty syncs the root
	FolderID string `json:"folderId,omitempty" jsonschema:"title=Folder Path,description=Folder path relative to drive root (e.g. Policies). Leave empty to sync the entire drive root."`
}

// runFolderSync lists document files in the configured folder and returns payload sets
func runFolderSync(ctx context.Context, _ types.OperationRequest, c *DriveClient, cfg FolderSync) ([]types.IngestPayloadSet, error) {
	items, err := listFolderDocs(ctx, c, parseFolderID(cfg.FolderID))
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

// folderChildrenURL returns the Graph API URL for listing children of the given folder
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
