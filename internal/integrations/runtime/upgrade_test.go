package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/integrations/types"
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

func TestUpgradeExclusions(t *testing.T) {
	t.Parallel()

	retired := types.NewCredentialRef[retiredCredential]()
	current := types.Replacing(types.NewCredentialRef[upgradeCredential](), retired, nil)
	undeclared := types.NewCredentialSlotID("undeclared")

	def := types.Definition{CredentialRegistrations: []types.CredentialRegistration{{Ref: current}}}

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

func TestConformCredential(t *testing.T) {
	t.Parallel()

	fillToken := func(_ context.Context, _ types.InstallationRequest, v *upgradeCredential) error {
		v.Token = "filled"

		return nil
	}

	noop := func(context.Context, types.InstallationRequest, *upgradeCredential) error { return nil }

	tests := []struct {
		name    string
		slot    types.CredentialSlot
		payload string
		want    string
		wantErr error
	}{
		{
			name:    "valid payload is unchanged",
			slot:    types.NewCredentialRef[upgradeCredential](),
			payload: `{"token":"t","region":"eu"}`,
			want:    `{"token":"t","region":"eu"}`,
		},
		{
			name:    "undeclared key is stripped",
			slot:    types.NewCredentialRef[upgradeCredential](),
			payload: `{"token":"t","region":"eu","legacy":1}`,
			want:    `{"region":"eu","token":"t"}`,
		},
		{
			name:    "missing required with default is filled",
			slot:    types.NewCredentialRef[upgradeCredential](),
			payload: `{"token":"t"}`,
			want:    `{"region":"us","token":"t"}`,
		},
		{
			name:    "missing required without default and no backfill is invalid",
			slot:    types.NewCredentialRef[upgradeCredential](),
			payload: `{"region":"eu"}`,
			wantErr: ErrCredentialInvalid,
		},
		{
			name:    "missing required without default is filled by the declared backfill",
			slot:    types.NewCredentialRef[upgradeCredential]().Backfilled(fillToken),
			payload: `{"region":"eu"}`,
			want:    `{"token":"filled","region":"eu"}`,
		},
		{
			name:    "backfill that leaves the payload invalid is rejected",
			slot:    types.NewCredentialRef[upgradeCredential]().Backfilled(noop),
			payload: `{"region":"eu"}`,
			wantErr: ErrCredentialInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := conformCredential(t.Context(), tc.slot, types.InstallationRequest{}, json.RawMessage(tc.payload))

			if tc.wantErr != nil {
				assert.Assert(t, errors.Is(err, tc.wantErr), "got %v", err)

				return
			}

			assert.NilError(t, err)
			assert.Equal(t, string(got), tc.want)
		})
	}
}
