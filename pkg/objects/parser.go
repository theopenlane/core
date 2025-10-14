package objects

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/nguyenthenguyen/docx"
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
	default:
		return string(content), nil
	}
}

func parseDocx(content []byte) (string, error) {
	reader := bytes.NewReader(content)

	doc, err := docx.ReadDocxFromMemory(reader, int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to read docx file: %w", err) // nolint:err113
	}

	defer doc.Close()

	return strings.TrimSpace(doc.Editable().GetContent()), nil
}
