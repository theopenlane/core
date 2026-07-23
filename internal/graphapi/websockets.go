package graphapi

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/coder/websocket"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
	authmw "github.com/theopenlane/core/pkg/middleware/auth"
)

var websocketConnectionTrackerContextKey = contextx.NewKey[time.Time]()

// CreateWebsocketClient creates a websocket transport with the appropriate settings
func (r *Resolver) CreateWebsocketClient() transport.Websocket {
	return transport.Websocket{
		KeepAlivePingInterval: r.websocketPingInterval,
		InitTimeout:           defaultInitTimeout,
		InitFunc:              r.webSocketInit,
		Implementation:        r.websocketImplementation(),
		CloseFunc: func(ctx context.Context, _ int) {
			t, ok := websocketConnectionTrackerContextKey.Get(ctx)
			if !ok {
				return
			}

			metrics.RecordSubscriptionClosed(time.Since(t).Seconds())
		},
	}
}

// createSSEClient creates a server-sent events transport with the appropriate settings
func (r *Resolver) createSSEClient() transport.SSE {
	return transport.SSE{
		KeepAlivePingInterval: r.sseKeepAliveInterval,
	}
}

// webSocketInit handles the websocket init payload for authentication and returns the context with the authenticated caller
func (r *Resolver) webSocketInit(
	ctx context.Context,
	initPayload transport.InitPayload,
) (context.Context, *transport.InitPayload, error) {
	caller, err := authmw.AuthenticateTransport(
		ctx,
		initPayload,
		r.authOptions,
	)
	if err != nil {
		logx.FromContext(ctx).Info().Str("error", err.Error()).Msg("failed to authenticate websocket init payload")

		return ctx, nil, err
	}

	logx.FromContext(ctx).Debug().Str("user_id", caller.SubjectID).Msg("websocket connection authenticated")

	ctx = auth.WithCaller(ctx, caller)
	ctx = websocketConnectionTrackerContextKey.Set(ctx, time.Now())

	metrics.RecordSubscriptionOpened()

	return ctx, nil, nil
}

// websocketImplementation returns a websocket implementation with the appropriate origin check
func (r *Resolver) websocketImplementation() transport.WebsocketImplementation {
	return originCheckedWebsocket{origins: r.origins}
}

// originCheckedWebsocket wraps the coder websocket implementation with an origin check
type originCheckedWebsocket struct {
	origins map[string]struct{}
}

// Accept validates the request origin before delegating the upgrade to the coder implementation
func (o originCheckedWebsocket) Accept(w http.ResponseWriter, req *http.Request, options transport.WebsocketAcceptOptions) (transport.WebsocketConn, error) {
	if !o.originAllowed(req) {
		return nil, ErrOriginNotAllowed
	}

	impl := transport.CoderWebsocketImplementation{
		// origin verification is handled above, matching origins against the full origin string rather than the host
		AcceptOptions: websocket.AcceptOptions{InsecureSkipVerify: true},
	}

	return impl.Accept(w, req, options)
}

// originAllowed reports whether the request origin is allowed based on the configured origins
func (o originCheckedWebsocket) originAllowed(req *http.Request) bool {
	if len(o.origins) == 0 {
		return true // allow all origins if none are set
	}

	origin := req.Header.Get(httpsling.HeaderOrigin)
	if origin == "" {
		return false
	}

	return checkOrigin(origin, o.origins)
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
