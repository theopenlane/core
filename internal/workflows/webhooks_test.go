package workflows

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/models"
)

func webhookDocument(params string) models.WorkflowDefinitionDocument {
	return models.WorkflowDefinitionDocument{
		Actions: []models.WorkflowAction{
			{Key: "hook", Type: "WEBHOOK", Params: json.RawMessage(params)},
		},
	}
}

func TestValidateWebhookDestinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		doc          models.WorkflowDefinitionDocument
		allowPrivate bool
		wantErr      error
	}{
		{
			name: "public webhook",
			doc:  webhookDocument(`{"url": "https://example.com/hook", "headers": {"X-Custom": "value"}}`),
		},
		{
			name: "non webhook action is ignored",
			doc: models.WorkflowDefinitionDocument{
				Actions: []models.WorkflowAction{
					{Key: "notify", Type: "NOTIFICATION", Params: json.RawMessage(`{"url": "http://169.254.169.254/"}`)},
				},
			},
		},
		{
			name: "lowercase webhook type still checked",
			doc: models.WorkflowDefinitionDocument{
				Actions: []models.WorkflowAction{
					{Key: "hook", Type: "webhook", Params: json.RawMessage(`{"url": "http://169.254.169.254/"}`)},
				},
			},
			wantErr: ErrWebhookURLNotPublic,
		},
		{
			name:    "gcp metadata hostname",
			doc:     webhookDocument(`{"url": "http://metadata.google.internal/computeMetadata/v1/instance/id"}`),
			wantErr: ErrWebhookURLNotPublic,
		},
		{
			name:    "link local metadata ip",
			doc:     webhookDocument(`{"url": "http://169.254.169.254/computeMetadata/v1/"}`),
			wantErr: ErrWebhookURLNotPublic,
		},
		{
			name:    "loopback ip",
			doc:     webhookDocument(`{"url": "http://127.0.0.2:8080/admin"}`),
			wantErr: ErrWebhookURLNotPublic,
		},
		{
			name:    "private ip",
			doc:     webhookDocument(`{"url": "https://10.0.0.5/hook"}`),
			wantErr: ErrWebhookURLNotPublic,
		},
		{
			name:    "metadata flavor header",
			doc:     webhookDocument(`{"url": "https://example.com/hook", "headers": {"Metadata-Flavor": "Google"}}`),
			wantErr: ErrWebhookHeaderNotAllowed,
		},
		{
			name:    "malformed params",
			doc:     webhookDocument(`{"url": 5}`),
			wantErr: ErrWebhookParamsInvalid,
		},
		{
			name:         "private address allowed by config",
			doc:          webhookDocument(`{"url": "http://127.0.0.1:8080/hook"}`),
			allowPrivate: true,
		},
		{
			name:         "metadata header still blocked when private addresses are allowed",
			doc:          webhookDocument(`{"url": "http://127.0.0.1:8080/hook", "headers": {"Metadata-Flavor": "Google"}}`),
			allowPrivate: true,
			wantErr:      ErrWebhookHeaderNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateWebhookDestinations(tt.doc, tt.allowPrivate)
			if tt.wantErr == nil {
				assert.NilError(t, err)
				return
			}

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
