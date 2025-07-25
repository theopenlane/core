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
	"github.com/theopenlane/core/internal/ent/generated/controlimplementationhistory"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// ControlImplementationHistoryQuery is the builder for querying ControlImplementationHistory entities.
type ControlImplementationHistoryQuery struct {
	config
	ctx        *QueryContext
	order      []controlimplementationhistory.OrderOption
	inters     []Interceptor
	predicates []predicate.ControlImplementationHistory
	loadTotal  []func(context.Context, []*ControlImplementationHistory) error
	modifiers  []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the ControlImplementationHistoryQuery builder.
func (cihq *ControlImplementationHistoryQuery) Where(ps ...predicate.ControlImplementationHistory) *ControlImplementationHistoryQuery {
	cihq.predicates = append(cihq.predicates, ps...)
	return cihq
}

// Limit the number of records to be returned by this query.
func (cihq *ControlImplementationHistoryQuery) Limit(limit int) *ControlImplementationHistoryQuery {
	cihq.ctx.Limit = &limit
	return cihq
}

// Offset to start from.
func (cihq *ControlImplementationHistoryQuery) Offset(offset int) *ControlImplementationHistoryQuery {
	cihq.ctx.Offset = &offset
	return cihq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (cihq *ControlImplementationHistoryQuery) Unique(unique bool) *ControlImplementationHistoryQuery {
	cihq.ctx.Unique = &unique
	return cihq
}

// Order specifies how the records should be ordered.
func (cihq *ControlImplementationHistoryQuery) Order(o ...controlimplementationhistory.OrderOption) *ControlImplementationHistoryQuery {
	cihq.order = append(cihq.order, o...)
	return cihq
}

// First returns the first ControlImplementationHistory entity from the query.
// Returns a *NotFoundError when no ControlImplementationHistory was found.
func (cihq *ControlImplementationHistoryQuery) First(ctx context.Context) (*ControlImplementationHistory, error) {
	nodes, err := cihq.Limit(1).All(setContextOp(ctx, cihq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{controlimplementationhistory.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (cihq *ControlImplementationHistoryQuery) FirstX(ctx context.Context) *ControlImplementationHistory {
	node, err := cihq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first ControlImplementationHistory ID from the query.
// Returns a *NotFoundError when no ControlImplementationHistory ID was found.
func (cihq *ControlImplementationHistoryQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = cihq.Limit(1).IDs(setContextOp(ctx, cihq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{controlimplementationhistory.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (cihq *ControlImplementationHistoryQuery) FirstIDX(ctx context.Context) string {
	id, err := cihq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single ControlImplementationHistory entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one ControlImplementationHistory entity is found.
// Returns a *NotFoundError when no ControlImplementationHistory entities are found.
func (cihq *ControlImplementationHistoryQuery) Only(ctx context.Context) (*ControlImplementationHistory, error) {
	nodes, err := cihq.Limit(2).All(setContextOp(ctx, cihq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{controlimplementationhistory.Label}
	default:
		return nil, &NotSingularError{controlimplementationhistory.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (cihq *ControlImplementationHistoryQuery) OnlyX(ctx context.Context) *ControlImplementationHistory {
	node, err := cihq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only ControlImplementationHistory ID in the query.
// Returns a *NotSingularError when more than one ControlImplementationHistory ID is found.
// Returns a *NotFoundError when no entities are found.
func (cihq *ControlImplementationHistoryQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = cihq.Limit(2).IDs(setContextOp(ctx, cihq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{controlimplementationhistory.Label}
	default:
		err = &NotSingularError{controlimplementationhistory.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (cihq *ControlImplementationHistoryQuery) OnlyIDX(ctx context.Context) string {
	id, err := cihq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of ControlImplementationHistories.
func (cihq *ControlImplementationHistoryQuery) All(ctx context.Context) ([]*ControlImplementationHistory, error) {
	ctx = setContextOp(ctx, cihq.ctx, ent.OpQueryAll)
	if err := cihq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*ControlImplementationHistory, *ControlImplementationHistoryQuery]()
	return withInterceptors[[]*ControlImplementationHistory](ctx, cihq, qr, cihq.inters)
}

// AllX is like All, but panics if an error occurs.
func (cihq *ControlImplementationHistoryQuery) AllX(ctx context.Context) []*ControlImplementationHistory {
	nodes, err := cihq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of ControlImplementationHistory IDs.
func (cihq *ControlImplementationHistoryQuery) IDs(ctx context.Context) (ids []string, err error) {
	if cihq.ctx.Unique == nil && cihq.path != nil {
		cihq.Unique(true)
	}
	ctx = setContextOp(ctx, cihq.ctx, ent.OpQueryIDs)
	if err = cihq.Select(controlimplementationhistory.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (cihq *ControlImplementationHistoryQuery) IDsX(ctx context.Context) []string {
	ids, err := cihq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (cihq *ControlImplementationHistoryQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, cihq.ctx, ent.OpQueryCount)
	if err := cihq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, cihq, querierCount[*ControlImplementationHistoryQuery](), cihq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (cihq *ControlImplementationHistoryQuery) CountX(ctx context.Context) int {
	count, err := cihq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (cihq *ControlImplementationHistoryQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, cihq.ctx, ent.OpQueryExist)
	switch _, err := cihq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("generated: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (cihq *ControlImplementationHistoryQuery) ExistX(ctx context.Context) bool {
	exist, err := cihq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the ControlImplementationHistoryQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (cihq *ControlImplementationHistoryQuery) Clone() *ControlImplementationHistoryQuery {
	if cihq == nil {
		return nil
	}
	return &ControlImplementationHistoryQuery{
		config:     cihq.config,
		ctx:        cihq.ctx.Clone(),
		order:      append([]controlimplementationhistory.OrderOption{}, cihq.order...),
		inters:     append([]Interceptor{}, cihq.inters...),
		predicates: append([]predicate.ControlImplementationHistory{}, cihq.predicates...),
		// clone intermediate query.
		sql:       cihq.sql.Clone(),
		path:      cihq.path,
		modifiers: append([]func(*sql.Selector){}, cihq.modifiers...),
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
//	client.ControlImplementationHistory.Query().
//		GroupBy(controlimplementationhistory.FieldHistoryTime).
//		Aggregate(generated.Count()).
//		Scan(ctx, &v)
func (cihq *ControlImplementationHistoryQuery) GroupBy(field string, fields ...string) *ControlImplementationHistoryGroupBy {
	cihq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &ControlImplementationHistoryGroupBy{build: cihq}
	grbuild.flds = &cihq.ctx.Fields
	grbuild.label = controlimplementationhistory.Label
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
//	client.ControlImplementationHistory.Query().
//		Select(controlimplementationhistory.FieldHistoryTime).
//		Scan(ctx, &v)
func (cihq *ControlImplementationHistoryQuery) Select(fields ...string) *ControlImplementationHistorySelect {
	cihq.ctx.Fields = append(cihq.ctx.Fields, fields...)
	sbuild := &ControlImplementationHistorySelect{ControlImplementationHistoryQuery: cihq}
	sbuild.label = controlimplementationhistory.Label
	sbuild.flds, sbuild.scan = &cihq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a ControlImplementationHistorySelect configured with the given aggregations.
func (cihq *ControlImplementationHistoryQuery) Aggregate(fns ...AggregateFunc) *ControlImplementationHistorySelect {
	return cihq.Select().Aggregate(fns...)
}

func (cihq *ControlImplementationHistoryQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range cihq.inters {
		if inter == nil {
			return fmt.Errorf("generated: uninitialized interceptor (forgotten import generated/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, cihq); err != nil {
				return err
			}
		}
	}
	for _, f := range cihq.ctx.Fields {
		if !controlimplementationhistory.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
		}
	}
	if cihq.path != nil {
		prev, err := cihq.path(ctx)
		if err != nil {
			return err
		}
		cihq.sql = prev
	}
	if controlimplementationhistory.Policy == nil {
		return errors.New("generated: uninitialized controlimplementationhistory.Policy (forgotten import generated/runtime?)")
	}
	if err := controlimplementationhistory.Policy.EvalQuery(ctx, cihq); err != nil {
		return err
	}
	return nil
}

func (cihq *ControlImplementationHistoryQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*ControlImplementationHistory, error) {
	var (
		nodes = []*ControlImplementationHistory{}
		_spec = cihq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*ControlImplementationHistory).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &ControlImplementationHistory{config: cihq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	_spec.Node.Schema = cihq.schemaConfig.ControlImplementationHistory
	ctx = internal.NewSchemaConfigContext(ctx, cihq.schemaConfig)
	if len(cihq.modifiers) > 0 {
		_spec.Modifiers = cihq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, cihq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	for i := range cihq.loadTotal {
		if err := cihq.loadTotal[i](ctx, nodes); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (cihq *ControlImplementationHistoryQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := cihq.querySpec()
	_spec.Node.Schema = cihq.schemaConfig.ControlImplementationHistory
	ctx = internal.NewSchemaConfigContext(ctx, cihq.schemaConfig)
	if len(cihq.modifiers) > 0 {
		_spec.Modifiers = cihq.modifiers
	}
	_spec.Node.Columns = cihq.ctx.Fields
	if len(cihq.ctx.Fields) > 0 {
		_spec.Unique = cihq.ctx.Unique != nil && *cihq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, cihq.driver, _spec)
}

func (cihq *ControlImplementationHistoryQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(controlimplementationhistory.Table, controlimplementationhistory.Columns, sqlgraph.NewFieldSpec(controlimplementationhistory.FieldID, field.TypeString))
	_spec.From = cihq.sql
	if unique := cihq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if cihq.path != nil {
		_spec.Unique = true
	}
	if fields := cihq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, controlimplementationhistory.FieldID)
		for i := range fields {
			if fields[i] != controlimplementationhistory.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := cihq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := cihq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := cihq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := cihq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (cihq *ControlImplementationHistoryQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(cihq.driver.Dialect())
	t1 := builder.Table(controlimplementationhistory.Table)
	columns := cihq.ctx.Fields
	if len(columns) == 0 {
		columns = controlimplementationhistory.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if cihq.sql != nil {
		selector = cihq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if cihq.ctx.Unique != nil && *cihq.ctx.Unique {
		selector.Distinct()
	}
	t1.Schema(cihq.schemaConfig.ControlImplementationHistory)
	ctx = internal.NewSchemaConfigContext(ctx, cihq.schemaConfig)
	selector.WithContext(ctx)
	for _, m := range cihq.modifiers {
		m(selector)
	}
	for _, p := range cihq.predicates {
		p(selector)
	}
	for _, p := range cihq.order {
		p(selector)
	}
	if offset := cihq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := cihq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (cihq *ControlImplementationHistoryQuery) Modify(modifiers ...func(s *sql.Selector)) *ControlImplementationHistorySelect {
	cihq.modifiers = append(cihq.modifiers, modifiers...)
	return cihq.Select()
}

// CountIDs returns the count of ids and allows for filtering of the query post retrieval by IDs
func (cihq *ControlImplementationHistoryQuery) CountIDs(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, cihq.ctx, ent.OpQueryIDs)
	if err := cihq.prepareQuery(ctx); err != nil {
		return 0, err
	}

	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return cihq.IDs(ctx)
	})

	ids, err := withInterceptors[[]string](ctx, cihq, qr, cihq.inters)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// ControlImplementationHistoryGroupBy is the group-by builder for ControlImplementationHistory entities.
type ControlImplementationHistoryGroupBy struct {
	selector
	build *ControlImplementationHistoryQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (cihgb *ControlImplementationHistoryGroupBy) Aggregate(fns ...AggregateFunc) *ControlImplementationHistoryGroupBy {
	cihgb.fns = append(cihgb.fns, fns...)
	return cihgb
}

// Scan applies the selector query and scans the result into the given value.
func (cihgb *ControlImplementationHistoryGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, cihgb.build.ctx, ent.OpQueryGroupBy)
	if err := cihgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ControlImplementationHistoryQuery, *ControlImplementationHistoryGroupBy](ctx, cihgb.build, cihgb, cihgb.build.inters, v)
}

func (cihgb *ControlImplementationHistoryGroupBy) sqlScan(ctx context.Context, root *ControlImplementationHistoryQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(cihgb.fns))
	for _, fn := range cihgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*cihgb.flds)+len(cihgb.fns))
		for _, f := range *cihgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*cihgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := cihgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// ControlImplementationHistorySelect is the builder for selecting fields of ControlImplementationHistory entities.
type ControlImplementationHistorySelect struct {
	*ControlImplementationHistoryQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (cihs *ControlImplementationHistorySelect) Aggregate(fns ...AggregateFunc) *ControlImplementationHistorySelect {
	cihs.fns = append(cihs.fns, fns...)
	return cihs
}

// Scan applies the selector query and scans the result into the given value.
func (cihs *ControlImplementationHistorySelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, cihs.ctx, ent.OpQuerySelect)
	if err := cihs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ControlImplementationHistoryQuery, *ControlImplementationHistorySelect](ctx, cihs.ControlImplementationHistoryQuery, cihs, cihs.inters, v)
}

func (cihs *ControlImplementationHistorySelect) sqlScan(ctx context.Context, root *ControlImplementationHistoryQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(cihs.fns))
	for _, fn := range cihs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*cihs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := cihs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (cihs *ControlImplementationHistorySelect) Modify(modifiers ...func(s *sql.Selector)) *ControlImplementationHistorySelect {
	cihs.modifiers = append(cihs.modifiers, modifiers...)
	return cihs
}
