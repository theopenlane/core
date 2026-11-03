package store

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
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
	assert.Check(t, err != nil)
	assert.Check(t, is.ErrorIs(err, ErrMissingParent))
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
	assert.Check(t, err != nil)
	assert.Check(t, is.ErrorContains(err, "organization ID is required for file upload"))
}

func TestGetOrgOwnerIDWithUserType(t *testing.T) {
	orgID, err := getOrgOwnerID(context.Background(), pkgobjects.File{
		CorrelatedObjectType: "user",
	})

	assert.Check(t, err)
	assert.Check(t, is.Equal(orgID, ""))
}

func TestGetOrgOwnerIDUsesAuthContext(t *testing.T) {
	user := &auth.Caller{
		OrganizationID: "01HYQZ5YTVJ0P2R2HF7N3W3MQZ",
	}

	ctx := auth.WithCaller(context.Background(), user)
	orgID, err := getOrgOwnerID(ctx, pkgobjects.File{
		CorrelatedObjectType: "program",
	})

	assert.Check(t, err)
	assert.Check(t, is.Equal(user.OrganizationID, orgID))
}

func TestTxHelpersReturnClients(t *testing.T) {
	client := &ent.Client{}
	ctx := ent.NewContext(context.Background(), client)

	assert.Check(t, is.Equal(client, txClientFromContext(ctx)))
	assert.Check(t, is.Nil(txFileClientFromContext(ctx)))

	fileClient := &ent.FileClient{}
	client.File = fileClient

	assert.Check(t, is.Equal(fileClient, txFileClientFromContext(ent.NewContext(context.Background(), client))))
}

func TestAddFilePermissionsNoFiles(t *testing.T) {
	ctx := context.Background()
	updated, err := AddFilePermissions(ctx)
	assert.Check(t, err)
	assert.Check(t, is.Equal(ctx, updated))
}

func TestTxClientFromContextEmpty(t *testing.T) {
	assert.Check(t, is.Nil(txClientFromContext(context.Background())))
	assert.Check(t, is.Nil(txFileClientFromContext(context.Background())))
}
