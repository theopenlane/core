package googleworkspace

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
)

var googleWorkspaceScopes = []string{
	"https://www.googleapis.com/auth/admin.directory.user.readonly",
	"https://www.googleapis.com/auth/admin.directory.group.readonly",
	"https://www.googleapis.com/auth/apps.groups.migration",
}

var googleWorkspaceAuthParams = map[string]string{
	"access_type": "offline",
	"prompt":      "consent",
}
