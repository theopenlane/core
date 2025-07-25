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

func (ec *executionContext) _TrustCenterSubprocessorBulkCreatePayload_trustCenterSubprocessors(ctx context.Context, field graphql.CollectedField, obj *model.TrustCenterSubprocessorBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_TrustCenterSubprocessorBulkCreatePayload_trustCenterSubprocessors(ctx, field)
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
		return obj.TrustCenterSubprocessors, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.TrustCenterSubprocessor)
	fc.Result = res
	return ec.marshalOTrustCenterSubprocessor2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐTrustCenterSubprocessorᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_TrustCenterSubprocessorBulkCreatePayload_trustCenterSubprocessors(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "TrustCenterSubprocessorBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_TrustCenterSubprocessor_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_TrustCenterSubprocessor_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_TrustCenterSubprocessor_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_TrustCenterSubprocessor_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_TrustCenterSubprocessor_updatedBy(ctx, field)
			case "subprocessorID":
				return ec.fieldContext_TrustCenterSubprocessor_subprocessorID(ctx, field)
			case "trustCenterID":
				return ec.fieldContext_TrustCenterSubprocessor_trustCenterID(ctx, field)
			case "countries":
				return ec.fieldContext_TrustCenterSubprocessor_countries(ctx, field)
			case "category":
				return ec.fieldContext_TrustCenterSubprocessor_category(ctx, field)
			case "trustCenter":
				return ec.fieldContext_TrustCenterSubprocessor_trustCenter(ctx, field)
			case "subprocessor":
				return ec.fieldContext_TrustCenterSubprocessor_subprocessor(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type TrustCenterSubprocessor", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _TrustCenterSubprocessorCreatePayload_trustCenterSubprocessor(ctx context.Context, field graphql.CollectedField, obj *model.TrustCenterSubprocessorCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_TrustCenterSubprocessorCreatePayload_trustCenterSubprocessor(ctx, field)
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
		return obj.TrustCenterSubprocessor, nil
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
	res := resTmp.(*generated.TrustCenterSubprocessor)
	fc.Result = res
	return ec.marshalNTrustCenterSubprocessor2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐTrustCenterSubprocessor(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_TrustCenterSubprocessorCreatePayload_trustCenterSubprocessor(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "TrustCenterSubprocessorCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_TrustCenterSubprocessor_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_TrustCenterSubprocessor_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_TrustCenterSubprocessor_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_TrustCenterSubprocessor_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_TrustCenterSubprocessor_updatedBy(ctx, field)
			case "subprocessorID":
				return ec.fieldContext_TrustCenterSubprocessor_subprocessorID(ctx, field)
			case "trustCenterID":
				return ec.fieldContext_TrustCenterSubprocessor_trustCenterID(ctx, field)
			case "countries":
				return ec.fieldContext_TrustCenterSubprocessor_countries(ctx, field)
			case "category":
				return ec.fieldContext_TrustCenterSubprocessor_category(ctx, field)
			case "trustCenter":
				return ec.fieldContext_TrustCenterSubprocessor_trustCenter(ctx, field)
			case "subprocessor":
				return ec.fieldContext_TrustCenterSubprocessor_subprocessor(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type TrustCenterSubprocessor", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _TrustCenterSubprocessorDeletePayload_deletedID(ctx context.Context, field graphql.CollectedField, obj *model.TrustCenterSubprocessorDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_TrustCenterSubprocessorDeletePayload_deletedID(ctx, field)
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

func (ec *executionContext) fieldContext_TrustCenterSubprocessorDeletePayload_deletedID(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "TrustCenterSubprocessorDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _TrustCenterSubprocessorUpdatePayload_trustCenterSubprocessor(ctx context.Context, field graphql.CollectedField, obj *model.TrustCenterSubprocessorUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_TrustCenterSubprocessorUpdatePayload_trustCenterSubprocessor(ctx, field)
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
		return obj.TrustCenterSubprocessor, nil
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
	res := resTmp.(*generated.TrustCenterSubprocessor)
	fc.Result = res
	return ec.marshalNTrustCenterSubprocessor2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐTrustCenterSubprocessor(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_TrustCenterSubprocessorUpdatePayload_trustCenterSubprocessor(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "TrustCenterSubprocessorUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_TrustCenterSubprocessor_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_TrustCenterSubprocessor_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_TrustCenterSubprocessor_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_TrustCenterSubprocessor_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_TrustCenterSubprocessor_updatedBy(ctx, field)
			case "subprocessorID":
				return ec.fieldContext_TrustCenterSubprocessor_subprocessorID(ctx, field)
			case "trustCenterID":
				return ec.fieldContext_TrustCenterSubprocessor_trustCenterID(ctx, field)
			case "countries":
				return ec.fieldContext_TrustCenterSubprocessor_countries(ctx, field)
			case "category":
				return ec.fieldContext_TrustCenterSubprocessor_category(ctx, field)
			case "trustCenter":
				return ec.fieldContext_TrustCenterSubprocessor_trustCenter(ctx, field)
			case "subprocessor":
				return ec.fieldContext_TrustCenterSubprocessor_subprocessor(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type TrustCenterSubprocessor", field.Name)
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

var trustCenterSubprocessorBulkCreatePayloadImplementors = []string{"TrustCenterSubprocessorBulkCreatePayload"}

func (ec *executionContext) _TrustCenterSubprocessorBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.TrustCenterSubprocessorBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, trustCenterSubprocessorBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("TrustCenterSubprocessorBulkCreatePayload")
		case "trustCenterSubprocessors":
			out.Values[i] = ec._TrustCenterSubprocessorBulkCreatePayload_trustCenterSubprocessors(ctx, field, obj)
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

var trustCenterSubprocessorCreatePayloadImplementors = []string{"TrustCenterSubprocessorCreatePayload"}

func (ec *executionContext) _TrustCenterSubprocessorCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.TrustCenterSubprocessorCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, trustCenterSubprocessorCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("TrustCenterSubprocessorCreatePayload")
		case "trustCenterSubprocessor":
			out.Values[i] = ec._TrustCenterSubprocessorCreatePayload_trustCenterSubprocessor(ctx, field, obj)
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

var trustCenterSubprocessorDeletePayloadImplementors = []string{"TrustCenterSubprocessorDeletePayload"}

func (ec *executionContext) _TrustCenterSubprocessorDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.TrustCenterSubprocessorDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, trustCenterSubprocessorDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("TrustCenterSubprocessorDeletePayload")
		case "deletedID":
			out.Values[i] = ec._TrustCenterSubprocessorDeletePayload_deletedID(ctx, field, obj)
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

var trustCenterSubprocessorUpdatePayloadImplementors = []string{"TrustCenterSubprocessorUpdatePayload"}

func (ec *executionContext) _TrustCenterSubprocessorUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.TrustCenterSubprocessorUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, trustCenterSubprocessorUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("TrustCenterSubprocessorUpdatePayload")
		case "trustCenterSubprocessor":
			out.Values[i] = ec._TrustCenterSubprocessorUpdatePayload_trustCenterSubprocessor(ctx, field, obj)
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

func (ec *executionContext) marshalNTrustCenterSubprocessorBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐTrustCenterSubprocessorBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.TrustCenterSubprocessorBulkCreatePayload) graphql.Marshaler {
	return ec._TrustCenterSubprocessorBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNTrustCenterSubprocessorBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐTrustCenterSubprocessorBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.TrustCenterSubprocessorBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._TrustCenterSubprocessorBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNTrustCenterSubprocessorCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐTrustCenterSubprocessorCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.TrustCenterSubprocessorCreatePayload) graphql.Marshaler {
	return ec._TrustCenterSubprocessorCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNTrustCenterSubprocessorCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐTrustCenterSubprocessorCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.TrustCenterSubprocessorCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._TrustCenterSubprocessorCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNTrustCenterSubprocessorDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐTrustCenterSubprocessorDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.TrustCenterSubprocessorDeletePayload) graphql.Marshaler {
	return ec._TrustCenterSubprocessorDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNTrustCenterSubprocessorDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐTrustCenterSubprocessorDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.TrustCenterSubprocessorDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._TrustCenterSubprocessorDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNTrustCenterSubprocessorUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐTrustCenterSubprocessorUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.TrustCenterSubprocessorUpdatePayload) graphql.Marshaler {
	return ec._TrustCenterSubprocessorUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNTrustCenterSubprocessorUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐTrustCenterSubprocessorUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.TrustCenterSubprocessorUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._TrustCenterSubprocessorUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
