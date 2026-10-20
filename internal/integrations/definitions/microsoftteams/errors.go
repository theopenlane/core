package microsoftteams

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("microsoftteams: oauth token missing")
	// ErrClientBuildFailed indicates the Graph SDK client could not be constructed
	ErrClientBuildFailed = errors.New("microsoftteams: client build failed")
	// ErrChannelMissing indicates the team_id or channel_id is missing
	ErrChannelMissing = errors.New("microsoftteams: team_id and channel_id are required")
	// ErrMessageEmpty indicates the message body is empty
	ErrMessageEmpty = errors.New("microsoftteams: message body is required")
	// ErrProfileLookupFailed indicates the Graph /me request failed
	ErrProfileLookupFailed = errors.New("microsoftteams: profile lookup failed")
	// ErrChannelMessageSendFailed indicates the Graph channel message request failed
	ErrChannelMessageSendFailed = errors.New("microsoftteams: channel message send failed")
	// ErrTeamsListFailed indicates the Graph joined teams request failed
	ErrTeamsListFailed = errors.New("microsoftteams: teams list failed")
	// ErrChannelsListFailed indicates the Graph team channels request failed
	ErrChannelsListFailed = errors.New("microsoftteams: channels list failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("microsoftteams: result encode failed")
	// ErrCredentialEncode indicates the credential could not be serialized
	ErrCredentialEncode = errors.New("microsoftteams: credential encode failed")
	// ErrTenantUnresolved indicates the access token did not carry a tenant identifier
	ErrTenantUnresolved = errors.New("microsoftteams: tenant unresolved")
	// ErrUserInputDecode indicates the installation user input could not be decoded
	ErrUserInputDecode = errors.New("microsoftteams: user input decode failed")
)
