//go:build ignore

// Verifies the GitHub App client against the real GitHub API by querying vulnerability alerts.
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
	"strings"
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

	client := graphql.NewClient("https://api.github.com/graphql", oauth2.NewClient(ctx, oauth2.StaticTokenSource(token)))

	// step 1: list repositories accessible to the installation
	repos, err := queryRepositories(ctx, client)
	if err != nil {
		fmt.Fprintf(os.Stderr, "repository query: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("found %d repositories, querying vulnerability alerts...\n\n", len(repos))

	// step 2: query vulnerability alerts for each repository
	totalAlerts := 0

	for _, repo := range repos {
		alerts, err := queryVulnerabilityAlerts(ctx, client, repo.NameWithOwner)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s: error: %v\n", repo.NameWithOwner, err)

			continue
		}

		if len(alerts) == 0 {
			continue
		}

		totalAlerts += len(alerts)
		fmt.Printf("  %s — %d alerts\n", repo.NameWithOwner, len(alerts))

		for _, a := range alerts {
			fmt.Printf("    #%-5d %-12s %-10s %s\n", a.Number, a.State, a.Severity, a.Advisory)
		}
	}

	fmt.Printf("\ntotal: %d alerts across %d repositories\n", totalAlerts, len(repos))
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

// repository query types

type repositoryNode struct {
	NameWithOwner string
	IsPrivate     bool
}

func queryRepositories(ctx context.Context, client *graphql.Client) ([]repositoryNode, error) {
	var query struct {
		Viewer struct {
			Repositories struct {
				Nodes    []repositoryNode
				PageInfo struct {
					EndCursor   string
					HasNextPage bool
				}
			} `graphql:"repositories(first: 100, orderBy: {field: UPDATED_AT, direction: DESC})"`
		}
	}

	if err := client.Query(ctx, &query, nil); err != nil {
		return nil, err
	}

	return query.Viewer.Repositories.Nodes, nil
}

// vulnerability alert query types

type alertSummary struct {
	Number   int
	State    string
	Severity string
	Advisory string
}

func queryVulnerabilityAlerts(ctx context.Context, client *graphql.Client, repository string) ([]alertSummary, error) {
	owner, name, ok := strings.Cut(repository, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repository format: %s", repository)
	}

	var alerts []alertSummary
	var after *graphql.String

	for {
		var query struct {
			Repository struct {
				VulnerabilityAlerts struct {
					Nodes []struct {
						Number                int
						State                 string
						SecurityVulnerability struct {
							Severity string
							Advisory struct {
								GHSAID  string `graphql:"ghsaId"`
								Summary string
							}
						}
					}
					PageInfo struct {
						EndCursor   string
						HasNextPage bool
					}
				} `graphql:"vulnerabilityAlerts(first: 50, after: $after)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables := map[string]any{
			"owner": graphql.String(owner),
			"name":  graphql.String(name),
			"after": after,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, err
		}

		for _, n := range query.Repository.VulnerabilityAlerts.Nodes {
			label := n.SecurityVulnerability.Advisory.GHSAID
			if n.SecurityVulnerability.Advisory.Summary != "" {
				label = n.SecurityVulnerability.Advisory.Summary
			}

			alerts = append(alerts, alertSummary{
				Number:   n.Number,
				State:    n.State,
				Severity: n.SecurityVulnerability.Severity,
				Advisory: label,
			})
		}

		if !query.Repository.VulnerabilityAlerts.PageInfo.HasNextPage {
			break
		}

		after = graphql.NewString(graphql.String(query.Repository.VulnerabilityAlerts.PageInfo.EndCursor))
	}

	return alerts, nil
}
