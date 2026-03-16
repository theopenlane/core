package azureentraid

const (
	azureAuthURL  = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	azureTokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
)

var azureEntraScopes = []string{
	"https://graph.microsoft.com/.default",
	"offline_access",
}
