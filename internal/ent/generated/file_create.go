// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/contact"
	"github.com/theopenlane/core/internal/ent/generated/documentdata"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/event"
	"github.com/theopenlane/core/internal/ent/generated/evidence"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/program"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
)

// FileCreate is the builder for creating a File entity.
type FileCreate struct {
	config
	mutation *FileMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (fc *FileCreate) SetCreatedAt(t time.Time) *FileCreate {
	fc.mutation.SetCreatedAt(t)
	return fc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (fc *FileCreate) SetNillableCreatedAt(t *time.Time) *FileCreate {
	if t != nil {
		fc.SetCreatedAt(*t)
	}
	return fc
}

// SetUpdatedAt sets the "updated_at" field.
func (fc *FileCreate) SetUpdatedAt(t time.Time) *FileCreate {
	fc.mutation.SetUpdatedAt(t)
	return fc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (fc *FileCreate) SetNillableUpdatedAt(t *time.Time) *FileCreate {
	if t != nil {
		fc.SetUpdatedAt(*t)
	}
	return fc
}

// SetCreatedBy sets the "created_by" field.
func (fc *FileCreate) SetCreatedBy(s string) *FileCreate {
	fc.mutation.SetCreatedBy(s)
	return fc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (fc *FileCreate) SetNillableCreatedBy(s *string) *FileCreate {
	if s != nil {
		fc.SetCreatedBy(*s)
	}
	return fc
}

// SetUpdatedBy sets the "updated_by" field.
func (fc *FileCreate) SetUpdatedBy(s string) *FileCreate {
	fc.mutation.SetUpdatedBy(s)
	return fc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (fc *FileCreate) SetNillableUpdatedBy(s *string) *FileCreate {
	if s != nil {
		fc.SetUpdatedBy(*s)
	}
	return fc
}

// SetDeletedAt sets the "deleted_at" field.
func (fc *FileCreate) SetDeletedAt(t time.Time) *FileCreate {
	fc.mutation.SetDeletedAt(t)
	return fc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (fc *FileCreate) SetNillableDeletedAt(t *time.Time) *FileCreate {
	if t != nil {
		fc.SetDeletedAt(*t)
	}
	return fc
}

// SetDeletedBy sets the "deleted_by" field.
func (fc *FileCreate) SetDeletedBy(s string) *FileCreate {
	fc.mutation.SetDeletedBy(s)
	return fc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (fc *FileCreate) SetNillableDeletedBy(s *string) *FileCreate {
	if s != nil {
		fc.SetDeletedBy(*s)
	}
	return fc
}

// SetTags sets the "tags" field.
func (fc *FileCreate) SetTags(s []string) *FileCreate {
	fc.mutation.SetTags(s)
	return fc
}

// SetProvidedFileName sets the "provided_file_name" field.
func (fc *FileCreate) SetProvidedFileName(s string) *FileCreate {
	fc.mutation.SetProvidedFileName(s)
	return fc
}

// SetProvidedFileExtension sets the "provided_file_extension" field.
func (fc *FileCreate) SetProvidedFileExtension(s string) *FileCreate {
	fc.mutation.SetProvidedFileExtension(s)
	return fc
}

// SetProvidedFileSize sets the "provided_file_size" field.
func (fc *FileCreate) SetProvidedFileSize(i int64) *FileCreate {
	fc.mutation.SetProvidedFileSize(i)
	return fc
}

// SetNillableProvidedFileSize sets the "provided_file_size" field if the given value is not nil.
func (fc *FileCreate) SetNillableProvidedFileSize(i *int64) *FileCreate {
	if i != nil {
		fc.SetProvidedFileSize(*i)
	}
	return fc
}

// SetPersistedFileSize sets the "persisted_file_size" field.
func (fc *FileCreate) SetPersistedFileSize(i int64) *FileCreate {
	fc.mutation.SetPersistedFileSize(i)
	return fc
}

// SetNillablePersistedFileSize sets the "persisted_file_size" field if the given value is not nil.
func (fc *FileCreate) SetNillablePersistedFileSize(i *int64) *FileCreate {
	if i != nil {
		fc.SetPersistedFileSize(*i)
	}
	return fc
}

// SetDetectedMimeType sets the "detected_mime_type" field.
func (fc *FileCreate) SetDetectedMimeType(s string) *FileCreate {
	fc.mutation.SetDetectedMimeType(s)
	return fc
}

// SetNillableDetectedMimeType sets the "detected_mime_type" field if the given value is not nil.
func (fc *FileCreate) SetNillableDetectedMimeType(s *string) *FileCreate {
	if s != nil {
		fc.SetDetectedMimeType(*s)
	}
	return fc
}

// SetMd5Hash sets the "md5_hash" field.
func (fc *FileCreate) SetMd5Hash(s string) *FileCreate {
	fc.mutation.SetMd5Hash(s)
	return fc
}

// SetNillableMd5Hash sets the "md5_hash" field if the given value is not nil.
func (fc *FileCreate) SetNillableMd5Hash(s *string) *FileCreate {
	if s != nil {
		fc.SetMd5Hash(*s)
	}
	return fc
}

// SetDetectedContentType sets the "detected_content_type" field.
func (fc *FileCreate) SetDetectedContentType(s string) *FileCreate {
	fc.mutation.SetDetectedContentType(s)
	return fc
}

// SetStoreKey sets the "store_key" field.
func (fc *FileCreate) SetStoreKey(s string) *FileCreate {
	fc.mutation.SetStoreKey(s)
	return fc
}

// SetNillableStoreKey sets the "store_key" field if the given value is not nil.
func (fc *FileCreate) SetNillableStoreKey(s *string) *FileCreate {
	if s != nil {
		fc.SetStoreKey(*s)
	}
	return fc
}

// SetCategoryType sets the "category_type" field.
func (fc *FileCreate) SetCategoryType(s string) *FileCreate {
	fc.mutation.SetCategoryType(s)
	return fc
}

// SetNillableCategoryType sets the "category_type" field if the given value is not nil.
func (fc *FileCreate) SetNillableCategoryType(s *string) *FileCreate {
	if s != nil {
		fc.SetCategoryType(*s)
	}
	return fc
}

// SetURI sets the "uri" field.
func (fc *FileCreate) SetURI(s string) *FileCreate {
	fc.mutation.SetURI(s)
	return fc
}

// SetNillableURI sets the "uri" field if the given value is not nil.
func (fc *FileCreate) SetNillableURI(s *string) *FileCreate {
	if s != nil {
		fc.SetURI(*s)
	}
	return fc
}

// SetStorageScheme sets the "storage_scheme" field.
func (fc *FileCreate) SetStorageScheme(s string) *FileCreate {
	fc.mutation.SetStorageScheme(s)
	return fc
}

// SetNillableStorageScheme sets the "storage_scheme" field if the given value is not nil.
func (fc *FileCreate) SetNillableStorageScheme(s *string) *FileCreate {
	if s != nil {
		fc.SetStorageScheme(*s)
	}
	return fc
}

// SetStorageVolume sets the "storage_volume" field.
func (fc *FileCreate) SetStorageVolume(s string) *FileCreate {
	fc.mutation.SetStorageVolume(s)
	return fc
}

// SetNillableStorageVolume sets the "storage_volume" field if the given value is not nil.
func (fc *FileCreate) SetNillableStorageVolume(s *string) *FileCreate {
	if s != nil {
		fc.SetStorageVolume(*s)
	}
	return fc
}

// SetStoragePath sets the "storage_path" field.
func (fc *FileCreate) SetStoragePath(s string) *FileCreate {
	fc.mutation.SetStoragePath(s)
	return fc
}

// SetNillableStoragePath sets the "storage_path" field if the given value is not nil.
func (fc *FileCreate) SetNillableStoragePath(s *string) *FileCreate {
	if s != nil {
		fc.SetStoragePath(*s)
	}
	return fc
}

// SetFileContents sets the "file_contents" field.
func (fc *FileCreate) SetFileContents(b []byte) *FileCreate {
	fc.mutation.SetFileContents(b)
	return fc
}

// SetID sets the "id" field.
func (fc *FileCreate) SetID(s string) *FileCreate {
	fc.mutation.SetID(s)
	return fc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (fc *FileCreate) SetNillableID(s *string) *FileCreate {
	if s != nil {
		fc.SetID(*s)
	}
	return fc
}

// AddUserIDs adds the "user" edge to the User entity by IDs.
func (fc *FileCreate) AddUserIDs(ids ...string) *FileCreate {
	fc.mutation.AddUserIDs(ids...)
	return fc
}

// AddUser adds the "user" edges to the User entity.
func (fc *FileCreate) AddUser(u ...*User) *FileCreate {
	ids := make([]string, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return fc.AddUserIDs(ids...)
}

// AddOrganizationIDs adds the "organization" edge to the Organization entity by IDs.
func (fc *FileCreate) AddOrganizationIDs(ids ...string) *FileCreate {
	fc.mutation.AddOrganizationIDs(ids...)
	return fc
}

// AddOrganization adds the "organization" edges to the Organization entity.
func (fc *FileCreate) AddOrganization(o ...*Organization) *FileCreate {
	ids := make([]string, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return fc.AddOrganizationIDs(ids...)
}

// AddGroupIDs adds the "groups" edge to the Group entity by IDs.
func (fc *FileCreate) AddGroupIDs(ids ...string) *FileCreate {
	fc.mutation.AddGroupIDs(ids...)
	return fc
}

// AddGroups adds the "groups" edges to the Group entity.
func (fc *FileCreate) AddGroups(g ...*Group) *FileCreate {
	ids := make([]string, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return fc.AddGroupIDs(ids...)
}

// AddContactIDs adds the "contact" edge to the Contact entity by IDs.
func (fc *FileCreate) AddContactIDs(ids ...string) *FileCreate {
	fc.mutation.AddContactIDs(ids...)
	return fc
}

// AddContact adds the "contact" edges to the Contact entity.
func (fc *FileCreate) AddContact(c ...*Contact) *FileCreate {
	ids := make([]string, len(c))
	for i := range c {
		ids[i] = c[i].ID
	}
	return fc.AddContactIDs(ids...)
}

// AddEntityIDs adds the "entity" edge to the Entity entity by IDs.
func (fc *FileCreate) AddEntityIDs(ids ...string) *FileCreate {
	fc.mutation.AddEntityIDs(ids...)
	return fc
}

// AddEntity adds the "entity" edges to the Entity entity.
func (fc *FileCreate) AddEntity(e ...*Entity) *FileCreate {
	ids := make([]string, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return fc.AddEntityIDs(ids...)
}

// AddUserSettingIDs adds the "user_setting" edge to the UserSetting entity by IDs.
func (fc *FileCreate) AddUserSettingIDs(ids ...string) *FileCreate {
	fc.mutation.AddUserSettingIDs(ids...)
	return fc
}

// AddUserSetting adds the "user_setting" edges to the UserSetting entity.
func (fc *FileCreate) AddUserSetting(u ...*UserSetting) *FileCreate {
	ids := make([]string, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return fc.AddUserSettingIDs(ids...)
}

// AddOrganizationSettingIDs adds the "organization_setting" edge to the OrganizationSetting entity by IDs.
func (fc *FileCreate) AddOrganizationSettingIDs(ids ...string) *FileCreate {
	fc.mutation.AddOrganizationSettingIDs(ids...)
	return fc
}

// AddOrganizationSetting adds the "organization_setting" edges to the OrganizationSetting entity.
func (fc *FileCreate) AddOrganizationSetting(o ...*OrganizationSetting) *FileCreate {
	ids := make([]string, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return fc.AddOrganizationSettingIDs(ids...)
}

// AddTemplateIDs adds the "template" edge to the Template entity by IDs.
func (fc *FileCreate) AddTemplateIDs(ids ...string) *FileCreate {
	fc.mutation.AddTemplateIDs(ids...)
	return fc
}

// AddTemplate adds the "template" edges to the Template entity.
func (fc *FileCreate) AddTemplate(t ...*Template) *FileCreate {
	ids := make([]string, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return fc.AddTemplateIDs(ids...)
}

// AddDocumentIDs adds the "document" edge to the DocumentData entity by IDs.
func (fc *FileCreate) AddDocumentIDs(ids ...string) *FileCreate {
	fc.mutation.AddDocumentIDs(ids...)
	return fc
}

// AddDocument adds the "document" edges to the DocumentData entity.
func (fc *FileCreate) AddDocument(d ...*DocumentData) *FileCreate {
	ids := make([]string, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return fc.AddDocumentIDs(ids...)
}

// AddProgramIDs adds the "program" edge to the Program entity by IDs.
func (fc *FileCreate) AddProgramIDs(ids ...string) *FileCreate {
	fc.mutation.AddProgramIDs(ids...)
	return fc
}

// AddProgram adds the "program" edges to the Program entity.
func (fc *FileCreate) AddProgram(p ...*Program) *FileCreate {
	ids := make([]string, len(p))
	for i := range p {
		ids[i] = p[i].ID
	}
	return fc.AddProgramIDs(ids...)
}

// AddEvidenceIDs adds the "evidence" edge to the Evidence entity by IDs.
func (fc *FileCreate) AddEvidenceIDs(ids ...string) *FileCreate {
	fc.mutation.AddEvidenceIDs(ids...)
	return fc
}

// AddEvidence adds the "evidence" edges to the Evidence entity.
func (fc *FileCreate) AddEvidence(e ...*Evidence) *FileCreate {
	ids := make([]string, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return fc.AddEvidenceIDs(ids...)
}

// AddEventIDs adds the "events" edge to the Event entity by IDs.
func (fc *FileCreate) AddEventIDs(ids ...string) *FileCreate {
	fc.mutation.AddEventIDs(ids...)
	return fc
}

// AddEvents adds the "events" edges to the Event entity.
func (fc *FileCreate) AddEvents(e ...*Event) *FileCreate {
	ids := make([]string, len(e))
	for i := range e {
		ids[i] = e[i].ID
	}
	return fc.AddEventIDs(ids...)
}

// AddTrustCenterSettingIDs adds the "trust_center_setting" edge to the TrustCenterSetting entity by IDs.
func (fc *FileCreate) AddTrustCenterSettingIDs(ids ...string) *FileCreate {
	fc.mutation.AddTrustCenterSettingIDs(ids...)
	return fc
}

// AddTrustCenterSetting adds the "trust_center_setting" edges to the TrustCenterSetting entity.
func (fc *FileCreate) AddTrustCenterSetting(t ...*TrustCenterSetting) *FileCreate {
	ids := make([]string, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return fc.AddTrustCenterSettingIDs(ids...)
}

// AddSubprocessorIDs adds the "subprocessor" edge to the Subprocessor entity by IDs.
func (fc *FileCreate) AddSubprocessorIDs(ids ...string) *FileCreate {
	fc.mutation.AddSubprocessorIDs(ids...)
	return fc
}

// AddSubprocessor adds the "subprocessor" edges to the Subprocessor entity.
func (fc *FileCreate) AddSubprocessor(s ...*Subprocessor) *FileCreate {
	ids := make([]string, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return fc.AddSubprocessorIDs(ids...)
}

// Mutation returns the FileMutation object of the builder.
func (fc *FileCreate) Mutation() *FileMutation {
	return fc.mutation
}

// Save creates the File in the database.
func (fc *FileCreate) Save(ctx context.Context) (*File, error) {
	if err := fc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, fc.sqlSave, fc.mutation, fc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (fc *FileCreate) SaveX(ctx context.Context) *File {
	v, err := fc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fc *FileCreate) Exec(ctx context.Context) error {
	_, err := fc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fc *FileCreate) ExecX(ctx context.Context) {
	if err := fc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (fc *FileCreate) defaults() error {
	if _, ok := fc.mutation.CreatedAt(); !ok {
		if file.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized file.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := file.DefaultCreatedAt()
		fc.mutation.SetCreatedAt(v)
	}
	if _, ok := fc.mutation.UpdatedAt(); !ok {
		if file.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized file.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := file.DefaultUpdatedAt()
		fc.mutation.SetUpdatedAt(v)
	}
	if _, ok := fc.mutation.Tags(); !ok {
		v := file.DefaultTags
		fc.mutation.SetTags(v)
	}
	if _, ok := fc.mutation.ID(); !ok {
		if file.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized file.DefaultID (forgotten import generated/runtime?)")
		}
		v := file.DefaultID()
		fc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (fc *FileCreate) check() error {
	if _, ok := fc.mutation.ProvidedFileName(); !ok {
		return &ValidationError{Name: "provided_file_name", err: errors.New(`generated: missing required field "File.provided_file_name"`)}
	}
	if _, ok := fc.mutation.ProvidedFileExtension(); !ok {
		return &ValidationError{Name: "provided_file_extension", err: errors.New(`generated: missing required field "File.provided_file_extension"`)}
	}
	if v, ok := fc.mutation.ProvidedFileSize(); ok {
		if err := file.ProvidedFileSizeValidator(v); err != nil {
			return &ValidationError{Name: "provided_file_size", err: fmt.Errorf(`generated: validator failed for field "File.provided_file_size": %w`, err)}
		}
	}
	if v, ok := fc.mutation.PersistedFileSize(); ok {
		if err := file.PersistedFileSizeValidator(v); err != nil {
			return &ValidationError{Name: "persisted_file_size", err: fmt.Errorf(`generated: validator failed for field "File.persisted_file_size": %w`, err)}
		}
	}
	if _, ok := fc.mutation.DetectedContentType(); !ok {
		return &ValidationError{Name: "detected_content_type", err: errors.New(`generated: missing required field "File.detected_content_type"`)}
	}
	return nil
}

func (fc *FileCreate) sqlSave(ctx context.Context) (*File, error) {
	if err := fc.check(); err != nil {
		return nil, err
	}
	_node, _spec := fc.createSpec()
	if err := sqlgraph.CreateNode(ctx, fc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected File.ID type: %T", _spec.ID.Value)
		}
	}
	fc.mutation.id = &_node.ID
	fc.mutation.done = true
	return _node, nil
}

func (fc *FileCreate) createSpec() (*File, *sqlgraph.CreateSpec) {
	var (
		_node = &File{config: fc.config}
		_spec = sqlgraph.NewCreateSpec(file.Table, sqlgraph.NewFieldSpec(file.FieldID, field.TypeString))
	)
	_spec.Schema = fc.schemaConfig.File
	if id, ok := fc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := fc.mutation.CreatedAt(); ok {
		_spec.SetField(file.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := fc.mutation.UpdatedAt(); ok {
		_spec.SetField(file.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := fc.mutation.CreatedBy(); ok {
		_spec.SetField(file.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := fc.mutation.UpdatedBy(); ok {
		_spec.SetField(file.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := fc.mutation.DeletedAt(); ok {
		_spec.SetField(file.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := fc.mutation.DeletedBy(); ok {
		_spec.SetField(file.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := fc.mutation.Tags(); ok {
		_spec.SetField(file.FieldTags, field.TypeJSON, value)
		_node.Tags = value
	}
	if value, ok := fc.mutation.ProvidedFileName(); ok {
		_spec.SetField(file.FieldProvidedFileName, field.TypeString, value)
		_node.ProvidedFileName = value
	}
	if value, ok := fc.mutation.ProvidedFileExtension(); ok {
		_spec.SetField(file.FieldProvidedFileExtension, field.TypeString, value)
		_node.ProvidedFileExtension = value
	}
	if value, ok := fc.mutation.ProvidedFileSize(); ok {
		_spec.SetField(file.FieldProvidedFileSize, field.TypeInt64, value)
		_node.ProvidedFileSize = value
	}
	if value, ok := fc.mutation.PersistedFileSize(); ok {
		_spec.SetField(file.FieldPersistedFileSize, field.TypeInt64, value)
		_node.PersistedFileSize = value
	}
	if value, ok := fc.mutation.DetectedMimeType(); ok {
		_spec.SetField(file.FieldDetectedMimeType, field.TypeString, value)
		_node.DetectedMimeType = value
	}
	if value, ok := fc.mutation.Md5Hash(); ok {
		_spec.SetField(file.FieldMd5Hash, field.TypeString, value)
		_node.Md5Hash = value
	}
	if value, ok := fc.mutation.DetectedContentType(); ok {
		_spec.SetField(file.FieldDetectedContentType, field.TypeString, value)
		_node.DetectedContentType = value
	}
	if value, ok := fc.mutation.StoreKey(); ok {
		_spec.SetField(file.FieldStoreKey, field.TypeString, value)
		_node.StoreKey = value
	}
	if value, ok := fc.mutation.CategoryType(); ok {
		_spec.SetField(file.FieldCategoryType, field.TypeString, value)
		_node.CategoryType = value
	}
	if value, ok := fc.mutation.URI(); ok {
		_spec.SetField(file.FieldURI, field.TypeString, value)
		_node.URI = value
	}
	if value, ok := fc.mutation.StorageScheme(); ok {
		_spec.SetField(file.FieldStorageScheme, field.TypeString, value)
		_node.StorageScheme = value
	}
	if value, ok := fc.mutation.StorageVolume(); ok {
		_spec.SetField(file.FieldStorageVolume, field.TypeString, value)
		_node.StorageVolume = value
	}
	if value, ok := fc.mutation.StoragePath(); ok {
		_spec.SetField(file.FieldStoragePath, field.TypeString, value)
		_node.StoragePath = value
	}
	if value, ok := fc.mutation.FileContents(); ok {
		_spec.SetField(file.FieldFileContents, field.TypeBytes, value)
		_node.FileContents = value
	}
	if nodes := fc.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.UserTable,
			Columns: file.UserPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.UserFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.OrganizationIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.OrganizationTable,
			Columns: file.OrganizationPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organization.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.OrganizationFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.GroupsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.GroupsTable,
			Columns: file.GroupsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(group.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.GroupFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.ContactIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.ContactTable,
			Columns: file.ContactPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(contact.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.ContactFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.EntityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.EntityTable,
			Columns: file.EntityPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(entity.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.EntityFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.UserSettingIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.UserSettingTable,
			Columns: file.UserSettingPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(usersetting.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.UserSettingFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.OrganizationSettingIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.OrganizationSettingTable,
			Columns: file.OrganizationSettingPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(organizationsetting.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.OrganizationSettingFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.TemplateIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.TemplateTable,
			Columns: file.TemplatePrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(template.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.TemplateFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.DocumentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.DocumentTable,
			Columns: file.DocumentPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(documentdata.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.DocumentDataFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.ProgramIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.ProgramTable,
			Columns: file.ProgramPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(program.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.ProgramFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.EvidenceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.EvidenceTable,
			Columns: file.EvidencePrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(evidence.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.EvidenceFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.EventsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   file.EventsTable,
			Columns: file.EventsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(event.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.FileEvents
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.TrustCenterSettingIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.TrustCenterSettingTable,
			Columns: file.TrustCenterSettingPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(trustcentersetting.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.TrustCenterSettingFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := fc.mutation.SubprocessorIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   file.SubprocessorTable,
			Columns: file.SubprocessorPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(subprocessor.FieldID, field.TypeString),
			},
		}
		edge.Schema = fc.schemaConfig.SubprocessorFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// FileCreateBulk is the builder for creating many File entities in bulk.
type FileCreateBulk struct {
	config
	err      error
	builders []*FileCreate
}

// Save creates the File entities in the database.
func (fcb *FileCreateBulk) Save(ctx context.Context) ([]*File, error) {
	if fcb.err != nil {
		return nil, fcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(fcb.builders))
	nodes := make([]*File, len(fcb.builders))
	mutators := make([]Mutator, len(fcb.builders))
	for i := range fcb.builders {
		func(i int, root context.Context) {
			builder := fcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*FileMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, fcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, fcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, fcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (fcb *FileCreateBulk) SaveX(ctx context.Context) []*File {
	v, err := fcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (fcb *FileCreateBulk) Exec(ctx context.Context) error {
	_, err := fcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fcb *FileCreateBulk) ExecX(ctx context.Context) {
	if err := fcb.Exec(ctx); err != nil {
		panic(err)
	}
}
