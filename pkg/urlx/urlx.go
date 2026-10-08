package urlx

import (
	"context"
	"net/url"

	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/shortlinks"
)

// tokenQueryParam is the query parameter carrying a signed token on generated links
const tokenQueryParam = "token"

// TokenURL appends a token query parameter to the base URL
func TokenURL(baseURL url.URL, token string) string {
	return baseURL.ResolveReference(&url.URL{
		RawQuery: url.Values{tokenQueryParam: []string{token}}.Encode(),
	}).String()
}

// Shorten creates a shortlink for req.URL. The original URL is returned unchanged when sl is
// nil, when the context suppresses shortening, or when the service call fails, so delivery of a
// link never depends on the service being reachable
func Shorten(ctx context.Context, sl *shortlinks.Client, req shortlinks.CreateRequest) string {
	if sl == nil || req.URL == "" || shortlinks.DisabledFromContext(ctx) {
		return req.URL
	}

	shortened, err := sl.Create(ctx, req)
	if err != nil {
		// the full link may carry a confidential token, so only its host and the purpose are logged
		host, _ := NormalizeHostname(req.URL)
		logx.FromContext(ctx).Error().Str("host", host).Str("purpose", string(req.Metadata.Purpose)).Err(err).Msg("failed to shorten URL, using original")

		return req.URL
	}

	return shortened
}
