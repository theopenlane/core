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

func (ec *executionContext) _DocumentDataBulkCreatePayload_documentData(ctx context.Context, field graphql.CollectedField, obj *model.DocumentDataBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_DocumentDataBulkCreatePayload_documentData(ctx, field)
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
		return obj.DocumentData, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.DocumentData)
	fc.Result = res
	return ec.marshalODocumentData2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐDocumentDataᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_DocumentDataBulkCreatePayload_documentData(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "DocumentDataBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_DocumentData_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_DocumentData_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_DocumentData_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_DocumentData_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_DocumentData_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_DocumentData_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_DocumentData_ownerID(ctx, field)
			case "templateID":
				return ec.fieldContext_DocumentData_templateID(ctx, field)
			case "data":
				return ec.fieldContext_DocumentData_data(ctx, field)
			case "owner":
				return ec.fieldContext_DocumentData_owner(ctx, field)
			case "template":
				return ec.fieldContext_DocumentData_template(ctx, field)
			case "entities":
				return ec.fieldContext_DocumentData_entities(ctx, field)
			case "files":
				return ec.fieldContext_DocumentData_files(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type DocumentData", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _DocumentDataCreatePayload_documentData(ctx context.Context, field graphql.CollectedField, obj *model.DocumentDataCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_DocumentDataCreatePayload_documentData(ctx, field)
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
		return obj.DocumentData, nil
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
	res := resTmp.(*generated.DocumentData)
	fc.Result = res
	return ec.marshalNDocumentData2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐDocumentData(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_DocumentDataCreatePayload_documentData(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "DocumentDataCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_DocumentData_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_DocumentData_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_DocumentData_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_DocumentData_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_DocumentData_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_DocumentData_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_DocumentData_ownerID(ctx, field)
			case "templateID":
				return ec.fieldContext_DocumentData_templateID(ctx, field)
			case "data":
				return ec.fieldContext_DocumentData_data(ctx, field)
			case "owner":
				return ec.fieldContext_DocumentData_owner(ctx, field)
			case "template":
				return ec.fieldContext_DocumentData_template(ctx, field)
			case "entities":
				return ec.fieldContext_DocumentData_entities(ctx, field)
			case "files":
				return ec.fieldContext_DocumentData_files(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type DocumentData", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _DocumentDataDeletePayload_deletedID(ctx context.Context, field graphql.CollectedField, obj *model.DocumentDataDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_DocumentDataDeletePayload_deletedID(ctx, field)
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

func (ec *executionContext) fieldContext_DocumentDataDeletePayload_deletedID(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "DocumentDataDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _DocumentDataUpdatePayload_documentData(ctx context.Context, field graphql.CollectedField, obj *model.DocumentDataUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_DocumentDataUpdatePayload_documentData(ctx, field)
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
		return obj.DocumentData, nil
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
	res := resTmp.(*generated.DocumentData)
	fc.Result = res
	return ec.marshalNDocumentData2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐDocumentData(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_DocumentDataUpdatePayload_documentData(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "DocumentDataUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_DocumentData_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_DocumentData_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_DocumentData_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_DocumentData_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_DocumentData_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_DocumentData_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_DocumentData_ownerID(ctx, field)
			case "templateID":
				return ec.fieldContext_DocumentData_templateID(ctx, field)
			case "data":
				return ec.fieldContext_DocumentData_data(ctx, field)
			case "owner":
				return ec.fieldContext_DocumentData_owner(ctx, field)
			case "template":
				return ec.fieldContext_DocumentData_template(ctx, field)
			case "entities":
				return ec.fieldContext_DocumentData_entities(ctx, field)
			case "files":
				return ec.fieldContext_DocumentData_files(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type DocumentData", field.Name)
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

var documentDataBulkCreatePayloadImplementors = []string{"DocumentDataBulkCreatePayload"}

func (ec *executionContext) _DocumentDataBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.DocumentDataBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, documentDataBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("DocumentDataBulkCreatePayload")
		case "documentData":
			out.Values[i] = ec._DocumentDataBulkCreatePayload_documentData(ctx, field, obj)
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

var documentDataCreatePayloadImplementors = []string{"DocumentDataCreatePayload"}

func (ec *executionContext) _DocumentDataCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.DocumentDataCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, documentDataCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("DocumentDataCreatePayload")
		case "documentData":
			out.Values[i] = ec._DocumentDataCreatePayload_documentData(ctx, field, obj)
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

var documentDataDeletePayloadImplementors = []string{"DocumentDataDeletePayload"}

func (ec *executionContext) _DocumentDataDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.DocumentDataDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, documentDataDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("DocumentDataDeletePayload")
		case "deletedID":
			out.Values[i] = ec._DocumentDataDeletePayload_deletedID(ctx, field, obj)
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

var documentDataUpdatePayloadImplementors = []string{"DocumentDataUpdatePayload"}

func (ec *executionContext) _DocumentDataUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.DocumentDataUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, documentDataUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("DocumentDataUpdatePayload")
		case "documentData":
			out.Values[i] = ec._DocumentDataUpdatePayload_documentData(ctx, field, obj)
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

func (ec *executionContext) marshalNDocumentDataBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐDocumentDataBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.DocumentDataBulkCreatePayload) graphql.Marshaler {
	return ec._DocumentDataBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNDocumentDataBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐDocumentDataBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.DocumentDataBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._DocumentDataBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNDocumentDataCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐDocumentDataCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.DocumentDataCreatePayload) graphql.Marshaler {
	return ec._DocumentDataCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNDocumentDataCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐDocumentDataCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.DocumentDataCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._DocumentDataCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNDocumentDataDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐDocumentDataDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.DocumentDataDeletePayload) graphql.Marshaler {
	return ec._DocumentDataDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNDocumentDataDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐDocumentDataDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.DocumentDataDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._DocumentDataDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNDocumentDataUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐDocumentDataUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.DocumentDataUpdatePayload) graphql.Marshaler {
	return ec._DocumentDataUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNDocumentDataUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐDocumentDataUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.DocumentDataUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._DocumentDataUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
