//go:build test

package integrations

import (
	"context"
	"encoding/json"
	"path"
	"testing/fstest"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// VersionedRegistry registers builders on a registry whose definitions carry a version minted now, so a registry created later is newer than one created earlier
func VersionedRegistry(builders ...registry.Builder) (*registry.Registry, error) {
	plain := registry.New()
	if err := plain.RegisterAll(builders...); err != nil {
		return nil, err
	}

	snapshots := fstest.MapFS{}

	for _, def := range plain.Definitions() {
		hash, err := registry.SurfaceHash(def)
		if err != nil {
			return nil, err
		}

		data, err := json.Marshal(registry.Snapshot{Hash: hash, Version: ulid.Make().String()})
		if err != nil {
			return nil, err
		}

		snapshots[path.Join("surfaces", def.ID+".json")] = &fstest.MapFile{Data: data}
	}

	reg := registry.New(registry.WithSnapshots(snapshots))

	return reg, reg.RegisterAll(builders...)
}

// backfilledFilter is the default filter for version 3 user input
const backfilledFilter = "*"

// tokenV1 is the version 1 token credential shape
type tokenV1 struct {
	// AccessToken is the version 1 field name for the token
	AccessToken string `json:"accessToken"`
}

// tokenV2 is the version 2 token credential shape
type tokenV2 struct {
	// Token is the API token
	Token string `json:"token"`
}

// tokenV3 is the version 3 token credential shape
type tokenV3 struct {
	// Token is the API token
	Token string `json:"token"`
	// Region is the required region
	Region string `json:"region" jsonschema:"required"`
}

var (
	// TokenV1 is the version 1 token connection
	TokenV1 = types.ConnectionOf[tokenV1]()
	// TokenV2 is the version 2 token connection
	TokenV2 = types.ConnectionOf[tokenV2]().Replacing(TokenV1).Upgraded(upgradeTokenV2)
	// TokenV4 is the version 4 token connection
	TokenV4 = types.ConnectionOf[tokenV3]()
	// TokenV3 is the version 3 token connection
	TokenV3 = TokenV4.Replacing(TokenV2).Replacing(TokenV1).Upgraded(upgradeTokenV3)
)

// upgradeTokenV2 maps a version 1 payload onto the version 2 token shape
func upgradeTokenV2(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) (tokenV2, error) {
	switch from {
	case TokenV1.Connection().Credential.Name:
		old, err := jsonx.Decode[tokenV1](stored)
		if err != nil {
			return tokenV2{}, err
		}

		return tokenV2{Token: old.AccessToken}, nil
	default:
		return jsonx.Decode[tokenV2](stored)
	}
}

// upgradeTokenV3 maps version 1 and 2 payloads onto the version 3 token shape and fills the region from the installation
func upgradeTokenV3(_ context.Context, req types.InstallationRequest, from string, stored json.RawMessage) (tokenV3, error) {
	var (
		current tokenV3
		err     error
	)

	switch from {
	case TokenV1.Connection().Credential.Name:
		var old tokenV1

		old, err = jsonx.Decode[tokenV1](stored)
		current = tokenV3{Token: old.AccessToken}
	case TokenV2.Connection().Credential.Name:
		var old tokenV2

		old, err = jsonx.Decode[tokenV2](stored)
		current = tokenV3{Token: old.Token}
	default:
		current, err = jsonx.Decode[tokenV3](stored)
	}

	if err != nil {
		return tokenV3{}, err
	}

	if current.Region == "" {
		current.Region = req.Integration.ID
	}

	return current, nil
}

// TokenV1Set returns the version 1 token credential payload
func TokenV1Set(token string) types.CredentialSet {
	return credentialSet(tokenV1{AccessToken: token})
}

// TokenV3Set returns the version 3 token credential payload
func TokenV3Set(token, region string) types.CredentialSet {
	return credentialSet(tokenV3{Token: token, Region: region})
}

// userInputV1 is the version 1 user input layout
type userInputV1 struct {
	// FilterExpr is a free-form filter expression
	FilterExpr string `json:"filterExpr,omitempty"`
}

// userInputV2 is the version 2 user input layout
type userInputV2 struct {
	// Mode is the recurring loop operation selector
	Mode string `json:"mode,omitempty"`
	// FilterExpr is a free-form filter expression
	FilterExpr string `json:"filterExpr,omitempty"`
}

// userInputV3 is the version 3 user input layout
type userInputV3 struct {
	// Mode is the recurring loop operation selector
	Mode string `json:"mode,omitempty"`
	// Filter is the required filter expression
	Filter string `json:"filter" jsonschema:"required"`
}

var (
	// UserInputV1 is the version 1 user input layout
	UserInputV1 = types.UserInputRefOf[userInputV1]()
	// UserInputV2 is the version 2 user input layout, taking over version 1 documents
	UserInputV2 = types.UserInputRefOf[userInputV2]().Upgraded(upgradeUserInputV2)
	// UserInputV4 is the version 4 user input layout, the version 3 layout with its upgrade removed
	UserInputV4 = types.UserInputRefOf[userInputV3]()
	// UserInputV3 is the version 3 user input layout, taking over version 1 and 2 documents
	UserInputV3 = UserInputV4.Upgraded(upgradeUserInputV3)
)

// upgradeUserInputV2 maps a version 1 document onto the version 2 user input layout
func upgradeUserInputV2(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) (userInputV2, error) {
	switch from {
	case UserInputV1.Name():
		old, err := jsonx.Decode[userInputV1](stored)
		if err != nil {
			return userInputV2{}, err
		}

		return userInputV2{FilterExpr: old.FilterExpr}, nil
	default:
		return jsonx.Decode[userInputV2](stored)
	}
}

// upgradeUserInputV3 maps version 1 and 2 documents onto the version 3 user input layout and defaults the filter
func upgradeUserInputV3(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) (userInputV3, error) {
	var (
		current userInputV3
		err     error
	)

	switch from {
	case UserInputV1.Name():
		var old userInputV1

		old, err = jsonx.Decode[userInputV1](stored)
		current = userInputV3{Filter: old.FilterExpr}
	case UserInputV2.Name():
		var old userInputV2

		old, err = jsonx.Decode[userInputV2](stored)
		current = userInputV3{Mode: old.Mode, Filter: old.FilterExpr}
	default:
		current, err = jsonx.Decode[userInputV3](stored)
	}

	if err != nil {
		return userInputV3{}, err
	}

	if current.Filter == "" {
		current.Filter = backfilledFilter
	}

	return current, nil
}

// versionMetadata is the version 3 installation metadata
type versionMetadata struct {
	// ExternalID is the installation's identifier
	ExternalID string `json:"externalId"`
	// Region is the resolved version 3 token credential's region
	Region string `json:"region"`
}

// InstallationIdentity returns the installation identity
func (m versionMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{ExternalID: m.ExternalID}
}

// versionMetadataV3 is the version 3 installation metadata layout
var versionMetadataV3 = types.InstallationOf[versionMetadata]()

// syncCfg is the stored input layout of the reconcile operation shared by versions 1 and 2
type syncCfg struct {
	types.OperationSettings
	// Pattern is the record pattern the version 1 and 2 operation filters on
	Pattern string `json:"pattern,omitempty"`
}

// syncCfgV3 is the stored input layout of the version 3 reconcile operation
type syncCfgV3 struct {
	types.OperationSettings
	// Filter is the record filter the version 3 operation applies
	Filter string `json:"filter,omitempty"`
}

var (
	// SyncOp is the reconcile operation used by versions 1 and 2
	SyncOp = types.OperationRefOf[syncCfg]().HandlesRequest(idleCycle[syncCfg]).Policy(types.ExecutionPolicy{Inline: true})
	// SyncOpV4 is the version 4 reconcile operation, the version 3 operation with its replacement removed
	SyncOpV4 = types.OperationRefOf[syncCfgV3]().HandlesRequest(syncHandlerV3).Policy(types.ExecutionPolicy{Inline: true})
	// SyncOpV3 is the version 3 reconcile operation, taking over the stored input, runs, and health of SyncOp
	SyncOpV3 = SyncOpV4.Replacing(SyncOp).Upgraded(upgradeSyncCfgV3)
)

// upgradeSyncCfgV3 maps a document stored under the version 1 and 2 operation onto the version 3 config layout
func upgradeSyncCfgV3(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) (syncCfgV3, error) {
	switch from {
	case SyncOp.Name():
		old, err := jsonx.Decode[syncCfg](stored)
		if err != nil {
			return syncCfgV3{}, err
		}

		return syncCfgV3{Filter: old.Pattern}, nil
	default:
		return jsonx.Decode[syncCfgV3](stored)
	}
}

// SyncConfigV3 is the channel receiving each version 3 reconcile handler's resolved config
var SyncConfigV3 = make(chan json.RawMessage, 4)

// syncHandlerV3 records the operation's resolved config
func syncHandlerV3(_ context.Context, req types.OperationRequest, _ syncCfgV3) (json.RawMessage, error) {
	SyncConfigV3 <- jsonx.CloneRawMessage(req.Config)

	return nil, nil
}

var (
	// WebhookV1V2 is the webhook contract name used by versions 1 and 2
	WebhookV1V2 = types.NewWebhookRef("version-events.v1")
	// WebhookV4 is the version 4 webhook contract
	WebhookV4 = types.NewWebhookRef("version-events.v2")
	// WebhookV3 is the version 3 webhook contract
	WebhookV3 = WebhookV4.Replacing(WebhookV1V2)
)

// BuilderV1 returns the version 1 shared test definition
func BuilderV1() registry.Builder {
	return sharedVersionBuilder(UserInputV1, TokenV1, func(c tokenV1) string { return c.AccessToken })
}

// BuilderV2 returns the version 2 shared test definition
func BuilderV2() registry.Builder {
	return sharedVersionBuilder(UserInputV2, TokenV2, func(c tokenV2) string { return c.Token })
}

// sharedVersionBuilder returns a shared test definition builder for one version
func sharedVersionBuilder[In, Cred any](input types.UserInputRef[In], token types.ConnectionRef[Cred], tokenOf func(Cred) string) registry.Builder {
	return func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			UserInput:    input.Registration(),
			Installation: types.InstallationOf[versionMetadata]().Registration(),
			Connections: []types.Connector{
				token.
					Name("Test Token").
					Provides(tokenClient(tokenOf)).
					Verified(func(_ context.Context, req types.ConnectionRequest[Cred], _ *Client) (versionMetadata, error) {
						return versionMetadata{ExternalID: req.Integration.ID}, nil
					}).
					Disconnects("", nil),
			},
			Operations: []types.OperationRegistration{
				SyncOp.Registration(),
			},
			Webhooks: []types.WebhookRegistration{
				WebhookV1V2.Registration(types.WebhookRegistration{}),
			},
		}, nil
	}
}

// BuilderV3 returns the version 3 shared test definition
func BuilderV3() registry.Builder {
	return latestVersionBuilder(UserInputV3, TokenV3, SyncOpV3, WebhookV3)
}

// BuilderV4 returns the version 4 shared test definition
func BuilderV4() registry.Builder {
	return latestVersionBuilder(UserInputV4, TokenV4, SyncOpV4, WebhookV4)
}

// latestVersionBuilder returns a shared test definition builder for the latest version
func latestVersionBuilder(input types.UserInputRef[userInputV3], token types.ConnectionRef[tokenV3], operation types.OperationRef[syncCfgV3], webhook types.WebhookRef) registry.Builder {
	return func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			UserInput:    input.Registration(),
			Installation: versionMetadataV3.Registration(),
			Connections: []types.Connector{
				token.
					Name("Test Token").
					Provides(tokenClient(func(c tokenV3) string { return c.Token })).
					Verified(func(_ context.Context, req types.ConnectionRequest[tokenV3], _ *Client) (versionMetadata, error) {
						return versionMetadata{ExternalID: req.Integration.ID, Region: req.Credential.Region}, nil
					}).
					Disconnects("", nil),
			},
			Operations: []types.OperationRegistration{
				operation.Registration(),
			},
			Webhooks: []types.WebhookRegistration{
				webhook.Registration(types.WebhookRegistration{}),
			},
		}, nil
	}
}
