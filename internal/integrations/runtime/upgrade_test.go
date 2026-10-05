package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
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

func TestConformDocuments(t *testing.T) {
	t.Parallel()

	definition := types.NewDefinitionRef("test-def")
	retiredOp := types.OperationRefOf[retiredOperationConfigCfg]()
	currentOp := types.OperationRefOf[upgradeOperationConfigCfg]().
		Replacing(retiredOp).
		Upgraded(func(_ context.Context, _ types.InstallationRequest, from string, stored json.RawMessage) (upgradeOperationConfigCfg, error) {
			switch from {
			case retiredOp.Name():
				old, err := jsonx.Decode[retiredOperationConfigCfg](stored)
				if err != nil {
					return upgradeOperationConfigCfg{}, err
				}

				return upgradeOperationConfigCfg{Region: old.Zone}, nil
			default:
				return jsonx.Decode[upgradeOperationConfigCfg](stored)
			}
		})

	operationDef := types.Definition{Operations: []types.OperationRegistration{currentOp.Registration(definition, types.OperationRegistration{})}}

	retiredSlot := types.NewCredentialRef[retiredCredential]("retiredCredential")
	currentSlot := types.NewCredentialRef[upgradeCredential]("upgradeCredential").Replacing(retiredSlot).Upgraded(upgradeRetiredCredential(retiredSlot))

	credentialDef := types.Definition{CredentialRegistrations: []types.CredentialRegistration{currentSlot.Registration(types.CredentialRegistration{})}}

	userInput := types.UserInputRefOf[upgradeUserInput]()
	userInputDef := types.Definition{UserInput: userInput.Registration()}

	hooked := types.UserInputRefOf[upgradeUserInput]().Upgraded(func(context.Context, types.InstallationRequest, string, json.RawMessage) (upgradeUserInput, error) {
		return upgradeUserInput{}, ErrUpgradeHookCalled
	})
	hookedDef := types.Definition{UserInput: hooked.Registration()}

	tests := []struct {
		name    string
		kind    documentKind
		stored  map[string]json.RawMessage
		want    map[string]string
		wantErr error
	}{
		{
			name:   "declared operation is conformed keeping the uniform settings",
			kind:   operationInputKind(operationDef),
			stored: map[string]json.RawMessage{currentOp.Name(): json.RawMessage(`{"disable":true,"region":"eu","legacy":1}`)},
			want:   map[string]string{currentOp.Name(): `{"disable":true,"region":"eu"}`},
		},
		{
			name:   "retired operation name is upgraded onto its replacement",
			kind:   operationInputKind(operationDef),
			stored: map[string]json.RawMessage{retiredOp.Name(): json.RawMessage(`{"filterExpr":"payload.active","zone":"eu"}`)},
			want:   map[string]string{currentOp.Name(): `{"filterExpr":"payload.active","region":"eu"}`},
		},
		{
			name: "retired document does not overwrite an existing replacement document",
			kind: operationInputKind(operationDef),
			stored: map[string]json.RawMessage{
				retiredOp.Name(): json.RawMessage(`{"zone":"eu"}`),
				currentOp.Name(): json.RawMessage(`{"region":"us"}`),
			},
			want: map[string]string{currentOp.Name(): `{"region":"us"}`},
		},
		{
			name:   "undeclared operation is left untouched",
			kind:   operationInputKind(operationDef),
			stored: map[string]json.RawMessage{"unknown": json.RawMessage(`{"zone":"eu"}`)},
			want:   map[string]string{"unknown": `{"zone":"eu"}`},
		},
		{
			name:    "upgrade that still fails validation is rejected",
			kind:    operationInputKind(operationDef),
			stored:  map[string]json.RawMessage{retiredOp.Name(): json.RawMessage(`{"zone":""}`)},
			wantErr: types.ErrOperationConfigInvalid,
		},
		{
			name: "a retired document is dropped unconformed when its replacement is already stored",
			kind: operationInputKind(operationDef),
			stored: map[string]json.RawMessage{
				retiredOp.Name(): json.RawMessage(`{"zone":""}`),
				currentOp.Name(): json.RawMessage(`{"region":"us"}`),
			},
			want: map[string]string{currentOp.Name(): `{"region":"us"}`},
		},
		{
			name:   "retired credential slot is upgraded onto its replacement",
			kind:   credentialKind(credentialDef),
			stored: map[string]json.RawMessage{retiredSlot.String(): json.RawMessage(`{"accessToken":"t"}`)},
			want:   map[string]string{currentSlot.String(): `{"token":"t","region":""}`},
		},
		{
			name: "retired credential slot is dropped unconformed when its replacement is already stored",
			kind: credentialKind(credentialDef),
			stored: map[string]json.RawMessage{
				retiredSlot.String(): json.RawMessage(`{"accessToken":""}`),
				currentSlot.String(): json.RawMessage(`{"token":"t","region":"eu"}`),
			},
			want: map[string]string{currentSlot.String(): `{"token":"t","region":"eu"}`},
		},
		{
			name:   "undeclared credential slot is left untouched",
			kind:   credentialKind(credentialDef),
			stored: map[string]json.RawMessage{"unknown": json.RawMessage(`{"accessToken":"t"}`)},
			want:   map[string]string{"unknown": `{"accessToken":"t"}`},
		},
		{
			name:    "credential that still fails validation is rejected with the credential sentinel",
			kind:    credentialKind(credentialDef),
			stored:  map[string]json.RawMessage{retiredSlot.String(): json.RawMessage(`{"accessToken":""}`)},
			wantErr: ErrCredentialInvalid,
		},
		{
			name:   "user input stored under an empty layout is persisted under the current name",
			kind:   userInputKind(userInputDef),
			stored: map[string]json.RawMessage{"": json.RawMessage(`{"region":"eu"}`)},
			want:   map[string]string{userInput.Name(): `{"region":"eu"}`},
		},
		{
			name:   "user input stored under the current layout is conformed in place",
			kind:   userInputKind(userInputDef),
			stored: map[string]json.RawMessage{userInput.Name(): json.RawMessage(`{"region":"eu","legacy":1}`)},
			want:   map[string]string{userInput.Name(): `{"region":"eu"}`},
		},
		{
			name:    "user input that fails validation is rejected with the user input sentinel",
			kind:    userInputKind(userInputDef),
			stored:  map[string]json.RawMessage{"": json.RawMessage(`{"region":""}`)},
			wantErr: ErrUserInputInvalid,
		},
		{
			name: "absent document never visits the hook",
			kind: userInputKind(hookedDef),
			want: map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, failed := conformDocuments(t.Context(), types.InstallationRequest{}, tc.kind, tc.stored)
			documents := func(docs map[string]json.RawMessage) map[string]string {
				return lo.MapValues(docs, func(doc json.RawMessage, _ string) string { return string(doc) })
			}

			if tc.wantErr != nil {
				err := joinFailures(failed)
				assert.Assert(t, errors.Is(err, tc.wantErr), "got %v", err)
				assert.DeepEqual(t, documents(got), documents(tc.stored))

				return
			}

			assert.Equal(t, len(failed), 0, "failed: %v", failed)
			assert.DeepEqual(t, documents(got), tc.want)
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
		{
			name:   "a present empty document still passes through the upgrade hook",
			schema: jsonx.SchemaFrom[retiredCredential](),
			upgrade: func(context.Context, types.InstallationRequest, string, json.RawMessage) (json.RawMessage, error) {
				return nil, ErrUpgradeHookCalled
			},
			sentinel: ErrCredentialInvalid,
			wantErr:  ErrUpgradeHookCalled,
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

func TestLegacyDocuments(t *testing.T) {
	t.Parallel()

	definition := types.NewDefinitionRef("legacy-def")
	userInput := types.UserInputRefOf[upgradeUserInput]()
	sectioned := types.NewOperationRef[upgradeOperationConfigCfg]("DirectorySync")
	flat := types.NewOperationRef[retiredOperationConfigCfg]("AssetSync")
	payload := types.NewOperationPayload[upgradeUserInput]("Notify")

	def := types.Definition{
		UserInput: userInput.Registration(),
		Operations: []types.OperationRegistration{
			sectioned.Registration(definition, types.OperationRegistration{}),
			flat.Registration(definition, types.OperationRegistration{}),
			payload.Registration(definition, types.OperationRegistration{}),
		},
	}

	legacy := json.RawMessage(`{"region":"eu","directorySync":{"region":"us","disable":true}}`)

	t.Run("main's client config seeds user input and each stored operation by its camelCase section, else the whole document", func(t *testing.T) {
		t.Parallel()

		gotInput, gotConfig := legacyDocuments(&ent.Integration{Config: openapi.IntegrationConfig{ClientConfig: legacy}}, def)

		assert.Equal(t, gotInput.Layout, userInput.Name())
		assert.Equal(t, string(gotInput.Data), string(legacy))
		assert.DeepEqual(t, lo.MapValues(gotConfig.Operations, func(doc json.RawMessage, _ string) string { return string(doc) }), map[string]string{
			sectioned.Name(): `{"disable":true,"region":"us"}`,
			flat.Name():      string(legacy),
		})
	})

	t.Run("an installation with stored user input keeps its documents", func(t *testing.T) {
		t.Parallel()

		stored := types.IntegrationUserInput{Layout: userInput.Name(), Data: json.RawMessage(`{"region":"ca"}`)}

		gotInput, gotConfig := legacyDocuments(&ent.Integration{UserInput: stored, Config: openapi.IntegrationConfig{ClientConfig: legacy}}, def)

		assert.DeepEqual(t, gotInput, stored)
		assert.Equal(t, len(gotConfig.Operations), 0)
	})
}

func TestRetiredHealth(t *testing.T) {
	t.Parallel()

	definition := types.NewDefinitionRef("health-def")
	retired := types.NewOperationRef[retiredOperationConfigCfg]("old")
	current := types.NewOperationRef[upgradeOperationConfigCfg]("current").Replacing(retired)

	def := types.Definition{Operations: []types.OperationRegistration{current.Registration(definition, types.OperationRegistration{})}}

	assert.DeepEqual(t, retiredHealth(map[string]string{"old": "failing", "gone": "undeclared"}, def), map[string]string{"current": "failing"})
	assert.Assert(t, retiredHealth(map[string]string{"gone": "undeclared"}, def) == nil)
}
