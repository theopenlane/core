//go:build cli

package speccli

import (
	"strings"

	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/core/pkg/objects/storage"
)

// UploadFromPath opens the given path and returns a graphql.Upload usable with mutations.
// An empty path returns (nil, nil) so callers can treat the field as optional.
func UploadFromPath(path string) (*graphql.Upload, error) {
	cleaned := strings.TrimSpace(path)
	if cleaned == "" {
		return nil, nil
	}

	file, err := storage.NewUploadFile(cleaned)
	if err != nil {
		return nil, err
	}

	return &graphql.Upload{
		File:        file.RawFile,
		Filename:    file.OriginalName,
		Size:        file.Size,
		ContentType: file.ContentType,
	}, nil
}
