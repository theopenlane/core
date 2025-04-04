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
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/task"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// NoteQuery is the builder for querying Note entities.
type NoteQuery struct {
	config
	ctx            *QueryContext
	order          []note.OrderOption
	inters         []Interceptor
	predicates     []predicate.Note
	withOwner      *OrganizationQuery
	withTask       *TaskQuery
	withFiles      *FileQuery
	withFKs        bool
	loadTotal      []func(context.Context, []*Note) error
	modifiers      []func(*sql.Selector)
	withNamedFiles map[string]*FileQuery
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the NoteQuery builder.
func (nq *NoteQuery) Where(ps ...predicate.Note) *NoteQuery {
	nq.predicates = append(nq.predicates, ps...)
	return nq
}

// Limit the number of records to be returned by this query.
func (nq *NoteQuery) Limit(limit int) *NoteQuery {
	nq.ctx.Limit = &limit
	return nq
}

// Offset to start from.
func (nq *NoteQuery) Offset(offset int) *NoteQuery {
	nq.ctx.Offset = &offset
	return nq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (nq *NoteQuery) Unique(unique bool) *NoteQuery {
	nq.ctx.Unique = &unique
	return nq
}

// Order specifies how the records should be ordered.
func (nq *NoteQuery) Order(o ...note.OrderOption) *NoteQuery {
	nq.order = append(nq.order, o...)
	return nq
}

// QueryOwner chains the current query on the "owner" edge.
func (nq *NoteQuery) QueryOwner() *OrganizationQuery {
	query := (&OrganizationClient{config: nq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := nq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := nq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(note.Table, note.FieldID, selector),
			sqlgraph.To(organization.Table, organization.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, note.OwnerTable, note.OwnerColumn),
		)
		schemaConfig := nq.schemaConfig
		step.To.Schema = schemaConfig.Organization
		step.Edge.Schema = schemaConfig.Note
		fromU = sqlgraph.SetNeighbors(nq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryTask chains the current query on the "task" edge.
func (nq *NoteQuery) QueryTask() *TaskQuery {
	query := (&TaskClient{config: nq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := nq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := nq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(note.Table, note.FieldID, selector),
			sqlgraph.To(task.Table, task.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, note.TaskTable, note.TaskColumn),
		)
		schemaConfig := nq.schemaConfig
		step.To.Schema = schemaConfig.Task
		step.Edge.Schema = schemaConfig.Note
		fromU = sqlgraph.SetNeighbors(nq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryFiles chains the current query on the "files" edge.
func (nq *NoteQuery) QueryFiles() *FileQuery {
	query := (&FileClient{config: nq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := nq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := nq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(note.Table, note.FieldID, selector),
			sqlgraph.To(file.Table, file.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, note.FilesTable, note.FilesColumn),
		)
		schemaConfig := nq.schemaConfig
		step.To.Schema = schemaConfig.File
		step.Edge.Schema = schemaConfig.File
		fromU = sqlgraph.SetNeighbors(nq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Note entity from the query.
// Returns a *NotFoundError when no Note was found.
func (nq *NoteQuery) First(ctx context.Context) (*Note, error) {
	nodes, err := nq.Limit(1).All(setContextOp(ctx, nq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{note.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (nq *NoteQuery) FirstX(ctx context.Context) *Note {
	node, err := nq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Note ID from the query.
// Returns a *NotFoundError when no Note ID was found.
func (nq *NoteQuery) FirstID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = nq.Limit(1).IDs(setContextOp(ctx, nq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{note.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (nq *NoteQuery) FirstIDX(ctx context.Context) string {
	id, err := nq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Note entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Note entity is found.
// Returns a *NotFoundError when no Note entities are found.
func (nq *NoteQuery) Only(ctx context.Context) (*Note, error) {
	nodes, err := nq.Limit(2).All(setContextOp(ctx, nq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{note.Label}
	default:
		return nil, &NotSingularError{note.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (nq *NoteQuery) OnlyX(ctx context.Context) *Note {
	node, err := nq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Note ID in the query.
// Returns a *NotSingularError when more than one Note ID is found.
// Returns a *NotFoundError when no entities are found.
func (nq *NoteQuery) OnlyID(ctx context.Context) (id string, err error) {
	var ids []string
	if ids, err = nq.Limit(2).IDs(setContextOp(ctx, nq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{note.Label}
	default:
		err = &NotSingularError{note.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (nq *NoteQuery) OnlyIDX(ctx context.Context) string {
	id, err := nq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Notes.
func (nq *NoteQuery) All(ctx context.Context) ([]*Note, error) {
	ctx = setContextOp(ctx, nq.ctx, ent.OpQueryAll)
	if err := nq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Note, *NoteQuery]()
	return withInterceptors[[]*Note](ctx, nq, qr, nq.inters)
}

// AllX is like All, but panics if an error occurs.
func (nq *NoteQuery) AllX(ctx context.Context) []*Note {
	nodes, err := nq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Note IDs.
func (nq *NoteQuery) IDs(ctx context.Context) (ids []string, err error) {
	if nq.ctx.Unique == nil && nq.path != nil {
		nq.Unique(true)
	}
	ctx = setContextOp(ctx, nq.ctx, ent.OpQueryIDs)
	if err = nq.Select(note.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (nq *NoteQuery) IDsX(ctx context.Context) []string {
	ids, err := nq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (nq *NoteQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, nq.ctx, ent.OpQueryCount)
	if err := nq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, nq, querierCount[*NoteQuery](), nq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (nq *NoteQuery) CountX(ctx context.Context) int {
	count, err := nq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (nq *NoteQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, nq.ctx, ent.OpQueryExist)
	switch _, err := nq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("generated: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (nq *NoteQuery) ExistX(ctx context.Context) bool {
	exist, err := nq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the NoteQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (nq *NoteQuery) Clone() *NoteQuery {
	if nq == nil {
		return nil
	}
	return &NoteQuery{
		config:     nq.config,
		ctx:        nq.ctx.Clone(),
		order:      append([]note.OrderOption{}, nq.order...),
		inters:     append([]Interceptor{}, nq.inters...),
		predicates: append([]predicate.Note{}, nq.predicates...),
		withOwner:  nq.withOwner.Clone(),
		withTask:   nq.withTask.Clone(),
		withFiles:  nq.withFiles.Clone(),
		// clone intermediate query.
		sql:       nq.sql.Clone(),
		path:      nq.path,
		modifiers: append([]func(*sql.Selector){}, nq.modifiers...),
	}
}

// WithOwner tells the query-builder to eager-load the nodes that are connected to
// the "owner" edge. The optional arguments are used to configure the query builder of the edge.
func (nq *NoteQuery) WithOwner(opts ...func(*OrganizationQuery)) *NoteQuery {
	query := (&OrganizationClient{config: nq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	nq.withOwner = query
	return nq
}

// WithTask tells the query-builder to eager-load the nodes that are connected to
// the "task" edge. The optional arguments are used to configure the query builder of the edge.
func (nq *NoteQuery) WithTask(opts ...func(*TaskQuery)) *NoteQuery {
	query := (&TaskClient{config: nq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	nq.withTask = query
	return nq
}

// WithFiles tells the query-builder to eager-load the nodes that are connected to
// the "files" edge. The optional arguments are used to configure the query builder of the edge.
func (nq *NoteQuery) WithFiles(opts ...func(*FileQuery)) *NoteQuery {
	query := (&FileClient{config: nq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	nq.withFiles = query
	return nq
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
//	client.Note.Query().
//		GroupBy(note.FieldCreatedAt).
//		Aggregate(generated.Count()).
//		Scan(ctx, &v)
func (nq *NoteQuery) GroupBy(field string, fields ...string) *NoteGroupBy {
	nq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &NoteGroupBy{build: nq}
	grbuild.flds = &nq.ctx.Fields
	grbuild.label = note.Label
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
//	client.Note.Query().
//		Select(note.FieldCreatedAt).
//		Scan(ctx, &v)
func (nq *NoteQuery) Select(fields ...string) *NoteSelect {
	nq.ctx.Fields = append(nq.ctx.Fields, fields...)
	sbuild := &NoteSelect{NoteQuery: nq}
	sbuild.label = note.Label
	sbuild.flds, sbuild.scan = &nq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a NoteSelect configured with the given aggregations.
func (nq *NoteQuery) Aggregate(fns ...AggregateFunc) *NoteSelect {
	return nq.Select().Aggregate(fns...)
}

func (nq *NoteQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range nq.inters {
		if inter == nil {
			return fmt.Errorf("generated: uninitialized interceptor (forgotten import generated/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, nq); err != nil {
				return err
			}
		}
	}
	for _, f := range nq.ctx.Fields {
		if !note.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
		}
	}
	if nq.path != nil {
		prev, err := nq.path(ctx)
		if err != nil {
			return err
		}
		nq.sql = prev
	}
	if note.Policy == nil {
		return errors.New("generated: uninitialized note.Policy (forgotten import generated/runtime?)")
	}
	if err := note.Policy.EvalQuery(ctx, nq); err != nil {
		return err
	}
	return nil
}

func (nq *NoteQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Note, error) {
	var (
		nodes       = []*Note{}
		withFKs     = nq.withFKs
		_spec       = nq.querySpec()
		loadedTypes = [3]bool{
			nq.withOwner != nil,
			nq.withTask != nil,
			nq.withFiles != nil,
		}
	)
	if nq.withTask != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, note.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Note).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Note{config: nq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	_spec.Node.Schema = nq.schemaConfig.Note
	ctx = internal.NewSchemaConfigContext(ctx, nq.schemaConfig)
	if len(nq.modifiers) > 0 {
		_spec.Modifiers = nq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, nq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := nq.withOwner; query != nil {
		if err := nq.loadOwner(ctx, query, nodes, nil,
			func(n *Note, e *Organization) { n.Edges.Owner = e }); err != nil {
			return nil, err
		}
	}
	if query := nq.withTask; query != nil {
		if err := nq.loadTask(ctx, query, nodes, nil,
			func(n *Note, e *Task) { n.Edges.Task = e }); err != nil {
			return nil, err
		}
	}
	if query := nq.withFiles; query != nil {
		if err := nq.loadFiles(ctx, query, nodes,
			func(n *Note) { n.Edges.Files = []*File{} },
			func(n *Note, e *File) { n.Edges.Files = append(n.Edges.Files, e) }); err != nil {
			return nil, err
		}
	}
	for name, query := range nq.withNamedFiles {
		if err := nq.loadFiles(ctx, query, nodes,
			func(n *Note) { n.appendNamedFiles(name) },
			func(n *Note, e *File) { n.appendNamedFiles(name, e) }); err != nil {
			return nil, err
		}
	}
	for i := range nq.loadTotal {
		if err := nq.loadTotal[i](ctx, nodes); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (nq *NoteQuery) loadOwner(ctx context.Context, query *OrganizationQuery, nodes []*Note, init func(*Note), assign func(*Note, *Organization)) error {
	ids := make([]string, 0, len(nodes))
	nodeids := make(map[string][]*Note)
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
func (nq *NoteQuery) loadTask(ctx context.Context, query *TaskQuery, nodes []*Note, init func(*Note), assign func(*Note, *Task)) error {
	ids := make([]string, 0, len(nodes))
	nodeids := make(map[string][]*Note)
	for i := range nodes {
		if nodes[i].task_comments == nil {
			continue
		}
		fk := *nodes[i].task_comments
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(task.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "task_comments" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (nq *NoteQuery) loadFiles(ctx context.Context, query *FileQuery, nodes []*Note, init func(*Note), assign func(*Note, *File)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[string]*Note)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.File(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(note.FilesColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.note_files
		if fk == nil {
			return fmt.Errorf(`foreign-key "note_files" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "note_files" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (nq *NoteQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := nq.querySpec()
	_spec.Node.Schema = nq.schemaConfig.Note
	ctx = internal.NewSchemaConfigContext(ctx, nq.schemaConfig)
	if len(nq.modifiers) > 0 {
		_spec.Modifiers = nq.modifiers
	}
	_spec.Node.Columns = nq.ctx.Fields
	if len(nq.ctx.Fields) > 0 {
		_spec.Unique = nq.ctx.Unique != nil && *nq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, nq.driver, _spec)
}

func (nq *NoteQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(note.Table, note.Columns, sqlgraph.NewFieldSpec(note.FieldID, field.TypeString))
	_spec.From = nq.sql
	if unique := nq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if nq.path != nil {
		_spec.Unique = true
	}
	if fields := nq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, note.FieldID)
		for i := range fields {
			if fields[i] != note.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if nq.withOwner != nil {
			_spec.Node.AddColumnOnce(note.FieldOwnerID)
		}
	}
	if ps := nq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := nq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := nq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := nq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (nq *NoteQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(nq.driver.Dialect())
	t1 := builder.Table(note.Table)
	columns := nq.ctx.Fields
	if len(columns) == 0 {
		columns = note.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if nq.sql != nil {
		selector = nq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if nq.ctx.Unique != nil && *nq.ctx.Unique {
		selector.Distinct()
	}
	t1.Schema(nq.schemaConfig.Note)
	ctx = internal.NewSchemaConfigContext(ctx, nq.schemaConfig)
	selector.WithContext(ctx)
	for _, m := range nq.modifiers {
		m(selector)
	}
	for _, p := range nq.predicates {
		p(selector)
	}
	for _, p := range nq.order {
		p(selector)
	}
	if offset := nq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := nq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (nq *NoteQuery) Modify(modifiers ...func(s *sql.Selector)) *NoteSelect {
	nq.modifiers = append(nq.modifiers, modifiers...)
	return nq.Select()
}

// WithNamedFiles tells the query-builder to eager-load the nodes that are connected to the "files"
// edge with the given name. The optional arguments are used to configure the query builder of the edge.
func (nq *NoteQuery) WithNamedFiles(name string, opts ...func(*FileQuery)) *NoteQuery {
	query := (&FileClient{config: nq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	if nq.withNamedFiles == nil {
		nq.withNamedFiles = make(map[string]*FileQuery)
	}
	nq.withNamedFiles[name] = query
	return nq
}

// CountIDs returns the count of ids and allows for filtering of the query post retrieval by IDs
func (nq *NoteQuery) CountIDs(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, nq.ctx, ent.OpQueryIDs)
	if err := nq.prepareQuery(ctx); err != nil {
		return 0, err
	}

	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return nq.IDs(ctx)
	})

	ids, err := withInterceptors[[]string](ctx, nq, qr, nq.inters)
	if err != nil {
		return 0, err
	}

	return len(ids), nil
}

// NoteGroupBy is the group-by builder for Note entities.
type NoteGroupBy struct {
	selector
	build *NoteQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (ngb *NoteGroupBy) Aggregate(fns ...AggregateFunc) *NoteGroupBy {
	ngb.fns = append(ngb.fns, fns...)
	return ngb
}

// Scan applies the selector query and scans the result into the given value.
func (ngb *NoteGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ngb.build.ctx, ent.OpQueryGroupBy)
	if err := ngb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*NoteQuery, *NoteGroupBy](ctx, ngb.build, ngb, ngb.build.inters, v)
}

func (ngb *NoteGroupBy) sqlScan(ctx context.Context, root *NoteQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(ngb.fns))
	for _, fn := range ngb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*ngb.flds)+len(ngb.fns))
		for _, f := range *ngb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*ngb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ngb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// NoteSelect is the builder for selecting fields of Note entities.
type NoteSelect struct {
	*NoteQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ns *NoteSelect) Aggregate(fns ...AggregateFunc) *NoteSelect {
	ns.fns = append(ns.fns, fns...)
	return ns
}

// Scan applies the selector query and scans the result into the given value.
func (ns *NoteSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ns.ctx, ent.OpQuerySelect)
	if err := ns.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*NoteQuery, *NoteSelect](ctx, ns.NoteQuery, ns, ns.inters, v)
}

func (ns *NoteSelect) sqlScan(ctx context.Context, root *NoteQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ns.fns))
	for _, fn := range ns.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ns.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ns.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (ns *NoteSelect) Modify(modifiers ...func(s *sql.Selector)) *NoteSelect {
	ns.modifiers = append(ns.modifiers, modifiers...)
	return ns
}
