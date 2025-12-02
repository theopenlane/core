package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	pkgobjects "github.com/theopenlane/core/pkg/objects"
	ent "github.com/theopenlane/ent/generated"
	"github.com/theopenlane/iam/auth"
)

func TestAddFilePermissionsMissingParent(t *testing.T) {
	ctx := context.Background()

	file := pkgobjects.File{
		ID:     "file-1",
		Parent: pkgobjects.ParentObject{},
	}

	ctx = pkgobjects.WriteFilesToContext(ctx, pkgobjects.Files{
		"default": {file},
	})

	_, err := AddFilePermissions(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrMissingParent)
}

func TestAddFilePermissionsAvatarMissingOrg(t *testing.T) {
	ctx := context.Background()

	file := pkgobjects.File{
		ID: "file-2",
		Parent: pkgobjects.ParentObject{
			ID:   "parent-id",
			Type: "evidence",
		},
		FieldName: "avatarFile",
	}

	ctx = pkgobjects.WriteFilesToContext(ctx, pkgobjects.Files{
		"avatarFile": {file},
	})

	_, err := AddFilePermissions(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, auth.ErrNoAuthUser)
}

func TestGetOrgOwnerIDWithUserType(t *testing.T) {
	orgID, err := getOrgOwnerID(context.Background(), pkgobjects.File{
		CorrelatedObjectType: "user",
	})

	require.NoError(t, err)
	require.Empty(t, orgID)
}

func TestGetOrgOwnerIDUsesAuthContext(t *testing.T) {
	user := &auth.AuthenticatedUser{
		OrganizationID: "01HYQZ5YTVJ0P2R2HF7N3W3MQZ",
	}

	ctx := auth.WithAuthenticatedUser(context.Background(), user)
	orgID, err := getOrgOwnerID(ctx, pkgobjects.File{
		CorrelatedObjectType: "program",
	})

	require.NoError(t, err)
	require.Equal(t, user.OrganizationID, orgID)
}

func TestTxHelpersReturnClients(t *testing.T) {
	client := &ent.Client{}
	ctx := ent.NewContext(context.Background(), client)

	require.Equal(t, client, txClientFromContext(ctx))
	require.Nil(t, txFileClientFromContext(ctx))

	fileClient := &ent.FileClient{}
	client.File = fileClient

	require.Equal(t, fileClient, txFileClientFromContext(ent.NewContext(context.Background(), client)))
}

func TestAddFilePermissionsNoFiles(t *testing.T) {
	ctx := context.Background()
	updated, err := AddFilePermissions(ctx)
	require.NoError(t, err)
	require.Equal(t, ctx, updated)
}

func TestTxClientFromContextEmpty(t *testing.T) {
	require.Nil(t, txClientFromContext(context.Background()))
	require.Nil(t, txFileClientFromContext(context.Background()))
}
