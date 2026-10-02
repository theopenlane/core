package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// upgradeCredential is the credential type with a required non-empty token and a defaulted region
type upgradeCredential struct {
	// Token is the required non-empty token
	Token string `json:"token" jsonschema:"required,minLength=1"`
	// Region is the required region defaulted when absent
	Region string `json:"region" jsonschema:"required,default=us"`
}

// retiredCredential is the shape an earlier definition version stored the token under
type retiredCredential struct {
	// AccessToken is the retired field name for the token
	AccessToken string `json:"accessToken"`
}

// upgradeUserInput is the user input type with a required non-empty region
type upgradeUserInput struct {
	// Region is the required non-empty region
	Region string `json:"region" jsonschema:"required,minLength=1"`
}

// upgradeCredentialRegion is a credential type whose region is required by presence only
type upgradeCredentialRegion struct {
	// Token is the token field
	Token string `json:"token"`
	// Region is required by presence only
	Region string `json:"region" jsonschema:"required"`
}

// upgradeOperationConfigCfg is the current operation config type with a required non-empty region
type upgradeOperationConfigCfg struct {
	types.OperationSettings
	// Region is the required non-empty region
	Region string `json:"region" jsonschema:"required,minLength=1"`
}

// retiredOperationConfigCfg is the shape an earlier definition version stored the region under
type retiredOperationConfigCfg struct {
	types.OperationSettings
	// Zone is the retired field name for the region
	Zone string `json:"zone"`
}

// upgradeRetiredCredential maps a payload stored under the retired slot onto the current credential layout
func upgradeRetiredCredential(retired types.CredentialRef[retiredCredential]) func(context.Context, types.InstallationRequest, string, json.RawMessage) (upgradeCredential, error) {
	return func(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) (upgradeCredential, error) {
		switch from {
		case retired.ID().String():
			old, err := jsonx.Decode[retiredCredential](stored)
			if err != nil {
				return upgradeCredential{}, err
			}

			return upgradeCredential{Token: old.AccessToken}, nil
		default:
			return jsonx.Decode[upgradeCredential](stored)
		}
	}
}

func TestUpgradeOperationDocuments(t *testing.T) {
	t.Parallel()

	definition := types.NewDefinitionRef("test-def")
	retired := types.OperationRefOf[retiredOperationConfigCfg]()
	current := types.OperationRefOf[upgradeOperationConfigCfg]().
		Replacing(retired).
		Upgraded(func(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) (upgradeOperationConfigCfg, error) {
			switch from {
			case retired.Name():
				old, err := jsonx.Decode[retiredOperationConfigCfg](stored)
				if err != nil {
					return upgradeOperationConfigCfg{}, err
				}

				return upgradeOperationConfigCfg{Region: old.Zone}, nil
			default:
				return jsonx.Decode[upgradeOperationConfigCfg](stored)
			}
		})

	registration := current.Registration(definition, types.OperationRegistration{})

	def := types.Definition{Operations: []types.OperationRegistration{registration}}

	tests := []struct {
		name    string
		stored  map[string]json.RawMessage
		want    map[string]string
		wantErr error
	}{
		{
			name:   "declared operation is conformed keeping the uniform settings",
			stored: map[string]json.RawMessage{current.Name(): json.RawMessage(`{"disable":true,"region":"eu","legacy":1}`)},
			want:   map[string]string{current.Name(): `{"disable":true,"region":"eu"}`},
		},
		{
			name:   "retired operation name is upgraded onto its replacement",
			stored: map[string]json.RawMessage{retired.Name(): json.RawMessage(`{"filterExpr":"payload.active","zone":"eu"}`)},
			want:   map[string]string{current.Name(): `{"filterExpr":"payload.active","region":"eu"}`},
		},
		{
			name: "retired document does not overwrite an existing replacement document",
			stored: map[string]json.RawMessage{
				retired.Name(): json.RawMessage(`{"zone":"eu"}`),
				current.Name(): json.RawMessage(`{"region":"us"}`),
			},
			want: map[string]string{current.Name(): `{"region":"us"}`},
		},
		{
			name:   "undeclared operation is left untouched",
			stored: map[string]json.RawMessage{"unknown": json.RawMessage(`{"zone":"eu"}`)},
			want:   map[string]string{"unknown": `{"zone":"eu"}`},
		},
		{
			name:    "upgrade that still fails validation is rejected",
			stored:  map[string]json.RawMessage{retired.Name(): json.RawMessage(`{"zone":""}`)},
			wantErr: types.ErrOperationConfigInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := upgradeOperationDocuments(t.Context(), types.InstallationRequest{}, def, tc.stored)

			if tc.wantErr != nil {
				assert.Assert(t, errors.Is(err, tc.wantErr), "got %v", err)

				return
			}

			assert.NilError(t, err)
			assert.DeepEqual(t, lo.MapValues(got, func(doc json.RawMessage, _ string) string { return string(doc) }), tc.want)
		})
	}
}

func TestUpgradeExclusions(t *testing.T) {
	t.Parallel()

	retired := types.NewCredentialRef[retiredCredential]("retiredCredential")
	current := types.NewCredentialRef[upgradeCredential]("upgradeCredential").Replacing(retired)
	undeclared := types.NewCredentialSlotID("undeclared")

	def := types.Definition{CredentialRegistrations: []types.CredentialRegistration{current.Registration(types.CredentialRegistration{
		Schema: jsonx.SchemaFrom[upgradeCredential](),
	})}}

	tests := []struct {
		name string
		skip []types.CredentialSlotID
		want []string
	}{
		{
			name: "no skip excludes nothing",
			want: []string{},
		},
		{
			name: "declared slot excludes itself and the slots it replaces",
			skip: []types.CredentialSlotID{current.ID()},
			want: []string{retired.String(), current.String()},
		},
		{
			name: "undeclared slot excludes only itself",
			skip: []types.CredentialSlotID{undeclared},
			want: []string{undeclared.String()},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := lo.Map(upgradeExclusions(def, tc.skip), func(slot types.CredentialSlotID, _ int) string {
				return slot.String()
			})

			assert.DeepEqual(t, got, tc.want)
		})
	}
}

func TestConformStored(t *testing.T) {
	t.Parallel()

	retired := types.NewCredentialRef[retiredCredential]("retiredCredential")
	credentialSchema := jsonx.SchemaFrom[upgradeCredential]()

	filling := types.NewCredentialRef[upgradeCredential]("upgradeCredential").Upgraded(func(_ context.Context, _ types.InstallationRequest, _ string, stored json.RawMessage) (upgradeCredential, error) {
		value, err := jsonx.Decode[upgradeCredential](stored)
		value.Token = "filled"

		return value, err
	}).Registration(types.CredentialRegistration{})

	idle := types.NewCredentialRef[upgradeCredential]("upgradeCredential").Upgraded(func(_ context.Context, _ types.InstallationRequest, _ string, stored json.RawMessage) (upgradeCredential, error) {
		return jsonx.Decode[upgradeCredential](stored)
	}).Registration(types.CredentialRegistration{})

	retagging := types.NewCredentialRef[upgradeCredential]("upgradeCredential").Upgraded(func(_ context.Context, _ types.InstallationRequest, _ string, stored json.RawMessage) (upgradeCredential, error) {
		value, err := jsonx.Decode[upgradeCredential](stored)
		value.Region = "override"

		return value, err
	}).Registration(types.CredentialRegistration{})

	renaming := types.NewCredentialRef[upgradeCredential]("upgradeCredential").Replacing(retired).Upgraded(upgradeRetiredCredential(retired)).Registration(types.CredentialRegistration{})

	regionSchema := jsonx.SchemaFrom[upgradeCredentialRegion]()
	regionFilling := types.NewCredentialRef[upgradeCredentialRegion]("upgradeCredentialRegion").Upgraded(func(_ context.Context, _ types.InstallationRequest, _ string, stored json.RawMessage) (upgradeCredentialRegion, error) {
		value, err := jsonx.Decode[upgradeCredentialRegion](stored)
		if value.Region == "" {
			value.Region = "resolved"
		}

		return value, err
	}).Registration(types.CredentialRegistration{})

	tests := []struct {
		name     string
		schema   json.RawMessage
		upgrade  types.UpgradeFunc
		from     string
		sentinel error
		payload  string
		want     string
		wantErr  error
	}{
		{
			name:     "valid payload is unchanged",
			schema:   credentialSchema,
			sentinel: ErrCredentialInvalid,
			payload:  `{"token":"t","region":"eu"}`,
			want:     `{"token":"t","region":"eu"}`,
		},
		{
			name:     "undeclared key is stripped",
			schema:   credentialSchema,
			sentinel: ErrCredentialInvalid,
			payload:  `{"token":"t","region":"eu","legacy":1}`,
			want:     `{"region":"eu","token":"t"}`,
		},
		{
			name:     "missing required with default is filled",
			schema:   credentialSchema,
			sentinel: ErrCredentialInvalid,
			payload:  `{"token":"t"}`,
			want:     `{"region":"us","token":"t"}`,
		},
		{
			name:     "missing required without default and no upgrade is invalid",
			schema:   credentialSchema,
			sentinel: ErrCredentialInvalid,
			payload:  `{"region":"eu"}`,
			wantErr:  ErrCredentialInvalid,
		},
		{
			name:     "missing required without default is filled by the declared upgrade",
			schema:   credentialSchema,
			upgrade:  filling.Upgrade,
			sentinel: ErrCredentialInvalid,
			payload:  `{"region":"eu"}`,
			want:     `{"token":"filled","region":"eu"}`,
		},
		{
			name:     "upgrade that leaves the payload invalid is rejected",
			schema:   credentialSchema,
			upgrade:  idle.Upgrade,
			sentinel: ErrCredentialInvalid,
			payload:  `{"region":"eu"}`,
			wantErr:  ErrCredentialInvalid,
		},
		{
			name:     "user input schema fails with the user input sentinel",
			schema:   jsonx.SchemaFrom[upgradeUserInput](),
			sentinel: ErrUserInputInvalid,
			payload:  `{"region":""}`,
			wantErr:  ErrUserInputInvalid,
		},
		{
			name:     "declared upgrade runs even when the stored payload already validates",
			schema:   credentialSchema,
			upgrade:  retagging.Upgrade,
			sentinel: ErrCredentialInvalid,
			payload:  `{"token":"t","region":"eu"}`,
			want:     `{"token":"t","region":"override"}`,
		},
		{
			name:     "payload stored under a retired slot is mapped by the declared upgrade",
			schema:   credentialSchema,
			upgrade:  renaming.Upgrade,
			from:     retired.ID().String(),
			sentinel: ErrCredentialInvalid,
			payload:  `{"accessToken":"t"}`,
			want:     `{"token":"t","region":""}`,
		},
		{
			name:     "a required field present but empty is filled by the declared upgrade",
			schema:   regionSchema,
			upgrade:  regionFilling.Upgrade,
			sentinel: ErrCredentialInvalid,
			payload:  `{"token":"t","region":""}`,
			want:     `{"token":"t","region":"resolved"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := conformStored(t.Context(), types.InstallationRequest{}, tc.schema, tc.upgrade, nil, tc.from, json.RawMessage(tc.payload), tc.sentinel)

			if tc.wantErr != nil {
				assert.Assert(t, errors.Is(err, tc.wantErr), "got %v", err)

				return
			}

			assert.NilError(t, err)
			assert.Equal(t, string(got), tc.want)
		})
	}
}
