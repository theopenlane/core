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

const docxMIMEType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

// Handle adapts the document export to the generic operation registration boundary
func Handle() types.OperationHandler {
	return operations.Handle(oneDriveClient, documentExportOperation)
}

// Run fetches OneDrive item metadata and returns an HTML representation of the document
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

	log.Info().Str("mime_type", cfg.MimeType).Msg("onedrive: fetched item metadata")

	name := lo.FromPtr(item.GetName())
	cfg.Name = strings.TrimSuffix(name, path.Ext(name))

	cfg.HTML = htmlContent(ctx, c.TS, c.Cfg, cfg.FileID, cfg.MimeType)
	log.Info().Bool("has_html", cfg.HTML != "").Int("html_len", len(cfg.HTML)).Msg("onedrive: export html result")

	return nil
}

// htmlContent returns HTML for the document. Behaviour is controlled by cfg.ContentMode:
//   - "iframe": returns an embeddable <iframe> via the Graph preview API (no DOCX download)
//   - "html" (default): downloads the DOCX and parses it locally (or via Document Intelligence)
func htmlContent(ctx context.Context, ts oauth2.TokenSource, cfg Config, itemID, mimeType string) string {
	log := logx.FromContext(ctx).With().Str("item_id", itemID).Logger()

	if cfg.ContentMode == "iframe" {
		return fetchIframeEmbed(ctx, ts, itemID)
	}

	if mimeType != docxMIMEType {
		log.Warn().Str("mime_type", mimeType).Msg("onedrive: unsupported mime type for content extraction")
		return ""
	}

	downloadURL, err := getPreAuthDownloadURL(ctx, ts, itemID)
	if err != nil || downloadURL == "" {
		log.Warn().Err(err).Msg("onedrive: could not get pre-authenticated download URL, falling back to content endpoint")
		return fetchDocxViaContentEndpoint(ctx, ts, cfg, itemID)
	}

	return fetchDocxHTML(ctx, ts, cfg, itemID, downloadURL)
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

// getPreAuthDownloadURL fetches the @microsoft.graph.downloadUrl OData annotation via a raw
// HTTP request. The kiota SDK strips OData annotations during deserialization, so we decode
// the JSON ourselves to reliably extract the pre-authenticated URL.
func getPreAuthDownloadURL(ctx context.Context, ts oauth2.TokenSource, itemID string) (string, error) {
	tok, err := ts.Token()
	if err != nil {
		return "", err
	}

	u := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s?$select=id,@microsoft.graph.downloadUrl", itemID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("graph returned %d fetching download URL", resp.StatusCode) //nolint:err113
	}

	var result struct {
		DownloadURL string `json:"@microsoft.graph.downloadUrl"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.DownloadURL, nil
}

// bearerTransport is an http.RoundTripper that adds a Bearer token to every request
type bearerTransport struct {
	token string
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(req)
}

// fetchDocxHTML downloads the DOCX from the pre-authenticated URL and converts it to HTML.
// The pre-authenticated URL (@microsoft.graph.downloadUrl) must be fetched without a Bearer
// token — the auth is embedded in the URL parameters and an explicit header causes 401.
func fetchDocxHTML(ctx context.Context, ts oauth2.TokenSource, cfg Config, itemID, downloadURL string) string {
	log := logx.FromContext(ctx).With().Str("item_id", itemID).Logger()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: failed to build docx download request")
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: docx download failed")
		return ""
	}

	defer resp.Body.Close()

	log.Info().Int("status", resp.StatusCode).Str("host", resp.Request.URL.Host).Msg("onedrive: docx download response")

	if resp.StatusCode != http.StatusOK {
		log.Warn().Int("status", resp.StatusCode).Msg("onedrive: docx download returned non-200")

		// Pre-authenticated URL may have expired; fall back to fetching a fresh one via Graph
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return fetchDocxViaContentEndpoint(ctx, ts, cfg, itemID)
		}

		return ""
	}

	docxBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: failed to read docx response body")
		return ""
	}

	return docxBytesToHTML(ctx, cfg, itemID, docxBytes)
}

// fetchDocxViaContentEndpoint is a fallback for when the pre-authenticated URL has expired.
// It fetches a fresh download URL by stopping at Graph's redirect and then fetching the
// CDN URL without auth.
func fetchDocxViaContentEndpoint(ctx context.Context, ts oauth2.TokenSource, cfg Config, itemID string) string {
	log := logx.FromContext(ctx).With().Str("item_id", itemID).Logger()
	log.Info().Msg("onedrive: falling back to content endpoint for docx download")

	tok, err := ts.Token()
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: failed to get token for content endpoint fallback")
		return ""
	}

	contentURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/content", itemID)

	noRedirectClient := &http.Client{
		Transport: &bearerTransport{token: tok.AccessToken},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, contentURL, nil)
	if err != nil {
		return ""
	}

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: content endpoint request failed")
		return ""
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently {
		log.Warn().Int("status", resp.StatusCode).Msg("onedrive: content endpoint did not redirect")
		return ""
	}

	cdnURL := resp.Header.Get("Location")
	if cdnURL == "" {
		log.Warn().Msg("onedrive: content endpoint redirect had no Location header")
		return ""
	}

	dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, cdnURL, nil)
	if err != nil {
		return ""
	}

	dlResp, err := http.DefaultClient.Do(dlReq)
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: cdn fetch failed")
		return ""
	}

	defer dlResp.Body.Close()

	log.Info().Int("status", dlResp.StatusCode).Str("host", dlResp.Request.URL.Host).Msg("onedrive: cdn response")

	if dlResp.StatusCode != http.StatusOK {
		log.Warn().Int("status", dlResp.StatusCode).Msg("onedrive: cdn returned non-200")
		return ""
	}

	docxBytes, err := io.ReadAll(dlResp.Body)
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: failed to read cdn response body")
		return ""
	}

	return docxBytesToHTML(ctx, cfg, itemID, docxBytes)
}

// docxBytesToHTML converts raw DOCX bytes to HTML. When cfg has Document Intelligence
// credentials set, it delegates to the Azure prebuilt-layout model; otherwise it falls
// back to the local XML parser in docparse.go.
func docxBytesToHTML(ctx context.Context, cfg Config, itemID string, docxBytes []byte) string {
	log := logx.FromContext(ctx).With().Str("item_id", itemID).Logger()

	if cfg.DocumentIntelligenceEndpoint != "" && cfg.DocumentIntelligenceKey != "" {
		result, err := analyzeWithDocumentIntelligence(ctx, cfg.DocumentIntelligenceEndpoint, cfg.DocumentIntelligenceKey, docxBytes)
		if err != nil {
			log.Warn().Err(err).Msg("onedrive: document intelligence failed, falling back to local parsing")
		} else {
			return result
		}
	}

	result, err := extractDocxHTML(docxBytes)
	if err != nil {
		log.Warn().Err(err).Msg("onedrive: xml extraction failed")
		return ""
	}

	return result
}
