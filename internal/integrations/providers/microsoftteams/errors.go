package microsoftteams

import "errors"

var (
	// ErrTeamsChannelMissing indicates the team/channel identifiers are missing
	ErrTeamsChannelMissing = errors.New("microsoft teams: team_id and channel_id required")
	// ErrTeamsMessageEmpty indicates the Teams message body is empty
	ErrTeamsMessageEmpty = errors.New("microsoft teams: message body required")
	// ErrTeamsMessageFormatInvalid indicates an unsupported body format was provided
	ErrTeamsMessageFormatInvalid = errors.New("microsoft teams: message body format invalid")
)
