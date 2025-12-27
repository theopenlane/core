//go:build examples

package openlane

import (
	"context"

	openlane "github.com/theopenlane/go-client"
)

// FileBase64Response captures the GraphQL response for a file base64 lookup.
type FileBase64Response struct {
	File FileBase64Payload `json:"file"`
}

// FileBase64Payload represents the file payload returned by the base64 query.
type FileBase64Payload struct {
	ID                  string  `json:"id"`
	Base64              *string `json:"base64,omitempty"`
	DetectedContentType *string `json:"detectedContentType,omitempty"`
	ProvidedFileName    *string `json:"providedFileName,omitempty"`
}

// FetchFileBase64 queries the Openlane API for the base64 contents of a file.
func FetchFileBase64(ctx context.Context, client *openlane.Client, fileID string) (*FileBase64Response, error) {
	vars := map[string]any{
		"fileId": fileID,
	}

	const query = `query GetFileBase64 ($fileId: ID!) {
	file(id: $fileId) {
		id
		base64
		detectedContentType
		providedFileName
	}
}
`

	var res FileBase64Response
	if err := client.Client.Post(ctx, "GetFileBase64", query, &res, vars); err != nil {
		return nil, err
	}

	return &res, nil
}
