package urlx

import (
	"context"
	"net/url"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/shortlinks"
)

// BuildTokenURL appends a token query parameter to the base URL and optionally
// shortens the result via the shortlinks client. If sl is nil or shortening
// fails, the full-length URL is returned (graceful degradation)
func BuildTokenURL(ctx context.Context, sl *shortlinks.Client, baseURL url.URL, token string) (string, error) {
	resolved := baseURL.ResolveReference(&url.URL{
		RawQuery: url.Values{"token": []string{token}}.Encode(),
	})

	regularLink := resolved.String()

	if sl == nil {
		return regularLink, nil
	}

	shortenedURL, err := sl.Create(ctx, regularLink, "")
	if err != nil {
		// don't log the full link as it contains a confidential token, just log the base URL
		logx.FromContext(ctx).Error().Str("baseURL", baseURL.String()).Err(err).Msg("failed to shorten URL, using original")

		return regularLink, nil
	}

	return shortenedURL, nil
}
