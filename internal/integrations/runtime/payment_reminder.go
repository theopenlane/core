package runtime

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
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

// HandlePaymentReminders queries organizations without payment methods past the
// configured interval, marks them for deletion, and dispatches notification emails
// to admin/owner members and the billing contact. Returns the number of emails
// dispatched as the delta for adaptive scheduling
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

	settings, err := db.OrganizationSetting.Query().
		Where(
			organizationsetting.PendingDeletionAtIsNil(),
			organizationsetting.PaymentMethodAddedEQ(false),
			organizationsetting.HasOrganizationWith(
				organization.PersonalOrg(false),
				organization.Not(
					organization.HasOrgSubscriptionsWith(
						orgsubscription.ActiveEQ(true),
					),
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

	for _, setting := range settings {
		org := setting.Edges.Organization
		if org == nil {
			continue
		}

		if !isPastPaymentInterval(org.CreatedAt, cfg.PaymentMethodInterval) {
			continue
		}

		if cfg.DryRun {
			logger.Info().Str("organization_id", org.ID).Str("organization_name", org.Name).Msg("dry run: would schedule for deletion")
			dispatched++

			continue
		}

		pendingDeletionAt := time.Now().AddDate(0, 0, int(cfg.DeletionDays))

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

			if _, err := r.Dispatch(systemCtx, operations.DispatchRequest{
				DefinitionID: emaildef.DefinitionID.ID(),
				Operation:    emaildef.OrgDeletionNoticeOp.Name(),
				Config:       config,
				RunType:      enums.IntegrationRunTypeScheduled,
				Runtime:      true,
			}); err != nil {
				logger.Error().Err(err).Str("organization_id", org.ID).Str("email", rcp.email).Msg("failed to dispatch deletion notice email")
				continue
			}

			dispatched++
		}

		logger.Info().Str("organization_id", org.ID).Int("recipients", len(recipients)).Msg("payment reminder dispatched")
	}

	return dispatched, nil
}

// paymentReminderRecipient holds the recipient details for a payment reminder email
type paymentReminderRecipient struct {
	email     string
	firstName string
	lastName  string
}

// isPastPaymentInterval reports whether the org creation time is past the
// configured payment method interval
func isPastPaymentInterval(createdAt time.Time, intervalDays uint8) bool {
	if intervalDays == 0 {
		return true
	}

	return time.Since(createdAt) >= time.Duration(intervalDays)*24*time.Hour
}
