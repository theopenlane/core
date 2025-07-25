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
	"github.com/theopenlane/core/internal/ent/generated/event"
	"github.com/theopenlane/core/internal/ent/generated/hush"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// HushQuery is the builder for querying Hush entities.
type HushQuery struct {
	config
	ctx                   *QueryContext
	order                 []hush.OrderOption
	inters                []Interceptor
	predicates            []predicate.Hush
	withOwner             *OrganizationQuery
	withIntegrations      *IntegrationQuery
	withEvents            *EventQuery
	loadTotal             []func(context.Context, []*Hush) error
	modifiers             []func(*sql.Selector)
	withNamedIntegrations map[string]*IntegrationQuery
	withNamedEvents       map[string]*EventQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the HushQuery builder.
func (hq *HushQuery) Where(ps ...predicate.Hush) *HushQuery {
	hq.predicates = append(hq.predicates, ps...)
	return hq
}

// Limit the number of records to be returned by this query.
func (hq *HushQuery) Limit(limit int) *HushQuery {
	hq.ctx.Limit = &limit
	return hq
}

// Offset to start from.
func (hq *HushQuery) Offset(offset int) *HushQuery {
	hq.ctx.Offset = &offset
	return hq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (hq *HushQuery) Unique(unique bool) *HushQuery {
	hq.ctx.Unique = &unique
	return hq
}

// Order specifies how the records should be ordered.
func (hq *HushQuery) Order(o ...hush.OrderOption) *HushQuery {
	hq.order = append(hq.order, o...)
	return hq
}

// QueryOwner chains the current query on the "owner" edge.
func (hq *HushQuery) QueryOwner() *OrganizationQuery {
	query := (&OrganizationClient{config: hq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := hq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := hq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(hush.Table, hush.FieldID, selector),
			sqlgraph.To(organization.Table, organization.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, hush.OwnerTable, hush.OwnerColumn),
		)
		schemaConfig := hq.schemaConfig
		step.To.Schema = schemaConfig.Organization
		step.Edge.Schema = schemaConfig.Hush
		fromU = sqlgraph.SetNeighbors(hq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryIntegrations chains the current query on the "integrations" edge.
func (hq *HushQuery) QueryIntegrations() *IntegrationQuery {
	query := (&IntegrationClient{config: hq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := hq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := hq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(hush.Table, hush.FieldID, selector),
			sqlgraph.To(integration.Table, integration.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, hush.IntegrationsTable, hush.IntegrationsPrimaryKey...),
		)
		schemaConfig := hq.schemaConfig
		step.To.Schema = schemaConfig.Integration
		step.Edge.Schema = schemaConfig.IntegrationSecrets
		fromU = sqlgraph.SetNeighbors(hq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryEvents chains the current query on the "events" edge.
func (hq *HushQuery) QueryEvents() *EventQuery {
	query := (&EventClient{config: hq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := hq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := hq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(hush.Table, hush.FieldID, selector),
			sqlgraph.To(event.Table, event.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, hush.EventsTable, hush.EventsPrimaryKey...),
		)
		schemaConfig := hq.schemaConfig
		step.To.Schema = schemaConfig.Event
		step.Edge.Schema = schemaConfig.HushEvents
		fromU = sqlgraph.SetNeighbors(hq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Hush entity from the query.
// Returns a *NotFoundError when no Hush was found.
func (hq *HushQuery) First(ctx context.Context) (*Hush, error) {
	nodes, err := hq.Limit(1).All(setContextOp(ctx, hq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{hush.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (hq *HushQuery) FirstX(ctx context.Context) *Hush {
	node, err := hq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Hush ID from the query.
// Returns a *NotFoundError when no Hush ID was found.
func (hq *HushQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = hq.Limit(1).IDs(setContextOp(ctx, hq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{hush.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (hq *HushQuery) FirstIDX(ctx context.Context) string {
	id, err := hq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Hush entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Hush entity is found.
// Returns a *NotFoundError when no Hush entities are found.
func (hq *HushQuery) Only(ctx context.Context) (*Hush, error) {
	nodes, err := hq.Limit(2).All(setContextOp(ctx, hq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{hush.Label}
	default:
		return nil, &NotSingularError{hush.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (hq *HushQuery) OnlyX(ctx context.Context) *Hush {
	node, err := hq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Hush ID in the query.
// Returns a *NotSingularError when more than one Hush ID is found.
// Returns a *NotFoundError when no entities are found.
func (hq *HushQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = hq.Limit(2).IDs(setContextOp(ctx, hq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{hush.Label}
	default:
		err = &NotSingularError{hush.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (hq *HushQuery) OnlyIDX(ctx context.Context) string {
	id, err := hq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Hushes.
func (hq *HushQuery) All(ctx context.Context) ([]*Hush, error) {
	ctx = setContextOp(ctx, hq.ctx, ent.OpQueryAll)
	if err := hq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Hush, *HushQuery]()
	return withInterceptors[[]*Hush](ctx, hq, qr, hq.inters)
}

// AllX is like All, but panics if an error occurs.
func (hq *HushQuery) AllX(ctx context.Context) []*Hush {
	nodes, err := hq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Hush IDs.
func (hq *HushQuery) IDs(ctx context.Context) (ids []string, err error) {
	if hq.ctx.Unique == nil && hq.path != nil {
		hq.Unique(true)
	}
	ctx = setContextOp(ctx, hq.ctx, ent.OpQueryIDs)
	if err = hq.Select(hush.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (hq *HushQuery) IDsX(ctx context.Context) []string {
	ids, err := hq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (hq *HushQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, hq.ctx, ent.OpQueryCount)
	if err := hq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, hq, querierCount[*HushQuery](), hq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (hq *HushQuery) CountX(ctx context.Context) int {
	count, err := hq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (hq *HushQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, hq.ctx, ent.OpQueryExist)
	switch _, err := hq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("generated: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (hq *HushQuery) ExistX(ctx context.Context) bool {
	exist, err := hq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the HushQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (hq *HushQuery) Clone() *HushQuery {
	if hq == nil {
		return nil
	}
	return &HushQuery{
		config:           hq.config,
		ctx:              hq.ctx.Clone(),
		order:            append([]hush.OrderOption{}, hq.order...),
		inters:           append([]Interceptor{}, hq.inters...),
		predicates:       append([]predicate.Hush{}, hq.predicates...),
		withOwner:        hq.withOwner.Clone(),
		withIntegrations: hq.withIntegrations.Clone(),
		withEvents:       hq.withEvents.Clone(),
		// clone intermediate query.
		sql:       hq.sql.Clone(),
		path:      hq.path,
		modifiers: append([]func(*sql.Selector){}, hq.modifiers...),
	}
}

// WithOwner tells the query-builder to eager-load the nodes that are connected to
// the "owner" edge. The optional arguments are used to configure the query builder of the edge.
func (hq *HushQuery) WithOwner(opts ...func(*OrganizationQuery)) *HushQuery {
	query := (&OrganizationClient{config: hq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	hq.withOwner = query
	return hq
}

// WithIntegrations tells the query-builder to eager-load the nodes that are connected to
// the "integrations" edge. The optional arguments are used to configure the query builder of the edge.
func (hq *HushQuery) WithIntegrations(opts ...func(*IntegrationQuery)) *HushQuery {
	query := (&IntegrationClient{config: hq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	hq.withIntegrations = query
	return hq
}

// WithEvents tells the query-builder to eager-load the nodes that are connected to
// the "events" edge. The optional arguments are used to configure the query builder of the edge.
func (hq *HushQuery) WithEvents(opts ...func(*EventQuery)) *HushQuery {
	query := (&EventClient{config: hq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	hq.withEvents = query
	return hq
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
//	client.Hush.Query().
//		GroupBy(hush.FieldCreatedAt).
//		Aggregate(generated.Count()).
//		Scan(ctx, &v)
func (hq *HushQuery) GroupBy(field string, fields ...string) *HushGroupBy {
	hq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &HushGroupBy{build: hq}
	grbuild.flds = &hq.ctx.Fields
	grbuild.label = hush.Label
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
//	client.Hush.Query().
//		Select(hush.FieldCreatedAt).
//		Scan(ctx, &v)
func (hq *HushQuery) Select(fields ...string) *HushSelect {
	hq.ctx.Fields = append(hq.ctx.Fields, fields...)
	sbuild := &HushSelect{HushQuery: hq}
	sbuild.label = hush.Label
	sbuild.flds, sbuild.scan = &hq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a HushSelect configured with the given aggregations.
func (hq *HushQuery) Aggregate(fns ...AggregateFunc) *HushSelect {
	return hq.Select().Aggregate(fns...)
}

func (hq *HushQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range hq.inters {
		if inter == nil {
			return fmt.Errorf("generated: uninitialized interceptor (forgotten import generated/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, hq); err != nil {
				return err
			}
		}
	}
	for _, f := range hq.ctx.Fields {
		if !hush.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
		}
	}
	if hq.path != nil {
		prev, err := hq.path(ctx)
		if err != nil {
			return err
		}
		hq.sql = prev
	}
	if hush.Policy == nil {
		return errors.New("generated: uninitialized hush.Policy (forgotten import generated/runtime?)")
	}
	if err := hush.Policy.EvalQuery(ctx, hq); err != nil {
		return err
	}
	return nil
}

func (hq *HushQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Hush, error) {
	var (
		nodes       = []*Hush{}
		_spec       = hq.querySpec()
		loadedTypes = [3]bool{
			hq.withOwner != nil,
			hq.withIntegrations != nil,
			hq.withEvents != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Hush).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Hush{config: hq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	_spec.Node.Schema = hq.schemaConfig.Hush
	ctx = internal.NewSchemaConfigContext(ctx, hq.schemaConfig)
	if len(hq.modifiers) > 0 {
		_spec.Modifiers = hq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, hq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := hq.withOwner; query != nil {
		if err := hq.loadOwner(ctx, query, nodes, nil,
			func(n *Hush, e *Organization) { n.Edges.Owner = e }); err != nil {
			return nil, err
		}
	}
	if query := hq.withIntegrations; query != nil {
		if err := hq.loadIntegrations(ctx, query, nodes,
			func(n *Hush) { n.Edges.Integrations = []*Integration{} },
			func(n *Hush, e *Integration) { n.Edges.Integrations = append(n.Edges.Integrations, e) }); err != nil {
			return nil, err
		}
	}
	if query := hq.withEvents; query != nil {
		if err := hq.loadEvents(ctx, query, nodes,
			func(n *Hush) { n.Edges.Events = []*Event{} },
			func(n *Hush, e *Event) { n.Edges.Events = append(n.Edges.Events, e) }); err != nil {
			return nil, err
		}
	}
	for name, query := range hq.withNamedIntegrations {
		if err := hq.loadIntegrations(ctx, query, nodes,
			func(n *Hush) { n.appendNamedIntegrations(name) },
			func(n *Hush, e *Integration) { n.appendNamedIntegrations(name, e) }); err != nil {
			return nil, err
		}
	}
	for name, query := range hq.withNamedEvents {
		if err := hq.loadEvents(ctx, query, nodes,
			func(n *Hush) { n.appendNamedEvents(name) },
			func(n *Hush, e *Event) { n.appendNamedEvents(name, e) }); err != nil {
			return nil, err
		}
	}
	for i := range hq.loadTotal {
		if err := hq.loadTotal[i](ctx, nodes); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (hq *HushQuery) loadOwner(ctx context.Context, query *OrganizationQuery, nodes []*Hush, init func(*Hush), assign func(*Hush, *Organization)) error {
	ids := make([]string, 0, len(nodes))
	nodeids := make(map[string][]*Hush)
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
func (hq *HushQuery) loadIntegrations(ctx context.Context, query *IntegrationQuery, nodes []*Hush, init func(*Hush), assign func(*Hush, *Integration)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[string]*Hush)
	nids := make(map[string]map[*Hush]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(hush.IntegrationsTable)
		joinT.Schema(hq.schemaConfig.IntegrationSecrets)
		s.Join(joinT).On(s.C(integration.FieldID), joinT.C(hush.IntegrationsPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(hush.IntegrationsPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(hush.IntegrationsPrimaryKey[1]))
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
					nids[inValue] = map[*Hush]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*Integration](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "integrations" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}
func (hq *HushQuery) loadEvents(ctx context.Context, query *EventQuery, nodes []*Hush, init func(*Hush), assign func(*Hush, *Event)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[string]*Hush)
	nids := make(map[string]map[*Hush]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(hush.EventsTable)
		joinT.Schema(hq.schemaConfig.HushEvents)
		s.Join(joinT).On(s.C(event.FieldID), joinT.C(hush.EventsPrimaryKey[1]))
		s.Where(sql.InValues(joinT.C(hush.EventsPrimaryKey[0]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(hush.EventsPrimaryKey[0]))
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
					nids[inValue] = map[*Hush]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*Event](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "events" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}

func (hq *HushQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := hq.querySpec()
	_spec.Node.Schema = hq.schemaConfig.Hush
	ctx = internal.NewSchemaConfigContext(ctx, hq.schemaConfig)
	if len(hq.modifiers) > 0 {
		_spec.Modifiers = hq.modifiers
	}
	_spec.Node.Columns = hq.ctx.Fields
	if len(hq.ctx.Fields) > 0 {
		_spec.Unique = hq.ctx.Unique != nil && *hq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, hq.driver, _spec)
}

func (hq *HushQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(hush.Table, hush.Columns, sqlgraph.NewFieldSpec(hush.FieldID, field.TypeString))
	_spec.From = hq.sql
	if unique := hq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if hq.path != nil {
		_spec.Unique = true
	}
	if fields := hq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, hush.FieldID)
		for i := range fields {
			if fields[i] != hush.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if hq.withOwner != nil {
			_spec.Node.AddColumnOnce(hush.FieldOwnerID)
		}
	}
	if ps := hq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := hq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := hq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := hq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (hq *HushQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(hq.driver.Dialect())
	t1 := builder.Table(hush.Table)
	columns := hq.ctx.Fields
	if len(columns) == 0 {
		columns = hush.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if hq.sql != nil {
		selector = hq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if hq.ctx.Unique != nil && *hq.ctx.Unique {
		selector.Distinct()
	}
	t1.Schema(hq.schemaConfig.Hush)
	ctx = internal.NewSchemaConfigContext(ctx, hq.schemaConfig)
	selector.WithContext(ctx)
	for _, m := range hq.modifiers {
		m(selector)
	}
	for _, p := range hq.predicates {
		p(selector)
	}
	for _, p := range hq.order {
		p(selector)
	}
	if offset := hq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := hq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (hq *HushQuery) Modify(modifiers ...func(s *sql.Selector)) *HushSelect {
	hq.modifiers = append(hq.modifiers, modifiers...)
	return hq.Select()
}

// WithNamedIntegrations tells the query-builder to eager-load the nodes that are connected to the "integrations"
// edge with the given name. The optional arguments are used to configure the query builder of the edge.
func (hq *HushQuery) WithNamedIntegrations(name string, opts ...func(*IntegrationQuery)) *HushQuery {
	query := (&IntegrationClient{config: hq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	if hq.withNamedIntegrations == nil {
		hq.withNamedIntegrations = make(map[string]*IntegrationQuery)
	}
	hq.withNamedIntegrations[name] = query
	return hq
}

// WithNamedEvents tells the query-builder to eager-load the nodes that are connected to the "events"
// edge with the given name. The optional arguments are used to configure the query builder of the edge.
func (hq *HushQuery) WithNamedEvents(name string, opts ...func(*EventQuery)) *HushQuery {
	query := (&EventClient{config: hq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	if hq.withNamedEvents == nil {
		hq.withNamedEvents = make(map[string]*EventQuery)
	}
	hq.withNamedEvents[name] = query
	return hq
}

// CountIDs returns the count of ids and allows for filtering of the query post retrieval by IDs
func (hq *HushQuery) CountIDs(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, hq.ctx, ent.OpQueryIDs)
	if err := hq.prepareQuery(ctx); err != nil {
		return 0, err
	}

	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return hq.IDs(ctx)
	})

	ids, err := withInterceptors[[]string](ctx, hq, qr, hq.inters)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// HushGroupBy is the group-by builder for Hush entities.
type HushGroupBy struct {
	selector
	build *HushQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (hgb *HushGroupBy) Aggregate(fns ...AggregateFunc) *HushGroupBy {
	hgb.fns = append(hgb.fns, fns...)
	return hgb
}

// Scan applies the selector query and scans the result into the given value.
func (hgb *HushGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, hgb.build.ctx, ent.OpQueryGroupBy)
	if err := hgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*HushQuery, *HushGroupBy](ctx, hgb.build, hgb, hgb.build.inters, v)
}

func (hgb *HushGroupBy) sqlScan(ctx context.Context, root *HushQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(hgb.fns))
	for _, fn := range hgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*hgb.flds)+len(hgb.fns))
		for _, f := range *hgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*hgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := hgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// HushSelect is the builder for selecting fields of Hush entities.
type HushSelect struct {
	*HushQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (hs *HushSelect) Aggregate(fns ...AggregateFunc) *HushSelect {
	hs.fns = append(hs.fns, fns...)
	return hs
}

// Scan applies the selector query and scans the result into the given value.
func (hs *HushSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, hs.ctx, ent.OpQuerySelect)
	if err := hs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*HushQuery, *HushSelect](ctx, hs.HushQuery, hs, hs.inters, v)
}

func (hs *HushSelect) sqlScan(ctx context.Context, root *HushQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(hs.fns))
	for _, fn := range hs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*hs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := hs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (hs *HushSelect) Modify(modifiers ...func(s *sql.Selector)) *HushSelect {
	hs.modifiers = append(hs.modifiers, modifiers...)
	return hs
}
