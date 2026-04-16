package summarizer

import (
	"bytes"
	"context"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	goldmarkparser "github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
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

	if strings.TrimSpace(sanitizedSentence) == "" {
		return "", nil
	}

	return s.impl.Summarize(ctx, sanitizedSentence)
}

func mdToHTML(md []byte) []byte {
	gm := goldmark.New(
		goldmark.WithParserOptions(
			goldmarkparser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := gm.Convert(md, &buf); err != nil {
		return md
	}

	return buf.Bytes()
}
