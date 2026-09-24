package hooks

import (
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
)

// init registers the branded message op as the email used by notify specs that list the email
// channel without naming their own operation
func init() {
	entityops.DefaultEmail = entityops.EmailVia(email.DefinitionID, email.BrandedMessageOp, defaultNotificationEmail).Batched()
}

// defaultNotificationEmail mirrors the in-app notification as a branded email: the title becomes
// the subject and headline, the body the single paragraph, and a url in the data the button
func defaultNotificationEmail(_ entityops.Invocation, _ entityops.MutationPayload, _ json.RawMessage, recipient entityops.EmailRecipient) (email.BrandedMessageRequest, error) {
	request := email.BrandedMessageRequest{
		RecipientInfo: emailRecipientInfo(recipient.Users...),
		Subject:       recipient.Title,
		Preheader:     recipient.Body,
		Title:         recipient.Title,
		Intros:        []string{recipient.Body},
	}

	if link, ok := recipient.Data["url"].(string); ok && link != "" {
		request.ButtonText = "View"
		request.ButtonLink = link
	}

	return request, nil
}

// emailRecipientInfo builds the email recipient block: a single user is addressed by name, several
// users share one message with every address on the recipient list
func emailRecipientInfo(users ...*generated.User) email.RecipientInfo {
	if len(users) == 0 {
		return email.RecipientInfo{}
	}

	if len(users) == 1 {
		return email.RecipientInfo{Email: users[0].Email, FirstName: users[0].FirstName, LastName: users[0].LastName}
	}

	return email.RecipientInfo{
		Email:      users[0].Email,
		Recipients: lo.Map(users, func(user *generated.User, _ int) string { return user.Email }),
	}
}
