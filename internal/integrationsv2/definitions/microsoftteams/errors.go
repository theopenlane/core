package microsoftteams

import "errors"

var (
	// ErrOAuthTokenMissing indicates the OAuth access token is missing
	ErrOAuthTokenMissing = errors.New("microsoftteams: oauth token missing")
	// ErrClientType indicates the provided client is not the expected type
	ErrClientType = errors.New("microsoftteams: unexpected client type")
	// ErrChannelMissing indicates the team_id or channel_id is missing
	ErrChannelMissing = errors.New("microsoftteams: team_id and channel_id are required")
	// ErrMessageEmpty indicates the message body is empty
	ErrMessageEmpty = errors.New("microsoftteams: message body is required")
	// ErrBodyFormatInvalid indicates the body format is not text or html
	ErrBodyFormatInvalid = errors.New("microsoftteams: body_format must be text or html")
)
