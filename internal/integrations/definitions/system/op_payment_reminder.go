package system

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v84"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/consts"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	slackdef "github.com/theopenlane/core/internal/integrations/definitions/slack"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// reminderStaggerDifference spaces successive deletion notice emails apart to avoid
// bursting the email provider
const reminderStaggerDifference = 30 * time.Second

// Handle adapts the payment reminder sweep to the generic operation registration boundary;
// the receiver carries the operator defaults and request config overlays a copy
func (p PaymentReminderSweep) Handle() types.OperationHandler {
	return func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
		sweep := p

		if err := jsonx.UnmarshalIfPresent(req.Config, &sweep); err != nil {
			return nil, ErrOperationConfigInvalid
		}

		processed, err := sweep.Run(ctx, req)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(types.ScheduledCycleResult{Processed: processed}, ErrResultEncode)
	}
}

// Run executes one payment reminder sweep and returns the number of dispatched notifications
func (p PaymentReminderSweep) Run(ctx context.Context, req types.OperationRequest) (int, error) {
	db := req.DB
	logger := logx.FromContext(ctx)

	if p.PaymentMethodInterval == 0 {
		p.PaymentMethodInterval = DefaultPaymentMethodInterval
	}

	if p.DeletionDays == 0 {
		p.DeletionDays = DefaultDeletionDays
	}

	systemCtx := systemSweepContext(ctx)

	now := time.Now()
	canceledBefore := now.Add(-time.Duration(p.PaymentMethodInterval) * 24 * time.Hour)

	settings, err := db.OrganizationSetting.Query().
		Where(
			organizationsetting.PendingDeletionAtIsNil(),
			organizationsetting.HasOrganizationWith(
				organization.IDNEQ(consts.SystemAdminOrgID),
				organization.PersonalOrg(false),
				organization.Not(
					organization.HasOrgSubscriptionsWith(activeOrTrialingSubscriptionPredicates()...),
				),
				organization.HasOrgSubscriptionsWith(
					orgsubscription.ActiveEQ(false),
					orgsubscription.UpdatedAtLTE(canceledBefore),
					orgsubscription.StripeSubscriptionStatusNEQ(string(stripe.SubscriptionStatusTrialing)),
					orgsubscription.StripeSubscriptionStatusNEQ(string(stripe.SubscriptionStatusActive)),
				),
			),
		).
		WithOrganization().
		All(systemCtx)
	if err != nil {
		logger.Error().Err(err).Msg("failed querying organization settings for payment reminders")
		return 0, err
	}

	if len(settings) == 0 {
		return 0, nil
	}

	logger.Info().Int("count", len(settings)).Msg("organization settings eligible for payment reminder check")

	dispatched := 0
	emailQueueOffset := 0
	orgsToDelete := make([]string, 0, len(settings))

	for _, setting := range settings {
		org := setting.Edges.Organization
		if org == nil {
			continue
		}

		if p.DryRun {
			logger.Info().Str("organization_id", org.ID).Str("organization_name", org.Name).Msg("dry run: would schedule for deletion")
			orgsToDelete = append(orgsToDelete, fmt.Sprintf("%s (%s)", org.Name, org.ID))
			dispatched++

			continue
		}

		pendingDeletionAt := now.AddDate(0, 0, int(p.DeletionDays))

		if err := db.OrganizationSetting.UpdateOneID(setting.ID).
			SetPendingDeletionAt(models.DateTime(pendingDeletionAt)).
			Exec(systemCtx); err != nil {
			logger.Error().Err(err).Str("organization_id", org.ID).Str("setting_id", setting.ID).Msg("failed to set pending_deletion_at")
			continue
		}

		orgsToDelete = append(orgsToDelete, fmt.Sprintf("%s (%s)", org.Name, org.ID))

		members, err := db.OrgMembership.Query().
			Where(
				orgmembership.OrganizationIDEQ(org.ID),
				orgmembership.RoleIn(enums.RoleAdmin, enums.RoleOwner),
			).
			WithUser().
			All(systemCtx)
		if err != nil {
			logger.Error().Err(err).Str("organization_id", org.ID).Msg("failed to fetch admin members")
			continue
		}

		recipients := paymentReminderRecipients(members, setting.BillingEmail)

		for _, rcp := range recipients {
			input := emaildef.OrgDeletionNoticeEmail{
				RecipientInfo: emaildef.RecipientInfo{
					Email:     rcp.email,
					FirstName: rcp.firstName,
					LastName:  rcp.lastName,
				},
				OrgName:      org.Name,
				DeletionDate: pendingDeletionAt,
			}

			config, err := json.Marshal(input)
			if err != nil {
				logger.Error().Err(err).Str("organization_id", org.ID).Msg("failed to marshal deletion notice email")
				continue
			}

			scheduledAt := now.Add(time.Duration(emailQueueOffset) * reminderStaggerDifference)

			if _, err := req.Dispatch(systemCtx, types.DispatchRequest{
				DefinitionID: emaildef.DefinitionID.ID(),
				Operation:    emaildef.OrgDeletionNoticeOp.Name(),
				Config:       config,
				RunType:      enums.IntegrationRunTypeScheduled,
				Runtime:      true,
				ScheduledAt:  &scheduledAt,
			}); err != nil {
				logger.Error().Err(err).Str("organization_id", org.ID).Str("email", rcp.email).Msg("failed to dispatch deletion notice email")
				continue
			}

			emailQueueOffset++
			dispatched++
		}

		logger.Info().Str("organization_id", org.ID).Int("recipients", len(recipients)).Msg("payment reminder dispatched")
	}

	if len(orgsToDelete) > 0 {
		config, err := json.Marshal(slackdef.OrganizationsPendingDeletionMessage{
			Count:         len(orgsToDelete),
			Organizations: orgsToDelete,
			DryRun:        p.DryRun,
		})
		if err != nil {
			logger.Error().Err(err).Msg("failed to marshal org deletion reminder slack notification")
			return dispatched, err
		}

		if _, err := req.Dispatch(systemCtx, types.DispatchRequest{
			DefinitionID: slackdef.DefinitionID.ID(),
			Operation:    slackdef.OrganizationsPendingDeletionOp.Name(),
			Config:       config,
			RunType:      enums.IntegrationRunTypeEvent,
			Runtime:      true,
		}); err != nil {
			logger.Error().Err(err).Int("count", len(orgsToDelete)).Msg("failed to dispatch org deletion reminder slack notification")
		}
	}

	return dispatched, nil
}

type paymentReminderRecipient struct {
	email     string
	firstName string
	lastName  string
}

// paymentReminderRecipients collects the unique admin/owner members and billing contact
// eligible for a deletion notice
func paymentReminderRecipients(members []*ent.OrgMembership, billingEmail string) []paymentReminderRecipient {
	recipients := make([]paymentReminderRecipient, 0, len(members)+1)

	for _, member := range members {
		user := member.Edges.User
		if user == nil || user.Email == "" {
			continue
		}

		recipients = append(recipients, paymentReminderRecipient{
			email:     user.Email,
			firstName: user.FirstName,
			lastName:  user.LastName,
		})
	}

	billingEmail = strings.TrimSpace(billingEmail)
	if billingEmail != "" {
		recipients = append(recipients, paymentReminderRecipient{
			email:     billingEmail,
			firstName: "Billing",
			lastName:  "Admin",
		})
	}

	return lo.UniqBy(recipients, func(rcp paymentReminderRecipient) string {
		return strings.ToLower(strings.TrimSpace(rcp.email))
	})
}
