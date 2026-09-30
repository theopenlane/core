//go:build test

package integrations

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// backfilledFilter is the default filter the version 3 user input backfills when the stored input carries none
const backfilledFilter = "*"

// tokenV1 is the version 1 token credential shape
type tokenV1 struct {
	// AccessToken is the version 1 field name for the token
	AccessToken string `json:"accessToken"`
}

// tokenV2 is the version 2 token credential shape, renaming the stored token field
type tokenV2 struct {
	// Token is the API token
	Token string `json:"token"`
}

// tokenV3 is the version 3 token credential shape, adding a required region backfilled from the installation id
type tokenV3 struct {
	// Token is the API token
	Token string `json:"token"`
	// Region is the required region, backfilled from the installation id when empty
	Region string `json:"region" jsonschema:"required"`
}

var (
	// TokenV1 is the version 1 token credential slot
	TokenV1 = types.CredentialRefOf[tokenV1]()
	// TokenV2 is the version 2 token credential slot, taking over payloads stored under TokenV1
	TokenV2 = types.CredentialRefOf[tokenV2]().Replacing(TokenV1, func(o tokenV1) tokenV2 { return tokenV2{Token: o.AccessToken} })
	// TokenV4 is the version 4 token credential slot, the version 3 slot with no replacement or backfill declared
	TokenV4 = types.CredentialRefOf[tokenV3]()
	// TokenV3 is the version 3 token credential slot, taking over payloads stored under TokenV2 or, chained, directly under TokenV1, backfilling the region from the installation id
	TokenV3 = TokenV4.
		Replacing(TokenV2, func(o tokenV2) tokenV3 { return tokenV3{Token: o.Token} }).
		Replacing(TokenV1, func(o tokenV1) tokenV3 { return tokenV3{Token: o.AccessToken} }).
		Backfilled(func(_ context.Context, req types.InstallationRequest, c *tokenV3) error {
			if c.Region == "" {
				c.Region = req.Integration.ID
			}

			return nil
		})
)

// TokenV1Set builds the version 1 token credential payload
func TokenV1Set(token string) types.CredentialSet {
	return credentialSet(tokenV1{AccessToken: token})
}

// TokenV3Set builds the version 3 token credential payload
func TokenV3Set(token, region string) types.CredentialSet {
	return credentialSet(tokenV3{Token: token, Region: region})
}

// userInputV1 is the version 1 user input layout, a single optional filter expression
type userInputV1 struct {
	// FilterExpr is a free-form filter expression
	FilterExpr string `json:"filterExpr,omitempty"`
}

// userInputV2 is the version 2 user input layout, adding a scheduling mode alongside the version 1 filter expression
type userInputV2 struct {
	// Mode selects which recurring loop operation is active
	Mode string `json:"mode,omitempty"`
	// FilterExpr is a free-form filter expression
	FilterExpr string `json:"filterExpr,omitempty"`
}

// userInputV3 is the version 3 user input layout, renaming the version 2 filter expression to a required filter
type userInputV3 struct {
	// Mode selects which recurring loop operation is active
	Mode string `json:"mode,omitempty"`
	// Filter is the required filter expression, renamed from the version 2 layout's filterExpr
	Filter string `json:"filter" jsonschema:"required"`
}

// userInputV1Ref names the version 1 user input layout
var userInputV1Ref = types.NewUserInputRef[userInputV1]("version-input.v1")

// userInputV2Ref names the version 2 user input layout so the version 3 layout can retire it explicitly
var userInputV2Ref = types.NewUserInputRef[userInputV2]("version-input.v2")

// userInputV4Ref is the version 4 user input layout, the version 3 layout with no replacement or backfill declared
var userInputV4Ref = types.NewUserInputRef[userInputV3]("version-input.v3")

// UserInputV3 is the version 3 user input layout, taking over the version 2 layout's stored filter expression as its required filter and backfilling it when still empty
var UserInputV3 = userInputV4Ref.
	Replacing(userInputV2Ref, func(o userInputV2) userInputV3 { return userInputV3{Mode: o.Mode, Filter: o.FilterExpr} }).
	Backfilled(func(_ context.Context, _ types.InstallationRequest, c *userInputV3) error {
		if c.Filter == "" {
			c.Filter = backfilledFilter
		}

		return nil
	})

// versionMetadata is the installation metadata the version 3 connection derives from the resolved token credential
type versionMetadata struct {
	// ExternalID is the installation's own identifier, echoed back as the normalized display identity
	ExternalID string `json:"externalId"`
	// Region is the resolved version 3 token credential's region
	Region string `json:"region"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m versionMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{ExternalID: m.ExternalID}
}

// versionMetadataV3 derives version 3 installation metadata from the bound version 3 token credential
var versionMetadataV3 = types.NewInstallationRef(func(_ context.Context, req types.InstallationRequest) (versionMetadata, bool, error) {
	cred, ok, err := TokenV3.Resolve(req.Credentials)
	if err != nil || !ok {
		return versionMetadata{}, false, nil
	}

	return versionMetadata{ExternalID: req.Integration.ID, Region: cred.Region}, true, nil
})

// syncCfg is the config for the reconcile operation shared by versions 1 and 2
type syncCfg struct{}

// syncCfgV3 is the resolved config for the version 3 reconcile operation, projected from the installation's version 3 user input
type syncCfgV3 struct {
	// Filter is the filter carried over from the installation's user input
	Filter string `json:"filter,omitempty"`
}

var (
	// SyncOp is the reconcile operation name used by versions 1 and 2
	SyncOp = types.OperationRefOf[syncCfg]().HandlesRequest(idleCycle[syncCfg])
	// SyncOpV4 is the version 4 reconcile operation, the version 3 operation with no replacement declared
	SyncOpV4 = types.OperationRefOf[syncCfgV3]().HandlesRequest(syncHandlerV3)
	// SyncOpV3 is the reconcile operation name version 3 renames SyncOp to
	SyncOpV3 = SyncOpV4.Replacing(SyncOp)
)

// SyncConfigV3 receives the resolved config each time the version 3 reconcile handler runs
var SyncConfigV3 = make(chan json.RawMessage, 4)

// syncHandlerV3 records the operation's resolved config and performs no work
func syncHandlerV3(_ context.Context, req types.OperationRequest, _ syncCfgV3) (json.RawMessage, error) {
	SyncConfigV3 <- jsonx.CloneRawMessage(req.Config)

	return nil, nil
}

// syncConfigV3From projects the version 3 reconcile config from the installation's flat version 3 user input
func syncConfigV3From(userInput json.RawMessage) json.RawMessage {
	var input userInputV3

	if err := jsonx.UnmarshalIfPresent(userInput, &input); err != nil {
		return nil
	}

	config, err := jsonx.ToRawMessage(syncCfgV3{Filter: input.Filter})
	if err != nil {
		return nil
	}

	return config
}

var (
	// WebhookV1V2 is the webhook contract name used by versions 1 and 2
	WebhookV1V2 = types.NewWebhookRef("version-events.v1")
	// WebhookV4 is the version 4 webhook contract, the version 3 contract with no replacement declared
	WebhookV4 = types.NewWebhookRef("version-events.v2")
	// WebhookV3 is the webhook contract version 3 renames WebhookV1V2 to
	WebhookV3 = WebhookV4.Replacing(WebhookV1V2)
)

// BuilderV1 registers the version 1 shared test definition
func BuilderV1() registry.Builder {
	return sharedVersionBuilder(userInputV1Ref, TokenV1, func(c tokenV1) string { return c.AccessToken })
}

// BuilderV2 registers the version 2 shared test definition, taking over version 1's credential and reusing its operation and webhook contract
func BuilderV2() registry.Builder {
	return sharedVersionBuilder(userInputV2Ref, TokenV2, func(c tokenV2) string { return c.Token })
}

// sharedVersionBuilder registers a shared test definition version over the given user input layout and token slot, whose client reads the token through tokenOf, sharing the version 1 operation and webhook contract
func sharedVersionBuilder[In, Cred any](input types.UserInputRef[In], token types.CredentialRef[Cred], tokenOf func(Cred) string) registry.Builder {
	return func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			UserInput: input.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				token.Registration(types.CredentialRegistration{
					Name: "Test Token",
				}),
			},
			Connections: []types.ConnectionRegistration{
				types.NewConnectionRef(token).Registration(types.ConnectionRegistration{
					Name:       "Test Token",
					Disconnect: &types.DisconnectRegistration{},
				}),
			},
			HealthCheck: types.CredentialHealthCheck(healthHandler),
			Clients: []types.ClientRegistration{
				types.ClientRefOf[*Client]().Using(token).Registration(tokenClient(token, tokenOf), types.ClientRegistration{
					Description: "Version client built from the token credential.",
				}),
			},
			Operations: []types.OperationRegistration{
				SyncOp.Registration(DefinitionID, types.OperationRegistration{
					Policy: types.ExecutionPolicy{Inline: true},
				}),
			},
			Webhooks: []types.WebhookRegistration{
				WebhookV1V2.Registration(types.WebhookRegistration{}),
			},
		}, nil
	}
}

// BuilderV3 registers the version 3 shared test definition, chain-taking over version 2's and version 1's credential, and renaming the shared user input layout, reconcile operation, and webhook contract
func BuilderV3() registry.Builder {
	return latestVersionBuilder(UserInputV3, TokenV3, SyncOpV3, WebhookV3)
}

// BuilderV4 registers the version 4 shared test definition, version 3's names and schemas with every replacement and backfill declaration removed
func BuilderV4() registry.Builder {
	return latestVersionBuilder(userInputV4Ref, TokenV4, SyncOpV4, WebhookV4)
}

// latestVersionBuilder registers a shared test definition over the version 3 user input layout, token slot, reconcile operation, and webhook contract, carrying whatever lifecycle declarations the given refs hold
func latestVersionBuilder(input types.UserInputRef[userInputV3], token types.CredentialRef[tokenV3], operation types.OperationRef[syncCfgV3], webhook types.WebhookRef) registry.Builder {
	return func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			UserInput: input.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				token.Registration(types.CredentialRegistration{
					Name: "Test Token",
				}),
			},
			Connections: []types.ConnectionRegistration{
				types.NewConnectionRef(token).Registration(types.ConnectionRegistration{
					Name:       "Test Token",
					Disconnect: &types.DisconnectRegistration{},
				}),
			},
			HealthCheck:  types.CredentialHealthCheck(healthHandler),
			Installation: versionMetadataV3.Registration(),
			Clients: []types.ClientRegistration{
				types.ClientRefOf[*Client]().Using(token).Registration(tokenClient(token, func(c tokenV3) string { return c.Token }), types.ClientRegistration{
					Description: "Version 3 client built from the version 3 token credential.",
				}),
			},
			Operations: []types.OperationRegistration{
				operation.Registration(DefinitionID, types.OperationRegistration{
					Policy:         types.ExecutionPolicy{Inline: true},
					ConfigResolver: syncConfigV3From,
				}),
			},
			Webhooks: []types.WebhookRegistration{
				webhook.Registration(types.WebhookRegistration{}),
			},
		}, nil
	}
}
