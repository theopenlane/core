package graphapi

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/workflows"
	mwauth "github.com/theopenlane/core/pkg/middleware/auth"
	"github.com/theopenlane/core/pkg/gala"
)

// WithTrustCenterCnameTarget sets the trust center cname target for the resolver
func (r Resolver) WithTrustCenterCnameTarget(cname string) *Resolver {
	r.trustCenterCnameTarget = cname

	return &r
}

// WithTrustCenterDefaultDomain sets the default trust center domain for the resolver
func (r Resolver) WithTrustCenterDefaultDomain(domain string) *Resolver {
	r.defaultTrustCenterDomain = domain

	return &r
}

// WithExtensions enables or disables graph extensions
func (r Resolver) WithExtensions(enabled bool) *Resolver {
	r.extensionsEnabled = enabled

	return &r
}

// WithDevelopment sets the resolver to development mode
// when isDevelopment is false, introspection will be disabled
func (r Resolver) WithDevelopment(dev bool) *Resolver {
	r.isDevelopment = dev

	return &r
}

// WithAllowedOrigins sets the allowed origins for websocket connections
func (r Resolver) WithAllowedOrigins(origins []string) *Resolver {
	r.origins = make(map[string]struct{}, len(origins))
	for _, o := range origins {
		r.origins[o] = struct{}{}
	}

	return &r
}

// WithAuthOptions sets the auth options for the resolver
func (r Resolver) WithAuthOptions(options ...mwauth.Option) *Resolver {
	opts := mwauth.NewAuthOptions(options...)
	r.authOptions = &opts

	return &r
}

// WithComplexityLimitConfig sets the complexity limit for the resolver
func (r Resolver) WithComplexityLimitConfig(limit int) *Resolver {
	r.complexityLimit = limit

	return &r
}

// WithMaxResultLimit sets the max result limit in the config for the resolvers
func (r Resolver) WithMaxResultLimit(limit int) *Resolver {
	r.maxResultLimit = &limit

	return &r
}

// WithWorkflowsConfig sets the workflows config for CEL validation in resolvers.
func (r Resolver) WithWorkflowsConfig(cfg workflows.Config) *Resolver {
	r.workflowsConfig = cfg

	return &r
}

// WithWebsocketPingInterval sets the websocket ping interval for the resolver
func (r Resolver) WithWebsocketPingInterval(interval time.Duration) *Resolver {
	r.websocketPingInterval = interval

	return &r
}

// WithSSEKeepAliveInterval sets the sse keep-alive interval for the resolver
func (r Resolver) WithSSEKeepAliveInterval(interval time.Duration) *Resolver {
	r.sseKeepAliveInterval = interval

	return &r
}

// WithComplexityLimit adds a complexity limit middleware to the handler
func (r *Resolver) WithComplexityLimit(h *handler.Server) {
	// prevent complex queries except the introspection query
	h.Use(common.NewComplexityLimitWithMetrics(func(_ context.Context, rc *graphql.OperationContext) int {
		if rc != nil && rc.OperationName == "IntrospectionQuery" {
			return common.IntrospectionComplexity
		}

		if rc.OperationName == "GlobalSearch" {
			// allow more complexity for the global search
			// e.g. if the complexity limit is 100, we allow 500 for the global search
			return r.complexityLimit * 5 //nolint:mnd
		}

		if r.complexityLimit > 0 {
			return r.complexityLimit
		}

		return common.DefaultComplexityLimit
	}))
}

// WithPool adds a worker pool to the resolver for parallel processing
func (r *Resolver) WithPool(maxWorkers int) {
	r.pool = gala.NewPool(
		gala.WithWorkers(maxWorkers),
		gala.WithPoolName("graphapi-worker-pool"),
	)
}
