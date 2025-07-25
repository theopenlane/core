// Code generated by ent, DO NOT EDIT.

package generated

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersettinghistory"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx/history"
)

// TrustCenterSettingHistoryCreate is the builder for creating a TrustCenterSettingHistory entity.
type TrustCenterSettingHistoryCreate struct {
	config
	mutation *TrustCenterSettingHistoryMutation
	hooks    []Hook
}

// SetHistoryTime sets the "history_time" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetHistoryTime(t time.Time) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetHistoryTime(t)
	return tcshc
}

// SetNillableHistoryTime sets the "history_time" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableHistoryTime(t *time.Time) *TrustCenterSettingHistoryCreate {
	if t != nil {
		tcshc.SetHistoryTime(*t)
	}
	return tcshc
}

// SetRef sets the "ref" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetRef(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetRef(s)
	return tcshc
}

// SetNillableRef sets the "ref" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableRef(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetRef(*s)
	}
	return tcshc
}

// SetOperation sets the "operation" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetOperation(ht history.OpType) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetOperation(ht)
	return tcshc
}

// SetCreatedAt sets the "created_at" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetCreatedAt(t time.Time) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetCreatedAt(t)
	return tcshc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableCreatedAt(t *time.Time) *TrustCenterSettingHistoryCreate {
	if t != nil {
		tcshc.SetCreatedAt(*t)
	}
	return tcshc
}

// SetUpdatedAt sets the "updated_at" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetUpdatedAt(t time.Time) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetUpdatedAt(t)
	return tcshc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableUpdatedAt(t *time.Time) *TrustCenterSettingHistoryCreate {
	if t != nil {
		tcshc.SetUpdatedAt(*t)
	}
	return tcshc
}

// SetCreatedBy sets the "created_by" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetCreatedBy(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetCreatedBy(s)
	return tcshc
}

// SetNillableCreatedBy sets the "created_by" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableCreatedBy(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetCreatedBy(*s)
	}
	return tcshc
}

// SetUpdatedBy sets the "updated_by" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetUpdatedBy(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetUpdatedBy(s)
	return tcshc
}

// SetNillableUpdatedBy sets the "updated_by" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableUpdatedBy(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetUpdatedBy(*s)
	}
	return tcshc
}

// SetDeletedAt sets the "deleted_at" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetDeletedAt(t time.Time) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetDeletedAt(t)
	return tcshc
}

// SetNillableDeletedAt sets the "deleted_at" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableDeletedAt(t *time.Time) *TrustCenterSettingHistoryCreate {
	if t != nil {
		tcshc.SetDeletedAt(*t)
	}
	return tcshc
}

// SetDeletedBy sets the "deleted_by" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetDeletedBy(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetDeletedBy(s)
	return tcshc
}

// SetNillableDeletedBy sets the "deleted_by" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableDeletedBy(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetDeletedBy(*s)
	}
	return tcshc
}

// SetTrustCenterID sets the "trust_center_id" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetTrustCenterID(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetTrustCenterID(s)
	return tcshc
}

// SetNillableTrustCenterID sets the "trust_center_id" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableTrustCenterID(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetTrustCenterID(*s)
	}
	return tcshc
}

// SetTitle sets the "title" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetTitle(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetTitle(s)
	return tcshc
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableTitle(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetTitle(*s)
	}
	return tcshc
}

// SetOverview sets the "overview" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetOverview(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetOverview(s)
	return tcshc
}

// SetNillableOverview sets the "overview" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableOverview(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetOverview(*s)
	}
	return tcshc
}

// SetLogoRemoteURL sets the "logo_remote_url" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetLogoRemoteURL(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetLogoRemoteURL(s)
	return tcshc
}

// SetNillableLogoRemoteURL sets the "logo_remote_url" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableLogoRemoteURL(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetLogoRemoteURL(*s)
	}
	return tcshc
}

// SetLogoLocalFileID sets the "logo_local_file_id" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetLogoLocalFileID(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetLogoLocalFileID(s)
	return tcshc
}

// SetNillableLogoLocalFileID sets the "logo_local_file_id" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableLogoLocalFileID(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetLogoLocalFileID(*s)
	}
	return tcshc
}

// SetFaviconRemoteURL sets the "favicon_remote_url" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetFaviconRemoteURL(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetFaviconRemoteURL(s)
	return tcshc
}

// SetNillableFaviconRemoteURL sets the "favicon_remote_url" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableFaviconRemoteURL(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetFaviconRemoteURL(*s)
	}
	return tcshc
}

// SetFaviconLocalFileID sets the "favicon_local_file_id" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetFaviconLocalFileID(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetFaviconLocalFileID(s)
	return tcshc
}

// SetNillableFaviconLocalFileID sets the "favicon_local_file_id" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableFaviconLocalFileID(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetFaviconLocalFileID(*s)
	}
	return tcshc
}

// SetThemeMode sets the "theme_mode" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetThemeMode(ectm enums.TrustCenterThemeMode) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetThemeMode(ectm)
	return tcshc
}

// SetNillableThemeMode sets the "theme_mode" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableThemeMode(ectm *enums.TrustCenterThemeMode) *TrustCenterSettingHistoryCreate {
	if ectm != nil {
		tcshc.SetThemeMode(*ectm)
	}
	return tcshc
}

// SetPrimaryColor sets the "primary_color" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetPrimaryColor(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetPrimaryColor(s)
	return tcshc
}

// SetNillablePrimaryColor sets the "primary_color" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillablePrimaryColor(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetPrimaryColor(*s)
	}
	return tcshc
}

// SetFont sets the "font" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetFont(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetFont(s)
	return tcshc
}

// SetNillableFont sets the "font" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableFont(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetFont(*s)
	}
	return tcshc
}

// SetForegroundColor sets the "foreground_color" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetForegroundColor(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetForegroundColor(s)
	return tcshc
}

// SetNillableForegroundColor sets the "foreground_color" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableForegroundColor(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetForegroundColor(*s)
	}
	return tcshc
}

// SetBackgroundColor sets the "background_color" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetBackgroundColor(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetBackgroundColor(s)
	return tcshc
}

// SetNillableBackgroundColor sets the "background_color" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableBackgroundColor(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetBackgroundColor(*s)
	}
	return tcshc
}

// SetAccentColor sets the "accent_color" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetAccentColor(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetAccentColor(s)
	return tcshc
}

// SetNillableAccentColor sets the "accent_color" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableAccentColor(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetAccentColor(*s)
	}
	return tcshc
}

// SetID sets the "id" field.
func (tcshc *TrustCenterSettingHistoryCreate) SetID(s string) *TrustCenterSettingHistoryCreate {
	tcshc.mutation.SetID(s)
	return tcshc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (tcshc *TrustCenterSettingHistoryCreate) SetNillableID(s *string) *TrustCenterSettingHistoryCreate {
	if s != nil {
		tcshc.SetID(*s)
	}
	return tcshc
}

// Mutation returns the TrustCenterSettingHistoryMutation object of the builder.
func (tcshc *TrustCenterSettingHistoryCreate) Mutation() *TrustCenterSettingHistoryMutation {
	return tcshc.mutation
}

// Save creates the TrustCenterSettingHistory in the database.
func (tcshc *TrustCenterSettingHistoryCreate) Save(ctx context.Context) (*TrustCenterSettingHistory, error) {
	if err := tcshc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, tcshc.sqlSave, tcshc.mutation, tcshc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (tcshc *TrustCenterSettingHistoryCreate) SaveX(ctx context.Context) *TrustCenterSettingHistory {
	v, err := tcshc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tcshc *TrustCenterSettingHistoryCreate) Exec(ctx context.Context) error {
	_, err := tcshc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcshc *TrustCenterSettingHistoryCreate) ExecX(ctx context.Context) {
	if err := tcshc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tcshc *TrustCenterSettingHistoryCreate) defaults() error {
	if _, ok := tcshc.mutation.HistoryTime(); !ok {
		if trustcentersettinghistory.DefaultHistoryTime == nil {
			return fmt.Errorf("generated: uninitialized trustcentersettinghistory.DefaultHistoryTime (forgotten import generated/runtime?)")
		}
		v := trustcentersettinghistory.DefaultHistoryTime()
		tcshc.mutation.SetHistoryTime(v)
	}
	if _, ok := tcshc.mutation.CreatedAt(); !ok {
		if trustcentersettinghistory.DefaultCreatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentersettinghistory.DefaultCreatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentersettinghistory.DefaultCreatedAt()
		tcshc.mutation.SetCreatedAt(v)
	}
	if _, ok := tcshc.mutation.UpdatedAt(); !ok {
		if trustcentersettinghistory.DefaultUpdatedAt == nil {
			return fmt.Errorf("generated: uninitialized trustcentersettinghistory.DefaultUpdatedAt (forgotten import generated/runtime?)")
		}
		v := trustcentersettinghistory.DefaultUpdatedAt()
		tcshc.mutation.SetUpdatedAt(v)
	}
	if _, ok := tcshc.mutation.ThemeMode(); !ok {
		v := trustcentersettinghistory.DefaultThemeMode
		tcshc.mutation.SetThemeMode(v)
	}
	if _, ok := tcshc.mutation.ID(); !ok {
		if trustcentersettinghistory.DefaultID == nil {
			return fmt.Errorf("generated: uninitialized trustcentersettinghistory.DefaultID (forgotten import generated/runtime?)")
		}
		v := trustcentersettinghistory.DefaultID()
		tcshc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (tcshc *TrustCenterSettingHistoryCreate) check() error {
	if _, ok := tcshc.mutation.HistoryTime(); !ok {
		return &ValidationError{Name: "history_time", err: errors.New(`generated: missing required field "TrustCenterSettingHistory.history_time"`)}
	}
	if _, ok := tcshc.mutation.Operation(); !ok {
		return &ValidationError{Name: "operation", err: errors.New(`generated: missing required field "TrustCenterSettingHistory.operation"`)}
	}
	if v, ok := tcshc.mutation.Operation(); ok {
		if err := trustcentersettinghistory.OperationValidator(v); err != nil {
			return &ValidationError{Name: "operation", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSettingHistory.operation": %w`, err)}
		}
	}
	if v, ok := tcshc.mutation.ThemeMode(); ok {
		if err := trustcentersettinghistory.ThemeModeValidator(v); err != nil {
			return &ValidationError{Name: "theme_mode", err: fmt.Errorf(`generated: validator failed for field "TrustCenterSettingHistory.theme_mode": %w`, err)}
		}
	}
	return nil
}

func (tcshc *TrustCenterSettingHistoryCreate) sqlSave(ctx context.Context) (*TrustCenterSettingHistory, error) {
	if err := tcshc.check(); err != nil {
		return nil, err
	}
	_node, _spec := tcshc.createSpec()
	if err := sqlgraph.CreateNode(ctx, tcshc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(string); ok {
			_node.ID = id
		} else {
			return nil, fmt.Errorf("unexpected TrustCenterSettingHistory.ID type: %T", _spec.ID.Value)
		}
	}
	tcshc.mutation.id = &_node.ID
	tcshc.mutation.done = true
	return _node, nil
}

func (tcshc *TrustCenterSettingHistoryCreate) createSpec() (*TrustCenterSettingHistory, *sqlgraph.CreateSpec) {
	var (
		_node = &TrustCenterSettingHistory{config: tcshc.config}
		_spec = sqlgraph.NewCreateSpec(trustcentersettinghistory.Table, sqlgraph.NewFieldSpec(trustcentersettinghistory.FieldID, field.TypeString))
	)
	_spec.Schema = tcshc.schemaConfig.TrustCenterSettingHistory
	if id, ok := tcshc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := tcshc.mutation.HistoryTime(); ok {
		_spec.SetField(trustcentersettinghistory.FieldHistoryTime, field.TypeTime, value)
		_node.HistoryTime = value
	}
	if value, ok := tcshc.mutation.Ref(); ok {
		_spec.SetField(trustcentersettinghistory.FieldRef, field.TypeString, value)
		_node.Ref = value
	}
	if value, ok := tcshc.mutation.Operation(); ok {
		_spec.SetField(trustcentersettinghistory.FieldOperation, field.TypeEnum, value)
		_node.Operation = value
	}
	if value, ok := tcshc.mutation.CreatedAt(); ok {
		_spec.SetField(trustcentersettinghistory.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := tcshc.mutation.UpdatedAt(); ok {
		_spec.SetField(trustcentersettinghistory.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := tcshc.mutation.CreatedBy(); ok {
		_spec.SetField(trustcentersettinghistory.FieldCreatedBy, field.TypeString, value)
		_node.CreatedBy = value
	}
	if value, ok := tcshc.mutation.UpdatedBy(); ok {
		_spec.SetField(trustcentersettinghistory.FieldUpdatedBy, field.TypeString, value)
		_node.UpdatedBy = value
	}
	if value, ok := tcshc.mutation.DeletedAt(); ok {
		_spec.SetField(trustcentersettinghistory.FieldDeletedAt, field.TypeTime, value)
		_node.DeletedAt = value
	}
	if value, ok := tcshc.mutation.DeletedBy(); ok {
		_spec.SetField(trustcentersettinghistory.FieldDeletedBy, field.TypeString, value)
		_node.DeletedBy = value
	}
	if value, ok := tcshc.mutation.TrustCenterID(); ok {
		_spec.SetField(trustcentersettinghistory.FieldTrustCenterID, field.TypeString, value)
		_node.TrustCenterID = value
	}
	if value, ok := tcshc.mutation.Title(); ok {
		_spec.SetField(trustcentersettinghistory.FieldTitle, field.TypeString, value)
		_node.Title = value
	}
	if value, ok := tcshc.mutation.Overview(); ok {
		_spec.SetField(trustcentersettinghistory.FieldOverview, field.TypeString, value)
		_node.Overview = value
	}
	if value, ok := tcshc.mutation.LogoRemoteURL(); ok {
		_spec.SetField(trustcentersettinghistory.FieldLogoRemoteURL, field.TypeString, value)
		_node.LogoRemoteURL = &value
	}
	if value, ok := tcshc.mutation.LogoLocalFileID(); ok {
		_spec.SetField(trustcentersettinghistory.FieldLogoLocalFileID, field.TypeString, value)
		_node.LogoLocalFileID = &value
	}
	if value, ok := tcshc.mutation.FaviconRemoteURL(); ok {
		_spec.SetField(trustcentersettinghistory.FieldFaviconRemoteURL, field.TypeString, value)
		_node.FaviconRemoteURL = &value
	}
	if value, ok := tcshc.mutation.FaviconLocalFileID(); ok {
		_spec.SetField(trustcentersettinghistory.FieldFaviconLocalFileID, field.TypeString, value)
		_node.FaviconLocalFileID = &value
	}
	if value, ok := tcshc.mutation.ThemeMode(); ok {
		_spec.SetField(trustcentersettinghistory.FieldThemeMode, field.TypeEnum, value)
		_node.ThemeMode = value
	}
	if value, ok := tcshc.mutation.PrimaryColor(); ok {
		_spec.SetField(trustcentersettinghistory.FieldPrimaryColor, field.TypeString, value)
		_node.PrimaryColor = value
	}
	if value, ok := tcshc.mutation.Font(); ok {
		_spec.SetField(trustcentersettinghistory.FieldFont, field.TypeString, value)
		_node.Font = value
	}
	if value, ok := tcshc.mutation.ForegroundColor(); ok {
		_spec.SetField(trustcentersettinghistory.FieldForegroundColor, field.TypeString, value)
		_node.ForegroundColor = value
	}
	if value, ok := tcshc.mutation.BackgroundColor(); ok {
		_spec.SetField(trustcentersettinghistory.FieldBackgroundColor, field.TypeString, value)
		_node.BackgroundColor = value
	}
	if value, ok := tcshc.mutation.AccentColor(); ok {
		_spec.SetField(trustcentersettinghistory.FieldAccentColor, field.TypeString, value)
		_node.AccentColor = value
	}
	return _node, _spec
}

// TrustCenterSettingHistoryCreateBulk is the builder for creating many TrustCenterSettingHistory entities in bulk.
type TrustCenterSettingHistoryCreateBulk struct {
	config
	err      error
	builders []*TrustCenterSettingHistoryCreate
}

// Save creates the TrustCenterSettingHistory entities in the database.
func (tcshcb *TrustCenterSettingHistoryCreateBulk) Save(ctx context.Context) ([]*TrustCenterSettingHistory, error) {
	if tcshcb.err != nil {
		return nil, tcshcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(tcshcb.builders))
	nodes := make([]*TrustCenterSettingHistory, len(tcshcb.builders))
	mutators := make([]Mutator, len(tcshcb.builders))
	for i := range tcshcb.builders {
		func(i int, root context.Context) {
			builder := tcshcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TrustCenterSettingHistoryMutation)
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
					_, err = mutators[i+1].Mutate(root, tcshcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tcshcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, tcshcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tcshcb *TrustCenterSettingHistoryCreateBulk) SaveX(ctx context.Context) []*TrustCenterSettingHistory {
	v, err := tcshcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tcshcb *TrustCenterSettingHistoryCreateBulk) Exec(ctx context.Context) error {
	_, err := tcshcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcshcb *TrustCenterSettingHistoryCreateBulk) ExecX(ctx context.Context) {
	if err := tcshcb.Exec(ctx); err != nil {
		panic(err)
	}
}
