package pdftext

import (
	"fmt"
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/pkg/pdftext/pdftest"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name      string
		pdf       []byte
		wantErr   bool
		wantPages int
		wantText  []string
	}{
		{
			name:      "simple font",
			pdf:       pdftest.Build([][]string{{"Hello world", "Second line"}, {"Page two"}}, false),
			wantPages: 2,
			wantText:  []string{"Hello world", "Second line", "Page two"},
		},
		{
			name:      "composite font",
			pdf:       pdftest.Build([][]string{{"Composite font text", "SOC 2 Type II"}}, true),
			wantPages: 1,
			wantText:  []string{"Composite font text", "SOC 2 Type II"},
		},
		{
			name:     "positioned glyphs",
			pdf:      pdftest.BuildGlyphs([][]string{{"SOC 2 Type 1 Report", "Independent Service Auditor's Report"}}),
			wantText: []string{"SOC 2 Type 1 Report", "Independent Service Auditor's Report"},
		},
		{
			name:    "not a pdf",
			pdf:     []byte("not a pdf"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document, err := Open(t.Context(), tt.pdf)
			if tt.wantErr {
				assert.Check(t, err != nil)
				return
			}

			assert.NilError(t, err)

			if tt.wantPages > 0 {
				assert.Check(t, document.PageCount() == tt.wantPages)
			}

			text := string(document.Text(t.Context()))
			for _, want := range tt.wantText {
				assert.Check(t, strings.Contains(text, want), "missing %q in %q", want, text)
			}
		})
	}
}

func TestReadLiteralString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
		wantNext int
		wantErr  error
	}{
		{
			name:     "escapes and nested parentheses",
			input:    `a\(b\)c \\ nested (inner) end) tail`,
			want:     `a(b)c \ nested (inner) end`,
			wantNext: len(`a\(b\)c \\ nested (inner) end)`),
		},
		{
			name:     "octal escapes",
			input:    `\101\102)`,
			want:     "AB",
			wantNext: len(`\101\102)`),
		},
		{
			name:    "unterminated",
			input:   `never closed`,
			wantErr: errUnterminated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, next, err := readLiteralString([]byte(tt.input), 0)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NilError(t, err)
			assert.Check(t, string(decoded) == tt.want)
			assert.Check(t, next == tt.wantNext)
		})
	}
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

func TestParseToUnicode(t *testing.T) {
	tests := []struct {
		name          string
		cmap          string
		wantCodeBytes int
		want          map[uint32]string
		wantUnmapped  []uint32
	}{
		{
			name: "bfchar",
			cmap: `
1 begincodespacerange <0000> <FFFF> endcodespacerange
2 beginbfchar
<0003> <0041>
<0004> <00E9>
endbfchar`,
			wantCodeBytes: 2,
			want:          map[uint32]string{0x0003: "A", 0x0004: "é"},
		},
		{
			name: "bfrange with offset and array forms",
			cmap: `
1 begincodespacerange <00> <FF> endcodespacerange
2 beginbfrange
<10> <12> <0061>
<20> <21> [<0058> <0059>]
endbfrange`,
			wantCodeBytes: 1,
			want:          map[uint32]string{0x10: "a", 0x11: "b", 0x12: "c", 0x20: "X", 0x21: "Y"},
		},
		{
			name: "bfchar skips an empty source code",
			cmap: `
1 begincodespacerange <00> <FF> endcodespacerange
2 beginbfchar
<> <0041>
<03> <0042>
endbfchar`,
			wantCodeBytes: 1,
			want:          map[uint32]string{0x03: "B"},
			wantUnmapped:  []uint32{0x00},
		},
		{
			name: "bfrange skips entries with an empty low or high code",
			cmap: `
1 begincodespacerange <00> <FF> endcodespacerange
3 beginbfrange
<> <12> <0061>
<20> <> <0058>
<30> <31> <0041>
endbfrange`,
			wantCodeBytes: 1,
			want:          map[uint32]string{0x30: "A", 0x31: "B"},
			wantUnmapped:  []uint32{0x00, 0x10, 0x12, 0x20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmap := parseToUnicode([]byte(tt.cmap))

			assert.Check(t, cmap.codeBytes == tt.wantCodeBytes)

			for code, want := range tt.want {
				assert.Check(t, cmap.unicode[code] == want, "code %#x: got %q, want %q", code, cmap.unicode[code], want)
			}

			for _, code := range tt.wantUnmapped {
				_, mapped := cmap.unicode[code]
				assert.Check(t, !mapped, "code %#x should not be mapped", code)
			}
		})
	}
}

func TestCodeValue(t *testing.T) {
	tests := []struct {
		name string
		code []byte
		want uint32
	}{
		{name: "empty", code: []byte{}, want: 0},
		{name: "one byte", code: []byte{0x41}, want: 0x41},
		{name: "two bytes", code: []byte{0x01, 0x02}, want: 0x0102},
		{name: "longer codes use the first byte", code: []byte{0x07, 0x08, 0x09}, want: 0x07},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Check(t, codeValue(tt.code) == tt.want)
		})
	}
}

// quartzGlyph draws one glyph in its own text block, the way macOS Quartz exports text
func quartzGlyph(y, x float64, glyph string) string {
	return fmt.Sprintf("q 0.75 0 0 -0.75 0 %g cm BT 48 0 0 -48 %g 48 Tm /F1 1 Tf (%s) Tj ET Q ", y, x, glyph)
}

func TestDecodeContent(t *testing.T) {
	simpleFont := func() map[string]*fontDecoder {
		return map[string]*fontDecoder{"F1": {codeBytes: 1}}
	}

	widthFont := func() map[string]*fontDecoder {
		return map[string]*fontDecoder{"F1": {codeBytes: 1, widths: map[uint32]float64{}, defaultWidth: 500}}
	}

	tests := []struct {
		name        string
		fonts       map[string]*fontDecoder
		content     string
		contains    []string
		notContains []string
	}{
		{
			name:     "TJ kerning gaps become word breaks",
			fonts:    simpleFont(),
			content:  `BT /F1 10 Tf [(SO) -20 (C 2) -400 (report)] TJ ET`,
			contains: []string{"SOC 2 report"},
		},
		{
			name:     "small Td step is not a word break",
			fonts:    widthFont(),
			content:  `BT /F1 10 Tf (S) Tj 5 0 Td (O) Tj 5.2 0 Td (C) Tj 8 0 Td (2) Tj 5 -12 Td (x) Tj ET`,
			contains: []string{"SOC 2 x"},
		},
		{
			name:        "glyph per block joins words",
			fonts:       widthFont(),
			content:     quartzGlyph(700, 48, "S") + quartzGlyph(700, 72, "O") + quartzGlyph(700, 96, "C") + quartzGlyph(700, 140, "2"),
			contains:    []string{"SOC 2"},
			notContains: []string{"S O"},
		},
		{
			name:     "glyph per block breaks lines",
			fonts:    widthFont(),
			content:  quartzGlyph(700, 48, "a") + quartzGlyph(700, 72, "b") + quartzGlyph(640, 48, "c"),
			contains: []string{"ab c"},
		},
		{
			name:     "Td without widths breaks",
			fonts:    simpleFont(),
			content:  `BT /F1 10 Tf (one) Tj 20 0 Td (two) Tj ET`,
			contains: []string{"one two"},
		},
		{
			name:     "composite font",
			fonts:    map[string]*fontDecoder{"F2": {codeBytes: 2, unicode: map[uint32]string{0x0001: "H", 0x0002: "i"}}},
			content:  `BT /F2 10 Tf <00010002> Tj ET`,
			contains: []string{"Hi"},
		},
		{
			name:     "unknown font falls back to raw bytes",
			fonts:    map[string]*fontDecoder{},
			content:  `BT /Missing 10 Tf (raw) Tj ET`,
			contains: []string{"raw"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := string((&contentScope{fonts: tt.fonts}).decode([]byte(tt.content)))

			for _, want := range tt.contains {
				assert.Check(t, strings.Contains(text, want), "missing %q in %q", want, text)
			}

			for _, unwanted := range tt.notContains {
				assert.Check(t, !strings.Contains(text, unwanted), "unexpected %q in %q", unwanted, text)
			}
		})
	}
}

func TestFontDecoderDecode(t *testing.T) {
	wideCodespace := parseToUnicode([]byte(`
1 begincodespacerange <0000> <FFFF> endcodespacerange
2 beginbfchar
<41> <0041>
<42> <0042>
endbfchar`))

	tests := []struct {
		name    string
		decoder *fontDecoder
		input   []byte
		want    string
	}{
		{
			name:    "simple font with a two byte codespace",
			decoder: &fontDecoder{codeBytes: 1, unicode: wideCodespace.unicode},
			input:   []byte("AB"),
			want:    "AB",
		},
		{
			name:    "simple font falls back for unmapped codes",
			decoder: &fontDecoder{codeBytes: 1, unicode: map[uint32]string{0x41: "Z"}},
			input:   []byte("AB"),
			want:    "ZB",
		},
		{
			name:    "composite font drops unmapped codes",
			decoder: &fontDecoder{codeBytes: 2, unicode: map[uint32]string{0x0001: "H"}},
			input:   []byte{0x00, 0x01, 0x00, 0x09},
			want:    "H",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Check(t, string(tt.decoder.decode(tt.input)) == tt.want)
		})
	}
}
