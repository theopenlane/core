package microsoftteams

const (
	teamsAuthURL  = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	teamsTokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)

var teamsScopes = []string{
	"https://graph.microsoft.com/User.Read",
	"https://graph.microsoft.com/Team.ReadBasic.All",
	"https://graph.microsoft.com/Channel.ReadBasic.All",
	"https://graph.microsoft.com/ChannelMessage.Send",
	"offline_access",
}
