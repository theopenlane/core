package summarizer

import (
	"context"
	"errors"
)

var (
	ErrSentenceEmpty = errors.New("you cannot summarize an empty string")
)

type Summarizer interface {
	// Summarize takes in a long text and returns a summarized
	// version of the text.
	Summarize(context.Context, string) (string, error)
}
