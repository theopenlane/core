package oidc

import "errors"

var (
	// ErrOrganizationIDRequired indicates no organization was supplied to bind the assertion to
	ErrOrganizationIDRequired = errors.New("oidc: organization ID required")
	// ErrAudienceRequired indicates no federation audience was supplied
	ErrAudienceRequired = errors.New("oidc: audience required")
	// ErrEndpointRequired indicates no token exchange endpoint was supplied
	ErrEndpointRequired = errors.New("oidc: token exchange endpoint required")
	// ErrExchangeFailed indicates the RFC 8693 token exchange failed
	ErrExchangeFailed = errors.New("oidc: token exchange failed")
)
