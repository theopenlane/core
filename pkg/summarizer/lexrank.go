package summarizer

import (
	"context"
	"strings"
	"sync"

	"github.com/didasy/tldr"
)

type LexRankSummarizer struct {
	maxSentences int
	bag          *tldr.Bag
	// *tldr.Bag is not thread-safe
	mu sync.Mutex
}

// NewLexRankSummarizer returns a summarization engine using the lexrank algorithm
func NewLexRankSummarizer(maxSentences int) *LexRankSummarizer {
	return &LexRankSummarizer{
		bag:          tldr.New(),
		mu:           sync.Mutex{},
		maxSentences: maxSentences,
	}
}

// Summarize returns a shortened version of the provided string using the lexrank algorithm
func (l *LexRankSummarizer) Summarize(ctx context.Context, s string) (string, error) {
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
