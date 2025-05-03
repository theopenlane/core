package summarizer

import (
	"context"

	"github.com/microcosm-cc/bluemonday"

	"github.com/theopenlane/core/internal/ent/entconfig"
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

func NewSummarizer(cfg entconfig.Config) (*Client, error) {
	sanitizer := bluemonday.StrictPolicy()

	switch cfg.Summarizer.Type {
	case entconfig.SummarizerTypeLexrank:
		return &Client{
			impl:      newLexRankSummarizer(cfg.Summarizer.MaximumSentences),
			sanitizer: sanitizer,
		}, nil

	case entconfig.SummarizerTypeLlm:
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
	sanitizedSentence := s.sanitizer.Sanitize(sentence)

	return s.impl.Summarize(ctx, sanitizedSentence)
}
