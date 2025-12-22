package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/nguyenthenguyen/docx"
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
	default:
		return &ParsedDocument{Data: data}, nil
	}
}

// parseDocx extracts and returns the text content from a DOCX file
func parseDocx(content []byte) (string, error) {
	reader := bytes.NewReader(content)

	doc, err := docx.ReadDocxFromMemory(reader, int64(len(content)))
	if err != nil {
		return "", fmt.Errorf("failed to read docx file: %w", err) //nolint:err113
	}

	defer doc.Close()

	return strings.TrimSpace(doc.Editable().GetContent()), nil
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
