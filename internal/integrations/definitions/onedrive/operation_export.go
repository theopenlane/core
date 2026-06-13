package onedrive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// Handle adapts the document export to the generic operation registration boundary
func Handle() types.OperationHandler {
	return operations.Handle(oneDriveClient, documentExportOperation)
}

// Export fetches OneDrive item metadata and returns either an iframe embed or PDF bytes
func (c DriveClient) Export(ctx context.Context, cfg *operations.DocumentExport) error {
	if cfg.FileID == "" {
		return ErrExportFailed
	}

	log := logx.FromContext(ctx).With().Str("item_id", cfg.FileID).Logger()

	itemURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s", cfg.FileID)

	item, err := c.Graph.Drives().ByDriveId("me").Items().ByDriveItemId(cfg.FileID).WithUrl(itemURL).Get(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("onedrive: failed to get item metadata")
		return ErrExportFailed
	}

	if item.GetFile() != nil {
		cfg.MimeType = lo.FromPtr(item.GetFile().GetMimeType())
	}

	name := lo.FromPtr(item.GetName())
	cfg.Name = strings.TrimSuffix(name, path.Ext(name))

	if c.Cfg.ContentMode == "pdf" {
		pdfBytes, err := fetchPDF(ctx, c.TS, cfg.FileID)
		if err != nil {
			log.Warn().Err(err).Msg("onedrive: pdf export failed")
			return ErrExportFailed
		}

		cfg.PDF = pdfBytes
		log.Info().Int("pdf_len", len(cfg.PDF)).Msg("onedrive: pdf export result")

		return nil
	}

	cfg.HTML = fetchIframeEmbed(ctx, c.TS, cfg.FileID)
	log.Info().Bool("has_html", cfg.HTML != "").Msg("onedrive: iframe export result")

	return nil
}

// fetchIframeEmbed calls the Graph preview API and returns an <iframe> pointing at the
// embeddable Office preview URL. Works for both personal and business accounts.
func fetchIframeEmbed(ctx context.Context, ts oauth2.TokenSource, itemID string) string {
	log := logx.FromContext(ctx).With().Str("item_id", itemID).Logger()

	tok, err := ts.Token()
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: failed to get token for iframe embed")
		return ""
	}

	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/preview", itemID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader("{}"))
	if err != nil {
		return ""
	}

	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: preview API request failed")
		return ""
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().Int("status", resp.StatusCode).Msg("onedrive: preview API returned non-200")
		return ""
	}

	var result struct {
		GetURL string `json:"getUrl"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.GetURL == "" {
		log.Warn().Err(err).Msg("onedrive: failed to decode preview response or empty getUrl")
		return ""
	}

	return `<iframe src="` + result.GetURL + `" width="100%" height="600" frameborder="0" allowfullscreen></iframe>`
}

// fetchPDF downloads the item as a PDF using the Graph convert-to-PDF endpoint
func fetchPDF(ctx context.Context, ts oauth2.TokenSource, itemID string) ([]byte, error) {
	log := logx.FromContext(ctx).With().Str("item_id", itemID).Logger()

	tok, err := ts.Token()
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/content?format=pdf", itemID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	// http.DefaultClient strips the Authorization header on cross-domain redirects,
	// so the CDN pre-authenticated URL receives no credentials
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: pdf download request failed")
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warn().Int("status", resp.StatusCode).Msg("onedrive: pdf download returned non-200")
		return nil, fmt.Errorf("graph returned %d fetching pdf", resp.StatusCode) //nolint:err113
	}

	return io.ReadAll(resp.Body)
}
