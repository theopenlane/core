package microsoftteams

const (
	// teamsAuthURL is the Microsoft identity platform authorization endpoint for the Teams OAuth flow
	teamsAuthURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	// teamsTokenURL is the Microsoft identity platform token endpoint for the Teams OAuth flow
	teamsTokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)

// teamsScopes lists the Microsoft Graph OAuth scopes requested for the Teams flow
var teamsScopes = []string{
	"https://graph.microsoft.com/User.Read",
	"https://graph.microsoft.com/Team.ReadBasic.All",
	"https://graph.microsoft.com/Channel.ReadBasic.All",
	"https://graph.microsoft.com/ChannelMessage.Send",
	"offline_access",
}
