package summarizer

import (
	"errors"
	"testing"

	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Summarize(t *testing.T) {
	client, err := NewSummarizer(Config{
		Type:             TypeLexrank,
		MaximumSentences: 10,
	})
	require.NoError(t, err)

	tt := []struct {
		name      string
		input     string
		wantEmpty bool
	}{
		{
			name:      "empty string",
			input:     "",
			wantEmpty: true,
		},
		{
			name:      "whitespace only",
			input:     "   \n\t\n  ",
			wantEmpty: true,
		},
		{
			name:      "image only markdown",
			input:     "![alt text](http://example.com/img.png)",
			wantEmpty: true,
		},
		{
			name:      "multiple images only",
			input:     "![a](http://example.com/1.png)\n\n![b](http://example.com/2.png)",
			wantEmpty: true,
		},
		{
			name:  "html content preserves text",
			input: "<p>This is raw HTML that should be sanitized before summarization.</p>",
		},
		{
			name:  "plain text",
			input: "This is a plain text sentence that should be summarized properly.",
		},
		{
			name:  "markdown with heading and paragraph",
			input: "# Heading\n\nThis is a paragraph with enough content to be summarized by the algorithm.",
		},
		{
			name:  "markdown with bold and italic",
			input: "This sentence has **bold** and *italic* formatting that should be stripped.",
		},
		{
			name:  "markdown list",
			input: "- first item in the list\n- second item in the list\n- third item in the list",
		},
		{
			name:  "markdown link",
			input: "Visit [the documentation](http://example.com) for more details on this topic.",
		},
		{
			name:  "mixed markdown with images and text",
			input: "# Report\n\n![chart](http://example.com/chart.png)\n\nThe chart above shows the quarterly results.",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Summarize(t.Context(), tc.input)

			require.NoError(t, err)
			assert.False(t, errors.Is(err, ErrSentenceEmpty), "Client.Summarize must never surface ErrSentenceEmpty")

			if tc.wantEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
			}
		})
	}
}

// TestClient_Summarize_UGCSanitizedInput mirrors the real data path: documents
// imported via file upload or URL fetch are sanitized with bluemonday.UGCPolicy()
// before being stored as details. UGCPolicy preserves HTML tags like <p>, <strong>,
// <a>, etc. The summarizer must handle this HTML-containing input and extract text
// for summarization — not discard it.
func TestClient_Summarize_UGCSanitizedInput(t *testing.T) {
	client, err := NewSummarizer(Config{
		Type:             TypeLexrank,
		MaximumSentences: 10,
	})
	require.NoError(t, err)

	ugc := bluemonday.UGCPolicy()

	tt := []struct {
		name string
		raw  string
	}{
		{
			name: "paragraph with heading",
			raw:  "<h1>Policy Document</h1><p>This policy governs the use of company resources and outlines acceptable behavior.</p>",
		},
		{
			name: "bold and paragraph tags",
			raw:  "<p>Section 1: <strong>Overview</strong></p><p>The quick brown fox jumps over the lazy dog.</p>",
		},
		{
			name: "unordered list",
			raw:  "<ul><li>Item one has content</li><li>Item two has content</li><li>Item three has content</li></ul>",
		},
		{
			name: "paragraph with link",
			raw:  `<p>Contact us at <a href="mailto:support@example.com">support@example.com</a> for more details.</p>`,
		},
		{
			name: "nested divs with content",
			raw:  "<div><h2>Section Title</h2><p>This section contains important compliance requirements for all employees.</p></div>",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// simulate the import path: UGCPolicy sanitizes first, result is stored as details
			details := ugc.Sanitize(tc.raw)

			result, err := client.Summarize(t.Context(), details)
			require.NoError(t, err, "summarizing UGC-sanitized input must not error")
			assert.False(t, errors.Is(err, ErrSentenceEmpty),
				"UGC-sanitized HTML must not produce ErrSentenceEmpty")
			assert.NotEmpty(t, result, "UGC-sanitized input with real text content must produce a summary")
		})
	}
}

// TestClient_Summarize_NoSentenceEmptyError covers inputs that legitimately contain
// no extractable text after the full pipeline. These must return ("", nil) rather
// than surfacing ErrSentenceEmpty to the caller.
func TestClient_Summarize_NoSentenceEmptyError(t *testing.T) {
	client, err := NewSummarizer(Config{
		Type:             TypeLexrank,
		MaximumSentences: 10,
	})
	require.NoError(t, err)

	inputs := []struct {
		name  string
		input string
	}{
		{"image only", "![alt](http://example.com/img.png)"},
		{"horizontal rule", "---"},
		{"html comment", "<!-- comment -->"},
		{"image in heading", "# ![icon](icon.png) "},
	}

	for _, tc := range inputs {
		t.Run(tc.name, func(t *testing.T) {
			result, err := client.Summarize(t.Context(), tc.input)
			require.NoError(t, err)
			assert.False(t, errors.Is(err, ErrSentenceEmpty),
				"input %q must not produce ErrSentenceEmpty", tc.input)
			assert.Empty(t, result)
		})
	}
}

func TestMdToHTML(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		wantSub string
	}{
		{
			name:    "plain text wrapped in paragraph",
			input:   "hello world",
			wantSub: "<p>hello world</p>",
		},
		{
			name:    "heading gets id",
			input:   "# Title",
			wantSub: `<h1 id="title">Title</h1>`,
		},
		{
			name:    "bold renders strong",
			input:   "**bold**",
			wantSub: "<strong>bold</strong>",
		},
		{
			name:    "empty input returns empty",
			input:   "",
			wantSub: "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := string(mdToHTML([]byte(tc.input)))
			assert.Contains(t, result, tc.wantSub)
		})
	}
}
