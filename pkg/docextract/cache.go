package docextract

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/theopenlane/core/v2/pkg/logx"
)

// CacheID is the trailing identifier of a cached content resource name and not the full path
func CacheID(name string) string {
	if index := strings.LastIndex(name, "/"); index >= 0 {
		return name[index+1:]
	}

	return name
}

// CacheDocument stores the document and system instruction as cached content for the given
// lifetime so every request for the same document references the cache instead of re-sending it
func (c *Client) CacheDocument(ctx context.Context, pdf io.Reader, ttl time.Duration) (string, error) {
	if c.systemInstruction == "" {
		return "", ErrSystemInstructionRequired
	}

	doc, err := c.provider.PrepareDocument(ctx, pdf)
	if err != nil {
		return "", err
	}

	// the cache holds its own reference to an uploaded file, the file record is not needed afterwards
	defer doc.Release(ctx)

	name, err := c.provider.CreateCache(ctx, CacheRequest{
		Document:          doc,
		SystemInstruction: c.systemInstruction,
		TTL:               ttl,
	})
	if err != nil {
		return "", err
	}

	logx.FromContext(ctx).Debug().Str("cache", CacheID(name)).Dur("ttl", ttl).Msg("docextract: document cached")

	return name, nil
}

// ReleaseDocumentCache deletes cached content created by CacheDocument; a cache that already
// expired is not an error
func (c *Client) ReleaseDocumentCache(ctx context.Context, name string) error {
	return c.provider.DeleteCache(ctx, name)
}
