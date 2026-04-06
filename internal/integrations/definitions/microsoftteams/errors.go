package microsoftteams

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("microsoftteams: oauth token missing")
	// ErrChannelMissing indicates the team_id or channel_id is missing
	ErrChannelMissing = errors.New("microsoftteams: team_id and channel_id are required")
	// ErrMessageEmpty indicates the message body is empty
	ErrMessageEmpty = errors.New("microsoftteams: message body is required")
	// ErrProfileLookupFailed indicates the Graph /me request failed
	ErrProfileLookupFailed = errors.New("microsoftteams: profile lookup failed")
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("microsoftteams: operation config invalid")
	// ErrChannelMessageSendFailed indicates the Graph channel message request failed
	ErrChannelMessageSendFailed = errors.New("microsoftteams: channel message send failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("microsoftteams: result encode failed")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("microsoftteams: credential encode failed")
	// ErrCredentialDecode indicates the credential could not be deserialized
	ErrCredentialDecode = errors.New("microsoftteams: credential decode failed")
)
