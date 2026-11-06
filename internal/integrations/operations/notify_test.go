package operations

import (
	"errors"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/enums"
)

func TestImportOutcome(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		result    IngestResult
		ingestErr error
		detail    string
		url       string
		wantTitle string
		wantBody  string
		wantData  map[string]any
	}{
		{
			name:      "completed",
			result:    IngestResult{Attempted: 3, Succeeded: 3, Persisted: 3},
			wantTitle: "Demo data import completed",
			wantBody:  "Imported 3 record(s)",
			wantData:  map[string]any{"succeeded": 3, "persisted": 3, "skipped": 0, "failed": 0, "failure_summary": ""},
		},
		{
			name:      "completed with detail and url",
			result:    IngestResult{Attempted: 3, Succeeded: 3, Persisted: 3},
			detail:    "you can remove this data from your organization settings",
			url:       "/organization-settings/general-settings#demo-data",
			wantTitle: "Demo data import completed",
			wantBody:  "Imported 3 record(s), you can remove this data from your organization settings",
			wantData:  map[string]any{"succeeded": 3, "persisted": 3, "skipped": 0, "failed": 0, "failure_summary": "", "url": "/organization-settings/general-settings#demo-data"},
		},
		{
			name:      "matched existing records",
			result:    IngestResult{Attempted: 3, Succeeded: 3, Persisted: 2, Skipped: 1},
			wantTitle: "Demo data import completed",
			wantBody:  "Imported 2 record(s), 1 record(s) matched existing records and were left in place",
			wantData:  map[string]any{"succeeded": 3, "persisted": 2, "skipped": 1, "failed": 0, "failure_summary": ""},
		},
		{
			name:      "partial failure with detail",
			result:    IngestResult{Attempted: 3, Succeeded: 2, Persisted: 2, Failed: 1, Failures: []RecordFailure{{Schema: "control", Resource: "demo#1", Message: "invalid status"}}},
			detail:    "contact support",
			wantTitle: "Demo data import completed",
			wantBody:  "Imported 2 record(s), 1 record(s) could not be imported, contact support",
			wantData:  map[string]any{"succeeded": 2, "persisted": 2, "skipped": 0, "failed": 1, "failure_summary": "1 of 3 records failed to import; first failure: control demo#1: invalid status"},
		},
		{
			name:      "failed with detail",
			ingestErr: errors.New("bucket unreachable"),
			detail:    "contact support",
			wantTitle: "Demo data import failed",
			wantBody:  "The import could not be completed, contact support",
			wantData:  map[string]any{"succeeded": 0, "persisted": 0, "skipped": 0, "failed": 0, "failure_summary": ""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content := importOutcome("Demo data", tc.result, tc.ingestErr, tc.detail, tc.url)

			assert.Equal(t, tc.wantTitle, content.TitleTemplate)
			assert.Equal(t, tc.wantBody, content.BodyTemplate)
			assert.DeepEqual(t, tc.wantData, content.Data)
			assert.Equal(t, enums.NotificationTopic(""), content.Topic)
		})
	}
}

func TestDeleteOutcome(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		deleted   map[string]int
		deleteErr error
		wantTitle string
		wantBody  string
	}{
		{
			name:      "completed",
			deleted:   map[string]int{"program": 1, "control": 4},
			wantTitle: "Demo data removed",
			wantBody:  "Removed 5 record(s)",
		},
		{
			name:      "stopped",
			deleted:   map[string]int{"procedure": 2},
			deleteErr: errors.New("delete failed"),
			wantTitle: "Demo data removal failed",
			wantBody:  "Removed 2 record(s) before the removal stopped, run it again to remove the rest",
		},
		{
			name:      "nothing deleted",
			deleted:   map[string]int{},
			wantTitle: "Demo data removed",
			wantBody:  "Removed 0 record(s)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			content := deleteOutcome("Demo data", tc.deleted, tc.deleteErr)

			assert.Equal(t, tc.wantTitle, content.TitleTemplate)
			assert.Equal(t, tc.wantBody, content.BodyTemplate)
			assert.DeepEqual(t, map[string]any{"deleted": tc.deleted}, content.Data)
		})
	}
}
