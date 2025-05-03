package summarizer

import (
	"context"
	"errors"
	"strings"

	"github.com/didasy/tldr"
)

var (
	ErrSentenceEmpty             = errors.New("you cannot summarize an empty string")
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

	return strings.Join(vals, " "), nil
}
