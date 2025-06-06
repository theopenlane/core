// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlgenerated

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/vektah/gqlparser/v2/ast"
)

// region    ************************** generated!.gotpl **************************

// endregion ************************** generated!.gotpl **************************

// region    ***************************** args.gotpl *****************************

// endregion ***************************** args.gotpl *****************************

// region    ************************** directives.gotpl **************************

// endregion ************************** directives.gotpl **************************

// region    **************************** field.gotpl *****************************

func (ec *executionContext) _UserSettingBulkCreatePayload_userSettings(ctx context.Context, field graphql.CollectedField, obj *model.UserSettingBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_UserSettingBulkCreatePayload_userSettings(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (any, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.UserSettings, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.UserSetting)
	fc.Result = res
	return ec.marshalOUserSetting2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐUserSettingᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_UserSettingBulkCreatePayload_userSettings(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "UserSettingBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_UserSetting_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_UserSetting_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_UserSetting_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_UserSetting_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_UserSetting_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_UserSetting_tags(ctx, field)
			case "userID":
				return ec.fieldContext_UserSetting_userID(ctx, field)
			case "locked":
				return ec.fieldContext_UserSetting_locked(ctx, field)
			case "silencedAt":
				return ec.fieldContext_UserSetting_silencedAt(ctx, field)
			case "suspendedAt":
				return ec.fieldContext_UserSetting_suspendedAt(ctx, field)
			case "status":
				return ec.fieldContext_UserSetting_status(ctx, field)
			case "emailConfirmed":
				return ec.fieldContext_UserSetting_emailConfirmed(ctx, field)
			case "isWebauthnAllowed":
				return ec.fieldContext_UserSetting_isWebauthnAllowed(ctx, field)
			case "isTfaEnabled":
				return ec.fieldContext_UserSetting_isTfaEnabled(ctx, field)
			case "user":
				return ec.fieldContext_UserSetting_user(ctx, field)
			case "defaultOrg":
				return ec.fieldContext_UserSetting_defaultOrg(ctx, field)
			case "files":
				return ec.fieldContext_UserSetting_files(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type UserSetting", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _UserSettingCreatePayload_userSetting(ctx context.Context, field graphql.CollectedField, obj *model.UserSettingCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_UserSettingCreatePayload_userSetting(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (any, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.UserSetting, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(*generated.UserSetting)
	fc.Result = res
	return ec.marshalNUserSetting2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐUserSetting(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_UserSettingCreatePayload_userSetting(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "UserSettingCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_UserSetting_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_UserSetting_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_UserSetting_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_UserSetting_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_UserSetting_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_UserSetting_tags(ctx, field)
			case "userID":
				return ec.fieldContext_UserSetting_userID(ctx, field)
			case "locked":
				return ec.fieldContext_UserSetting_locked(ctx, field)
			case "silencedAt":
				return ec.fieldContext_UserSetting_silencedAt(ctx, field)
			case "suspendedAt":
				return ec.fieldContext_UserSetting_suspendedAt(ctx, field)
			case "status":
				return ec.fieldContext_UserSetting_status(ctx, field)
			case "emailConfirmed":
				return ec.fieldContext_UserSetting_emailConfirmed(ctx, field)
			case "isWebauthnAllowed":
				return ec.fieldContext_UserSetting_isWebauthnAllowed(ctx, field)
			case "isTfaEnabled":
				return ec.fieldContext_UserSetting_isTfaEnabled(ctx, field)
			case "user":
				return ec.fieldContext_UserSetting_user(ctx, field)
			case "defaultOrg":
				return ec.fieldContext_UserSetting_defaultOrg(ctx, field)
			case "files":
				return ec.fieldContext_UserSetting_files(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type UserSetting", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _UserSettingUpdatePayload_userSetting(ctx context.Context, field graphql.CollectedField, obj *model.UserSettingUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_UserSettingUpdatePayload_userSetting(ctx, field)
	if err != nil {
		return graphql.Null
	}
	ctx = graphql.WithFieldContext(ctx, fc)
	defer func() {
		if r := recover(); r != nil {
			ec.Error(ctx, ec.Recover(ctx, r))
			ret = graphql.Null
		}
	}()
	resTmp, err := ec.ResolverMiddleware(ctx, func(rctx context.Context) (any, error) {
		ctx = rctx // use context from middleware stack in children
		return obj.UserSetting, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		if !graphql.HasFieldError(ctx, fc) {
			ec.Errorf(ctx, "must not be null")
		}
		return graphql.Null
	}
	res := resTmp.(*generated.UserSetting)
	fc.Result = res
	return ec.marshalNUserSetting2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐUserSetting(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_UserSettingUpdatePayload_userSetting(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "UserSettingUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_UserSetting_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_UserSetting_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_UserSetting_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_UserSetting_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_UserSetting_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_UserSetting_tags(ctx, field)
			case "userID":
				return ec.fieldContext_UserSetting_userID(ctx, field)
			case "locked":
				return ec.fieldContext_UserSetting_locked(ctx, field)
			case "silencedAt":
				return ec.fieldContext_UserSetting_silencedAt(ctx, field)
			case "suspendedAt":
				return ec.fieldContext_UserSetting_suspendedAt(ctx, field)
			case "status":
				return ec.fieldContext_UserSetting_status(ctx, field)
			case "emailConfirmed":
				return ec.fieldContext_UserSetting_emailConfirmed(ctx, field)
			case "isWebauthnAllowed":
				return ec.fieldContext_UserSetting_isWebauthnAllowed(ctx, field)
			case "isTfaEnabled":
				return ec.fieldContext_UserSetting_isTfaEnabled(ctx, field)
			case "user":
				return ec.fieldContext_UserSetting_user(ctx, field)
			case "defaultOrg":
				return ec.fieldContext_UserSetting_defaultOrg(ctx, field)
			case "files":
				return ec.fieldContext_UserSetting_files(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type UserSetting", field.Name)
		},
	}
	return fc, nil
}

// endregion **************************** field.gotpl *****************************

// region    **************************** input.gotpl *****************************

// endregion **************************** input.gotpl *****************************

// region    ************************** interface.gotpl ***************************

// endregion ************************** interface.gotpl ***************************

// region    **************************** object.gotpl ****************************

var userSettingBulkCreatePayloadImplementors = []string{"UserSettingBulkCreatePayload"}

func (ec *executionContext) _UserSettingBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.UserSettingBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, userSettingBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("UserSettingBulkCreatePayload")
		case "userSettings":
			out.Values[i] = ec._UserSettingBulkCreatePayload_userSettings(ctx, field, obj)
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}
	out.Dispatch(ctx)
	if out.Invalids > 0 {
		return graphql.Null
	}

	atomic.AddInt32(&ec.deferred, int32(len(deferred)))

	for label, dfs := range deferred {
		ec.processDeferredGroup(graphql.DeferredGroup{
			Label:    label,
			Path:     graphql.GetPath(ctx),
			FieldSet: dfs,
			Context:  ctx,
		})
	}

	return out
}

var userSettingCreatePayloadImplementors = []string{"UserSettingCreatePayload"}

func (ec *executionContext) _UserSettingCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.UserSettingCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, userSettingCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("UserSettingCreatePayload")
		case "userSetting":
			out.Values[i] = ec._UserSettingCreatePayload_userSetting(ctx, field, obj)
			if out.Values[i] == graphql.Null {
				out.Invalids++
			}
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}
	out.Dispatch(ctx)
	if out.Invalids > 0 {
		return graphql.Null
	}

	atomic.AddInt32(&ec.deferred, int32(len(deferred)))

	for label, dfs := range deferred {
		ec.processDeferredGroup(graphql.DeferredGroup{
			Label:    label,
			Path:     graphql.GetPath(ctx),
			FieldSet: dfs,
			Context:  ctx,
		})
	}

	return out
}

var userSettingUpdatePayloadImplementors = []string{"UserSettingUpdatePayload"}

func (ec *executionContext) _UserSettingUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.UserSettingUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, userSettingUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("UserSettingUpdatePayload")
		case "userSetting":
			out.Values[i] = ec._UserSettingUpdatePayload_userSetting(ctx, field, obj)
			if out.Values[i] == graphql.Null {
				out.Invalids++
			}
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}
	out.Dispatch(ctx)
	if out.Invalids > 0 {
		return graphql.Null
	}

	atomic.AddInt32(&ec.deferred, int32(len(deferred)))

	for label, dfs := range deferred {
		ec.processDeferredGroup(graphql.DeferredGroup{
			Label:    label,
			Path:     graphql.GetPath(ctx),
			FieldSet: dfs,
			Context:  ctx,
		})
	}

	return out
}

// endregion **************************** object.gotpl ****************************

// region    ***************************** type.gotpl *****************************

func (ec *executionContext) marshalNUserSettingBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐUserSettingBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.UserSettingBulkCreatePayload) graphql.Marshaler {
	return ec._UserSettingBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNUserSettingBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐUserSettingBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.UserSettingBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._UserSettingBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNUserSettingCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐUserSettingCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.UserSettingCreatePayload) graphql.Marshaler {
	return ec._UserSettingCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNUserSettingCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐUserSettingCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.UserSettingCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._UserSettingCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNUserSettingUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐUserSettingUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.UserSettingUpdatePayload) graphql.Marshaler {
	return ec._UserSettingUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNUserSettingUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐUserSettingUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.UserSettingUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._UserSettingUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
