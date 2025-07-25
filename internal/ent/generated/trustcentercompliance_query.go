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
	"github.com/theopenlane/core/internal/ent/generated/trustcentercompliance"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// TrustCenterComplianceQuery is the builder for querying TrustCenterCompliance entities.
type TrustCenterComplianceQuery struct {
	config
	ctx        *QueryContext
	order      []trustcentercompliance.OrderOption
	inters     []Interceptor
	predicates []predicate.TrustCenterCompliance
	loadTotal  []func(context.Context, []*TrustCenterCompliance) error
	modifiers  []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the TrustCenterComplianceQuery builder.
func (tccq *TrustCenterComplianceQuery) Where(ps ...predicate.TrustCenterCompliance) *TrustCenterComplianceQuery {
	tccq.predicates = append(tccq.predicates, ps...)
	return tccq
}

// Limit the number of records to be returned by this query.
func (tccq *TrustCenterComplianceQuery) Limit(limit int) *TrustCenterComplianceQuery {
	tccq.ctx.Limit = &limit
	return tccq
}

// Offset to start from.
func (tccq *TrustCenterComplianceQuery) Offset(offset int) *TrustCenterComplianceQuery {
	tccq.ctx.Offset = &offset
	return tccq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (tccq *TrustCenterComplianceQuery) Unique(unique bool) *TrustCenterComplianceQuery {
	tccq.ctx.Unique = &unique
	return tccq
}

// Order specifies how the records should be ordered.
func (tccq *TrustCenterComplianceQuery) Order(o ...trustcentercompliance.OrderOption) *TrustCenterComplianceQuery {
	tccq.order = append(tccq.order, o...)
	return tccq
}

// First returns the first TrustCenterCompliance entity from the query.
// Returns a *NotFoundError when no TrustCenterCompliance was found.
func (tccq *TrustCenterComplianceQuery) First(ctx context.Context) (*TrustCenterCompliance, error) {
	nodes, err := tccq.Limit(1).All(setContextOp(ctx, tccq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{trustcentercompliance.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (tccq *TrustCenterComplianceQuery) FirstX(ctx context.Context) *TrustCenterCompliance {
	node, err := tccq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first TrustCenterCompliance ID from the query.
// Returns a *NotFoundError when no TrustCenterCompliance ID was found.
func (tccq *TrustCenterComplianceQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = tccq.Limit(1).IDs(setContextOp(ctx, tccq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{trustcentercompliance.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (tccq *TrustCenterComplianceQuery) FirstIDX(ctx context.Context) string {
	id, err := tccq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single TrustCenterCompliance entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one TrustCenterCompliance entity is found.
// Returns a *NotFoundError when no TrustCenterCompliance entities are found.
func (tccq *TrustCenterComplianceQuery) Only(ctx context.Context) (*TrustCenterCompliance, error) {
	nodes, err := tccq.Limit(2).All(setContextOp(ctx, tccq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{trustcentercompliance.Label}
	default:
		return nil, &NotSingularError{trustcentercompliance.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (tccq *TrustCenterComplianceQuery) OnlyX(ctx context.Context) *TrustCenterCompliance {
	node, err := tccq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only TrustCenterCompliance ID in the query.
// Returns a *NotSingularError when more than one TrustCenterCompliance ID is found.
// Returns a *NotFoundError when no entities are found.
func (tccq *TrustCenterComplianceQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = tccq.Limit(2).IDs(setContextOp(ctx, tccq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{trustcentercompliance.Label}
	default:
		err = &NotSingularError{trustcentercompliance.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (tccq *TrustCenterComplianceQuery) OnlyIDX(ctx context.Context) string {
	id, err := tccq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of TrustCenterCompliances.
func (tccq *TrustCenterComplianceQuery) All(ctx context.Context) ([]*TrustCenterCompliance, error) {
	ctx = setContextOp(ctx, tccq.ctx, ent.OpQueryAll)
	if err := tccq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*TrustCenterCompliance, *TrustCenterComplianceQuery]()
	return withInterceptors[[]*TrustCenterCompliance](ctx, tccq, qr, tccq.inters)
}

// AllX is like All, but panics if an error occurs.
func (tccq *TrustCenterComplianceQuery) AllX(ctx context.Context) []*TrustCenterCompliance {
	nodes, err := tccq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of TrustCenterCompliance IDs.
func (tccq *TrustCenterComplianceQuery) IDs(ctx context.Context) (ids []string, err error) {
	if tccq.ctx.Unique == nil && tccq.path != nil {
		tccq.Unique(true)
	}
	ctx = setContextOp(ctx, tccq.ctx, ent.OpQueryIDs)
	if err = tccq.Select(trustcentercompliance.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (tccq *TrustCenterComplianceQuery) IDsX(ctx context.Context) []string {
	ids, err := tccq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (tccq *TrustCenterComplianceQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, tccq.ctx, ent.OpQueryCount)
	if err := tccq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, tccq, querierCount[*TrustCenterComplianceQuery](), tccq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (tccq *TrustCenterComplianceQuery) CountX(ctx context.Context) int {
	count, err := tccq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (tccq *TrustCenterComplianceQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, tccq.ctx, ent.OpQueryExist)
	switch _, err := tccq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("generated: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (tccq *TrustCenterComplianceQuery) ExistX(ctx context.Context) bool {
	exist, err := tccq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the TrustCenterComplianceQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (tccq *TrustCenterComplianceQuery) Clone() *TrustCenterComplianceQuery {
	if tccq == nil {
		return nil
	}
	return &TrustCenterComplianceQuery{
		config:     tccq.config,
		ctx:        tccq.ctx.Clone(),
		order:      append([]trustcentercompliance.OrderOption{}, tccq.order...),
		inters:     append([]Interceptor{}, tccq.inters...),
		predicates: append([]predicate.TrustCenterCompliance{}, tccq.predicates...),
		// clone intermediate query.
		sql:       tccq.sql.Clone(),
		path:      tccq.path,
		modifiers: append([]func(*sql.Selector){}, tccq.modifiers...),
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		CreatedAt time.Time `json:"created_at,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.TrustCenterCompliance.Query().
//		GroupBy(trustcentercompliance.FieldCreatedAt).
//		Aggregate(generated.Count()).
//		Scan(ctx, &v)
func (tccq *TrustCenterComplianceQuery) GroupBy(field string, fields ...string) *TrustCenterComplianceGroupBy {
	tccq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &TrustCenterComplianceGroupBy{build: tccq}
	grbuild.flds = &tccq.ctx.Fields
	grbuild.label = trustcentercompliance.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		CreatedAt time.Time `json:"created_at,omitempty"`
//	}
//
//	client.TrustCenterCompliance.Query().
//		Select(trustcentercompliance.FieldCreatedAt).
//		Scan(ctx, &v)
func (tccq *TrustCenterComplianceQuery) Select(fields ...string) *TrustCenterComplianceSelect {
	tccq.ctx.Fields = append(tccq.ctx.Fields, fields...)
	sbuild := &TrustCenterComplianceSelect{TrustCenterComplianceQuery: tccq}
	sbuild.label = trustcentercompliance.Label
	sbuild.flds, sbuild.scan = &tccq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a TrustCenterComplianceSelect configured with the given aggregations.
func (tccq *TrustCenterComplianceQuery) Aggregate(fns ...AggregateFunc) *TrustCenterComplianceSelect {
	return tccq.Select().Aggregate(fns...)
}

func (tccq *TrustCenterComplianceQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range tccq.inters {
		if inter == nil {
			return fmt.Errorf("generated: uninitialized interceptor (forgotten import generated/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, tccq); err != nil {
				return err
			}
		}
	}
	for _, f := range tccq.ctx.Fields {
		if !trustcentercompliance.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
		}
	}
	if tccq.path != nil {
		prev, err := tccq.path(ctx)
		if err != nil {
			return err
		}
		tccq.sql = prev
	}
	if trustcentercompliance.Policy == nil {
		return errors.New("generated: uninitialized trustcentercompliance.Policy (forgotten import generated/runtime?)")
	}
	if err := trustcentercompliance.Policy.EvalQuery(ctx, tccq); err != nil {
		return err
	}
	return nil
}

func (tccq *TrustCenterComplianceQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*TrustCenterCompliance, error) {
	var (
		nodes = []*TrustCenterCompliance{}
		_spec = tccq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*TrustCenterCompliance).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &TrustCenterCompliance{config: tccq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	_spec.Node.Schema = tccq.schemaConfig.TrustCenterCompliance
	ctx = internal.NewSchemaConfigContext(ctx, tccq.schemaConfig)
	if len(tccq.modifiers) > 0 {
		_spec.Modifiers = tccq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, tccq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	for i := range tccq.loadTotal {
		if err := tccq.loadTotal[i](ctx, nodes); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (tccq *TrustCenterComplianceQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := tccq.querySpec()
	_spec.Node.Schema = tccq.schemaConfig.TrustCenterCompliance
	ctx = internal.NewSchemaConfigContext(ctx, tccq.schemaConfig)
	if len(tccq.modifiers) > 0 {
		_spec.Modifiers = tccq.modifiers
	}
	_spec.Node.Columns = tccq.ctx.Fields
	if len(tccq.ctx.Fields) > 0 {
		_spec.Unique = tccq.ctx.Unique != nil && *tccq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, tccq.driver, _spec)
}

func (tccq *TrustCenterComplianceQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(trustcentercompliance.Table, trustcentercompliance.Columns, sqlgraph.NewFieldSpec(trustcentercompliance.FieldID, field.TypeString))
	_spec.From = tccq.sql
	if unique := tccq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if tccq.path != nil {
		_spec.Unique = true
	}
	if fields := tccq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, trustcentercompliance.FieldID)
		for i := range fields {
			if fields[i] != trustcentercompliance.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := tccq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := tccq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := tccq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := tccq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (tccq *TrustCenterComplianceQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(tccq.driver.Dialect())
	t1 := builder.Table(trustcentercompliance.Table)
	columns := tccq.ctx.Fields
	if len(columns) == 0 {
		columns = trustcentercompliance.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if tccq.sql != nil {
		selector = tccq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if tccq.ctx.Unique != nil && *tccq.ctx.Unique {
		selector.Distinct()
	}
	t1.Schema(tccq.schemaConfig.TrustCenterCompliance)
	ctx = internal.NewSchemaConfigContext(ctx, tccq.schemaConfig)
	selector.WithContext(ctx)
	for _, m := range tccq.modifiers {
		m(selector)
	}
	for _, p := range tccq.predicates {
		p(selector)
	}
	for _, p := range tccq.order {
		p(selector)
	}
	if offset := tccq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := tccq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (tccq *TrustCenterComplianceQuery) Modify(modifiers ...func(s *sql.Selector)) *TrustCenterComplianceSelect {
	tccq.modifiers = append(tccq.modifiers, modifiers...)
	return tccq.Select()
}

// CountIDs returns the count of ids and allows for filtering of the query post retrieval by IDs
func (tccq *TrustCenterComplianceQuery) CountIDs(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, tccq.ctx, ent.OpQueryIDs)
	if err := tccq.prepareQuery(ctx); err != nil {
		return 0, err
	}

	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return tccq.IDs(ctx)
	})

	ids, err := withInterceptors[[]string](ctx, tccq, qr, tccq.inters)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// TrustCenterComplianceGroupBy is the group-by builder for TrustCenterCompliance entities.
type TrustCenterComplianceGroupBy struct {
	selector
	build *TrustCenterComplianceQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (tccgb *TrustCenterComplianceGroupBy) Aggregate(fns ...AggregateFunc) *TrustCenterComplianceGroupBy {
	tccgb.fns = append(tccgb.fns, fns...)
	return tccgb
}

// Scan applies the selector query and scans the result into the given value.
func (tccgb *TrustCenterComplianceGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, tccgb.build.ctx, ent.OpQueryGroupBy)
	if err := tccgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TrustCenterComplianceQuery, *TrustCenterComplianceGroupBy](ctx, tccgb.build, tccgb, tccgb.build.inters, v)
}

func (tccgb *TrustCenterComplianceGroupBy) sqlScan(ctx context.Context, root *TrustCenterComplianceQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(tccgb.fns))
	for _, fn := range tccgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*tccgb.flds)+len(tccgb.fns))
		for _, f := range *tccgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*tccgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := tccgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// TrustCenterComplianceSelect is the builder for selecting fields of TrustCenterCompliance entities.
type TrustCenterComplianceSelect struct {
	*TrustCenterComplianceQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (tccs *TrustCenterComplianceSelect) Aggregate(fns ...AggregateFunc) *TrustCenterComplianceSelect {
	tccs.fns = append(tccs.fns, fns...)
	return tccs
}

// Scan applies the selector query and scans the result into the given value.
func (tccs *TrustCenterComplianceSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, tccs.ctx, ent.OpQuerySelect)
	if err := tccs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TrustCenterComplianceQuery, *TrustCenterComplianceSelect](ctx, tccs.TrustCenterComplianceQuery, tccs, tccs.inters, v)
}

func (tccs *TrustCenterComplianceSelect) sqlScan(ctx context.Context, root *TrustCenterComplianceQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(tccs.fns))
	for _, fn := range tccs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*tccs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := tccs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (tccs *TrustCenterComplianceSelect) Modify(modifiers ...func(s *sql.Selector)) *TrustCenterComplianceSelect {
	tccs.modifiers = append(tccs.modifiers, modifiers...)
	return tccs
}
