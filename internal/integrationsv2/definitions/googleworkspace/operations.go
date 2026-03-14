package googleworkspace

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	googleWorkspaceDirectoryDefaultPageSize = int64(200)
	googleWorkspaceHealthMaxResults         = int64(1)
	googleWorkspaceDirectoryFields          = "nextPageToken,users(id,primaryEmail,name/fullName,orgUnitPath,suspended,isEnforcedIn2Sv,isEnrolledIn2Sv,lastLoginTime)"
)

type googleWorkspaceHealthDetails struct {
	UserCount int `json:"userCount"`
}

type googleWorkspaceDirectoryUser struct {
	ID           string `json:"id"`
	PrimaryEmail string `json:"primaryEmail"`
	FullName     string `json:"fullName"`
	OrgUnitPath  string `json:"orgUnitPath"`
	Suspended    bool   `json:"suspended"`
}

type googleWorkspaceDirectorySyncDetails struct {
	UsersTotal int                            `json:"users_total"`
	Samples    []googleWorkspaceDirectoryUser `json:"samples"`
}

// buildWorkspaceClient builds the Google Workspace Admin SDK client for one installation
func buildWorkspaceClient(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	token := req.Credential.OAuthAccessToken
	if token == "" {
		return nil, ErrOAuthTokenMissing
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	svc, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("googleworkspace: admin service build failed: %w", err)
	}

	return svc, nil
}

// runHealthOperation calls users.list with maxResults=1 to verify workspace access
func runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	svc, ok := client.(*admin.Service)
	if !ok {
		return nil, ErrClientType
	}

	resp, err := svc.Users.List().
		Customer("my_customer").
		MaxResults(googleWorkspaceHealthMaxResults).
		Projection("basic").
		ViewType("admin_view").
		Fields(googleapi.Field("users(id),nextPageToken")).
		Do()
	if err != nil {
		return nil, fmt.Errorf("googleworkspace: health check failed: %w", err)
	}

	return jsonx.ToRawMessage(googleWorkspaceHealthDetails{UserCount: len(resp.Users)})
}

type googleWorkspaceDirectoryConfig struct {
	Customer string `json:"customer"`
	Domain   string `json:"domain"`
	Query    string `json:"query"`
}

// runDirectorySyncOperation collects Google Workspace directory users
func runDirectorySyncOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, config json.RawMessage) (json.RawMessage, error) {
	svc, ok := client.(*admin.Service)
	if !ok {
		return nil, ErrClientType
	}

	var cfg googleWorkspaceDirectoryConfig
	if err := jsonx.UnmarshalIfPresent(config, &cfg); err != nil {
		return nil, err
	}

	customer := "my_customer"
	if cfg.Customer != "" {
		customer = cfg.Customer
	}

	var allUsers []*admin.User
	pageToken := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		call := svc.Users.List().
			MaxResults(googleWorkspaceDirectoryDefaultPageSize).
			Projection("full").
			ViewType("admin_view").
			Fields(googleapi.Field(googleWorkspaceDirectoryFields))

		if cfg.Domain != "" {
			call = call.Domain(cfg.Domain)
		} else {
			call = call.Customer(customer)
		}

		if cfg.Query != "" {
			call = call.Query(cfg.Query)
		}

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("googleworkspace: directory users fetch failed: %w", err)
		}

		allUsers = append(allUsers, resp.Users...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	sampleSize := min(len(allUsers), 10)
	samples := make([]googleWorkspaceDirectoryUser, 0, sampleSize)
	for _, u := range allUsers[:sampleSize] {
		if u == nil {
			continue
		}

		fullName := ""
		if u.Name != nil {
			fullName = u.Name.FullName
		}

		samples = append(samples, googleWorkspaceDirectoryUser{
			ID:           u.Id,
			PrimaryEmail: u.PrimaryEmail,
			FullName:     fullName,
			OrgUnitPath:  u.OrgUnitPath,
			Suspended:    u.Suspended,
		})
	}

	return jsonx.ToRawMessage(googleWorkspaceDirectorySyncDetails{
		UsersTotal: len(allUsers),
		Samples:    samples,
	})
}
