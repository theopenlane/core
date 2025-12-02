package summarizer

import (
	"context"
	"errors"
	"strings"

	"github.com/didasy/tldr"
)

var (
	// ErrSentenceEmpty is used to denote required sentences that needs to be summarized
	ErrSentenceEmpty = errors.New("you cannot summarize an empty string")

	// ErrUnsupportedSummarizerType is used to denote an summarization algorithm we do not support at the moment
	ErrUnsupportedSummarizerType = errors.New("unsupported summarizer type")
)

type lexRankSummarizer struct {
	maxSentences int
}

// newLexRankSummarizer returns a summarization engine using the lexrank algorithm
func newLexRankSummarizer(maxSentences int) *lexRankSummarizer {
	return &lexRankSummarizer{
		maxSentences: maxSentences,
	}
}

// Summarize returns a shortened version of the provided string using the lexrank algorithm
func (l *lexRankSummarizer) Summarize(_ context.Context, s string) (string, error) {
	if strings.TrimSpace(s) == "" {
		return "", ErrSentenceEmpty
	}

	vals, err := tldr.New().Summarize(s, l.maxSentences)
	if err != nil {
		return "", err
	}

	summarized := strings.Join(vals, " ")
	// very very short strings like <= 30 chars cannot be summarized with this algorithm
	// so just return the same details here. It is short already
	// so works fine as a "summary"
	if len(summarized) == 0 {
		summarized = s
	}

	return summarized, nil
}
