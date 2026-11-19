package interceptors

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/intercept"
	"github.com/theopenlane/core/v2/pkg/objects/storage/proxy"
)

// shareLinkURLField is the computed share link field holding the share URL
const shareLinkURLField = "url"

// InterceptorShareLinkURL sets the url field on share link query results when it is requested, for the links the
// caller created; the URL carries the share token, so it is never handed to another member of the organization
func InterceptorShareLinkURL() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.ShareLinkFunc(func(ctx context.Context, q *generated.ShareLinkQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			if q.ObjectManager == nil || !graphutils.CheckForRequestedField(ctx, shareLinkURLField) {
				return v, nil
			}

			caller, ok := auth.CallerFromContext(ctx)
			if !ok {
				return v, nil
			}

			links, ok := v.([]*generated.ShareLink)
			if !ok {
				return v, nil
			}

			baseURL := q.ObjectManager.Links().BaseURL

			for _, link := range links {
				setShareLinkURL(link, caller, baseURL)
			}

			return v, nil
		})
	})
}

// setShareLinkURL sets the share URL on the link when the caller is the member who created it
func setShareLinkURL(link *generated.ShareLink, caller *auth.Caller, baseURL string) {
	if link.CreatedBy != caller.SubjectID {
		return
	}

	link.URL = proxy.ComposeShareURL(baseURL, link.ID, link.Token)
}
