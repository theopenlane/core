package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	ent "github.com/theopenlane/core/internal/ent/generated"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
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
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMissingParent)
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
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrMissingOrganizationID)
}

func TestGetOrgOwnerIDWithUserType(t *testing.T) {
	orgID, err := getOrgOwnerID(context.Background(), pkgobjects.File{
		CorrelatedObjectType: "user",
	})

	assert.NoError(t, err)
	assert.Empty(t, orgID)
}

func TestGetOrgOwnerIDUsesAuthContext(t *testing.T) {
	user := &auth.Caller{
		OrganizationID: "01HYQZ5YTVJ0P2R2HF7N3W3MQZ",
	}

	ctx := auth.WithCaller(context.Background(), user)
	orgID, err := getOrgOwnerID(ctx, pkgobjects.File{
		CorrelatedObjectType: "program",
	})

	assert.NoError(t, err)
	assert.Equal(t, user.OrganizationID, orgID)
}

func TestTxHelpersReturnClients(t *testing.T) {
	client := &ent.Client{}
	ctx := ent.NewContext(context.Background(), client)

	assert.Equal(t, client, txClientFromContext(ctx))
	assert.Nil(t, txFileClientFromContext(ctx))

	fileClient := &ent.FileClient{}
	client.File = fileClient

	assert.Equal(t, fileClient, txFileClientFromContext(ent.NewContext(context.Background(), client)))
}

func TestAddFilePermissionsNoFiles(t *testing.T) {
	ctx := context.Background()
	updated, err := AddFilePermissions(ctx)
	assert.NoError(t, err)
	assert.Equal(t, ctx, updated)
}

func TestTxClientFromContextEmpty(t *testing.T) {
	assert.Nil(t, txClientFromContext(context.Background()))
	assert.Nil(t, txFileClientFromContext(context.Background()))
}
