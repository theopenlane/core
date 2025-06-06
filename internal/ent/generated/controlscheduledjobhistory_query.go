// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/controlscheduledjobhistory"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// ControlScheduledJobHistoryQuery is the builder for querying ControlScheduledJobHistory entities.
type ControlScheduledJobHistoryQuery struct {
	config
	ctx        *QueryContext
	order      []controlscheduledjobhistory.OrderOption
	inters     []Interceptor
	predicates []predicate.ControlScheduledJobHistory
	loadTotal  []func(context.Context, []*ControlScheduledJobHistory) error
	modifiers  []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the ControlScheduledJobHistoryQuery builder.
func (csjhq *ControlScheduledJobHistoryQuery) Where(ps ...predicate.ControlScheduledJobHistory) *ControlScheduledJobHistoryQuery {
	csjhq.predicates = append(csjhq.predicates, ps...)
	return csjhq
}

// Limit the number of records to be returned by this query.
func (csjhq *ControlScheduledJobHistoryQuery) Limit(limit int) *ControlScheduledJobHistoryQuery {
	csjhq.ctx.Limit = &limit
	return csjhq
}

// Offset to start from.
func (csjhq *ControlScheduledJobHistoryQuery) Offset(offset int) *ControlScheduledJobHistoryQuery {
	csjhq.ctx.Offset = &offset
	return csjhq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (csjhq *ControlScheduledJobHistoryQuery) Unique(unique bool) *ControlScheduledJobHistoryQuery {
	csjhq.ctx.Unique = &unique
	return csjhq
}

// Order specifies how the records should be ordered.
func (csjhq *ControlScheduledJobHistoryQuery) Order(o ...controlscheduledjobhistory.OrderOption) *ControlScheduledJobHistoryQuery {
	csjhq.order = append(csjhq.order, o...)
	return csjhq
}

// First returns the first ControlScheduledJobHistory entity from the query.
// Returns a *NotFoundError when no ControlScheduledJobHistory was found.
func (csjhq *ControlScheduledJobHistoryQuery) First(ctx context.Context) (*ControlScheduledJobHistory, error) {
	nodes, err := csjhq.Limit(1).All(setContextOp(ctx, csjhq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{controlscheduledjobhistory.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (csjhq *ControlScheduledJobHistoryQuery) FirstX(ctx context.Context) *ControlScheduledJobHistory {
	node, err := csjhq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first ControlScheduledJobHistory ID from the query.
// Returns a *NotFoundError when no ControlScheduledJobHistory ID was found.
func (csjhq *ControlScheduledJobHistoryQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = csjhq.Limit(1).IDs(setContextOp(ctx, csjhq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{controlscheduledjobhistory.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (csjhq *ControlScheduledJobHistoryQuery) FirstIDX(ctx context.Context) string {
	id, err := csjhq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single ControlScheduledJobHistory entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one ControlScheduledJobHistory entity is found.
// Returns a *NotFoundError when no ControlScheduledJobHistory entities are found.
func (csjhq *ControlScheduledJobHistoryQuery) Only(ctx context.Context) (*ControlScheduledJobHistory, error) {
	nodes, err := csjhq.Limit(2).All(setContextOp(ctx, csjhq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{controlscheduledjobhistory.Label}
	default:
		return nil, &NotSingularError{controlscheduledjobhistory.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (csjhq *ControlScheduledJobHistoryQuery) OnlyX(ctx context.Context) *ControlScheduledJobHistory {
	node, err := csjhq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only ControlScheduledJobHistory ID in the query.
// Returns a *NotSingularError when more than one ControlScheduledJobHistory ID is found.
// Returns a *NotFoundError when no entities are found.
func (csjhq *ControlScheduledJobHistoryQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = csjhq.Limit(2).IDs(setContextOp(ctx, csjhq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{controlscheduledjobhistory.Label}
	default:
		err = &NotSingularError{controlscheduledjobhistory.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (csjhq *ControlScheduledJobHistoryQuery) OnlyIDX(ctx context.Context) string {
	id, err := csjhq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of ControlScheduledJobHistories.
func (csjhq *ControlScheduledJobHistoryQuery) All(ctx context.Context) ([]*ControlScheduledJobHistory, error) {
	ctx = setContextOp(ctx, csjhq.ctx, ent.OpQueryAll)
	if err := csjhq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*ControlScheduledJobHistory, *ControlScheduledJobHistoryQuery]()
	return withInterceptors[[]*ControlScheduledJobHistory](ctx, csjhq, qr, csjhq.inters)
}

// AllX is like All, but panics if an error occurs.
func (csjhq *ControlScheduledJobHistoryQuery) AllX(ctx context.Context) []*ControlScheduledJobHistory {
	nodes, err := csjhq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of ControlScheduledJobHistory IDs.
func (csjhq *ControlScheduledJobHistoryQuery) IDs(ctx context.Context) (ids []string, err error) {
	if csjhq.ctx.Unique == nil && csjhq.path != nil {
		csjhq.Unique(true)
	}
	ctx = setContextOp(ctx, csjhq.ctx, ent.OpQueryIDs)
	if err = csjhq.Select(controlscheduledjobhistory.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (csjhq *ControlScheduledJobHistoryQuery) IDsX(ctx context.Context) []string {
	ids, err := csjhq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (csjhq *ControlScheduledJobHistoryQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, csjhq.ctx, ent.OpQueryCount)
	if err := csjhq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, csjhq, querierCount[*ControlScheduledJobHistoryQuery](), csjhq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (csjhq *ControlScheduledJobHistoryQuery) CountX(ctx context.Context) int {
	count, err := csjhq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (csjhq *ControlScheduledJobHistoryQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, csjhq.ctx, ent.OpQueryExist)
	switch _, err := csjhq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("generated: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (csjhq *ControlScheduledJobHistoryQuery) ExistX(ctx context.Context) bool {
	exist, err := csjhq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the ControlScheduledJobHistoryQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (csjhq *ControlScheduledJobHistoryQuery) Clone() *ControlScheduledJobHistoryQuery {
	if csjhq == nil {
		return nil
	}
	return &ControlScheduledJobHistoryQuery{
		config:     csjhq.config,
		ctx:        csjhq.ctx.Clone(),
		order:      append([]controlscheduledjobhistory.OrderOption{}, csjhq.order...),
		inters:     append([]Interceptor{}, csjhq.inters...),
		predicates: append([]predicate.ControlScheduledJobHistory{}, csjhq.predicates...),
		// clone intermediate query.
		sql:       csjhq.sql.Clone(),
		path:      csjhq.path,
		modifiers: append([]func(*sql.Selector){}, csjhq.modifiers...),
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
//	client.ControlScheduledJobHistory.Query().
//		GroupBy(controlscheduledjobhistory.FieldHistoryTime).
//		Aggregate(generated.Count()).
//		Scan(ctx, &v)
func (csjhq *ControlScheduledJobHistoryQuery) GroupBy(field string, fields ...string) *ControlScheduledJobHistoryGroupBy {
	csjhq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &ControlScheduledJobHistoryGroupBy{build: csjhq}
	grbuild.flds = &csjhq.ctx.Fields
	grbuild.label = controlscheduledjobhistory.Label
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
//	client.ControlScheduledJobHistory.Query().
//		Select(controlscheduledjobhistory.FieldHistoryTime).
//		Scan(ctx, &v)
func (csjhq *ControlScheduledJobHistoryQuery) Select(fields ...string) *ControlScheduledJobHistorySelect {
	csjhq.ctx.Fields = append(csjhq.ctx.Fields, fields...)
	sbuild := &ControlScheduledJobHistorySelect{ControlScheduledJobHistoryQuery: csjhq}
	sbuild.label = controlscheduledjobhistory.Label
	sbuild.flds, sbuild.scan = &csjhq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a ControlScheduledJobHistorySelect configured with the given aggregations.
func (csjhq *ControlScheduledJobHistoryQuery) Aggregate(fns ...AggregateFunc) *ControlScheduledJobHistorySelect {
	return csjhq.Select().Aggregate(fns...)
}

func (csjhq *ControlScheduledJobHistoryQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range csjhq.inters {
		if inter == nil {
			return fmt.Errorf("generated: uninitialized interceptor (forgotten import generated/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, csjhq); err != nil {
				return err
			}
		}
	}
	for _, f := range csjhq.ctx.Fields {
		if !controlscheduledjobhistory.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
		}
	}
	if csjhq.path != nil {
		prev, err := csjhq.path(ctx)
		if err != nil {
			return err
		}
		csjhq.sql = prev
	}
	return nil
}

func (csjhq *ControlScheduledJobHistoryQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*ControlScheduledJobHistory, error) {
	var (
		nodes = []*ControlScheduledJobHistory{}
		_spec = csjhq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*ControlScheduledJobHistory).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &ControlScheduledJobHistory{config: csjhq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	_spec.Node.Schema = csjhq.schemaConfig.ControlScheduledJobHistory
	ctx = internal.NewSchemaConfigContext(ctx, csjhq.schemaConfig)
	if len(csjhq.modifiers) > 0 {
		_spec.Modifiers = csjhq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, csjhq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	for i := range csjhq.loadTotal {
		if err := csjhq.loadTotal[i](ctx, nodes); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (csjhq *ControlScheduledJobHistoryQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := csjhq.querySpec()
	_spec.Node.Schema = csjhq.schemaConfig.ControlScheduledJobHistory
	ctx = internal.NewSchemaConfigContext(ctx, csjhq.schemaConfig)
	if len(csjhq.modifiers) > 0 {
		_spec.Modifiers = csjhq.modifiers
	}
	_spec.Node.Columns = csjhq.ctx.Fields
	if len(csjhq.ctx.Fields) > 0 {
		_spec.Unique = csjhq.ctx.Unique != nil && *csjhq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, csjhq.driver, _spec)
}

func (csjhq *ControlScheduledJobHistoryQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(controlscheduledjobhistory.Table, controlscheduledjobhistory.Columns, sqlgraph.NewFieldSpec(controlscheduledjobhistory.FieldID, field.TypeString))
	_spec.From = csjhq.sql
	if unique := csjhq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if csjhq.path != nil {
		_spec.Unique = true
	}
	if fields := csjhq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, controlscheduledjobhistory.FieldID)
		for i := range fields {
			if fields[i] != controlscheduledjobhistory.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := csjhq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := csjhq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := csjhq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := csjhq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (csjhq *ControlScheduledJobHistoryQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(csjhq.driver.Dialect())
	t1 := builder.Table(controlscheduledjobhistory.Table)
	columns := csjhq.ctx.Fields
	if len(columns) == 0 {
		columns = controlscheduledjobhistory.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if csjhq.sql != nil {
		selector = csjhq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if csjhq.ctx.Unique != nil && *csjhq.ctx.Unique {
		selector.Distinct()
	}
	t1.Schema(csjhq.schemaConfig.ControlScheduledJobHistory)
	ctx = internal.NewSchemaConfigContext(ctx, csjhq.schemaConfig)
	selector.WithContext(ctx)
	for _, m := range csjhq.modifiers {
		m(selector)
	}
	for _, p := range csjhq.predicates {
		p(selector)
	}
	for _, p := range csjhq.order {
		p(selector)
	}
	if offset := csjhq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := csjhq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (csjhq *ControlScheduledJobHistoryQuery) Modify(modifiers ...func(s *sql.Selector)) *ControlScheduledJobHistorySelect {
	csjhq.modifiers = append(csjhq.modifiers, modifiers...)
	return csjhq.Select()
}

// CountIDs returns the count of ids and allows for filtering of the query post retrieval by IDs
func (csjhq *ControlScheduledJobHistoryQuery) CountIDs(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, csjhq.ctx, ent.OpQueryIDs)
	if err := csjhq.prepareQuery(ctx); err != nil {
		return 0, err
	}

	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return csjhq.IDs(ctx)
	})

	ids, err := withInterceptors[[]string](ctx, csjhq, qr, csjhq.inters)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// ControlScheduledJobHistoryGroupBy is the group-by builder for ControlScheduledJobHistory entities.
type ControlScheduledJobHistoryGroupBy struct {
	selector
	build *ControlScheduledJobHistoryQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (csjhgb *ControlScheduledJobHistoryGroupBy) Aggregate(fns ...AggregateFunc) *ControlScheduledJobHistoryGroupBy {
	csjhgb.fns = append(csjhgb.fns, fns...)
	return csjhgb
}

// Scan applies the selector query and scans the result into the given value.
func (csjhgb *ControlScheduledJobHistoryGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, csjhgb.build.ctx, ent.OpQueryGroupBy)
	if err := csjhgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ControlScheduledJobHistoryQuery, *ControlScheduledJobHistoryGroupBy](ctx, csjhgb.build, csjhgb, csjhgb.build.inters, v)
}

func (csjhgb *ControlScheduledJobHistoryGroupBy) sqlScan(ctx context.Context, root *ControlScheduledJobHistoryQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(csjhgb.fns))
	for _, fn := range csjhgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*csjhgb.flds)+len(csjhgb.fns))
		for _, f := range *csjhgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*csjhgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := csjhgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// ControlScheduledJobHistorySelect is the builder for selecting fields of ControlScheduledJobHistory entities.
type ControlScheduledJobHistorySelect struct {
	*ControlScheduledJobHistoryQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (csjhs *ControlScheduledJobHistorySelect) Aggregate(fns ...AggregateFunc) *ControlScheduledJobHistorySelect {
	csjhs.fns = append(csjhs.fns, fns...)
	return csjhs
}

// Scan applies the selector query and scans the result into the given value.
func (csjhs *ControlScheduledJobHistorySelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, csjhs.ctx, ent.OpQuerySelect)
	if err := csjhs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ControlScheduledJobHistoryQuery, *ControlScheduledJobHistorySelect](ctx, csjhs.ControlScheduledJobHistoryQuery, csjhs, csjhs.inters, v)
}

func (csjhs *ControlScheduledJobHistorySelect) sqlScan(ctx context.Context, root *ControlScheduledJobHistoryQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(csjhs.fns))
	for _, fn := range csjhs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*csjhs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := csjhs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (csjhs *ControlScheduledJobHistorySelect) Modify(modifiers ...func(s *sql.Selector)) *ControlScheduledJobHistorySelect {
	csjhs.modifiers = append(csjhs.modifiers, modifiers...)
	return csjhs
}
