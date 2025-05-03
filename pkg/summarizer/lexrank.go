package summarizer

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/didasy/tldr"
)

var (
	ErrSentenceEmpty             = errors.New("you cannot summarize an empty string")
	ErrUnsupportedSummarizerType = errors.New("unsupported summarizer type")
)

type lexRankSummarizer struct {
	maxSentences int
	bag          *tldr.Bag
	// *tldr.Bag is not thread-safe
	mu sync.Mutex
}

// newLexRankSummarizer returns a summarization engine using the lexrank algorithm
func newLexRankSummarizer(maxSentences int) *lexRankSummarizer {
	return &lexRankSummarizer{
		bag:          tldr.New(),
		mu:           sync.Mutex{},
		maxSentences: maxSentences,
	}
}

// Summarize returns a shortened version of the provided string using the lexrank algorithm
func (l *lexRankSummarizer) Summarize(ctx context.Context, s string) (string, error) {
	if strings.TrimSpace(s) == "" {
		return "", ErrSentenceEmpty
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	vals, err := l.bag.Summarize(s, l.maxSentences)
	if err != nil {
		return "", err
	}

	return strings.Join(vals, " "), nil
}
