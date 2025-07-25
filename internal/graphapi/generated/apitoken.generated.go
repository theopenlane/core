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

func (ec *executionContext) _APITokenBulkCreatePayload_apiTokens(ctx context.Context, field graphql.CollectedField, obj *model.APITokenBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_APITokenBulkCreatePayload_apiTokens(ctx, field)
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
		return obj.APITokens, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.APIToken)
	fc.Result = res
	return ec.marshalOAPIToken2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐAPITokenᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_APITokenBulkCreatePayload_apiTokens(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "APITokenBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_APIToken_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_APIToken_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_APIToken_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_APIToken_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_APIToken_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_APIToken_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_APIToken_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_APIToken_name(ctx, field)
			case "token":
				return ec.fieldContext_APIToken_token(ctx, field)
			case "expiresAt":
				return ec.fieldContext_APIToken_expiresAt(ctx, field)
			case "description":
				return ec.fieldContext_APIToken_description(ctx, field)
			case "scopes":
				return ec.fieldContext_APIToken_scopes(ctx, field)
			case "lastUsedAt":
				return ec.fieldContext_APIToken_lastUsedAt(ctx, field)
			case "isActive":
				return ec.fieldContext_APIToken_isActive(ctx, field)
			case "revokedReason":
				return ec.fieldContext_APIToken_revokedReason(ctx, field)
			case "revokedBy":
				return ec.fieldContext_APIToken_revokedBy(ctx, field)
			case "revokedAt":
				return ec.fieldContext_APIToken_revokedAt(ctx, field)
			case "ssoAuthorizations":
				return ec.fieldContext_APIToken_ssoAuthorizations(ctx, field)
			case "owner":
				return ec.fieldContext_APIToken_owner(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type APIToken", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _APITokenCreatePayload_apiToken(ctx context.Context, field graphql.CollectedField, obj *model.APITokenCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_APITokenCreatePayload_apiToken(ctx, field)
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
		return obj.APIToken, nil
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
	res := resTmp.(*generated.APIToken)
	fc.Result = res
	return ec.marshalNAPIToken2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐAPIToken(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_APITokenCreatePayload_apiToken(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "APITokenCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_APIToken_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_APIToken_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_APIToken_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_APIToken_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_APIToken_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_APIToken_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_APIToken_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_APIToken_name(ctx, field)
			case "token":
				return ec.fieldContext_APIToken_token(ctx, field)
			case "expiresAt":
				return ec.fieldContext_APIToken_expiresAt(ctx, field)
			case "description":
				return ec.fieldContext_APIToken_description(ctx, field)
			case "scopes":
				return ec.fieldContext_APIToken_scopes(ctx, field)
			case "lastUsedAt":
				return ec.fieldContext_APIToken_lastUsedAt(ctx, field)
			case "isActive":
				return ec.fieldContext_APIToken_isActive(ctx, field)
			case "revokedReason":
				return ec.fieldContext_APIToken_revokedReason(ctx, field)
			case "revokedBy":
				return ec.fieldContext_APIToken_revokedBy(ctx, field)
			case "revokedAt":
				return ec.fieldContext_APIToken_revokedAt(ctx, field)
			case "ssoAuthorizations":
				return ec.fieldContext_APIToken_ssoAuthorizations(ctx, field)
			case "owner":
				return ec.fieldContext_APIToken_owner(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type APIToken", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _APITokenDeletePayload_deletedID(ctx context.Context, field graphql.CollectedField, obj *model.APITokenDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_APITokenDeletePayload_deletedID(ctx, field)
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

func (ec *executionContext) fieldContext_APITokenDeletePayload_deletedID(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "APITokenDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _APITokenUpdatePayload_apiToken(ctx context.Context, field graphql.CollectedField, obj *model.APITokenUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_APITokenUpdatePayload_apiToken(ctx, field)
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
		return obj.APIToken, nil
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
	res := resTmp.(*generated.APIToken)
	fc.Result = res
	return ec.marshalNAPIToken2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐAPIToken(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_APITokenUpdatePayload_apiToken(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "APITokenUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_APIToken_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_APIToken_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_APIToken_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_APIToken_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_APIToken_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_APIToken_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_APIToken_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_APIToken_name(ctx, field)
			case "token":
				return ec.fieldContext_APIToken_token(ctx, field)
			case "expiresAt":
				return ec.fieldContext_APIToken_expiresAt(ctx, field)
			case "description":
				return ec.fieldContext_APIToken_description(ctx, field)
			case "scopes":
				return ec.fieldContext_APIToken_scopes(ctx, field)
			case "lastUsedAt":
				return ec.fieldContext_APIToken_lastUsedAt(ctx, field)
			case "isActive":
				return ec.fieldContext_APIToken_isActive(ctx, field)
			case "revokedReason":
				return ec.fieldContext_APIToken_revokedReason(ctx, field)
			case "revokedBy":
				return ec.fieldContext_APIToken_revokedBy(ctx, field)
			case "revokedAt":
				return ec.fieldContext_APIToken_revokedAt(ctx, field)
			case "ssoAuthorizations":
				return ec.fieldContext_APIToken_ssoAuthorizations(ctx, field)
			case "owner":
				return ec.fieldContext_APIToken_owner(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type APIToken", field.Name)
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

var aPITokenBulkCreatePayloadImplementors = []string{"APITokenBulkCreatePayload"}

func (ec *executionContext) _APITokenBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.APITokenBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, aPITokenBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("APITokenBulkCreatePayload")
		case "apiTokens":
			out.Values[i] = ec._APITokenBulkCreatePayload_apiTokens(ctx, field, obj)
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

var aPITokenCreatePayloadImplementors = []string{"APITokenCreatePayload"}

func (ec *executionContext) _APITokenCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.APITokenCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, aPITokenCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("APITokenCreatePayload")
		case "apiToken":
			out.Values[i] = ec._APITokenCreatePayload_apiToken(ctx, field, obj)
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

var aPITokenDeletePayloadImplementors = []string{"APITokenDeletePayload"}

func (ec *executionContext) _APITokenDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.APITokenDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, aPITokenDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("APITokenDeletePayload")
		case "deletedID":
			out.Values[i] = ec._APITokenDeletePayload_deletedID(ctx, field, obj)
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

var aPITokenUpdatePayloadImplementors = []string{"APITokenUpdatePayload"}

func (ec *executionContext) _APITokenUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.APITokenUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, aPITokenUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("APITokenUpdatePayload")
		case "apiToken":
			out.Values[i] = ec._APITokenUpdatePayload_apiToken(ctx, field, obj)
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

func (ec *executionContext) marshalNAPITokenBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐAPITokenBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.APITokenBulkCreatePayload) graphql.Marshaler {
	return ec._APITokenBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNAPITokenBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐAPITokenBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.APITokenBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._APITokenBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNAPITokenCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐAPITokenCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.APITokenCreatePayload) graphql.Marshaler {
	return ec._APITokenCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNAPITokenCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐAPITokenCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.APITokenCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._APITokenCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNAPITokenDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐAPITokenDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.APITokenDeletePayload) graphql.Marshaler {
	return ec._APITokenDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNAPITokenDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐAPITokenDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.APITokenDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._APITokenDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNAPITokenUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐAPITokenUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.APITokenUpdatePayload) graphql.Marshaler {
	return ec._APITokenUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNAPITokenUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐAPITokenUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.APITokenUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._APITokenUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
