package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fumiama/go-docx"
	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"gopkg.in/yaml.v3"
)

const (
	// MIMEDetectionBufferSize defines the buffer size for MIME type detection
	MIMEDetectionBufferSize = 512
)

// DetectContentType detects the MIME type of the provided reader using gabriel-vasile/mimetype library
func DetectContentType(reader io.ReadSeeker) (string, error) {
	// Seek to beginning
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	// Detect content type using mimetype library for better accuracy
	mimeType, err := mimetype.DetectReader(reader)
	if err != nil {
		return "", err
	}

	// Seek back to beginning
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	return mimeType.String(), nil
}

// ParsedDocument represents a document parsed with its frontmatter and data
type ParsedDocument struct {
	// Frontmatter contains metadata extracted from the document, only for markdown files
	Frontmatter *Frontmatter
	// Data contains the parsed content of the document
	Data any
}

// ParseDocument parses a document based on its MIME type
func ParseDocument(reader io.Reader, mimeType string) (*ParsedDocument, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	switch {
	case strings.Contains(mimeType, "json"):
		var result any
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrJSONParseFailed, err)
		}
		return &ParsedDocument{Data: result}, nil
	case strings.Contains(mimeType, "yaml") || strings.Contains(mimeType, "yml"):
		var result any
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrYAMLParseFailed, err)
		}
		return &ParsedDocument{Data: result}, nil
	case strings.Contains(mimeType, "application/vnd.openxmlformats-officedocument.wordprocessingml.document"):
		text, err := parseDocx(data)
		if err != nil {
			return nil, err
		}

		return &ParsedDocument{Data: text}, nil
	case strings.Contains(mimeType, "text/plain"):
		fm, body, err := ParseFrontmatter(data)
		if err != nil {
			return nil, err
		}

		return &ParsedDocument{Frontmatter: fm, Data: string(body)}, nil
	case strings.Contains(mimeType, "text/markdown"), strings.Contains(mimeType, "text/x-markdown"):
		fm, body, err := ParseFrontmatter(data)
		if err != nil {
			return nil, err
		}

		return &ParsedDocument{Frontmatter: fm, Data: body}, nil
	case strings.Contains(mimeType, "text/html"):
		text, err := parseHTML(data)
		if err != nil {
			return nil, err
		}

		return &ParsedDocument{Data: text}, nil
	default:
		return &ParsedDocument{Data: data}, nil
	}
}

func getHeadingLevel(p *docx.Paragraph) int {
	if p.Properties == nil || p.Properties.Style == nil {
		return 0
	}

	style := strings.ToLower(p.Properties.Style.Val)

	if levelStr, found := strings.CutPrefix(style, "heading"); found {
		if level := parseHeadingLevel(levelStr); level > 0 {
			return level
		}
	}

	switch style {
	case "title":
		return 1 //nolint:mnd
	case "subtitle":
		return 2 //nolint:mnd
	case "h1":
		return 1 //nolint:mnd
	case "h2":
		return 2 //nolint:mnd
	case "h3":
		return 3 //nolint:mnd
	case "h4":
		return 4 //nolint:mnd
	case "h5":
		return 5 //nolint:mnd
	case "h6":
		return 6 //nolint:mnd
	}

	return 0
}

func parseHeadingLevel(s string) int {
	if len(s) == 1 && s[0] >= '1' && s[0] <= '6' {
		return int(s[0] - '0')
	}

	return 0
}

// parseDocx extracts and returns the text content from a DOCX file preserving paragraph structure
func parseDocx(content []byte) (string, error) {
	reader := bytes.NewReader(content)

	doc, err := docx.Parse(reader, int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrDOCXParseFailed, err)
	}

	var paragraphs []string

	for _, item := range doc.Document.Body.Items {
		switch p := item.(type) {
		case *docx.Paragraph:

			if text := p.String(); text != "" {
				if lvl := getHeadingLevel(p); lvl > 0 {
					prefix := strings.Repeat("#", lvl)
					paragraphs = append(paragraphs, prefix+" "+text)
				} else {
					paragraphs = append(paragraphs, text)
				}
			}

		case *docx.Table:
			for rowIdx, row := range p.TableRows {
				var cells []string

				for _, cell := range row.TableCells {
					var cellContent []string

					for _, para := range cell.Paragraphs {
						if t := para.String(); t != "" {
							cellContent = append(cellContent, t)
						}
					}

					cells = append(cells, strings.Join(cellContent, " "))
				}

				if len(cells) > 0 {
					line := "| " + strings.Join(cells, " | ") + " |"
					paragraphs = append(paragraphs, line)
					// If this is the first row, add a markdown separator
					if rowIdx == 0 {
						separator := "|"
						for range cells {
							separator += " --- |"
						}
						paragraphs = append(paragraphs, separator)
					}
				}
			}

			paragraphs = append(paragraphs, "")
		}
	}

	return strings.Join(paragraphs, "\n"), nil
}

// htmlSkipElements lists the tags whose contents should be discarded entirely
// (script/style and friends) when extracting plain text.
var htmlSkipElements = map[atom.Atom]struct{}{
	atom.Script:   {},
	atom.Style:    {},
	atom.Noscript: {},
	atom.Template: {},
	atom.Head:     {},
}

// htmlBlockElements lists the tags that should produce a paragraph break in the
// extracted text. Inline tags (<span>, <b>, <i>, <a>, …) intentionally do not
// appear here; their text content concatenates inline.
var htmlBlockElements = map[atom.Atom]struct{}{
	atom.Address: {}, atom.Article: {}, atom.Aside: {}, atom.Blockquote: {},
	atom.Br: {}, atom.Caption: {}, atom.Dd: {}, atom.Div: {},
	atom.Dl: {}, atom.Dt: {}, atom.Figcaption: {}, atom.Figure: {},
	atom.Footer: {}, atom.Form: {}, atom.H1: {}, atom.H2: {},
	atom.H3: {}, atom.H4: {}, atom.H5: {}, atom.H6: {},
	atom.Header: {}, atom.Hr: {}, atom.Li: {}, atom.Main: {},
	atom.Nav: {}, atom.Ol: {}, atom.P: {}, atom.Pre: {},
	atom.Section: {}, atom.Table: {}, atom.Tr: {}, atom.Ul: {},
}

// parseHTML extracts readable plain text from an HTML document. The full HTML
// stays in object storage for FilePreview to render; this is the text used for
// search indexing and details summaries, so it must (a) decode HTML entities
// and (b) emit paragraph breaks between block elements — neither of which a
// pure tag-stripper does correctly.
func parseHTML(content []byte) (string, error) {
	node, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrHTMLParseFailed, err)
	}

	var buf strings.Builder

	walkHTMLText(node, &buf)

	return collapseWhitespace(buf.String()), nil
}

// walkHTMLText writes the textual content of n (and its descendants) to buf,
// emitting newlines around block-level elements and skipping non-content tags.
func walkHTMLText(n *html.Node, buf *strings.Builder) {
	if n.Type == html.TextNode {
		buf.WriteString(n.Data)

		return
	}

	if n.Type == html.ElementNode {
		if _, skip := htmlSkipElements[n.DataAtom]; skip {
			return
		}

		if _, block := htmlBlockElements[n.DataAtom]; block {
			buf.WriteByte('\n')
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkHTMLText(c, buf)
	}

	if n.Type == html.ElementNode {
		if _, block := htmlBlockElements[n.DataAtom]; block {
			buf.WriteByte('\n')
		}
	}
}

// collapseWhitespace normalizes runs of spaces/tabs within each line, drops
// empty lines, and joins the remaining lines with a single newline. The result
// is human-readable plain text suited for indexing and previewing.
func collapseWhitespace(s string) string {
	var out strings.Builder

	for _, raw := range strings.Split(s, "\n") {
		line := strings.Join(strings.Fields(raw), " ")
		if line == "" {
			continue
		}

		if out.Len() > 0 {
			out.WriteByte('\n')
		}

		out.WriteString(line)
	}

	return out.String()
}

// NewUploadFile creates a new File from a file path
func NewUploadFile(path string) (*File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	// Detect content type
	mimeType, err := DetectContentType(file)
	if err != nil {
		// Fallback to default
		mimeType = "application/octet-stream"
	}

	// Reset to beginning of file
	if _, err := file.Seek(0, 0); err != nil {
		file.Close()
		return nil, err
	}

	return &File{
		RawFile:      file,
		CreatedAt:    time.Now(),
		OriginalName: stat.Name(),
		FileMetadata: FileMetadata{
			Size:        stat.Size(),
			ContentType: mimeType,
			Key:         "file_upload",
		},
	}, nil
}
