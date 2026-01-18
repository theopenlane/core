package common //nolint:revive

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	// AuthExtensionKey is the key used to store the auth data in the extensions map
	AuthExtensionKey = "auth"
	// ServerLatencyExtensionKey is the key used to store the server latency in the extensions map
	ServerLatencyExtensionKey = "server_latency"
	// TraceExtensionKey is the key used to store the trace id in the extensions map
	TraceExtensionKey = "trace_id"
	// ModulesExtensionKey is the key used to store the missing module in the extensions map
	ModulesExtensionKey = "missing_module"
)

// Auth contains the authentication data to be added to the extensions map
type Auth struct {
	// AuthenticationType is the type of authentication used, e.g. JWT, API key, etc.
	AuthenticationType auth.AuthenticationType `json:"authentication_type,omitempty"`
	// AuthorizedOrganizations is the organization ID(s) of the authenticated user
	AuthorizedOrganizations []string `json:"authorized_organization,omitempty"`
	// AccessToken is the access token used for authentication, if the user did an action (e.g. created a new organization)
	// that updated the access token, this will be the new access token
	AccessToken string `json:"access_token,omitempty"`
	// RefreshToken is the refresh token used for authentication, if the user did an action (e.g. created a new organization)
	// that updated the refresh token, this will be the new refresh token
	RefreshToken string `json:"refresh_token,omitempty"`
	// SessionID is the session token used for authentication
	SessionID string `json:"session_id,omitempty"`
}

// AddAllExtensions adds all the extensions to the server including auth, latency and trace
func AddAllExtensions(h *handler.Server) {
	// add the auth extension
	authExtension(h)
	// add the latency extension
	latencyExtension(h)
	// add the trace extension
	traceExtension(h)
	// add the modules extension
	modulesExtension(h)
}

// modulesExtension adds the missing module value if it exists into the extension map of the response
func modulesExtension(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		resp := next(ctx)

		// do not wrap subscriptions responses, they must return nil to close the connection
		optctx := graphql.GetOperationContext(ctx)
		if optctx.Operation.Operation == ast.Subscription {
			return resp
		}

		resp = initExtensionResponse(resp)

		if missingModule := getMissingModuleFromErrors(resp.Errors); missingModule != "" {
			resp.Extensions[ModulesExtensionKey] = missingModule
		}

		return resp
	})
}

func getMissingModuleFromErrors(errorList gqlerror.List) string {
	for _, err := range errorList {
		if customErr, ok := err.Err.(gqlerrors.CustomErrorType); ok {
			if module := customErr.Module(); module.String() != "" {
				return module.String()
			}
		}
	}

	return ""
}

// authExtension adds the auth data to the extensions map in the response
func authExtension(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		resp := next(ctx)

		// do not wrap subscriptions responses, they must return nil to close the connection
		optctx := graphql.GetOperationContext(ctx)
		if optctx.Operation.Operation == ast.Subscription {
			return resp
		}

		resp = initExtensionResponse(resp)

		resp.Extensions[AuthExtensionKey] = getAuthData(ctx)

		return resp
	})
}

// latencyExtension adds the server latency to the extensions map in the response
func latencyExtension(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		// do not wrap subscriptions responses, they must return nil to close the connection
		optctx := graphql.GetOperationContext(ctx)
		if optctx.Operation.Operation == ast.Subscription {
			return next(ctx)
		}

		start := time.Now()
		resp := next(ctx)
		latency := time.Since(start).String()

		resp = initExtensionResponse(resp)

		resp.Extensions[ServerLatencyExtensionKey] = latency

		return resp
	})
}

// traceExtension adds the trace id to the extensions map in the response
func traceExtension(h *handler.Server) {
	h.AroundResponses(func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		resp := next(ctx)

		// do not wrap subscriptions responses, they must return nil to close the connection
		optctx := graphql.GetOperationContext(ctx)
		if optctx.Operation.Operation == ast.Subscription {
			return resp
		}

		traceID := getRequestID(ctx)

		resp = initExtensionResponse(resp)

		resp.Extensions[TraceExtensionKey] = traceID

		return resp
	})
}

// initExtensionResponse initializes the extensions map in the response to avoid nil pointer panics
func initExtensionResponse(resp *graphql.Response) *graphql.Response {
	if resp == nil {
		resp = &graphql.Response{}
	}

	if resp.Extensions == nil {
		resp.Extensions = make(map[string]interface{})
	}

	return resp
}

// getRequestID retrieves the trace request id from the context
// if the echo context is not available an empty string is returned
func getRequestID(ctx context.Context) string {
	c, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return ""
	}

	// Prefer context value if set by middleware
	if rid := c.Get(echo.HeaderXRequestID); rid != nil {
		if s, ok := rid.(string); ok {
			return s
		}
	}

	// clone to ensure we don't have concurrent map access issues
	reqHeader := c.Request().Header.Clone()
	respHeader := c.Response().Header().Clone()

	requestID := reqHeader.Get(echo.HeaderXRequestID) // request-id generated by reverse-proxy
	if requestID == "" {
		// missed request-id from proxy, got generated one by middleware.RequestID()
		requestID = respHeader.Get(echo.HeaderXRequestID)
	}

	return requestID
}

// getAuthData retrieves the auth data from the context if available
// all errors are ignored because the auth data is optional
func getAuthData(ctx context.Context) Auth {
	ac, _ := auth.GetAuthenticatedUserFromContext(ctx)
	if ac == nil {
		// return early to prevent nil pointer panics
		return Auth{}
	}

	at, _ := auth.AccessTokenFromContext(ctx)
	rt, _ := auth.RefreshTokenFromContext(ctx)
	session, _ := sessions.SessionToken(ctx)

	return Auth{
		AuthenticationType:      ac.AuthenticationType,
		AuthorizedOrganizations: ac.OrganizationIDs,
		AccessToken:             at,
		RefreshToken:            rt,
		SessionID:               session,
	}
}
