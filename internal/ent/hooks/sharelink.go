package hooks

import (
	"context"
	"net/mail"
	"time"

	"entgo.io/ent"
	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/pkg/objects/storage/proxy"
)

// HookShareLinkCreate refuses share links when storage links are disabled, the organization has not enabled external
// sharing, or the caller cannot read every shared R2 file; it normalizes the recipients, decides whether the share URL
// opens the landing page, stores the opaque share token with its secret, and sets the share URL on the created link
func HookShareLinkCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ShareLinkFunc(func(ctx context.Context, m *generated.ShareLinkMutation) (generated.Value, error) {
			if !m.Client().ObjectManager.Links().Enabled {
				return nil, ErrShareLinksDisabled
			}

			orgID, err := shareLinkOrganization(ctx, m)
			if err != nil {
				return nil, err
			}

			if err := checkExternalSharingEnabled(ctx, m, orgID); err != nil {
				return nil, err
			}

			expiresAt, err := shareLinkExpiry(m)
			if err != nil {
				return nil, err
			}

			shared, err := sharedFiles(ctx, m)
			if err != nil {
				return nil, err
			}

			recipients, err := normalizeShareLinkRecipients(m)
			if err != nil {
				return nil, err
			}

			// a page is needed to pick between files and to enter the recipient verification code, and a capped link
			// opens one so only a deliberate download counts as a use
			requested, _ := m.LandingPage()
			maxUses, _ := m.MaxUses()
			m.SetLandingPage(requested || shared > 1 || recipients > 0 || maxUses != proxy.UnlimitedShareUses)

			id, _ := m.ID()

			tokenOrgID, err := ulid.Parse(orgID)
			if err != nil {
				return nil, err
			}

			shareToken, err := tokens.NewShareToken(id, tokenOrgID, time.Until(expiresAt))
			if err != nil {
				return nil, err
			}

			signature, secret, err := shareToken.Sign()
			if err != nil {
				return nil, err
			}

			// the organization id travels in the token so the share link is looked up within its organization
			token := proxy.ComposeShareToken(orgID, signature)

			m.SetToken(token)
			m.SetSecret(secret)
			// the stored expiry is the signed one so the token can be rebuilt from the row
			m.SetExpiresAt(shareToken.ExpiresAt)

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			link, ok := v.(*generated.ShareLink)
			if !ok {
				return v, nil
			}

			link.URL = proxy.ComposeShareURL(m.Client().ObjectManager.Links().BaseURL, id, token)

			return link, nil
		})
	}, ent.OpCreate)
}

// shareLinkOrganization returns the organization the share link is created in, from the owner set by the org-owned
// mixin or the caller's organization when the owner was not derived
func shareLinkOrganization(ctx context.Context, m *generated.ShareLinkMutation) (string, error) {
	if ownerID, ok := m.OwnerID(); ok && ownerID != "" {
		return ownerID, nil
	}

	return auth.GetOrganizationIDFromContext(ctx)
}

// checkExternalSharingEnabled refuses the share link when the organization setting does not allow external sharing
func checkExternalSharingEnabled(ctx context.Context, m *generated.ShareLinkMutation, orgID string) error {
	setting, err := m.Client().OrganizationSetting.Query().
		Where(organizationsetting.OrganizationID(orgID)).
		Select(organizationsetting.FieldExternalSharingEnabled).
		Only(ctx)
	if err != nil {
		return err
	}

	if !setting.ExternalSharingEnabled {
		return ErrShareLinkExternalSharingDisabled
	}

	return nil
}

// shareLinkExpiry returns the share link expiry, refusing one in the past or further out than the maximum
func shareLinkExpiry(m *generated.ShareLinkMutation) (time.Time, error) {
	expiresAt, ok := m.ExpiresAt()
	if !ok {
		return time.Time{}, ErrShareLinkExpiryRequired
	}

	now := time.Now()

	switch {
	case expiresAt.Before(now):
		return time.Time{}, ErrShareLinkExpiryPast
	case expiresAt.After(now.Add(proxy.MaxShareDuration)):
		return time.Time{}, ErrShareLinkExpiryTooLong
	default:
		return expiresAt, nil
	}
}

// sharedFiles loads the files the share link grants under the caller's context, so a file the caller cannot read
// refuses the share link rather than being shared blindly, refuses a file that is not stored on R2, and returns the
// number of files shared
func sharedFiles(ctx context.Context, m *generated.ShareLinkMutation) (int, error) {
	fileIDs := lo.Uniq(m.FilesIDs())

	switch {
	case len(fileIDs) == 0:
		return 0, ErrShareLinkFileAccess
	case len(fileIDs) > proxy.MaxShareFiles:
		return 0, ErrShareLinkTooManyFiles
	}

	files, err := m.Client().File.Query().Where(file.IDIn(fileIDs...)).All(ctx)
	if err != nil {
		return 0, err
	}

	if len(files) != len(fileIDs) {
		return 0, ErrShareLinkFileAccess
	}

	if lo.SomeBy(files, func(f *generated.File) bool { return f.StorageProvider != string(storagetypes.R2Provider) }) {
		return 0, ErrShareLinkFileNotOnR2
	}

	return len(files), nil
}

// normalizeShareLinkRecipients stores the lowercased, de-duplicated bare recipient addresses and returns how many
// the share link is bound to, refusing an entry that is not an email address
func normalizeShareLinkRecipients(m *generated.ShareLinkMutation) (int, error) {
	emails, ok := m.RecipientEmails()
	if !ok {
		return 0, nil
	}

	normalized := make([]string, 0, len(emails))

	for _, email := range emails {
		address, err := mail.ParseAddress(email)
		if err != nil {
			return 0, ErrShareLinkRecipientInvalid
		}

		normalized = append(normalized, proxy.NormalizeRecipientEmail(address.Address))
	}

	normalized = lo.Uniq(normalized)

	if len(normalized) == 0 {
		m.ResetRecipientEmails()

		return 0, nil
	}

	m.SetRecipientEmails(normalized)

	return len(normalized), nil
}

// HookShareLinkUpdate refuses to undo a revocation and refuses a download cap on a share link that serves its file
// directly, where every request would count as a use
func HookShareLinkUpdate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ShareLinkFunc(func(ctx context.Context, m *generated.ShareLinkMutation) (generated.Value, error) {
			if m.RevokedAtCleared() {
				return nil, ErrShareLinkRevocationPermanent
			}

			if maxUses, ok := m.MaxUses(); ok && maxUses != proxy.UnlimitedShareUses {
				landingPage, err := m.OldLandingPage(ctx)
				if err != nil {
					return nil, err
				}

				if !landingPage {
					return nil, ErrShareLinkDirectUseCap
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdateOne)
}
