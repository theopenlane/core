package entbridge

import (
	"context"
	"encoding/json"
	"errors"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// ErrNoDB indicates the runtime services do not expose an ent client
var ErrNoDB = errors.New("integrations: runtime services do not expose a database client")

// errServiceUnavailable is returned by the DB-only services for every non-database capability
var errServiceUnavailable = errors.New("integrations: runtime service unavailable outside the integration runtime")

// DBProvider is implemented by runtime services that expose the ent client to first-party definitions
type DBProvider interface {
	// DB returns the runtime-managed ent client
	DB() *ent.Client
}

// DB returns the ent client behind the runtime services, or nil when the services do not expose one
func DB(services types.RuntimeServices) *ent.Client {
	if provider, ok := services.(DBProvider); ok {
		return provider.DB()
	}

	return nil
}

// RequireDB returns the ent client behind the runtime services or ErrNoDB
func RequireDB(services types.RuntimeServices) (*ent.Client, error) {
	if db := DB(services); db != nil {
		return db, nil
	}

	return nil, ErrNoDB
}

// Installation converts an ent integration record into the ent-free installation view, nil for a nil record
func Installation(record *ent.Integration) *types.Installation {
	if record == nil {
		return nil
	}

	return &types.Installation{
		ID:                   record.ID,
		OwnerID:              record.OwnerID,
		Name:                 record.Name,
		Description:          record.Description,
		DefinitionID:         record.DefinitionID,
		DefinitionVersion:    record.DefinitionVersion,
		Family:               record.Family,
		Status:               record.Status,
		SystemOwned:          record.SystemOwned,
		PrimaryDirectory:     record.PrimaryDirectory,
		CampaignEmail:        record.CampaignEmail,
		Config:               record.Config,
		UserInput:            record.UserInput,
		OperationConfig:      record.OperationConfig,
		InstallationMetadata: record.InstallationMetadata,
		ProviderState:        record.ProviderState,
		Metadata:             record.Metadata,
		ExpiresAt:            record.ExpiresAt,
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.UpdatedAt,
	}
}

// Webhook converts an ent webhook record into the ent-free webhook endpoint view, nil for a nil record
func Webhook(record *ent.IntegrationWebhook) *types.WebhookEndpoint {
	if record == nil {
		return nil
	}

	return &types.WebhookEndpoint{
		ID:              record.ID,
		OwnerID:         record.OwnerID,
		IntegrationID:   record.IntegrationID,
		Name:            record.Name,
		Status:          record.Status,
		EndpointID:      record.EndpointID,
		EndpointURL:     record.EndpointURL,
		SecretToken:     record.SecretToken,
		AllowedEvents:   record.AllowedEvents,
		ExternalEventID: record.ExternalEventID,
		LastDeliveryID:  record.LastDeliveryID,
		Metadata:        record.Metadata,
	}
}

// Services returns runtime services that expose only the ent client, for callers that run definition
// code outside the integration runtime (tests, one-off handlers); every other capability errors
func Services(db *ent.Client) types.RuntimeServices {
	return dbOnlyServices{db: db}
}

// dbOnlyServices satisfies types.RuntimeServices with only a database client
type dbOnlyServices struct {
	// db is the ent client exposed through DBProvider
	db *ent.Client
}

// DB returns the wrapped ent client
func (s dbOnlyServices) DB() *ent.Client { return s.db }

// Gala returns nil; no event runtime is available
func (dbOnlyServices) Gala() *gala.Gala { return nil }

// Definitions returns an empty lookup
func (dbOnlyServices) Definitions() types.DefinitionLookup { return noDefinitions{} }

// ExecuteRuntimeOperation is unavailable outside the runtime
func (dbOnlyServices) ExecuteRuntimeOperation(context.Context, string, string, json.RawMessage) (json.RawMessage, error) {
	return nil, errServiceUnavailable
}

// Dispatch is unavailable outside the runtime
func (dbOnlyServices) Dispatch(context.Context, types.DispatchRequest) (types.DispatchResult, error) {
	return types.DispatchResult{}, errServiceUnavailable
}

// ReapExpiredInstallation is unavailable outside the runtime
func (dbOnlyServices) ReapExpiredInstallation(context.Context, string) (bool, error) {
	return false, errServiceUnavailable
}

// ResolveOwnerIntegration is unavailable outside the runtime
func (dbOnlyServices) ResolveOwnerIntegration(context.Context, string, string, ...func(*types.Installation) bool) (string, error) {
	return "", errServiceUnavailable
}

// noDefinitions is a definition lookup with no entries
type noDefinitions struct{}

// Definition reports no definition for any id
func (noDefinitions) Definition(string) (types.Definition, bool) { return types.Definition{}, false }

// recordContextKey keys the ent record the runtime is executing an operation against
type recordContextKey struct{}

// WithRecord returns a context carrying the ent record the runtime executes an operation against, so first-party
// definitions that hand the record to runtime machinery share the runtime's instance rather than a reloaded copy
func WithRecord(ctx context.Context, record *ent.Integration) context.Context {
	if record == nil {
		return ctx
	}

	return context.WithValue(ctx, recordContextKey{}, record)
}

// Record returns the ent integration record behind an installation view, nil for a nil view; it prefers the
// runtime's instance carried by WithRecord so in-place updates (such as ingest health tracking) stay visible to
// the runtime, and falls back to loading the row
func Record(ctx context.Context, db *ent.Client, installation *types.Installation) (*ent.Integration, error) {
	if installation == nil {
		return nil, nil
	}

	if record, ok := ctx.Value(recordContextKey{}).(*ent.Integration); ok && record.ID == installation.ID {
		return record, nil
	}

	return db.Integration.Get(ctx, installation.ID)
}
