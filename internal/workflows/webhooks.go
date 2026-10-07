package workflows

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

// blockedWebhookHeaders are lowercased request headers that cloud metadata services require
var blockedWebhookHeaders = map[string]struct{}{
	"metadata-flavor":                      {},
	"x-google-metadata-request":            {},
	"metadata":                             {},
	"x-aws-ec2-metadata-token":             {},
	"x-aws-ec2-metadata-token-ttl-seconds": {},
}

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

		if err := validateWebhookDestination(params, allowPrivateAddresses); err != nil {
			return fmt.Errorf("%w: %s", err, action.Key)
		}
	}

	return nil
}

func validateWebhookDestination(params WebhookActionParams, allowPrivateAddresses bool) error {
	for name := range params.Headers {
		if _, blocked := blockedWebhookHeaders[strings.ToLower(strings.TrimSpace(name))]; blocked {
			return fmt.Errorf("%w: %s", ErrWebhookHeaderNotAllowed, name)
		}
	}

	if allowPrivateAddresses || strings.TrimSpace(params.URL) == "" {
		return nil
	}

	if _, err := urlx.ValidatePublicURL(params.URL); err != nil {
		return fmt.Errorf("%w: %w", ErrWebhookURLNotPublic, err)
	}

	return nil
}
