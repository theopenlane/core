package slack

const (
	slackAuthURL  = "https://slack.com/oauth/v2/authorize"
	slackTokenURL = "https://slack.com/api/oauth.v2.access"
)

var slackScopes = []string{
	"chat:write",
	"chat:write.public",
	"chat:write.customize",
	"team:read",
	"users:read",
}
