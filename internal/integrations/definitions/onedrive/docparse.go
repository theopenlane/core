package onedrive

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"
)

const wordMLNamespace = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

// xmlBlock is a single renderable document element produced by the Word XML parser
type xmlBlock struct {
	// htmlContent holds inline HTML (may include <strong>, <em>) or a pre-rendered table
	htmlContent string
	heading     int  // 0 = paragraph, 1–6 = heading level
	listItem    bool
	listLevel   int // 0-based indent
	listNumID   int // Word numId — used to group items into the same list
	ordered     bool
	isTable     bool // when true, htmlContent is a complete <table>…</table> fragment
}

type tableXMLCell struct {
	content string
	span    int // colspan from w:gridSpan (0 or 1 means no span)
}

// extractDocxHTML opens docxBytes as a DOCX ZIP archive, parses word/numbering.xml for
// list ordering information, and converts word/document.xml to HTML.
func extractDocxHTML(docxBytes []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		return "", fmt.Errorf("not a valid zip: %w", err)
	}

	numOrdered := parseNumberingXMLFromZip(zr)

	for _, f := range zr.File {
		if f.Name != "word/document.xml" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return "", err
		}

		defer rc.Close()

		return parseWordDocumentToHTML(rc, numOrdered)
	}

	return "", fmt.Errorf("word/document.xml not found in docx") //nolint:err113
}

func parseNumberingXMLFromZip(zr *zip.Reader) map[int]bool {
	for _, f := range zr.File {
		if f.Name != "word/numbering.xml" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil
		}

		defer rc.Close()

		return parseNumberingXML(rc)
	}

	return nil
}

// parseNumberingXML reads word/numbering.xml and returns a map from Word numId to
// isOrdered. It inspects the numFmt of indent level 0 for each abstractNum to determine
// whether the list is ordered (decimal, roman, etc.) or unordered (bullet).
func parseNumberingXML(r io.Reader) map[int]bool {
	dec := xml.NewDecoder(r)

	abstractOrdered := map[int]bool{} // abstractNumId → ordered
	numAbstract := map[int]int{}      // numId → abstractNumId

	currentAbstractID := -1
	currentNumID := -1
	inLvl0 := false

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			if elem.Name.Space != wordMLNamespace {
				continue
			}

			switch elem.Name.Local {
			case "abstractNum":
				currentAbstractID = -1

				for _, a := range elem.Attr {
					if a.Name.Local == "abstractNumId" {
						if n, convErr := strconv.Atoi(a.Value); convErr == nil {
							currentAbstractID = n
						}
					}
				}

			case "lvl":
				inLvl0 = false

				for _, a := range elem.Attr {
					if a.Name.Local == "ilvl" && a.Value == "0" {
						inLvl0 = true
					}
				}

			case "numFmt":
				if inLvl0 && currentAbstractID >= 0 {
					for _, a := range elem.Attr {
						if a.Name.Local == "val" {
							abstractOrdered[currentAbstractID] = isOrderedNumFmt(a.Value)
						}
					}
				}

			case "num":
				currentNumID = -1

				for _, a := range elem.Attr {
					if a.Name.Local == "numId" {
						if n, convErr := strconv.Atoi(a.Value); convErr == nil {
							currentNumID = n
						}
					}
				}

			case "abstractNumId":
				if currentNumID >= 0 {
					for _, a := range elem.Attr {
						if a.Name.Local == "val" {
							if n, convErr := strconv.Atoi(a.Value); convErr == nil {
								numAbstract[currentNumID] = n
							}
						}
					}
				}
			}

		case xml.EndElement:
			if elem.Name.Space != wordMLNamespace {
				continue
			}

			switch elem.Name.Local {
			case "abstractNum":
				currentAbstractID = -1
			case "lvl":
				inLvl0 = false
			case "num":
				currentNumID = -1
			}
		}
	}

	result := make(map[int]bool, len(numAbstract))
	for numID, abstractID := range numAbstract {
		result[numID] = abstractOrdered[abstractID]
	}

	return result
}

func isOrderedNumFmt(val string) bool {
	switch val {
	case "decimal", "upperRoman", "lowerRoman", "upperLetter", "lowerLetter",
		"ordinal", "cardinalText", "ordinalText", "decimalZero":
		return true
	}

	return false
}

// parseWordDocumentToHTML streams word/document.xml and produces HTML with headings,
// bold, italic, ordered/unordered lists, colspan, and tables.
func parseWordDocumentToHTML(r io.Reader, numOrdered map[int]bool) (string, error) {
	dec := xml.NewDecoder(r)

	var blocks []xmlBlock

	// paragraph state (only active when !inTable)
	var (
		inPara    bool
		inPPr     bool
		inNumPr   bool
		paraRuns  []string
		paraStyle string
		paraNumPr bool
		paraIlvl  int
		paraNumID int
	)

	// table state
	var (
		inTable   bool
		tableRows [][]tableXMLCell
		rowCells  []tableXMLCell
		inCell    bool
		inTcPr    bool
		cellRuns  []string
		cellSpan  int
	)

	// run state
	var (
		inRun     bool
		inRPr     bool
		runBold   bool
		runItalic bool
	)

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}

		if err != nil {
			continue
		}

		switch elem := tok.(type) {
		case xml.StartElement:
			if elem.Name.Space != wordMLNamespace {
				continue
			}

			switch elem.Name.Local {
			// ── table ────────────────────────────────────────────────────
			case "tbl":
				inTable = true
				tableRows = nil

			case "tr":
				rowCells = nil

			case "tc":
				inCell = true
				inTcPr = false
				cellRuns = nil
				cellSpan = 1

			case "tcPr":
				inTcPr = true

			case "gridSpan":
				if inTcPr {
					for _, a := range elem.Attr {
						if a.Name.Local == "val" {
							if n, convErr := strconv.Atoi(a.Value); convErr == nil && n > 1 {
								cellSpan = n
							}
						}
					}
				}

			// ── paragraph ─────────────────────────────────────────────────
			case "p":
				if !inTable {
					inPara = true
					paraRuns = nil
					paraStyle = ""
					paraNumPr = false
					paraIlvl = 0
					paraNumID = 0
				}

			case "pPr":
				inPPr = true

			case "pStyle":
				if inPPr {
					for _, a := range elem.Attr {
						if a.Name.Local == "val" {
							paraStyle = a.Value
						}
					}
				}

			case "numPr":
				if inPPr {
					inNumPr = true
					paraNumPr = true
				}

			case "ilvl":
				if inNumPr {
					for _, a := range elem.Attr {
						if a.Name.Local == "val" {
							if n, convErr := strconv.Atoi(a.Value); convErr == nil {
								paraIlvl = n
							}
						}
					}
				}

			case "numId":
				if inNumPr {
					for _, a := range elem.Attr {
						if a.Name.Local == "val" {
							if n, convErr := strconv.Atoi(a.Value); convErr == nil {
								paraNumID = n
							}
						}
					}
				}

			// ── run ───────────────────────────────────────────────────────
			case "r":
				inRun = true
				runBold = false
				runItalic = false

			case "rPr":
				if inRun {
					inRPr = true
				}

			case "b":
				if inRPr {
					runBold = toggleOn(elem.Attr)
				}

			case "i":
				if inRPr {
					runItalic = toggleOn(elem.Attr)
				}

			// ── text content ──────────────────────────────────────────────
			case "t":
				var content struct {
					Text string `xml:",chardata"`
				}

				if decErr := dec.DecodeElement(&content, &elem); decErr != nil || content.Text == "" {
					continue
				}

				piece := html.EscapeString(content.Text)

				if runItalic {
					piece = "<em>" + piece + "</em>"
				}

				if runBold {
					piece = "<strong>" + piece + "</strong>"
				}

				if inCell {
					cellRuns = append(cellRuns, piece)
				} else if inPara {
					paraRuns = append(paraRuns, piece)
				}
			}

		case xml.EndElement:
			if elem.Name.Space != wordMLNamespace {
				continue
			}

			switch elem.Name.Local {
			case "rPr":
				inRPr = false

			case "r":
				inRun = false

			case "pPr":
				inPPr = false
				inNumPr = false

			case "numPr":
				inNumPr = false

			case "tcPr":
				inTcPr = false

			case "p":
				if inPara && !inTable {
					content := strings.Join(paraRuns, "")
					if strings.TrimSpace(stripTags(content)) != "" {
						blocks = append(blocks, xmlBlock{
							htmlContent: content,
							heading:     xmlHeadingLevel(paraStyle),
							listItem:    paraNumPr && paraNumID > 0,
							listLevel:   paraIlvl,
							listNumID:   paraNumID,
							ordered:     paraNumPr && numOrdered[paraNumID],
						})
					}

					inPara = false
				}

			case "tc":
				if inCell {
					rowCells = append(rowCells, tableXMLCell{
						content: strings.Join(cellRuns, ""),
						span:    cellSpan,
					})
					inCell = false
				}

			case "tr":
				if len(rowCells) > 0 {
					tableRows = append(tableRows, rowCells)
				}

			case "tbl":
				if len(tableRows) > 0 {
					blocks = append(blocks, xmlBlock{
						htmlContent: renderXMLTable(tableRows),
						isTable:     true,
					})
				}

				inTable = false
			}
		}
	}

	return renderBlocks(blocks), nil
}

// toggleOn returns false only when a Word toggle element explicitly sets w:val="false" or
// w:val="0", which overrides the default-on behaviour of elements like <w:b/> and <w:i/>
func toggleOn(attrs []xml.Attr) bool {
	for _, a := range attrs {
		if a.Name.Local == "val" && (a.Value == "false" || a.Value == "0") {
			return false
		}
	}

	return true
}

// stripTags is used only to test whether formatted content is visually empty after
// removing any inline HTML tags added by the run formatter
func stripTags(s string) string {
	var b strings.Builder

	inTag := false

	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}

	return b.String()
}

// renderBlocks converts the flat slice of xmlBlocks to HTML, managing the open/close of
// <ul>/<ol> elements as list items are encountered and the list context changes.
func renderBlocks(blocks []xmlBlock) string {
	type listEntry struct {
		numID   int
		level   int
		ordered bool
	}

	var listStack []listEntry
	var b strings.Builder

	closeListsTo := func(depth int) {
		for len(listStack) > depth {
			top := listStack[len(listStack)-1]
			listStack = listStack[:len(listStack)-1]

			if top.ordered {
				b.WriteString("</ol>\n")
			} else {
				b.WriteString("</ul>\n")
			}
		}
	}

	closeLists := func() { closeListsTo(0) }

	for _, block := range blocks {
		if block.isTable {
			closeLists()
			b.WriteString(block.htmlContent)
			continue
		}

		if block.heading > 0 {
			closeLists()
			tag := fmt.Sprintf("h%d", block.heading)
			b.WriteString("<" + tag + ">" + block.htmlContent + "</" + tag + ">\n")
			continue
		}

		if !block.listItem {
			closeLists()
			b.WriteString("<p>" + block.htmlContent + "</p>\n")
			continue
		}

		// List item: adjust the open list stack to match the current item's context
		if len(listStack) > 0 {
			top := listStack[len(listStack)-1]

			if top.numID != block.listNumID {
				// Different list definition — close everything and start fresh
				closeLists()
			} else if block.listLevel < top.level {
				// Dedenting — close inner levels
				closeListsTo(block.listLevel + 1)
			}
		}

		// Open new levels until we reach the item's indent depth
		for len(listStack) <= block.listLevel {
			entry := listEntry{
				numID:   block.listNumID,
				level:   len(listStack),
				ordered: block.ordered,
			}
			listStack = append(listStack, entry)

			if block.ordered {
				b.WriteString("<ol>\n")
			} else {
				b.WriteString("<ul>\n")
			}
		}

		b.WriteString("<li>" + block.htmlContent + "</li>\n")
	}

	closeLists()

	return b.String()
}

func renderXMLTable(rows [][]tableXMLCell) string {
	if len(rows) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString("<table>\n")

	for i, row := range rows {
		b.WriteString("<tr>")

		for _, cell := range row {
			// treat the first row as column headers
			tag := "td"
			if i == 0 {
				tag = "th"
			}

			open := "<" + tag

			if cell.span > 1 {
				open += fmt.Sprintf(` colspan="%d"`, cell.span)
			}

			b.WriteString(open + ">" + cell.content + "</" + tag + ">")
		}

		b.WriteString("</tr>\n")
	}

	b.WriteString("</table>\n")

	return b.String()
}

// xmlHeadingLevel maps a Word paragraph style name to an HTML heading level (1–6),
// returning 0 for body text
func xmlHeadingLevel(style string) int {
	s := strings.ToLower(strings.TrimSpace(style))

	switch {
	case s == "title":
		return 1
	case strings.HasPrefix(s, "heading"):
		n := strings.TrimPrefix(s, "heading")
		if len(n) == 1 && n[0] >= '1' && n[0] <= '6' {
			return int(n[0] - '0')
		}
	}

	return 0
}
