package corejobs

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gomutex/godocx"
	"github.com/gomutex/godocx/docx"
)

func marshalToDocxFormatFull(nodes []map[string]any, metadata *ExportMetadata) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	// Create a new DOCX document
	doc, err := godocx.NewDocument()
	if err != nil {
		return nil, fmt.Errorf("failed to create DOCX document: %w", err)
	}

	// Optional: title at top
	if metadata != nil && metadata.Title != "" {
		// Heading level 0 = Title in godocx
		if _, err := doc.AddHeading(metadata.Title, 0); err != nil {
			return nil, fmt.Errorf("failed to add title heading: %w", err)
		}
	}

	// metadata block at top
	if metadata != nil {
		if err := addDocxMetadata(doc, metadata); err != nil {
			return nil, err
		}
	}

	// SINGLE DOCUMENT EXPORT
	if len(nodes) == 1 {
		node := nodes[0]

		// Document name as heading
		if name, ok := node["name"].(string); ok && name != "" {
			if _, err := doc.AddHeading(name, 1); err != nil {
				return nil, fmt.Errorf("failed to add node heading: %w", err)
			}
		}

		// Details section
		if details, ok := node["details"].(string); ok && details != "" {
			if _, err := doc.AddHeading("Details", 2); err != nil {
				return nil, fmt.Errorf("failed to add details heading: %w", err)
			}

			detailText := cleanHTML(details)
			for _, line := range strings.Split(detailText, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				p := doc.AddParagraph(line)
				_ = p // if you want, you can format the paragraph/run here
			}
		}

		// Metadata table for remaining fields
		if len(node) > 0 {
			if _, err := doc.AddHeading("Metadata", 2); err != nil {
				return nil, fmt.Errorf("failed to add metadata heading: %w", err)
			}

			tbl := doc.AddTable()
			// Optional style: must exist in default template
			tbl.Style("TableGrid")

			// Header row
			hdr := tbl.AddRow()
			hdr.AddCell().AddParagraph("Field")
			hdr.AddCell().AddParagraph("Value")

			for key, value := range node {
				if key == "details" || key == "name" {
					continue
				}

				row := tbl.AddRow()
				row.AddCell().AddParagraph(key)
				row.AddCell().AddParagraph(cleanHTML(value))
			}
		}
	} else {
		// High level "Export Contents" heading
		if _, err := doc.AddHeading("Export Contents", 1); err != nil {
			return nil, fmt.Errorf("failed to add export contents heading: %w", err)
		}

		// Index list
		for i, node := range nodes {
			var label string
			if name, ok := node["name"].(string); ok && name != "" {
				label = fmt.Sprintf("%d. %s", i+1, name)
			} else {
				label = fmt.Sprintf("%d. Item %d", i+1, i+1)
			}
			doc.AddParagraph(label)
		}

		// Per-item sections
		for i, node := range nodes {
			doc.AddPageBreak() // separate each item

			// Section heading
			var sectionTitle string
			if name, ok := node["name"].(string); ok && name != "" {
				sectionTitle = fmt.Sprintf("Section %d: %s", i+1, name)
			} else {
				sectionTitle = fmt.Sprintf("Section %d", i+1)
			}
			if _, err := doc.AddHeading(sectionTitle, 2); err != nil {
				return nil, fmt.Errorf("failed to add section heading: %w", err)
			}

			// Optional details block in each section
			if details, ok := node["details"].(string); ok && details != "" {
				if _, err := doc.AddHeading("Details", 3); err != nil {
					return nil, fmt.Errorf("failed to add section details heading: %w", err)
				}
				detailText := cleanHTML(details)
				for _, line := range strings.Split(detailText, "\n") {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					doc.AddParagraph(line)
				}
			}

			// Table for all fields
			tbl := doc.AddTable()
			tbl.Style("TableGrid")

			hdr := tbl.AddRow()
			hdr.AddCell().AddParagraph("Field")
			hdr.AddCell().AddParagraph("Value")

			for key, value := range node {
				row := tbl.AddRow()
				row.AddCell().AddParagraph(key)
				row.AddCell().AddParagraph(cleanHTML(value))
			}
		}
	}

	// Write document to bytes
	var buf bytes.Buffer
	if err := doc.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write DOCX document: %w", err)
	}

	return buf.Bytes(), nil
}

// addDocxMetadata adds metadata section to DOCX document
func addDocxMetadata(doc *docx.RootDoc, metadata *ExportMetadata) error {
	if metadata == nil {
		return nil
	}

	if metadata.Title != "" {
		doc.AddParagraph(fmt.Sprintf("Title: %s", metadata.Title))
	}

	if metadata.CreatedAt != "" {
		doc.AddParagraph(fmt.Sprintf("Created: %s", metadata.CreatedAt))
	}

	if metadata.CreatedBy != "" {
		doc.AddParagraph(fmt.Sprintf("Created By: %s", metadata.CreatedBy))
	}

	if metadata.ExportType != "" {
		doc.AddParagraph(fmt.Sprintf("Type: %s", metadata.ExportType.String()))
	}

	return nil
}
