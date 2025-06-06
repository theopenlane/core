// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/jobrunner"
	"github.com/theopenlane/core/internal/ent/generated/jobrunnertoken"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// JobRunnerTokenQuery is the builder for querying JobRunnerToken entities.
type JobRunnerTokenQuery struct {
	config
	ctx                 *QueryContext
	order               []jobrunnertoken.OrderOption
	inters              []Interceptor
	predicates          []predicate.JobRunnerToken
	withOwner           *OrganizationQuery
	withJobRunners      *JobRunnerQuery
	loadTotal           []func(context.Context, []*JobRunnerToken) error
	modifiers           []func(*sql.Selector)
	withNamedJobRunners map[string]*JobRunnerQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the JobRunnerTokenQuery builder.
func (jrtq *JobRunnerTokenQuery) Where(ps ...predicate.JobRunnerToken) *JobRunnerTokenQuery {
	jrtq.predicates = append(jrtq.predicates, ps...)
	return jrtq
}

// Limit the number of records to be returned by this query.
func (jrtq *JobRunnerTokenQuery) Limit(limit int) *JobRunnerTokenQuery {
	jrtq.ctx.Limit = &limit
	return jrtq
}

// Offset to start from.
func (jrtq *JobRunnerTokenQuery) Offset(offset int) *JobRunnerTokenQuery {
	jrtq.ctx.Offset = &offset
	return jrtq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (jrtq *JobRunnerTokenQuery) Unique(unique bool) *JobRunnerTokenQuery {
	jrtq.ctx.Unique = &unique
	return jrtq
}

// Order specifies how the records should be ordered.
func (jrtq *JobRunnerTokenQuery) Order(o ...jobrunnertoken.OrderOption) *JobRunnerTokenQuery {
	jrtq.order = append(jrtq.order, o...)
	return jrtq
}

// QueryOwner chains the current query on the "owner" edge.
func (jrtq *JobRunnerTokenQuery) QueryOwner() *OrganizationQuery {
	query := (&OrganizationClient{config: jrtq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := jrtq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := jrtq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(jobrunnertoken.Table, jobrunnertoken.FieldID, selector),
			sqlgraph.To(organization.Table, organization.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, jobrunnertoken.OwnerTable, jobrunnertoken.OwnerColumn),
		)
		schemaConfig := jrtq.schemaConfig
		step.To.Schema = schemaConfig.Organization
		step.Edge.Schema = schemaConfig.JobRunnerToken
		fromU = sqlgraph.SetNeighbors(jrtq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryJobRunners chains the current query on the "job_runners" edge.
func (jrtq *JobRunnerTokenQuery) QueryJobRunners() *JobRunnerQuery {
	query := (&JobRunnerClient{config: jrtq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := jrtq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := jrtq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(jobrunnertoken.Table, jobrunnertoken.FieldID, selector),
			sqlgraph.To(jobrunner.Table, jobrunner.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, jobrunnertoken.JobRunnersTable, jobrunnertoken.JobRunnersPrimaryKey...),
		)
		schemaConfig := jrtq.schemaConfig
		step.To.Schema = schemaConfig.JobRunner
		step.Edge.Schema = schemaConfig.JobRunnerJobRunnerTokens
		fromU = sqlgraph.SetNeighbors(jrtq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first JobRunnerToken entity from the query.
// Returns a *NotFoundError when no JobRunnerToken was found.
func (jrtq *JobRunnerTokenQuery) First(ctx context.Context) (*JobRunnerToken, error) {
	nodes, err := jrtq.Limit(1).All(setContextOp(ctx, jrtq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{jobrunnertoken.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (jrtq *JobRunnerTokenQuery) FirstX(ctx context.Context) *JobRunnerToken {
	node, err := jrtq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first JobRunnerToken ID from the query.
// Returns a *NotFoundError when no JobRunnerToken ID was found.
func (jrtq *JobRunnerTokenQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = jrtq.Limit(1).IDs(setContextOp(ctx, jrtq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{jobrunnertoken.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (jrtq *JobRunnerTokenQuery) FirstIDX(ctx context.Context) string {
	id, err := jrtq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single JobRunnerToken entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one JobRunnerToken entity is found.
// Returns a *NotFoundError when no JobRunnerToken entities are found.
func (jrtq *JobRunnerTokenQuery) Only(ctx context.Context) (*JobRunnerToken, error) {
	nodes, err := jrtq.Limit(2).All(setContextOp(ctx, jrtq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{jobrunnertoken.Label}
	default:
		return nil, &NotSingularError{jobrunnertoken.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (jrtq *JobRunnerTokenQuery) OnlyX(ctx context.Context) *JobRunnerToken {
	node, err := jrtq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only JobRunnerToken ID in the query.
// Returns a *NotSingularError when more than one JobRunnerToken ID is found.
// Returns a *NotFoundError when no entities are found.
func (jrtq *JobRunnerTokenQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = jrtq.Limit(2).IDs(setContextOp(ctx, jrtq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{jobrunnertoken.Label}
	default:
		err = &NotSingularError{jobrunnertoken.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (jrtq *JobRunnerTokenQuery) OnlyIDX(ctx context.Context) string {
	id, err := jrtq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of JobRunnerTokens.
func (jrtq *JobRunnerTokenQuery) All(ctx context.Context) ([]*JobRunnerToken, error) {
	ctx = setContextOp(ctx, jrtq.ctx, ent.OpQueryAll)
	if err := jrtq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*JobRunnerToken, *JobRunnerTokenQuery]()
	return withInterceptors[[]*JobRunnerToken](ctx, jrtq, qr, jrtq.inters)
}

// AllX is like All, but panics if an error occurs.
func (jrtq *JobRunnerTokenQuery) AllX(ctx context.Context) []*JobRunnerToken {
	nodes, err := jrtq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of JobRunnerToken IDs.
func (jrtq *JobRunnerTokenQuery) IDs(ctx context.Context) (ids []string, err error) {
	if jrtq.ctx.Unique == nil && jrtq.path != nil {
		jrtq.Unique(true)
	}
	ctx = setContextOp(ctx, jrtq.ctx, ent.OpQueryIDs)
	if err = jrtq.Select(jobrunnertoken.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (jrtq *JobRunnerTokenQuery) IDsX(ctx context.Context) []string {
	ids, err := jrtq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (jrtq *JobRunnerTokenQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, jrtq.ctx, ent.OpQueryCount)
	if err := jrtq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, jrtq, querierCount[*JobRunnerTokenQuery](), jrtq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (jrtq *JobRunnerTokenQuery) CountX(ctx context.Context) int {
	count, err := jrtq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (jrtq *JobRunnerTokenQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, jrtq.ctx, ent.OpQueryExist)
	switch _, err := jrtq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("generated: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (jrtq *JobRunnerTokenQuery) ExistX(ctx context.Context) bool {
	exist, err := jrtq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the JobRunnerTokenQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (jrtq *JobRunnerTokenQuery) Clone() *JobRunnerTokenQuery {
	if jrtq == nil {
		return nil
	}
	return &JobRunnerTokenQuery{
		config:         jrtq.config,
		ctx:            jrtq.ctx.Clone(),
		order:          append([]jobrunnertoken.OrderOption{}, jrtq.order...),
		inters:         append([]Interceptor{}, jrtq.inters...),
		predicates:     append([]predicate.JobRunnerToken{}, jrtq.predicates...),
		withOwner:      jrtq.withOwner.Clone(),
		withJobRunners: jrtq.withJobRunners.Clone(),
		// clone intermediate query.
		sql:       jrtq.sql.Clone(),
		path:      jrtq.path,
		modifiers: append([]func(*sql.Selector){}, jrtq.modifiers...),
	}
}

// WithOwner tells the query-builder to eager-load the nodes that are connected to
// the "owner" edge. The optional arguments are used to configure the query builder of the edge.
func (jrtq *JobRunnerTokenQuery) WithOwner(opts ...func(*OrganizationQuery)) *JobRunnerTokenQuery {
	query := (&OrganizationClient{config: jrtq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	jrtq.withOwner = query
	return jrtq
}

// WithJobRunners tells the query-builder to eager-load the nodes that are connected to
// the "job_runners" edge. The optional arguments are used to configure the query builder of the edge.
func (jrtq *JobRunnerTokenQuery) WithJobRunners(opts ...func(*JobRunnerQuery)) *JobRunnerTokenQuery {
	query := (&JobRunnerClient{config: jrtq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	jrtq.withJobRunners = query
	return jrtq
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
//	client.JobRunnerToken.Query().
//		GroupBy(jobrunnertoken.FieldCreatedAt).
//		Aggregate(generated.Count()).
//		Scan(ctx, &v)
func (jrtq *JobRunnerTokenQuery) GroupBy(field string, fields ...string) *JobRunnerTokenGroupBy {
	jrtq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &JobRunnerTokenGroupBy{build: jrtq}
	grbuild.flds = &jrtq.ctx.Fields
	grbuild.label = jobrunnertoken.Label
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
//	client.JobRunnerToken.Query().
//		Select(jobrunnertoken.FieldCreatedAt).
//		Scan(ctx, &v)
func (jrtq *JobRunnerTokenQuery) Select(fields ...string) *JobRunnerTokenSelect {
	jrtq.ctx.Fields = append(jrtq.ctx.Fields, fields...)
	sbuild := &JobRunnerTokenSelect{JobRunnerTokenQuery: jrtq}
	sbuild.label = jobrunnertoken.Label
	sbuild.flds, sbuild.scan = &jrtq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a JobRunnerTokenSelect configured with the given aggregations.
func (jrtq *JobRunnerTokenQuery) Aggregate(fns ...AggregateFunc) *JobRunnerTokenSelect {
	return jrtq.Select().Aggregate(fns...)
}

func (jrtq *JobRunnerTokenQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range jrtq.inters {
		if inter == nil {
			return fmt.Errorf("generated: uninitialized interceptor (forgotten import generated/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, jrtq); err != nil {
				return err
			}
		}
	}
	for _, f := range jrtq.ctx.Fields {
		if !jobrunnertoken.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
		}
	}
	if jrtq.path != nil {
		prev, err := jrtq.path(ctx)
		if err != nil {
			return err
		}
		jrtq.sql = prev
	}
	if jobrunnertoken.Policy == nil {
		return errors.New("generated: uninitialized jobrunnertoken.Policy (forgotten import generated/runtime?)")
	}
	if err := jobrunnertoken.Policy.EvalQuery(ctx, jrtq); err != nil {
		return err
	}
	return nil
}

func (jrtq *JobRunnerTokenQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*JobRunnerToken, error) {
	var (
		nodes       = []*JobRunnerToken{}
		_spec       = jrtq.querySpec()
		loadedTypes = [2]bool{
			jrtq.withOwner != nil,
			jrtq.withJobRunners != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*JobRunnerToken).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &JobRunnerToken{config: jrtq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	_spec.Node.Schema = jrtq.schemaConfig.JobRunnerToken
	ctx = internal.NewSchemaConfigContext(ctx, jrtq.schemaConfig)
	if len(jrtq.modifiers) > 0 {
		_spec.Modifiers = jrtq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, jrtq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := jrtq.withOwner; query != nil {
		if err := jrtq.loadOwner(ctx, query, nodes, nil,
			func(n *JobRunnerToken, e *Organization) { n.Edges.Owner = e }); err != nil {
			return nil, err
		}
	}
	if query := jrtq.withJobRunners; query != nil {
		if err := jrtq.loadJobRunners(ctx, query, nodes,
			func(n *JobRunnerToken) { n.Edges.JobRunners = []*JobRunner{} },
			func(n *JobRunnerToken, e *JobRunner) { n.Edges.JobRunners = append(n.Edges.JobRunners, e) }); err != nil {
			return nil, err
		}
	}
	for name, query := range jrtq.withNamedJobRunners {
		if err := jrtq.loadJobRunners(ctx, query, nodes,
			func(n *JobRunnerToken) { n.appendNamedJobRunners(name) },
			func(n *JobRunnerToken, e *JobRunner) { n.appendNamedJobRunners(name, e) }); err != nil {
			return nil, err
		}
	}
	for i := range jrtq.loadTotal {
		if err := jrtq.loadTotal[i](ctx, nodes); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (jrtq *JobRunnerTokenQuery) loadOwner(ctx context.Context, query *OrganizationQuery, nodes []*JobRunnerToken, init func(*JobRunnerToken), assign func(*JobRunnerToken, *Organization)) error {
	ids := make([]string, 0, len(nodes))
	nodeids := make(map[string][]*JobRunnerToken)
	for i := range nodes {
		fk := nodes[i].OwnerID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(organization.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "owner_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (jrtq *JobRunnerTokenQuery) loadJobRunners(ctx context.Context, query *JobRunnerQuery, nodes []*JobRunnerToken, init func(*JobRunnerToken), assign func(*JobRunnerToken, *JobRunner)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[string]*JobRunnerToken)
	nids := make(map[string]map[*JobRunnerToken]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(jobrunnertoken.JobRunnersTable)
		joinT.Schema(jrtq.schemaConfig.JobRunnerJobRunnerTokens)
		s.Join(joinT).On(s.C(jobrunner.FieldID), joinT.C(jobrunnertoken.JobRunnersPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(jobrunnertoken.JobRunnersPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(jobrunnertoken.JobRunnersPrimaryKey[1]))
		s.AppendSelect(columns...)
		s.SetDistinct(false)
	})
	if err := query.prepareQuery(ctx); err != nil {
		return err
	}
	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return query.sqlAll(ctx, func(_ context.Context, spec *sqlgraph.QuerySpec) {
			assign := spec.Assign
			values := spec.ScanValues
			spec.ScanValues = func(columns []string) ([]any, error) {
				values, err := values(columns[1:])
				if err != nil {
					return nil, err
				}
				return append([]any{new(sql.NullString)}, values...), nil
			}
			spec.Assign = func(columns []string, values []any) error {
				outValue := values[0].(*sql.NullString).String
				inValue := values[1].(*sql.NullString).String
				if nids[inValue] == nil {
					nids[inValue] = map[*JobRunnerToken]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*JobRunner](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "job_runners" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}

func (jrtq *JobRunnerTokenQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := jrtq.querySpec()
	_spec.Node.Schema = jrtq.schemaConfig.JobRunnerToken
	ctx = internal.NewSchemaConfigContext(ctx, jrtq.schemaConfig)
	if len(jrtq.modifiers) > 0 {
		_spec.Modifiers = jrtq.modifiers
	}
	_spec.Node.Columns = jrtq.ctx.Fields
	if len(jrtq.ctx.Fields) > 0 {
		_spec.Unique = jrtq.ctx.Unique != nil && *jrtq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, jrtq.driver, _spec)
}

func (jrtq *JobRunnerTokenQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(jobrunnertoken.Table, jobrunnertoken.Columns, sqlgraph.NewFieldSpec(jobrunnertoken.FieldID, field.TypeString))
	_spec.From = jrtq.sql
	if unique := jrtq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if jrtq.path != nil {
		_spec.Unique = true
	}
	if fields := jrtq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, jobrunnertoken.FieldID)
		for i := range fields {
			if fields[i] != jobrunnertoken.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if jrtq.withOwner != nil {
			_spec.Node.AddColumnOnce(jobrunnertoken.FieldOwnerID)
		}
	}
	if ps := jrtq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := jrtq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := jrtq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := jrtq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (jrtq *JobRunnerTokenQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(jrtq.driver.Dialect())
	t1 := builder.Table(jobrunnertoken.Table)
	columns := jrtq.ctx.Fields
	if len(columns) == 0 {
		columns = jobrunnertoken.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if jrtq.sql != nil {
		selector = jrtq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if jrtq.ctx.Unique != nil && *jrtq.ctx.Unique {
		selector.Distinct()
	}
	t1.Schema(jrtq.schemaConfig.JobRunnerToken)
	ctx = internal.NewSchemaConfigContext(ctx, jrtq.schemaConfig)
	selector.WithContext(ctx)
	for _, m := range jrtq.modifiers {
		m(selector)
	}
	for _, p := range jrtq.predicates {
		p(selector)
	}
	for _, p := range jrtq.order {
		p(selector)
	}
	if offset := jrtq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := jrtq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (jrtq *JobRunnerTokenQuery) Modify(modifiers ...func(s *sql.Selector)) *JobRunnerTokenSelect {
	jrtq.modifiers = append(jrtq.modifiers, modifiers...)
	return jrtq.Select()
}

// WithNamedJobRunners tells the query-builder to eager-load the nodes that are connected to the "job_runners"
// edge with the given name. The optional arguments are used to configure the query builder of the edge.
func (jrtq *JobRunnerTokenQuery) WithNamedJobRunners(name string, opts ...func(*JobRunnerQuery)) *JobRunnerTokenQuery {
	query := (&JobRunnerClient{config: jrtq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	if jrtq.withNamedJobRunners == nil {
		jrtq.withNamedJobRunners = make(map[string]*JobRunnerQuery)
	}
	jrtq.withNamedJobRunners[name] = query
	return jrtq
}

// CountIDs returns the count of ids and allows for filtering of the query post retrieval by IDs
func (jrtq *JobRunnerTokenQuery) CountIDs(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, jrtq.ctx, ent.OpQueryIDs)
	if err := jrtq.prepareQuery(ctx); err != nil {
		return 0, err
	}

	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return jrtq.IDs(ctx)
	})

	ids, err := withInterceptors[[]string](ctx, jrtq, qr, jrtq.inters)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// JobRunnerTokenGroupBy is the group-by builder for JobRunnerToken entities.
type JobRunnerTokenGroupBy struct {
	selector
	build *JobRunnerTokenQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (jrtgb *JobRunnerTokenGroupBy) Aggregate(fns ...AggregateFunc) *JobRunnerTokenGroupBy {
	jrtgb.fns = append(jrtgb.fns, fns...)
	return jrtgb
}

// Scan applies the selector query and scans the result into the given value.
func (jrtgb *JobRunnerTokenGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, jrtgb.build.ctx, ent.OpQueryGroupBy)
	if err := jrtgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*JobRunnerTokenQuery, *JobRunnerTokenGroupBy](ctx, jrtgb.build, jrtgb, jrtgb.build.inters, v)
}

func (jrtgb *JobRunnerTokenGroupBy) sqlScan(ctx context.Context, root *JobRunnerTokenQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(jrtgb.fns))
	for _, fn := range jrtgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*jrtgb.flds)+len(jrtgb.fns))
		for _, f := range *jrtgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*jrtgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := jrtgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// JobRunnerTokenSelect is the builder for selecting fields of JobRunnerToken entities.
type JobRunnerTokenSelect struct {
	*JobRunnerTokenQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (jrts *JobRunnerTokenSelect) Aggregate(fns ...AggregateFunc) *JobRunnerTokenSelect {
	jrts.fns = append(jrts.fns, fns...)
	return jrts
}

// Scan applies the selector query and scans the result into the given value.
func (jrts *JobRunnerTokenSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, jrts.ctx, ent.OpQuerySelect)
	if err := jrts.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*JobRunnerTokenQuery, *JobRunnerTokenSelect](ctx, jrts.JobRunnerTokenQuery, jrts, jrts.inters, v)
}

func (jrts *JobRunnerTokenSelect) sqlScan(ctx context.Context, root *JobRunnerTokenQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(jrts.fns))
	for _, fn := range jrts.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*jrts.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := jrts.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (jrts *JobRunnerTokenSelect) Modify(modifiers ...func(s *sql.Selector)) *JobRunnerTokenSelect {
	jrts.modifiers = append(jrts.modifiers, modifiers...)
	return jrts
}
