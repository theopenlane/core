package hooks

import (
	"context"
	"fmt"
	"slices"

	"entgo.io/ent/dialect/sql"
	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/v2/internal/ent/validator"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/ssoutils"

	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
)

func init() { registerListeners(NDAAutoApprovalListeners) }

type ndaApprovalSource struct {
	SkipNDAApprovalNotification bool `json:"skipNDAApprovalNotification,omitempty"`
}

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
	oc, _ := gala.OperationContextFromContext(inv.Context)
	source, _ := gala.DecodeAttributes[ndaApprovalSource](oc)
	if source.SkipNDAApprovalNotification {
		return nil
	}

	req, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.TrustCenterNDARequest.Get)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	if payload.Operation != entityops.OpCreate {
		proposedStatus, ok := payload.StringValue(trustcenterndarequest.FieldStatus)

		// for resolved requests, only notify on a status change
		if req.Status != enums.TrustCenterNDARequestStatusPendingApproval && (!ok || proposedStatus != req.Status.String()) {
			return nil
		}
	}

	if req.Status == enums.TrustCenterNDARequestStatusPendingApproval {
		req, err = processApproval(inv.Context, inv.Client, req)
		if err != nil {
			return err
		}
	}

	switch req.Status {
	// we have requested here because that will be the state still if auto approval rules not enabled
	// so user will always get the email if org has no auto approval set up or if it was auto approved
	case enums.TrustCenterNDARequestStatusRequested, enums.TrustCenterNDARequestStatusApproved:

		err := sendSystemEmail(inv.Context, emaildef.TCNDARequestOp.Name(), emaildef.TrustCenterNDARequestEmail{
			RecipientInfo: emaildef.RecipientInfo{Email: req.Email},
			RequestID:     req.ID,
			TrustCenterID: req.TrustCenterID,
		})

		if err != nil {
			logx.FromContext(inv.Context).Error().Err(err).
				Msg("failed to send trustcenter nda approval email")
		}

	case enums.TrustCenterNDARequestStatusNeedsApproval:

		tc, err := inv.Client.TrustCenter.Query().Where(
			trustcenter.ID(req.TrustCenterID),
		).
			WithSetting(func(q *generated.TrustCenterSettingQuery) {
				q.Where(trustcentersetting.EnvironmentEQ(enums.TrustCenterEnvironmentLive))
			}).
			Only(inv.Context)
		if err != nil {
			return err
		}

		if err := createNDARequestMutationNotification(inv.Context, inv.Client, req, tc.OwnerID, string(inv.Envelope.ID)); err != nil {
			return err
		}

		return sendNDAApprovalRequestEmails(inv.Context, inv.Client, req, tc)

	case enums.TrustCenterNDARequestStatusSigned:

		err := sendSystemEmail(inv.Context, emaildef.TCAuthOp.Name(), emaildef.TrustCenterAuthEmail{
			RecipientInfo: emaildef.RecipientInfo{Email: req.Email},
			RequestID:     req.ID,
			TrustCenterID: req.TrustCenterID,
		})

		if err != nil {
			logx.FromContext(inv.Context).Error().Err(err).
				Msg("failed to send trustcenter nda signed/auth email")
		}

	}

	return nil
}

func processApproval(ctx context.Context, client *generated.Client, request *generated.TrustCenterNDARequest) (*generated.TrustCenterNDARequest, error) {
	tc, err := client.TrustCenter.Query().
		Where(trustcenter.ID(request.TrustCenterID)).
		WithSetting().
		Only(ctx)
	if err != nil {
		return nil, err
	}

	settings, err := tc.Edges.SettingOrErr()
	if err != nil {
		return nil, err
	}

	approvalRules := settings.AutoApprovalRules

	ok := !settings.NdaApprovalRequired

	// fine to do this as we auto set all to pending approval in the hook
	status := enums.TrustCenterNDARequestStatusNeedsApproval

	if settings.NdaApprovalRequired && settings.EnableAutoApproval {
		ok, err = evaluateRules(ctx, client, request, &approvalRules)
		if err != nil {
			return nil, err
		}

		switch {
		case ok:
			status = enums.TrustCenterNDARequestStatusApproved

		case !approvalRules.ManualApprovalOnFailure:

			status = enums.TrustCenterNDARequestStatusDeclined
		}
	}

	// saving this will retrigger the listener, so be sure to skip it
	oc, _ := gala.OperationContextFromContext(ctx)
	source, _ := gala.DecodeAttributes[ndaApprovalSource](oc)
	source.SkipNDAApprovalNotification = true

	if err := gala.SetAttributes(&oc, source); err != nil {
		return nil, err
	}

	ctx = gala.WithOperationContext(ctx, oc)

	return client.TrustCenterNDARequest.UpdateOneID(request.ID).
		Where(
			trustcenterndarequest.StatusEQ(enums.TrustCenterNDARequestStatusPendingApproval),
			trustcenterndarequest.EmailEQ(request.Email),
		).
		SetStatus(status).
		SetAutoApproved(ok).
		ClearApprovedByUserID().
		Save(ctx)
}

func evaluateRules(ctx context.Context, client *generated.Client, request *generated.TrustCenterNDARequest, setting *models.TrustCenterNDARequestSetting) (bool, error) {
	domain := ssoutils.EmailDomain(request.Email)

	if err := validator.ValidateDomains()([]string{domain}); err != nil {
		return false, nil
	}

	approvedByList, matchedList := validateDomainList(domain, setting)
	if matchedList {
		return approvedByList, nil
	}

	if ok, err := validateContact(ctx, client, request.Email, setting); err != nil || ok {
		return ok, err
	}

	if ok, err := validateDomainFromNDARequest(ctx, client, request, domain, setting); err != nil || ok {
		return ok, err
	}

	if ok, err := validateContactDomain(ctx, client, domain, setting); err != nil || ok {
		return ok, err
	}

	var result *emailverifier.Result
	if client.EmailVerifier != nil {

		var err error
		result, err = client.EmailVerifier.Client.Verify(request.Email)
		if err != nil {

			// we do not want to fail the job. Instead log and return here
			logx.FromContext(ctx).Error().Err(err).
				Str("email", request.Email).
				Msg("could not verify email")

			return false, nil
		}
	}

	return validateEmailRules(result, setting), nil
}

func validateEmailRules(result *emailverifier.Result, setting *models.TrustCenterNDARequestSetting) bool {
	if result == nil || !result.Syntax.Valid {
		return false
	}

	if setting.AllowDisposableEmail && result.Disposable {
		return true
	}

	if setting.AllowRoleAccount && result.RoleAccount {
		return true
	}

	if setting.WorkEmailOnly && !result.Free && !result.Disposable {
		return true
	}

	return false
}

func validateDomainList(domain string, setting *models.TrustCenterNDARequestSetting) (isApproved, didMatch bool) {
	if setting.UseDomainBlocklist && slices.Contains(setting.DomainBlocklist, domain) {
		return false, true
	}

	if setting.UseDomainAllowlist && slices.Contains(setting.DomainAllowlist, domain) {
		return true, true
	}

	return false, false
}

func validateContact(ctx context.Context, client *generated.Client, email string, setting *models.TrustCenterNDARequestSetting) (bool, error) {
	if !setting.ApproveIfContactExists {
		return false, nil
	}

	return client.Contact.Query().
		Where(
			contact.StatusEQ(enums.UserStatusActive),
			contact.EmailEqualFold(email),
		).
		Exist(ctx)
}

func validateDomainFromNDARequest(ctx context.Context, client *generated.Client, request *generated.TrustCenterNDARequest, domain string, setting *models.TrustCenterNDARequestSetting) (bool, error) {
	if !setting.ApproveFromExistingRequestDomain {
		return false, nil
	}

	return client.TrustCenterNDARequest.Query().
		Where(
			trustcenterndarequest.TrustCenterID(request.TrustCenterID),
			trustcenterndarequest.IDNEQ(request.ID),
			trustcenterndarequest.StatusIn(enums.TrustCenterNDARequestStatusApproved, enums.TrustCenterNDARequestStatusSigned),
			sql.FieldHasSuffixFold(trustcenterndarequest.FieldEmail, domain),
		).
		Exist(ctx)
}

func validateContactDomain(ctx context.Context, client *generated.Client, domain string, setting *models.TrustCenterNDARequestSetting) (bool, error) {
	if !setting.ApproveFromContactDomain {
		return false, nil
	}

	return client.Contact.Query().
		Where(
			contact.StatusEQ(enums.UserStatusActive),
			sql.FieldHasSuffixFold(contact.FieldEmail, domain),
		).
		Exist(ctx)
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
			"url":             entityops.ConsoleLanding(generated.TypeTrustCenterNDARequest),
		},
	}

	_, err := client.Notification.Create().SetID(notificationID).SetInput(input).Save(ctx)
	if IsUniqueConstraintError(err) {
		return nil
	}

	return err
}
