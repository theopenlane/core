//go:build test

package integrations

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// DefinitionID is the id of the shared test integration definition
var DefinitionID = types.NewDefinitionRef("def_01K0TESTDEF0000000000000001")

var (
	// RepoSyncOp is the async client-resolving operation
	RepoSyncOp = types.OperationRefOf[repoSync]().Handles(testClient, repoSyncHandler)
	// ValidatedOp is the inline operation with a required config field
	ValidatedOp = types.OperationRefOf[validatedRun]().HandlesRequest(validatedHandler)
	// RecurringOp is the healthy idle loop
	RecurringOp = types.OperationRefOf[recurringCycle]().HandlesRequest(idleCycle[recurringCycle])
	// ExhaustingOp is the always-failing loop
	ExhaustingOp = types.OperationRefOf[exhaustingCycle]().HandlesRequest(failingCycle)
	// UnresolvableOp is the client-resolving loop seeded without a credential
	UnresolvableOp = types.OperationRefOf[unresolvableCycle]().Handles(testClient, idleClientCycle)

	// LegacyTokenCredential is the unregistered legacy token credential slot
	LegacyTokenCredential = types.CredentialRefOf[legacyTokenCred]()
	// TokenCredential is the token slot the test client is built from
	TokenCredential = types.CredentialRefOf[tokenCred]().Replacing(LegacyTokenCredential, func(l legacyTokenCred) tokenCred { return tokenCred{Token: l.AccessToken} })
	// OAuthCredential is the auth-managed slot filled by the OAuth fixture
	OAuthCredential = types.CredentialRefOf[oauthTokenCred]()
	// ServiceAccountCredential is the strict-schema service account credential slot
	ServiceAccountCredential = types.CredentialRefOf[serviceAccountCred]().Backfilled(func(_ context.Context, req types.InstallationRequest, c *serviceAccountCred) error {
		if c.ServiceAccountEmail == "" {
			c.ServiceAccountEmail = req.Integration.ID + "@backfilled.example.com"
		}

		return nil
	})

	// testClient is the client built from the token credential
	testClient = types.ClientRefOf[*Client]().Using(TokenCredential)

	// oauthConnection is the OAuth connection mode
	oauthConnection = types.NewConnectionRef(OAuthCredential)
	// tokenConnection is the token connection mode
	tokenConnection = types.NewConnectionRef(TokenCredential)
	// serviceAccountConnection is the service account connection mode
	serviceAccountConnection = types.NewConnectionRef(ServiceAccountCredential)

	// WebhookAlertCreated is the webhook event contract
	WebhookAlertCreated = types.NewWebhookEventRef[webhookAlertEnvelope]("alert.created")

	// userInput is the shared test definition's user input layout
	userInput = types.NewUserInputRef[UserInput]("test-input")
)

const (
	// ModeRecurring is the healthy idle loop mode
	ModeRecurring = "recurring"
	// ModeExhausting is the always-failing loop mode
	ModeExhausting = "exhausting"
	// ModeUnresolvable is the client-resolving loop mode without a credential
	ModeUnresolvable = "unresolvable"
)

const (
	// FailProjectID is the project id value that fails the health check
	FailProjectID = "fail-project"
	// FailToken is the token value that fails the health check
	FailToken = "fail"
)

const (
	recurringInterval        = time.Hour
	exhaustingInterval       = time.Millisecond
	exhaustingMaxErrorStreak = 3
)

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
	// Mode is the recurring loop operation selector
	Mode string `json:"mode,omitempty" jsonschema:"title=Scheduling Mode"`
	// FilterExpr is a free-form filter expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression"`
}

type webhookAlertEnvelope struct{}

// ModeInput returns the installation user input selecting one scheduling mode
func ModeInput(mode string) json.RawMessage {
	return lo.Must(json.Marshal(UserInput{Mode: mode}))
}

// legacyTokenCred is the token shape stored under the retired slot
type legacyTokenCred struct {
	// AccessToken is the legacy field name for the token
	AccessToken string `json:"accessToken"`
}

// credentialSet returns a credential payload marshaled from value
func credentialSet(value any) types.CredentialSet {
	return types.CredentialSet{Data: lo.Must(json.Marshal(value))}
}

// TokenCredentialSet returns the token credential payload
func TokenCredentialSet(token string) types.CredentialSet {
	return credentialSet(tokenCred{Token: token})
}

// ServiceAccountCredentialSet returns the strict credential payload
func ServiceAccountCredentialSet(projectID, email string) types.CredentialSet {
	return credentialSet(serviceAccountCred{ProjectID: projectID, ServiceAccountEmail: email})
}
