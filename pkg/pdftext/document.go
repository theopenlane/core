package pdftext

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"strconv"
	"unicode/utf16"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// ContentType is the mime type of a pdf
const ContentType = "application/pdf"

// Document is an opened pdf with its page count and per-page text extraction
type Document struct {
	ctx *model.Context
}

// Open reads and validates the pdf once so page counting and text extraction share the parse
func Open(pdf []byte) (*Document, error) {
	conf := model.NewDefaultConfiguration()

	ctx, err := api.ReadAndValidate(bytes.NewReader(pdf), conf)
	if err == nil {
		return &Document{ctx: ctx}, nil
	}

	// some real world pdfs fail validation on a single malformed annotation while the page tree is intact
	ctx, readErr := api.ReadContext(bytes.NewReader(pdf), conf)
	if readErr != nil {
		return nil, err
	}

	if pageErr := ctx.EnsurePageCount(); pageErr != nil || ctx.PageCount == 0 {
		return nil, err
	}

	return &Document{ctx: ctx}, nil
}

// PageCount returns the number of pages in the document
func (d *Document) PageCount() int {
	return d.ctx.PageCount
}

// Text returns the text of every page; pages that cannot be read are skipped
func (d *Document) Text() []byte {
	var text bytes.Buffer

	for index := range d.ctx.PageCount {
		// pdf pages are numbered from one
		pageNr := index + 1

		pageDict, _, attrs, err := d.ctx.PageDict(pageNr, false)
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

	return text.Bytes()
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
	return &contentScope{doc: d, resources: resources, fonts: d.resourceFonts(resources), depth: depth, visited: visited}
}

// formText decodes the text of the named form xobject, watermarked pdfs often wrap every page this way
func (s *contentScope) formText(name string) []byte {
	if s.doc == nil || s.resources == nil || s.depth >= maxFormDepth {
		return nil
	}

	xobjects, found := s.resources.Find("XObject")
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

	if subtype := stream.NameEntry("Subtype"); subtype == nil || *subtype != "Form" {
		return nil
	}

	if err := stream.Decode(); err != nil {
		return nil
	}

	formResources := s.resources
	if entry, found := stream.Find("Resources"); found {
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
}

// decode turns a raw string operand into text using the font's code width and unicode map
func (f *fontDecoder) decode(raw []byte) []byte {
	if f == nil {
		return raw
	}

	var out bytes.Buffer

	for i := 0; i+f.codeBytes <= len(raw); i += f.codeBytes {
		code := codeValue(raw[i : i+f.codeBytes])

		if mapped, ok := f.unicode[code]; ok {
			out.WriteString(mapped)

			continue
		}

		// an unmapped single byte code reads as latin text
		if f.codeBytes == 1 {
			out.WriteRune(rune(raw[i]))
		}
	}

	return out.Bytes()
}

// compositeCodeBytes is the width of a character code in a composite font
const compositeCodeBytes = 2

// codeValue packs one or two bytes into a character code
func codeValue(b []byte) uint32 {
	if len(b) == compositeCodeBytes {
		return uint32(binary.BigEndian.Uint16(b))
	}

	return uint32(b[0])
}

// resourceFonts builds a decoder for every font in the resources, keyed by resource name
func (d *Document) resourceFonts(resources types.Dict) map[string]*fontDecoder {
	fonts := map[string]*fontDecoder{}

	if resources == nil {
		return fonts
	}

	fontEntry, found := resources.Find("Font")
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

	if subtype := fontDict.NameEntry("Subtype"); subtype != nil && *subtype == "Type0" {
		decoder.codeBytes = 2
	}

	toUnicode, found := fontDict.Find("ToUnicode")
	if !found {
		return decoder
	}

	stream, _, err := d.ctx.DereferenceStreamDict(toUnicode)
	if err != nil || stream == nil {
		return decoder
	}

	if err := stream.Decode(); err != nil {
		return decoder
	}

	// the codespace is unreliable, simple fonts often declare <0000> <FFFF> while mapping one byte codes
	decoder.unicode = parseToUnicode(stream.Content).unicode

	return decoder
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
	for i := start; i < len(tokens); i++ {
		if tokens[i].kind == tokenKeyword {
			return i
		}

		if tokens[i].kind == tokenHex && result.codeBytes == 0 {
			result.codeBytes = len(tokens[i].raw)
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

		if tokens[i].kind == tokenHex && tokens[i+1].kind == tokenHex {
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
		if hi < lo || hi-lo > math.MaxUint16 {
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

	shifted := make([]byte, len(base))
	copy(shifted, base)

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

	for i := 0; i+1 < len(raw); i += utf16UnitBytes {
		units = append(units, binary.BigEndian.Uint16(raw[i:]))
	}

	return string(utf16.Decode(units))
}

// decode walks a content stream and decodes every text showing operator with the font selected
// by the most recent Tf, separating runs with spaces and following Do into form xobjects
func (s *contentScope) decode(content []byte) []byte {
	var (
		text     bytes.Buffer
		operands []token
		current  *fontDecoder
	)

	for _, tok := range tokenize(content) {
		if tok.kind != tokenKeyword {
			operands = append(operands, tok)

			continue
		}

		switch tok.text {
		case "Tf":
			if len(operands) >= 2 && operands[len(operands)-2].kind == tokenName {
				current = s.fonts[operands[len(operands)-2].text]
			}
		case "Tj", "'", "\"":
			if last := lastString(operands); last != nil {
				text.Write(current.decode(last.raw))
				text.WriteByte(' ')
			}
		case "TJ":
			writeTJ(&text, operands, current)
		case "T*", "Td", "TD", "Tm", "ET":
			text.WriteByte(' ')
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

// writeTJ decodes each string in a TJ array, inserting a space where the kerning gap is wide enough to be a word break
func writeTJ(text *bytes.Buffer, operands []token, current *fontDecoder) {
	for _, operand := range operands {
		switch operand.kind {
		case tokenString, tokenHex:
			text.Write(current.decode(operand.raw))
		case tokenNumber:
			if operand.number < -wordGapThreshold {
				text.WriteByte(' ')
			}
		}
	}

	text.WriteByte(' ')
}

// lastString returns the last string operand, which is what Tj and its quote variants show
func lastString(operands []token) *token {
	for i := len(operands) - 1; i >= 0; i-- {
		if operands[i].kind == tokenString || operands[i].kind == tokenHex {
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

// token is one lexical token; raw holds decoded string bytes, text holds names and keywords
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
			end := i + 1
			for end < len(input) && !isDelimiter(input[end]) {
				end++
			}

			tokens = append(tokens, token{kind: tokenName, text: string(input[i+1 : end])})
			i = end
		default:
			end := i
			for end < len(input) && !isDelimiter(input[end]) {
				end++
			}

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
	switch c {
	case ' ', '\t', '\n', '\r', '\f', 0:
		return true
	default:
		return false
	}
}

// isDelimiter reports whether c ends a bare word
func isDelimiter(c byte) bool {
	if isWhitespace(c) {
		return true
	}

	switch c {
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}

// decodeHex turns a hex pdf string into bytes, ignoring anything that is not hex
func decodeHex(hexText []byte) []byte {
	cleaned := make([]byte, 0, len(hexText))

	for _, c := range hexText {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f', c >= 'A' && c <= 'F':
			cleaned = append(cleaned, c)
		}
	}

	if len(cleaned)%hexDigitsPerByte == 1 {
		cleaned = append(cleaned, '0')
	}

	decoded := make([]byte, len(cleaned)/hexDigitsPerByte)

	for i := range decoded {
		decoded[i] = hexNibble(cleaned[hexDigitsPerByte*i])<<nibbleBits | hexNibble(cleaned[hexDigitsPerByte*i+1])
	}

	return decoded
}

const (
	// hexDigitsPerByte is how many hex digits encode one byte
	hexDigitsPerByte = 2
	// nibbleBits is the width of one hex digit in bits
	nibbleBits = 4
	// hexLetterOffset is the value of the first hex letter digit
	hexLetterOffset = 10
)

// hexNibble converts one hex digit to its value
func hexNibble(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + hexLetterOffset
	default:
		return c - 'A' + hexLetterOffset
	}
}
