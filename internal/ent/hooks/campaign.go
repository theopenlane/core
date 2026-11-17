package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	emaildef "github.com/theopenlane/core/v2/internal/integrations/definitions/email"
)

// HookCampaignEmailTemplate rejects linking an email template that does not fit the campaign type
func HookCampaignEmailTemplate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.CampaignFunc(func(ctx context.Context, m *generated.CampaignMutation) (generated.Value, error) {
			if err := validateCampaignEmailTemplate(ctx, m); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// validateCampaignEmailTemplate checks the template key against the campaign type when either one changes
func validateCampaignEmailTemplate(ctx context.Context, m *generated.CampaignMutation) error {
	templateID, templateSet := m.EmailTemplateID()
	campaignType, typeSet := m.CampaignType()

	if (!templateSet && !typeSet) || m.EmailTemplateIDCleared() {
		return nil
	}

	if m.Op().Is(ent.OpUpdateOne) {
		var err error

		if !templateSet {
			if templateID, err = m.OldEmailTemplateID(ctx); err != nil {
				return err
			}
		}

		if !typeSet {
			if campaignType, err = m.OldCampaignType(ctx); err != nil {
				return err
			}
		}
	}

	if templateID == "" || campaignType == enums.CampaignTypeTrustCenterUpdate {
		return nil
	}

	key, err := m.Client().EmailTemplate.Query().
		Where(emailtemplate.ID(templateID)).
		Select(emailtemplate.FieldKey).
		String(ctx)
	if err != nil {
		return err
	}

	return emaildef.ValidateCampaignTemplateKey(campaignType, key)
}
