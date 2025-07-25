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

func (ec *executionContext) _ExportBulkCreatePayload_exports(ctx context.Context, field graphql.CollectedField, obj *model.ExportBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ExportBulkCreatePayload_exports(ctx, field)
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
		return obj.Exports, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.Export)
	fc.Result = res
	return ec.marshalOExport2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐExportᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ExportBulkCreatePayload_exports(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ExportBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Export_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Export_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Export_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Export_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Export_updatedBy(ctx, field)
			case "ownerID":
				return ec.fieldContext_Export_ownerID(ctx, field)
			case "exportType":
				return ec.fieldContext_Export_exportType(ctx, field)
			case "format":
				return ec.fieldContext_Export_format(ctx, field)
			case "status":
				return ec.fieldContext_Export_status(ctx, field)
			case "requestorID":
				return ec.fieldContext_Export_requestorID(ctx, field)
			case "fields":
				return ec.fieldContext_Export_fields(ctx, field)
			case "filters":
				return ec.fieldContext_Export_filters(ctx, field)
			case "errorMessage":
				return ec.fieldContext_Export_errorMessage(ctx, field)
			case "owner":
				return ec.fieldContext_Export_owner(ctx, field)
			case "events":
				return ec.fieldContext_Export_events(ctx, field)
			case "files":
				return ec.fieldContext_Export_files(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Export", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _ExportBulkDeletePayload_deletedIDs(ctx context.Context, field graphql.CollectedField, obj *model.ExportBulkDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ExportBulkDeletePayload_deletedIDs(ctx, field)
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
		return obj.DeletedIDs, nil
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
	res := resTmp.([]string)
	fc.Result = res
	return ec.marshalNID2ᚕstringᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ExportBulkDeletePayload_deletedIDs(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ExportBulkDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _ExportCreatePayload_export(ctx context.Context, field graphql.CollectedField, obj *model.ExportCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ExportCreatePayload_export(ctx, field)
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
		return obj.Export, nil
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
	res := resTmp.(*generated.Export)
	fc.Result = res
	return ec.marshalNExport2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐExport(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ExportCreatePayload_export(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ExportCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Export_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Export_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Export_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Export_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Export_updatedBy(ctx, field)
			case "ownerID":
				return ec.fieldContext_Export_ownerID(ctx, field)
			case "exportType":
				return ec.fieldContext_Export_exportType(ctx, field)
			case "format":
				return ec.fieldContext_Export_format(ctx, field)
			case "status":
				return ec.fieldContext_Export_status(ctx, field)
			case "requestorID":
				return ec.fieldContext_Export_requestorID(ctx, field)
			case "fields":
				return ec.fieldContext_Export_fields(ctx, field)
			case "filters":
				return ec.fieldContext_Export_filters(ctx, field)
			case "errorMessage":
				return ec.fieldContext_Export_errorMessage(ctx, field)
			case "owner":
				return ec.fieldContext_Export_owner(ctx, field)
			case "events":
				return ec.fieldContext_Export_events(ctx, field)
			case "files":
				return ec.fieldContext_Export_files(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Export", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _ExportDeletePayload_deletedID(ctx context.Context, field graphql.CollectedField, obj *model.ExportDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ExportDeletePayload_deletedID(ctx, field)
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

func (ec *executionContext) fieldContext_ExportDeletePayload_deletedID(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ExportDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _ExportUpdatePayload_export(ctx context.Context, field graphql.CollectedField, obj *model.ExportUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ExportUpdatePayload_export(ctx, field)
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
		return obj.Export, nil
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
	res := resTmp.(*generated.Export)
	fc.Result = res
	return ec.marshalNExport2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐExport(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ExportUpdatePayload_export(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ExportUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Export_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Export_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Export_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Export_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Export_updatedBy(ctx, field)
			case "ownerID":
				return ec.fieldContext_Export_ownerID(ctx, field)
			case "exportType":
				return ec.fieldContext_Export_exportType(ctx, field)
			case "format":
				return ec.fieldContext_Export_format(ctx, field)
			case "status":
				return ec.fieldContext_Export_status(ctx, field)
			case "requestorID":
				return ec.fieldContext_Export_requestorID(ctx, field)
			case "fields":
				return ec.fieldContext_Export_fields(ctx, field)
			case "filters":
				return ec.fieldContext_Export_filters(ctx, field)
			case "errorMessage":
				return ec.fieldContext_Export_errorMessage(ctx, field)
			case "owner":
				return ec.fieldContext_Export_owner(ctx, field)
			case "events":
				return ec.fieldContext_Export_events(ctx, field)
			case "files":
				return ec.fieldContext_Export_files(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Export", field.Name)
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

var exportBulkCreatePayloadImplementors = []string{"ExportBulkCreatePayload"}

func (ec *executionContext) _ExportBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ExportBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, exportBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ExportBulkCreatePayload")
		case "exports":
			out.Values[i] = ec._ExportBulkCreatePayload_exports(ctx, field, obj)
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

var exportBulkDeletePayloadImplementors = []string{"ExportBulkDeletePayload"}

func (ec *executionContext) _ExportBulkDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ExportBulkDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, exportBulkDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ExportBulkDeletePayload")
		case "deletedIDs":
			out.Values[i] = ec._ExportBulkDeletePayload_deletedIDs(ctx, field, obj)
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

var exportCreatePayloadImplementors = []string{"ExportCreatePayload"}

func (ec *executionContext) _ExportCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ExportCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, exportCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ExportCreatePayload")
		case "export":
			out.Values[i] = ec._ExportCreatePayload_export(ctx, field, obj)
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

var exportDeletePayloadImplementors = []string{"ExportDeletePayload"}

func (ec *executionContext) _ExportDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ExportDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, exportDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ExportDeletePayload")
		case "deletedID":
			out.Values[i] = ec._ExportDeletePayload_deletedID(ctx, field, obj)
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

var exportUpdatePayloadImplementors = []string{"ExportUpdatePayload"}

func (ec *executionContext) _ExportUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ExportUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, exportUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ExportUpdatePayload")
		case "export":
			out.Values[i] = ec._ExportUpdatePayload_export(ctx, field, obj)
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

func (ec *executionContext) marshalNExportBulkDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐExportBulkDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.ExportBulkDeletePayload) graphql.Marshaler {
	return ec._ExportBulkDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNExportBulkDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐExportBulkDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.ExportBulkDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ExportBulkDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNExportCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐExportCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.ExportCreatePayload) graphql.Marshaler {
	return ec._ExportCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNExportCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐExportCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.ExportCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ExportCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNExportDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐExportDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.ExportDeletePayload) graphql.Marshaler {
	return ec._ExportDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNExportDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐExportDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.ExportDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ExportDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNExportUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐExportUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.ExportUpdatePayload) graphql.Marshaler {
	return ec._ExportUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNExportUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐExportUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.ExportUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ExportUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
