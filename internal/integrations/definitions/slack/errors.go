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
	// ErrAuthTestFailed indicates auth.test failed
	ErrAuthTestFailed = errors.New("slack: auth test failed")
	// ErrTeamInfoFailed indicates team.info failed
	ErrTeamInfoFailed = errors.New("slack: team info failed")
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("slack: operation config invalid")
	// ErrConversationsListFailed indicates conversations.list failed
	ErrConversationsListFailed = errors.New("slack: conversations list failed")
	// ErrMessageSendFailed indicates chat.postMessage failed
	ErrMessageSendFailed = errors.New("slack: message send failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("slack: result encode failed")
	// ErrUsersFetchFailed indicates the workspace users list request failed
	ErrUsersFetchFailed = errors.New("slack: users fetch failed")
	// ErrPayloadEncode indicates a collected Slack payload could not be serialized for ingest
	ErrPayloadEncode = errors.New("slack: ingest payload encode failed")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("slack: credential encode failed")
	// ErrCredentialDecode indicates the credential could not be deserialized
	ErrCredentialDecode = errors.New("slack: credential decode failed")
)
