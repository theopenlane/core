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

func (ec *executionContext) _ProgramBulkCreatePayload_programs(ctx context.Context, field graphql.CollectedField, obj *model.ProgramBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ProgramBulkCreatePayload_programs(ctx, field)
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
		return obj.Programs, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.Program)
	fc.Result = res
	return ec.marshalOProgram2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐProgramᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ProgramBulkCreatePayload_programs(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ProgramBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Program_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Program_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Program_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Program_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Program_updatedBy(ctx, field)
			case "displayID":
				return ec.fieldContext_Program_displayID(ctx, field)
			case "tags":
				return ec.fieldContext_Program_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_Program_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_Program_name(ctx, field)
			case "description":
				return ec.fieldContext_Program_description(ctx, field)
			case "status":
				return ec.fieldContext_Program_status(ctx, field)
			case "programType":
				return ec.fieldContext_Program_programType(ctx, field)
			case "frameworkName":
				return ec.fieldContext_Program_frameworkName(ctx, field)
			case "startDate":
				return ec.fieldContext_Program_startDate(ctx, field)
			case "endDate":
				return ec.fieldContext_Program_endDate(ctx, field)
			case "auditorReady":
				return ec.fieldContext_Program_auditorReady(ctx, field)
			case "auditorWriteComments":
				return ec.fieldContext_Program_auditorWriteComments(ctx, field)
			case "auditorReadComments":
				return ec.fieldContext_Program_auditorReadComments(ctx, field)
			case "auditFirm":
				return ec.fieldContext_Program_auditFirm(ctx, field)
			case "auditor":
				return ec.fieldContext_Program_auditor(ctx, field)
			case "auditorEmail":
				return ec.fieldContext_Program_auditorEmail(ctx, field)
			case "owner":
				return ec.fieldContext_Program_owner(ctx, field)
			case "blockedGroups":
				return ec.fieldContext_Program_blockedGroups(ctx, field)
			case "editors":
				return ec.fieldContext_Program_editors(ctx, field)
			case "viewers":
				return ec.fieldContext_Program_viewers(ctx, field)
			case "controls":
				return ec.fieldContext_Program_controls(ctx, field)
			case "subcontrols":
				return ec.fieldContext_Program_subcontrols(ctx, field)
			case "controlObjectives":
				return ec.fieldContext_Program_controlObjectives(ctx, field)
			case "internalPolicies":
				return ec.fieldContext_Program_internalPolicies(ctx, field)
			case "procedures":
				return ec.fieldContext_Program_procedures(ctx, field)
			case "risks":
				return ec.fieldContext_Program_risks(ctx, field)
			case "tasks":
				return ec.fieldContext_Program_tasks(ctx, field)
			case "notes":
				return ec.fieldContext_Program_notes(ctx, field)
			case "files":
				return ec.fieldContext_Program_files(ctx, field)
			case "evidence":
				return ec.fieldContext_Program_evidence(ctx, field)
			case "narratives":
				return ec.fieldContext_Program_narratives(ctx, field)
			case "actionPlans":
				return ec.fieldContext_Program_actionPlans(ctx, field)
			case "users":
				return ec.fieldContext_Program_users(ctx, field)
			case "members":
				return ec.fieldContext_Program_members(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Program", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _ProgramCreatePayload_program(ctx context.Context, field graphql.CollectedField, obj *model.ProgramCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ProgramCreatePayload_program(ctx, field)
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
		return obj.Program, nil
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
	res := resTmp.(*generated.Program)
	fc.Result = res
	return ec.marshalNProgram2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐProgram(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ProgramCreatePayload_program(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ProgramCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Program_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Program_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Program_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Program_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Program_updatedBy(ctx, field)
			case "displayID":
				return ec.fieldContext_Program_displayID(ctx, field)
			case "tags":
				return ec.fieldContext_Program_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_Program_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_Program_name(ctx, field)
			case "description":
				return ec.fieldContext_Program_description(ctx, field)
			case "status":
				return ec.fieldContext_Program_status(ctx, field)
			case "programType":
				return ec.fieldContext_Program_programType(ctx, field)
			case "frameworkName":
				return ec.fieldContext_Program_frameworkName(ctx, field)
			case "startDate":
				return ec.fieldContext_Program_startDate(ctx, field)
			case "endDate":
				return ec.fieldContext_Program_endDate(ctx, field)
			case "auditorReady":
				return ec.fieldContext_Program_auditorReady(ctx, field)
			case "auditorWriteComments":
				return ec.fieldContext_Program_auditorWriteComments(ctx, field)
			case "auditorReadComments":
				return ec.fieldContext_Program_auditorReadComments(ctx, field)
			case "auditFirm":
				return ec.fieldContext_Program_auditFirm(ctx, field)
			case "auditor":
				return ec.fieldContext_Program_auditor(ctx, field)
			case "auditorEmail":
				return ec.fieldContext_Program_auditorEmail(ctx, field)
			case "owner":
				return ec.fieldContext_Program_owner(ctx, field)
			case "blockedGroups":
				return ec.fieldContext_Program_blockedGroups(ctx, field)
			case "editors":
				return ec.fieldContext_Program_editors(ctx, field)
			case "viewers":
				return ec.fieldContext_Program_viewers(ctx, field)
			case "controls":
				return ec.fieldContext_Program_controls(ctx, field)
			case "subcontrols":
				return ec.fieldContext_Program_subcontrols(ctx, field)
			case "controlObjectives":
				return ec.fieldContext_Program_controlObjectives(ctx, field)
			case "internalPolicies":
				return ec.fieldContext_Program_internalPolicies(ctx, field)
			case "procedures":
				return ec.fieldContext_Program_procedures(ctx, field)
			case "risks":
				return ec.fieldContext_Program_risks(ctx, field)
			case "tasks":
				return ec.fieldContext_Program_tasks(ctx, field)
			case "notes":
				return ec.fieldContext_Program_notes(ctx, field)
			case "files":
				return ec.fieldContext_Program_files(ctx, field)
			case "evidence":
				return ec.fieldContext_Program_evidence(ctx, field)
			case "narratives":
				return ec.fieldContext_Program_narratives(ctx, field)
			case "actionPlans":
				return ec.fieldContext_Program_actionPlans(ctx, field)
			case "users":
				return ec.fieldContext_Program_users(ctx, field)
			case "members":
				return ec.fieldContext_Program_members(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Program", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _ProgramDeletePayload_deletedID(ctx context.Context, field graphql.CollectedField, obj *model.ProgramDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ProgramDeletePayload_deletedID(ctx, field)
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

func (ec *executionContext) fieldContext_ProgramDeletePayload_deletedID(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ProgramDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _ProgramUpdatePayload_program(ctx context.Context, field graphql.CollectedField, obj *model.ProgramUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_ProgramUpdatePayload_program(ctx, field)
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
		return obj.Program, nil
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
	res := resTmp.(*generated.Program)
	fc.Result = res
	return ec.marshalNProgram2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐProgram(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_ProgramUpdatePayload_program(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "ProgramUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Program_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Program_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Program_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Program_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Program_updatedBy(ctx, field)
			case "displayID":
				return ec.fieldContext_Program_displayID(ctx, field)
			case "tags":
				return ec.fieldContext_Program_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_Program_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_Program_name(ctx, field)
			case "description":
				return ec.fieldContext_Program_description(ctx, field)
			case "status":
				return ec.fieldContext_Program_status(ctx, field)
			case "programType":
				return ec.fieldContext_Program_programType(ctx, field)
			case "frameworkName":
				return ec.fieldContext_Program_frameworkName(ctx, field)
			case "startDate":
				return ec.fieldContext_Program_startDate(ctx, field)
			case "endDate":
				return ec.fieldContext_Program_endDate(ctx, field)
			case "auditorReady":
				return ec.fieldContext_Program_auditorReady(ctx, field)
			case "auditorWriteComments":
				return ec.fieldContext_Program_auditorWriteComments(ctx, field)
			case "auditorReadComments":
				return ec.fieldContext_Program_auditorReadComments(ctx, field)
			case "auditFirm":
				return ec.fieldContext_Program_auditFirm(ctx, field)
			case "auditor":
				return ec.fieldContext_Program_auditor(ctx, field)
			case "auditorEmail":
				return ec.fieldContext_Program_auditorEmail(ctx, field)
			case "owner":
				return ec.fieldContext_Program_owner(ctx, field)
			case "blockedGroups":
				return ec.fieldContext_Program_blockedGroups(ctx, field)
			case "editors":
				return ec.fieldContext_Program_editors(ctx, field)
			case "viewers":
				return ec.fieldContext_Program_viewers(ctx, field)
			case "controls":
				return ec.fieldContext_Program_controls(ctx, field)
			case "subcontrols":
				return ec.fieldContext_Program_subcontrols(ctx, field)
			case "controlObjectives":
				return ec.fieldContext_Program_controlObjectives(ctx, field)
			case "internalPolicies":
				return ec.fieldContext_Program_internalPolicies(ctx, field)
			case "procedures":
				return ec.fieldContext_Program_procedures(ctx, field)
			case "risks":
				return ec.fieldContext_Program_risks(ctx, field)
			case "tasks":
				return ec.fieldContext_Program_tasks(ctx, field)
			case "notes":
				return ec.fieldContext_Program_notes(ctx, field)
			case "files":
				return ec.fieldContext_Program_files(ctx, field)
			case "evidence":
				return ec.fieldContext_Program_evidence(ctx, field)
			case "narratives":
				return ec.fieldContext_Program_narratives(ctx, field)
			case "actionPlans":
				return ec.fieldContext_Program_actionPlans(ctx, field)
			case "users":
				return ec.fieldContext_Program_users(ctx, field)
			case "members":
				return ec.fieldContext_Program_members(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Program", field.Name)
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

var programBulkCreatePayloadImplementors = []string{"ProgramBulkCreatePayload"}

func (ec *executionContext) _ProgramBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ProgramBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, programBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ProgramBulkCreatePayload")
		case "programs":
			out.Values[i] = ec._ProgramBulkCreatePayload_programs(ctx, field, obj)
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

var programCreatePayloadImplementors = []string{"ProgramCreatePayload"}

func (ec *executionContext) _ProgramCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ProgramCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, programCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ProgramCreatePayload")
		case "program":
			out.Values[i] = ec._ProgramCreatePayload_program(ctx, field, obj)
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

var programDeletePayloadImplementors = []string{"ProgramDeletePayload"}

func (ec *executionContext) _ProgramDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ProgramDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, programDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ProgramDeletePayload")
		case "deletedID":
			out.Values[i] = ec._ProgramDeletePayload_deletedID(ctx, field, obj)
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

var programUpdatePayloadImplementors = []string{"ProgramUpdatePayload"}

func (ec *executionContext) _ProgramUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.ProgramUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, programUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("ProgramUpdatePayload")
		case "program":
			out.Values[i] = ec._ProgramUpdatePayload_program(ctx, field, obj)
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

func (ec *executionContext) marshalNProgramBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐProgramBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.ProgramBulkCreatePayload) graphql.Marshaler {
	return ec._ProgramBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNProgramBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐProgramBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.ProgramBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ProgramBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNProgramCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐProgramCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.ProgramCreatePayload) graphql.Marshaler {
	return ec._ProgramCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNProgramCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐProgramCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.ProgramCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ProgramCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNProgramDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐProgramDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.ProgramDeletePayload) graphql.Marshaler {
	return ec._ProgramDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNProgramDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐProgramDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.ProgramDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ProgramDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNProgramUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐProgramUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.ProgramUpdatePayload) graphql.Marshaler {
	return ec._ProgramUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNProgramUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐProgramUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.ProgramUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._ProgramUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
