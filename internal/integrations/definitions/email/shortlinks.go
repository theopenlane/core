package email

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/shortlinks"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

// runtimeShortlinks returns the shortlinks client carried by the runtime DB client
func runtimeShortlinks(req types.OperationRequest) *shortlinks.Client {
	if req.DB == nil {
		return nil
	}

	return req.DB.Shortlinks
}

// linkRecipient is the address a link is attributed to
func linkRecipient(recipient RecipientInfo) string {
	if len(recipient.Recipients) > 1 {
		return ""
	}

	return recipient.Email
}

// trackingLink shortens a tracking-only link
func trackingLink(ctx context.Context, req types.OperationRequest, link string, meta shortlinks.Metadata, scope string) string {
	return urlx.Shorten(ctx, runtimeShortlinks(req), shortlinks.CreateRequest{
		URL:      link,
		Slug:     shortlinks.DeterministicSlug(string(meta.Purpose), link, meta.RecipientEmail, meta.OrganizationID, meta.UserID, meta.TrustCenterID, scope),
		Metadata: meta,
	})
}
