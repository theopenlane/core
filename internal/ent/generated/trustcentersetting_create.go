// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/pkg/enums"
)

// TrustCenterSettingCreate is the builder for creating a TrustCenterSetting entity.
type TrustCenterSettingCreate struct {
	config
	mutation *TrustCenterSettingMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (tcsc *TrustCenterSettingCreate) SetCreatedAt(t time.Time) *TrustCenterSettingCreate {
	tcsc.mutation.SetCreatedAt(t)
	return tcsc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableCreatedAt(t *time.Time) *TrustCenterSettingCreate {
	if t != nil {
		tcsc.SetCreatedAt(*t)
	}
	return tcsc
}

// SetUpdatedAt sets the "updated_at" field.
func (tcsc *TrustCenterSettingCreate) SetUpdatedAt(t time.Time) *TrustCenterSettingCreate {
	tcsc.mutation.SetUpdatedAt(t)
	return tcsc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableUpdatedAt(t *time.Time) *TrustCenterSettingCreate {
	if t != nil {
		tcsc.SetUpdatedAt(*t)
	}
	return tcsc
}

// SetCreatedBy sets the "created_by" field.
func (tcsc *TrustCenterSettingCreate) SetCreatedBy(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetCreatedBy(s)
	return tcsc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableCreatedBy(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetCreatedBy(*s)
	}
	return tcsc
}

// SetUpdatedBy sets the "updated_by" field.
func (tcsc *TrustCenterSettingCreate) SetUpdatedBy(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetUpdatedBy(s)
	return tcsc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableUpdatedBy(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetUpdatedBy(*s)
	}
	return tcsc
}

// SetDeletedAt sets the "deleted_at" field.
func (tcsc *TrustCenterSettingCreate) SetDeletedAt(t time.Time) *TrustCenterSettingCreate {
	tcsc.mutation.SetDeletedAt(t)
	return tcsc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableDeletedAt(t *time.Time) *TrustCenterSettingCreate {
	if t != nil {
		tcsc.SetDeletedAt(*t)
	}
	return tcsc
}

// SetDeletedBy sets the "deleted_by" field.
func (tcsc *TrustCenterSettingCreate) SetDeletedBy(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetDeletedBy(s)
	return tcsc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableDeletedBy(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetDeletedBy(*s)
	}
	return tcsc
}

// SetTrustCenterID sets the "trust_center_id" field.
func (tcsc *TrustCenterSettingCreate) SetTrustCenterID(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetTrustCenterID(s)
	return tcsc
}

// SetNillableTrustCenterID sets the "trust_center_id" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableTrustCenterID(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetTrustCenterID(*s)
	}
	return tcsc
}

// SetTitle sets the "title" field.
func (tcsc *TrustCenterSettingCreate) SetTitle(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetTitle(s)
	return tcsc
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableTitle(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetTitle(*s)
	}
	return tcsc
}

// SetOverview sets the "overview" field.
func (tcsc *TrustCenterSettingCreate) SetOverview(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetOverview(s)
	return tcsc
}

// SetNillableOverview sets the "overview" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableOverview(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetOverview(*s)
	}
	return tcsc
}

// SetLogoRemoteURL sets the "logo_remote_url" field.
func (tcsc *TrustCenterSettingCreate) SetLogoRemoteURL(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetLogoRemoteURL(s)
	return tcsc
}

// SetNillableLogoRemoteURL sets the "logo_remote_url" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableLogoRemoteURL(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetLogoRemoteURL(*s)
	}
	return tcsc
}

// SetLogoLocalFileID sets the "logo_local_file_id" field.
func (tcsc *TrustCenterSettingCreate) SetLogoLocalFileID(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetLogoLocalFileID(s)
	return tcsc
}

// SetNillableLogoLocalFileID sets the "logo_local_file_id" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableLogoLocalFileID(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetLogoLocalFileID(*s)
	}
	return tcsc
}

// SetFaviconRemoteURL sets the "favicon_remote_url" field.
func (tcsc *TrustCenterSettingCreate) SetFaviconRemoteURL(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetFaviconRemoteURL(s)
	return tcsc
}

// SetNillableFaviconRemoteURL sets the "favicon_remote_url" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableFaviconRemoteURL(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetFaviconRemoteURL(*s)
	}
	return tcsc
}

// SetFaviconLocalFileID sets the "favicon_local_file_id" field.
func (tcsc *TrustCenterSettingCreate) SetFaviconLocalFileID(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetFaviconLocalFileID(s)
	return tcsc
}

// SetNillableFaviconLocalFileID sets the "favicon_local_file_id" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableFaviconLocalFileID(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetFaviconLocalFileID(*s)
	}
	return tcsc
}

// SetThemeMode sets the "theme_mode" field.
func (tcsc *TrustCenterSettingCreate) SetThemeMode(ectm enums.TrustCenterThemeMode) *TrustCenterSettingCreate {
	tcsc.mutation.SetThemeMode(ectm)
	return tcsc
}

// SetNillableThemeMode sets the "theme_mode" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableThemeMode(ectm *enums.TrustCenterThemeMode) *TrustCenterSettingCreate {
	if ectm != nil {
		tcsc.SetThemeMode(*ectm)
	}
	return tcsc
}

// SetPrimaryColor sets the "primary_color" field.
func (tcsc *TrustCenterSettingCreate) SetPrimaryColor(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetPrimaryColor(s)
	return tcsc
}

// SetNillablePrimaryColor sets the "primary_color" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillablePrimaryColor(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetPrimaryColor(*s)
	}
	return tcsc
}

// SetFont sets the "font" field.
func (tcsc *TrustCenterSettingCreate) SetFont(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetFont(s)
	return tcsc
}

// SetNillableFont sets the "font" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableFont(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetFont(*s)
	}
	return tcsc
}

// SetForegroundColor sets the "foreground_color" field.
func (tcsc *TrustCenterSettingCreate) SetForegroundColor(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetForegroundColor(s)
	return tcsc
}

// SetNillableForegroundColor sets the "foreground_color" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableForegroundColor(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetForegroundColor(*s)
	}
	return tcsc
}

// SetBackgroundColor sets the "background_color" field.
func (tcsc *TrustCenterSettingCreate) SetBackgroundColor(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetBackgroundColor(s)
	return tcsc
}

// SetNillableBackgroundColor sets the "background_color" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableBackgroundColor(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetBackgroundColor(*s)
	}
	return tcsc
}

// SetAccentColor sets the "accent_color" field.
func (tcsc *TrustCenterSettingCreate) SetAccentColor(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetAccentColor(s)
	return tcsc
}

// SetNillableAccentColor sets the "accent_color" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableAccentColor(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetAccentColor(*s)
	}
	return tcsc
}

// SetID sets the "id" field.
func (tcsc *TrustCenterSettingCreate) SetID(s string) *TrustCenterSettingCreate {
	tcsc.mutation.SetID(s)
	return tcsc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableID(s *string) *TrustCenterSettingCreate {
	if s != nil {
		tcsc.SetID(*s)
	}
	return tcsc
}

// SetTrustCenter sets the "trust_center" edge to the TrustCenter entity.
func (tcsc *TrustCenterSettingCreate) SetTrustCenter(t *TrustCenter) *TrustCenterSettingCreate {
	return tcsc.SetTrustCenterID(t.ID)
}

// AddFileIDs adds the "files" edge to the File entity by IDs.
func (tcsc *TrustCenterSettingCreate) AddFileIDs(ids ...string) *TrustCenterSettingCreate {
	tcsc.mutation.AddFileIDs(ids...)
	return tcsc
}

// AddFiles adds the "files" edges to the File entity.
func (tcsc *TrustCenterSettingCreate) AddFiles(f ...*File) *TrustCenterSettingCreate {
	ids := make([]string, len(f))
	for i := range f {
		ids[i] = f[i].ID
	}
	return tcsc.AddFileIDs(ids...)
}

// SetLogoFileID sets the "logo_file" edge to the File entity by ID.
func (tcsc *TrustCenterSettingCreate) SetLogoFileID(id string) *TrustCenterSettingCreate {
	tcsc.mutation.SetLogoFileID(id)
	return tcsc
}

// SetNillableLogoFileID sets the "logo_file" edge to the File entity by ID if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableLogoFileID(id *string) *TrustCenterSettingCreate {
	if id != nil {
		tcsc = tcsc.SetLogoFileID(*id)
	}
	return tcsc
}

// SetLogoFile sets the "logo_file" edge to the File entity.
func (tcsc *TrustCenterSettingCreate) SetLogoFile(f *File) *TrustCenterSettingCreate {
	return tcsc.SetLogoFileID(f.ID)
}

// SetFaviconFileID sets the "favicon_file" edge to the File entity by ID.
func (tcsc *TrustCenterSettingCreate) SetFaviconFileID(id string) *TrustCenterSettingCreate {
	tcsc.mutation.SetFaviconFileID(id)
	return tcsc
}

// SetNillableFaviconFileID sets the "favicon_file" edge to the File entity by ID if the given value is not nil.
func (tcsc *TrustCenterSettingCreate) SetNillableFaviconFileID(id *string) *TrustCenterSettingCreate {
	if id != nil {
		tcsc = tcsc.SetFaviconFileID(*id)
	}
	return tcsc
}

// SetFaviconFile sets the "favicon_file" edge to the File entity.
func (tcsc *TrustCenterSettingCreate) SetFaviconFile(f *File) *TrustCenterSettingCreate {
	return tcsc.SetFaviconFileID(f.ID)
}

// Mutation returns the TrustCenterSettingMutation object of the builder.
func (tcsc *TrustCenterSettingCreate) Mutation() *TrustCenterSettingMutation {
	return tcsc.mutation
}

// Save creates the TrustCenterSetting in the database.
func (tcsc *TrustCenterSettingCreate) Save(ctx context.Context) (*TrustCenterSetting, error) {
	if err := tcsc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, tcsc.sqlSave, tcsc.mutation, tcsc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (tcsc *TrustCenterSettingCreate) SaveX(ctx context.Context) *TrustCenterSetting {
	v, err := tcsc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tcsc *TrustCenterSettingCreate) Exec(ctx context.Context) error {
	_, err := tcsc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcsc *TrustCenterSettingCreate) ExecX(ctx context.Context) {
	if err := tcsc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tcsc *TrustCenterSettingCreate) defaults() error {
	if _, ok := tcsc.mutation.CreatedAt(); !ok {
		if trustcentersetting.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentersetting.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentersetting.DefaultCreatedAt()
		tcsc.mutation.SetCreatedAt(v)
	}
	if _, ok := tcsc.mutation.UpdatedAt(); !ok {
		if trustcentersetting.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentersetting.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentersetting.DefaultUpdatedAt()
		tcsc.mutation.SetUpdatedAt(v)
	}
	if _, ok := tcsc.mutation.ThemeMode(); !ok {
		v := trustcentersetting.DefaultThemeMode
		tcsc.mutation.SetThemeMode(v)
	}
	if _, ok := tcsc.mutation.ID(); !ok {
		if trustcentersetting.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized trustcentersetting.DefaultID (forgotten import generated/runtime?)")
		}
		v := trustcentersetting.DefaultID()
		tcsc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (tcsc *TrustCenterSettingCreate) check() error {
	if v, ok := tcsc.mutation.TrustCenterID(); ok {
		if err := trustcentersetting.TrustCenterIDValidator(v); err != nil {
			return &ValidationError{Name: "trust_center_id", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSetting.trust_center_id": %w`, err)}
		}
	}
	if v, ok := tcsc.mutation.Title(); ok {
		if err := trustcentersetting.TitleValidator(v); err != nil {
			return &ValidationError{Name: "title", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSetting.title": %w`, err)}
		}
	}
	if v, ok := tcsc.mutation.Overview(); ok {
		if err := trustcentersetting.OverviewValidator(v); err != nil {
			return &ValidationError{Name: "overview", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSetting.overview": %w`, err)}
		}
	}
	if v, ok := tcsc.mutation.LogoRemoteURL(); ok {
		if err := trustcentersetting.LogoRemoteURLValidator(v); err != nil {
			return &ValidationError{Name: "logo_remote_url", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSetting.logo_remote_url": %w`, err)}
		}
	}
	if v, ok := tcsc.mutation.FaviconRemoteURL(); ok {
		if err := trustcentersetting.FaviconRemoteURLValidator(v); err != nil {
			return &ValidationError{Name: "favicon_remote_url", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSetting.favicon_remote_url": %w`, err)}
		}
	}
	if v, ok := tcsc.mutation.ThemeMode(); ok {
		if err := trustcentersetting.ThemeModeValidator(v); err != nil {
			return &ValidationError{Name: "theme_mode", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSetting.theme_mode": %w`, err)}
		}
	}
	return nil
}

func (tcsc *TrustCenterSettingCreate) sqlSave(ctx context.Context) (*TrustCenterSetting, error) {
	if err := tcsc.check(); err != nil {
		return nil, err
	}
	_node, _spec := tcsc.createSpec()
	if err := sqlgraph.CreateNode(ctx, tcsc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected TrustCenterSetting.ID type: %T", _spec.ID.Value)
		}
	}
	tcsc.mutation.id = &_node.ID
	tcsc.mutation.done = true
	return _node, nil
}

func (tcsc *TrustCenterSettingCreate) createSpec() (*TrustCenterSetting, *sqlgraph.CreateSpec) {
	var (
		_node = &TrustCenterSetting{config: tcsc.config}
		_spec = sqlgraph.NewCreateSpec(trustcentersetting.Table, sqlgraph.NewFieldSpec(trustcentersetting.FieldID, field.TypeString))
	)
	_spec.Schema = tcsc.schemaConfig.TrustCenterSetting
	if id, ok := tcsc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := tcsc.mutation.CreatedAt(); ok {
		_spec.SetField(trustcentersetting.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := tcsc.mutation.UpdatedAt(); ok {
		_spec.SetField(trustcentersetting.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := tcsc.mutation.CreatedBy(); ok {
		_spec.SetField(trustcentersetting.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := tcsc.mutation.UpdatedBy(); ok {
		_spec.SetField(trustcentersetting.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := tcsc.mutation.DeletedAt(); ok {
		_spec.SetField(trustcentersetting.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := tcsc.mutation.DeletedBy(); ok {
		_spec.SetField(trustcentersetting.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := tcsc.mutation.Title(); ok {
		_spec.SetField(trustcentersetting.FieldTitle, field.TypeString, value)
		_node.Title = value
	}
	if value, ok := tcsc.mutation.Overview(); ok {
		_spec.SetField(trustcentersetting.FieldOverview, field.TypeString, value)
		_node.Overview = value
	}
	if value, ok := tcsc.mutation.LogoRemoteURL(); ok {
		_spec.SetField(trustcentersetting.FieldLogoRemoteURL, field.TypeString, value)
		_node.LogoRemoteURL = &value
	}
	if value, ok := tcsc.mutation.FaviconRemoteURL(); ok {
		_spec.SetField(trustcentersetting.FieldFaviconRemoteURL, field.TypeString, value)
		_node.FaviconRemoteURL = &value
	}
	if value, ok := tcsc.mutation.ThemeMode(); ok {
		_spec.SetField(trustcentersetting.FieldThemeMode, field.TypeEnum, value)
		_node.ThemeMode = value
	}
	if value, ok := tcsc.mutation.PrimaryColor(); ok {
		_spec.SetField(trustcentersetting.FieldPrimaryColor, field.TypeString, value)
		_node.PrimaryColor = value
	}
	if value, ok := tcsc.mutation.Font(); ok {
		_spec.SetField(trustcentersetting.FieldFont, field.TypeString, value)
		_node.Font = value
	}
	if value, ok := tcsc.mutation.ForegroundColor(); ok {
		_spec.SetField(trustcentersetting.FieldForegroundColor, field.TypeString, value)
		_node.ForegroundColor = value
	}
	if value, ok := tcsc.mutation.BackgroundColor(); ok {
		_spec.SetField(trustcentersetting.FieldBackgroundColor, field.TypeString, value)
		_node.BackgroundColor = value
	}
	if value, ok := tcsc.mutation.AccentColor(); ok {
		_spec.SetField(trustcentersetting.FieldAccentColor, field.TypeString, value)
		_node.AccentColor = value
	}
	if nodes := tcsc.mutation.TrustCenterIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   trustcentersetting.TrustCenterTable,
			Columns: []string{trustcentersetting.TrustCenterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(trustcenter.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsc.schemaConfig.TrustCenterSetting
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.TrustCenterID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tcsc.mutation.FilesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   trustcentersetting.FilesTable,
			Columns: trustcentersetting.FilesPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsc.schemaConfig.TrustCenterSettingFiles
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tcsc.mutation.LogoFileIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   trustcentersetting.LogoFileTable,
			Columns: []string{trustcentersetting.LogoFileColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsc.schemaConfig.TrustCenterSetting
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.LogoLocalFileID = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tcsc.mutation.FaviconFileIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   trustcentersetting.FaviconFileTable,
			Columns: []string{trustcentersetting.FaviconFileColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(file.FieldID, field.TypeString),
			},
		}
		edge.Schema = tcsc.schemaConfig.TrustCenterSetting
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.FaviconLocalFileID = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// TrustCenterSettingCreateBulk is the builder for creating many TrustCenterSetting entities in bulk.
type TrustCenterSettingCreateBulk struct {
	config
	err      error
	builders []*TrustCenterSettingCreate
}

// Save creates the TrustCenterSetting entities in the database.
func (tcscb *TrustCenterSettingCreateBulk) Save(ctx context.Context) ([]*TrustCenterSetting, error) {
	if tcscb.err != nil {
		return nil, tcscb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(tcscb.builders))
	nodes := make([]*TrustCenterSetting, len(tcscb.builders))
	mutators := make([]Mutator, len(tcscb.builders))
	for i := range tcscb.builders {
		func(i int, root context.Context) {
			builder := tcscb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TrustCenterSettingMutation)
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
					_, err = mutators[i+1].Mutate(root, tcscb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tcscb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, tcscb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tcscb *TrustCenterSettingCreateBulk) SaveX(ctx context.Context) []*TrustCenterSetting {
	v, err := tcscb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tcscb *TrustCenterSettingCreateBulk) Exec(ctx context.Context) error {
	_, err := tcscb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcscb *TrustCenterSettingCreateBulk) ExecX(ctx context.Context) {
	if err := tcscb.Exec(ctx); err != nil {
		panic(err)
	}
}
