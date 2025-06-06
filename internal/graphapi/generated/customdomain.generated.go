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

func (ec *executionContext) _CustomDomainBulkCreatePayload_customDomains(ctx context.Context, field graphql.CollectedField, obj *model.CustomDomainBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_CustomDomainBulkCreatePayload_customDomains(ctx, field)
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
		return obj.CustomDomains, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.CustomDomain)
	fc.Result = res
	return ec.marshalOCustomDomain2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐCustomDomainᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_CustomDomainBulkCreatePayload_customDomains(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "CustomDomainBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_CustomDomain_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_CustomDomain_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_CustomDomain_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_CustomDomain_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_CustomDomain_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_CustomDomain_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_CustomDomain_ownerID(ctx, field)
			case "cnameRecord":
				return ec.fieldContext_CustomDomain_cnameRecord(ctx, field)
			case "mappableDomainID":
				return ec.fieldContext_CustomDomain_mappableDomainID(ctx, field)
			case "dnsVerificationID":
				return ec.fieldContext_CustomDomain_dnsVerificationID(ctx, field)
			case "owner":
				return ec.fieldContext_CustomDomain_owner(ctx, field)
			case "mappableDomain":
				return ec.fieldContext_CustomDomain_mappableDomain(ctx, field)
			case "dnsVerification":
				return ec.fieldContext_CustomDomain_dnsVerification(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type CustomDomain", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _CustomDomainCreatePayload_customDomain(ctx context.Context, field graphql.CollectedField, obj *model.CustomDomainCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_CustomDomainCreatePayload_customDomain(ctx, field)
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
		return obj.CustomDomain, nil
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
	res := resTmp.(*generated.CustomDomain)
	fc.Result = res
	return ec.marshalNCustomDomain2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐCustomDomain(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_CustomDomainCreatePayload_customDomain(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "CustomDomainCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_CustomDomain_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_CustomDomain_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_CustomDomain_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_CustomDomain_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_CustomDomain_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_CustomDomain_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_CustomDomain_ownerID(ctx, field)
			case "cnameRecord":
				return ec.fieldContext_CustomDomain_cnameRecord(ctx, field)
			case "mappableDomainID":
				return ec.fieldContext_CustomDomain_mappableDomainID(ctx, field)
			case "dnsVerificationID":
				return ec.fieldContext_CustomDomain_dnsVerificationID(ctx, field)
			case "owner":
				return ec.fieldContext_CustomDomain_owner(ctx, field)
			case "mappableDomain":
				return ec.fieldContext_CustomDomain_mappableDomain(ctx, field)
			case "dnsVerification":
				return ec.fieldContext_CustomDomain_dnsVerification(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type CustomDomain", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _CustomDomainDeletePayload_deletedID(ctx context.Context, field graphql.CollectedField, obj *model.CustomDomainDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_CustomDomainDeletePayload_deletedID(ctx, field)
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

func (ec *executionContext) fieldContext_CustomDomainDeletePayload_deletedID(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "CustomDomainDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _CustomDomainUpdatePayload_customDomain(ctx context.Context, field graphql.CollectedField, obj *model.CustomDomainUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_CustomDomainUpdatePayload_customDomain(ctx, field)
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
		return obj.CustomDomain, nil
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
	res := resTmp.(*generated.CustomDomain)
	fc.Result = res
	return ec.marshalNCustomDomain2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐCustomDomain(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_CustomDomainUpdatePayload_customDomain(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "CustomDomainUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_CustomDomain_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_CustomDomain_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_CustomDomain_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_CustomDomain_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_CustomDomain_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_CustomDomain_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_CustomDomain_ownerID(ctx, field)
			case "cnameRecord":
				return ec.fieldContext_CustomDomain_cnameRecord(ctx, field)
			case "mappableDomainID":
				return ec.fieldContext_CustomDomain_mappableDomainID(ctx, field)
			case "dnsVerificationID":
				return ec.fieldContext_CustomDomain_dnsVerificationID(ctx, field)
			case "owner":
				return ec.fieldContext_CustomDomain_owner(ctx, field)
			case "mappableDomain":
				return ec.fieldContext_CustomDomain_mappableDomain(ctx, field)
			case "dnsVerification":
				return ec.fieldContext_CustomDomain_dnsVerification(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type CustomDomain", field.Name)
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

var customDomainBulkCreatePayloadImplementors = []string{"CustomDomainBulkCreatePayload"}

func (ec *executionContext) _CustomDomainBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.CustomDomainBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, customDomainBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("CustomDomainBulkCreatePayload")
		case "customDomains":
			out.Values[i] = ec._CustomDomainBulkCreatePayload_customDomains(ctx, field, obj)
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

var customDomainCreatePayloadImplementors = []string{"CustomDomainCreatePayload"}

func (ec *executionContext) _CustomDomainCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.CustomDomainCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, customDomainCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("CustomDomainCreatePayload")
		case "customDomain":
			out.Values[i] = ec._CustomDomainCreatePayload_customDomain(ctx, field, obj)
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

var customDomainDeletePayloadImplementors = []string{"CustomDomainDeletePayload"}

func (ec *executionContext) _CustomDomainDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.CustomDomainDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, customDomainDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("CustomDomainDeletePayload")
		case "deletedID":
			out.Values[i] = ec._CustomDomainDeletePayload_deletedID(ctx, field, obj)
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

var customDomainUpdatePayloadImplementors = []string{"CustomDomainUpdatePayload"}

func (ec *executionContext) _CustomDomainUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.CustomDomainUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, customDomainUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("CustomDomainUpdatePayload")
		case "customDomain":
			out.Values[i] = ec._CustomDomainUpdatePayload_customDomain(ctx, field, obj)
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

func (ec *executionContext) marshalNCustomDomainBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐCustomDomainBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.CustomDomainBulkCreatePayload) graphql.Marshaler {
	return ec._CustomDomainBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNCustomDomainBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐCustomDomainBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.CustomDomainBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._CustomDomainBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNCustomDomainCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐCustomDomainCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.CustomDomainCreatePayload) graphql.Marshaler {
	return ec._CustomDomainCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNCustomDomainCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐCustomDomainCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.CustomDomainCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._CustomDomainCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNCustomDomainDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐCustomDomainDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.CustomDomainDeletePayload) graphql.Marshaler {
	return ec._CustomDomainDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNCustomDomainDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐCustomDomainDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.CustomDomainDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._CustomDomainDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNCustomDomainUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐCustomDomainUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.CustomDomainUpdatePayload) graphql.Marshaler {
	return ec._CustomDomainUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNCustomDomainUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐCustomDomainUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.CustomDomainUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._CustomDomainUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
