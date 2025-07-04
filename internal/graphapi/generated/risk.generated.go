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

func (ec *executionContext) _RiskBulkCreatePayload_risks(ctx context.Context, field graphql.CollectedField, obj *model.RiskBulkCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_RiskBulkCreatePayload_risks(ctx, field)
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
		return obj.Risks, nil
	})
	if err != nil {
		ec.Error(ctx, err)
		return graphql.Null
	}
	if resTmp == nil {
		return graphql.Null
	}
	res := resTmp.([]*generated.Risk)
	fc.Result = res
	return ec.marshalORisk2ᚕᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐRiskᚄ(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_RiskBulkCreatePayload_risks(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "RiskBulkCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Risk_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Risk_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Risk_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Risk_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Risk_updatedBy(ctx, field)
			case "displayID":
				return ec.fieldContext_Risk_displayID(ctx, field)
			case "tags":
				return ec.fieldContext_Risk_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_Risk_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_Risk_name(ctx, field)
			case "status":
				return ec.fieldContext_Risk_status(ctx, field)
			case "riskType":
				return ec.fieldContext_Risk_riskType(ctx, field)
			case "category":
				return ec.fieldContext_Risk_category(ctx, field)
			case "impact":
				return ec.fieldContext_Risk_impact(ctx, field)
			case "likelihood":
				return ec.fieldContext_Risk_likelihood(ctx, field)
			case "score":
				return ec.fieldContext_Risk_score(ctx, field)
			case "mitigation":
				return ec.fieldContext_Risk_mitigation(ctx, field)
			case "details":
				return ec.fieldContext_Risk_details(ctx, field)
			case "businessCosts":
				return ec.fieldContext_Risk_businessCosts(ctx, field)
			case "stakeholderID":
				return ec.fieldContext_Risk_stakeholderID(ctx, field)
			case "delegateID":
				return ec.fieldContext_Risk_delegateID(ctx, field)
			case "owner":
				return ec.fieldContext_Risk_owner(ctx, field)
			case "blockedGroups":
				return ec.fieldContext_Risk_blockedGroups(ctx, field)
			case "editors":
				return ec.fieldContext_Risk_editors(ctx, field)
			case "viewers":
				return ec.fieldContext_Risk_viewers(ctx, field)
			case "controls":
				return ec.fieldContext_Risk_controls(ctx, field)
			case "subcontrols":
				return ec.fieldContext_Risk_subcontrols(ctx, field)
			case "procedures":
				return ec.fieldContext_Risk_procedures(ctx, field)
			case "internalPolicies":
				return ec.fieldContext_Risk_internalPolicies(ctx, field)
			case "programs":
				return ec.fieldContext_Risk_programs(ctx, field)
			case "actionPlans":
				return ec.fieldContext_Risk_actionPlans(ctx, field)
			case "tasks":
				return ec.fieldContext_Risk_tasks(ctx, field)
			case "assets":
				return ec.fieldContext_Risk_assets(ctx, field)
			case "entities":
				return ec.fieldContext_Risk_entities(ctx, field)
			case "scans":
				return ec.fieldContext_Risk_scans(ctx, field)
			case "stakeholder":
				return ec.fieldContext_Risk_stakeholder(ctx, field)
			case "delegate":
				return ec.fieldContext_Risk_delegate(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Risk", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _RiskCreatePayload_risk(ctx context.Context, field graphql.CollectedField, obj *model.RiskCreatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_RiskCreatePayload_risk(ctx, field)
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
		return obj.Risk, nil
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
	res := resTmp.(*generated.Risk)
	fc.Result = res
	return ec.marshalNRisk2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐRisk(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_RiskCreatePayload_risk(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "RiskCreatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Risk_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Risk_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Risk_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Risk_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Risk_updatedBy(ctx, field)
			case "displayID":
				return ec.fieldContext_Risk_displayID(ctx, field)
			case "tags":
				return ec.fieldContext_Risk_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_Risk_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_Risk_name(ctx, field)
			case "status":
				return ec.fieldContext_Risk_status(ctx, field)
			case "riskType":
				return ec.fieldContext_Risk_riskType(ctx, field)
			case "category":
				return ec.fieldContext_Risk_category(ctx, field)
			case "impact":
				return ec.fieldContext_Risk_impact(ctx, field)
			case "likelihood":
				return ec.fieldContext_Risk_likelihood(ctx, field)
			case "score":
				return ec.fieldContext_Risk_score(ctx, field)
			case "mitigation":
				return ec.fieldContext_Risk_mitigation(ctx, field)
			case "details":
				return ec.fieldContext_Risk_details(ctx, field)
			case "businessCosts":
				return ec.fieldContext_Risk_businessCosts(ctx, field)
			case "stakeholderID":
				return ec.fieldContext_Risk_stakeholderID(ctx, field)
			case "delegateID":
				return ec.fieldContext_Risk_delegateID(ctx, field)
			case "owner":
				return ec.fieldContext_Risk_owner(ctx, field)
			case "blockedGroups":
				return ec.fieldContext_Risk_blockedGroups(ctx, field)
			case "editors":
				return ec.fieldContext_Risk_editors(ctx, field)
			case "viewers":
				return ec.fieldContext_Risk_viewers(ctx, field)
			case "controls":
				return ec.fieldContext_Risk_controls(ctx, field)
			case "subcontrols":
				return ec.fieldContext_Risk_subcontrols(ctx, field)
			case "procedures":
				return ec.fieldContext_Risk_procedures(ctx, field)
			case "internalPolicies":
				return ec.fieldContext_Risk_internalPolicies(ctx, field)
			case "programs":
				return ec.fieldContext_Risk_programs(ctx, field)
			case "actionPlans":
				return ec.fieldContext_Risk_actionPlans(ctx, field)
			case "tasks":
				return ec.fieldContext_Risk_tasks(ctx, field)
			case "assets":
				return ec.fieldContext_Risk_assets(ctx, field)
			case "entities":
				return ec.fieldContext_Risk_entities(ctx, field)
			case "scans":
				return ec.fieldContext_Risk_scans(ctx, field)
			case "stakeholder":
				return ec.fieldContext_Risk_stakeholder(ctx, field)
			case "delegate":
				return ec.fieldContext_Risk_delegate(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Risk", field.Name)
		},
	}
	return fc, nil
}

func (ec *executionContext) _RiskDeletePayload_deletedID(ctx context.Context, field graphql.CollectedField, obj *model.RiskDeletePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_RiskDeletePayload_deletedID(ctx, field)
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

func (ec *executionContext) fieldContext_RiskDeletePayload_deletedID(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "RiskDeletePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			return nil, errors.New("field of type ID does not have child fields")
		},
	}
	return fc, nil
}

func (ec *executionContext) _RiskUpdatePayload_risk(ctx context.Context, field graphql.CollectedField, obj *model.RiskUpdatePayload) (ret graphql.Marshaler) {
	fc, err := ec.fieldContext_RiskUpdatePayload_risk(ctx, field)
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
		return obj.Risk, nil
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
	res := resTmp.(*generated.Risk)
	fc.Result = res
	return ec.marshalNRisk2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋentᚋgeneratedᚐRisk(ctx, field.Selections, res)
}

func (ec *executionContext) fieldContext_RiskUpdatePayload_risk(_ context.Context, field graphql.CollectedField) (fc *graphql.FieldContext, err error) {
	fc = &graphql.FieldContext{
		Object:     "RiskUpdatePayload",
		Field:      field,
		IsMethod:   false,
		IsResolver: false,
		Child: func(ctx context.Context, field graphql.CollectedField) (*graphql.FieldContext, error) {
			switch field.Name {
			case "id":
				return ec.fieldContext_Risk_id(ctx, field)
			case "createdAt":
				return ec.fieldContext_Risk_createdAt(ctx, field)
			case "updatedAt":
				return ec.fieldContext_Risk_updatedAt(ctx, field)
			case "createdBy":
				return ec.fieldContext_Risk_createdBy(ctx, field)
			case "updatedBy":
				return ec.fieldContext_Risk_updatedBy(ctx, field)
			case "displayID":
				return ec.fieldContext_Risk_displayID(ctx, field)
			case "tags":
				return ec.fieldContext_Risk_tags(ctx, field)
			case "ownerID":
				return ec.fieldContext_Risk_ownerID(ctx, field)
			case "name":
				return ec.fieldContext_Risk_name(ctx, field)
			case "status":
				return ec.fieldContext_Risk_status(ctx, field)
			case "riskType":
				return ec.fieldContext_Risk_riskType(ctx, field)
			case "category":
				return ec.fieldContext_Risk_category(ctx, field)
			case "impact":
				return ec.fieldContext_Risk_impact(ctx, field)
			case "likelihood":
				return ec.fieldContext_Risk_likelihood(ctx, field)
			case "score":
				return ec.fieldContext_Risk_score(ctx, field)
			case "mitigation":
				return ec.fieldContext_Risk_mitigation(ctx, field)
			case "details":
				return ec.fieldContext_Risk_details(ctx, field)
			case "businessCosts":
				return ec.fieldContext_Risk_businessCosts(ctx, field)
			case "stakeholderID":
				return ec.fieldContext_Risk_stakeholderID(ctx, field)
			case "delegateID":
				return ec.fieldContext_Risk_delegateID(ctx, field)
			case "owner":
				return ec.fieldContext_Risk_owner(ctx, field)
			case "blockedGroups":
				return ec.fieldContext_Risk_blockedGroups(ctx, field)
			case "editors":
				return ec.fieldContext_Risk_editors(ctx, field)
			case "viewers":
				return ec.fieldContext_Risk_viewers(ctx, field)
			case "controls":
				return ec.fieldContext_Risk_controls(ctx, field)
			case "subcontrols":
				return ec.fieldContext_Risk_subcontrols(ctx, field)
			case "procedures":
				return ec.fieldContext_Risk_procedures(ctx, field)
			case "internalPolicies":
				return ec.fieldContext_Risk_internalPolicies(ctx, field)
			case "programs":
				return ec.fieldContext_Risk_programs(ctx, field)
			case "actionPlans":
				return ec.fieldContext_Risk_actionPlans(ctx, field)
			case "tasks":
				return ec.fieldContext_Risk_tasks(ctx, field)
			case "assets":
				return ec.fieldContext_Risk_assets(ctx, field)
			case "entities":
				return ec.fieldContext_Risk_entities(ctx, field)
			case "scans":
				return ec.fieldContext_Risk_scans(ctx, field)
			case "stakeholder":
				return ec.fieldContext_Risk_stakeholder(ctx, field)
			case "delegate":
				return ec.fieldContext_Risk_delegate(ctx, field)
			}
			return nil, fmt.Errorf("no field named %q was found under type Risk", field.Name)
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

var riskBulkCreatePayloadImplementors = []string{"RiskBulkCreatePayload"}

func (ec *executionContext) _RiskBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.RiskBulkCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, riskBulkCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("RiskBulkCreatePayload")
		case "risks":
			out.Values[i] = ec._RiskBulkCreatePayload_risks(ctx, field, obj)
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

var riskCreatePayloadImplementors = []string{"RiskCreatePayload"}

func (ec *executionContext) _RiskCreatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.RiskCreatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, riskCreatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("RiskCreatePayload")
		case "risk":
			out.Values[i] = ec._RiskCreatePayload_risk(ctx, field, obj)
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

var riskDeletePayloadImplementors = []string{"RiskDeletePayload"}

func (ec *executionContext) _RiskDeletePayload(ctx context.Context, sel ast.SelectionSet, obj *model.RiskDeletePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, riskDeletePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("RiskDeletePayload")
		case "deletedID":
			out.Values[i] = ec._RiskDeletePayload_deletedID(ctx, field, obj)
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

var riskUpdatePayloadImplementors = []string{"RiskUpdatePayload"}

func (ec *executionContext) _RiskUpdatePayload(ctx context.Context, sel ast.SelectionSet, obj *model.RiskUpdatePayload) graphql.Marshaler {
	fields := graphql.CollectFields(ec.OperationContext, sel, riskUpdatePayloadImplementors)

	out := graphql.NewFieldSet(fields)
	deferred := make(map[string]*graphql.FieldSet)
	for i, field := range fields {
		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("RiskUpdatePayload")
		case "risk":
			out.Values[i] = ec._RiskUpdatePayload_risk(ctx, field, obj)
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

func (ec *executionContext) marshalNRiskBulkCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐRiskBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.RiskBulkCreatePayload) graphql.Marshaler {
	return ec._RiskBulkCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNRiskBulkCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐRiskBulkCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.RiskBulkCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._RiskBulkCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNRiskCreatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐRiskCreatePayload(ctx context.Context, sel ast.SelectionSet, v model.RiskCreatePayload) graphql.Marshaler {
	return ec._RiskCreatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNRiskCreatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐRiskCreatePayload(ctx context.Context, sel ast.SelectionSet, v *model.RiskCreatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._RiskCreatePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNRiskDeletePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐRiskDeletePayload(ctx context.Context, sel ast.SelectionSet, v model.RiskDeletePayload) graphql.Marshaler {
	return ec._RiskDeletePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNRiskDeletePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐRiskDeletePayload(ctx context.Context, sel ast.SelectionSet, v *model.RiskDeletePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._RiskDeletePayload(ctx, sel, v)
}

func (ec *executionContext) marshalNRiskUpdatePayload2githubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐRiskUpdatePayload(ctx context.Context, sel ast.SelectionSet, v model.RiskUpdatePayload) graphql.Marshaler {
	return ec._RiskUpdatePayload(ctx, sel, &v)
}

func (ec *executionContext) marshalNRiskUpdatePayload2ᚖgithubᚗcomᚋtheopenlaneᚋcoreᚋinternalᚋgraphapiᚋmodelᚐRiskUpdatePayload(ctx context.Context, sel ast.SelectionSet, v *model.RiskUpdatePayload) graphql.Marshaler {
	if v == nil {
		if !graphql.HasFieldError(ctx, graphql.GetFieldContext(ctx)) {
			ec.Errorf(ctx, "the requested element is null which the schema does not allow")
		}
		return graphql.Null
	}
	return ec._RiskUpdatePayload(ctx, sel, v)
}

// endregion ***************************** type.gotpl *****************************
