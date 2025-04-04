// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/dialect/sql/sqljson"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/documentdata"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/pkg/enums"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// TemplateUpdate is the builder for updating Template entities.
type TemplateUpdate struct {
	config
	hooks     []Hook
	mutation  *TemplateMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the TemplateUpdate builder.
func (tu *TemplateUpdate) Where(ps ...predicate.Template) *TemplateUpdate {
	tu.mutation.Where(ps...)
	return tu
}

// SetUpdatedAt sets the "updated_at" field.
func (tu *TemplateUpdate) SetUpdatedAt(t time.Time) *TemplateUpdate {
	tu.mutation.SetUpdatedAt(t)
	return tu
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (tu *TemplateUpdate) ClearUpdatedAt() *TemplateUpdate {
	tu.mutation.ClearUpdatedAt()
	return tu
}

// SetUpdatedBy sets the "updated_by" field.
func (tu *TemplateUpdate) SetUpdatedBy(s string) *TemplateUpdate {
	tu.mutation.SetUpdatedBy(s)
	return tu
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tu *TemplateUpdate) SetNillableUpdatedBy(s *string) *TemplateUpdate {
	if s != nil {
		tu.SetUpdatedBy(*s)
	}
	return tu
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (tu *TemplateUpdate) ClearUpdatedBy() *TemplateUpdate {
	tu.mutation.ClearUpdatedBy()
	return tu
}

// SetDeletedAt sets the "deleted_at" field.
func (tu *TemplateUpdate) SetDeletedAt(t time.Time) *TemplateUpdate {
	tu.mutation.SetDeletedAt(t)
	return tu
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tu *TemplateUpdate) SetNillableDeletedAt(t *time.Time) *TemplateUpdate {
	if t != nil {
		tu.SetDeletedAt(*t)
	}
	return tu
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (tu *TemplateUpdate) ClearDeletedAt() *TemplateUpdate {
	tu.mutation.ClearDeletedAt()
	return tu
}

// SetDeletedBy sets the "deleted_by" field.
func (tu *TemplateUpdate) SetDeletedBy(s string) *TemplateUpdate {
	tu.mutation.SetDeletedBy(s)
	return tu
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tu *TemplateUpdate) SetNillableDeletedBy(s *string) *TemplateUpdate {
	if s != nil {
		tu.SetDeletedBy(*s)
	}
	return tu
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (tu *TemplateUpdate) ClearDeletedBy() *TemplateUpdate {
	tu.mutation.ClearDeletedBy()
	return tu
}

// SetTags sets the "tags" field.
func (tu *TemplateUpdate) SetTags(s []string) *TemplateUpdate {
	tu.mutation.SetTags(s)
	return tu
}

// AppendTags appends s to the "tags" field.
func (tu *TemplateUpdate) AppendTags(s []string) *TemplateUpdate {
	tu.mutation.AppendTags(s)
	return tu
}

// ClearTags clears the value of the "tags" field.
func (tu *TemplateUpdate) ClearTags() *TemplateUpdate {
	tu.mutation.ClearTags()
	return tu
}

// SetOwnerID sets the "owner_id" field.
func (tu *TemplateUpdate) SetOwnerID(s string) *TemplateUpdate {
	tu.mutation.SetOwnerID(s)
	return tu
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (tu *TemplateUpdate) SetNillableOwnerID(s *string) *TemplateUpdate {
	if s != nil {
		tu.SetOwnerID(*s)
	}
	return tu
}

// ClearOwnerID clears the value of the "owner_id" field.
func (tu *TemplateUpdate) ClearOwnerID() *TemplateUpdate {
	tu.mutation.ClearOwnerID()
	return tu
}

// SetName sets the "name" field.
func (tu *TemplateUpdate) SetName(s string) *TemplateUpdate {
	tu.mutation.SetName(s)
	return tu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (tu *TemplateUpdate) SetNillableName(s *string) *TemplateUpdate {
	if s != nil {
		tu.SetName(*s)
	}
	return tu
}

// SetTemplateType sets the "template_type" field.
func (tu *TemplateUpdate) SetTemplateType(et enums.DocumentType) *TemplateUpdate {
	tu.mutation.SetTemplateType(et)
	return tu
}

// SetNillableTemplateType sets the "template_type" field if the given value is not nil.
func (tu *TemplateUpdate) SetNillableTemplateType(et *enums.DocumentType) *TemplateUpdate {
	if et != nil {
		tu.SetTemplateType(*et)
	}
	return tu
}

// SetDescription sets the "description" field.
func (tu *TemplateUpdate) SetDescription(s string) *TemplateUpdate {
	tu.mutation.SetDescription(s)
	return tu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (tu *TemplateUpdate) SetNillableDescription(s *string) *TemplateUpdate {
	if s != nil {
		tu.SetDescription(*s)
	}
	return tu
}

// ClearDescription clears the value of the "description" field.
func (tu *TemplateUpdate) ClearDescription() *TemplateUpdate {
	tu.mutation.ClearDescription()
	return tu
}

// SetJsonconfig sets the "jsonconfig" field.
func (tu *TemplateUpdate) SetJsonconfig(m map[string]interface{}) *TemplateUpdate {
	tu.mutation.SetJsonconfig(m)
	return tu
}

// SetUischema sets the "uischema" field.
func (tu *TemplateUpdate) SetUischema(m map[string]interface{}) *TemplateUpdate {
	tu.mutation.SetUischema(m)
	return tu
}

// ClearUischema clears the value of the "uischema" field.
func (tu *TemplateUpdate) ClearUischema() *TemplateUpdate {
	tu.mutation.ClearUischema()
	return tu
}

// SetOwner sets the "owner" edge to the Organization entity.
func (tu *TemplateUpdate) SetOwner(o *Organization) *TemplateUpdate {
	return tu.SetOwnerID(o.ID)
}

// AddDocumentIDs adds the "documents" edge to the DocumentData entity by IDs.
func (tu *TemplateUpdate) AddDocumentIDs(ids ...string) *TemplateUpdate {
	tu.mutation.AddDocumentIDs(ids...)
	return tu
}

// AddDocuments adds the "documents" edges to the DocumentData entity.
func (tu *TemplateUpdate) AddDocuments(d ...*DocumentData) *TemplateUpdate {
	ids := make([]string, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return tu.AddDocumentIDs(ids...)
}

// AddFileIDs adds the "files" edge to the File entity by IDs.
func (tu *TemplateUpdate) AddFileIDs(ids ...string) *TemplateUpdate {
	tu.mutation.AddFileIDs(ids...)
	return tu
}

// AddFiles adds the "files" edges to the File entity.
func (tu *TemplateUpdate) AddFiles(f ...*File) *TemplateUpdate {
	ids := make([]string, len(f))
	for i := range f {
		ids[i] = f[i].ID
	}
	return tu.AddFileIDs(ids...)
}

// Mutation returns the TemplateMutation object of the builder.
func (tu *TemplateUpdate) Mutation() *TemplateMutation {
	return tu.mutation
}

// ClearOwner clears the "owner" edge to the Organization entity.
func (tu *TemplateUpdate) ClearOwner() *TemplateUpdate {
	tu.mutation.ClearOwner()
	return tu
}

// ClearDocuments clears all "documents" edges to the DocumentData entity.
func (tu *TemplateUpdate) ClearDocuments() *TemplateUpdate {
	tu.mutation.ClearDocuments()
	return tu
}

// RemoveDocumentIDs removes the "documents" edge to DocumentData entities by IDs.
func (tu *TemplateUpdate) RemoveDocumentIDs(ids ...string) *TemplateUpdate {
	tu.mutation.RemoveDocumentIDs(ids...)
	return tu
}

// RemoveDocuments removes "documents" edges to DocumentData entities.
func (tu *TemplateUpdate) RemoveDocuments(d ...*DocumentData) *TemplateUpdate {
	ids := make([]string, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return tu.RemoveDocumentIDs(ids...)
}

// ClearFiles clears all "files" edges to the File entity.
func (tu *TemplateUpdate) ClearFiles() *TemplateUpdate {
	tu.mutation.ClearFiles()
	return tu
}

// RemoveFileIDs removes the "files" edge to File entities by IDs.
func (tu *TemplateUpdate) RemoveFileIDs(ids ...string) *TemplateUpdate {
	tu.mutation.RemoveFileIDs(ids...)
	return tu
}

// RemoveFiles removes "files" edges to File entities.
func (tu *TemplateUpdate) RemoveFiles(f ...*File) *TemplateUpdate {
	ids := make([]string, len(f))
	for i := range f {
		ids[i] = f[i].ID
	}
	return tu.RemoveFileIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (tu *TemplateUpdate) Save(ctx context.Context) (int, error) {
	if err := tu.defaults(); err != nil {
		return 0, err
	}
	return withHooks(ctx, tu.sqlSave, tu.mutation, tu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tu *TemplateUpdate) SaveX(ctx context.Context) int {
	affected, err := tu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (tu *TemplateUpdate) Exec(ctx context.Context) error {
	_, err := tu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tu *TemplateUpdate) ExecX(ctx context.Context) {
	if err := tu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tu *TemplateUpdate) defaults() error {
	if _, ok := tu.mutation.UpdatedAt(); !ok && !tu.mutation.UpdatedAtCleared() {
		if template.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized template.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := template.UpdateDefaultUpdatedAt()
		tu.mutation.SetUpdatedAt(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (tu *TemplateUpdate) check() error {
	if v, ok := tu.mutation.OwnerID(); ok {
		if err := template.OwnerIDValidator(v); err != nil {
			return &ValidationError{Name: "owner_id", err: fmt.Errorf(`generated: validator failed for field "Template.owner_id": %w`, err)}
		}
	}
	if v, ok := tu.mutation.Name(); ok {
		if err := template.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`generated: validator failed for field "Template.name": %w`, err)}
		}
	}
	if v, ok := tu.mutation.TemplateType(); ok {
		if err := template.TemplateTypeValidator(v); err != nil {
			return &ValidationError{Name: "template_type", err: fmt.Errorf(`generated: validator failed for field "Template.template_type": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (tu *TemplateUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *TemplateUpdate {
	tu.modifiers = append(tu.modifiers, modifiers...)
	return tu
}

func (tu *TemplateUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := tu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(template.Table, template.Columns, sqlgraph.NewFieldSpec(template.FieldID, field.TypeString))
	if ps := tu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if tu.mutation.CreatedAtCleared() {
		_spec.ClearField(template.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := tu.mutation.UpdatedAt(); ok {
		_spec.SetField(template.FieldUpdatedAt, field.TypeTime, value)
	}
	if tu.mutation.UpdatedAtCleared() {
		_spec.ClearField(template.FieldUpdatedAt, field.TypeTime)
	}
	if tu.mutation.CreatedByCleared() {
		_spec.ClearField(template.FieldCreatedBy, field.TypeString)
	}
	if value, ok := tu.mutation.UpdatedBy(); ok {
		_spec.SetField(template.FieldUpdatedBy, field.TypeString, value)
	}
	if tu.mutation.UpdatedByCleared() {
		_spec.ClearField(template.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := tu.mutation.DeletedAt(); ok {
		_spec.SetField(template.FieldDeletedAt, field.TypeTime, value)
	}
	if tu.mutation.DeletedAtCleared() {
		_spec.ClearField(template.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := tu.mutation.DeletedBy(); ok {
		_spec.SetField(template.FieldDeletedBy, field.TypeString, value)
	}
	if tu.mutation.DeletedByCleared() {
		_spec.ClearField(template.FieldDeletedBy, field.TypeString)
	}
	if value, ok := tu.mutation.Tags(); ok {
		_spec.SetField(template.FieldTags, field.TypeJSON, value)
	}
	if value, ok := tu.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, template.FieldTags, value)
		})
	}
	if tu.mutation.TagsCleared() {
		_spec.ClearField(template.FieldTags, field.TypeJSON)
	}
	if value, ok := tu.mutation.Name(); ok {
		_spec.SetField(template.FieldName, field.TypeString, value)
	}
	if value, ok := tu.mutation.TemplateType(); ok {
		_spec.SetField(template.FieldTemplateType, field.TypeEnum, value)
	}
	if value, ok := tu.mutation.Description(); ok {
		_spec.SetField(template.FieldDescription, field.TypeString, value)
	}
	if tu.mutation.DescriptionCleared() {
		_spec.ClearField(template.FieldDescription, field.TypeString)
	}
	if value, ok := tu.mutation.Jsonconfig(); ok {
		_spec.SetField(template.FieldJsonconfig, field.TypeJSON, value)
	}
	if value, ok := tu.mutation.Uischema(); ok {
		_spec.SetField(template.FieldUischema, field.TypeJSON, value)
	}
	if tu.mutation.UischemaCleared() {
		_spec.ClearField(template.FieldUischema, field.TypeJSON)
	}
	if tu.mutation.OwnerCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   template.OwnerTable,
			Columns: []string{template.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = tu.schemaConfig.Template
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tu.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   template.OwnerTable,
			Columns: []string{template.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = tu.schemaConfig.Template
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if tu.mutation.DocumentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   template.DocumentsTable,
			Columns: []string{template.DocumentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(documentdata.FieldID, field.TypeString),
			},
		}
		edge.Schema = tu.schemaConfig.DocumentData
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tu.mutation.RemovedDocumentsIDs(); len(nodes) > 0 && !tu.mutation.DocumentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   template.DocumentsTable,
			Columns: []string{template.DocumentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(documentdata.FieldID, field.TypeString),
			},
		}
		edge.Schema = tu.schemaConfig.DocumentData
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tu.mutation.DocumentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   template.DocumentsTable,
			Columns: []string{template.DocumentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(documentdata.FieldID, field.TypeString),
			},
		}
		edge.Schema = tu.schemaConfig.DocumentData
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if tu.mutation.FilesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   template.FilesTable,
			Columns: template.FilesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = tu.schemaConfig.TemplateFiles
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tu.mutation.RemovedFilesIDs(); len(nodes) > 0 && !tu.mutation.FilesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   template.FilesTable,
			Columns: template.FilesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = tu.schemaConfig.TemplateFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tu.mutation.FilesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   template.FilesTable,
			Columns: template.FilesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = tu.schemaConfig.TemplateFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.Node.Schema = tu.schemaConfig.Template
	ctx = internal.NewSchemaConfigContext(ctx, tu.schemaConfig)
	_spec.AddModifiers(tu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, tu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{template.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	tu.mutation.done = true
	return n, nil
}

// TemplateUpdateOne is the builder for updating a single Template entity.
type TemplateUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *TemplateMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetUpdatedAt sets the "updated_at" field.
func (tuo *TemplateUpdateOne) SetUpdatedAt(t time.Time) *TemplateUpdateOne {
	tuo.mutation.SetUpdatedAt(t)
	return tuo
}

// ClearUpdatedAt clears the value of the "updated_at" field.
func (tuo *TemplateUpdateOne) ClearUpdatedAt() *TemplateUpdateOne {
	tuo.mutation.ClearUpdatedAt()
	return tuo
}

// SetUpdatedBy sets the "updated_by" field.
func (tuo *TemplateUpdateOne) SetUpdatedBy(s string) *TemplateUpdateOne {
	tuo.mutation.SetUpdatedBy(s)
	return tuo
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tuo *TemplateUpdateOne) SetNillableUpdatedBy(s *string) *TemplateUpdateOne {
	if s != nil {
		tuo.SetUpdatedBy(*s)
	}
	return tuo
}

// ClearUpdatedBy clears the value of the "updated_by" field.
func (tuo *TemplateUpdateOne) ClearUpdatedBy() *TemplateUpdateOne {
	tuo.mutation.ClearUpdatedBy()
	return tuo
}

// SetDeletedAt sets the "deleted_at" field.
func (tuo *TemplateUpdateOne) SetDeletedAt(t time.Time) *TemplateUpdateOne {
	tuo.mutation.SetDeletedAt(t)
	return tuo
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tuo *TemplateUpdateOne) SetNillableDeletedAt(t *time.Time) *TemplateUpdateOne {
	if t != nil {
		tuo.SetDeletedAt(*t)
	}
	return tuo
}

// ClearDeletedAt clears the value of the "deleted_at" field.
func (tuo *TemplateUpdateOne) ClearDeletedAt() *TemplateUpdateOne {
	tuo.mutation.ClearDeletedAt()
	return tuo
}

// SetDeletedBy sets the "deleted_by" field.
func (tuo *TemplateUpdateOne) SetDeletedBy(s string) *TemplateUpdateOne {
	tuo.mutation.SetDeletedBy(s)
	return tuo
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tuo *TemplateUpdateOne) SetNillableDeletedBy(s *string) *TemplateUpdateOne {
	if s != nil {
		tuo.SetDeletedBy(*s)
	}
	return tuo
}

// ClearDeletedBy clears the value of the "deleted_by" field.
func (tuo *TemplateUpdateOne) ClearDeletedBy() *TemplateUpdateOne {
	tuo.mutation.ClearDeletedBy()
	return tuo
}

// SetTags sets the "tags" field.
func (tuo *TemplateUpdateOne) SetTags(s []string) *TemplateUpdateOne {
	tuo.mutation.SetTags(s)
	return tuo
}

// AppendTags appends s to the "tags" field.
func (tuo *TemplateUpdateOne) AppendTags(s []string) *TemplateUpdateOne {
	tuo.mutation.AppendTags(s)
	return tuo
}

// ClearTags clears the value of the "tags" field.
func (tuo *TemplateUpdateOne) ClearTags() *TemplateUpdateOne {
	tuo.mutation.ClearTags()
	return tuo
}

// SetOwnerID sets the "owner_id" field.
func (tuo *TemplateUpdateOne) SetOwnerID(s string) *TemplateUpdateOne {
	tuo.mutation.SetOwnerID(s)
	return tuo
}

// SetNillableOwnerID sets the "owner_id" field if the given value is not nil.
func (tuo *TemplateUpdateOne) SetNillableOwnerID(s *string) *TemplateUpdateOne {
	if s != nil {
		tuo.SetOwnerID(*s)
	}
	return tuo
}

// ClearOwnerID clears the value of the "owner_id" field.
func (tuo *TemplateUpdateOne) ClearOwnerID() *TemplateUpdateOne {
	tuo.mutation.ClearOwnerID()
	return tuo
}

// SetName sets the "name" field.
func (tuo *TemplateUpdateOne) SetName(s string) *TemplateUpdateOne {
	tuo.mutation.SetName(s)
	return tuo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (tuo *TemplateUpdateOne) SetNillableName(s *string) *TemplateUpdateOne {
	if s != nil {
		tuo.SetName(*s)
	}
	return tuo
}

// SetTemplateType sets the "template_type" field.
func (tuo *TemplateUpdateOne) SetTemplateType(et enums.DocumentType) *TemplateUpdateOne {
	tuo.mutation.SetTemplateType(et)
	return tuo
}

// SetNillableTemplateType sets the "template_type" field if the given value is not nil.
func (tuo *TemplateUpdateOne) SetNillableTemplateType(et *enums.DocumentType) *TemplateUpdateOne {
	if et != nil {
		tuo.SetTemplateType(*et)
	}
	return tuo
}

// SetDescription sets the "description" field.
func (tuo *TemplateUpdateOne) SetDescription(s string) *TemplateUpdateOne {
	tuo.mutation.SetDescription(s)
	return tuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (tuo *TemplateUpdateOne) SetNillableDescription(s *string) *TemplateUpdateOne {
	if s != nil {
		tuo.SetDescription(*s)
	}
	return tuo
}

// ClearDescription clears the value of the "description" field.
func (tuo *TemplateUpdateOne) ClearDescription() *TemplateUpdateOne {
	tuo.mutation.ClearDescription()
	return tuo
}

// SetJsonconfig sets the "jsonconfig" field.
func (tuo *TemplateUpdateOne) SetJsonconfig(m map[string]interface{}) *TemplateUpdateOne {
	tuo.mutation.SetJsonconfig(m)
	return tuo
}

// SetUischema sets the "uischema" field.
func (tuo *TemplateUpdateOne) SetUischema(m map[string]interface{}) *TemplateUpdateOne {
	tuo.mutation.SetUischema(m)
	return tuo
}

// ClearUischema clears the value of the "uischema" field.
func (tuo *TemplateUpdateOne) ClearUischema() *TemplateUpdateOne {
	tuo.mutation.ClearUischema()
	return tuo
}

// SetOwner sets the "owner" edge to the Organization entity.
func (tuo *TemplateUpdateOne) SetOwner(o *Organization) *TemplateUpdateOne {
	return tuo.SetOwnerID(o.ID)
}

// AddDocumentIDs adds the "documents" edge to the DocumentData entity by IDs.
func (tuo *TemplateUpdateOne) AddDocumentIDs(ids ...string) *TemplateUpdateOne {
	tuo.mutation.AddDocumentIDs(ids...)
	return tuo
}

// AddDocuments adds the "documents" edges to the DocumentData entity.
func (tuo *TemplateUpdateOne) AddDocuments(d ...*DocumentData) *TemplateUpdateOne {
	ids := make([]string, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return tuo.AddDocumentIDs(ids...)
}

// AddFileIDs adds the "files" edge to the File entity by IDs.
func (tuo *TemplateUpdateOne) AddFileIDs(ids ...string) *TemplateUpdateOne {
	tuo.mutation.AddFileIDs(ids...)
	return tuo
}

// AddFiles adds the "files" edges to the File entity.
func (tuo *TemplateUpdateOne) AddFiles(f ...*File) *TemplateUpdateOne {
	ids := make([]string, len(f))
	for i := range f {
		ids[i] = f[i].ID
	}
	return tuo.AddFileIDs(ids...)
}

// Mutation returns the TemplateMutation object of the builder.
func (tuo *TemplateUpdateOne) Mutation() *TemplateMutation {
	return tuo.mutation
}

// ClearOwner clears the "owner" edge to the Organization entity.
func (tuo *TemplateUpdateOne) ClearOwner() *TemplateUpdateOne {
	tuo.mutation.ClearOwner()
	return tuo
}

// ClearDocuments clears all "documents" edges to the DocumentData entity.
func (tuo *TemplateUpdateOne) ClearDocuments() *TemplateUpdateOne {
	tuo.mutation.ClearDocuments()
	return tuo
}

// RemoveDocumentIDs removes the "documents" edge to DocumentData entities by IDs.
func (tuo *TemplateUpdateOne) RemoveDocumentIDs(ids ...string) *TemplateUpdateOne {
	tuo.mutation.RemoveDocumentIDs(ids...)
	return tuo
}

// RemoveDocuments removes "documents" edges to DocumentData entities.
func (tuo *TemplateUpdateOne) RemoveDocuments(d ...*DocumentData) *TemplateUpdateOne {
	ids := make([]string, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return tuo.RemoveDocumentIDs(ids...)
}

// ClearFiles clears all "files" edges to the File entity.
func (tuo *TemplateUpdateOne) ClearFiles() *TemplateUpdateOne {
	tuo.mutation.ClearFiles()
	return tuo
}

// RemoveFileIDs removes the "files" edge to File entities by IDs.
func (tuo *TemplateUpdateOne) RemoveFileIDs(ids ...string) *TemplateUpdateOne {
	tuo.mutation.RemoveFileIDs(ids...)
	return tuo
}

// RemoveFiles removes "files" edges to File entities.
func (tuo *TemplateUpdateOne) RemoveFiles(f ...*File) *TemplateUpdateOne {
	ids := make([]string, len(f))
	for i := range f {
		ids[i] = f[i].ID
	}
	return tuo.RemoveFileIDs(ids...)
}

// Where appends a list predicates to the TemplateUpdate builder.
func (tuo *TemplateUpdateOne) Where(ps ...predicate.Template) *TemplateUpdateOne {
	tuo.mutation.Where(ps...)
	return tuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (tuo *TemplateUpdateOne) Select(field string, fields ...string) *TemplateUpdateOne {
	tuo.fields = append([]string{field}, fields...)
	return tuo
}

// Save executes the query and returns the updated Template entity.
func (tuo *TemplateUpdateOne) Save(ctx context.Context) (*Template, error) {
	if err := tuo.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, tuo.sqlSave, tuo.mutation, tuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (tuo *TemplateUpdateOne) SaveX(ctx context.Context) *Template {
	node, err := tuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (tuo *TemplateUpdateOne) Exec(ctx context.Context) error {
	_, err := tuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tuo *TemplateUpdateOne) ExecX(ctx context.Context) {
	if err := tuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tuo *TemplateUpdateOne) defaults() error {
	if _, ok := tuo.mutation.UpdatedAt(); !ok && !tuo.mutation.UpdatedAtCleared() {
		if template.UpdateDefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized template.UpdateDefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := template.UpdateDefaultUpdatedAt()
		tuo.mutation.SetUpdatedAt(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (tuo *TemplateUpdateOne) check() error {
	if v, ok := tuo.mutation.OwnerID(); ok {
		if err := template.OwnerIDValidator(v); err != nil {
			return &ValidationError{Name: "owner_id", err: fmt.Errorf(`generated: validator failed for field "Template.owner_id": %w`, err)}
		}
	}
	if v, ok := tuo.mutation.Name(); ok {
		if err := template.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`generated: validator failed for field "Template.name": %w`, err)}
		}
	}
	if v, ok := tuo.mutation.TemplateType(); ok {
		if err := template.TemplateTypeValidator(v); err != nil {
			return &ValidationError{Name: "template_type", err: fmt.Errorf(`generated: validator failed for field "Template.template_type": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (tuo *TemplateUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *TemplateUpdateOne {
	tuo.modifiers = append(tuo.modifiers, modifiers...)
	return tuo
}

func (tuo *TemplateUpdateOne) sqlSave(ctx context.Context) (_node *Template, err error) {
	if err := tuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(template.Table, template.Columns, sqlgraph.NewFieldSpec(template.FieldID, field.TypeString))
	id, ok := tuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`generated: missing "Template.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := tuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, template.FieldID)
		for _, f := range fields {
			if !template.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("generated: invalid field %q for query", f)}
			}
			if f != template.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := tuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if tuo.mutation.CreatedAtCleared() {
		_spec.ClearField(template.FieldCreatedAt, field.TypeTime)
	}
	if value, ok := tuo.mutation.UpdatedAt(); ok {
		_spec.SetField(template.FieldUpdatedAt, field.TypeTime, value)
	}
	if tuo.mutation.UpdatedAtCleared() {
		_spec.ClearField(template.FieldUpdatedAt, field.TypeTime)
	}
	if tuo.mutation.CreatedByCleared() {
		_spec.ClearField(template.FieldCreatedBy, field.TypeString)
	}
	if value, ok := tuo.mutation.UpdatedBy(); ok {
		_spec.SetField(template.FieldUpdatedBy, field.TypeString, value)
	}
	if tuo.mutation.UpdatedByCleared() {
		_spec.ClearField(template.FieldUpdatedBy, field.TypeString)
	}
	if value, ok := tuo.mutation.DeletedAt(); ok {
		_spec.SetField(template.FieldDeletedAt, field.TypeTime, value)
	}
	if tuo.mutation.DeletedAtCleared() {
		_spec.ClearField(template.FieldDeletedAt, field.TypeTime)
	}
	if value, ok := tuo.mutation.DeletedBy(); ok {
		_spec.SetField(template.FieldDeletedBy, field.TypeString, value)
	}
	if tuo.mutation.DeletedByCleared() {
		_spec.ClearField(template.FieldDeletedBy, field.TypeString)
	}
	if value, ok := tuo.mutation.Tags(); ok {
		_spec.SetField(template.FieldTags, field.TypeJSON, value)
	}
	if value, ok := tuo.mutation.AppendedTags(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, template.FieldTags, value)
		})
	}
	if tuo.mutation.TagsCleared() {
		_spec.ClearField(template.FieldTags, field.TypeJSON)
	}
	if value, ok := tuo.mutation.Name(); ok {
		_spec.SetField(template.FieldName, field.TypeString, value)
	}
	if value, ok := tuo.mutation.TemplateType(); ok {
		_spec.SetField(template.FieldTemplateType, field.TypeEnum, value)
	}
	if value, ok := tuo.mutation.Description(); ok {
		_spec.SetField(template.FieldDescription, field.TypeString, value)
	}
	if tuo.mutation.DescriptionCleared() {
		_spec.ClearField(template.FieldDescription, field.TypeString)
	}
	if value, ok := tuo.mutation.Jsonconfig(); ok {
		_spec.SetField(template.FieldJsonconfig, field.TypeJSON, value)
	}
	if value, ok := tuo.mutation.Uischema(); ok {
		_spec.SetField(template.FieldUischema, field.TypeJSON, value)
	}
	if tuo.mutation.UischemaCleared() {
		_spec.ClearField(template.FieldUischema, field.TypeJSON)
	}
	if tuo.mutation.OwnerCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   template.OwnerTable,
			Columns: []string{template.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = tuo.schemaConfig.Template
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tuo.mutation.OwnerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   template.OwnerTable,
			Columns: []string{template.OwnerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = tuo.schemaConfig.Template
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if tuo.mutation.DocumentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   template.DocumentsTable,
			Columns: []string{template.DocumentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(documentdata.FieldID, field.TypeString),
			},
		}
		edge.Schema = tuo.schemaConfig.DocumentData
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tuo.mutation.RemovedDocumentsIDs(); len(nodes) > 0 && !tuo.mutation.DocumentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   template.DocumentsTable,
			Columns: []string{template.DocumentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(documentdata.FieldID, field.TypeString),
			},
		}
		edge.Schema = tuo.schemaConfig.DocumentData
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tuo.mutation.DocumentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   template.DocumentsTable,
			Columns: []string{template.DocumentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(documentdata.FieldID, field.TypeString),
			},
		}
		edge.Schema = tuo.schemaConfig.DocumentData
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if tuo.mutation.FilesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   template.FilesTable,
			Columns: template.FilesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = tuo.schemaConfig.TemplateFiles
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tuo.mutation.RemovedFilesIDs(); len(nodes) > 0 && !tuo.mutation.FilesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   template.FilesTable,
			Columns: template.FilesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = tuo.schemaConfig.TemplateFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := tuo.mutation.FilesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   template.FilesTable,
			Columns: template.FilesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = tuo.schemaConfig.TemplateFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.Node.Schema = tuo.schemaConfig.Template
	ctx = internal.NewSchemaConfigContext(ctx, tuo.schemaConfig)
	_spec.AddModifiers(tuo.modifiers...)
	_node = &Template{config: tuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, tuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{template.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	tuo.mutation.done = true
	return _node, nil
}
