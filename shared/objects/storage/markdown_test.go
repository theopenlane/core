package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/shared/objects/storage"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedFM    *storage.Frontmatter
		expectedBody  string
		expectedError bool
	}{
		{
			name:  "valid frontmatter with content",
			input: "---\ntitle: Test Title\nopenlane_id: 01K9P6G47035WRQ6967THJRFYA\n---\n\n# Main Content\n\nThis is the body.",
			expectedFM: &storage.Frontmatter{
				Title: "Test Title",
			},
			expectedBody:  "# Main Content\n\nThis is the body.",
			expectedError: false,
		},
		{
			name:          "no frontmatter",
			input:         "# Just Content\n\nNo frontmatter here.",
			expectedFM:    nil,
			expectedBody:  "# Just Content\n\nNo frontmatter here.",
			expectedError: false,
		},
		{
			name:          "incomplete frontmatter - only one separator",
			input:         "---\ntitle: Test\n\nContent without closing separator",
			expectedFM:    nil,
			expectedBody:  "---\ntitle: Test\n\nContent without closing separator",
			expectedError: false,
		},
		{
			name:  "empty frontmatter",
			input: "---\n---\n\nContent after empty frontmatter",
			expectedFM: &storage.Frontmatter{
				Title: "",
			},
			expectedBody:  "Content after empty frontmatter",
			expectedError: false,
		},
		{
			name:  "frontmatter with only title",
			input: "---\ntitle: Only Title\n---\n\nBody content",
			expectedFM: &storage.Frontmatter{
				Title: "Only Title",
			},
			expectedBody:  "Body content",
			expectedError: false,
		},
		{
			name:  "frontmatter with extra fields",
			input: "---\nopenlane_id: 01K9P6G47035WRQ6967THJRFYA\n---\n\nBody content",
			expectedFM: &storage.Frontmatter{
				Title: "",
			},
			expectedBody:  "Body content",
			expectedError: false,
		},
		{
			name:          "invalid yaml in frontmatter",
			input:         "---\ntitle: Test\ninvalid: [unclosed\n---\n\nContent",
			expectedFM:    nil,
			expectedBody:  "",
			expectedError: true,
		},
		{
			name:  "frontmatter with extra whitespace",
			input: "---\n  title: Spaced Title  \n  openlane_id: spaced-id  \n---\n\n  Body with spaces  \n",
			expectedFM: &storage.Frontmatter{
				Title: "Spaced Title",
			},
			expectedBody:  "Body with spaces",
			expectedError: false,
		},
		{
			name:          "content starting with --- but not frontmatter",
			input:         "--- This is not YAML frontmatter\nJust regular content",
			expectedFM:    nil,
			expectedBody:  "--- This is not YAML frontmatter\nJust regular content",
			expectedError: false,
		},
		{
			name:  "empty body after frontmatter",
			input: "---\ntitle: Test\n---\n",
			expectedFM: &storage.Frontmatter{
				Title: "Test",
			},
			expectedBody:  "",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := storage.ParseFrontmatter([]byte(tt.input))

			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.expectedFM == nil {
				assert.Nil(t, fm)
			} else {
				require.NotNil(t, fm)
				assert.Equal(t, tt.expectedFM.Title, fm.Title)
			}

			assert.Equal(t, tt.expectedBody, string(body))
		})
	}
}
