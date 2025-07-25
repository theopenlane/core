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
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// TrustCenterSettingQuery is the builder for querying TrustCenterSetting entities.
type TrustCenterSettingQuery struct {
	config
	ctx             *QueryContext
	order           []trustcentersetting.OrderOption
	inters          []Interceptor
	predicates      []predicate.TrustCenterSetting
	withTrustCenter *TrustCenterQuery
	withFiles       *FileQuery
	withLogoFile    *FileQuery
	withFaviconFile *FileQuery
	loadTotal       []func(context.Context, []*TrustCenterSetting) error
	modifiers       []func(*sql.Selector)
	withNamedFiles  map[string]*FileQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the TrustCenterSettingQuery builder.
func (tcsq *TrustCenterSettingQuery) Where(ps ...predicate.TrustCenterSetting) *TrustCenterSettingQuery {
	tcsq.predicates = append(tcsq.predicates, ps...)
	return tcsq
}

// Limit the number of records to be returned by this query.
func (tcsq *TrustCenterSettingQuery) Limit(limit int) *TrustCenterSettingQuery {
	tcsq.ctx.Limit = &limit
	return tcsq
}

// Offset to start from.
func (tcsq *TrustCenterSettingQuery) Offset(offset int) *TrustCenterSettingQuery {
	tcsq.ctx.Offset = &offset
	return tcsq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (tcsq *TrustCenterSettingQuery) Unique(unique bool) *TrustCenterSettingQuery {
	tcsq.ctx.Unique = &unique
	return tcsq
}

// Order specifies how the records should be ordered.
func (tcsq *TrustCenterSettingQuery) Order(o ...trustcentersetting.OrderOption) *TrustCenterSettingQuery {
	tcsq.order = append(tcsq.order, o...)
	return tcsq
}

// QueryTrustCenter chains the current query on the "trust_center" edge.
func (tcsq *TrustCenterSettingQuery) QueryTrustCenter() *TrustCenterQuery {
	query := (&TrustCenterClient{config: tcsq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := tcsq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := tcsq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(trustcentersetting.Table, trustcentersetting.FieldID, selector),
			sqlgraph.To(trustcenter.Table, trustcenter.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, trustcentersetting.TrustCenterTable, trustcentersetting.TrustCenterColumn),
		)
		schemaConfig := tcsq.schemaConfig
		step.To.Schema = schemaConfig.TrustCenter
		step.Edge.Schema = schemaConfig.TrustCenterSetting
		fromU = sqlgraph.SetNeighbors(tcsq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryFiles chains the current query on the "files" edge.
func (tcsq *TrustCenterSettingQuery) QueryFiles() *FileQuery {
	query := (&FileClient{config: tcsq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := tcsq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := tcsq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(trustcentersetting.Table, trustcentersetting.FieldID, selector),
			sqlgraph.To(file.Table, file.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, trustcentersetting.FilesTable, trustcentersetting.FilesPrimaryKey...),
		)
		schemaConfig := tcsq.schemaConfig
		step.To.Schema = schemaConfig.File
		step.Edge.Schema = schemaConfig.TrustCenterSettingFiles
		fromU = sqlgraph.SetNeighbors(tcsq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryLogoFile chains the current query on the "logo_file" edge.
func (tcsq *TrustCenterSettingQuery) QueryLogoFile() *FileQuery {
	query := (&FileClient{config: tcsq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := tcsq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := tcsq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(trustcentersetting.Table, trustcentersetting.FieldID, selector),
			sqlgraph.To(file.Table, file.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, trustcentersetting.LogoFileTable, trustcentersetting.LogoFileColumn),
		)
		schemaConfig := tcsq.schemaConfig
		step.To.Schema = schemaConfig.File
		step.Edge.Schema = schemaConfig.TrustCenterSetting
		fromU = sqlgraph.SetNeighbors(tcsq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryFaviconFile chains the current query on the "favicon_file" edge.
func (tcsq *TrustCenterSettingQuery) QueryFaviconFile() *FileQuery {
	query := (&FileClient{config: tcsq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := tcsq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := tcsq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(trustcentersetting.Table, trustcentersetting.FieldID, selector),
			sqlgraph.To(file.Table, file.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, trustcentersetting.FaviconFileTable, trustcentersetting.FaviconFileColumn),
		)
		schemaConfig := tcsq.schemaConfig
		step.To.Schema = schemaConfig.File
		step.Edge.Schema = schemaConfig.TrustCenterSetting
		fromU = sqlgraph.SetNeighbors(tcsq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first TrustCenterSetting entity from the query.
// Returns a *NotFoundError when no TrustCenterSetting was found.
func (tcsq *TrustCenterSettingQuery) First(ctx context.Context) (*TrustCenterSetting, error) {
	nodes, err := tcsq.Limit(1).All(setContextOp(ctx, tcsq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{trustcentersetting.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (tcsq *TrustCenterSettingQuery) FirstX(ctx context.Context) *TrustCenterSetting {
	node, err := tcsq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first TrustCenterSetting ID from the query.
// Returns a *NotFoundError when no TrustCenterSetting ID was found.
func (tcsq *TrustCenterSettingQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = tcsq.Limit(1).IDs(setContextOp(ctx, tcsq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{trustcentersetting.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (tcsq *TrustCenterSettingQuery) FirstIDX(ctx context.Context) string {
	id, err := tcsq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single TrustCenterSetting entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one TrustCenterSetting entity is found.
// Returns a *NotFoundError when no TrustCenterSetting entities are found.
func (tcsq *TrustCenterSettingQuery) Only(ctx context.Context) (*TrustCenterSetting, error) {
	nodes, err := tcsq.Limit(2).All(setContextOp(ctx, tcsq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{trustcentersetting.Label}
	default:
		return nil, &NotSingularError{trustcentersetting.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (tcsq *TrustCenterSettingQuery) OnlyX(ctx context.Context) *TrustCenterSetting {
	node, err := tcsq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only TrustCenterSetting ID in the query.
// Returns a *NotSingularError when more than one TrustCenterSetting ID is found.
// Returns a *NotFoundError when no entities are found.
func (tcsq *TrustCenterSettingQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = tcsq.Limit(2).IDs(setContextOp(ctx, tcsq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{trustcentersetting.Label}
	default:
		err = &NotSingularError{trustcentersetting.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (tcsq *TrustCenterSettingQuery) OnlyIDX(ctx context.Context) string {
	id, err := tcsq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of TrustCenterSettings.
func (tcsq *TrustCenterSettingQuery) All(ctx context.Context) ([]*TrustCenterSetting, error) {
	ctx = setContextOp(ctx, tcsq.ctx, ent.OpQueryAll)
	if err := tcsq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*TrustCenterSetting, *TrustCenterSettingQuery]()
	return withInterceptors[[]*TrustCenterSetting](ctx, tcsq, qr, tcsq.inters)
}

// AllX is like All, but panics if an error occurs.
func (tcsq *TrustCenterSettingQuery) AllX(ctx context.Context) []*TrustCenterSetting {
	nodes, err := tcsq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of TrustCenterSetting IDs.
func (tcsq *TrustCenterSettingQuery) IDs(ctx context.Context) (ids []string, err error) {
	if tcsq.ctx.Unique == nil && tcsq.path != nil {
		tcsq.Unique(true)
	}
	ctx = setContextOp(ctx, tcsq.ctx, ent.OpQueryIDs)
	if err = tcsq.Select(trustcentersetting.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (tcsq *TrustCenterSettingQuery) IDsX(ctx context.Context) []string {
	ids, err := tcsq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (tcsq *TrustCenterSettingQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, tcsq.ctx, ent.OpQueryCount)
	if err := tcsq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, tcsq, querierCount[*TrustCenterSettingQuery](), tcsq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (tcsq *TrustCenterSettingQuery) CountX(ctx context.Context) int {
	count, err := tcsq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (tcsq *TrustCenterSettingQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, tcsq.ctx, ent.OpQueryExist)
	switch _, err := tcsq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("generated: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (tcsq *TrustCenterSettingQuery) ExistX(ctx context.Context) bool {
	exist, err := tcsq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the TrustCenterSettingQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (tcsq *TrustCenterSettingQuery) Clone() *TrustCenterSettingQuery {
	if tcsq == nil {
		return nil
	}
	return &TrustCenterSettingQuery{
		config:          tcsq.config,
		ctx:             tcsq.ctx.Clone(),
		order:           append([]trustcentersetting.OrderOption{}, tcsq.order...),
		inters:          append([]Interceptor{}, tcsq.inters...),
		predicates:      append([]predicate.TrustCenterSetting{}, tcsq.predicates...),
		withTrustCenter: tcsq.withTrustCenter.Clone(),
		withFiles:       tcsq.withFiles.Clone(),
		withLogoFile:    tcsq.withLogoFile.Clone(),
		withFaviconFile: tcsq.withFaviconFile.Clone(),
		// clone intermediate query.
		sql:       tcsq.sql.Clone(),
		path:      tcsq.path,
		modifiers: append([]func(*sql.Selector){}, tcsq.modifiers...),
	}
}

// WithTrustCenter tells the query-builder to eager-load the nodes that are connected to
// the "trust_center" edge. The optional arguments are used to configure the query builder of the edge.
func (tcsq *TrustCenterSettingQuery) WithTrustCenter(opts ...func(*TrustCenterQuery)) *TrustCenterSettingQuery {
	query := (&TrustCenterClient{config: tcsq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	tcsq.withTrustCenter = query
	return tcsq
}

// WithFiles tells the query-builder to eager-load the nodes that are connected to
// the "files" edge. The optional arguments are used to configure the query builder of the edge.
func (tcsq *TrustCenterSettingQuery) WithFiles(opts ...func(*FileQuery)) *TrustCenterSettingQuery {
	query := (&FileClient{config: tcsq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	tcsq.withFiles = query
	return tcsq
}

// WithLogoFile tells the query-builder to eager-load the nodes that are connected to
// the "logo_file" edge. The optional arguments are used to configure the query builder of the edge.
func (tcsq *TrustCenterSettingQuery) WithLogoFile(opts ...func(*FileQuery)) *TrustCenterSettingQuery {
	query := (&FileClient{config: tcsq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	tcsq.withLogoFile = query
	return tcsq
}

// WithFaviconFile tells the query-builder to eager-load the nodes that are connected to
// the "favicon_file" edge. The optional arguments are used to configure the query builder of the edge.
func (tcsq *TrustCenterSettingQuery) WithFaviconFile(opts ...func(*FileQuery)) *TrustCenterSettingQuery {
	query := (&FileClient{config: tcsq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	tcsq.withFaviconFile = query
	return tcsq
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
//	client.TrustCenterSetting.Query().
//		GroupBy(trustcentersetting.FieldCreatedAt).
//		Aggregate(generated.Count()).
//		Scan(ctx, &v)
func (tcsq *TrustCenterSettingQuery) GroupBy(field string, fields ...string) *TrustCenterSettingGroupBy {
	tcsq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &TrustCenterSettingGroupBy{build: tcsq}
	grbuild.flds = &tcsq.ctx.Fields
	grbuild.label = trustcentersetting.Label
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
//	client.TrustCenterSetting.Query().
//		Select(trustcentersetting.FieldCreatedAt).
//		Scan(ctx, &v)
func (tcsq *TrustCenterSettingQuery) Select(fields ...string) *TrustCenterSettingSelect {
	tcsq.ctx.Fields = append(tcsq.ctx.Fields, fields...)
	sbuild := &TrustCenterSettingSelect{TrustCenterSettingQuery: tcsq}
	sbuild.label = trustcentersetting.Label
	sbuild.flds, sbuild.scan = &tcsq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a TrustCenterSettingSelect configured with the given aggregations.
func (tcsq *TrustCenterSettingQuery) Aggregate(fns ...AggregateFunc) *TrustCenterSettingSelect {
	return tcsq.Select().Aggregate(fns...)
}

func (tcsq *TrustCenterSettingQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range tcsq.inters {
		if inter == nil {
			return fmt.Errorf("generated: uninitialized interceptor (forgotten import generated/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, tcsq); err != nil {
				return err
			}
		}
	}
	for _, f := range tcsq.ctx.Fields {
		if !trustcentersetting.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
		}
	}
	if tcsq.path != nil {
		prev, err := tcsq.path(ctx)
		if err != nil {
			return err
		}
		tcsq.sql = prev
	}
	if trustcentersetting.Policy == nil {
		return errors.New("generated: uninitialized trustcentersetting.Policy (forgotten import generated/runtime?)")
	}
	if err := trustcentersetting.Policy.EvalQuery(ctx, tcsq); err != nil {
		return err
	}
	return nil
}

func (tcsq *TrustCenterSettingQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*TrustCenterSetting, error) {
	var (
		nodes       = []*TrustCenterSetting{}
		_spec       = tcsq.querySpec()
		loadedTypes = [4]bool{
			tcsq.withTrustCenter != nil,
			tcsq.withFiles != nil,
			tcsq.withLogoFile != nil,
			tcsq.withFaviconFile != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*TrustCenterSetting).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &TrustCenterSetting{config: tcsq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	_spec.Node.Schema = tcsq.schemaConfig.TrustCenterSetting
	ctx = internal.NewSchemaConfigContext(ctx, tcsq.schemaConfig)
	if len(tcsq.modifiers) > 0 {
		_spec.Modifiers = tcsq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, tcsq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := tcsq.withTrustCenter; query != nil {
		if err := tcsq.loadTrustCenter(ctx, query, nodes, nil,
			func(n *TrustCenterSetting, e *TrustCenter) { n.Edges.TrustCenter = e }); err != nil {
			return nil, err
		}
	}
	if query := tcsq.withFiles; query != nil {
		if err := tcsq.loadFiles(ctx, query, nodes,
			func(n *TrustCenterSetting) { n.Edges.Files = []*File{} },
			func(n *TrustCenterSetting, e *File) { n.Edges.Files = append(n.Edges.Files, e) }); err != nil {
			return nil, err
		}
	}
	if query := tcsq.withLogoFile; query != nil {
		if err := tcsq.loadLogoFile(ctx, query, nodes, nil,
			func(n *TrustCenterSetting, e *File) { n.Edges.LogoFile = e }); err != nil {
			return nil, err
		}
	}
	if query := tcsq.withFaviconFile; query != nil {
		if err := tcsq.loadFaviconFile(ctx, query, nodes, nil,
			func(n *TrustCenterSetting, e *File) { n.Edges.FaviconFile = e }); err != nil {
			return nil, err
		}
	}
	for name, query := range tcsq.withNamedFiles {
		if err := tcsq.loadFiles(ctx, query, nodes,
			func(n *TrustCenterSetting) { n.appendNamedFiles(name) },
			func(n *TrustCenterSetting, e *File) { n.appendNamedFiles(name, e) }); err != nil {
			return nil, err
		}
	}
	for i := range tcsq.loadTotal {
		if err := tcsq.loadTotal[i](ctx, nodes); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (tcsq *TrustCenterSettingQuery) loadTrustCenter(ctx context.Context, query *TrustCenterQuery, nodes []*TrustCenterSetting, init func(*TrustCenterSetting), assign func(*TrustCenterSetting, *TrustCenter)) error {
	ids := make([]string, 0, len(nodes))
	nodeids := make(map[string][]*TrustCenterSetting)
	for i := range nodes {
		fk := nodes[i].TrustCenterID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(trustcenter.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "trust_center_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (tcsq *TrustCenterSettingQuery) loadFiles(ctx context.Context, query *FileQuery, nodes []*TrustCenterSetting, init func(*TrustCenterSetting), assign func(*TrustCenterSetting, *File)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[string]*TrustCenterSetting)
	nids := make(map[string]map[*TrustCenterSetting]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(trustcentersetting.FilesTable)
		joinT.Schema(tcsq.schemaConfig.TrustCenterSettingFiles)
		s.Join(joinT).On(s.C(file.FieldID), joinT.C(trustcentersetting.FilesPrimaryKey[1]))
		s.Where(sql.InValues(joinT.C(trustcentersetting.FilesPrimaryKey[0]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(trustcentersetting.FilesPrimaryKey[0]))
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
					nids[inValue] = map[*TrustCenterSetting]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*File](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "files" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}
func (tcsq *TrustCenterSettingQuery) loadLogoFile(ctx context.Context, query *FileQuery, nodes []*TrustCenterSetting, init func(*TrustCenterSetting), assign func(*TrustCenterSetting, *File)) error {
	ids := make([]string, 0, len(nodes))
	nodeids := make(map[string][]*TrustCenterSetting)
	for i := range nodes {
		if nodes[i].LogoLocalFileID == nil {
			continue
		}
		fk := *nodes[i].LogoLocalFileID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(file.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "logo_local_file_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (tcsq *TrustCenterSettingQuery) loadFaviconFile(ctx context.Context, query *FileQuery, nodes []*TrustCenterSetting, init func(*TrustCenterSetting), assign func(*TrustCenterSetting, *File)) error {
	ids := make([]string, 0, len(nodes))
	nodeids := make(map[string][]*TrustCenterSetting)
	for i := range nodes {
		if nodes[i].FaviconLocalFileID == nil {
			continue
		}
		fk := *nodes[i].FaviconLocalFileID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(file.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "favicon_local_file_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (tcsq *TrustCenterSettingQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := tcsq.querySpec()
	_spec.Node.Schema = tcsq.schemaConfig.TrustCenterSetting
	ctx = internal.NewSchemaConfigContext(ctx, tcsq.schemaConfig)
	if len(tcsq.modifiers) > 0 {
		_spec.Modifiers = tcsq.modifiers
	}
	_spec.Node.Columns = tcsq.ctx.Fields
	if len(tcsq.ctx.Fields) > 0 {
		_spec.Unique = tcsq.ctx.Unique != nil && *tcsq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, tcsq.driver, _spec)
}

func (tcsq *TrustCenterSettingQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(trustcentersetting.Table, trustcentersetting.Columns, sqlgraph.NewFieldSpec(trustcentersetting.FieldID, field.TypeString))
	_spec.From = tcsq.sql
	if unique := tcsq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if tcsq.path != nil {
		_spec.Unique = true
	}
	if fields := tcsq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, trustcentersetting.FieldID)
		for i := range fields {
			if fields[i] != trustcentersetting.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if tcsq.withTrustCenter != nil {
			_spec.Node.AddColumnOnce(trustcentersetting.FieldTrustCenterID)
		}
		if tcsq.withLogoFile != nil {
			_spec.Node.AddColumnOnce(trustcentersetting.FieldLogoLocalFileID)
		}
		if tcsq.withFaviconFile != nil {
			_spec.Node.AddColumnOnce(trustcentersetting.FieldFaviconLocalFileID)
		}
	}
	if ps := tcsq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := tcsq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := tcsq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := tcsq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (tcsq *TrustCenterSettingQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(tcsq.driver.Dialect())
	t1 := builder.Table(trustcentersetting.Table)
	columns := tcsq.ctx.Fields
	if len(columns) == 0 {
		columns = trustcentersetting.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if tcsq.sql != nil {
		selector = tcsq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if tcsq.ctx.Unique != nil && *tcsq.ctx.Unique {
		selector.Distinct()
	}
	t1.Schema(tcsq.schemaConfig.TrustCenterSetting)
	ctx = internal.NewSchemaConfigContext(ctx, tcsq.schemaConfig)
	selector.WithContext(ctx)
	for _, m := range tcsq.modifiers {
		m(selector)
	}
	for _, p := range tcsq.predicates {
		p(selector)
	}
	for _, p := range tcsq.order {
		p(selector)
	}
	if offset := tcsq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := tcsq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (tcsq *TrustCenterSettingQuery) Modify(modifiers ...func(s *sql.Selector)) *TrustCenterSettingSelect {
	tcsq.modifiers = append(tcsq.modifiers, modifiers...)
	return tcsq.Select()
}

// WithNamedFiles tells the query-builder to eager-load the nodes that are connected to the "files"
// edge with the given name. The optional arguments are used to configure the query builder of the edge.
func (tcsq *TrustCenterSettingQuery) WithNamedFiles(name string, opts ...func(*FileQuery)) *TrustCenterSettingQuery {
	query := (&FileClient{config: tcsq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	if tcsq.withNamedFiles == nil {
		tcsq.withNamedFiles = make(map[string]*FileQuery)
	}
	tcsq.withNamedFiles[name] = query
	return tcsq
}

// CountIDs returns the count of ids and allows for filtering of the query post retrieval by IDs
func (tcsq *TrustCenterSettingQuery) CountIDs(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, tcsq.ctx, ent.OpQueryIDs)
	if err := tcsq.prepareQuery(ctx); err != nil {
		return 0, err
	}

	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return tcsq.IDs(ctx)
	})

	ids, err := withInterceptors[[]string](ctx, tcsq, qr, tcsq.inters)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// TrustCenterSettingGroupBy is the group-by builder for TrustCenterSetting entities.
type TrustCenterSettingGroupBy struct {
	selector
	build *TrustCenterSettingQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (tcsgb *TrustCenterSettingGroupBy) Aggregate(fns ...AggregateFunc) *TrustCenterSettingGroupBy {
	tcsgb.fns = append(tcsgb.fns, fns...)
	return tcsgb
}

// Scan applies the selector query and scans the result into the given value.
func (tcsgb *TrustCenterSettingGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, tcsgb.build.ctx, ent.OpQueryGroupBy)
	if err := tcsgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TrustCenterSettingQuery, *TrustCenterSettingGroupBy](ctx, tcsgb.build, tcsgb, tcsgb.build.inters, v)
}

func (tcsgb *TrustCenterSettingGroupBy) sqlScan(ctx context.Context, root *TrustCenterSettingQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(tcsgb.fns))
	for _, fn := range tcsgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*tcsgb.flds)+len(tcsgb.fns))
		for _, f := range *tcsgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*tcsgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := tcsgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// TrustCenterSettingSelect is the builder for selecting fields of TrustCenterSetting entities.
type TrustCenterSettingSelect struct {
	*TrustCenterSettingQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (tcss *TrustCenterSettingSelect) Aggregate(fns ...AggregateFunc) *TrustCenterSettingSelect {
	tcss.fns = append(tcss.fns, fns...)
	return tcss
}

// Scan applies the selector query and scans the result into the given value.
func (tcss *TrustCenterSettingSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, tcss.ctx, ent.OpQuerySelect)
	if err := tcss.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*TrustCenterSettingQuery, *TrustCenterSettingSelect](ctx, tcss.TrustCenterSettingQuery, tcss, tcss.inters, v)
}

func (tcss *TrustCenterSettingSelect) sqlScan(ctx context.Context, root *TrustCenterSettingQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(tcss.fns))
	for _, fn := range tcss.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*tcss.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := tcss.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (tcss *TrustCenterSettingSelect) Modify(modifiers ...func(s *sql.Selector)) *TrustCenterSettingSelect {
	tcss.modifiers = append(tcss.modifiers, modifiers...)
	return tcss
}
