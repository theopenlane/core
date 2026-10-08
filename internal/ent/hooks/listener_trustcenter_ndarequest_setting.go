package hooks

import (
	"context"
	"fmt"
	"slices"

	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/v2/internal/ent/validator"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/ssoutils"

	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
)

func init() { registerListeners(NDAAutoApprovalListeners) }

func NDAAutoApprovalListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema: entityops.SchemaTrustCenterNDARequest,
			Operations: []string{
				entityops.OpCreate,
				entityops.OpUpdate,
				entityops.OpUpdateOne,
			},
			Fields: []string{
				trustcenterndarequest.FieldStatus,
				trustcenterndarequest.FieldEmail,
			},
			Handle: handleNDARequestApproval,
			RowMatch: []entityops.FieldMatch{
				{
					Field: trustcenterndarequest.FieldStatus,
					In: []string{
						string(enums.TrustCenterNDARequestStatusPendingApproval),
					},
				},
			},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return auth.NewOrgInternalCaller(restored.OrganizationID)
			},
			ContextKeys: []func(context.Context) context.Context{
				func(ctx context.Context) context.Context {
					// remove the anon trustcenter key from the ctx
					return auth.ActiveTrustCenterIDKey.Set(ctx, "")
				},
			},
		},
	}
}

func handleNDARequestApproval(inv entityops.Invocation, payload entityops.MutationPayload) error {
	req, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.TrustCenterNDARequest.Get)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	if req.Status == enums.TrustCenterNDARequestStatusPendingApproval {
		req, err = processApproval(inv.Context, inv.Client, req)
		if err != nil {
			return err
		}
	}

	switch req.Status {
	case enums.TrustCenterNDARequestStatusRequested, enums.TrustCenterNDARequestStatusApproved:

		return sendSystemEmail(inv.Context, emaildef.TCNDARequestOp.Name(), emaildef.TrustCenterNDARequestEmail{
			RecipientInfo: emaildef.RecipientInfo{Email: req.Email},
			RequestID:     req.ID,
			TrustCenterID: req.TrustCenterID,
		})

	case enums.TrustCenterNDARequestStatusNeedsApproval:

		tc, err := inv.Client.TrustCenter.Query().Where(trustcenter.ID(req.TrustCenterID)).Only(inv.Context)
		if err != nil {
			return err
		}
		if err := createNDARequestMutationNotification(inv.Context, inv.Client, req, tc.OwnerID, string(inv.Envelope.ID)); err != nil {
			return err
		}
		return sendNDAApprovalRequestEmails(inv.Context, inv.Client, req, tc)

	case enums.TrustCenterNDARequestStatusSigned:

		status, _ := payload.StringValue(trustcenterndarequest.FieldStatus)
		if payload.Operation != entityops.OpCreate || status == enums.TrustCenterNDARequestStatusSigned.String() {
			return nil
		}

		return sendSystemEmail(inv.Context, emaildef.TCAuthOp.Name(), emaildef.TrustCenterAuthEmail{
			RecipientInfo: emaildef.RecipientInfo{Email: req.Email},
			RequestID:     req.ID,
			TrustCenterID: req.TrustCenterID,
		})
	}

	return nil
}

func processApproval(ctx context.Context, client *generated.Client, request *generated.TrustCenterNDARequest) (*generated.TrustCenterNDARequest, error) {
	settings, err := fetchNDARequestSetting(ctx, client, request.TrustCenterID)
	if err != nil {
		return nil, err
	}

	ndaSettings := settings.AutoApprovalRules

	status := enums.TrustCenterNDARequestStatusNeedsApproval

	isApproved := !settings.NdaApprovalRequired

	if settings.NdaApprovalRequired {
		isApproved, err = evaluateRules(ctx, client, request, &ndaSettings)
		if err != nil {
			return nil, err
		}
	}

	if isApproved {
		status = enums.TrustCenterNDARequestStatusApproved
	} else if !ndaSettings.ManualApprovalOnFailure {
		status = enums.TrustCenterNDARequestStatusDeclined
	}

	return client.TrustCenterNDARequest.UpdateOneID(request.ID).
		Where(
			trustcenterndarequest.StatusEQ(enums.TrustCenterNDARequestStatusPendingApproval),
			trustcenterndarequest.EmailEQ(request.Email),
		).
		SetStatus(status).
		ClearApprovedByUserID().
		Save(ctx)
}

func fetchNDARequestSetting(ctx context.Context, client *generated.Client, id string) (*generated.TrustCenterSetting, error) {
	tc, err := client.TrustCenter.Query().
		Where(trustcenter.ID(id)).
		WithSetting().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return tc.Edges.SettingOrErr()
}

func evaluateRules(ctx context.Context, client *generated.Client, request *generated.TrustCenterNDARequest, setting *models.TrustCenterNDARequestSetting) (bool, error) {
	domain := ssoutils.EmailDomain(request.Email)

	if err := validator.ValidateDomains()([]string{domain}); err != nil {
		return false, nil
	}

	if client.EmailVerifier == nil {
		return false, nil
	}

	result, err := client.EmailVerifier.Client.Verify(request.Email)
	if err != nil {
		return false, err
	}

	if !result.Syntax.Valid {
		return false, nil
	}

	if result.Disposable && !setting.AllowDisposableEmail {
		return false, nil
	}

	if result.RoleAccount && !setting.AllowRoleAccount {
		return false, nil
	}

	isWorkDomain := !result.Free && !result.Disposable

	if setting.WorkEmailOnly && !isWorkDomain {
		return false, nil
	}

	if setting.UseDomainBlocklist && slices.Contains(setting.DomainBlocklist, domain) {
		return false, nil
	}

	if setting.UseDomainAllowlist && slices.Contains(setting.DomainAllowlist, domain) {
		return true, nil
	}

	if setting.ApproveIfContactExists {

		ok, err := client.Contact.Query().
			Where(
				contact.StatusEQ(enums.UserStatusActive),
				contact.EmailEqualFold(request.Email),
			).
			Exist(ctx)
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}
	}

	if setting.ApproveFromExistingRequestDomain && isWorkDomain {

		ok, err := client.TrustCenterNDARequest.Query().
			Where(
				trustcenterndarequest.TrustCenterID(request.TrustCenterID),
				trustcenterndarequest.IDNEQ(request.ID),
				trustcenterndarequest.StatusIn(enums.TrustCenterNDARequestStatusApproved, enums.TrustCenterNDARequestStatusSigned),
				// match existing nda requests if another approved/signed one already exists using the same email domain
				sql.FieldHasSuffixFold(trustcenterndarequest.FieldEmail, domain),
			).
			Exist(ctx)
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}
	}

	if setting.ApproveFromContactDomain && isWorkDomain {
		ok, err := client.Contact.Query().
			Where(
				contact.StatusEQ(enums.UserStatusActive),
				// match existing contacts by domain
				sql.FieldHasSuffixFold(contact.FieldEmail, domain),
			).
			Exist(ctx)
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}
	}

	return !setting.UseDomainAllowlist && !setting.ApproveIfContactExists &&
		!setting.ApproveFromExistingRequestDomain && !setting.ApproveFromContactDomain, nil
}

func createNDARequestMutationNotification(ctx context.Context, client *generated.Client, ndaRequest *generated.TrustCenterNDARequest, ownerID, notificationID string) error {
	if _, err := client.Notification.Get(ctx, notificationID); err == nil {
		return nil
	} else if !generated.IsNotFound(err) {
		return err
	}

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

	_, err := client.Notification.Create().SetID(notificationID).SetInput(input).Save(ctx)
	if IsUniqueConstraintError(err) {
		return nil
	}

	return err
}
