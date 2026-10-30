package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/entfga"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/logx"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/group"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/report"
	"github.com/theopenlane/core/v2/internal/ent/generated/reportmembership"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	gqlgenerated "github.com/theopenlane/core/v2/internal/graphapi/generated"
)

var (
	gqlSchema = gqlgenerated.NewExecutableSchema(gqlgenerated.Config{}).Schema()

	// ErrReportIsPrivate is returned when adding members to a private report
	ErrReportIsPrivate = errors.New("you cannot add or update membership of a private report. change visibility to restricted first")
)

// isQueryValid checks if the provided query string is valid against the server. it also validates
// the variables too
func isQueryValid(query string, variables map[string]any) error {
	document, documentErr := gqlparser.LoadQueryWithRules(gqlSchema, query, nil)
	if len(documentErr) > 0 {
		return documentErr
	}

	// only queries should ever be allowed + one query per report
	if len(document.Operations) != 1 || document.Operations[0].Operation != ast.Query {
		return errors.New("reporting must contain exactly one query operation") //nolint:err113
	}

	_, err := validator.VariableValues(gqlSchema, document.Operations[0], variables)
	return err
}

// HookValidateReportQuery makes sure to validate the provided query and variables that they can be executed
// against the server
func HookValidateReportQuery() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ReportFunc(func(ctx context.Context, m *generated.ReportMutation) (generated.Value, error) {
			query, queryExists := m.Query()
			variables, variablesExist := m.Variables()

			isUpdateOp := m.Op().Is(ent.OpUpdateOne)

			// if nothing changed in an update operation, we want to return early
			// validation should only kick in if it is a new creation or either the query/variable
			// has changed
			if isUpdateOp && !queryExists && !variablesExist && !m.VariablesCleared() {
				return next.Mutate(ctx, m)
			}

			if isUpdateOp {
				var err error

				if !queryExists {
					query, err = m.OldQuery(ctx)
					if err != nil {
						logx.FromContext(ctx).Error().Err(err).Msg("failed to get old report query")

						return nil, err
					}
				}

				if !variablesExist && !m.VariablesCleared() {
					variables, err = m.OldVariables(ctx)
					if err != nil {
						logx.FromContext(ctx).Error().Err(err).Msg("failed to get old report variables")

						return nil, err
					}
				}
			}

			if err := isQueryValid(query, variables); err != nil {
				return nil, err
			}

			// clients can always reformat if they need to
			m.SetQuery(strings.TrimSpace(query))
			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookReportAuthz manages the tuples for access to the report across users + organization based off their visibility
func HookReportAuthz() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ReportFunc(func(ctx context.Context, m *generated.ReportMutation) (ent.Value, error) {

			requestedVisibility, visibilityExists := m.Visibility()

			addedGroupAccess := len(m.ViewersIDs()) > 0 || len(m.EditorsIDs()) > 0

			// we only care about running this hook if visibility changes or new groups have been added
			// regular members go through the reportmembership schema
			if m.Op().Is(ent.OpUpdateOne) && !visibilityExists && !addedGroupAccess {
				return next.Mutate(ctx, m)
			}

			// "userless" requests must explicitly use organization visibility
			if m.Op().Is(ent.OpCreate) && auth.IsUserlessContext(ctx) && requestedVisibility != enums.ReportVisibilityOrganization {
				return nil, fmt.Errorf("%w: userless reports require organization visibility", ErrInvalidInput)
			}

			if m.Op().Is(ent.OpCreate) && !visibilityExists {
				requestedVisibility = enums.ReportVisibilityPrivate
				m.SetVisibility(requestedVisibility)
			}

			oldVisibility, err := getReportVisibility(ctx, m)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).
					Msg("could not fetch the old visibility of this report")

				return nil, err
			}

			if m.Op().Is(ent.OpUpdateOne) && !visibilityExists {
				requestedVisibility = oldVisibility
			}

			// reports required to have restricted visibility for it to have members or groups
			if addedGroupAccess && requestedVisibility == enums.ReportVisibilityPrivate {
				return nil, ErrReportIsPrivate
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to mutate report")

				return v, err
			}

			// if visibility did not change. no need to sync tuples
			if m.Op().Is(ent.OpUpdateOne) && requestedVisibility == oldVisibility {
				return v, nil
			}

			report, ok := v.(*generated.Report)
			if !ok {
				return nil, ErrInternalServerError
			}

			if err := createDefaultReportMember(ctx, m, report); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("report_id", report.ID).Msg("failed to create report creator membership")

				return nil, err
			}

			if err := syncReportAccessTuples(ctx, m, report, oldVisibility); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("report_id", report.ID).Msg("failed to synchronize report access")

				return nil, err
			}

			return v, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// HookReportGroupPermissions makes sure direct mutationson groups do not bypass
// the visibility rules of reports
func HookReportGroupPermissions() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.GroupFunc(func(ctx context.Context, m *generated.GroupMutation) (ent.Value, error) {

			ids := append(m.ReportEditorsIDs(), m.ReportViewersIDs()...)
			if len(ids) == 0 {
				return next.Mutate(ctx, m)
			}

			ok, err := m.Client().Report.Query().
				Where(
					report.IDIn(ids...),
					report.VisibilityEQ(enums.ReportVisibilityPrivate)).
				Exist(auth.WithInternalReadContext(ctx))
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to check report visibility for group permissions")

				return nil, err
			}

			if ok {
				return nil, ErrReportIsPrivate
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne)
}

func getReportVisibility(ctx context.Context, m *generated.ReportMutation) (enums.ReportVisibility, error) {
	if !m.Op().Is(ent.OpUpdateOne) {
		return "", nil
	}

	return m.OldVisibility(ctx)
}

func createDefaultReportMember(ctx context.Context, m *generated.ReportMutation, report *generated.Report) error {
	if !m.Op().Is(ent.OpCreate) || auth.IsUserlessContext(ctx) {
		return nil
	}

	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil {
		return err
	}

	internalCtx := auth.WithInternalOperationContext(ctx)
	client := m.Client()

	id, err := client.OrgMembership.Query().
		Where(
			orgmembership.UserID(userID),
			orgmembership.OrganizationID(report.OwnerID),
		).
		OnlyID(internalCtx)
	if err != nil || id == "" {
		return ErrUserNotInOrg
	}

	input := generated.CreateReportMembershipInput{
		UserID:   userID,
		ReportID: report.ID,
		Role:     &enums.RoleOwner,
	}

	return client.ReportMembership.Create().
		SetInput(input).
		SetOrgMembershipID(id).
		Exec(internalCtx)
}

func syncReportAccessTuples(ctx context.Context, m *generated.ReportMutation, storedReport *generated.Report,
	previousVisibility enums.ReportVisibility) error {

	var writes, deletes []fgax.TupleKey

	isCreateOp := m.Op().Is(ent.OpCreate)

	isVisibilityChanged := m.Op().Is(ent.OpUpdateOne) && previousVisibility != storedReport.Visibility

	orgAccessTuple := fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:       storedReport.OwnerID,
		SubjectType:     generated.TypeOrganization,
		SubjectRelation: fgax.MemberRelation,
		ObjectID:        storedReport.ID,
		ObjectType:      generated.TypeReport,
		Relation:        "organization_viewer",
	})

	// if visibility is now "organization", we want to create the org wide access tuple
	if (isCreateOp || isVisibilityChanged) && storedReport.Visibility == enums.ReportVisibilityOrganization {
		writes = append(writes, orgAccessTuple)
	}

	if isVisibilityChanged && previousVisibility == enums.ReportVisibilityOrganization {
		deletes = append(deletes, orgAccessTuple)
	}

	if err := wipeExistingPrivateAccess(ctx, m, storedReport); err != nil {
		return err
	}

	_, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, writes, deletes)
	return err
}

func wipeExistingPrivateAccess(ctx context.Context, m *generated.ReportMutation, storedReport *generated.Report) error {
	// changing to private visibility needs to clear out the groups + members
	if !m.Op().Is(ent.OpUpdateOne) || storedReport.Visibility != enums.ReportVisibilityPrivate {
		return nil
	}

	cleanupCtx := entfga.WithDeleteTuplesFirst(auth.WithInternalOperationContext(ctx))

	ids, err := m.Client().Group.Query().
		Where(group.Or(
			group.HasReportEditorsWith(report.ID(storedReport.ID)),
			group.HasReportViewersWith(report.ID(storedReport.ID)),
		)).
		IDs(auth.WithInternalReadContext(ctx))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get report groups for private visibility")

		return err
	}

	err = m.Client().Report.UpdateOneID(storedReport.ID).
		RemoveEditorIDs(ids...).
		RemoveViewerIDs(ids...).
		Exec(cleanupCtx)
	if err != nil {
		return err
	}

	_, err = m.Client().ReportMembership.Delete().
		Where(
			reportmembership.ReportID(storedReport.ID),
			reportmembership.UserIDNEQ(storedReport.CreatedBy),
		).
		Exec(cleanupCtx)
	if err != nil {
		return err
	}

	return nil
}
