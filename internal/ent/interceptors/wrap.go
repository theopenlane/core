package interceptors

import (
	"context"
	"errors"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/historygenerated"
	hintercept "github.com/theopenlane/core/internal/ent/historygenerated/intercept"
)

// These wrappers allow interceptors to work with multiple ent query packages (e.g., generated and historygenerated)
// and provide a common interface for query manipulation.

// Query is an interface that abstracts over different query types.
type Query interface {
	// Type returns the string representation of the query type.
	Type() string
	// Limit the number of records to be returned by this query.
	Limit(int)
	// Offset to start from.
	Offset(int)
	// Unique configures the query builder to filter duplicate records.
	Unique(bool)
	// Order specifies how the records should be ordered.
	Order(...func(*sql.Selector))
	// WhereP appends storage-level predicates to the query builder. Using this method, users
	// can use type-assertion to append predicates that do not depend on any generated package.
	WhereP(...func(*sql.Selector))
}

var (
	errUnknownQueryType = errors.New("unknown query type")
)

// NewInterceptQuery creates a new Query wrapper for the given query.
func NewInterceptQuery[T generated.Query | historygenerated.Query](query T) (Query, error) {
	// check the main schemas first, if not found check history schemas
	q, err := intercept.NewQuery(query)
	if err != nil {
		if strings.Contains(err.Error(), errUnknownQueryType.Error()) {
			return hintercept.NewQuery(query)
		}

		return nil, err
	}

	return q, nil
}

// The TraverseFunc type is an adapter to allow the use of ordinary function as Traverser.
// If f is a function with the appropriate signature, TraverseFunc(f) is a Traverser that calls f.
type TraverseFunc func(context.Context, Query) error

// Intercept is a dummy implementation of Intercept that returns the next Querier in the pipeline.
func (f TraverseFunc) Intercept(next ent.Querier) ent.Querier {
	return next
}

// Traverse calls f(ctx, q).
func (f TraverseFunc) Traverse(ctx context.Context, q ent.Query) error {
	query, err := NewInterceptQuery(q)
	if err != nil {
		return err
	}

	return f(ctx, query)
}

// type autoTraverseFunc struct {
// 	fn func(context.Context, Query) error
// }

// func (f *autoTraverseFunc) Intercept(next ent.Querier) ent.Querier {
// 	return next
// }

// func (f *autoTraverseFunc) Traverse(ctx context.Context, q ent.Query) error {
// 	wrappedQuery, err := NewInterceptQuery(q)
// 	if err != nil {
// 		return err
// 	}

// 	return f.fn(ctx, wrappedQuery)
// }
