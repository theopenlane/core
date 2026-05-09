package hooks

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/pkg/logx"
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

			n, err := m.Client().Template.Query().
				Select(template.FieldID).
				Where(template.KindEQ(enums.TemplateKindTrustCenterNda)).
				Count(ctx)
			if err != nil {
				return nil, err
			}

			if n == 0 {
				return nil, ErrNDATemplateRequired
			}

			email, _ := m.Email()

			queryCtx := ctx
			if _, isAnon := auth.ActiveTrustCenterIDKey.Get(ctx); isAnon {
				queryCtx = privacy.DecisionContext(ctx, privacy.Allow)
			}

			existingRequest, err := m.Client().TrustCenterNDARequest.Query().
				Where(
					trustcenterndarequest.TrustCenterIDEQ(trustCenterID),
					trustcenterndarequest.EmailEQ(email),
				).
				Only(queryCtx)
			if err != nil && !generated.IsNotFound(err) {
				return nil, err
			}

			if existingRequest != nil {
				return handleExistingNDARequest(ctx, queryCtx, m.Client(), existingRequest)
			}

			tc, err := m.Client().TrustCenter.Query().
				Where(trustcenter.IDEQ(trustCenterID)).
				WithSetting().
				Only(queryCtx)
			if err != nil {
				return nil, err
			}

			requiresApproval := tc.Edges.Setting != nil && tc.Edges.Setting.NdaApprovalRequired
			if requiresApproval {
				m.SetStatus(enums.TrustCenterNDARequestStatusNeedsApproval)
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			request, ok := v.(*generated.TrustCenterNDARequest)
			if !ok {
				return v, nil
			}

			if requiresApproval {
				if err := createNDARequestNotification(ctx, request, tc.OwnerID); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
				}
				return v, nil
			}

			if err := sendSystemEmail(ctx, m.Client(), emaildef.TCNDARequestOp.Name(), emaildef.TrustCenterNDARequestEmail{
				RecipientInfo: emaildef.RecipientInfo{Email: request.Email},
				RequestID:     request.ID,
				TrustCenterID: request.TrustCenterID,
			}); err != nil {
				return nil, err
			}

			return v, nil
		})
	}, ent.OpCreate)
}

func handleExistingNDARequest(ctx, queryCtx context.Context, client *generated.Client, existing *generated.TrustCenterNDARequest) (*generated.TrustCenterNDARequest, error) {
	switch existing.Status {
	case enums.TrustCenterNDARequestStatusSigned:
		if err := sendSystemEmail(ctx, client, emaildef.TCAuthOp.Name(), emaildef.TrustCenterAuthEmail{
			RecipientInfo: emaildef.RecipientInfo{Email: existing.Email},
			RequestID:     existing.ID,
			TrustCenterID: existing.TrustCenterID,
		}); err != nil {
			return nil, err
		}

		return existing, nil
	case enums.TrustCenterNDARequestStatusApproved, enums.TrustCenterNDARequestStatusRequested:
		if err := sendSystemEmail(ctx, client, emaildef.TCNDARequestOp.Name(), emaildef.TrustCenterNDARequestEmail{
			RecipientInfo: emaildef.RecipientInfo{Email: existing.Email},
			RequestID:     existing.ID,
			TrustCenterID: existing.TrustCenterID,
		}); err != nil {
			return nil, err
		}

		return existing, nil
	case enums.TrustCenterNDARequestStatusNeedsApproval:
		// if needs approval, recreate notification
		tc, err := getTrustCenter(ctx, existing.TrustCenterID)
		if err != nil {
			return nil, err
		}

		if err := createNDARequestNotification(ctx, existing, tc.OwnerID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
		}

		return existing, nil
	case enums.TrustCenterNDARequestStatusDeclined:
		// if previously declined, set to needs approval again to restart the process
		if err := transactionFromContext(ctx).TrustCenterNDARequest.UpdateOne(existing).SetStatus(enums.TrustCenterNDARequestStatusNeedsApproval).
			Exec(queryCtx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("email", existing.Email).Msg("failed to update NDA request status to needs approval")

			return nil, err
		}

		tc, err := getTrustCenter(ctx, existing.TrustCenterID)
		if err != nil {
			return nil, err
		}

		if err := createNDARequestNotification(ctx, existing, tc.OwnerID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
		}
	}

	// otherwise do nothing invalid
	return existing, nil
}

// getTrustCenter is a helper to get the trust center for an NDA request by ID
func getTrustCenter(ctx context.Context, trustCenterID string) (*generated.TrustCenter, error) {
	tc, err := transactionFromContext(ctx).TrustCenter.Query().
		Where(trustcenter.IDEQ(trustCenterID)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to get trust center for existing NDA request")
		return nil, err
	}

	return tc, nil
}

// HookTrustCenterNDARequestUpdate handles NDA request status updates - sends email when approved
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

			// if approved or signed, set the timestamp in the ISO8601 format
			now, err := models.ToDateTime(time.Now().UTC().Format(time.RFC3339))
			if err != nil {
				return nil, err
			}

			if status == enums.TrustCenterNDARequestStatusSigned {
				m.SetSignedAt(*now)

				retVal, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				return retVal, nil
			}

			m.SetApprovedAt(*now)

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			request, ok := v.(*generated.TrustCenterNDARequest)
			if !ok {
				return v, nil
			}

			if err := sendSystemEmail(ctx, m.Client(), emaildef.TCNDARequestOp.Name(), emaildef.TrustCenterNDARequestEmail{
				RecipientInfo: emaildef.RecipientInfo{Email: request.Email},
				RequestID:     request.ID,
				TrustCenterID: request.TrustCenterID,
			}); err != nil {
				return nil, err
			}

			return v, nil
		})
	}, ent.OpUpdateOne|ent.OpUpdate|ent.OpDeleteOne)
}

func handleNDARequestDelete(ctx context.Context, m *generated.TrustCenterNDARequestMutation) error {
	id, ok := m.ID()
	if !ok {
		logx.FromContext(ctx).Error().Msg("missing ID for deleted NDA request, unable to cleanup tuples")

		return nil
	}

	tcID, ok := m.TrustCenterID()
	if !ok && m.Op().Is(ent.OpUpdateOne) {

		oldTrustcenterID, err := m.OldTrustCenterID(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("missing trust center ID for deleted NDA request, unable to cleanup tuples")

			return err
		}

		tcID = oldTrustcenterID
	}

	// delete any tuples associated with the nda request
	deleteTuple := fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, id),
		SubjectType: "user",
		ObjectID:    tcID,
		ObjectType:  "trust_center",
		Relation:    "nda_signed",
	})

	if _, err := m.Authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{deleteTuple}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to delete relationship tuple for deleted NDA request")

		return ErrInternalServerError
	}

	return nil
}

func createNDARequestNotification(ctx context.Context, ndaRequest *generated.TrustCenterNDARequest, ownerID string) error {
	name := fmt.Sprintf("%s %s", ndaRequest.FirstName, ndaRequest.LastName)
	if name == " " {
		name = ndaRequest.Email
	}

	topic := enums.NotificationTopicApproval

	input := generated.CreateNotificationInput{
		NotificationType: enums.NotificationTypeOrganization,
		Title:            "New NDA Access Request",
		Body:             fmt.Sprintf("%s has requested access to private trust center documents.", name),
		ObjectType:       "trust_center_nda_request",
		OwnerID:          &ownerID,
		Topic:            &topic,
		Data: map[string]any{
			"nda_request_id":  ndaRequest.ID,
			"trust_center_id": ndaRequest.TrustCenterID,
			"email":           ndaRequest.Email,
			"url":             "trust-center/NDAs",
		},
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	_, err := transactionFromContext(ctx).Notification.Create().SetInput(input).Save(allowCtx)

	return err
}
