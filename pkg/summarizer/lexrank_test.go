package summarizer

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
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
			name:     "extremly short string",
			sentence: gofakeit.Sentence(2),
		},
		{
			name:     "short string",
			sentence: gofakeit.Sentence(200),
		},
		{
			name:     "long string",
			sentence: gofakeit.Sentence(1000),
		},
		{
			name:     "really long string",
			sentence: gofakeit.Sentence(10000),
		},
	}

	summarizer := newLexRankSummarizer(maxSentences)

	for _, v := range tt {
		t.Run("Test "+v.name, func(t *testing.T) {

			summarized, err := summarizer.Summarize(t.Context(), v.sentence)
			if v.hasError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, summarized)
		})
	}
}
