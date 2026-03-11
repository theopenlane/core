package emailruntime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
)

// validBaseData returns a data map that satisfies the required fields in ContextData.
func validBaseData() map[string]any {
	return map[string]any{
		"CompanyName": "Openlane",
		"Year":        2025,
		"FromEmail":   "no-reply@openlane.io",
		"URLS": map[string]any{
			"Root":    "https://app.openlane.io",
			"Product": "https://openlane.io",
			"Docs":    "https://docs.openlane.io",
		},
		"Recipient": map[string]any{
			"Email": "ada@example.com",
		},
	}
}

func TestTemplateContextSchema_TransactionalValidData(t *testing.T) {
	schema := TemplateContextSchema(enums.TemplateContextTransactional)
	require.NotNil(t, schema)

	err := validateTemplateData(schema, validBaseData())
	assert.NoError(t, err)
}

func TestTemplateContextSchema_TransactionalMissingRequired(t *testing.T) {
	schema := TemplateContextSchema(enums.TemplateContextTransactional)
	require.NotNil(t, schema)

	data := validBaseData()
	delete(data, "CompanyName")

	err := validateTemplateData(schema, data)
	assert.Error(t, err)
}

func TestTemplateContextSchema_CampaignValidData(t *testing.T) {
	schema := TemplateContextSchema(enums.TemplateContextCampaignRecipient)
	require.NotNil(t, schema)

	data := validBaseData()
	data["Campaign"] = map[string]any{
		"name": "Q1 Outreach",
	}

	err := validateTemplateData(schema, data)
	assert.NoError(t, err)
}

func TestTemplateContextSchema_WorkflowValidData(t *testing.T) {
	schema := TemplateContextSchema(enums.TemplateContextWorkflowAction)
	require.NotNil(t, schema)

	err := validateTemplateData(schema, validBaseData())
	assert.NoError(t, err)
}

func TestTemplateContextSchema_UnknownContextReturnsNil(t *testing.T) {
	assert.Nil(t, TemplateContextSchema(enums.TemplateContextInvalid))
}

func TestTemplateContextEntries_AllContextsCovered(t *testing.T) {
	entries := TemplateContextEntries()
	require.Len(t, entries, 3)

	for _, e := range entries {
		assert.NotEmpty(t, e.Label)
		assert.NotEmpty(t, e.Description)
		assert.NotNil(t, e.Schema, "context %s has nil schema", e.Context)
	}
}
