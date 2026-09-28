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
	// TokenV3 is the version 3 token credential slot, taking over payloads stored under TokenV2 or, chained, directly under TokenV1, backfilling the region from the installation id
	TokenV3 = types.CredentialRefOf[tokenV3]().
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
	raw, err := json.Marshal(tokenV1{AccessToken: token})
	if err != nil {
		panic(err)
	}

	return types.CredentialSet{Data: raw}
}

// TokenV2Set builds the version 2 token credential payload
func TokenV2Set(token string) types.CredentialSet {
	raw, err := json.Marshal(tokenV2{Token: token})
	if err != nil {
		panic(err)
	}

	return types.CredentialSet{Data: raw}
}

// TokenV3Set builds the version 3 token credential payload
func TokenV3Set(token, region string) types.CredentialSet {
	raw, err := json.Marshal(tokenV3{Token: token, Region: region})
	if err != nil {
		panic(err)
	}

	return types.CredentialSet{Data: raw}
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

// UserInputV3 is the version 3 user input layout, taking over the version 2 layout's stored filter expression as its required filter and backfilling it when still empty
var UserInputV3 = types.NewUserInputRef[userInputV3]("version-input.v3").
	Replacing(userInputV2Ref, func(o userInputV2) userInputV3 { return userInputV3{Mode: o.Mode, Filter: o.FilterExpr} }).
	Backfilled(func(_ context.Context, _ types.InstallationRequest, c *userInputV3) error {
		if c.Filter == "" {
			c.Filter = backfilledFilter
		}

		return nil
	})

// versionClient is the client the version fixtures build from their active token slot
type versionClient struct {
	// Token is the resolved API token
	Token string
}

// buildVersionClientV1 builds the client from the version 1 token credential
func buildVersionClientV1(_ context.Context, req types.ClientBuildRequest) (*versionClient, error) {
	cred, ok, err := TokenV1.Resolve(req.Credentials)
	if err != nil || !ok {
		return nil, ErrTokenMissing
	}

	return &versionClient{Token: cred.AccessToken}, nil
}

// buildVersionClientV2 builds the client from the version 2 token credential
func buildVersionClientV2(_ context.Context, req types.ClientBuildRequest) (*versionClient, error) {
	cred, ok, err := TokenV2.Resolve(req.Credentials)
	if err != nil || !ok {
		return nil, ErrTokenMissing
	}

	return &versionClient{Token: cred.Token}, nil
}

// buildVersionClientV3 builds the client from the version 3 token credential
func buildVersionClientV3(_ context.Context, req types.ClientBuildRequest) (*versionClient, error) {
	cred, ok, err := TokenV3.Resolve(req.Credentials)
	if err != nil || !ok {
		return nil, ErrTokenMissing
	}

	return &versionClient{Token: cred.Token}, nil
}

var (
	// versionClientV1 is the version 1 client, built from the version 1 token credential
	versionClientV1 = types.ClientRefOf[*versionClient]().Using(TokenV1)
	// versionClientV2 is the version 2 client, built from the version 2 token credential
	versionClientV2 = types.ClientRefOf[*versionClient]().Using(TokenV2)
	// versionClientV3 is the version 3 client, built from the version 3 token credential
	versionClientV3 = types.ClientRefOf[*versionClient]().Using(TokenV3)

	// tokenConnectionV1 is the version 1 connection mode selected by the version 1 token slot
	tokenConnectionV1 = types.NewConnectionRef(TokenV1)
	// tokenConnectionV2 is the version 2 connection mode selected by the version 2 token slot
	tokenConnectionV2 = types.NewConnectionRef(TokenV2)
	// tokenConnectionV3 is the version 3 connection mode selected by the version 3 token slot
	tokenConnectionV3 = types.NewConnectionRef(TokenV3)
)

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
	// SyncOpV3 is the reconcile operation name version 3 renames SyncOp to
	SyncOpV3 = types.OperationRefOf[syncCfgV3]().Replacing(SyncOp).HandlesRequest(syncHandlerV3)
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
	// WebhookV3 is the webhook contract version 3 renames WebhookV1V2 to
	WebhookV3 = types.NewWebhookRef("version-events.v2").Replacing(WebhookV1V2)
)

// BuilderV1 registers the version 1 shared test definition
func BuilderV1() registry.Builder {
	return func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			UserInput: userInputV1Ref.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				TokenV1.Registration(types.CredentialRegistration{
					Name: "Test Token",
				}),
			},
			Connections: []types.ConnectionRegistration{
				tokenConnectionV1.Registration(types.ConnectionRegistration{
					Name:       "Test Token",
					Disconnect: &types.DisconnectRegistration{},
				}),
			},
			HealthCheck: types.CredentialHealthCheck(healthHandler),
			Clients: []types.ClientRegistration{
				versionClientV1.Registration(buildVersionClientV1, types.ClientRegistration{
					Description: "Version 1 client built from the version 1 token credential.",
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

// BuilderV2 registers the version 2 shared test definition, taking over version 1's credential and reusing its operation and webhook contract
func BuilderV2() registry.Builder {
	return func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			UserInput: userInputV2Ref.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				TokenV2.Registration(types.CredentialRegistration{
					Name: "Test Token",
				}),
			},
			Connections: []types.ConnectionRegistration{
				tokenConnectionV2.Registration(types.ConnectionRegistration{
					Name:       "Test Token",
					Disconnect: &types.DisconnectRegistration{},
				}),
			},
			HealthCheck: types.CredentialHealthCheck(healthHandler),
			Clients: []types.ClientRegistration{
				versionClientV2.Registration(buildVersionClientV2, types.ClientRegistration{
					Description: "Version 2 client built from the version 2 token credential.",
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
	return func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			UserInput: UserInputV3.Registration(),
			CredentialRegistrations: []types.CredentialRegistration{
				TokenV3.Registration(types.CredentialRegistration{
					Name: "Test Token",
				}),
			},
			Connections: []types.ConnectionRegistration{
				tokenConnectionV3.Registration(types.ConnectionRegistration{
					Name:       "Test Token",
					Disconnect: &types.DisconnectRegistration{},
				}),
			},
			HealthCheck:  types.CredentialHealthCheck(healthHandler),
			Installation: versionMetadataV3.Registration(),
			Clients: []types.ClientRegistration{
				versionClientV3.Registration(buildVersionClientV3, types.ClientRegistration{
					Description: "Version 3 client built from the version 3 token credential.",
				}),
			},
			Operations: []types.OperationRegistration{
				SyncOpV3.Registration(DefinitionID, types.OperationRegistration{
					Policy:         types.ExecutionPolicy{Inline: true},
					ConfigResolver: syncConfigV3From,
				}),
			},
			Webhooks: []types.WebhookRegistration{
				WebhookV3.Registration(types.WebhookRegistration{}),
			},
		}, nil
	}
}
