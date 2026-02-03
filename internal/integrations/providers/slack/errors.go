package slack

import "errors"

var (
	// ErrAPIRequest indicates a Slack API request failed with a non-2xx status
	ErrAPIRequest = errors.New("slack: api request failed")
	// ErrSlackAPIError indicates the Slack API returned an error in the response body
	ErrSlackAPIError = errors.New("slack: api returned error")
	// ErrOAuthTokenMissing indicates the OAuth token is not present in the credential payload
	ErrOAuthTokenMissing = errors.New("slack: oauth token missing")
	// ErrAccessTokenEmpty indicates the access token field is empty
	ErrAccessTokenEmpty = errors.New("slack: access token empty")
	// ErrSlackChannelMissing indicates the Slack message channel is missing
	ErrSlackChannelMissing = errors.New("slack: message channel missing")
	// ErrSlackMessageEmpty indicates the Slack message is empty
	ErrSlackMessageEmpty = errors.New("slack: message text or blocks required")
)
