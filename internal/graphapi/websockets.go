package graphapi

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
	"github.com/theopenlane/core/pkg/logx"
	authmw "github.com/theopenlane/core/pkg/middleware/auth"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
)

// createWebsocketClient creates a websocket transport with the appropriate settings
func (r *Resolver) createWebsocketClient() transport.Websocket {
	return transport.Websocket{
		KeepAlivePingInterval: r.websocketPingInterval,
		InitTimeout:           defaultInitTimeout,
		InitFunc:              r.webSocketInit,
		Upgrader:              r.upgraderFunc(),
	}
}

// createSSEClient creates a server-sent events transport with the appropriate settings
func (r *Resolver) createSSEClient() transport.SSE {
	return transport.SSE{
		KeepAlivePingInterval: r.sseKeepAliveInterval,
	}
}

// webSocketInit handles the websocket init payload for authentication and returns the context with the authenticated user
func (r *Resolver) webSocketInit(
	ctx context.Context,
	initPayload transport.InitPayload,
) (context.Context, *transport.InitPayload, error) {
	au, err := authmw.AuthenticateTransport(
		ctx,
		initPayload,
		r.authOptions,
	)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to authenticate websocket init payload")

		return ctx, nil, err
	}

	ctx = auth.WithAuthenticatedUser(ctx, au)

	return ctx, nil, nil
}

// upgraderFunc returns a websocket upgrader with the appropriate origin check
func (r *Resolver) upgraderFunc() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(req *http.Request) bool {
			if len(r.origins) == 0 {
				return true // allow all origins if none are set
			}

			o := req.Header.Get(httpsling.HeaderOrigin)
			if o == "" {
				return false
			}

			return checkOrigin(o, r.origins)
		},
	}
}

// checkOrigin checks if the given origin is allowed based on the allowed origins map
func checkOrigin(o string, allowedOrigins map[string]struct{}) bool {
	// check if the origin is in the allowed list
	_, ok := allowedOrigins[o]
	if ok {
		return ok
	}

	// check the same list, but for blob patterns, used sometimes in preview environments
	// such as  https://*openlane*.vercel.app
	allowOriginPatterns := allowedOriginsPatterns(allowedOrigins)
	for _, re := range allowOriginPatterns {
		if match := re.MatchString(o); match {
			return true
		}
	}

	return ok

}

// allowedOriginsPatterns converts allowed origins with wildcards into regex patterns
// this matches the cors middleware behavior to allow the same origins for websockets
func allowedOriginsPatterns(allowedOrigins map[string]struct{}) []*regexp.Regexp {
	allowOriginPatterns := make([]*regexp.Regexp, 0, len(allowedOrigins))

	for origin := range allowedOrigins {
		pattern := regexp.QuoteMeta(origin)
		pattern = strings.ReplaceAll(pattern, "\\*", ".*")
		pattern = strings.ReplaceAll(pattern, "\\?", ".")
		pattern = "^" + pattern + "$"

		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		allowOriginPatterns = append(allowOriginPatterns, re)
	}

	return allowOriginPatterns
}
