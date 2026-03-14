package handlers

import (
	"context"

	"github.com/theopenlane/newman/compose"

	"github.com/theopenlane/core/internal/emailruntime"
)

// sendEmail composes and queues a notification-template-driven email.
// ownerID scopes the template lookup to the owning organization; pass empty string for system-scoped lookup.
// key identifies the NotificationTemplate record. recipient provides the To address and Recipient template variables.
// dataBuilder defines typed template data overrides; pass nil to use only the base config and recipient data.
// opts are applied to the ComposeRequest before dispatch and carry per-call extras (tags, attachments, etc.).
func (h *Handler) sendEmail(ctx context.Context, ownerID string, key string, recipient compose.Recipient, dataBuilder *emailruntime.TemplateData, opts ...emailruntime.SendOption) error {
	return emailruntime.Send(ctx, h.DBClient, ownerID, key, recipient, dataBuilder, opts...)
}
