package hooks

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/template"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/internal/httpserve/authmanager"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// HookTrustCenterNDARequestCreate handles new NDA request creation
func HookTrustCenterNDARequestCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterNDARequestFunc(func(ctx context.Context, m *generated.TrustCenterNDARequestMutation) (generated.Value, error) {
			trustCenterID, ok := m.TrustCenterID()
			if !ok || trustCenterID == "" {
				logx.FromContext(ctx).Error().Msg("trust center ID is required for NDA request")

				return nil, ErrTrustCenterIDRequired
			}

			// do not allow any status other than the default requested from anon users
			// intentionally using wide-ranging IsAnonymousFromContext and not trust center specific
			// to ensure other anon user types also cannot set this with an anon token
			requestedStatus, _ := m.Status()
			if requestedStatus != enums.TrustCenterNDARequestStatusRequested && auth.IsAnonymousFromContext(ctx) {
				logx.FromContext(ctx).Warn().Str("status", requestedStatus.String()).Msg("nda status attempted to be set from anon context")
				return nil, ErrNDARequestStatusNotAllowed
			}

			recordSigned := requestedStatus == enums.TrustCenterNDARequestStatusSigned

			ndaExists, err := m.Client().Template.Query().
				Where(template.KindEQ(enums.TemplateKindTrustCenterNda)).
				Exist(ctx)
			if err != nil {
				return nil, err
			}

			if !ndaExists {
				return nil, ErrNDATemplateRequired
			}

			email, _ := m.Email()

			queryCtx := ctx
			if auth.IsTrustCenterFromContext(ctx) {
				queryCtx = auth.WithInternalOperationContext(ctx)
			}

			existingRequest, err := m.Client().TrustCenterNDARequest.Query().
				Where(
					trustcenterndarequest.TrustCenterIDEQ(trustCenterID),
					trustcenterndarequest.EmailEqualFold(email),
				).
				Only(queryCtx)
			if err != nil && !generated.IsNotFound(err) {
				return nil, err
			}

			if existingRequest != nil {
				if recordSigned {
					return recordSignedNDARequest(ctx, queryCtx, m, existingRequest)
				}

				return existingRequest, nil
			}

			if recordSigned {
				if err := defaultSignedAt(m); err != nil {
					return nil, err
				}
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			request, ok := v.(*generated.TrustCenterNDARequest)
			if !ok {
				return v, nil
			}

			// a request created as already signed grants access directly, there is nothing to
			// request a signature for and nothing to approve
			if recordSigned {
				if err := grantNDASignedAccess(ctx, m, []*generated.TrustCenterNDARequest{request}); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to grant nda access for created signed request")
					return nil, err
				}

				return v, nil
			}

			return v, nil
		})
	}, ent.OpCreate)
}

// HookTrustCenterNDARequestUpdate handles NDA request status updates
func HookTrustCenterNDARequestUpdate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterNDARequestFunc(func(ctx context.Context, m *generated.TrustCenterNDARequestMutation) (generated.Value, error) {
			if isDeleteOp(ctx, m) {
				if err := handleNDARequestDelete(ctx, m); err != nil {
					return nil, err
				}

				return next.Mutate(ctx, m)
			}

			status, ok := m.Status()

			// on update one, check if status is set, if not get old status
			if m.Op().Is(ent.OpUpdateOne) && (!ok || status == "") {
				oldStatus, err := m.OldStatus(ctx)

				// if status isn't set on mutation, set to the old status
				if err == nil && status == "" {
					status = oldStatus
				}
			}

			if !ok || (status != enums.TrustCenterNDARequestStatusApproved && status != enums.TrustCenterNDARequestStatusSigned) {
				return next.Mutate(ctx, m)
			}

			if status == enums.TrustCenterNDARequestStatusSigned {
				if err := defaultSignedAt(m); err != nil {
					return nil, err
				}

				// resolve the targets first, the status predicate no longer matches after the update
				requests, err := ndaRequestsFromMutation(ctx, m)
				if err != nil {
					return nil, err
				}

				retVal, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				if err := grantNDASignedAccess(ctx, m, requests); err != nil {
					return nil, err
				}

				return retVal, nil
			}

			// if approved, set the timestamp in the ISO8601 format
			now, err := models.ToDateTime(time.Now().UTC().Format(time.RFC3339))
			if err != nil {
				return nil, err
			}

			m.SetApprovedAt(*now)

			autoApproved, _ := m.AutoApproved()
			if _, ok := m.ApprovedByUserID(); !ok && !autoApproved {
				userID, err := auth.GetSubjectIDFromContext(ctx)
				if err != nil || userID == "" {
					return nil, auth.ErrNoAuthUser
				}

				m.SetApprovedByUserID(userID)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdateOne|ent.OpUpdate|ent.OpDeleteOne|ent.OpDelete)
}

func handleNDARequestDelete(ctx context.Context, m *generated.TrustCenterNDARequestMutation) error {
	requests, err := ndaRequestsFromMutation(ctx, m)
	if err != nil {
		return err
	}

	if len(requests) == 0 {
		return nil
	}

	deleteTuples := make([]fgax.TupleKey, 0, len(requests))
	for _, request := range requests {
		deleteTuples = append(deleteTuples, ndaSignedTuple(request.ID, request.TrustCenterID))
	}

	if _, err := m.Authz.WriteTupleKeys(ctx, nil, deleteTuples); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to delete relationship tuple for deleted NDA request")

		return ErrInternalServerError
	}

	return nil
}

// defaultSignedAt stamps the signed timestamp only when the caller did not record one, so a
// backfilled signature keeps its historical date
func defaultSignedAt(m *generated.TrustCenterNDARequestMutation) error {
	if _, ok := m.SignedAt(); ok {
		return nil
	}

	now, err := models.ToDateTime(time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}

	m.SetSignedAt(*now)

	return nil
}

// ndaSignedTuple builds the tuple granting trust center document access to the signer of an NDA
// request; the subject is the anonymous identity minted for that request in the access email
func ndaSignedTuple(requestID, trustCenterID string) fgax.TupleKey {
	return fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, requestID),
		SubjectType: "user",
		ObjectID:    trustCenterID,
		ObjectType:  "trust_center",
		Relation:    "nda_signed",
	})
}

// grantNDASignedAccess writes the nda_signed tuples for requests that have been signed
func grantNDASignedAccess(ctx context.Context, m *generated.TrustCenterNDARequestMutation, requests []*generated.TrustCenterNDARequest) error {
	if len(requests) == 0 {
		return nil
	}

	tuples := make([]fgax.TupleKey, 0, len(requests))
	for _, request := range requests {
		tuples = append(tuples, ndaSignedTuple(request.ID, request.TrustCenterID))
	}

	if _, err := m.Authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create nda_signed relationship tuple")

		return ErrInternalServerError
	}

	return nil
}

// ndaRequestsFromMutation loads the id and trust center id of every request the mutation targets
func ndaRequestsFromMutation(ctx context.Context, m *generated.TrustCenterNDARequestMutation) ([]*generated.TrustCenterNDARequest, error) {
	var ids []string

	switch m.Op() {
	case ent.OpDelete, ent.OpUpdate:
		var err error

		ids, err = m.IDs(ctx)
		if err != nil {
			return nil, err
		}
	case ent.OpDeleteOne, ent.OpUpdateOne:
		id, ok := m.ID()
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrInvalidInput, "id is required")
		}

		ids = []string{id}
	}

	if len(ids) == 0 {
		return nil, nil
	}

	return m.Client().TrustCenterNDARequest.Query().
		Where(trustcenterndarequest.IDIn(ids...)).
		Select(trustcenterndarequest.FieldID, trustcenterndarequest.FieldTrustCenterID).
		All(auth.WithInternalReadContext(ctx))
}

// recordSignedNDARequest marks an existing request signed when an already signed NDA is recorded
// for an email that has requested access before
func recordSignedNDARequest(ctx, queryCtx context.Context, m *generated.TrustCenterNDARequestMutation, existing *generated.TrustCenterNDARequest) (*generated.TrustCenterNDARequest, error) {
	if existing.Status == enums.TrustCenterNDARequestStatusSigned {
		return existing, nil
	}

	update := transactionFromContext(ctx).TrustCenterNDARequest.UpdateOne(existing).
		SetStatus(enums.TrustCenterNDARequestStatusSigned)

	if signedAt, ok := m.SignedAt(); ok {
		update.SetSignedAt(signedAt)
	}

	if fileID, ok := m.FileID(); ok {
		update.SetFileID(fileID)
	}

	if documentDataID, ok := m.DocumentDataID(); ok {
		update.SetDocumentDataID(documentDataID)
	}

	return update.Save(queryCtx)
}

// ndaApproverRoles are the fallback organization roles notified when no NDA approver group is configured.
var ndaApproverRoles = []enums.Role{enums.RoleOwner, enums.RoleSuperAdmin, enums.RoleAdmin}

func sendNDAApprovalRequestEmails(ctx context.Context, client *generated.Client, ndaRequest *generated.TrustCenterNDARequest, tc *generated.TrustCenter) error {
	internalCtx := auth.WithInternalReadContext(ctx)

	org, err := client.Organization.Query().
		Where(organization.IDEQ(tc.OwnerID)).
		Select(organization.FieldDisplayName).
		Only(internalCtx)
	if err != nil {
		return err
	}

	emails, err := getNDAApproverEmails(ctx, client, tc.OwnerID, tc.Edges.Setting)
	if err != nil {
		return err
	}
	if len(emails) == 0 {
		return nil
	}

	requesterName := fmt.Sprintf("%s %s", ndaRequest.FirstName, ndaRequest.LastName)
	if requesterName == " " {
		requesterName = ""
	}

	err = sendSystemEmail(ctx, emaildef.TCNDAApprovalRequestOp.Name(), emaildef.TrustCenterNDAApprovalRequestEmail{
		RecipientInfo:  emaildef.RecipientInfo{Email: emails[0], Recipients: emails},
		OrgName:        org.DisplayName,
		OrgID:          tc.OwnerID,
		RequesterName:  requesterName,
		RequesterEmail: ndaRequest.Email,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("could not send email to organization approvers")
	}

	return nil
}

func getNDAApproverEmails(ctx context.Context, client *generated.Client, ownerID string, setting *generated.TrustCenterSetting) ([]string, error) {
	ids, err := getNDAApproverUserIDs(ctx, client, ownerID, setting)
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	internalCtx := auth.WithInternalReadContext(ctx)

	var emails []string

	err = client.User.Query().
		Where(user.IDIn(ids...)).
		Select(user.FieldEmail).
		Scan(internalCtx, &emails)
	if err != nil {
		return nil, err
	}

	return lo.Uniq(lo.Filter(emails, func(email string, _ int) bool {
		return email != ""
	})), nil
}

func getNDAApproverUserIDs(ctx context.Context, client *generated.Client, ownerID string, setting *generated.TrustCenterSetting) ([]string, error) {
	// approvers are read for the trust center owner org regardless of who made the request, both queries are pinned to it
	internalCtx := auth.WithInternalReadCrossOrgContext(ctx)

	if setting != nil && setting.NdaApproverGroupID != nil && *setting.NdaApproverGroupID != "" {
		var ids []string

		err := client.GroupMembership.Query().
			Where(groupmembership.GroupID(*setting.NdaApproverGroupID)).
			Select(groupmembership.FieldUserID).
			Scan(internalCtx, &ids)
		if err != nil {
			return nil, err
		}

		return lo.Uniq(ids), nil
	}

	var ids []string

	err := client.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(ownerID),
			orgmembership.RoleIn(ndaApproverRoles...),
		).
		Select(orgmembership.FieldUserID).
		Scan(internalCtx, &ids)
	if err != nil {
		return nil, err
	}

	return lo.Uniq(ids), nil
}
