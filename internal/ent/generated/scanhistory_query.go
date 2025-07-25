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
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/scanhistory"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// ScanHistoryQuery is the builder for querying ScanHistory entities.
type ScanHistoryQuery struct {
	config
	ctx        *QueryContext
	order      []scanhistory.OrderOption
	inters     []Interceptor
	predicates []predicate.ScanHistory
	loadTotal  []func(context.Context, []*ScanHistory) error
	modifiers  []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the ScanHistoryQuery builder.
func (shq *ScanHistoryQuery) Where(ps ...predicate.ScanHistory) *ScanHistoryQuery {
	shq.predicates = append(shq.predicates, ps...)
	return shq
}

// Limit the number of records to be returned by this query.
func (shq *ScanHistoryQuery) Limit(limit int) *ScanHistoryQuery {
	shq.ctx.Limit = &limit
	return shq
}

// Offset to start from.
func (shq *ScanHistoryQuery) Offset(offset int) *ScanHistoryQuery {
	shq.ctx.Offset = &offset
	return shq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (shq *ScanHistoryQuery) Unique(unique bool) *ScanHistoryQuery {
	shq.ctx.Unique = &unique
	return shq
}

// Order specifies how the records should be ordered.
func (shq *ScanHistoryQuery) Order(o ...scanhistory.OrderOption) *ScanHistoryQuery {
	shq.order = append(shq.order, o...)
	return shq
}

// First returns the first ScanHistory entity from the query.
// Returns a *NotFoundError when no ScanHistory was found.
func (shq *ScanHistoryQuery) First(ctx context.Context) (*ScanHistory, error) {
	nodes, err := shq.Limit(1).All(setContextOp(ctx, shq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{scanhistory.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (shq *ScanHistoryQuery) FirstX(ctx context.Context) *ScanHistory {
	node, err := shq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first ScanHistory ID from the query.
// Returns a *NotFoundError when no ScanHistory ID was found.
func (shq *ScanHistoryQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = shq.Limit(1).IDs(setContextOp(ctx, shq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{scanhistory.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (shq *ScanHistoryQuery) FirstIDX(ctx context.Context) string {
	id, err := shq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single ScanHistory entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one ScanHistory entity is found.
// Returns a *NotFoundError when no ScanHistory entities are found.
func (shq *ScanHistoryQuery) Only(ctx context.Context) (*ScanHistory, error) {
	nodes, err := shq.Limit(2).All(setContextOp(ctx, shq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{scanhistory.Label}
	default:
		return nil, &NotSingularError{scanhistory.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (shq *ScanHistoryQuery) OnlyX(ctx context.Context) *ScanHistory {
	node, err := shq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only ScanHistory ID in the query.
// Returns a *NotSingularError when more than one ScanHistory ID is found.
// Returns a *NotFoundError when no entities are found.
func (shq *ScanHistoryQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = shq.Limit(2).IDs(setContextOp(ctx, shq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{scanhistory.Label}
	default:
		err = &NotSingularError{scanhistory.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (shq *ScanHistoryQuery) OnlyIDX(ctx context.Context) string {
	id, err := shq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of ScanHistories.
func (shq *ScanHistoryQuery) All(ctx context.Context) ([]*ScanHistory, error) {
	ctx = setContextOp(ctx, shq.ctx, ent.OpQueryAll)
	if err := shq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*ScanHistory, *ScanHistoryQuery]()
	return withInterceptors[[]*ScanHistory](ctx, shq, qr, shq.inters)
}

// AllX is like All, but panics if an error occurs.
func (shq *ScanHistoryQuery) AllX(ctx context.Context) []*ScanHistory {
	nodes, err := shq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of ScanHistory IDs.
func (shq *ScanHistoryQuery) IDs(ctx context.Context) (ids []string, err error) {
	if shq.ctx.Unique == nil && shq.path != nil {
		shq.Unique(true)
	}
	ctx = setContextOp(ctx, shq.ctx, ent.OpQueryIDs)
	if err = shq.Select(scanhistory.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (shq *ScanHistoryQuery) IDsX(ctx context.Context) []string {
	ids, err := shq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (shq *ScanHistoryQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, shq.ctx, ent.OpQueryCount)
	if err := shq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, shq, querierCount[*ScanHistoryQuery](), shq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (shq *ScanHistoryQuery) CountX(ctx context.Context) int {
	count, err := shq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (shq *ScanHistoryQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, shq.ctx, ent.OpQueryExist)
	switch _, err := shq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("generated: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (shq *ScanHistoryQuery) ExistX(ctx context.Context) bool {
	exist, err := shq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the ScanHistoryQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (shq *ScanHistoryQuery) Clone() *ScanHistoryQuery {
	if shq == nil {
		return nil
	}
	return &ScanHistoryQuery{
		config:     shq.config,
		ctx:        shq.ctx.Clone(),
		order:      append([]scanhistory.OrderOption{}, shq.order...),
		inters:     append([]Interceptor{}, shq.inters...),
		predicates: append([]predicate.ScanHistory{}, shq.predicates...),
		// clone intermediate query.
		sql:       shq.sql.Clone(),
		path:      shq.path,
		modifiers: append([]func(*sql.Selector){}, shq.modifiers...),
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
//	client.ScanHistory.Query().
//		GroupBy(scanhistory.FieldHistoryTime).
//		Aggregate(generated.Count()).
//		Scan(ctx, &v)
func (shq *ScanHistoryQuery) GroupBy(field string, fields ...string) *ScanHistoryGroupBy {
	shq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &ScanHistoryGroupBy{build: shq}
	grbuild.flds = &shq.ctx.Fields
	grbuild.label = scanhistory.Label
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
//	client.ScanHistory.Query().
//		Select(scanhistory.FieldHistoryTime).
//		Scan(ctx, &v)
func (shq *ScanHistoryQuery) Select(fields ...string) *ScanHistorySelect {
	shq.ctx.Fields = append(shq.ctx.Fields, fields...)
	sbuild := &ScanHistorySelect{ScanHistoryQuery: shq}
	sbuild.label = scanhistory.Label
	sbuild.flds, sbuild.scan = &shq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a ScanHistorySelect configured with the given aggregations.
func (shq *ScanHistoryQuery) Aggregate(fns ...AggregateFunc) *ScanHistorySelect {
	return shq.Select().Aggregate(fns...)
}

func (shq *ScanHistoryQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range shq.inters {
		if inter == nil {
			return fmt.Errorf("generated: uninitialized interceptor (forgotten import generated/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, shq); err != nil {
				return err
			}
		}
	}
	for _, f := range shq.ctx.Fields {
		if !scanhistory.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
		}
	}
	if shq.path != nil {
		prev, err := shq.path(ctx)
		if err != nil {
			return err
		}
		shq.sql = prev
	}
	if scanhistory.Policy == nil {
		return errors.New("generated: uninitialized scanhistory.Policy (forgotten import generated/runtime?)")
	}
	if err := scanhistory.Policy.EvalQuery(ctx, shq); err != nil {
		return err
	}
	return nil
}

func (shq *ScanHistoryQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*ScanHistory, error) {
	var (
		nodes = []*ScanHistory{}
		_spec = shq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*ScanHistory).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &ScanHistory{config: shq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	_spec.Node.Schema = shq.schemaConfig.ScanHistory
	ctx = internal.NewSchemaConfigContext(ctx, shq.schemaConfig)
	if len(shq.modifiers) > 0 {
		_spec.Modifiers = shq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, shq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	for i := range shq.loadTotal {
		if err := shq.loadTotal[i](ctx, nodes); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (shq *ScanHistoryQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := shq.querySpec()
	_spec.Node.Schema = shq.schemaConfig.ScanHistory
	ctx = internal.NewSchemaConfigContext(ctx, shq.schemaConfig)
	if len(shq.modifiers) > 0 {
		_spec.Modifiers = shq.modifiers
	}
	_spec.Node.Columns = shq.ctx.Fields
	if len(shq.ctx.Fields) > 0 {
		_spec.Unique = shq.ctx.Unique != nil && *shq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, shq.driver, _spec)
}

func (shq *ScanHistoryQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(scanhistory.Table, scanhistory.Columns, sqlgraph.NewFieldSpec(scanhistory.FieldID, field.TypeString))
	_spec.From = shq.sql
	if unique := shq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if shq.path != nil {
		_spec.Unique = true
	}
	if fields := shq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, scanhistory.FieldID)
		for i := range fields {
			if fields[i] != scanhistory.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := shq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := shq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := shq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := shq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (shq *ScanHistoryQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(shq.driver.Dialect())
	t1 := builder.Table(scanhistory.Table)
	columns := shq.ctx.Fields
	if len(columns) == 0 {
		columns = scanhistory.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if shq.sql != nil {
		selector = shq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if shq.ctx.Unique != nil && *shq.ctx.Unique {
		selector.Distinct()
	}
	t1.Schema(shq.schemaConfig.ScanHistory)
	ctx = internal.NewSchemaConfigContext(ctx, shq.schemaConfig)
	selector.WithContext(ctx)
	for _, m := range shq.modifiers {
		m(selector)
	}
	for _, p := range shq.predicates {
		p(selector)
	}
	for _, p := range shq.order {
		p(selector)
	}
	if offset := shq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := shq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (shq *ScanHistoryQuery) Modify(modifiers ...func(s *sql.Selector)) *ScanHistorySelect {
	shq.modifiers = append(shq.modifiers, modifiers...)
	return shq.Select()
}

// CountIDs returns the count of ids and allows for filtering of the query post retrieval by IDs
func (shq *ScanHistoryQuery) CountIDs(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, shq.ctx, ent.OpQueryIDs)
	if err := shq.prepareQuery(ctx); err != nil {
		return 0, err
	}

	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return shq.IDs(ctx)
	})

	ids, err := withInterceptors[[]string](ctx, shq, qr, shq.inters)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// ScanHistoryGroupBy is the group-by builder for ScanHistory entities.
type ScanHistoryGroupBy struct {
	selector
	build *ScanHistoryQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (shgb *ScanHistoryGroupBy) Aggregate(fns ...AggregateFunc) *ScanHistoryGroupBy {
	shgb.fns = append(shgb.fns, fns...)
	return shgb
}

// Scan applies the selector query and scans the result into the given value.
func (shgb *ScanHistoryGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, shgb.build.ctx, ent.OpQueryGroupBy)
	if err := shgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ScanHistoryQuery, *ScanHistoryGroupBy](ctx, shgb.build, shgb, shgb.build.inters, v)
}

func (shgb *ScanHistoryGroupBy) sqlScan(ctx context.Context, root *ScanHistoryQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(shgb.fns))
	for _, fn := range shgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*shgb.flds)+len(shgb.fns))
		for _, f := range *shgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*shgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := shgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// ScanHistorySelect is the builder for selecting fields of ScanHistory entities.
type ScanHistorySelect struct {
	*ScanHistoryQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (shs *ScanHistorySelect) Aggregate(fns ...AggregateFunc) *ScanHistorySelect {
	shs.fns = append(shs.fns, fns...)
	return shs
}

// Scan applies the selector query and scans the result into the given value.
func (shs *ScanHistorySelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, shs.ctx, ent.OpQuerySelect)
	if err := shs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ScanHistoryQuery, *ScanHistorySelect](ctx, shs.ScanHistoryQuery, shs, shs.inters, v)
}

func (shs *ScanHistorySelect) sqlScan(ctx context.Context, root *ScanHistoryQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(shs.fns))
	for _, fn := range shs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*shs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := shs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (shs *ScanHistorySelect) Modify(modifiers ...func(s *sql.Selector)) *ScanHistorySelect {
	shs.modifiers = append(shs.modifiers, modifiers...)
	return shs
}
