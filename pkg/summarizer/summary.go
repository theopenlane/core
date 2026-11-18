package summarizer

import (
	"bytes"
	"context"
	"html"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark/v2/extension"
	goldmarkparser "github.com/yuin/goldmark/v2/parser"
	goldmarkhtml "github.com/yuin/goldmark/v2/renderer/html"
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

	headerTags := []string{
		"h1", "h2", "h3", "h4", "h5", "h6",
	}

	sanitizer.SkipElementsContent(headerTags...)

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

	sanitizedSentence = html.UnescapeString(s.sanitizer.Sanitize(sanitizedSentence))

	if strings.TrimSpace(sanitizedSentence) == "" {
		return "", nil
	}

	return s.impl.Summarize(ctx, sanitizedSentence)
}

func mdToHTML(md []byte) []byte {
	p := goldmarkparser.New(
		goldmarkparser.WithAutoHeadingID(),
		goldmarkparser.WithExtensions(extension.NewTableParser()),
	)
	r := goldmarkhtml.New(
		goldmarkhtml.WithUnsafe(),
		goldmarkhtml.WithExtensions(extension.NewTableHTMLRenderer()),
	)

	var buf bytes.Buffer
	if err := r.Render(&buf, md, p.Parse(md)); err != nil {
		return md
	}

	return buf.Bytes()
}
