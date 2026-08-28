package workflows

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
)

type nopDriver struct{}

func (nopDriver) Exec(context.Context, string, any, any) error  { return nil }
func (nopDriver) Query(context.Context, string, any, any) error { return nil }
func (d nopDriver) Tx(context.Context) (dialect.Tx, error)      { return dialect.NopTx(d), nil }
func (nopDriver) Close() error                                  { return nil }
func (nopDriver) Dialect() string                               { return dialect.SQLite }

// TestBuildObjectRefQuery verifies object ref query builder behavior
func TestBuildObjectRefQuery(t *testing.T) {
	query := &generated.WorkflowObjectRefQuery{}

	assert.Nil(t, buildObjectRefQuery(query, &Object{ID: "obj1"}))
	assert.Equal(t, query, buildObjectRefQuery(query, &Object{ID: "obj1", Type: enums.WorkflowObjectTypeControl}))
}

// TestObjectRefIDsErrors verifies error handling for object ref IDs
func TestObjectRefIDsErrors(t *testing.T) {
	ids, err := ObjectRefIDs(context.Background(), nil, &Object{ID: "obj1"})
	assert.ErrorIs(t, err, ErrNilClient)
	assert.Nil(t, ids)

	ids, err = ObjectRefIDs(context.Background(), &generated.Client{}, nil)
	assert.ErrorIs(t, err, ErrMissingObjectID)
	assert.Nil(t, ids)
}

// TestObjectRefIDsMissingOrg verifies auth context errors
func TestObjectRefIDsMissingOrg(t *testing.T) {
	ids, err := ObjectRefIDs(context.Background(), &generated.Client{}, &Object{ID: "obj1"})
	assert.ErrorIs(t, err, auth.ErrNoAuthUser)
	assert.Nil(t, ids)
}

// TestObjectRefIDsUnsupportedType verifies unsupported object type handling
func TestObjectRefIDsUnsupportedType(t *testing.T) {
	orgCtx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())
	client := generated.NewClient(generated.Driver(nopDriver{}))
	ids, err := ObjectRefIDs(orgCtx, client, &Object{ID: "obj1", Type: enums.WorkflowObjectType("Unsupported")})
	assert.ErrorIs(t, err, ErrUnsupportedObjectType)
	assert.Nil(t, ids)
}
