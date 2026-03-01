package googleworkspace

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

const (
	googleWorkspaceHealthOp           types.OperationName = "health.default"
	googleWorkspaceDirectorySyncOp    types.OperationName = types.OperationDirectorySync
	googleWorkspaceDirectoryAlertType string              = "directory_account"

	googleWorkspaceDirectoryUsersEndpoint string = "https://admin.googleapis.com/admin/directory/v1/users"
	googleWorkspaceDirectoryFields        string = "nextPageToken,users(id,primaryEmail,name/fullName,name/givenName,name/familyName,thumbnailPhotoUrl,organizations/title,organizations/department,orgUnitPath,suspended,archived,isEnforcedIn2Sv,isEnrolledIn2Sv,lastLoginTime)"

	googleWorkspaceDirectoryDefaultPageSize = 200
)

type googleUserinfoResponse struct {
	// Sub is the subject identifier for the user
	Sub string `json:"sub"`
	// Email is the primary email address
	Email string `json:"email"`
	// Name is the display name for the user
	Name string `json:"name"`
}

type googleWorkspaceDirectoryUsersResponse struct {
	// Users is the page of returned directory users
	Users []map[string]any `json:"users"`
	// NextPageToken is the pagination token for the next page
	NextPageToken string `json:"nextPageToken"`
}

type googleWorkspaceDirectoryConfig struct {
	// Customer overrides the default customer selector.
	Customer string `json:"customer"`
	// Domain scopes the user listing to a domain.
	Domain string `json:"domain"`
	// Query applies an Admin SDK user query filter.
	Query string `json:"query"`
	// OrgUnitPath filters results by org unit path.
	OrgUnitPath string `json:"orgUnitPath"`
}

// googleWorkspaceHealthDetails captures Google userinfo health payload details
type googleWorkspaceHealthDetails struct {
	// Sub is the subject identifier for the user
	Sub string `json:"sub,omitempty"`
	// Email is the primary email address
	Email string `json:"email,omitempty"`
	// Name is the display name for the user
	Name string `json:"name,omitempty"`
}

// googleWorkspaceDirectorySyncDetails captures directory sync operation details
type googleWorkspaceDirectorySyncDetails struct {
	// UsersTotal is the number of users returned by Google APIs
	UsersTotal int `json:"users_total"`
	// AlertsTotal is the number of emitted ingest envelopes
	AlertsTotal int `json:"alerts_total"`
	// Alerts is the emitted ingest envelopes
	Alerts []types.AlertEnvelope `json:"alerts"`
}

// googleWorkspaceOperations returns the Google Workspace operations supported by this provider
func googleWorkspaceOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(googleWorkspaceHealthOp, "Call Google OAuth userinfo to verify the workspace token", ClientGoogleWorkspaceAPI,
			operations.HealthCheckRunner(operations.TokenTypeOAuth, "https://www.googleapis.com/oauth2/v3/userinfo", "Google userinfo failed",
				func(resp googleUserinfoResponse) (string, any) {
					return fmt.Sprintf("Google token valid for %s", resp.Email), googleWorkspaceHealthDetails{
						Sub:   resp.Sub,
						Email: resp.Email,
						Name:  resp.Name,
					}
				})),
		{
			Name:        googleWorkspaceDirectorySyncOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect Google Workspace directory users and emit directory account envelopes",
			Client:      ClientGoogleWorkspaceAPI,
			Run:         runGoogleWorkspaceDirectorySync,
		},
	}
}

// runGoogleWorkspaceUsers returns a small sample of directory users for posture checks.
func runGoogleWorkspaceUsers(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, token, err := auth.ClientAndToken(input, auth.OAuthTokenFromPayload)
	if err != nil {
		return types.OperationResult{}, err
	}

	params := url.Values{}
	params.Set("customer", "my_customer")
	params.Set("maxResults", fmt.Sprintf("%d", googleWorkspaceDirectoryDefaultPageSize))
	params.Set("projection", "full")
	params.Set("viewType", "admin_view")
	params.Set("fields", googleWorkspaceDirectoryFields)

	config, err := operations.Decode[googleWorkspaceDirectoryConfig](input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}
	if config.Customer != "" {
		params.Set("customer", config.Customer)
	}
	if config.Domain != "" {
		params.Del("customer")
		params.Set("domain", config.Domain)
	}
	if config.Query != "" {
		params.Set("query", config.Query)
	}
	if config.OrgUnitPath != "" {
		params.Set("orgUnitPath", config.OrgUnitPath)
	}
	observedAt := time.Now().UTC().Format(time.RFC3339)
	envelopes := make([]types.AlertEnvelope, 0)
	totalUsers := 0
	pageToken := ""

	for {
		if pageToken == "" {
			params.Del("pageToken")
		} else {
			params.Set("pageToken", pageToken)
		}

		endpoint := googleWorkspaceDirectoryUsersEndpoint + "?" + params.Encode()

		var resp googleWorkspaceDirectoryUsersResponse
		if err := auth.GetJSONWithClient(ctx, client, endpoint, token, nil, &resp); err != nil {
			return operations.OperationFailure("Directory users fetch failed", err, nil)
		}

		for _, user := range resp.Users {
			user["observedAt"] = observedAt

			payload, err := json.Marshal(user)
			if err != nil {
				return operations.OperationFailure("Directory user payload encoding failed", err, nil)
			}

			envelopes = append(envelopes, types.AlertEnvelope{
				AlertType: googleWorkspaceDirectoryAlertType,
				Resource:  googleWorkspaceDirectoryResource(user),
				Payload:   payload,
			})
			totalUsers++
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return operations.OperationSuccess(fmt.Sprintf("Collected %d directory accounts", len(envelopes)), googleWorkspaceDirectorySyncDetails{
		UsersTotal:  totalUsers,
		AlertsTotal: len(envelopes),
		Alerts:      envelopes,
	}), nil
}

// googleWorkspaceDirectoryResource selects the envelope resource identity for a user
func googleWorkspaceDirectoryResource(user map[string]any) string {
	if primaryEmail, ok := user["primaryEmail"].(string); ok {
		return primaryEmail
	}
	if id, ok := user["id"].(string); ok {
		return id
	}

	return ""
}
