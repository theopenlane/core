//go:build test

package integrations

import (
	"encoding/json"
	"time"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// DefinitionID is the stable identifier for the shared test integration definition
var DefinitionID = types.NewDefinitionRef("def_01K0TESTDEF0000000000000001")

var (
	// HealthOp is the inline health check used as every connection's validation operation
	healthSchema, HealthOp = providerkit.OperationSchema[healthCheck]()
	// RepoSyncOp is the async client-resolving operation
	repoSyncSchema, RepoSyncOp = providerkit.OperationSchema[repoSync]()
	// ValidatedOp is the inline operation with a required config field
	validatedSchema, ValidatedOp = providerkit.OperationSchema[validatedRun]()
	// RecurringOp is the healthy idle loop
	recurringSchema, RecurringOp = providerkit.OperationSchema[recurringCycle]()
	// ExhaustingOp is the always-failing loop
	exhaustingSchema, ExhaustingOp = providerkit.OperationSchema[exhaustingCycle]()
	// UnresolvableOp is the client-resolving loop seeded without a credential
	unresolvableSchema, UnresolvableOp = providerkit.OperationSchema[unresolvableCycle]()

	// TokenCredential is the credential slot the test client is built from
	tokenSchema, TokenCredential = providerkit.CredentialSchema[tokenCred]()
	// OAuthCredential is the auth-managed slot filled by the OAuth fixture
	_, OAuthCredential = providerkit.CredentialSchema[oauthTokenCred]()
	// ServiceAccountCredential is the strict-schema slot used by config flows
	serviceAccountSchema, ServiceAccountCredential = providerkit.CredentialSchema[serviceAccountCred]()

	// testClient builds from the token credential
	testClient = types.NewClientRef[*Client]()

	// WebhookAlertCreated is the webhook event contract
	WebhookAlertCreated = types.NewWebhookEventRef[webhookAlertEnvelope]("alert.created")
)

const (
	// ModeRecurring seeds a healthy idle loop
	ModeRecurring = "recurring"
	// ModeExhausting seeds a loop whose every cycle fails
	ModeExhausting = "exhausting"
	// ModeUnresolvable seeds a client-resolving loop without a credential
	ModeUnresolvable = "unresolvable"
)

const (
	// FailProjectID fails the health check when set as the service-account project id
	FailProjectID = "fail-project"
	// FailToken fails the health check when set as the token value
	FailToken = "fail"
)

const (
	recurringInterval        = time.Hour
	exhaustingInterval       = time.Millisecond
	exhaustingMaxErrorStreak = 3
)

type healthCheck struct{}

type repoSync struct{}

// validatedRun is the config for the inline operation with a required field
type validatedRun struct {
	// Target is the required target field
	Target string `json:"target" jsonschema:"required"`
}

type recurringCycle struct{}

type exhaustingCycle struct{}

type unresolvableCycle struct{}

// tokenCred is the credential material the test client is built from
type tokenCred struct {
	// Token is the API token
	Token string `json:"token"`
}

// oauthTokenCred is the credential material minted by the OAuth fixture
type oauthTokenCred struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"access_token"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refresh_token,omitempty"`
}

// serviceAccountCred is the strict credential material used by config flows
type serviceAccountCred struct {
	// ProjectID is the required project identifier
	ProjectID string `json:"projectId" jsonschema:"required"`
	// ServiceAccountEmail is the required service account email
	ServiceAccountEmail string `json:"serviceAccountEmail" jsonschema:"required"`
}

// UserInput is the installation-scoped user input for the test definition
type UserInput struct {
	// Mode selects which recurring loop operation is active
	Mode string `json:"mode,omitempty" jsonschema:"title=Scheduling Mode"`
	// FilterExpr is a free-form filter expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression"`
}

type webhookAlertEnvelope struct{}

// ModeInput returns the installation user input selecting one scheduling mode
func ModeInput(mode string) json.RawMessage {
	raw, err := json.Marshal(UserInput{Mode: mode})
	if err != nil {
		panic(err)
	}

	return raw
}

// TokenCredentialSet builds the token credential payload
func TokenCredentialSet(token string) types.CredentialSet {
	raw, err := json.Marshal(tokenCred{Token: token})
	if err != nil {
		panic(err)
	}

	return types.CredentialSet{Data: raw}
}

// ServiceAccountCredentialSet builds the strict credential payload
func ServiceAccountCredentialSet(projectID, email string) types.CredentialSet {
	raw, err := json.Marshal(serviceAccountCred{ProjectID: projectID, ServiceAccountEmail: email})
	if err != nil {
		panic(err)
	}

	return types.CredentialSet{Data: raw}
}
