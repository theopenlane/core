package summarizer

import (
	"context"

	"github.com/theopenlane/core/internal/ent/entconfig"
)

type summarizer interface {
	// Summarize takes in a long text and returns a summarized
	// version of the text.
	Summarize(context.Context, string) (string, error)
}

type SummarizerClient struct {
	impl summarizer
}

func NewSummarizer(cfg entconfig.Config) (*SummarizerClient, error) {

	switch cfg.Summarizer.Type {
	case entconfig.SummarizerTypeLexrank:
		return &SummarizerClient{
			impl: newLexRankSummarizer(cfg.Summarizer.MaximumSentences),
		}, nil

	case entconfig.SummarizerTypeLlm:

		impl, err := newLLMSummarizer(cfg)
		if err != nil {
			return nil, err
		}

		return &SummarizerClient{
			impl: impl,
		}, nil
	}

	return nil, ErrUnsupportedSummarizerType
}
