// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/entitlementplanhistory"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// EntitlementPlanHistoryQuery is the builder for querying EntitlementPlanHistory entities.
type EntitlementPlanHistoryQuery struct {
	config
	ctx        *QueryContext
	order      []entitlementplanhistory.OrderOption
	inters     []Interceptor
	predicates []predicate.EntitlementPlanHistory
	modifiers  []func(*sql.Selector)
	loadTotal  []func(context.Context, []*EntitlementPlanHistory) error
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the EntitlementPlanHistoryQuery builder.
func (ephq *EntitlementPlanHistoryQuery) Where(ps ...predicate.EntitlementPlanHistory) *EntitlementPlanHistoryQuery {
	ephq.predicates = append(ephq.predicates, ps...)
	return ephq
}

// Limit the number of records to be returned by this query.
func (ephq *EntitlementPlanHistoryQuery) Limit(limit int) *EntitlementPlanHistoryQuery {
	ephq.ctx.Limit = &limit
	return ephq
}

// Offset to start from.
func (ephq *EntitlementPlanHistoryQuery) Offset(offset int) *EntitlementPlanHistoryQuery {
	ephq.ctx.Offset = &offset
	return ephq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (ephq *EntitlementPlanHistoryQuery) Unique(unique bool) *EntitlementPlanHistoryQuery {
	ephq.ctx.Unique = &unique
	return ephq
}

// Order specifies how the records should be ordered.
func (ephq *EntitlementPlanHistoryQuery) Order(o ...entitlementplanhistory.OrderOption) *EntitlementPlanHistoryQuery {
	ephq.order = append(ephq.order, o...)
	return ephq
}

// First returns the first EntitlementPlanHistory entity from the query.
// Returns a *NotFoundError when no EntitlementPlanHistory was found.
func (ephq *EntitlementPlanHistoryQuery) First(ctx context.Context) (*EntitlementPlanHistory, error) {
	nodes, err := ephq.Limit(1).All(setContextOp(ctx, ephq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{entitlementplanhistory.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (ephq *EntitlementPlanHistoryQuery) FirstX(ctx context.Context) *EntitlementPlanHistory {
	node, err := ephq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first EntitlementPlanHistory ID from the query.
// Returns a *NotFoundError when no EntitlementPlanHistory ID was found.
func (ephq *EntitlementPlanHistoryQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = ephq.Limit(1).IDs(setContextOp(ctx, ephq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{entitlementplanhistory.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (ephq *EntitlementPlanHistoryQuery) FirstIDX(ctx context.Context) string {
	id, err := ephq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single EntitlementPlanHistory entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one EntitlementPlanHistory entity is found.
// Returns a *NotFoundError when no EntitlementPlanHistory entities are found.
func (ephq *EntitlementPlanHistoryQuery) Only(ctx context.Context) (*EntitlementPlanHistory, error) {
	nodes, err := ephq.Limit(2).All(setContextOp(ctx, ephq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{entitlementplanhistory.Label}
	default:
		return nil, &NotSingularError{entitlementplanhistory.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (ephq *EntitlementPlanHistoryQuery) OnlyX(ctx context.Context) *EntitlementPlanHistory {
	node, err := ephq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only EntitlementPlanHistory ID in the query.
// Returns a *NotSingularError when more than one EntitlementPlanHistory ID is found.
// Returns a *NotFoundError when no entities are found.
func (ephq *EntitlementPlanHistoryQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = ephq.Limit(2).IDs(setContextOp(ctx, ephq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{entitlementplanhistory.Label}
	default:
		err = &NotSingularError{entitlementplanhistory.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (ephq *EntitlementPlanHistoryQuery) OnlyIDX(ctx context.Context) string {
	id, err := ephq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of EntitlementPlanHistories.
func (ephq *EntitlementPlanHistoryQuery) All(ctx context.Context) ([]*EntitlementPlanHistory, error) {
	ctx = setContextOp(ctx, ephq.ctx, ent.OpQueryAll)
	if err := ephq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*EntitlementPlanHistory, *EntitlementPlanHistoryQuery]()
	return withInterceptors[[]*EntitlementPlanHistory](ctx, ephq, qr, ephq.inters)
}

// AllX is like All, but panics if an error occurs.
func (ephq *EntitlementPlanHistoryQuery) AllX(ctx context.Context) []*EntitlementPlanHistory {
	nodes, err := ephq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of EntitlementPlanHistory IDs.
func (ephq *EntitlementPlanHistoryQuery) IDs(ctx context.Context) (ids []string, err error) {
	if ephq.ctx.Unique == nil && ephq.path != nil {
		ephq.Unique(true)
	}
	ctx = setContextOp(ctx, ephq.ctx, ent.OpQueryIDs)
	if err = ephq.Select(entitlementplanhistory.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (ephq *EntitlementPlanHistoryQuery) IDsX(ctx context.Context) []string {
	ids, err := ephq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (ephq *EntitlementPlanHistoryQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, ephq.ctx, ent.OpQueryCount)
	if err := ephq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, ephq, querierCount[*EntitlementPlanHistoryQuery](), ephq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (ephq *EntitlementPlanHistoryQuery) CountX(ctx context.Context) int {
	count, err := ephq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (ephq *EntitlementPlanHistoryQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, ephq.ctx, ent.OpQueryExist)
	switch _, err := ephq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("generated: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (ephq *EntitlementPlanHistoryQuery) ExistX(ctx context.Context) bool {
	exist, err := ephq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the EntitlementPlanHistoryQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (ephq *EntitlementPlanHistoryQuery) Clone() *EntitlementPlanHistoryQuery {
	if ephq == nil {
		return nil
	}
	return &EntitlementPlanHistoryQuery{
		config:     ephq.config,
		ctx:        ephq.ctx.Clone(),
		order:      append([]entitlementplanhistory.OrderOption{}, ephq.order...),
		inters:     append([]Interceptor{}, ephq.inters...),
		predicates: append([]predicate.EntitlementPlanHistory{}, ephq.predicates...),
		// clone intermediate query.
		sql:  ephq.sql.Clone(),
		path: ephq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		HistoryTime time.Time `json:"history_time,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.EntitlementPlanHistory.Query().
//		GroupBy(entitlementplanhistory.FieldHistoryTime).
//		Aggregate(generated.Count()).
//		Scan(ctx, &v)
func (ephq *EntitlementPlanHistoryQuery) GroupBy(field string, fields ...string) *EntitlementPlanHistoryGroupBy {
	ephq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &EntitlementPlanHistoryGroupBy{build: ephq}
	grbuild.flds = &ephq.ctx.Fields
	grbuild.label = entitlementplanhistory.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		HistoryTime time.Time `json:"history_time,omitempty"`
//	}
//
//	client.EntitlementPlanHistory.Query().
//		Select(entitlementplanhistory.FieldHistoryTime).
//		Scan(ctx, &v)
func (ephq *EntitlementPlanHistoryQuery) Select(fields ...string) *EntitlementPlanHistorySelect {
	ephq.ctx.Fields = append(ephq.ctx.Fields, fields...)
	sbuild := &EntitlementPlanHistorySelect{EntitlementPlanHistoryQuery: ephq}
	sbuild.label = entitlementplanhistory.Label
	sbuild.flds, sbuild.scan = &ephq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a EntitlementPlanHistorySelect configured with the given aggregations.
func (ephq *EntitlementPlanHistoryQuery) Aggregate(fns ...AggregateFunc) *EntitlementPlanHistorySelect {
	return ephq.Select().Aggregate(fns...)
}

func (ephq *EntitlementPlanHistoryQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range ephq.inters {
		if inter == nil {
			return fmt.Errorf("generated: uninitialized interceptor (forgotten import generated/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, ephq); err != nil {
				return err
			}
		}
	}
	for _, f := range ephq.ctx.Fields {
		if !entitlementplanhistory.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
		}
	}
	if ephq.path != nil {
		prev, err := ephq.path(ctx)
		if err != nil {
			return err
		}
		ephq.sql = prev
	}
	if entitlementplanhistory.Policy == nil {
		return errors.New("generated: uninitialized entitlementplanhistory.Policy (forgotten import generated/runtime?)")
	}
	if err := entitlementplanhistory.Policy.EvalQuery(ctx, ephq); err != nil {
		return err
	}
	return nil
}

func (ephq *EntitlementPlanHistoryQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*EntitlementPlanHistory, error) {
	var (
		nodes = []*EntitlementPlanHistory{}
		_spec = ephq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*EntitlementPlanHistory).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &EntitlementPlanHistory{config: ephq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	_spec.Node.Schema = ephq.schemaConfig.EntitlementPlanHistory
	ctx = internal.NewSchemaConfigContext(ctx, ephq.schemaConfig)
	if len(ephq.modifiers) > 0 {
		_spec.Modifiers = ephq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, ephq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	for i := range ephq.loadTotal {
		if err := ephq.loadTotal[i](ctx, nodes); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (ephq *EntitlementPlanHistoryQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := ephq.querySpec()
	_spec.Node.Schema = ephq.schemaConfig.EntitlementPlanHistory
	ctx = internal.NewSchemaConfigContext(ctx, ephq.schemaConfig)
	if len(ephq.modifiers) > 0 {
		_spec.Modifiers = ephq.modifiers
	}
	_spec.Node.Columns = ephq.ctx.Fields
	if len(ephq.ctx.Fields) > 0 {
		_spec.Unique = ephq.ctx.Unique != nil && *ephq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, ephq.driver, _spec)
}

func (ephq *EntitlementPlanHistoryQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(entitlementplanhistory.Table, entitlementplanhistory.Columns, sqlgraph.NewFieldSpec(entitlementplanhistory.FieldID, field.TypeString))
	_spec.From = ephq.sql
	if unique := ephq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if ephq.path != nil {
		_spec.Unique = true
	}
	if fields := ephq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, entitlementplanhistory.FieldID)
		for i := range fields {
			if fields[i] != entitlementplanhistory.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := ephq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := ephq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := ephq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := ephq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (ephq *EntitlementPlanHistoryQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(ephq.driver.Dialect())
	t1 := builder.Table(entitlementplanhistory.Table)
	columns := ephq.ctx.Fields
	if len(columns) == 0 {
		columns = entitlementplanhistory.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if ephq.sql != nil {
		selector = ephq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if ephq.ctx.Unique != nil && *ephq.ctx.Unique {
		selector.Distinct()
	}
	t1.Schema(ephq.schemaConfig.EntitlementPlanHistory)
	ctx = internal.NewSchemaConfigContext(ctx, ephq.schemaConfig)
	selector.WithContext(ctx)
	for _, p := range ephq.predicates {
		p(selector)
	}
	for _, p := range ephq.order {
		p(selector)
	}
	if offset := ephq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := ephq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// EntitlementPlanHistoryGroupBy is the group-by builder for EntitlementPlanHistory entities.
type EntitlementPlanHistoryGroupBy struct {
	selector
	build *EntitlementPlanHistoryQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (ephgb *EntitlementPlanHistoryGroupBy) Aggregate(fns ...AggregateFunc) *EntitlementPlanHistoryGroupBy {
	ephgb.fns = append(ephgb.fns, fns...)
	return ephgb
}

// Scan applies the selector query and scans the result into the given value.
func (ephgb *EntitlementPlanHistoryGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ephgb.build.ctx, ent.OpQueryGroupBy)
	if err := ephgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*EntitlementPlanHistoryQuery, *EntitlementPlanHistoryGroupBy](ctx, ephgb.build, ephgb, ephgb.build.inters, v)
}

func (ephgb *EntitlementPlanHistoryGroupBy) sqlScan(ctx context.Context, root *EntitlementPlanHistoryQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(ephgb.fns))
	for _, fn := range ephgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*ephgb.flds)+len(ephgb.fns))
		for _, f := range *ephgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*ephgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ephgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// EntitlementPlanHistorySelect is the builder for selecting fields of EntitlementPlanHistory entities.
type EntitlementPlanHistorySelect struct {
	*EntitlementPlanHistoryQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ephs *EntitlementPlanHistorySelect) Aggregate(fns ...AggregateFunc) *EntitlementPlanHistorySelect {
	ephs.fns = append(ephs.fns, fns...)
	return ephs
}

// Scan applies the selector query and scans the result into the given value.
func (ephs *EntitlementPlanHistorySelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ephs.ctx, ent.OpQuerySelect)
	if err := ephs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*EntitlementPlanHistoryQuery, *EntitlementPlanHistorySelect](ctx, ephs.EntitlementPlanHistoryQuery, ephs, ephs.inters, v)
}

func (ephs *EntitlementPlanHistorySelect) sqlScan(ctx context.Context, root *EntitlementPlanHistoryQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ephs.fns))
	for _, fn := range ephs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ephs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ephs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}