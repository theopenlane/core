package email

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/utils/jsonx"
)

func TestRegistrationAcceptsPerSendFields(t *testing.T) {
	dispatcher, ok := dispatcherIndex[brandedMessageKey]
	assert.Assert(t, ok)

	payload, err := json.Marshal(BrandedMessageRequest{
		RecipientInfo: RecipientInfo{Email: "user@example.com", FirstName: "Ada", LastName: "Lovelace"},
		Subject:       "Your report is ready to review",
		Title:         "Your report is ready to review",
		Tables:        []MessageTable{{Columns: []string{"Section"}, Rows: [][]string{{"controls"}}}},
	})
	assert.NilError(t, err)

	assert.NilError(t, jsonx.Validate(dispatcher.Registration().Input.Schema, payload))
}
