package docextract

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"google.golang.org/genai"

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

	doc, err := c.loadDocument(ctx, pdf)
	if err != nil {
		return "", err
	}

	// the cache holds its own reference to an uploaded file, the file record is not needed afterwards
	defer c.releaseDocument(ctx, doc)

	cache, err := c.Caches.Create(ctx, c.model, &genai.CreateCachedContentConfig{
		TTL:               ttl,
		Contents:          []*genai.Content{genai.NewContentFromParts([]*genai.Part{doc.part}, genai.RoleUser)},
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: c.systemInstruction}}},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create document cache: %w", err)
	}

	logx.FromContext(ctx).Debug().Str("cache", CacheID(cache.Name)).Dur("ttl", ttl).Msg("docextract: document cached")

	return cache.Name, nil
}

// ReleaseDocumentCache deletes cached content created by CacheDocument; a cache that already
// expired is not an error
func (c *Client) ReleaseDocumentCache(ctx context.Context, name string) error {
	if name == "" {
		return nil
	}

	if _, err := c.Caches.Delete(ctx, name, nil); err != nil {
		return fmt.Errorf("failed to delete document cache: %w", err)
	}

	return nil
}
