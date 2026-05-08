package slack

import "errors"

var (
	// ErrOAuthTokenMissing indicates the Slack OAuth access token is missing from the credential
	ErrOAuthTokenMissing = errors.New("slack: oauth token missing")
	// ErrBotTokenMissing indicates the Slack bot token is missing from the credential
	ErrBotTokenMissing = errors.New("slack: bot token missing")
	// ErrNoCredentialResolved indicates neither OAuth nor bot token credential was found
	ErrNoCredentialResolved = errors.New("slack: no credential resolved")
	// ErrChannelMissing indicates the Slack channel is missing from the operation config
	ErrChannelMissing = errors.New("slack: channel missing")
	// ErrMessageEmpty indicates the Slack message has no content
	ErrMessageEmpty = errors.New("slack: message must have text, blocks, or attachments")
	// ErrClientType indicates the provided client is not a Slack client
	ErrClientType = errors.New("slack: unexpected client type")
	// ErrAuthTestFailed indicates auth.test failed
	ErrAuthTestFailed = errors.New("slack: auth test failed")
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("slack: operation config invalid")
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
	// ErrInstallationMetadataDecode indicates installation metadata could not be decoded from credential data
	ErrInstallationMetadataDecode = errors.New("slack: installation metadata decode failed")
	// ErrTeamIDMissing indicates the Slack team ID is missing
	ErrTeamIDMissing = errors.New("slack: installation id missing")
	// ErrInstallationMetadataEncode indicates installation metadata could not be encoded
	ErrInstallationMetadataEncode = errors.New("slack: installation metadata encode failed")
	// ErrClientBuildFailed indicates a Slack runtime or customer client could not be constructed
	ErrClientBuildFailed = errors.New("slack: client build failed")
	// ErrRuntimeConfigInvalid indicates the runtime Slack configuration is missing required fields
	ErrRuntimeConfigInvalid = errors.New("slack: runtime config invalid")
	// ErrRuntimeConfigDecode indicates the runtime Slack configuration could not be deserialized
	ErrRuntimeConfigDecode = errors.New("slack: runtime config decode failed")
	// ErrDefaultChannelMissing indicates the installation has no default channel configured for system messages
	ErrDefaultChannelMissing = errors.New("slack: default channel missing")
	// ErrTemplateRenderFailed indicates a system message template could not be rendered
	ErrTemplateRenderFailed = errors.New("slack: template render failed")
	// ErrInstallationInputDecode indicates the installation input payload could not be deserialized
	ErrInstallationInputDecode = errors.New("slack: installation input decode failed")
)
