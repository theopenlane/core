package hooks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

func TestRequirePublicWebhookURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		action  models.WorkflowAction
		wantErr error
	}{
		{
			name:   "public webhook url",
			action: models.WorkflowAction{Type: enums.WorkflowActionTypeWebhook.String(), Params: json.RawMessage(`{"url":"https://hooks.example.com/x"}`)},
		},
		{
			name:    "loopback webhook url",
			action:  models.WorkflowAction{Type: enums.WorkflowActionTypeWebhook.String(), Params: json.RawMessage(`{"url":"http://127.0.0.1:8080/x"}`)},
			wantErr: urlx.ErrNonPublicHost,
		},
		{
			name:    "lowercase webhook type still checked",
			action:  models.WorkflowAction{Type: "webhook", Params: json.RawMessage(`{"url":"http://169.254.169.254/"}`)},
			wantErr: urlx.ErrNonPublicHost,
		},
		{
			name:    "unparsable webhook url",
			action:  models.WorkflowAction{Type: enums.WorkflowActionTypeWebhook.String(), Params: json.RawMessage(`{"url":"not-a-url"}`)},
			wantErr: urlx.ErrUnsupportedScheme,
		},
		{
			name:   "non webhook action ignored",
			action: models.WorkflowAction{Type: enums.WorkflowActionTypeNotification.String(), Params: json.RawMessage(`{"url":"http://127.0.0.1/"}`)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := requirePublicWebhookURL(tc.action)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}
