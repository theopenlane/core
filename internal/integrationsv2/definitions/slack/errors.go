package slack

import "errors"

var (
	// ErrOAuthTokenMissing indicates the Slack OAuth access token is missing from the credential
	ErrOAuthTokenMissing = errors.New("slack: oauth token missing")
	// ErrChannelMissing indicates the Slack channel is missing from the operation config
	ErrChannelMissing = errors.New("slack: channel missing")
	// ErrMessageEmpty indicates the Slack message has no content
	ErrMessageEmpty = errors.New("slack: message must have text, blocks, or attachments")
	// ErrClientType indicates the provided client is not a Slack client
	ErrClientType = errors.New("slack: unexpected client type")
)
