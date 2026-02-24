package hooks

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// CampaignEmailContextKey carries optional campaign-specific email tagging data.
// When present, hooks can add the IDs as email tags for webhook correlation.
type CampaignEmailContextKey struct {
	CampaignID       string
	CampaignTargetID string
}

var campaignEmailContextKey = contextx.NewKey[CampaignEmailContextKey]()

// WithCampaignEmailContext attaches campaign email metadata to the context.
func WithCampaignEmailContext(ctx context.Context, data CampaignEmailContextKey) context.Context {
	return campaignEmailContextKey.Set(ctx, data)
}

// CampaignEmailContextFrom returns campaign email metadata when present.
func CampaignEmailContextFrom(ctx context.Context) (CampaignEmailContextKey, bool) {
	return campaignEmailContextKey.Get(ctx)
}
