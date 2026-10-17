package microsoft

import "fmt"

const (
	// loginBaseURL is the Microsoft identity platform base URL for one tenant segment
	loginBaseURL = "https://login.microsoftonline.com/%s"
	// CommonTenant is the tenant segment that accepts both work/school and personal Microsoft accounts
	CommonTenant = "common"
	// OrganizationsTenant is the tenant segment that accepts any work/school account
	OrganizationsTenant = "organizations"
)

// loginURL returns the Microsoft identity platform URL for a tenant and path
func loginURL(tenant, path string) string {
	return fmt.Sprintf(loginBaseURL, tenant) + path
}

// AuthorizeURL returns the OAuth2 v2.0 authorize endpoint for a tenant
func AuthorizeURL(tenant string) string {
	return loginURL(tenant, "/oauth2/v2.0/authorize")
}

// TokenURL returns the OAuth2 v2.0 token endpoint for a tenant
func TokenURL(tenant string) string {
	return loginURL(tenant, "/oauth2/v2.0/token")
}

// AdminConsentURL returns the v2.0 admin consent endpoint for a tenant
func AdminConsentURL(tenant string) string {
	return loginURL(tenant, "/v2.0/adminconsent")
}
