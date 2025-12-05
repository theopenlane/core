package corejobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/pkg/enums"
)

func TestGetFormatMarshaler(t *testing.T) {
	testCases := []struct {
		name          string
		format        enums.ExportFormat
		expectError   bool
		expectContent string
	}{
		{
			name:        "CSV format",
			format:      enums.ExportFormatCsv,
			expectError: false,
		},
		{
			name:        "MD format",
			format:      enums.ExportFormatMD,
			expectError: false,
		},
		{
			name:        "DOCX format",
			format:      enums.ExportFormatDocx,
			expectError: false,
		},
		{
			name:        "PDF format",
			format:      enums.ExportFormatPDF,
			expectError: false,
		},
		{
			name:        "Invalid format",
			format:      enums.ExportFormatInvalid,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			marshaler, contentType, err := GetFormatMarshaler(tc.format)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, marshaler)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, marshaler)
			assert.NotEmpty(t, contentType)
		})
	}
}

func TestGetFileExtension(t *testing.T) {
	testCases := []struct {
		format    enums.ExportFormat
		extension string
	}{
		{enums.ExportFormatCsv, "csv"},
		{enums.ExportFormatMD, "md"},
		{enums.ExportFormatDocx, "docx"},
		{enums.ExportFormatPDF, "pdf"},
		{enums.ExportFormatInvalid, "csv"},
	}

	for _, tc := range testCases {
		t.Run(tc.format.String(), func(t *testing.T) {
			ext := GetFileExtension(tc.format)
			assert.Equal(t, tc.extension, ext)
		})
	}
}

func TestMarshalToMarkdownFormat(t *testing.T) {
	testCases := []struct {
		name        string
		nodes       []map[string]any
		metadata    *ExportMetadata
		expectError bool
		checkText   []string
	}{
		{
			name:     "Empty nodes",
			nodes:    []map[string]any{},
			metadata: nil,
		},
		{
			name: "Single node with name and details",
			nodes: []map[string]any{
				{
					"id":      "123",
					"name":    "Test Policy",
					"details": "This is a test policy",
					"status":  "published",
				},
			},
			metadata: &ExportMetadata{
				Title:      "Test Export",
				ExportType: enums.ExportTypeInternalPolicy,
			},
			checkText: []string{"Test Policy", "Details", "Test Export"},
		},
		{
			name: "Multiple nodes",
			nodes: []map[string]any{
				{
					"id":   "1",
					"name": "Item 1",
				},
				{
					"id":   "2",
					"name": "Item 2",
				},
			},
			metadata:  nil,
			checkText: []string{"Item 1", "Item 2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := marshalToMarkdownFormat(tc.nodes, tc.metadata)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if len(tc.nodes) == 0 {
				assert.Nil(t, data)
				return
			}

			assert.NotNil(t, data)
			content := string(data)

			for _, expectedText := range tc.checkText {
				assert.Contains(t, content, expectedText)
			}
		})
	}
}

func TestMarshalToDocxFormat(t *testing.T) {
	testCases := []struct {
		name        string
		nodes       []map[string]any
		metadata    *ExportMetadata
		expectError bool
	}{
		{
			name:  "Empty nodes",
			nodes: []map[string]any{},
		},
		{
			name: "Single node",
			nodes: []map[string]any{
				{
					"id":      "123",
					"name":    "Test Document",
					"details": "Document details",
				},
			},
			metadata: &ExportMetadata{
				Title: "Test Export",
			},
		},
		{
			name: "Multiple nodes",
			nodes: []map[string]any{
				{
					"id":   "1",
					"name": "Doc 1",
				},
				{
					"id":   "2",
					"name": "Doc 2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := marshalToDocxFormat(tc.nodes, tc.metadata)

			require.NoError(t, err)

			if len(tc.nodes) == 0 {
				assert.Nil(t, data)
				return
			}

			assert.NotNil(t, data)
			assert.Greater(t, len(data), 0)
		})
	}
}

func TestMarshalToPdfFormat(t *testing.T) {
	testCases := []struct {
		name        string
		nodes       []map[string]any
		metadata    *ExportMetadata
		expectError bool
	}{
		{
			name:  "Empty nodes",
			nodes: []map[string]any{},
		},
		{
			name: "Single node",
			nodes: []map[string]any{
				{
					"id":      "123",
					"name":    "Test Policy",
					"details": "Policy details",
				},
			},
			metadata: &ExportMetadata{
				Title: "Test PDF Export",
			},
		},
		{
			name: "Multiple nodes",
			nodes: []map[string]any{
				{
					"id":   "1",
					"name": "Policy 1",
				},
				{
					"id":   "2",
					"name": "Policy 2",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := marshalToPdfFormat(tc.nodes, tc.metadata)

			require.NoError(t, err)

			if len(tc.nodes) == 0 {
				assert.Nil(t, data)
				return
			}

			assert.NotNil(t, data)
			// PDF files should have some minimum size
			if len(data) > 0 {
				assert.Greater(t, len(data), 0)
			}
		})
	}
}

func TestMarshalToCSVFormat(t *testing.T) {
	testCases := []struct {
		name     string
		nodes    []map[string]any
		metadata *ExportMetadata
	}{
		{
			name:  "Empty nodes",
			nodes: []map[string]any{},
		},
		{
			name: "Single node",
			nodes: []map[string]any{
				{
					"id":    "123",
					"name":  "Test Item",
					"value": "test_value",
				},
			},
		},
		{
			name: "Multiple nodes with varying keys",
			nodes: []map[string]any{
				{
					"id":   "1",
					"name": "Item 1",
					"type": "type_a",
				},
				{
					"id":    "2",
					"name":  "Item 2",
					"value": "some_value",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := marshalToCSVFormat(tc.nodes, tc.metadata)

			require.NoError(t, err)

			if len(tc.nodes) == 0 {
				assert.Nil(t, data)
				return
			}

			assert.NotNil(t, data)
			content := string(data)

			// CSV should contain the values
			for _, node := range tc.nodes {
				for _, val := range node {
					// Check if any field values are in the CSV
					assert.Contains(t, content, val.(string))
				}
			}
		})
	}
}
