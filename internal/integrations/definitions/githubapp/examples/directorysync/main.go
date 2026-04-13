//go:build ignore

// Debugs GitHub App directory sync by querying organizations and members via GraphQL.
// Mirrors the production queryViewerOrganizations and queryOrganizationMembers paths,
// then also queries the REST installation metadata to compare what the app can see.
//
//	go run main.go <installation_id> [config_path]
//
// installation_id: the GitHub App installation to authenticate as
// config_path:     path to the koanf config yaml (default: config/.config.yaml, relative to repo root)
package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	gh "github.com/google/go-github/v84/github"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run main.go <installation_id> [config_path]")
		os.Exit(1)
	}

	installationID, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil || installationID == 0 {
		fmt.Fprintf(os.Stderr, "invalid installation_id: %s\n", os.Args[1])
		os.Exit(1)
	}

	cfgPath := "config/.config.yaml"
	if len(os.Args) > 2 {
		cfgPath = os.Args[2]
	}

	appID, privateKey, err := loadAppConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	fmt.Printf("minting JWT for app %s...\n", appID)

	jwtToken, err := mintJWT(appID, privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "jwt: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("requesting installation token for installation %d...\n", installationID)

	token, err := mintInstallationToken(ctx, installationID, jwtToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "installation token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("token expires at %s\n\n", token.Expiry.Format(time.RFC3339))

	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	gqlClient := graphql.NewClient("https://api.github.com/graphql", httpClient)
	restClient := gh.NewClient(httpClient)

	// step 1: check who the viewer is
	fmt.Println("=== viewer identity ===")

	var viewerQuery struct {
		Viewer struct {
			Login     string
			TypeName  string `graphql:"__typename"`
			AvatarURL string `graphql:"avatarUrl"`
		}
	}

	if err := gqlClient.Query(ctx, &viewerQuery, nil); err != nil {
		fmt.Fprintf(os.Stderr, "viewer query: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("viewer login: %s (type: %s)\n\n", viewerQuery.Viewer.Login, viewerQuery.Viewer.TypeName)

	// step 2: query viewer.organizations (this is what production does)
	fmt.Println("=== viewer.organizations (GraphQL — production path) ===")

	var orgQuery struct {
		Viewer struct {
			Organizations struct {
				TotalCount int
				Nodes      []struct {
					Login string
				}
			} `graphql:"organizations(first: 100)"`
		}
	}

	if err := gqlClient.Query(ctx, &orgQuery, nil); err != nil {
		fmt.Fprintf(os.Stderr, "viewer organizations query: %v\n", err)
	} else {
		fmt.Printf("total organizations via viewer: %d\n", orgQuery.Viewer.Organizations.TotalCount)
		for _, org := range orgQuery.Viewer.Organizations.Nodes {
			fmt.Printf("  - %s\n", org.Login)
		}

		if orgQuery.Viewer.Organizations.TotalCount == 0 {
			fmt.Println("  (empty — this is the bug: app bot is not a member of any org)")
		}
	}

	fmt.Println()

	// step 3: query installation metadata via app JWT (requires app-level auth, not installation token)
	fmt.Println("=== REST: installation details (via app JWT) ===")

	jwtHTTPClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwtToken}))
	jwtRestClient := gh.NewClient(jwtHTTPClient)

	installation, resp, err := jwtRestClient.Apps.GetInstallation(ctx, installationID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get installation: %v\n", err)
	} else {
		resp.Body.Close()

		fmt.Printf("installation target type: %s\n", installation.GetTargetType())
		if installation.GetAccount() != nil {
			fmt.Printf("installation account: %s (type: %s)\n", installation.GetAccount().GetLogin(), installation.GetAccount().GetType())
		}
	}

	fmt.Println()

	// step 4: derive orgs from repo owners (works with installation token)
	fmt.Println("=== orgs derived from viewer.repositories ===")

	var repoQuery struct {
		Viewer struct {
			Repositories struct {
				Nodes []struct {
					NameWithOwner string
					Owner         struct {
						Login    string
						TypeName string `graphql:"__typename"`
					}
				}
			} `graphql:"repositories(first: 100, orderBy: {field: UPDATED_AT, direction: DESC})"`
		}
	}

	if err := gqlClient.Query(ctx, &repoQuery, nil); err != nil {
		fmt.Fprintf(os.Stderr, "repo query: %v\n", err)
	} else {
		orgsFromRepos := map[string]string{}
		for _, r := range repoQuery.Viewer.Repositories.Nodes {
			orgsFromRepos[r.Owner.Login] = r.Owner.TypeName
		}

		fmt.Printf("unique owners from repos: %d\n", len(orgsFromRepos))
		for login, typename := range orgsFromRepos {
			fmt.Printf("  - %s (type: %s)\n", login, typename)
		}
	}

	fmt.Println()

	// step 5: try querying the org directly using the account from the installation
	var orgLogin string
	if installation != nil && installation.GetAccount() != nil {
		orgLogin = installation.GetAccount().GetLogin()
	}

	if orgLogin != "" {
		fmt.Printf("=== direct organization query: %s (GraphQL) ===\n", orgLogin)

		var directOrgQuery struct {
			Organization struct {
				Login           string
				MembersWithRole struct {
					TotalCount int
					Nodes      []struct {
						Login                          string
						Email                          string
						Name                           string
						OrganizationVerifiedDomainEmails []string `graphql:"organizationVerifiedDomainEmails(login: $login)"`
					}
				} `graphql:"membersWithRole(first: 10)"`
			} `graphql:"organization(login: $login)"`
		}

		variables := map[string]any{
			"login": graphql.String(orgLogin),
		}

		if err := gqlClient.Query(ctx, &directOrgQuery, variables); err != nil {
			fmt.Fprintf(os.Stderr, "direct org members query: %v\n", err)
		} else {
			fmt.Printf("members via direct org query: %d\n", directOrgQuery.Organization.MembersWithRole.TotalCount)
			for _, m := range directOrgQuery.Organization.MembersWithRole.Nodes {
				fmt.Printf("  - %s <%s> (%s) verified_emails=%v\n", m.Login, m.Email, m.Name, m.OrganizationVerifiedDomainEmails)
			}
		}

		fmt.Println()

		// step 6: query SAML/SCIM external identities (requires SSO and org admin permissions)
		fmt.Printf("=== SAML/SCIM external identities for %s ===\n", orgLogin)

		var samlQuery struct {
			Organization struct {
				SamlIdentityProvider *struct {
					ExternalIdentities struct {
						TotalCount int
						Nodes      []struct {
							Guid         string
							SamlIdentity *struct {
								NameID   *string `graphql:"nameId"`
								Username *string
								Emails   []struct {
									Value string
								}
								GivenName  *string
								FamilyName *string
								Groups     []string
							}
							ScimIdentity *struct {
								Username   *string
								Emails     []struct {
									Value string
								}
								GivenName  *string
								FamilyName *string
								Groups     []string
							}
							User *struct {
								Login string
								Name  string
								Email string
							}
						}
					} `graphql:"externalIdentities(first: 100)"`
				}
			} `graphql:"organization(login: $login)"`
		}

		samlVars := map[string]any{
			"login": graphql.String(orgLogin),
		}

		if err := gqlClient.Query(ctx, &samlQuery, samlVars); err != nil {
			fmt.Fprintf(os.Stderr, "SAML identity query: %v\n", err)
			fmt.Println("  (this likely requires organization_administration:read permission)")
		} else if samlQuery.Organization.SamlIdentityProvider == nil {
			fmt.Println("  (no SAML identity provider configured for this org)")
		} else {
			ids := samlQuery.Organization.SamlIdentityProvider.ExternalIdentities
			fmt.Printf("external identities: %d\n", ids.TotalCount)
			for _, id := range ids.Nodes {
				login := "(unlinked)"
				if id.User != nil {
					login = id.User.Login
				}

				samlEmail := ""
				if id.SamlIdentity != nil && id.SamlIdentity.NameID != nil {
					samlEmail = *id.SamlIdentity.NameID
				}

				scimUser := ""
				if id.ScimIdentity != nil && id.ScimIdentity.Username != nil {
					scimUser = *id.ScimIdentity.Username
				}

				fmt.Printf("  - github=%s  saml_nameid=%s  scim_user=%s  guid=%s\n", login, samlEmail, scimUser, id.Guid)

				if id.SamlIdentity != nil {
					for _, e := range id.SamlIdentity.Emails {
						fmt.Printf("      saml_email: %s\n", e.Value)
					}
				}

				if id.ScimIdentity != nil {
					for _, e := range id.ScimIdentity.Emails {
						fmt.Printf("      scim_email: %s\n", e.Value)
					}
				}
			}
		}

		fmt.Println()

		// step 7: also try REST org members for comparison
		fmt.Printf("=== REST: org members for %s ===\n", orgLogin)

		members, resp, err := restClient.Organizations.ListMembers(ctx, orgLogin, &gh.ListMembersOptions{
			ListOptions: gh.ListOptions{PerPage: 10},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "REST list members: %v (status: %d)\n", err, resp.StatusCode)
		} else {
			resp.Body.Close()

			fmt.Printf("members via REST: %d (first page)\n", len(members))
			for _, m := range members {
				fmt.Printf("  - %s\n", m.GetLogin())
			}
		}
	}

	fmt.Println()

	// step 7: check installation permissions (what scopes do we have?)
	fmt.Println("=== installation permissions ===")

	if installation != nil && installation.Permissions != nil {
		perms := installation.Permissions
		fmt.Printf("  members:        %s\n", stringOrNone(perms.Members))
		fmt.Printf("  organization:   %s\n", stringOrNone(perms.OrganizationAdministration))
		fmt.Printf("  administration: %s\n", stringOrNone(perms.Administration))
		fmt.Printf("  metadata:       %s\n", stringOrNone(perms.Metadata))
	} else {
		fmt.Println("  (no permission data available)")
	}
}

func stringOrNone(s *string) string {
	if s == nil {
		return "(not set)"
	}

	return *s
}

func loadAppConfig(path string) (string, string, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return "", "", fmt.Errorf("load %s: %w", path, err)
	}

	appID := k.String("integrations.githubapp.appid")
	privateKey := k.String("integrations.githubapp.privatekey")

	if appID == "" || privateKey == "" {
		return "", "", fmt.Errorf("integrations.githubapp.appid and privatekey are required in %s", path)
	}

	return appID, privateKey, nil
}

func mintJWT(appID, privateKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	var key any
	var err error

	key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("parse private key: %w", err)
		}
	}

	now := time.Now()

	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		Issuer:    appID,
		IssuedAt:  jwt.NewNumericDate(now.Add(-30 * time.Second)),
		ExpiresAt: jwt.NewNumericDate(now.Add(9 * time.Minute)),
	}).SignedString(key)
	if err != nil {
		return "", fmt.Errorf("sign JWT: %w", err)
	}

	return signed, nil
}

func mintInstallationToken(ctx context.Context, installationID int64, jwtToken string) (*oauth2.Token, error) {
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwtToken}))
	client := gh.NewClient(httpClient)

	tok, resp, err := client.Apps.CreateInstallationToken(ctx, installationID, &gh.InstallationTokenOptions{})
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("github responded %d: %w", resp.StatusCode, err)
		}

		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return &oauth2.Token{
		AccessToken: tok.GetToken(),
		TokenType:   "Bearer",
		Expiry:      tok.GetExpiresAt().Time,
	}, nil
}
