package hooks

import (
	"context"
	"fmt"
	"slices"
	"time"

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
	"github.com/theopenlane/core/v2/pkg/ssoutils"
	"github.com/theopenlane/logx"

	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
)

func init() { registerListeners(NDAAutoApprovalListeners) }

// ndaNotificationCooldown is the window within which a repeat request for the same NDA request and
// status notifies nobody, so a lost link can be resent after the window while repeat submissions
// through the anonymous create cannot flood the requester or the approvers
const ndaNotificationCooldown = 5 * time.Minute

// ndaNotifyingStatuses are the request statuses that notify the requester or the approvers
var ndaNotifyingStatuses = []enums.TrustCenterNDARequestStatus{
	enums.TrustCenterNDARequestStatusRequested,
	enums.TrustCenterNDARequestStatusApproved,
	enums.TrustCenterNDARequestStatusNeedsApproval,
	enums.TrustCenterNDARequestStatusSigned,
}

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

	// updating may contain either email or status but both are not required to be available at the same time
	// so we need to fetch the actual data and cannot rely on just status from the payload
	req, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.TrustCenterNDARequest.Get)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	// if this is an update mutation and the status field didn't change, we don't need to run the approval evaluation
	if payload.Operation != entityops.OpCreate && !payload.FieldChanged(trustcenterndarequest.FieldStatus) {
		return nil
	}

	if req.Status == enums.TrustCenterNDARequestStatusRequested {
		req, err = processApproval(inv.Context, inv.Client, req)
		if err != nil {
			return err
		}
	}

	if !slices.Contains(ndaNotifyingStatuses, req.Status) {
		return nil
	}

	allowed, err := allowNDANotification(inv.Context, req)
	if err != nil {
		return err
	}

	if !allowed {
		logx.FromContext(inv.Context).Debug().Str("status", req.Status.String()).Msg("nda request notification already sent within the cooldown, skipping")

		return nil
	}

	switch req.Status {
	// requested status is the default case and we still fall in here if auto approval is not turned on
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

// allowNDANotification consumes the single notification slot for the request and status within the
// cooldown; the status is part of the key so a transition always notifies and only a repeat of the
// same status is held back. Without an integration runtime no email can be sent, so nothing is bounded
func allowNDANotification(ctx context.Context, req *generated.TrustCenterNDARequest) (bool, error) {
	rt := intruntime.Default()
	if rt == nil {
		return true, nil
	}

	return rt.AllowN(ctx, ndaNotificationKey(req), 1, 1, ndaNotificationCooldown)
}

// ndaNotificationKey scopes the notification window to one request and status
func ndaNotificationKey(req *generated.TrustCenterNDARequest) string {
	return fmt.Sprintf("trustcenter:nda:%s:%s", req.ID, req.Status)
}

func processApproval(ctx context.Context, client *generated.Client, request *generated.TrustCenterNDARequest) (*generated.TrustCenterNDARequest, error) {
	setting, err := client.TrustCenterSetting.Query().
		Where(
			trustcentersetting.EnvironmentEQ(enums.TrustCenterEnvironmentLive),
			trustcentersetting.TrustCenterID(request.TrustCenterID),
		).Only(ctx)
	if err != nil {
		return nil, err
	}

	if !setting.NdaApprovalRequired {
		return request, nil
	}

	rules := setting.AutoApprovalRules

	ok := false

	status := enums.TrustCenterNDARequestStatusNeedsApproval

	if setting.EnableAutoApproval {
		ok, err = evaluateRules(ctx, client, request, rules)
		if err != nil {
			return nil, err
		}

		if ok {
			status = enums.TrustCenterNDARequestStatusApproved
		} else if !rules.ManualApprovalOnFailure {
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
			trustcenterndarequest.StatusEQ(request.Status),
			trustcenterndarequest.EmailEQ(request.Email),
		).
		SetStatus(status).
		SetAutoApproved(ok).
		ClearApprovedByUserID().
		Save(ctx)
}

func evaluateRules(ctx context.Context, client *generated.Client, req *generated.TrustCenterNDARequest, rules models.TrustCenterNDARequestSetting) (bool, error) {

	domain := ssoutils.EmailDomain(req.Email)

	if err := validator.ValidateDomains()([]string{domain}); err != nil {
		return false, nil
	}

	approvedByList, matchedList := validateDomainList(domain, rules)
	if matchedList {
		return approvedByList, nil
	}

	if ok, err := validateContact(ctx, client, req.Email, rules); err != nil || ok {
		return ok, err
	}

	if isBlocked := validateEmailRules(ctx, client, rules, req); isBlocked {
		return false, nil
	}

	if ok, err := validateDomainFromNDARequest(ctx, client, req, domain, rules); err != nil || ok {
		return ok, err
	}

	return validateContactDomain(ctx, client, domain, rules)
}

func validateEmailRules(ctx context.Context, client *generated.Client,
	setting models.TrustCenterNDARequestSetting, req *generated.TrustCenterNDARequest) bool {

	var result *emailverifier.Result

	if client.EmailVerifier != nil {
		var err error
		result, err = client.EmailVerifier.Client.Verify(req.Email)
		if err != nil {
			// we do not want to fail the job. Instead log and return here
			logx.FromContext(ctx).Error().Err(err).
				Str("email", req.Email).
				Msg("could not verify email")

			return true
		}
	}

	if result == nil || !result.Syntax.Valid {
		return true
	}

	if setting.WorkEmailOnly && (result.Free || result.Disposable) {
		return true
	}

	if result.Disposable && !setting.AllowDisposableEmail {
		return true
	}

	return result.RoleAccount && !setting.AllowRoleAccount
}

func validateDomainList(domain string, rules models.TrustCenterNDARequestSetting) (isApproved, didMatch bool) {
	if rules.UseDomainBlocklist && slices.Contains(rules.DomainBlocklist, domain) {
		return false, true
	}

	if rules.UseDomainAllowlist && slices.Contains(rules.DomainAllowlist, domain) {
		return true, true
	}

	return false, false
}

func validateContact(ctx context.Context, client *generated.Client, email string, rules models.TrustCenterNDARequestSetting) (bool, error) {
	if !rules.ApproveIfContactExists {
		return false, nil
	}

	return client.Contact.Query().
		Where(
			contact.StatusEQ(enums.UserStatusActive),
			contact.EmailEqualFold(email),
		).
		Exist(ctx)
}

func validateDomainFromNDARequest(ctx context.Context, client *generated.Client, req *generated.TrustCenterNDARequest,
	domain string, rules models.TrustCenterNDARequestSetting) (bool, error) {

	if !rules.ApproveFromExistingRequestDomain {
		return false, nil
	}

	return client.TrustCenterNDARequest.Query().
		Where(
			trustcenterndarequest.TrustCenterID(req.TrustCenterID),
			trustcenterndarequest.IDNEQ(req.ID),
			trustcenterndarequest.StatusIn(enums.TrustCenterNDARequestStatusApproved, enums.TrustCenterNDARequestStatusSigned),
			sql.FieldHasSuffixFold(trustcenterndarequest.FieldEmail, domain),
		).
		Exist(ctx)
}

func validateContactDomain(ctx context.Context, client *generated.Client, domain string, rules models.TrustCenterNDARequestSetting) (bool, error) {
	if !rules.ApproveFromContactDomain {
		return false, nil
	}

	return client.Contact.Query().
		Where(
			contact.StatusEQ(enums.UserStatusActive),
			sql.FieldHasSuffixFold(contact.FieldEmail, domain),
		).
		Exist(ctx)
}

func createNDARequestMutationNotification(ctx context.Context, client *generated.Client, req *generated.TrustCenterNDARequest, ownerID, notificationID string) error {
	if _, err := client.Notification.Get(ctx, notificationID); err == nil {
		return nil
	} else if !generated.IsNotFound(err) {
		return err
	}

	name := fmt.Sprintf("%s %s", req.FirstName, req.LastName)
	if name == " " {
		name = req.Email
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
			"nda_request_id":  req.ID,
			"trust_center_id": req.TrustCenterID,
			"email":           req.Email,
			"url":             entityops.ConsoleLanding(generated.TypeTrustCenterNDARequest),
		},
	}

	err := client.Notification.Create().SetID(notificationID).SetInput(input).Exec(ctx)
	if IsUniqueConstraintError(err) {
		return nil
	}

	return err
}
