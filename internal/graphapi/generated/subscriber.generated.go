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

func (ec *executionContext) _SubscriberBulkCreatePayload_subscribers(ctx context.Context, field graphql.CollectedField, obj *model.SubscriberBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_SubscriberBulkCreatePayload_subscribers(ctx, field)
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
		return obj.Subscribers, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.Subscriber)
	fc.Result = res
	return ec.marshalOSubscriber2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐSubscriberᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_SubscriberBulkCreatePayload_subscribers(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "SubscriberBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Subscriber_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Subscriber_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Subscriber_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Subscriber_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Subscriber_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_Subscriber_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_Subscriber_ownerID(ctx, field)
			case "email":
				return ec.fieldContext_Subscriber_email(ctx, field)
			case "phoneNumber":
				return ec.fieldContext_Subscriber_phoneNumber(ctx, field)
			case "verifiedEmail":
				return ec.fieldContext_Subscriber_verifiedEmail(ctx, field)
			case "verifiedPhone":
				return ec.fieldContext_Subscriber_verifiedPhone(ctx, field)
			case "active":
				return ec.fieldContext_Subscriber_active(ctx, field)
			case "unsubscribed":
				return ec.fieldContext_Subscriber_unsubscribed(ctx, field)
			case "sendAttempts":
				return ec.fieldContext_Subscriber_sendAttempts(ctx, field)
			case "owner":
				return ec.fieldContext_Subscriber_owner(ctx, field)
			case "events":
				return ec.fieldContext_Subscriber_events(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Subscriber", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _SubscriberCreatePayload_subscriber(ctx context.Context, field graphql.CollectedField, obj *model.SubscriberCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_SubscriberCreatePayload_subscriber(ctx, field)
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
		return obj.Subscriber, nil
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
	res := resTmp.(*generated.Subscriber)
	fc.Result = res
	return ec.marshalNSubscriber2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐSubscriber(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_SubscriberCreatePayload_subscriber(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "SubscriberCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Subscriber_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Subscriber_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Subscriber_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Subscriber_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Subscriber_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_Subscriber_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_Subscriber_ownerID(ctx, field)
			case "email":
				return ec.fieldContext_Subscriber_email(ctx, field)
			case "phoneNumber":
				return ec.fieldContext_Subscriber_phoneNumber(ctx, field)
			case "verifiedEmail":
				return ec.fieldContext_Subscriber_verifiedEmail(ctx, field)
			case "verifiedPhone":
				return ec.fieldContext_Subscriber_verifiedPhone(ctx, field)
			case "active":
				return ec.fieldContext_Subscriber_active(ctx, field)
			case "unsubscribed":
				return ec.fieldContext_Subscriber_unsubscribed(ctx, field)
			case "sendAttempts":
				return ec.fieldContext_Subscriber_sendAttempts(ctx, field)
			case "owner":
				return ec.fieldContext_Subscriber_owner(ctx, field)
			case "events":
				return ec.fieldContext_Subscriber_events(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Subscriber", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _SubscriberDeletePayload_email(ctx context.Context, field graphql.CollectedField, obj *model.SubscriberDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_SubscriberDeletePayload_email(ctx, field)
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
		return obj.Email, nil
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
	return ec.marshalNString2string(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_SubscriberDeletePayload_email(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "SubscriberDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type String does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _SubscriberUpdatePayload_subscriber(ctx context.Context, field graphql.CollectedField, obj *model.SubscriberUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_SubscriberUpdatePayload_subscriber(ctx, field)
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
		return obj.Subscriber, nil
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
	res := resTmp.(*generated.Subscriber)
	fc.Result = res
	return ec.marshalNSubscriber2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐSubscriber(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_SubscriberUpdatePayload_subscriber(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "SubscriberUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Subscriber_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Subscriber_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Subscriber_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Subscriber_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Subscriber_updatedBy(ctx, field)
			case "tags":
				return ec.fieldContext_Subscriber_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_Subscriber_ownerID(ctx, field)
			case "email":
				return ec.fieldContext_Subscriber_email(ctx, field)
			case "phoneNumber":
				return ec.fieldContext_Subscriber_phoneNumber(ctx, field)
			case "verifiedEmail":
				return ec.fieldContext_Subscriber_verifiedEmail(ctx, field)
			case "verifiedPhone":
				return ec.fieldContext_Subscriber_verifiedPhone(ctx, field)
			case "active":
				return ec.fieldContext_Subscriber_active(ctx, field)
			case "unsubscribed":
				return ec.fieldContext_Subscriber_unsubscribed(ctx, field)
			case "sendAttempts":
				return ec.fieldContext_Subscriber_sendAttempts(ctx, field)
			case "owner":
				return ec.fieldContext_Subscriber_owner(ctx, field)
			case "events":
				return ec.fieldContext_Subscriber_events(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Subscriber", field.Name)
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

var subscriberBulkCreatePayloadImplementors = []string{"SubscriberBulkCreatePayload"}

func (ec *executionContext) _SubscriberBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.SubscriberBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, subscriberBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("SubscriberBulkCreatePayload")
		case "subscribers":
			out.Values[i] = ec._SubscriberBulkCreatePayload_subscribers(ctx, field, obj)
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

var subscriberCreatePayloadImplementors = []string{"SubscriberCreatePayload"}

func (ec *executionContext) _SubscriberCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.SubscriberCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, subscriberCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("SubscriberCreatePayload")
		case "subscriber":
			out.Values[i] = ec._SubscriberCreatePayload_subscriber(ctx, field, obj)
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

var subscriberDeletePayloadImplementors = []string{"SubscriberDeletePayload"}

func (ec *executionContext) _SubscriberDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.SubscriberDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, subscriberDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("SubscriberDeletePayload")
		case "email":
			out.Values[i] = ec._SubscriberDeletePayload_email(ctx, field, obj)
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

var subscriberUpdatePayloadImplementors = []string{"SubscriberUpdatePayload"}

func (ec *executionContext) _SubscriberUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.SubscriberUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, subscriberUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("SubscriberUpdatePayload")
		case "subscriber":
			out.Values[i] = ec._SubscriberUpdatePayload_subscriber(ctx, field, obj)
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

func (ec *executionContext) marshalNSubscriberBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐSubscriberBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.SubscriberBulkCreatePayload) graphql.Marshaler {
	return ec._SubscriberBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNSubscriberBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐSubscriberBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.SubscriberBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._SubscriberBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNSubscriberCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐSubscriberCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.SubscriberCreatePayload) graphql.Marshaler {
	return ec._SubscriberCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNSubscriberCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐSubscriberCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.SubscriberCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._SubscriberCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNSubscriberDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐSubscriberDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.SubscriberDeletePayload) graphql.Marshaler {
	return ec._SubscriberDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNSubscriberDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐSubscriberDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.SubscriberDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._SubscriberDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNSubscriberUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐSubscriberUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.SubscriberUpdatePayload) graphql.Marshaler {
	return ec._SubscriberUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNSubscriberUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐSubscriberUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.SubscriberUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._SubscriberUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
