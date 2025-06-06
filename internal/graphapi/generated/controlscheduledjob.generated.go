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

func (ec *executionContext) _ControlScheduledJobBulkCreatePayload_controlScheduledJobs(ctx context.Context, field graphql.CollectedField, obj *model.ControlScheduledJobBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ControlScheduledJobBulkCreatePayload_controlScheduledJobs(ctx, field)
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
		return obj.ControlScheduledJobs, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.ControlScheduledJob)
	fc.Result = res
	return ec.marshalOControlScheduledJob2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐControlScheduledJobᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ControlScheduledJobBulkCreatePayload_controlScheduledJobs(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ControlScheduledJobBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_ControlScheduledJob_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_ControlScheduledJob_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_ControlScheduledJob_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_ControlScheduledJob_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_ControlScheduledJob_updatedBy(ctx, field)
			case "ownerID":
				return ec.fieldContext_ControlScheduledJob_ownerID(ctx, field)
			case "jobID":
				return ec.fieldContext_ControlScheduledJob_jobID(ctx, field)
			case "configuration":
				return ec.fieldContext_ControlScheduledJob_configuration(ctx, field)
			case "cadence":
				return ec.fieldContext_ControlScheduledJob_cadence(ctx, field)
			case "cron":
				return ec.fieldContext_ControlScheduledJob_cron(ctx, field)
			case "jobRunnerID":
				return ec.fieldContext_ControlScheduledJob_jobRunnerID(ctx, field)
			case "owner":
				return ec.fieldContext_ControlScheduledJob_owner(ctx, field)
			case "job":
				return ec.fieldContext_ControlScheduledJob_job(ctx, field)
			case "controls":
				return ec.fieldContext_ControlScheduledJob_controls(ctx, field)
			case "subcontrols":
				return ec.fieldContext_ControlScheduledJob_subcontrols(ctx, field)
			case "jobRunner":
				return ec.fieldContext_ControlScheduledJob_jobRunner(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type ControlScheduledJob", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _ControlScheduledJobCreatePayload_controlScheduledJob(ctx context.Context, field graphql.CollectedField, obj *model.ControlScheduledJobCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ControlScheduledJobCreatePayload_controlScheduledJob(ctx, field)
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
		return obj.ControlScheduledJob, nil
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
	res := resTmp.(*generated.ControlScheduledJob)
	fc.Result = res
	return ec.marshalNControlScheduledJob2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐControlScheduledJob(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ControlScheduledJobCreatePayload_controlScheduledJob(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ControlScheduledJobCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_ControlScheduledJob_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_ControlScheduledJob_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_ControlScheduledJob_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_ControlScheduledJob_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_ControlScheduledJob_updatedBy(ctx, field)
			case "ownerID":
				return ec.fieldContext_ControlScheduledJob_ownerID(ctx, field)
			case "jobID":
				return ec.fieldContext_ControlScheduledJob_jobID(ctx, field)
			case "configuration":
				return ec.fieldContext_ControlScheduledJob_configuration(ctx, field)
			case "cadence":
				return ec.fieldContext_ControlScheduledJob_cadence(ctx, field)
			case "cron":
				return ec.fieldContext_ControlScheduledJob_cron(ctx, field)
			case "jobRunnerID":
				return ec.fieldContext_ControlScheduledJob_jobRunnerID(ctx, field)
			case "owner":
				return ec.fieldContext_ControlScheduledJob_owner(ctx, field)
			case "job":
				return ec.fieldContext_ControlScheduledJob_job(ctx, field)
			case "controls":
				return ec.fieldContext_ControlScheduledJob_controls(ctx, field)
			case "subcontrols":
				return ec.fieldContext_ControlScheduledJob_subcontrols(ctx, field)
			case "jobRunner":
				return ec.fieldContext_ControlScheduledJob_jobRunner(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type ControlScheduledJob", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _ControlScheduledJobDeletePayload_deletedID(ctx context.Context, field graphql.CollectedField, obj *model.ControlScheduledJobDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ControlScheduledJobDeletePayload_deletedID(ctx, field)
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

func (ec *executionContext) fieldContext_ControlScheduledJobDeletePayload_deletedID(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ControlScheduledJobDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _ControlScheduledJobUpdatePayload_controlScheduledJob(ctx context.Context, field graphql.CollectedField, obj *model.ControlScheduledJobUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ControlScheduledJobUpdatePayload_controlScheduledJob(ctx, field)
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
		return obj.ControlScheduledJob, nil
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
	res := resTmp.(*generated.ControlScheduledJob)
	fc.Result = res
	return ec.marshalNControlScheduledJob2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐControlScheduledJob(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ControlScheduledJobUpdatePayload_controlScheduledJob(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ControlScheduledJobUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_ControlScheduledJob_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_ControlScheduledJob_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_ControlScheduledJob_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_ControlScheduledJob_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_ControlScheduledJob_updatedBy(ctx, field)
			case "ownerID":
				return ec.fieldContext_ControlScheduledJob_ownerID(ctx, field)
			case "jobID":
				return ec.fieldContext_ControlScheduledJob_jobID(ctx, field)
			case "configuration":
				return ec.fieldContext_ControlScheduledJob_configuration(ctx, field)
			case "cadence":
				return ec.fieldContext_ControlScheduledJob_cadence(ctx, field)
			case "cron":
				return ec.fieldContext_ControlScheduledJob_cron(ctx, field)
			case "jobRunnerID":
				return ec.fieldContext_ControlScheduledJob_jobRunnerID(ctx, field)
			case "owner":
				return ec.fieldContext_ControlScheduledJob_owner(ctx, field)
			case "job":
				return ec.fieldContext_ControlScheduledJob_job(ctx, field)
			case "controls":
				return ec.fieldContext_ControlScheduledJob_controls(ctx, field)
			case "subcontrols":
				return ec.fieldContext_ControlScheduledJob_subcontrols(ctx, field)
			case "jobRunner":
				return ec.fieldContext_ControlScheduledJob_jobRunner(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type ControlScheduledJob", field.Name)
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

var controlScheduledJobBulkCreatePayloadImplementors = []string{"ControlScheduledJobBulkCreatePayload"}

func (ec *executionContext) _ControlScheduledJobBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ControlScheduledJobBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, controlScheduledJobBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ControlScheduledJobBulkCreatePayload")
		case "controlScheduledJobs":
			out.Values[i] = ec._ControlScheduledJobBulkCreatePayload_controlScheduledJobs(ctx, field, obj)
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

var controlScheduledJobCreatePayloadImplementors = []string{"ControlScheduledJobCreatePayload"}

func (ec *executionContext) _ControlScheduledJobCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ControlScheduledJobCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, controlScheduledJobCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ControlScheduledJobCreatePayload")
		case "controlScheduledJob":
			out.Values[i] = ec._ControlScheduledJobCreatePayload_controlScheduledJob(ctx, field, obj)
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

var controlScheduledJobDeletePayloadImplementors = []string{"ControlScheduledJobDeletePayload"}

func (ec *executionContext) _ControlScheduledJobDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ControlScheduledJobDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, controlScheduledJobDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ControlScheduledJobDeletePayload")
		case "deletedID":
			out.Values[i] = ec._ControlScheduledJobDeletePayload_deletedID(ctx, field, obj)
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

var controlScheduledJobUpdatePayloadImplementors = []string{"ControlScheduledJobUpdatePayload"}

func (ec *executionContext) _ControlScheduledJobUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ControlScheduledJobUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, controlScheduledJobUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ControlScheduledJobUpdatePayload")
		case "controlScheduledJob":
			out.Values[i] = ec._ControlScheduledJobUpdatePayload_controlScheduledJob(ctx, field, obj)
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

func (ec *executionContext) marshalNControlScheduledJobBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐControlScheduledJobBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.ControlScheduledJobBulkCreatePayload) graphql.Marshaler {
	return ec._ControlScheduledJobBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNControlScheduledJobBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐControlScheduledJobBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.ControlScheduledJobBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ControlScheduledJobBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNControlScheduledJobCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐControlScheduledJobCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.ControlScheduledJobCreatePayload) graphql.Marshaler {
	return ec._ControlScheduledJobCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNControlScheduledJobCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐControlScheduledJobCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.ControlScheduledJobCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ControlScheduledJobCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNControlScheduledJobDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐControlScheduledJobDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.ControlScheduledJobDeletePayload) graphql.Marshaler {
	return ec._ControlScheduledJobDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNControlScheduledJobDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐControlScheduledJobDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.ControlScheduledJobDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ControlScheduledJobDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNControlScheduledJobUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐControlScheduledJobUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.ControlScheduledJobUpdatePayload) graphql.Marshaler {
	return ec._ControlScheduledJobUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNControlScheduledJobUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐControlScheduledJobUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.ControlScheduledJobUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ControlScheduledJobUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
