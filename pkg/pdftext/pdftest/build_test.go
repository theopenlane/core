package pdftest

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/pkg/pdftext"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name      string
		pdf       []byte
		wantPages int
		wantText  []string
	}{
		{
			name:      "simple font",
			pdf:       Build([][]string{{"Hello world", "Second line"}, {"Page two"}}, false),
			wantPages: 2,
			wantText:  []string{"Hello world", "Second line", "Page two"},
		},
		{
			name:      "simple font escapes string delimiters",
			pdf:       Build([][]string{{`Backslash \ and (parens)`}}, false),
			wantPages: 1,
			wantText:  []string{`Backslash \ and (parens)`},
		},
		{
			name:      "composite font",
			pdf:       Build([][]string{{"Composite font text", "SOC 2 Type II"}}, true),
			wantPages: 1,
			wantText:  []string{"Composite font text", "SOC 2 Type II"},
		},
		{
			name:      "positioned glyphs",
			pdf:       BuildGlyphs([][]string{{"SOC 2 Type 1 Report"}, {"Independent Service Auditor's Report"}}),
			wantPages: 2,
			wantText:  []string{"SOC 2 Type 1 Report", "Independent Service Auditor's Report"},
		},
		{
			name:      "repeated pages are numbered",
			pdf:       Repeat([]string{"Acceptable Use Policy"}, 3),
			wantPages: 3,
			wantText:  []string{"Acceptable Use Policy", "Page 1 of 3", "Page 2 of 3", "Page 3 of 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := model.NewDefaultConfiguration()

			// the hand written xref table and object offsets must produce a pdf that pdfcpu accepts
			assert.NilError(t, api.Validate(t.Context(), bytes.NewReader(tt.pdf), conf, nil))

			pages, err := api.PageCount(t.Context(), bytes.NewReader(tt.pdf), conf)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(pages, tt.wantPages))

			document, err := pdftext.Open(t.Context(), tt.pdf)
			assert.NilError(t, err)

			text := string(document.Text(t.Context()))
			for _, want := range tt.wantText {
				assert.Check(t, strings.Contains(text, want), "missing %q in %q", want, text)
			}
		})
	}
}

func TestContentStream(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		mode     layout
		contains []string
	}{
		{
			name:     "simple layout escapes delimiters in literal strings",
			lines:    []string{`a\b(c)`},
			mode:     layoutSimple,
			contains: []string{`(a\\b\(c\)) Tj T*`},
		},
		{
			name:     "composite layout writes two byte hex strings",
			lines:    []string{"Hi"},
			mode:     layoutComposite,
			contains: []string{"<00480069> Tj T*"},
		},
		{
			name:     "glyph layout draws one Tj per glyph",
			lines:    []string{"Hi"},
			mode:     layoutGlyphs,
			contains: []string{"0 0 Td <0048> Tj ", "5.5 0 Td <0069> Tj "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := contentStream(tt.lines, tt.mode)

			assert.Check(t, strings.HasPrefix(stream, "<< /Length "))
			assert.Check(t, strings.Contains(stream, "BT /F1 11 Tf 72 740 Td 14 TL"))
			assert.Check(t, strings.HasSuffix(stream, "ET\nendstream"))

			for _, want := range tt.contains {
				assert.Check(t, strings.Contains(stream, want), "missing %q in %q", want, stream)
			}
		})
	}
}

func TestGlyphLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{name: "empty", line: "", want: ""},
		{name: "steps by the previous glyph width", line: "AB", want: "0 0 Td <0041> Tj 5.5 0 Td <0042> Tj "},
		{name: "spaces widen the step instead of drawing", line: "A B", want: "0 0 Td <0041> Tj 8.25 0 Td <0042> Tj "},
		{name: "leading space is a step before the first glyph", line: " A", want: "2.75 0 Td <0041> Tj "},
		{name: "digit and narrow widths", line: "1I2", want: "0 0 Td <0031> Tj 4.95 0 Td <0049> Tj 2.75 0 Td <0032> Tj "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Check(t, is.Equal(glyphLine(tt.line), tt.want))
		})
	}
}

func TestGlyphWidth(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want float64
	}{
		{name: "digit", r: '7', want: digitGlyphWidth},
		{name: "narrow I", r: 'I', want: narrowGlyphWidth},
		{name: "lowercase i uses default", r: 'i', want: defaultGlyphWidth},
		{name: "letter", r: 'W', want: defaultGlyphWidth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Check(t, is.Equal(glyphWidth(tt.r), tt.want))
		})
	}
}

func TestTwoByteHex(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "empty", text: "", want: ""},
		{name: "ascii", text: "AZ", want: "0041005A"},
		{name: "printable range edges", text: " ~", want: "0020007E"},
		{name: "control characters become question marks", text: "\t", want: "003F"},
		{name: "non ascii becomes question marks", text: "é", want: "003F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Check(t, is.Equal(twoByteHex(tt.text), tt.want))
		})
	}
}
