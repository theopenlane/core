package workflows

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/validator"
)

// ValidateWebhookDestinations rejects webhook actions that send cloud metadata headers or, unless
// allowPrivateAddresses is set, target non-public destinations
func ValidateWebhookDestinations(doc models.WorkflowDefinitionDocument, allowPrivateAddresses bool) error {
	for _, action := range doc.Actions {
		actionType := enums.ToWorkflowActionType(action.Type)
		if actionType == nil || *actionType != enums.WorkflowActionTypeWebhook || len(action.Params) == 0 {
			continue
		}

		var params WebhookActionParams
		if err := json.Unmarshal(action.Params, &params); err != nil {
			return fmt.Errorf("%w: %s", ErrWebhookParamsInvalid, action.Key)
		}

		if err := validator.ValidateOutboundHeaders()(params.Headers); err != nil {
			return fmt.Errorf("%w: %s", err, action.Key)
		}

		if allowPrivateAddresses || strings.TrimSpace(params.URL) == "" {
			continue
		}

		if err := validator.ValidatePublicURL()(params.URL); err != nil {
			return fmt.Errorf("%w: %s", err, action.Key)
		}
	}

	return nil
}
