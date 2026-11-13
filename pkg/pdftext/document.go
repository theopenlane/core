package pdftext

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"iter"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/matrix"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/samber/lo"
	"golang.org/x/text/encoding/charmap"
)

// pdfcpu otherwise writes config.yml under $HOME/.config and panics when that is unwritable
// the stateless defaults match config.yml, revisit if we need custom fonts or trusted certs
func init() {
	model.ConfigPath = "disable"
}

// ContentType is the mime type of a pdf
const ContentType = "application/pdf"

// pdf name values the decoder compares against
const (
	// subtypeForm is the Subtype of a form xobject, a reusable content stream drawn with Do
	subtypeForm = "Form"
	// subtypeType0 is the Subtype of a composite font, whose character codes are two bytes
	subtypeType0 = "Type0"
	// encodingIdentityH is the composite font encoding where each two byte code is its cid
	encodingIdentityH = "Identity-H"
	// encodingMacRoman is the base encoding of fonts in pdfs written on macOS, whose curly quotes and
	// symbols sit above 0x7F where latin-1 would misread them
	encodingMacRoman = "MacRomanEncoding"
	// encodingWinAnsi is the Windows code page 1252 base encoding
	encodingWinAnsi = "WinAnsiEncoding"
)

// pdf dictionary keys the decoder reads
const (
	// keyXObject is the resources entry holding the page's named xobjects
	keyXObject = "XObject"
	// keyResources is the resources a page or form xobject draws with
	keyResources = "Resources"
	// keyFont is the resources entry holding the named fonts
	keyFont = "Font"
	// keyToUnicode is the CMap stream that maps a font's character codes to unicode
	keyToUnicode = "ToUnicode"
	// keyFirstChar is the first character code a simple font's Widths covers
	keyFirstChar = "FirstChar"
	// keyWidths is the glyph advances of a simple font, starting at FirstChar
	keyWidths = "Widths"
	// keyEncoding is how a font's character codes are encoded
	keyEncoding = "Encoding"
	// keyBaseEncoding is the named encoding an Encoding dictionary starts from
	keyBaseEncoding = "BaseEncoding"
	// keyDescendantFonts is the cid font a composite font draws its glyphs from
	keyDescendantFonts = "DescendantFonts"
	// keyDW is the default glyph advance of a cid font
	keyDW = "DW"
	// keyW is the per cid glyph advances of a cid font
	keyW = "W"
)

// IsContentType reports whether a mime or content type is a pdf, ignoring parameters and case
func IsContentType(mimeType string) bool {
	return strings.Contains(strings.ToLower(mimeType), ContentType)
}

// Document is an opened pdf with its page count and per-page text extraction
type Document struct {
	ctx *model.Context
}

// Open reads and validates the pdf once so page counting and text extraction share the parse
func Open(ctx context.Context, pdf []byte) (*Document, error) {
	conf := model.NewDefaultConfiguration()

	pdfCtx, err := api.ReadAndValidate(ctx, bytes.NewReader(pdf), conf)
	if err == nil {
		return &Document{ctx: pdfCtx}, nil
	}

	// some real world pdfs fail validation on a single malformed annotation while the page tree is intact
	pdfCtx, readErr := api.ReadContext(ctx, bytes.NewReader(pdf), conf)
	if readErr != nil {
		return nil, err
	}

	if pageErr := pdfCtx.EnsurePageCount(); pageErr != nil || pdfCtx.PageCount == 0 {
		return nil, err
	}

	return &Document{ctx: pdfCtx}, nil
}

// PageCount returns the number of pages in the document
func (d *Document) PageCount() int {
	return d.ctx.PageCount
}

// Text returns the text of every page, cleaned of layout padding. Pages that cannot be read are skipped
func (d *Document) Text(ctx context.Context) []byte {
	var text bytes.Buffer

	for index := range d.ctx.PageCount {
		// pdf pages are numbered from one
		pageNr := index + 1

		pageDict, _, attrs, err := d.ctx.PageDict(ctx, pageNr, false)
		if err != nil || pageDict == nil {
			continue
		}

		content, err := d.ctx.PageContent(pageDict, pageNr)
		if err != nil || len(content) == 0 {
			continue
		}

		var resources types.Dict
		if attrs != nil {
			resources = attrs.Resources
		}

		text.Write(d.newScope(resources, 0, map[string]bool{}).decode(content))
		text.WriteByte('\n')
	}

	return []byte(normalizeText(text.String()))
}

var (
	// dotLeaderPattern matches table of contents dot leaders: four or more dots, optionally separated
	// by single spaces, which would otherwise be read as runs of empty sentences
	dotLeaderPattern = regexp.MustCompile(`\.(?: ?\.){3,}`)
	// spaceRunPattern matches runs of horizontal whitespace, including non-breaking spaces, that pdf
	// layouts use for padding
	spaceRunPattern = regexp.MustCompile(`[ \t\x{00A0}]{2,}`)
)

// normalizeText removes dot leaders, collapses whitespace padding, trims each line and drops empty
// lines produced between pdf text objects
func normalizeText(s string) string {
	var kept []string

	for line := range strings.Lines(s) {
		line = dotLeaderPattern.ReplaceAllString(line, " ")
		line = spaceRunPattern.ReplaceAllString(line, " ")

		if line = strings.TrimSpace(line); line != "" {
			kept = append(kept, line)
		}
	}

	return strings.Join(kept, "\n")
}

// maxFormDepth bounds how deeply nested form xobjects are followed
const maxFormDepth = 8

// contentScope is the resources in effect for one content stream, a page or a form xobject
type contentScope struct {
	doc       *Document
	resources types.Dict
	fonts     map[string]*fontDecoder
	depth     int
	visited   map[string]bool
}

// newScope builds the scope for a content stream drawn with the given resources
func (d *Document) newScope(resources types.Dict, depth int, visited map[string]bool) *contentScope {
	return &contentScope{
		doc:       d,
		resources: resources,
		fonts:     d.resourceFonts(resources),
		depth:     depth,
		visited:   visited,
	}
}

// formText decodes the text of the named form xobject, watermarked pdfs often wrap every page this way
func (s *contentScope) formText(name string) []byte {
	if s.doc == nil || s.resources == nil || s.depth >= maxFormDepth {
		return nil
	}

	xobjects, found := s.resources.Find(keyXObject)
	if !found {
		return nil
	}

	xobjectDict, err := s.doc.ctx.DereferenceDict(xobjects)
	if err != nil || xobjectDict == nil {
		return nil
	}

	ref, found := xobjectDict.Find(name)
	if !found {
		return nil
	}

	key := ref.String()
	if s.visited[key] {
		return nil
	}

	stream, _, err := s.doc.ctx.DereferenceStreamDict(ref)
	if err != nil || stream == nil {
		return nil
	}

	if lo.FromPtr(stream.Subtype()) != subtypeForm {
		return nil
	}

	if err := stream.DecodeWithLimit(s.doc.ctx.DecodeLimit()); err != nil {
		return nil
	}

	formResources := s.resources

	if entry, found := stream.Find(keyResources); found {
		if dict, err := s.doc.ctx.DereferenceDict(entry); err == nil && dict != nil {
			formResources = dict
		}
	}

	s.visited[key] = true
	defer delete(s.visited, key)

	return s.doc.newScope(formResources, s.depth+1, s.visited).decode(stream.Content)
}

// fontDecoder maps the byte codes of one font to unicode text
type fontDecoder struct {
	// codeBytes is how many bytes make up one character code, 1 for simple fonts and 2 for composite fonts
	codeBytes int
	// unicode maps a character code to its text when the font carries a ToUnicode map
	unicode map[uint32]string
	// widths maps a character code to its advance in glyph space, nil when the font gives no widths
	widths map[uint32]float64
	// defaultWidth is the advance of a code missing from widths
	defaultWidth float64
	// charset decodes the single byte codes of a simple font that has no unicode mapping, nil reads them as latin-1
	charset *charmap.Charmap
}

// advance returns the width of a raw string operand in glyph space, false when the font gives no widths
func (f *fontDecoder) advance(raw []byte) (float64, bool) {
	if f == nil || f.widths == nil {
		return 0, false
	}

	var total float64

	for code := range f.codes(raw) {
		width, ok := f.widths[codeValue(code)]
		if !ok {
			width = f.defaultWidth
		}

		total += width
	}

	return total, true
}

// decode turns a raw string operand into text using the font's code width and unicode map
func (f *fontDecoder) decode(raw []byte) []byte {
	if f == nil {
		return raw
	}

	var out bytes.Buffer

	for code := range f.codes(raw) {
		if mapped, ok := f.unicode[codeValue(code)]; ok {
			out.WriteString(mapped)

			continue
		}

		// an unmapped single byte code reads through the font's encoding, or as latin-1 without one
		if f.codeBytes == 1 {
			out.WriteRune(f.decodeByte(code[0]))
		}
	}

	return out.Bytes()
}

// decodeByte maps a single byte code to its character through the font's encoding
func (f *fontDecoder) decodeByte(b byte) rune {
	if f.charset == nil {
		return rune(b)
	}

	return f.charset.DecodeByte(b)
}

// encodingName returns the named encoding of a simple font, either the Encoding name itself or the
// BaseEncoding of an Encoding dictionary, empty when the font names none
func (d *Document) encodingName(fontDict types.Dict) string {
	entry, found := fontDict.Find(keyEncoding)
	if !found {
		return ""
	}

	encoding, err := d.ctx.Dereference(entry)
	if err != nil {
		return ""
	}

	switch v := encoding.(type) {
	case types.Name:
		return v.Value()
	case types.Dict:
		return lo.FromPtr(v.NameEntry(keyBaseEncoding))
	default:
		return ""
	}
}

// charsetForEncoding maps a pdf base encoding name to its character set, nil for encodings read as latin-1
func charsetForEncoding(name string) *charmap.Charmap {
	switch name {
	case encodingMacRoman:
		return charmap.Macintosh
	case encodingWinAnsi:
		return charmap.Windows1252
	default:
		return nil
	}
}

// codes yields each whole character code in raw, dropping a trailing partial code
func (f *fontDecoder) codes(raw []byte) iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		if f.codeBytes < 1 {
			return
		}

		for code := range slices.Chunk(raw, f.codeBytes) {
			if len(code) < f.codeBytes || !yield(code) {
				return
			}
		}
	}
}

// compositeCodeBytes is the width of a character code in a composite font
const compositeCodeBytes = 2

// codeValue packs one or two bytes into a character code. An empty code, such as an empty hex
// string in a malformed CMap, is zero
func codeValue(b []byte) uint32 {
	switch len(b) {
	case 0:
		return 0
	case compositeCodeBytes:
		return uint32(binary.BigEndian.Uint16(b))
	default:
		return uint32(b[0])
	}
}

// resourceFonts builds a decoder for every font in the resources, keyed by resource name
func (d *Document) resourceFonts(resources types.Dict) map[string]*fontDecoder {
	fonts := map[string]*fontDecoder{}

	if resources == nil {
		return fonts
	}

	fontEntry, found := resources.Find(keyFont)
	if !found {
		return fonts
	}

	fontDicts, err := d.ctx.DereferenceDict(fontEntry)
	if err != nil {
		return fonts
	}

	for name, ref := range fontDicts {
		fontDict, err := d.ctx.DereferenceDict(ref)
		if err != nil || fontDict == nil {
			continue
		}

		fonts[name] = d.fontDecoder(fontDict)
	}

	return fonts
}

// fontDecoder reads a font dict into a decoder: composite fonts use two byte codes and need a
// ToUnicode map, simple fonts use one byte codes and read as latin text without one
func (d *Document) fontDecoder(fontDict types.Dict) *fontDecoder {
	decoder := &fontDecoder{codeBytes: 1}

	if lo.FromPtr(fontDict.Subtype()) == subtypeType0 {
		decoder.codeBytes = compositeCodeBytes
		decoder.widths, decoder.defaultWidth = d.compositeWidths(fontDict)
	} else {
		decoder.widths = d.simpleWidths(fontDict)
		decoder.charset = charsetForEncoding(d.encodingName(fontDict))
	}

	toUnicode, found := fontDict.Find(keyToUnicode)
	if !found {
		return decoder
	}

	stream, _, err := d.ctx.DereferenceStreamDict(toUnicode)
	if err != nil || stream == nil {
		return decoder
	}

	if err := stream.DecodeWithLimit(d.ctx.DecodeLimit()); err != nil {
		return decoder
	}

	// the codespace is unreliable, simple fonts often declare <0000> <FFFF> while mapping one byte codes
	decoder.unicode = parseToUnicode(stream.Content).unicode

	return decoder
}

const (
	// glyphSpaceUnits is how many glyph space units make one unit of text space
	glyphSpaceUnits = 1000
	// defaultCIDWidth is the advance of a cid font glyph when the font sets no DW
	defaultCIDWidth = 1000
	// widthRunArity is the number of W array items in a `first [w1 w2 ...]` entry
	widthRunArity = 2
	// widthRangeArity is the number of W array items in a `first last w` entry
	widthRangeArity = 3
)

// simpleWidths reads the FirstChar and Widths entries of a simple font, nil when either is missing
func (d *Document) simpleWidths(fontDict types.Dict) map[uint32]float64 {
	first := fontDict.IntEntry(keyFirstChar)

	entry, found := fontDict.Find(keyWidths)
	if first == nil || *first < 0 || !found {
		return nil
	}

	widthArray, err := d.ctx.DereferenceArray(entry)
	if err != nil || len(widthArray) == 0 {
		return nil
	}

	widths := make(map[uint32]float64, len(widthArray))

	for i, item := range widthArray {
		width, err := d.ctx.DereferenceNumber(item)
		if err != nil {
			return nil
		}

		widths[uint32(*first+i)] = width //nolint:gosec // G115: FirstChar is checked non-negative above
	}

	return widths
}

// compositeWidths reads the W and DW entries of a Type0 font's descendant. W is keyed by cid, so
// widths are only returned for Identity-H where a code is its cid
func (d *Document) compositeWidths(fontDict types.Dict) (map[uint32]float64, float64) {
	if lo.FromPtr(fontDict.NameEntry(keyEncoding)) != encodingIdentityH {
		return nil, 0
	}

	entry, found := fontDict.Find(keyDescendantFonts)
	if !found {
		return nil, 0
	}

	descendants, err := d.ctx.DereferenceArray(entry)
	if err != nil || len(descendants) == 0 {
		return nil, 0
	}

	descendant, err := d.ctx.DereferenceDict(descendants[0])
	if err != nil || descendant == nil {
		return nil, 0
	}

	defaultWidth := float64(defaultCIDWidth)

	if dw, found := descendant.Find(keyDW); found {
		if width, err := d.ctx.DereferenceNumber(dw); err == nil {
			defaultWidth = width
		}
	}

	widths := map[uint32]float64{}

	if w, found := descendant.Find(keyW); found {
		if widthArray, err := d.ctx.DereferenceArray(w); err == nil {
			d.cidWidths(widthArray, widths)
		}
	}

	return widths, defaultWidth
}

// cidWidths fills widths from a W array, whose entries are either `first [w1 w2 ...]` or `first last w`
func (d *Document) cidWidths(widthArray types.Array, widths map[uint32]float64) {
	for i := 0; i+1 < len(widthArray); {
		first, err := d.ctx.DereferenceNumber(widthArray[i])
		if err != nil || first < 0 || first > math.MaxUint16 {
			return
		}

		if run, err := d.ctx.DereferenceArray(widthArray[i+1]); err == nil {
			for offset, item := range run {
				if width, err := d.ctx.DereferenceNumber(item); err == nil {
					widths[uint32(first)+uint32(offset)] = width //nolint:gosec // G115: first is bounded to a uint16 above and run lengths fit a pdf array
				}
			}

			i += widthRunArity

			continue
		}

		if i+2 >= len(widthArray) {
			return
		}

		last, lastErr := d.ctx.DereferenceNumber(widthArray[i+1])
		width, widthErr := d.ctx.DereferenceNumber(widthArray[i+2])

		if lastErr != nil || widthErr != nil || last < first || last > math.MaxUint16 {
			return
		}

		for code := uint32(first); code <= uint32(last); code++ {
			widths[code] = width
		}

		i += widthRangeArity
	}
}

// toUnicodeMap is the parsed content of a ToUnicode CMap stream
type toUnicodeMap struct {
	codeBytes int
	unicode   map[uint32]string
}

// parseToUnicode reads the codespace, bfchar, and bfrange sections of a ToUnicode CMap
func parseToUnicode(cmap []byte) toUnicodeMap {
	result := toUnicodeMap{unicode: map[uint32]string{}}
	tokens := tokenize(cmap)

	for i := 0; i < len(tokens); i++ {
		if tokens[i].kind != tokenKeyword {
			continue
		}

		switch tokens[i].text {
		case "begincodespacerange":
			i = parseCodespace(tokens, i+1, &result)
		case "beginbfchar":
			i = parseBFChar(tokens, i+1, &result)
		case "beginbfrange":
			i = parseBFRange(tokens, i+1, &result)
		}
	}

	return result
}

// parseCodespace records the code width from the first codespace range and returns the index of its end keyword
func parseCodespace(tokens []token, start int, result *toUnicodeMap) int {
	for i, tok := range tokens[start:] {
		if tok.kind == tokenKeyword {
			return start + i
		}

		if tok.kind == tokenHex && result.codeBytes == 0 {
			result.codeBytes = len(tok.raw)
		}
	}

	return len(tokens)
}

// parseBFChar records single code mappings and returns the index of the end keyword
func parseBFChar(tokens []token, start int, result *toUnicodeMap) int {
	for i := start; i+1 < len(tokens); i += 2 {
		if tokens[i].kind == tokenKeyword {
			return i
		}

		// an empty source code maps nothing, so it is skipped rather than read as code zero
		if tokens[i].kind == tokenHex && tokens[i+1].kind == tokenHex && len(tokens[i].raw) > 0 {
			result.unicode[codeValue(tokens[i].raw)] = utf16Text(tokens[i+1].raw)
		}
	}

	return len(tokens)
}

// parseBFRange records code range mappings, either to a starting code or to an array of
// destinations, and returns the index of the end keyword
func parseBFRange(tokens []token, start int, result *toUnicodeMap) int {
	i := start

	for i+2 < len(tokens) {
		if tokens[i].kind == tokenKeyword {
			return i
		}

		if tokens[i].kind != tokenHex || tokens[i+1].kind != tokenHex {
			i++

			continue
		}

		lo, hi := codeValue(tokens[i].raw), codeValue(tokens[i+1].raw)
		if len(tokens[i].raw) == 0 || len(tokens[i+1].raw) == 0 || hi < lo || hi-lo > math.MaxUint16 {
			i += bfRangeArity

			continue
		}

		switch tokens[i+2].kind {
		case tokenHex:
			base := tokens[i+2].raw

			for code := lo; code <= hi; code++ {
				result.unicode[code] = utf16Text(offsetUTF16(base, uint16(code-lo))) //nolint:gosec // G115: the range guard above bounds code-lo to a uint16
			}

			i += bfRangeArity
		case tokenArrayOpen:
			j := i + bfRangeArity

			for code := lo; j < len(tokens) && tokens[j].kind != tokenArrayClose; j++ {
				if tokens[j].kind == tokenHex && code <= hi {
					result.unicode[code] = utf16Text(tokens[j].raw)
					code++
				}
			}

			i = j + 1
		default:
			i += bfRangeArity
		}
	}

	return len(tokens)
}

// bfRangeArity is the number of tokens in one bfrange entry: low code, high code, destination
const bfRangeArity = 3

// offsetUTF16 adds delta to the last code unit of a utf16 destination, the bfrange convention
func offsetUTF16(base []byte, delta uint16) []byte {
	if len(base) < utf16UnitBytes || delta == 0 {
		return base
	}

	shifted := slices.Clone(base)

	last := binary.BigEndian.Uint16(shifted[len(shifted)-utf16UnitBytes:]) + delta
	binary.BigEndian.PutUint16(shifted[len(shifted)-utf16UnitBytes:], last)

	return shifted
}

// utf16UnitBytes is the width of one utf16 code unit
const utf16UnitBytes = 2

// utf16Text decodes big endian utf16 bytes into a string, treating a lone byte as a single unit
func utf16Text(raw []byte) string {
	if len(raw) == 1 {
		return string(rune(raw[0]))
	}

	units := make([]uint16, 0, len(raw)/utf16UnitBytes)

	for unit := range slices.Chunk(raw, utf16UnitBytes) {
		if len(unit) == utf16UnitBytes {
			units = append(units, binary.BigEndian.Uint16(unit))
		}
	}

	return string(utf16.Decode(units))
}

// decode walks a content stream and decodes every text showing operator with the font selected
// by the most recent Tf, separating runs with spaces where they break and following Do into form xobjects
func (s *contentScope) decode(content []byte) []byte {
	var (
		text     bytes.Buffer
		operands []token
	)

	cursor := newTextCursor()

	for _, tok := range tokenize(content) {
		if tok.kind != tokenKeyword {
			operands = append(operands, tok)

			continue
		}

		switch tok.text {
		case "Tf":
			if len(operands) >= 2 && operands[len(operands)-2].kind == tokenName {
				cursor.font = s.fonts[operands[len(operands)-2].text]
				cursor.size = operands[len(operands)-1].number
			}
		case "Tj":
			if last := lastString(operands); last != nil {
				cursor.show(&text, last.raw)
			}
		case "'", "\"":
			writeBreak(&text)
			cursor.reset()
			cursor.placed = false

			if last := lastString(operands); last != nil {
				cursor.show(&text, last.raw)
			}
		case "TJ":
			writeTJ(&text, operands, cursor.font)
			cursor.measured = false
			cursor.placed = false
		case "Td", "TD":
			cursor.move(&text, operands)
		case "T*":
			writeBreak(&text)
			cursor.reset()
			cursor.placed = false
		case "q":
			cursor.saved = append(cursor.saved, cursor.ctm)
		case "Q":
			if n := len(cursor.saved); n > 0 {
				cursor.ctm, cursor.saved = cursor.saved[n-1], cursor.saved[:n-1]
			}
		case "cm":
			if m, ok := matrixOperands(operands); ok {
				cursor.ctm = m.Multiply(cursor.ctm)
			}
		case "BT":
			cursor.tm, cursor.tlm = matrix.IdentMatrix, matrix.IdentMatrix
			cursor.reset()
			cursor.pending = true
		case "Tm":
			if m, ok := matrixOperands(operands); ok {
				cursor.tm, cursor.tlm = m, m
			}

			cursor.reset()
			cursor.pending = true
		case "ET":
			cursor.reset()
			cursor.pending = true
		case "Do":
			if len(operands) >= 1 && operands[len(operands)-1].kind == tokenName {
				text.Write(s.formText(operands[len(operands)-1].text))
			}
		}

		operands = operands[:0]
	}

	return text.Bytes()
}

// wordGapThreshold is the TJ kerning adjustment, in thousandths of text space, treated as a word break
const wordGapThreshold = 180

// textCursor tracks the glyphs drawn since the last positioning operator so a Td that only steps
// past them, as renderers that place every glyph on its own do, is not read as a word break
type textCursor struct {
	font *fontDecoder
	size float64
	// drawn is the advance in text space of the glyphs shown since the last positioning operator
	drawn float64
	// measured is false when a glyph since the last positioning operator had no known width
	measured bool
	// ctm is the current transformation matrix, saved by q and restored by Q
	ctm matrix.Matrix
	// saved is the stack of transformation matrices q pushed
	saved []matrix.Matrix
	// tm and tlm are the text and text line matrices
	tm, tlm matrix.Matrix
	// pending is set when a Tm, BT, or ET may have started a new word, decided by where the next glyph lands
	pending bool
	// placed is true when end and height locate the last glyph drawn on the page
	placed bool
	// end is where the last glyph drawn ended on the page
	end types.Point
	// height is the font size of the last glyph drawn, on the page
	height float64
}

// newTextCursor starts a cursor with no transformation
func newTextCursor() textCursor {
	return textCursor{ctm: matrix.IdentMatrix, tm: matrix.IdentMatrix, tlm: matrix.IdentMatrix}
}

// show writes the decoded text of a string operand and adds its advance. A font without widths
// falls back to ending every run with a space
func (c *textCursor) show(text *bytes.Buffer, raw []byte) {
	if c.pending {
		if !c.continues(origin(c.tm.Multiply(c.ctm))) {
			writeBreak(text)
		}

		c.pending = false
	}

	text.Write(c.font.decode(raw))

	width, ok := c.font.advance(raw)
	if !ok {
		writeBreak(text)

		c.measured = false
		c.placed = false

		return
	}

	advance := width / glyphSpaceUnits * c.size
	c.drawn += advance
	c.tm = translation(advance, 0).Multiply(c.tm)

	render := c.tm.Multiply(c.ctm)
	c.end, c.height, c.placed = origin(render), math.Abs(c.size)*yScale(render), true
}

// continues reports whether a glyph starting at start carries on the word the last glyph drawn ended: on the
// same baseline, not jumping back, and no further on than a kerning gap
func (c *textCursor) continues(start types.Point) bool {
	if !c.placed || c.height == 0 {
		return false
	}

	gap := start.X - c.end.X

	return math.Abs(start.Y-c.end.Y) <= c.height/2 && gap >= -c.height/4 && gap <= wordGapThreshold*c.height/glyphSpaceUnits
}

// move handles Td and TD, writing a space when the move starts a new line or leaves a gap wider
// than kerning past the glyphs already drawn
func (c *textCursor) move(text *bytes.Buffer, operands []token) {
	tx, ty, ok := moveOffsets(operands)
	gap := math.Abs(tx - c.drawn)

	if !ok || !c.measured || ty != 0 || gap > wordGapThreshold*math.Abs(c.size)/glyphSpaceUnits {
		writeBreak(text)
	}

	if ok {
		c.tlm = translation(tx, ty).Multiply(c.tlm)
		c.tm = c.tlm
	}

	c.reset()
}

// translation moves positions by tx and ty
func translation(tx, ty float64) matrix.Matrix {
	return matrix.Matrix{{1, 0, 0}, {0, 1, 0}, {tx, ty, 1}}
}

// origin is where a matrix maps the start of the current glyph
func origin(m matrix.Matrix) types.Point {
	return m.Transform(types.Point{})
}

// yScale is how much the matrix stretches a unit of height
func yScale(m matrix.Matrix) float64 {
	return math.Hypot(m[1][0], m[1][1])
}

// matrixOperands returns the matrix [a b c d e f] given by the six numbers of a cm or Tm
func matrixOperands(operands []token) (matrix.Matrix, bool) {
	const size = 6

	if len(operands) < size {
		return matrix.Matrix{}, false
	}

	var n [size]float64

	for i, operand := range operands[len(operands)-size:] {
		if operand.kind != tokenNumber {
			return matrix.Matrix{}, false
		}

		n[i] = operand.number
	}

	return matrix.Matrix{{n[0], n[1], 0}, {n[2], n[3], 0}, {n[4], n[5], 1}}, true
}

// reset starts measuring from a fresh position
func (c *textCursor) reset() {
	c.drawn = 0
	c.measured = true
}

// moveOffsets returns the tx and ty operands of a Td or TD
func moveOffsets(operands []token) (float64, float64, bool) {
	if len(operands) < 2 { //nolint:mnd
		return 0, 0, false
	}

	tx, ty := operands[len(operands)-2], operands[len(operands)-1]
	if tx.kind != tokenNumber || ty.kind != tokenNumber {
		return 0, 0, false
	}

	return tx.number, ty.number, true
}

// writeTJ decodes each string in a TJ array, inserting a space where the kerning gap is wide enough to be a word break
func writeTJ(text *bytes.Buffer, operands []token, current *fontDecoder) {
	for _, operand := range operands {
		switch operand.kind {
		case tokenString, tokenHex:
			text.Write(current.decode(operand.raw))
		case tokenNumber:
			if operand.number < -wordGapThreshold {
				writeBreak(text)
			}
		}
	}

	writeBreak(text)
}

// writeBreak writes a word break unless the text already ends in one
func writeBreak(text *bytes.Buffer) {
	if bytes.HasSuffix(text.Bytes(), []byte{' '}) {
		return
	}

	text.WriteByte(' ')
}

// lastString returns the last string operand, which is what Tj and its quote variants show
func lastString(operands []token) *token {
	for i, operand := range slices.Backward(operands) {
		if operand.kind == tokenString || operand.kind == tokenHex {
			return &operands[i]
		}
	}

	return nil
}

// tokenKind classifies a token from a content stream or CMap
type tokenKind int

const (
	tokenKeyword tokenKind = iota
	tokenName
	tokenNumber
	tokenString
	tokenHex
	tokenArrayOpen
	tokenArrayClose
	tokenDictOpen
	tokenDictClose
)

// token is one lexical token. raw holds decoded string bytes and text holds names and keywords
type token struct {
	kind   tokenKind
	text   string
	raw    []byte
	number float64
}

// errUnterminated marks a string that ran past the end of the input
var errUnterminated = errors.New("pdftext: unterminated pdf string")

// tokenize splits pdf syntax into tokens, decoding literal and hex strings as it goes
func tokenize(input []byte) []token {
	var tokens []token

	for i := 0; i < len(input); {
		c := input[i]

		switch {
		case isWhitespace(c):
			i++
		case c == '%':
			for i < len(input) && input[i] != '\n' && input[i] != '\r' {
				i++
			}
		case c == '(':
			raw, next, err := readLiteralString(input, i+1)
			if err != nil {
				return tokens
			}

			tokens = append(tokens, token{kind: tokenString, raw: raw})
			i = next
		case c == '<' && i+1 < len(input) && input[i+1] == '<':
			tokens = append(tokens, token{kind: tokenDictOpen})
			i += 2
		case c == '>' && i+1 < len(input) && input[i+1] == '>':
			tokens = append(tokens, token{kind: tokenDictClose})
			i += 2
		case c == '<':
			end := bytes.IndexByte(input[i:], '>')
			if end < 0 {
				return tokens
			}

			tokens = append(tokens, token{kind: tokenHex, raw: decodeHex(input[i+1 : i+end])})
			i += end + 1
		case c == '[':
			tokens = append(tokens, token{kind: tokenArrayOpen})
			i++
		case c == ']':
			tokens = append(tokens, token{kind: tokenArrayClose})
			i++
		case c == '{' || c == '}' || c == ')' || c == '>':
			i++
		case c == '/':
			end := wordEnd(input, i+1)

			tokens = append(tokens, token{kind: tokenName, text: string(input[i+1 : end])})
			i = end
		default:
			end := wordEnd(input, i)

			if end == i {
				i++

				continue
			}

			word := string(input[i:end])
			if number, err := strconv.ParseFloat(word, 64); err == nil {
				tokens = append(tokens, token{kind: tokenNumber, number: number})
			} else {
				tokens = append(tokens, token{kind: tokenKeyword, text: word})
			}

			i = end
		}
	}

	return tokens
}

// wordEnd returns the index of the first delimiter at or after start, or the end of the input
func wordEnd(input []byte, start int) int {
	for end := start; end < len(input); end++ {
		if isDelimiter(input[end]) {
			return end
		}
	}

	return len(input)
}

// readLiteralString decodes a parenthesised string starting after the opening paren, handling
// nesting and backslash escapes, and returns the index after the closing paren
func readLiteralString(input []byte, start int) ([]byte, int, error) {
	var out bytes.Buffer

	depth := 1

	for i := start; i < len(input); i++ {
		c := input[i]

		switch c {
		case '\\':
			if i+1 >= len(input) {
				return nil, 0, errUnterminated
			}

			i++

			switch esc := input[i]; esc {
			case 'n':
				out.WriteByte('\n')
			case 'r', 't', 'f', 'b':
				out.WriteByte(' ')
			case '\n', '\r':
			default:
				if esc >= '0' && esc <= '7' {
					value, consumed := octalEscape(input[i:])
					out.WriteByte(value)

					i += consumed - 1
				} else {
					out.WriteByte(esc)
				}
			}
		case '(':
			depth++

			out.WriteByte(c)
		case ')':
			depth--
			if depth == 0 {
				return out.Bytes(), i + 1, nil
			}

			out.WriteByte(c)
		default:
			out.WriteByte(c)
		}
	}

	return nil, 0, errUnterminated
}

const (
	// octalBase is the radix of a pdf octal escape
	octalBase = 8
	// maxOctalDigits is the most digits a pdf octal escape carries
	maxOctalDigits = 3
)

// octalEscape reads up to three octal digits and returns the byte and how many digits were used
func octalEscape(input []byte) (byte, int) {
	value, consumed := 0, 0

	for consumed < maxOctalDigits && consumed < len(input) && input[consumed] >= '0' && input[consumed] <= '7' {
		value = value*octalBase + int(input[consumed]-'0')
		consumed++
	}

	return byte(value), consumed
}

// isWhitespace reports whether c separates tokens
func isWhitespace(c byte) bool {
	return strings.IndexByte(whitespaceChars, c) >= 0
}

// isDelimiter reports whether c ends a bare word
func isDelimiter(c byte) bool {
	return isWhitespace(c) || strings.IndexByte(delimiterChars, c) >= 0
}

const (
	// whitespaceChars are the bytes pdf syntax treats as whitespace
	whitespaceChars = " \t\n\r\f\x00"
	// delimiterChars are the bytes besides whitespace that end a bare word
	delimiterChars = "()<>[]{}/%"
)

// hexDigitsPerByte is how many hex digits encode one byte
const hexDigitsPerByte = 2

// decodeHex turns a hex pdf string into bytes, ignoring anything that is not hex and padding an
// odd final digit with zero, as the pdf spec requires
func decodeHex(hexText []byte) []byte {
	cleaned := bytes.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			return r
		}

		return -1
	}, hexText)

	if len(cleaned)%hexDigitsPerByte == 1 {
		cleaned = append(cleaned, '0')
	}

	decoded := make([]byte, hex.DecodedLen(len(cleaned)))

	// cleaned holds only hex digits in pairs, so decoding cannot fail
	n, _ := hex.Decode(decoded, cleaned)

	return decoded[:n]
}
