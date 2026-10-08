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
	"github.com/theopenlane/core/v2/internal/ent/validator"
	"github.com/theopenlane/core/v2/pkg/gala"
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
	case enums.TrustCenterNDARequestStatusRequested, enums.TrustCenterNDARequestStatusApproved:

		return sendSystemEmail(inv.Context, emaildef.TCNDARequestOp.Name(), emaildef.TrustCenterNDARequestEmail{
			RecipientInfo: emaildef.RecipientInfo{Email: req.Email},
			RequestID:     req.ID,
			TrustCenterID: req.TrustCenterID,
		})

	case enums.TrustCenterNDARequestStatusNeedsApproval:

		tc, err := inv.Client.TrustCenter.Query().Where(trustcenter.ID(req.TrustCenterID)).WithSetting().Only(inv.Context)
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

	isApproved := !settings.NdaApprovalRequired

	if settings.NdaApprovalRequired && settings.EnableAutoApproval {
		isApproved, err = evaluateRules(ctx, client, request, &approvalRules)
		if err != nil {
			return nil, err
		}
	}

	status := enums.TrustCenterNDARequestStatusNeedsApproval
	if isApproved {
		status = enums.TrustCenterNDARequestStatusApproved
	} else if !approvalRules.ManualApprovalOnFailure {
		status = enums.TrustCenterNDARequestStatusDeclined
	}

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
		SetAutoApproved(isApproved).
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
			return false, err
		}
	}

	if !validateEmailRules(result, setting) {
		return false, nil
	}

	return !setting.UseDomainAllowlist && !setting.ApproveIfContactExists &&
		!setting.ApproveFromExistingRequestDomain && !setting.ApproveFromContactDomain, nil
}

func validateEmailRules(result *emailverifier.Result, setting *models.TrustCenterNDARequestSetting) bool {
	if result == nil {
		return !setting.WorkEmailOnly && setting.AllowDisposableEmail && setting.AllowRoleAccount
	}

	if !result.Syntax.Valid {
		return false
	}

	if result.Disposable && !setting.AllowDisposableEmail {
		return false
	}

	if result.RoleAccount && !setting.AllowRoleAccount {
		return false
	}

	return !setting.WorkEmailOnly || (!result.Free && !result.Disposable)
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
