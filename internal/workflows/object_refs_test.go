package workflows

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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

	assert.Check(t, is.Nil(buildObjectRefQuery(query, &Object{ID: "obj1"})))
	assert.Check(t, is.Equal(query, buildObjectRefQuery(query, &Object{ID: "obj1", Type: enums.WorkflowObjectTypeControl})))
}

// TestObjectRefIDsErrors verifies error handling for object ref IDs
func TestObjectRefIDsErrors(t *testing.T) {
	ids, err := ObjectRefIDs(context.Background(), nil, &Object{ID: "obj1"})
	assert.Check(t, is.ErrorIs(err, ErrNilClient))
	assert.Check(t, is.Nil(ids))

	ids, err = ObjectRefIDs(context.Background(), &generated.Client{}, nil)
	assert.Check(t, is.ErrorIs(err, ErrMissingObjectID))
	assert.Check(t, is.Nil(ids))
}

// TestObjectRefIDsMissingOrg verifies auth context errors
func TestObjectRefIDsMissingOrg(t *testing.T) {
	ids, err := ObjectRefIDs(context.Background(), &generated.Client{}, &Object{ID: "obj1"})
	assert.Check(t, is.ErrorIs(err, auth.ErrNoAuthUser))
	assert.Check(t, is.Nil(ids))
}

// TestObjectRefIDsUnsupportedType verifies unsupported object type handling
func TestObjectRefIDsUnsupportedType(t *testing.T) {
	orgCtx := auth.NewTestContextWithOrgID(ulids.New().String(), ulids.New().String())
	client := generated.NewClient(generated.Driver(nopDriver{}))
	ids, err := ObjectRefIDs(orgCtx, client, &Object{ID: "obj1", Type: enums.WorkflowObjectType("Unsupported")})
	assert.Check(t, is.ErrorIs(err, ErrUnsupportedObjectType))
	assert.Check(t, is.Nil(ids))
}
