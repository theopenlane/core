package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
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

// ParseDocument parses a document based on its MIME type
func ParseDocument(reader io.Reader, mimeType string) (any, error) {
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
		return result, nil

	case strings.Contains(mimeType, "yaml") || strings.Contains(mimeType, "yml"):
		var result any
		if err := yaml.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrYAMLParseFailed, err)
		}
		return result, nil

	case strings.Contains(mimeType, "text/plain"):
		return string(data), nil

	default:
		return data, nil
	}
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
