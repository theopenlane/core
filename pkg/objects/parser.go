package objects

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/unidoc/unioffice/document"
)

var (
	errEmptyContentProvided = errors.New("empty content provided")
)

// ParseDocument parses the document content based on the provided mime type
// and returns the extracted text content as a string.
// only supports .docx, .txt, .md, .mdx files
func ParseDocument(content []byte, mimeType string) (string, error) {
	if len(content) == 0 {
		return "", errEmptyContentProvided
	}

	switch mimeType {
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return parseDocx(content)
	case "text/markdown":
		return parseMarkdown(content)
	default:
		return string(content), nil
	}
}

func parseDocx(content []byte) (string, error) {
	reader := bytes.NewReader(content)
	doc, err := document.Read(reader, int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to read docx file: %w", err) // nolint:err113
	}

	defer doc.Close()

	var w strings.Builder

	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			w.WriteString(run.Text())
			w.WriteString(" ")
		}

		w.WriteString("\n")
	}

	return strings.TrimSpace(w.String()), nil
}

func parseMarkdown(content []byte) (string, error) {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	flags := html.CommonFlags | html.HrefTargetBlank
	renderer := html.NewRenderer(html.RendererOptions{Flags: flags})

	doc := p.Parse(content)

	return string(markdown.Render(doc, renderer)), nil
}
