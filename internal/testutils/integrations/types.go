//go:build test

package integrations

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// DefinitionID is the id of the shared test integration definition
var DefinitionID = types.NewDefinitionRef("def_01K0TESTDEF0000000000000001")

var (
	// RepoSyncOp is the async client-resolving operation
	RepoSyncOp = types.OperationPayloadOf[repoSync]().Handles(repoSyncHandler)
	// ValidatedOp is the inline operation with a required config field
	ValidatedOp = types.OperationPayloadOf[validatedRun]().HandlesRequest(validatedHandler).Policy(types.ExecutionPolicy{Inline: true})
	// RecurringOp is the healthy idle loop
	RecurringOp = types.OperationRefOf[recurringCycle]().
			HandlesRequest(idleCycle[recurringCycle]).
			Policy(types.ExecutionPolicy{Reconcile: true}).
			Schedule(&gala.Schedule{MinInterval: recurringInterval})
	// ExhaustingOp is the always-failing loop
	ExhaustingOp = types.OperationRefOf[exhaustingCycle]().
			HandlesRequest(failingCycle).
			Policy(types.ExecutionPolicy{Reconcile: true}).
			Schedule(&gala.Schedule{MinInterval: exhaustingInterval, MaxErrorStreak: exhaustingMaxErrorStreak})
	// UnresolvableOp is the client-resolving loop seeded without a credential
	UnresolvableOp = types.OperationRefOf[unresolvableCycle]().
			Handles(idleClientCycle).
			Policy(types.ExecutionPolicy{Reconcile: true}).
			Schedule(&gala.Schedule{MinInterval: recurringInterval})

	// LegacyToken is the unregistered legacy token connection
	LegacyToken = types.ConnectionOf[legacyTokenCred]()
	// Token is the token connection the test client is built from, taking over the legacy connection's payloads
	Token = types.ConnectionOf[tokenCred]().Replacing(LegacyToken).Upgraded(upgradeLegacyToken)
	// OAuth is the auth-managed connection filled by the OAuth fixture
	OAuth = types.ConnectionOf[oauthTokenCred]()
	// ServiceAccount is the strict-schema service account connection
	ServiceAccount = types.ConnectionOf[serviceAccountCred]().Upgraded(func(_ context.Context, req types.InstallationRequest, _ string, stored json.RawMessage) (serviceAccountCred, error) {
		c, err := jsonx.Decode[serviceAccountCred](stored)
		if err != nil {
			return serviceAccountCred{}, err
		}

		if c.ServiceAccountEmail == "" {
			c.ServiceAccountEmail = req.Integration.ID + "@backfilled.example.com"
		}

		return c, nil
	})

	// installation is the installation metadata layout of the shared test definition
	installation = types.InstallationOf[testMetadata]()

	// WebhookAlertCreated is the webhook event contract
	WebhookAlertCreated = types.NewWebhookEventRef[webhookAlertEnvelope]("alert.created")
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

// testMetadata is the installation metadata of the shared test definition
type testMetadata struct{}

// validatedRun is the config for the inline operation with a required field
type validatedRun struct {
	// Target is the required target field
	Target string `json:"target" jsonschema:"required"`
}

type recurringCycle struct {
	types.OperationSettings
}

type exhaustingCycle struct {
	types.OperationSettings
}

type unresolvableCycle struct {
	types.OperationSettings
}

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

type webhookAlertEnvelope struct{}

// ModeOperationConfig returns the per-operation input disabling every reconcile loop except the one selected by mode
func ModeOperationConfig(mode string) map[string]json.RawMessage {
	loops := map[string]string{
		ModeRecurring:    RecurringOp.Name(),
		ModeExhausting:   ExhaustingOp.Name(),
		ModeUnresolvable: UnresolvableOp.Name(),
	}

	config := make(map[string]json.RawMessage, len(loops))

	for loopMode, operation := range loops {
		config[operation] = lo.Must(json.Marshal(types.OperationSettings{Disable: loopMode != mode}))
	}

	return config
}

// legacyTokenCred is the token shape stored under the retired slot
type legacyTokenCred struct {
	// AccessToken is the legacy field name for the token
	AccessToken string `json:"accessToken"`
}

// upgradeLegacyToken maps a payload stored under the legacy token slot onto the current token shape
func upgradeLegacyToken(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) (tokenCred, error) {
	switch from {
	case LegacyToken.Connection().Credential.Name:
		legacy, err := jsonx.Decode[legacyTokenCred](stored)
		if err != nil {
			return tokenCred{}, err
		}

		return tokenCred{Token: legacy.AccessToken}, nil
	default:
		return jsonx.Decode[tokenCred](stored)
	}
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
