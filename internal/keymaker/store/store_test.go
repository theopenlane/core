package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/keystore"
)

func TestEncodeTokens(t *testing.T) {
	t.Parallel()

	data, err := EncodeTokens("slack", tokensBundle{
		AccessToken:  "access",
		RefreshToken: "refresh",
		Attributes: map[string]string{
			keystore.ExpiresAtField: "2024-01-02T03:04:05Z",
		},
	})
	require.NoError(t, err)

	expected := map[string]string{
		"slack_" + keystore.AccessTokenField:  "access",
		"slack_" + keystore.RefreshTokenField: "refresh",
		"slack_" + keystore.ExpiresAtField:    "2024-01-02T03:04:05Z",
	}

	actual := make(map[string]string)
	require.NoError(t, json.Unmarshal(data, &actual))
	require.Equal(t, expected, actual)
}

func TestEntRepositoryMissingIdentifiers(t *testing.T) {
	t.Parallel()

	repo := &EntRepository{}

	_, err := repo.Integration(context.Background(), "", "")
	require.ErrorIs(t, err, ErrMissingIdentifiers)

	_, err = repo.Credentials(context.Background(), "", "")
	require.ErrorIs(t, err, ErrMissingIdentifiers)

	err = repo.SaveCredentials(context.Background(), CredentialRecord{})
	require.ErrorIs(t, err, ErrMissingIdentifiers)

	err = repo.DeleteCredentials(context.Background(), "", "")
	require.ErrorIs(t, err, ErrMissingIdentifiers)
}
