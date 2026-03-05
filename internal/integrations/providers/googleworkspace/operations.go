package googleworkspace

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"

	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	googleWorkspaceHealthOp           types.OperationName = types.OperationHealthDefault
	googleWorkspaceDirectorySyncOp    types.OperationName = types.OperationDirectorySync
	googleWorkspaceDirectoryAlertType string              = "directory_account"
	googleWorkspaceDirectoryFields    string              = "nextPageToken,users(id,primaryEmail,name/fullName,name/givenName,name/familyName,thumbnailPhotoUrl,organizations/title,organizations/department,orgUnitPath,suspended,archived,isEnforcedIn2Sv,isEnrolledIn2Sv,lastLoginTime)"

	googleWorkspaceDirectoryDefaultPageSize = int64(200)
	googleWorkspaceHealthMaxResults         = int64(1)
)

type googleWorkspaceDirectoryConfig struct {
	// Customer overrides the default customer selector.
	Customer string `json:"customer"`
	// Domain scopes the user listing to a domain.
	Domain string `json:"domain"`
	// Query applies an Admin SDK user query filter.
	Query string `json:"query"`
}

// googleWorkspaceHealthDetails captures health check details from the Admin SDK users list call
type googleWorkspaceHealthDetails struct {
	// UserCount is the number of users returned by the health probe
	UserCount int `json:"userCount"`
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

// googleWorkspaceUserPayload wraps an Admin SDK user with an observed timestamp for envelope encoding.
type googleWorkspaceUserPayload struct {
	*admin.User
	ObservedAt string `json:"observedAt"`
}

// googleWorkspaceOperations returns the Google Workspace operations supported by this provider
func googleWorkspaceOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		operations.HealthOperation(googleWorkspaceHealthOp, "Call Google Admin SDK users.list to verify the workspace token", ClientGoogleWorkspaceAPI, runGoogleWorkspaceHealth),
		{
			Name:        googleWorkspaceDirectorySyncOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect Google Workspace directory users and emit directory account envelopes",
			Client:      ClientGoogleWorkspaceAPI,
			Run:         runGoogleWorkspaceDirectorySync,
			Ingest: []types.IngestContract{
				{
					Schema: types.MappingSchemaDirectoryAccount,
				},
			},
		},
	}
}

// runGoogleWorkspaceHealth verifies Admin SDK reachability with a small users.list probe.
func runGoogleWorkspaceHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	svc, err := resolveGoogleWorkspaceClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := svc.Users.List().
		Customer("my_customer").
		MaxResults(googleWorkspaceHealthMaxResults).
		Projection("basic").
		ViewType("admin_view").
		Fields(googleapi.Field("users(id),nextPageToken")).
		Do()
	if err != nil {
		return operations.OperationFailure("Google Workspace health check failed", err, nil)
	}

	return operations.OperationSuccess(
		fmt.Sprintf("Google Workspace directory API reachable (%d sample users)", len(resp.Users)),
		googleWorkspaceHealthDetails{UserCount: len(resp.Users)},
	), nil
}

// runGoogleWorkspaceDirectorySync collects directory users and emits directory account envelopes.
func runGoogleWorkspaceDirectorySync(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	svc, err := resolveGoogleWorkspaceClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	config, err := operations.Decode[googleWorkspaceDirectoryConfig](input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	customer := "my_customer"
	if config.Customer != "" {
		customer = config.Customer
	}

	fetch := func(ctx context.Context, pageToken string) (operations.PageResult[*admin.User], error) {
		call := svc.Users.List().
			MaxResults(googleWorkspaceDirectoryDefaultPageSize).
			Projection("full").
			ViewType("admin_view").
			Fields(googleapi.Field(googleWorkspaceDirectoryFields))

		if config.Domain != "" {
			call = call.Domain(config.Domain)
		} else {
			call = call.Customer(customer)
		}

		if config.Query != "" {
			call = call.Query(config.Query)
		}

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return operations.PageResult[*admin.User]{}, err
		}

		return operations.PageResult[*admin.User]{
			Items:     resp.Users,
			NextToken: resp.NextPageToken,
		}, nil
	}

	allUsers, err := operations.CollectAll(ctx, fetch, 0)
	if err != nil {
		return operations.OperationFailure("Directory users fetch failed", err, nil)
	}

	observedAt := time.Now().UTC().Format(time.RFC3339)
	envelopes := make([]types.AlertEnvelope, 0, len(allUsers))

	for _, u := range allUsers {
		payload, err := json.Marshal(googleWorkspaceUserPayload{User: u, ObservedAt: observedAt})
		if err != nil {
			return operations.OperationFailure("Directory user payload encoding failed", err, nil)
		}

		envelopes = append(envelopes, types.AlertEnvelope{
			AlertType: googleWorkspaceDirectoryAlertType,
			Resource:  googleWorkspaceDirectoryResource(u),
			Payload:   payload,
		})
	}

	return operations.OperationSuccess(fmt.Sprintf("Collected %d directory accounts", len(envelopes)), googleWorkspaceDirectorySyncDetails{
		UsersTotal:  len(allUsers),
		AlertsTotal: len(envelopes),
		Alerts:      envelopes,
	}), nil
}

// googleWorkspaceDirectoryResource selects the envelope resource identity for a user
func googleWorkspaceDirectoryResource(u *admin.User) string {
	if u.PrimaryEmail != "" {
		return u.PrimaryEmail
	}

	return u.Id
}
