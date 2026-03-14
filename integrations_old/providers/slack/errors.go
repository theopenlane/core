package slack

import "errors"

var (
	// ErrSlackChannelMissing indicates the Slack message channel is missing
	ErrSlackChannelMissing = errors.New("slack: message channel missing")
	// ErrSlackMessageEmpty indicates the Slack message is empty
	ErrSlackMessageEmpty = errors.New("slack: message text or blocks required")
)
