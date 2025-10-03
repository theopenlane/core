package summarizer

import (
	"context"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

type summarizer interface {
	// Summarize takes in a long text and returns a summarized
	// version of the text.
	Summarize(context.Context, string) (string, error)
}

// Client takes in texts, strips out all html tags and
// tries to summarize it to be human readable and short
type Client struct {
	impl      summarizer
	sanitizer *bluemonday.Policy
}

// NewSummarizer returns a configured client based on the provided configuration
func NewSummarizer(cfg Config) (*Client, error) {
	sanitizer := bluemonday.StrictPolicy()

	switch cfg.Type {
	case TypeLexrank:
		return &Client{
			impl:      newLexRankSummarizer(cfg.MaximumSentences),
			sanitizer: sanitizer,
		}, nil

	case TypeLlm:
		impl, err := newLLMSummarizer(cfg)
		if err != nil {
			return nil, err
		}

		return &Client{
			impl:      impl,
			sanitizer: sanitizer,
		}, nil
	}

	return nil, ErrUnsupportedSummarizerType
}

// Summarize returns a shortened version of the provided string using the lexrank algorithm
func (s *Client) Summarize(ctx context.Context, sentence string) (string, error) {
	// also convert markdown to HTML before sanitizing
	sanitizedSentence := string(mdToHTML([]byte(sentence)))

	sanitizedSentence = s.sanitizer.Sanitize(sanitizedSentence)

	return s.impl.Summarize(ctx, sanitizedSentence)
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.SkipImages | html.SkipLinks | html.SkipHTML
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

// func htmlToText(htmlContent []byte) string {

// 	txt, err := html2text.FromString(string(htmlContent), html2text.Options{
// 		PrettyTables:      false, // tables often garble summarization; flatten them
// 		OmitLinks:         false, // keep link text; urls become [link]
// 		WordWrap:          0,     // don’t wrap; let your sentence splitter handle it
// 		BodyWidth:         0,
// 		PrettyTablesASCII: false,
// 	})
// 	if err != nil {
// 		return "", fmt.Errorf("html->text: %w", err)
// 	}

// 	return txt, nil
// }
