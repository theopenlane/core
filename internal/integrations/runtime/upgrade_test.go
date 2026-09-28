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

// retiredUserInput is the shape an earlier definition version stored the region under
type retiredUserInput struct {
	// Zone is the retired field name for the region
	Zone string `json:"zone"`
}

func TestUpgradeExclusions(t *testing.T) {
	t.Parallel()

	retired := types.NewCredentialRef[retiredCredential]("retiredCredential")
	current := types.NewCredentialRef[upgradeCredential]("upgradeCredential").Replacing(retired, nil)
	undeclared := types.NewCredentialSlotID("undeclared")

	def := types.Definition{CredentialRegistrations: []types.CredentialRegistration{{
		Ref:      current.ID(),
		Schema:   jsonx.SchemaFrom[upgradeCredential](),
		Replaces: current.Replaces(),
		Convert:  current.Convert,
	}}}

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

func TestConformPayload(t *testing.T) {
	t.Parallel()

	fillToken := func(_ context.Context, _ types.InstallationRequest, v *upgradeCredential) error {
		v.Token = "filled"

		return nil
	}

	noop := func(context.Context, types.InstallationRequest, *upgradeCredential) error { return nil }

	credentialSchema := jsonx.SchemaFrom[upgradeCredential]()
	filling := types.NewCredentialRef[upgradeCredential]("upgradeCredential").Backfilled(fillToken)
	idle := types.NewCredentialRef[upgradeCredential]("upgradeCredential").Backfilled(noop)

	tests := []struct {
		name     string
		schema   json.RawMessage
		backfill types.BackfillFunc
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
			name:     "missing required without default and no backfill is invalid",
			schema:   credentialSchema,
			sentinel: ErrCredentialInvalid,
			payload:  `{"region":"eu"}`,
			wantErr:  ErrCredentialInvalid,
		},
		{
			name:     "missing required without default is filled by the declared backfill",
			schema:   credentialSchema,
			backfill: filling.Backfill,
			sentinel: ErrCredentialInvalid,
			payload:  `{"region":"eu"}`,
			want:     `{"token":"filled","region":"eu"}`,
		},
		{
			name:     "backfill that leaves the payload invalid is rejected",
			schema:   credentialSchema,
			backfill: idle.Backfill,
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := conformPayload(t.Context(), types.InstallationRequest{}, tc.schema, tc.backfill, json.RawMessage(tc.payload), tc.sentinel)

			if tc.wantErr != nil {
				assert.Assert(t, errors.Is(err, tc.wantErr), "got %v", err)

				return
			}

			assert.NilError(t, err)
			assert.Equal(t, string(got), tc.want)
		})
	}
}

func TestConformUserInput(t *testing.T) {
	t.Parallel()

	retired := types.NewUserInputRef[retiredUserInput]("retiredUserInput")
	renamed := types.NewUserInputRef[upgradeUserInput]("upgradeUserInput").Replacing(retired, func(r retiredUserInput) upgradeUserInput {
		return upgradeUserInput{Region: r.Zone}
	})

	renamedInput := types.UserInputRegistration{
		Schema:   jsonx.SchemaFrom[upgradeUserInput](),
		Replaces: renamed.Replaces(),
		Convert:  renamed.Convert,
	}

	plainInput := types.UserInputRegistration{
		Schema: jsonx.SchemaFrom[upgradeUserInput](),
	}

	tests := []struct {
		name    string
		input   types.UserInputRegistration
		stored  string
		want    string
		wantErr error
	}{
		{
			name:   "valid stored input is conformed without conversion",
			input:  renamedInput,
			stored: `{"region":"eu","zone":"ignored"}`,
			want:   `{"region":"eu"}`,
		},
		{
			name:   "stored input in the retired layout is converted when it fails validation",
			input:  renamedInput,
			stored: `{"zone":"eu"}`,
			want:   `{"region":"eu"}`,
		},
		{
			name:    "invalid stored input with no replacements is rejected",
			input:   plainInput,
			stored:  `{"zone":"eu"}`,
			wantErr: ErrUserInputInvalid,
		},
		{
			name:    "conversion that still fails validation is rejected",
			input:   renamedInput,
			stored:  `{"zone":""}`,
			wantErr: ErrUserInputInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := conformUserInput(t.Context(), types.InstallationRequest{}, tc.input, json.RawMessage(tc.stored))

			if tc.wantErr != nil {
				assert.Assert(t, errors.Is(err, tc.wantErr), "got %v", err)

				return
			}

			assert.NilError(t, err)
			assert.Equal(t, string(got), tc.want)
		})
	}
}
