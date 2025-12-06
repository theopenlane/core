package corejobs

import (
	"bytes"
	"fmt"

	"github.com/theopenlane/core/pkg/enums"
)

// FormatMarshalerFunc is a function that marshals nodes to a specific format
type FormatMarshalerFunc func(nodes []map[string]any, metadata *ExportMetadata) ([]byte, error)

// ExportMetadata contains metadata for export formatting
type ExportMetadata struct {
	// Title of the export (e.g., document name)
	Title string
	// Description of the export
	Description string
	// ExportType determines how to format the data
	ExportType enums.ExportType
	// CreatedAt is the creation timestamp
	CreatedAt string
	// CreatedBy is the user who created the export
	CreatedBy string
}

// GetFormatMarshaler returns the appropriate marshaler function for the given format
func GetFormatMarshaler(format enums.ExportFormat) (FormatMarshalerFunc, string, error) {
	switch format {
	case enums.ExportFormatCsv:
		return marshalToCSVFormat, "text/csv", nil
	case enums.ExportFormatMD:
		return marshalToMarkdownFormat, "text/markdown", nil
	case enums.ExportFormatDocx:
		return marshalToDocxFormat, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", nil
	case enums.ExportFormatPDF:
		return marshalToPdfFormat, "application/pdf", nil
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// GetFileExtension returns the file extension for the given format
func GetFileExtension(format enums.ExportFormat) string {
	switch format {
	case enums.ExportFormatCsv:
		return "csv"
	case enums.ExportFormatMD:
		return "md"
	case enums.ExportFormatDocx:
		return "docx"
	case enums.ExportFormatPDF:
		return "pdf"
	default:
		return "csv"
	}
}

// marshalToCSVFormat converts nodes to CSV format using existing CSV marshaler
func marshalToCSVFormat(nodes []map[string]any, metadata *ExportMetadata) ([]byte, error) {
	// Create a temporary worker just for CSV marshaling
	w := &ExportContentWorker{}
	return w.marshalToCSV(nodes)
}

// marshalToMarkdownFormat converts nodes to Markdown format
func marshalToMarkdownFormat(nodes []map[string]any, metadata *ExportMetadata) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer

	// Write frontmatter (YAML metadata)
	buf.WriteString("---\n")
	if metadata != nil && metadata.Title != "" {
		fmt.Fprintf(&buf, "title: %s\n", metadata.Title)
	}
	if metadata != nil && metadata.Description != "" {
		fmt.Fprintf(&buf, "description: %s\n", metadata.Description)
	}
	if metadata != nil && metadata.CreatedAt != "" {
		fmt.Fprintf(&buf, "created_at: %s\n", metadata.CreatedAt)
	}
	if metadata != nil && metadata.CreatedBy != "" {
		fmt.Fprintf(&buf, "created_by: %s\n", metadata.CreatedBy)
	}
	buf.WriteString("---\n\n")

	// For single document exports
	if len(nodes) == 1 {
		node := nodes[0]
		return marshalSingleNodeToMarkdown(&buf, node, metadata)
	}

	// For multiple items, create a table of contents style markdown
	buf.WriteString("# Export Contents\n\n")
	for i, node := range nodes {
		fmt.Fprintf(&buf, "## Item %d\n\n", i+1)
		marshalNodeToMarkdownTable(&buf, node)
		buf.WriteString("\n\n")
	}

	return buf.Bytes(), nil
}

// marshalSingleNodeToMarkdown formats a single node as markdown content
func marshalSingleNodeToMarkdown(buf *bytes.Buffer, node map[string]any, metadata *ExportMetadata) ([]byte, error) {
	// Use the name field as the main heading if available
	if name, ok := node["name"].(string); ok && name != "" {
		fmt.Fprintf(buf, "# %s\n\n", name)
	}

	// Write details if available
	if details, ok := node["details"].(string); ok && details != "" {
		buf.WriteString("## Details\n\n")
		buf.WriteString(cleanHTML(details))
		buf.WriteString("\n\n")
	}

	// Write all other fields as a metadata table
	if len(node) > 0 {
		buf.WriteString("## Metadata\n\n")
		buf.WriteString("| Field | Value |\n")
		buf.WriteString("|-------|-------|\n")

		for key, value := range node {
			// Skip details as we already showed it
			if key == "details" || key == "name" {
				continue
			}
			val := cleanHTML(value)
			fmt.Fprintf(buf, "| %s | %s |\n", key, val)
		}
	}

	return buf.Bytes(), nil
}

// marshalNodeToMarkdownTable writes a node as a markdown table
func marshalNodeToMarkdownTable(buf *bytes.Buffer, node map[string]any) {
	if len(node) == 0 {
		return
	}

	buf.WriteString("| Field | Value |\n")
	buf.WriteString("|-------|-------|\n")

	for key, value := range node {
		val := cleanHTML(value)
		fmt.Fprintf(buf, "| %s | %s |\n", key, val)
	}
}

// marshalToDocxFormat converts nodes to DOCX format
func marshalToDocxFormat(nodes []map[string]any, metadata *ExportMetadata) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	// Use the full DOCX marshaling implementation
	return marshalToDocxFormatFull(nodes, metadata)
}

// marshalToPdfFormat converts nodes to PDF format
func marshalToPdfFormat(nodes []map[string]any, metadata *ExportMetadata) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	// Use the full PDF marshaling implementation
	return marshalToPdfFormatFull(nodes, metadata)
}
