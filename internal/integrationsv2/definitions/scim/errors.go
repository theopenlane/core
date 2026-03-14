package scim

import "errors"

var (
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("scim: unexpected client type")
	// ErrWebhookNotImplemented indicates the SCIM webhook handler is not yet implemented
	ErrWebhookNotImplemented = errors.New("scim: webhook handler not yet implemented")
)
