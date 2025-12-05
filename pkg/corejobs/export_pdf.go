package corejobs

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// marshalToPdfFormatFull converts nodes to PDF format with proper styling and layout
func marshalToPdfFormatFull(nodes []map[string]any, metadata *ExportMetadata) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	// For now, we'll create PDF by converting markdown-like content to PDF
	return createPDFFromContent(nodes, metadata)
}

// createPDFFromContent creates a PDF document from the given content
func createPDFFromContent(nodes []map[string]any, metadata *ExportMetadata) ([]byte, error) {
	var content strings.Builder

	// Add title page
	if metadata != nil && metadata.Title != "" {
		content.WriteString(fmt.Sprintf("TITLE: %s\n\n", metadata.Title))
	}

	// Add metadata
	if metadata != nil {
		if metadata.CreatedAt != "" {
			content.WriteString(fmt.Sprintf("Created: %s\n", metadata.CreatedAt))
		}
		if metadata.CreatedBy != "" {
			content.WriteString(fmt.Sprintf("Created By: %s\n", metadata.CreatedBy))
		}
		if metadata.ExportType != "" {
			content.WriteString(fmt.Sprintf("Type: %s\n\n", metadata.ExportType.String()))
		}
	}

	// For single document exports with rich formatting
	if len(nodes) == 1 {
		node := nodes[0]

		// Add document name as heading
		if name, ok := node["name"].(string); ok && name != "" {
			content.WriteString(fmt.Sprintf("== %s ==\n\n", name))
		}

		// Add details section
		if details, ok := node["details"].(string); ok && details != "" {
			content.WriteString("DETAILS\n")
			detailText := cleanHTML(details)
			content.WriteString(detailText)
			content.WriteString("\n\n")
		}

		// Add metadata section
		if len(node) > 0 {
			content.WriteString("METADATA\n")
			for key, value := range node {
				// Skip details and name as we already showed them
				if key == "details" || key == "name" {
					continue
				}
				content.WriteString(fmt.Sprintf("%s: %s\n", key, cleanHTML(value)))
			}
		}
	} else {
		// For multiple items
		content.WriteString("EXPORT CONTENTS\n\n")

		// Add index
		for i, node := range nodes {
			if name, ok := node["name"].(string); ok && name != "" {
				content.WriteString(fmt.Sprintf("%d. %s\n", i+1, name))
			} else {
				content.WriteString(fmt.Sprintf("%d. Item %d\n", i+1, i+1))
			}
		}
		content.WriteString("\n\n")

		// Add details for each item
		for i, node := range nodes {
			if name, ok := node["name"].(string); ok && name != "" {
				content.WriteString(fmt.Sprintf("== Section %d: %s ==\n\n", i+1, name))
			} else {
				content.WriteString(fmt.Sprintf("== Section %d ==\n\n", i+1))
			}

			for key, value := range node {
				content.WriteString(fmt.Sprintf("%s: %s\n", key, cleanHTML(value)))
			}
			content.WriteString("\n")
		}
	}

	// Build a minimal but valid PDF from plain text
	return createSimplePDF(content.String()), nil
}

// createSimplePDF creates a simple, single-page PDF from text content.
// It manually constructs a minimal valid PDF with a single Helvetica font
// and writes each line of text starting near the top-left of the page.
func createSimplePDF(content string) []byte {
	data, err := WriteStringToNewPDFFile(content)
	if err != nil {
		log.Error().Err(err).Msg("failed to build PDF bytes")
		return []byte{}
	}
	return data
}

// WriteStringToNewPDFFile builds a minimal, valid PDF from the given text.
// It does NOT depend on pdfcpu – everything is constructed manually.
// One page, A4 (612x792 points), Helvetica, 12pt.
func WriteStringToNewPDFFile(content string) ([]byte, error) {
	var buf bytes.Buffer

	// Helper closure
	write := func(s string) {
		_, _ = buf.WriteString(s)
	}

	// PDF header
	write("%PDF-1.4\n")

	// Offsets for xref table
	offsets := make([]int, 6)  

	// 1 0 obj: Catalog
	offsets[1] = buf.Len()
	write("1 0 obj\n")
	write("<< /Type /Catalog /Pages 2 0 R >>\n")
	write("endobj\n")

	// 2 0 obj: Pages
	offsets[2] = buf.Len()
	write("2 0 obj\n")
	write("<< /Type /Pages /Kids [3 0 R] /Count 1 >>\n")
	write("endobj\n")

	// 3 0 obj: Page
	offsets[3] = buf.Len()
	write("3 0 obj\n")
	write("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\n")
	write("endobj\n")

	// 4 0 obj: Content stream
	var contentBuf bytes.Buffer
	_, _ = contentBuf.WriteString("BT\n")
	_, _ = contentBuf.WriteString("/F1 12 Tf\n")
	_, _ = contentBuf.WriteString("72 750 Td\n") // start near top-left

	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Escape characters that are special in PDF string literals
		line = escapePDFString(line)
		if line == "" {
			// Even empty lines should move cursor down.
			if i == 0 {
				// For the first line if it's empty, just move down once.
				_, _ = contentBuf.WriteString("0 -14 Td\n")
			} else {
				_, _ = contentBuf.WriteString("0 -14 Td\n")
			}
			continue
		}

		// Draw text and then move down for next line
		_, _ = contentBuf.WriteString(fmt.Sprintf("(%s) Tj\n", line))
		// Move down by 14 points for next line
		_, _ = contentBuf.WriteString("0 -14 Td\n")
	}

	_, _ = contentBuf.WriteString("ET\n")

	streamBytes := contentBuf.Bytes()

	offsets[4] = buf.Len()
	write("4 0 obj\n")
	write(fmt.Sprintf("<< /Length %d >>\n", len(streamBytes)))
	write("stream\n")
	_, _ = buf.Write(streamBytes)
	write("\nendstream\n")
	write("endobj\n")

	// 5 0 obj: Font
	offsets[5] = buf.Len()
	write("5 0 obj\n")
	write("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n")
	write("endobj\n")

	// xref table
	xrefOffset := buf.Len()
	write("xref\n")
	write("0 6\n")
	write("0000000000 65535 f \n")

	for i := 1; i <= 5; i++ {
		write(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}

	// trailer
	write("trailer\n")
	write("<< /Size 6 /Root 1 0 R >>\n")
	write("startxref\n")
	write(fmt.Sprintf("%d\n", xrefOffset))
	write("%%EOF\n")

	return buf.Bytes(), nil
}

// escapePDFString escapes characters that are special in PDF string literals.
func escapePDFString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	return s
}
