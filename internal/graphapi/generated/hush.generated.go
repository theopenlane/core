// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlgenerated

import (
	"context"
	"errors"
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

func (ec *executionContext) _HushBulkCreatePayload_hushes(ctx context.Context, field graphql.CollectedField, obj *model.HushBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_HushBulkCreatePayload_hushes(ctx, field)
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
		return obj.Hushes, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.Hush)
	fc.Result = res
	return ec.marshalOHush2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐHushᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_HushBulkCreatePayload_hushes(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "HushBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Hush_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Hush_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Hush_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Hush_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Hush_updatedBy(ctx, field)
			case "ownerID":
				return ec.fieldContext_Hush_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_Hush_name(ctx, field)
			case "description":
				return ec.fieldContext_Hush_description(ctx, field)
			case "kind":
				return ec.fieldContext_Hush_kind(ctx, field)
			case "secretName":
				return ec.fieldContext_Hush_secretName(ctx, field)
			case "owner":
				return ec.fieldContext_Hush_owner(ctx, field)
			case "integrations":
				return ec.fieldContext_Hush_integrations(ctx, field)
			case "events":
				return ec.fieldContext_Hush_events(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Hush", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _HushCreatePayload_hush(ctx context.Context, field graphql.CollectedField, obj *model.HushCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_HushCreatePayload_hush(ctx, field)
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
		return obj.Hush, nil
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
	res := resTmp.(*generated.Hush)
	fc.Result = res
	return ec.marshalNHush2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐHush(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_HushCreatePayload_hush(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "HushCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Hush_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Hush_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Hush_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Hush_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Hush_updatedBy(ctx, field)
			case "ownerID":
				return ec.fieldContext_Hush_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_Hush_name(ctx, field)
			case "description":
				return ec.fieldContext_Hush_description(ctx, field)
			case "kind":
				return ec.fieldContext_Hush_kind(ctx, field)
			case "secretName":
				return ec.fieldContext_Hush_secretName(ctx, field)
			case "owner":
				return ec.fieldContext_Hush_owner(ctx, field)
			case "integrations":
				return ec.fieldContext_Hush_integrations(ctx, field)
			case "events":
				return ec.fieldContext_Hush_events(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Hush", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _HushDeletePayload_deletedID(ctx context.Context, field graphql.CollectedField, obj *model.HushDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_HushDeletePayload_deletedID(ctx, field)
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
		return obj.DeletedID, nil
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
	res := resTmp.(string)
	fc.Result = res
	return ec.marshalNID2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_HushDeletePayload_deletedID(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "HushDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _HushUpdatePayload_hush(ctx context.Context, field graphql.CollectedField, obj *model.HushUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_HushUpdatePayload_hush(ctx, field)
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
		return obj.Hush, nil
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
	res := resTmp.(*generated.Hush)
	fc.Result = res
	return ec.marshalNHush2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐHush(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_HushUpdatePayload_hush(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "HushUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Hush_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Hush_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Hush_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Hush_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Hush_updatedBy(ctx, field)
			case "ownerID":
				return ec.fieldContext_Hush_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_Hush_name(ctx, field)
			case "description":
				return ec.fieldContext_Hush_description(ctx, field)
			case "kind":
				return ec.fieldContext_Hush_kind(ctx, field)
			case "secretName":
				return ec.fieldContext_Hush_secretName(ctx, field)
			case "owner":
				return ec.fieldContext_Hush_owner(ctx, field)
			case "integrations":
				return ec.fieldContext_Hush_integrations(ctx, field)
			case "events":
				return ec.fieldContext_Hush_events(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Hush", field.Name)
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

var hushBulkCreatePayloadImplementors = []string{"HushBulkCreatePayload"}

func (ec *executionContext) _HushBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.HushBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, hushBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("HushBulkCreatePayload")
		case "hushes":
			out.Values[i] = ec._HushBulkCreatePayload_hushes(ctx, field, obj)
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

var hushCreatePayloadImplementors = []string{"HushCreatePayload"}

func (ec *executionContext) _HushCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.HushCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, hushCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("HushCreatePayload")
		case "hush":
			out.Values[i] = ec._HushCreatePayload_hush(ctx, field, obj)
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

var hushDeletePayloadImplementors = []string{"HushDeletePayload"}

func (ec *executionContext) _HushDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.HushDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, hushDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("HushDeletePayload")
		case "deletedID":
			out.Values[i] = ec._HushDeletePayload_deletedID(ctx, field, obj)
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

var hushUpdatePayloadImplementors = []string{"HushUpdatePayload"}

func (ec *executionContext) _HushUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.HushUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, hushUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("HushUpdatePayload")
		case "hush":
			out.Values[i] = ec._HushUpdatePayload_hush(ctx, field, obj)
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

func (ec *executionContext) marshalNHushBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐHushBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.HushBulkCreatePayload) graphql.Marshaler {
	return ec._HushBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNHushBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐHushBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.HushBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._HushBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNHushCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐHushCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.HushCreatePayload) graphql.Marshaler {
	return ec._HushCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNHushCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐHushCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.HushCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._HushCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNHushDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐHushDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.HushDeletePayload) graphql.Marshaler {
	return ec._HushDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNHushDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐHushDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.HushDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._HushDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNHushUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐHushUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.HushUpdatePayload) graphql.Marshaler {
	return ec._HushUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNHushUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐHushUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.HushUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._HushUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
