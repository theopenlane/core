package runtime

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/stripe/stripe-go/v84"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/consts"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/orgsubscription"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

const reminderStaggerDifference = 30 * time.Second

// SeedPaymentReminders starts the durable payment reminder polling loop
// after runtime listeners have been registered. It is a no-op when an active job
// already exists, preventing duplicate loops from accumulating across restarts
func (r *Runtime) SeedPaymentReminders(ctx context.Context) error {
	if !r.paymentReminderConfig.Enabled {
		logx.FromContext(ctx).Info().Msg("payment reminder listener disabled, skipping seed")
		return nil
	}

	active, err := r.Gala().HasActiveJobForTopic(ctx, operations.PaymentReminderTopic)
	if err != nil {
		return err
	}

	if active {
		logx.FromContext(ctx).Debug().Msg("payment reminder poller already active, skipping seed")
		return nil
	}

	receipt := r.Gala().EmitWithHeaders(ctx, operations.PaymentReminderTopic, operations.PaymentReminderEnvelope{}, gala.Headers{
		Tags: []string{"payment-reminders"},
	})

	return receipt.Err
}

// HandlePaymentReminders marks canceled organizations for deletion and dispatches
// notification emails to admin/owner members and the billing contact.
func (r *Runtime) HandlePaymentReminders(ctx context.Context, _ operations.PaymentReminderEnvelope) (int, error) {
	db := r.DB()
	logger := logx.FromContext(ctx)

	cfg := r.paymentReminderConfig
	if cfg.PaymentMethodInterval == 0 {
		cfg.PaymentMethodInterval = operations.DefaultPaymentMethodInterval
	}

	if cfg.DeletionDays == 0 {
		cfg.DeletionDays = operations.DefaultDeletionDays
	}

	systemCtx := auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		Capabilities: auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
	})

	now := time.Now()
	canceledBefore := now.Add(-time.Duration(cfg.PaymentMethodInterval) * 24 * time.Hour)

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

	for _, setting := range settings {
		org := setting.Edges.Organization
		if org == nil {
			continue
		}

		if cfg.DryRun {
			logger.Info().Str("organization_id", org.ID).Str("organization_name", org.Name).Msg("dry run: would schedule for deletion")
			dispatched++

			continue
		}

		pendingDeletionAt := now.AddDate(0, 0, int(cfg.DeletionDays))

		if err := db.OrganizationSetting.UpdateOneID(setting.ID).
			SetPendingDeletionAt(models.DateTime(pendingDeletionAt)).
			Exec(systemCtx); err != nil {
			logger.Error().Err(err).Str("organization_id", org.ID).Str("setting_id", setting.ID).Msg("failed to set pending_deletion_at")
			continue
		}

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

		billingEmail := strings.TrimSpace(setting.BillingEmail)
		if billingEmail != "" {
			recipients = append(recipients, paymentReminderRecipient{
				email:     billingEmail,
				firstName: "Billing",
				lastName:  "Admin",
			})
		}

		recipients = lo.UniqBy(recipients, func(rcp paymentReminderRecipient) string {
			return strings.ToLower(strings.TrimSpace(rcp.email))
		})

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

			if _, err := r.Dispatch(systemCtx, operations.DispatchRequest{
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

	return dispatched, nil
}

type paymentReminderRecipient struct {
	email     string
	firstName string
	lastName  string
}
