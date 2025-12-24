package testclient

import (
	"context"
	"net/http"

	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
)

// WithAuthorization adds the authorization header and session to the client request
func (a Authorization) WithAuthorization() clientv2.RequestInterceptor {
	return func(
		ctx context.Context,
		req *http.Request,
		gqlInfo *clientv2.GQLRequestInfo,
		res interface{},
		next clientv2.RequestInterceptorFunc,
	) error {
		// setting authorization header if its not already set
		a.SetAuthorizationHeader(req)

		// add session cookie
		a.SetSessionCookie(req)

		return next(ctx, req, gqlInfo, res)
	}
}

// WithCSRFTokenInterceptor adds a CSRF token interceptor to the client request
// in the header
func WithCSRFTokenInterceptor(token string) clientv2.RequestInterceptor {
	return func(
		ctx context.Context,
		req *http.Request,
		gqlInfo *clientv2.GQLRequestInfo,
		res interface{},
		next clientv2.RequestInterceptorFunc,
	) error {
		// set the CSRF token in the request header if it is not empty
		if token != "" {
			req.Header.Set(csrfHeader, token)
		}

		return next(ctx, req, gqlInfo, res)
	}
}

// WithLoggingInterceptor adds a http debug logging interceptor
func WithLoggingInterceptor() clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		log.Info().Interface("request body", req.Body).Msg("request body sent")

		return next(ctx, req, gqlInfo, res)
	}
}

// WithEmptyInterceptor adds an empty interceptor
func WithEmptyInterceptor() clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		return next(ctx, req, gqlInfo, res)
	}
}

// WithImpersonationInterceptor adds impersonation headers to the request
func WithImpersonationInterceptor(userID string, orgID string) clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		if userID != "" {
			req.Header.Set(auth.UserIDHeader, userID)
		}

		if orgID != "" {
			req.Header.Set(auth.OrganizationIDHeader, orgID)
		}

		return next(ctx, req, gqlInfo, res)
	}
}

// WithOrganizationHeader adds the organization id header to the request
// this is required when using personal access tokens that are authorized for more than one organization
func WithOrganizationHeader(orgID string) clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		if orgID != "" {
			req.Header.Set(auth.OrganizationIDHeader, orgID)
		}

		return next(ctx, req, gqlInfo, res)
	}
}
