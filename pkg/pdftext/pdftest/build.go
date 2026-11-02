// Package pdftest builds small pdfs in memory for tests that need a real text layer
package pdftest

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	// firstPageObject is the object number of the first page; 1 is the catalog and 2 the page tree
	firstPageObject = 3
	// objectsPerPage is the page dict plus its content stream
	objectsPerPage = 2
)

// layout is how a content stream draws its lines of text
type layout int

const (
	// layoutSimple shows each line as one literal string in a Type1 font
	layoutSimple layout = iota
	// layoutComposite shows each line as one hex string of two byte glyph ids
	layoutComposite
	// layoutGlyphs shows every glyph with its own Tj and steps to the next with Td
	layoutGlyphs
)

const (
	// fontSize is the size every test page sets its font to
	fontSize = 11
	// defaultGlyphWidth is the composite font's DW
	defaultGlyphWidth = 500
	// digitGlyphWidth is the composite font's W entry for 0 through 9
	digitGlyphWidth = 450
	// narrowGlyphWidth is the composite font's W entry for I
	narrowGlyphWidth = 250
	// wordGap is the extra Td distance a glyph layout leaves for a space, a quarter of the font size
	wordGap = fontSize / 4.0
	// glyphSpaceUnits is how many glyph space units make one unit of text space
	glyphSpaceUnits = 1000
)

// Build writes a pdf with one page per entry, each line of text on its own row; composite
// selects a Type0 font with a ToUnicode map so the text is encoded as two byte glyph ids
func Build(pages [][]string, composite bool) []byte {
	if composite {
		return build(pages, layoutComposite)
	}

	return build(pages, layoutSimple)
}

// BuildGlyphs writes a pdf in the composite font that places every glyph on its own, leaving word
// gaps only in the Td distance, the way some html to pdf renderers lay out text
func BuildGlyphs(pages [][]string) []byte {
	return build(pages, layoutGlyphs)
}

func build(pages [][]string, mode layout) []byte {
	var objects []string

	kids := make([]string, 0, len(pages))
	fontObject := firstPageObject + objectsPerPage*len(pages)

	for i := range pages {
		pageObject := firstPageObject + objectsPerPage*i
		contentObject := pageObject + 1
		kids = append(kids, fmt.Sprintf("%d 0 R", pageObject))

		objects = append(objects,
			fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R >>", fontObject, contentObject),
			contentStream(pages[i], mode),
		)
	}

	if mode != layoutSimple {
		objects = append(objects, compositeFontObjects(fontObject)...)
	} else {
		objects = append(objects, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	}

	objects = append([]string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kids, " "), len(pages)),
	}, objects...)

	var out bytes.Buffer

	out.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects)+1)

	for i, object := range objects {
		offsets[i+1] = out.Len()
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, object)
	}

	xref := out.Len()

	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)

	for _, offset := range offsets[1:] {
		fmt.Fprintf(&out, "%010d 00000 n \n", offset)
	}

	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)

	return out.Bytes()
}

// Repeat builds a pdf with the same lines on every one of count pages, numbering each page
func Repeat(lines []string, count int) []byte {
	pages := make([][]string, 0, count)

	for i := 1; i <= count; i++ {
		pages = append(pages, append(append([]string{}, lines...), fmt.Sprintf("Page %d of %d", i, count)))
	}

	return Build(pages, false)
}

func compositeFontObjects(fontObject int) []string {
	cmap := strings.Join([]string{
		"/CIDInit /ProcSet findresource begin",
		"12 dict begin",
		"begincmap",
		"/CMapName /Test-Identity-UCS def",
		"1 begincodespacerange",
		"<0000> <FFFF>",
		"endcodespacerange",
		"1 beginbfrange",
		"<0020> <007E> <0020>",
		"endbfrange",
		"endcmap",
		"CMapName currentdict /CMap defineresource pop",
		"end end",
	}, "\n")

	descendantObject := fontObject + 1
	descriptorObject := descendantObject + 1
	toUnicodeObject := descriptorObject + 1

	return []string{
		fmt.Sprintf("<< /Type /Font /Subtype /Type0 /BaseFont /TestSans /Encoding /Identity-H /DescendantFonts [%d 0 R] /ToUnicode %d 0 R >>", descendantObject, toUnicodeObject),
		fmt.Sprintf("<< /Type /Font /Subtype /CIDFontType2 /BaseFont /TestSans /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /FontDescriptor %d 0 R /DW %d /W [48 57 %d 73 [%d]] >>", descriptorObject, defaultGlyphWidth, digitGlyphWidth, narrowGlyphWidth),
		"<< /Type /FontDescriptor /FontName /TestSans /Flags 32 /FontBBox [0 -200 1000 900] /ItalicAngle 0 /Ascent 900 /Descent -200 /CapHeight 700 /StemV 80 >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(cmap), cmap),
	}
}

func contentStream(lines []string, mode layout) string {
	var stream strings.Builder

	fmt.Fprintf(&stream, "BT /F1 %d Tf 72 740 Td 14 TL\n", fontSize)

	for _, line := range lines {
		switch mode {
		case layoutComposite:
			fmt.Fprintf(&stream, "<%s> Tj T*\n", twoByteHex(line))
		case layoutGlyphs:
			fmt.Fprintf(&stream, "%sT*\n", glyphLine(line))
		default:
			fmt.Fprintf(&stream, "(%s) Tj T*\n", strings.NewReplacer(`\`, `\\`, "(", `\(`, ")", `\)`).Replace(line))
		}
	}

	stream.WriteString("ET")

	return fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", stream.Len(), stream.String())
}

// glyphLine draws each glyph of a line with its own Tj, stepping past the previous glyph with Td
// and folding spaces into the step instead of drawing them
func glyphLine(line string) string {
	var (
		out  strings.Builder
		step float64
	)

	for _, r := range line {
		if r == ' ' {
			step += wordGap

			continue
		}

		fmt.Fprintf(&out, "%g 0 Td <%s> Tj ", step, twoByteHex(string(r)))
		step = glyphWidth(r) / glyphSpaceUnits * fontSize
	}

	return out.String()
}

// glyphWidth mirrors the composite font's W and DW entries
func glyphWidth(r rune) float64 {
	switch {
	case r >= '0' && r <= '9':
		return digitGlyphWidth
	case r == 'I':
		return narrowGlyphWidth
	default:
		return defaultGlyphWidth
	}
}

func twoByteHex(text string) string {
	var out strings.Builder

	for _, r := range text {
		if r < 0x20 || r > 0x7e {
			r = '?'
		}

		fmt.Fprintf(&out, "%04X", r)
	}

	return out.String()
}
