package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/privacy/utils"
	"github.com/theopenlane/ent/validator"
)

// MutationWithEmail is an interface that mutations that require email validation must implement
type MutationWithEmail interface {
	Email() (string, bool)

	utils.GenericMutation
}

// HookEmailValidation runs on user mutations to validate email addresses to ensure they meet the configured criteria
// which could include checks for disposable, free, or role-based emails. Additionally, it can set a default avatar
// using Gravatar if no avatar is provided during user creation.
// This hook only accepts mutations that implement the MutationWithEmail interface or are Invite mutations,
// which used the recipient field.
func HookEmailValidation() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut, ok := m.(utils.GenericMutation)
			if !ok || mut == nil {
				return next.Mutate(ctx, m)
			}

			// skip if email validation is not enabled
			if !mut.Client().EntConfig.EmailValidation.Enabled {
				return next.Mutate(ctx, m)
			}

			email := getEmailFromMutation(m)
			if email == "" {
				// no email to validate, skip
				return next.Mutate(ctx, m)
			}

			// if email validation is enabled, verify the email address
			verified, res, err := mut.Client().EmailVerifier.VerifyEmailAddress(email)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("email", email).Msg("error verifying email address")

				return nil, validator.ErrEmailNotAllowed
			}

			if !verified {
				logx.FromContext(ctx).Error().Str("email", email).Interface("result", res).Msg("email address not allowed")

				return nil, validator.ErrEmailNotAllowed
			}

			// if we get a gravatar result, and the user did not provide an avatar, set the gravatar url
			if mut.Client().EntConfig.EmailValidation.EnableGravatarCheck && mut.Type() == generated.TypeUser {
				userMut, ok := m.(*generated.UserMutation)
				if !ok || userMut == nil {
					return next.Mutate(ctx, m)
				}

				avatarURL, _ := userMut.AvatarRemoteURL()
				if res != nil && res.Gravatar != nil && res.Gravatar.HasGravatar && avatarURL == "" {
					userMut.SetAvatarRemoteURL(res.Gravatar.GravatarUrl)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// getEmailFromMutation extracts the email from the mutation if it implements MutationWithEmail
// or if it is an InviteMutation. Returns an empty string if no email is found.
func getEmailFromMutation(m ent.Mutation) string {
	mut, ok := m.(MutationWithEmail)
	if ok && mut != nil {
		email, ok := mut.Email()
		if ok {
			return email
		}
	}

	inviteMut, ok := m.(*generated.InviteMutation)
	if ok && inviteMut != nil {
		email, ok := inviteMut.Recipient()
		if ok {
			return email
		}
	}

	return ""
}
