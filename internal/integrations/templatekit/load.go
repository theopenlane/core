package templatekit

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/notificationtemplate"
)

// LoadNotificationTemplate loads an active notification template by ID or key for the given owner
func LoadNotificationTemplate(ctx context.Context, client *generated.Client, ownerID, templateID, templateKey string) (*generated.NotificationTemplate, error) {
	if templateID == "" && templateKey == "" {
		return nil, ErrTemplateNotFound
	}

	query := client.NotificationTemplate.Query().
		Where(
			notificationtemplate.ActiveEQ(true),
			notificationtemplate.OwnerIDEQ(ownerID),
		)

	switch {
	case templateID != "":
		query = query.Where(notificationtemplate.IDEQ(templateID))
	case templateKey != "":
		query = query.Where(notificationtemplate.KeyEQ(templateKey))
	}

	template, err := query.Only(ctx)
	if generated.IsNotFound(err) {
		return nil, ErrTemplateNotFound
	}

	if err != nil {
		return nil, err
	}

	return template, nil
}
