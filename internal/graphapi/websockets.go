package graphapi

import (
	"context"
	"net/http"
	"time"

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
		InitTimeout:           10 * time.Second,
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

func (r *Resolver) webSocketInit(
	ctx context.Context,
	initPayload transport.InitPayload,
) (context.Context, *transport.InitPayload, error) {
	logx.FromContext(ctx).Warn().Msg("websocket init payload received")

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

			// check if the origin is in the allowed list
			_, ok := r.origins[o]

			return ok
		},
	}
}
