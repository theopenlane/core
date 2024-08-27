package openlaneclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Yamashou/gqlgenc/clientv2"
)

// WithAuthorizationAndSession adds the authorization header and session to the client request
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

// WithLoggingInterceptor adds a http debug logging interceptor
func WithLoggingInterceptor() clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		fmt.Println("Request header sent:", req.Header)
		fmt.Println("Request body sent:", req.Body)

		return next(ctx, req, gqlInfo, res)
	}
}

// WithEmptyInterceptor adds an empty interceptor
func WithEmptyInterceptor() clientv2.RequestInterceptor {
	return func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
		return next(ctx, req, gqlInfo, res)
	}
}
