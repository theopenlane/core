package pdftext

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/pdftext/pdftest"
)

func TestOpenSimpleFont(t *testing.T) {
	document, err := Open(pdftest.Build([][]string{{"Hello world", "Second line"}, {"Page two"}}, false))

	assert.NilError(t, err)
	assert.Check(t, document.PageCount() == 2)

	text := string(document.Text())

	assert.Check(t, strings.Contains(text, "Hello world"))
	assert.Check(t, strings.Contains(text, "Second line"))
	assert.Check(t, strings.Contains(text, "Page two"))
}

func TestOpenCompositeFont(t *testing.T) {
	document, err := Open(pdftest.Build([][]string{{"Composite font text", "SOC 2 Type II"}}, true))

	assert.NilError(t, err)
	assert.Check(t, document.PageCount() == 1)

	text := string(document.Text())

	assert.Check(t, strings.Contains(text, "Composite font text"))
	assert.Check(t, strings.Contains(text, "SOC 2 Type II"))
}

func TestOpenRejectsNonPDF(t *testing.T) {
	_, err := Open([]byte("not a pdf"))

	assert.Check(t, err != nil)
}

func TestLiteralStringEscapes(t *testing.T) {
	decoded, next, err := readLiteralString([]byte(`a\(b\)c \\ nested (inner) end) tail`), 0)

	assert.NilError(t, err)
	assert.Check(t, string(decoded) == `a(b)c \ nested (inner) end`)
	assert.Check(t, next == len(`a\(b\)c \\ nested (inner) end)`))
}

func TestLiteralStringOctalEscape(t *testing.T) {
	decoded, _, err := readLiteralString([]byte(`\101\102)`), 0)

	assert.NilError(t, err)
	assert.Check(t, string(decoded) == "AB")
}

func TestLiteralStringUnterminated(t *testing.T) {
	_, _, err := readLiteralString([]byte(`never closed`), 0)

	assert.ErrorIs(t, err, errUnterminated)
}

func TestTokenizeKinds(t *testing.T) {
	tokens := tokenize([]byte(`/F1 12 Tf [(A) -250 (B)] TJ <4142> Tj << /K 1 >> % comment`))

	kinds := make([]tokenKind, 0, len(tokens))
	for _, tok := range tokens {
		kinds = append(kinds, tok.kind)
	}

	assert.DeepEqual(t, kinds, []tokenKind{
		tokenName, tokenNumber, tokenKeyword,
		tokenArrayOpen, tokenString, tokenNumber, tokenString, tokenArrayClose, tokenKeyword,
		tokenHex, tokenKeyword,
		tokenDictOpen, tokenName, tokenNumber, tokenDictClose,
	})
	assert.Check(t, string(tokens[9].raw) == "AB")
}

func TestParseToUnicodeBFChar(t *testing.T) {
	cmap := parseToUnicode([]byte(`
1 begincodespacerange <0000> <FFFF> endcodespacerange
2 beginbfchar
<0003> <0041>
<0004> <00E9>
endbfchar`))

	assert.Check(t, cmap.codeBytes == 2)
	assert.Check(t, cmap.unicode[0x0003] == "A")
	assert.Check(t, cmap.unicode[0x0004] == "é")
}

func TestParseToUnicodeBFRange(t *testing.T) {
	cmap := parseToUnicode([]byte(`
1 begincodespacerange <00> <FF> endcodespacerange
2 beginbfrange
<10> <12> <0061>
<20> <21> [<0058> <0059>]
endbfrange`))

	assert.Check(t, cmap.codeBytes == 1)
	assert.Check(t, cmap.unicode[0x10] == "a")
	assert.Check(t, cmap.unicode[0x11] == "b")
	assert.Check(t, cmap.unicode[0x12] == "c")
	assert.Check(t, cmap.unicode[0x20] == "X")
	assert.Check(t, cmap.unicode[0x21] == "Y")
}

func TestDecodeContentTextWordGaps(t *testing.T) {
	fonts := map[string]*fontDecoder{"F1": {codeBytes: 1}}

	text := (&contentScope{fonts: fonts}).decode([]byte(`BT /F1 10 Tf [(SO) -20 (C 2) -400 (report)] TJ ET`))

	assert.Check(t, strings.Contains(string(text), "SOC 2 report"))
}

func TestDecodeContentTextCompositeFont(t *testing.T) {
	fonts := map[string]*fontDecoder{"F2": {codeBytes: 2, unicode: map[uint32]string{0x0001: "H", 0x0002: "i"}}}

	text := (&contentScope{fonts: fonts}).decode([]byte(`BT /F2 10 Tf <00010002> Tj ET`))

	assert.Check(t, strings.Contains(string(text), "Hi"))
}

func TestDecodeContentTextUnknownFont(t *testing.T) {
	text := (&contentScope{fonts: map[string]*fontDecoder{}}).decode([]byte(`BT /Missing 10 Tf (raw) Tj ET`))

	assert.Check(t, strings.Contains(string(text), "raw"))
}

func TestDecodeSimpleFontWithWideCodespace(t *testing.T) {
	cmap := parseToUnicode([]byte(`
1 begincodespacerange <0000> <FFFF> endcodespacerange
2 beginbfchar
<41> <0041>
<42> <0042>
endbfchar`))

	decoder := &fontDecoder{codeBytes: 1, unicode: cmap.unicode}

	assert.Check(t, string(decoder.decode([]byte("AB"))) == "AB")
}

func TestDecodeSimpleFontFallsBackForUnmappedCodes(t *testing.T) {
	decoder := &fontDecoder{codeBytes: 1, unicode: map[uint32]string{0x41: "Z"}}

	assert.Check(t, string(decoder.decode([]byte("AB"))) == "ZB")
}

func TestDecodeCompositeFontDropsUnmappedCodes(t *testing.T) {
	decoder := &fontDecoder{codeBytes: 2, unicode: map[uint32]string{0x0001: "H"}}

	assert.Check(t, string(decoder.decode([]byte{0x00, 0x01, 0x00, 0x09})) == "H")
}
