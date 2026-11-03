package summarizer

import (
	"gotest.tools/v3/assert"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestNewLexRankSummarizer_Summarize(t *testing.T) {

	maxSentences := 10

	tt := []struct {
		name     string
		sentence string
		hasError bool
	}{
		{
			name:     "empty string",
			hasError: true,
		},
		{
			name:     "extremely short string",
			sentence: gofakeit.LoremIpsumSentence(2),
		},
		{
			name:     "short string",
			sentence: gofakeit.Sentence(),
		},
		{
			name:     "long string",
			sentence: gofakeit.Paragraph(),
		},
		{
			name:     "really long string",
			sentence: gofakeit.LoremIpsumSentence(10000),
		},
	}

	summarizer := newLexRankSummarizer(maxSentences)

	for _, v := range tt {
		t.Run("Test "+v.name, func(t *testing.T) {

			summarized, err := summarizer.Summarize(t.Context(), v.sentence)
			if v.hasError {
				assert.Assert(t, err != nil)
				return
			}

			assert.NilError(t, err)
			assert.Assert(t, summarized != "")
		})
	}
}
