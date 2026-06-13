package onedrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	diAPIVersion    = "2024-11-30"
	diModel         = "prebuilt-layout"
	diMaxPolls      = 30
	diPollInterval  = 2 * time.Second
	diAnalyzePath   = "/documentintelligence/documentModels/" + diModel + ":analyze?api-version=" + diAPIVersion
)

// diSpan is the character-offset range of a content element within the full document text
type diSpan struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
}

type diParagraph struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Spans   []diSpan `json:"spans"`
}

type diCell struct {
	RowIndex    int      `json:"rowIndex"`
	ColumnIndex int      `json:"columnIndex"`
	Kind        string   `json:"kind"`
	Content     string   `json:"content"`
	Spans       []diSpan `json:"spans"`
}

type diTable struct {
	RowCount    int      `json:"rowCount"`
	ColumnCount int      `json:"columnCount"`
	Cells       []diCell `json:"cells"`
	Spans       []diSpan `json:"spans"`
}

type diAnalyzeResult struct {
	Paragraphs []diParagraph `json:"paragraphs"`
	Tables     []diTable     `json:"tables"`
}

type diOperationResponse struct {
	Status        string           `json:"status"`
	Error         *diErrorBody     `json:"error"`
	AnalyzeResult *diAnalyzeResult `json:"analyzeResult"`
}

type diErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// analyzeWithDocumentIntelligence submits docxBytes to the Azure Document Intelligence
// prebuilt-layout model and returns rendered HTML. The caller must supply a non-empty
// endpoint (e.g. "https://my-resource.cognitiveservices.azure.com") and API key.
func analyzeWithDocumentIntelligence(ctx context.Context, endpoint, apiKey string, docxBytes []byte) (string, error) {
	endpoint = strings.TrimRight(endpoint, "/")

	analyzeURL := endpoint + diAnalyzePath

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, analyzeURL, bytes.NewReader(docxBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", apiKey)
	req.Header.Set("Content-Type", docxMIMEType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("document intelligence analyze returned %d", resp.StatusCode) //nolint:err113
	}

	opURL := resp.Header.Get("Operation-Location")
	if opURL == "" {
		return "", fmt.Errorf("document intelligence: missing Operation-Location header") //nolint:err113
	}

	result, err := pollDocumentIntelligenceOperation(ctx, opURL, apiKey)
	if err != nil {
		return "", err
	}

	return diResultToHTML(result), nil
}

func pollDocumentIntelligenceOperation(ctx context.Context, opURL, apiKey string) (*diAnalyzeResult, error) {
	for range diMaxPolls {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(diPollInterval):
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, opURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Ocp-Apim-Subscription-Key", apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		var op diOperationResponse

		decErr := json.NewDecoder(resp.Body).Decode(&op)
		resp.Body.Close()

		if decErr != nil {
			return nil, decErr
		}

		switch op.Status {
		case "succeeded":
			return op.AnalyzeResult, nil
		case "failed":
			if op.Error != nil {
				return nil, fmt.Errorf("document intelligence failed: %s: %s", op.Error.Code, op.Error.Message) //nolint:err113
			}

			return nil, fmt.Errorf("document intelligence operation failed") //nolint:err113
		}
		// status == "running" or "notStarted" — keep polling
	}

	return nil, fmt.Errorf("document intelligence: timed out waiting for result") //nolint:err113
}

// diDocumentItem is a sortable union type — either a paragraph or a table — carrying
// its minimum span offset so that all elements can be sorted in document order.
type diDocumentItem struct {
	offset    int
	paragraph *diParagraph
	table     *diTable
}

func diResultToHTML(result *diAnalyzeResult) string {
	if result == nil {
		return ""
	}

	var items []diDocumentItem

	for i := range result.Paragraphs {
		p := &result.Paragraphs[i]

		if len(p.Spans) == 0 {
			continue
		}

		items = append(items, diDocumentItem{offset: p.Spans[0].Offset, paragraph: p})
	}

	for i := range result.Tables {
		t := &result.Tables[i]

		if len(t.Spans) == 0 {
			continue
		}

		items = append(items, diDocumentItem{offset: t.Spans[0].Offset, table: t})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].offset < items[j].offset })

	var b strings.Builder

	for _, item := range items {
		switch {
		case item.paragraph != nil:
			b.WriteString(diParagraphToHTML(item.paragraph))
		case item.table != nil:
			b.WriteString(diTableToHTML(item.table))
		}
	}

	return b.String()
}

func diParagraphToHTML(p *diParagraph) string {
	if strings.TrimSpace(p.Content) == "" {
		return ""
	}

	escaped := html.EscapeString(p.Content)

	switch p.Role {
	case "title":
		return "<h1>" + escaped + "</h1>\n"
	case "sectionHeading":
		return "<h2>" + escaped + "</h2>\n"
	case "pageHeader", "pageFooter", "pageNumber", "footnote":
		return ""
	default:
		return "<p>" + escaped + "</p>\n"
	}
}

func diTableToHTML(t *diTable) string {
	if t.RowCount == 0 || t.ColumnCount == 0 {
		return ""
	}

	// Build a 2-D grid so we can render cells by position
	grid := make([][]string, t.RowCount)
	kinds := make([][]string, t.RowCount)

	for r := range t.RowCount {
		grid[r] = make([]string, t.ColumnCount)
		kinds[r] = make([]string, t.ColumnCount)
	}

	for _, cell := range t.Cells {
		if cell.RowIndex < t.RowCount && cell.ColumnIndex < t.ColumnCount {
			grid[cell.RowIndex][cell.ColumnIndex] = html.EscapeString(cell.Content)
			kinds[cell.RowIndex][cell.ColumnIndex] = cell.Kind
		}
	}

	var b strings.Builder

	b.WriteString("<table>\n")

	for r, row := range grid {
		b.WriteString("<tr>")

		for c, cell := range row {
			tag := "td"
			if kinds[r][c] == "columnHeader" || kinds[r][c] == "rowHeader" || kinds[r][c] == "stub" {
				tag = "th"
			}

			b.WriteString("<" + tag + ">" + cell + "</" + tag + ">")
		}

		b.WriteString("</tr>\n")
	}

	b.WriteString("</table>\n")

	return b.String()
}
