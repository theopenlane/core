package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	gowebauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailverificationtoken"
	"github.com/theopenlane/core/internal/ent/generated/event"
	"github.com/theopenlane/core/internal/ent/generated/filedownloadtoken"
	"github.com/theopenlane/core/internal/ent/generated/invite"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/passwordresettoken"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/ent/generated/usersetting"
	"github.com/theopenlane/core/internal/ent/generated/webauthn"
	"github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/trustcenterurl"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/metrics"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	sso "github.com/theopenlane/core/pkg/ssoutils"
)

// updateUserLastSeen updates the last seen timestamp of the user and login method used
func (h *Handler) updateUserLastSeen(ctx context.Context, id string, authProvider enums.AuthProvider) error {
	if err := transaction.FromContext(ctx).
		User.
		UpdateOneID(id).
		SetLastSeen(time.Now()).
		SetLastLoginProvider(authProvider).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user last seen")

		return err
	}

	return nil
}

// createUser creates a user in the database based on the input and returns the user with user settings
func (h *Handler) createUser(ctx context.Context, input ent.CreateUserInput) (*ent.User, error) {
	meowuser, err := transaction.FromContext(ctx).User.Create().
		SetInput(input).
		Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating new user")

		return nil, err
	}

	metrics.RecordRegistration()

	return meowuser, nil
}

// updateSubscriberVerifiedEmail updates a subscriber in the database based on the input and sets it to
// active with a verified email. It does NOT clear the unsubscribed flag — an unsubscribed contact can
// only re-subscribe through the createSubscriber mutation, not by replaying a verify link. The token is
// expired (ttl set to now) so it is single-use and cannot be replayed
func (h *Handler) updateSubscriberVerifiedEmail(ctx context.Context, id string, input ent.UpdateSubscriberInput) error {
	err := transaction.FromContext(ctx).Subscriber.UpdateOneID(id).
		SetInput(input).
		SetActive(true).
		SetVerifiedEmail(true).
		SetTTL(time.Now()).
		Exec(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating subscriber verified")
		return err
	}

	return nil
}

// updateSubscriberVerificationToken rotates the subscriber's token fields and counts the resend against
// the send-attempt budget so repeated expired-link hits cannot send unbounded verification emails
func (h *Handler) updateSubscriberVerificationToken(ctx context.Context, subscriberID, token string, ttl time.Time, secret []byte) error {
	err := transaction.FromContext(ctx).Subscriber.UpdateOneID(subscriberID).
		SetToken(token).
		SetSecret(secret).
		SetTTL(ttl).
		AddSendAttempts(1).
		Exec(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating subscriber tokens")

		return err
	}

	return nil
}

// setSubscriberUnsubscribed marks the subscriber as unsubscribed; the HookSubscriberUpdated hook clears
// the active flag and resets send attempts. The token is expired (ttl set to now) so the link is
// single-use; re-presenting it is a harmless no-op via the unsubscribed flag
func (h *Handler) setSubscriberUnsubscribed(ctx context.Context, id string) error {
	err := transaction.FromContext(ctx).Subscriber.UpdateOneID(id).
		SetUnsubscribed(true).
		SetTTL(time.Now()).
		Exec(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error unsubscribing subscriber")

		return err
	}

	return nil
}

// createEmailVerificationToken creates a new email verification for the user
func (h *Handler) createEmailVerificationToken(ctx context.Context, user *User) (*ent.EmailVerificationToken, error) {
	ttl, err := time.Parse(time.RFC3339Nano, user.EmailVerificationExpires.String)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to parse ttl")
		return nil, err
	}

	meowtoken, err := transaction.FromContext(ctx).EmailVerificationToken.Create().
		SetOwnerID(user.ID).
		SetToken(user.EmailVerificationToken.String).
		SetTTL(ttl).
		SetEmail(user.Email).
		SetSecret(user.EmailVerificationSecret).
		Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating email verification token")

		return nil, err
	}

	return meowtoken, nil
}

// createPasswordResetToken creates a new password reset token for the user
func (h *Handler) createPasswordResetToken(ctx context.Context, user *User) (*ent.PasswordResetToken, error) {
	ttl, err := time.Parse(time.RFC3339Nano, user.PasswordResetExpires.String)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to parse ttl")
		return nil, err
	}

	meowtoken, err := transaction.FromContext(ctx).PasswordResetToken.Create().
		SetOwnerID(user.ID).
		SetToken(user.PasswordResetToken.String).
		SetTTL(ttl).
		SetEmail(user.Email).
		SetSecret(user.PasswordResetSecret).
		Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating password reset token")

		return nil, err
	}

	return meowtoken, nil
}

// getUserByEVToken returns the ent user with the user settings and email verification token fields based on the
// token in the request
func (h *Handler) getUserByEVToken(ctx context.Context, token string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).EmailVerificationToken.Query().
		Where(
			emailverificationtoken.Token(token),
		).
		QueryOwner().WithSetting().WithEmailVerificationTokens().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining user from email verification token")

		return nil, err
	}

	return user, nil
}

// getFilebyDownloadToken returns the ent file and download token based on the token in the request
func (h *Handler) getFilebyDownloadToken(ctx context.Context, token string) (*ent.File, *ent.FileDownloadToken, error) {
	tokenRecord, err := transaction.FromContext(ctx).FileDownloadToken.Query().
		Where(filedownloadtoken.Token(token)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining file download token")

		return nil, nil, err
	}

	if tokenRecord.FileID == nil || *tokenRecord.FileID == "" {
		logx.FromContext(ctx).Error().Msg("file download token missing file id")

		return nil, nil, ErrDownloadTokenMissingFile
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	fileRecord, err := transaction.FromContext(ctx).File.Get(allowCtx, *tokenRecord.FileID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining file from download token")

		return nil, nil, err
	}

	return fileRecord, tokenRecord, nil
}

// getUserByResetToken returns the ent user with the user settings and password reset tokens based on the
// token in the request
func (h *Handler) getUserByResetToken(ctx context.Context, token string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).PasswordResetToken.Query().
		Where(
			passwordresettoken.Token(token),
		).
		QueryOwner().WithSetting().WithPasswordResetTokens().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining user from reset token")

		return nil, err
	}

	return user, nil
}

// getUserByEmail returns the ent user with the user settings based on the email in the request
func (h *Handler) getUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).User.Query().WithSetting().
		Where(user.EmailEqualFold(email)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Str("email", email).Err(err).Msg("error obtaining user from email")

		return nil, err
	}

	return user, nil
}

// getUserByID returns the ent user with the user settings based on the email in the request
func (h *Handler) getUserByID(ctx context.Context, id string) (*ent.User, context.Context, error) {
	user, err := transaction.FromContext(ctx).User.Query().WithSetting().
		Where(user.ID(id)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining user from id")

		return nil, ctx, err
	}

	// set user in the viewer context for the rest of the request
	ctx = setAuthenticatedContext(ctx, user)

	// Add webauthn to the response
	webAuthns, err := user.QueryWebauthns().All(ctx)
	if err != nil {
		return user, ctx, err
	}

	user.Edges.Webauthns = webAuthns

	return user, ctx, nil
}

// addCredentialToUser adds a new webauthn credential to the user
func (h *Handler) addCredentialToUser(ctx context.Context, user *ent.User, credential gowebauthn.Credential) error {
	transports := []string{}
	for _, t := range credential.Transport {
		transports = append(transports, string(t))
	}

	count, err := transaction.FromContext(ctx).Webauthn.Query().Where(
		webauthn.OwnerID(user.ID),
	).Count(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error checking existing webauthn credentials")

		return err
	}

	if count >= h.OauthProvider.Webauthn.MaxDevices {
		logx.FromContext(ctx).Error().Err(err).Msg("max devices reached")

		return ErrMaxDeviceLimit
	}

	err = transaction.FromContext(ctx).Webauthn.Create().
		SetOwnerID(user.ID).
		SetTransports(transports).
		SetAttestationType(credential.AttestationType).
		SetAaguid(models.ToAAGUID(credential.Authenticator.AAGUID)).
		SetCredentialID(credential.ID).
		SetPublicKey(credential.PublicKey).
		SetBackupState(credential.Flags.BackupEligible).
		SetBackupEligible(credential.Flags.BackupEligible).
		SetUserPresent(credential.Flags.UserPresent).
		SetUserVerified(credential.Flags.UserVerified).
		SetSignCount(int32(credential.Authenticator.SignCount)). //nolint:gosec
		Exec(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating passkey")

		return err
	}

	return nil
}

// getUserDetailsByID returns the ent user with the user settings based on the user ID
func (h *Handler) getUserDetailsByID(ctx context.Context, userID string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).User.Query().WithSetting().Where(
		user.ID(userID),
	).Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error retrieving user")

		return nil, err
	}

	return user, nil
}

// getUserTFASettings returns the the user with their tfa settings based on the user ID
func (h *Handler) getUserTFASettings(ctx context.Context, userID string) (*ent.User, error) {
	user, err := transaction.FromContext(ctx).User.Query().Where(
		user.ID(userID),
	).WithTfaSettings().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error retrieving tfa settings for user")

		return nil, err
	}

	return user, nil
}

// updateRecoveryCodes updates the recovery codes for the user in their tfa settings
func (h *Handler) updateRecoveryCodes(ctx context.Context, tfaID string, codes []string) error {
	if err := transaction.FromContext(ctx).TFASetting.UpdateOneID(tfaID).
		SetRecoveryCodes(codes).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating recovery codes")

		return err
	}

	return nil
}

// getUserByInviteToken returns the ent user based on the invite token in the request
func (h *Handler) getUserByInviteToken(ctx context.Context, token string) (*ent.Invite, error) {
	recipient, err := transaction.FromContext(ctx).Invite.Query().
		Where(
			invite.Token(token),
		).WithOwner().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining user from token")

		return nil, err
	}

	return recipient, err
}

// countVerificationTokensUserByEmail counts number of existing email verification attempts before issuing a new one
func (h *Handler) countVerificationTokensUserByEmail(ctx context.Context, email string) (int, error) {
	attempts, err := transaction.FromContext(ctx).EmailVerificationToken.Query().WithOwner().Where(
		emailverificationtoken.And(
			emailverificationtoken.Email(email),
		)).Count(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error counting verification reset tokens")

		return 0, err
	}

	return attempts, nil
}

// expireAllVerificationTokensUserByEmail expires all existing email verification tokens before issuing a new one
func (h *Handler) expireAllVerificationTokensUserByEmail(ctx context.Context, email string) error {
	prs, err := transaction.FromContext(ctx).EmailVerificationToken.Query().WithOwner().Where(
		emailverificationtoken.And(
			emailverificationtoken.Email(email),
			emailverificationtoken.TTLGT(time.Now()),
		)).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining verification reset tokens")

		return err
	}

	for _, pr := range prs {
		if err := pr.Update().SetTTL(time.Now()).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error expiring verification token")

			return err
		}
	}

	return nil
}

// expireAllResetTokensUserByEmail expires all existing password reset tokens before issuing a new one
func (h *Handler) expireAllResetTokensUserByEmail(ctx context.Context, email string) error {
	prs, err := transaction.FromContext(ctx).PasswordResetToken.Query().WithOwner().Where(
		passwordresettoken.And(
			passwordresettoken.Email(email),
			passwordresettoken.TTLGT(time.Now()),
		)).All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining password reset tokens")

		return err
	}

	for _, pr := range prs {
		if err := pr.Update().SetTTL(time.Now()).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error expiring password reset token")

			return err
		}
	}

	return nil
}

// setEmailConfirmed sets the user setting field email_confirmed to true within a transaction
func (h *Handler) setEmailConfirmed(ctx context.Context, user *ent.User) error {
	if err := transaction.FromContext(ctx).UserSetting.
		UpdateOne(user.Edges.Setting).SetEmailConfirmed(true).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error setting email confirmed")

		return err
	}

	return nil
}

// updateUserPassword changes a updates a user's password in the database
func (h *Handler) updateUserPassword(ctx context.Context, id string, password string) error {
	if err := transaction.FromContext(ctx).User.UpdateOneID(id).SetPassword(password).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user password")

		return err
	}

	return nil
}

// CheckAndCreateUser takes a user with an OauthTooToken set in the context and checks if the user is already created
// if the user already exists, update last seen
func (h *Handler) CheckAndCreateUser(ctx context.Context, name, email string, provider enums.AuthProvider, image string) (*ent.User, error) {
	// check if users exists
	entUser, err := h.getUserByEmail(ctx, email)
	if err != nil {
		// if the user is not found, create now
		if ent.IsNotFound(err) {
			// create the input based on the provider
			input := createUserInput(name, email, provider, image)

			// create user in the database
			entUser, err = h.createUser(ctx, input)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error creating new user")

				return nil, err
			}

			// pull latest user settings to ensure any hook updates are included
			entUser.Edges.Setting, err = entUser.QuerySetting().Only(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error fetching user settings")

				return nil, err
			}

			// return newly created user
			return entUser, nil
		}

		return nil, err
	}

	// update last seen of user
	if err := h.updateUserLastSeen(ctx, entUser.ID, provider); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user last seen")

		return nil, err
	}

	// update the return
	entUser.LastLoginProvider = provider

	// update user avatar
	if err := h.updateUserAvatar(ctx, entUser, image); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user avatar")

		return nil, err
	}

	return entUser, nil
}

// createUserInput creates a new user input based on the name, email, image and provider
func createUserInput(name, email string, provider enums.AuthProvider, image string) ent.CreateUserInput {
	lastSeen := time.Now().UTC()

	// create new user input
	input := parseName(name)
	input.Email = email
	input.AuthProvider = &provider
	input.LastLoginProvider = &provider
	input.LastSeen = &lastSeen

	if image != "" {
		input.AvatarRemoteURL = &image
	}

	return input
}

func (h *Handler) updateUserAvatar(ctx context.Context, user *ent.User, image string) error {
	if image == "" {
		return nil
	}

	if user.AvatarRemoteURL != nil && *user.AvatarRemoteURL == image {
		return nil
	}

	if err := transaction.FromContext(ctx).
		User.UpdateOneID(user.ID).
		SetAvatarRemoteURL(image).
		Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error updating user avatar")
		return err
	}

	return nil
}

// setWebauthnAllowed sets the user setting field is_webauthn_allowed to true within a transaction
func (h *Handler) setWebauthnAllowed(ctx context.Context, user *ent.User) error {
	if err := transaction.FromContext(ctx).UserSetting.Update().SetIsWebauthnAllowed(true).
		Where(
			usersetting.UserID(user.ID),
		).Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error setting webauthn allowed")

		return err
	}

	return nil
}

// getSubscriberByToken returns the subscriber based on the token in the request
func (h *Handler) getSubscriberByToken(ctx context.Context, token string) (*ent.Subscriber, error) {
	subscriber, err := transaction.FromContext(ctx).Subscriber.Query().
		Where(
			subscriber.Token(token),
		).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining subscriber from token")

		return nil, err
	}

	return subscriber, nil
}

// getOrgByID returns the organization based on the id in the request
func (h *Handler) getOrgByID(ctx context.Context, id string) (*ent.Organization, error) {
	org, err := transaction.FromContext(ctx).Organization.Get(ctx, id)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error obtaining organization from id")

		return nil, err
	}

	return org, nil
}

// subscriberNotificationOrgName returns the display name identifying the source of a subscriber email:
// the trust center's configured company name, falling back to the owning organization's display name
// when the trust center has no company name set
func (h *Handler) subscriberNotificationOrgName(ctx context.Context, sub *ent.Subscriber) (string, error) {
	if name := h.trustCenterCompanyName(ctx, sub); name != "" {
		return name, nil
	}

	org, err := h.getOrgByID(ctx, sub.OwnerID)
	if err != nil {
		return "", err
	}

	return org.DisplayName, nil
}

// trustCenterCompanyName returns the configured company name for the subscriber's trust center, read
// through the anonymous trust center scope the public site uses: an anon trust center caller plus the
// active trust center id, which the trust center child interceptor scopes the query to. It is best
// effort and returns empty when the subscriber has no trust center, the name is unset, or the read fails
func (h *Handler) trustCenterCompanyName(ctx context.Context, sub *ent.Subscriber) string {
	if sub.TrustCenterID == nil || *sub.TrustCenterID == "" {
		return ""
	}

	tcCtx := auth.WithCaller(ctx, auth.NewTrustCenterBootstrapCaller(sub.OwnerID))
	tcCtx = auth.ActiveTrustCenterIDKey.Set(tcCtx, *sub.TrustCenterID)

	setting, err := transaction.FromContext(tcCtx).TrustCenterSetting.Query().
		Where(
			trustcentersetting.TrustCenterID(*sub.TrustCenterID),
			trustcentersetting.EnvironmentEQ(enums.TrustCenterEnvironmentLive),
		).
		First(tcCtx)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Msg("could not resolve trust center company name for subscriber, falling back to organization name")

		return ""
	}

	return setting.CompanyName
}

// subscriberTrustCenterDomain resolves a subscriber's trust center with its custom domain and live
// setting, read through the anonymous trust center scope, for composing public trust center links
// and the confirmation email branding. ok is false when the subscriber has no trust center or it
// cannot be resolved
func (h *Handler) subscriberTrustCenterDomain(ctx context.Context, sub *ent.Subscriber) (tc *ent.TrustCenter, customDomain string, ok bool) {
	if sub.TrustCenterID == nil || *sub.TrustCenterID == "" {
		return nil, "", false
	}

	tcCtx := auth.WithCaller(ctx, auth.NewTrustCenterBootstrapCaller(sub.OwnerID))
	tcCtx = auth.ActiveTrustCenterIDKey.Set(tcCtx, *sub.TrustCenterID)

	tc, err := transaction.FromContext(tcCtx).TrustCenter.Query().
		Where(trustcenter.ID(*sub.TrustCenterID)).
		WithCustomDomain().
		WithSetting().
		Only(tcCtx)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Msg("could not resolve trust center for subscriber link")

		return nil, "", false
	}

	if tc.Edges.CustomDomain != nil {
		customDomain = tc.Edges.CustomDomain.CnameRecord
	}

	return tc, customDomain, true
}

// subscriberTrustCenterLinks builds the tokenized trust center verify and unsubscribe links for a
// subscriber, so the links land on the trust center rather than the app console, along with the
// trust center branding for the confirmation email; zero values when the trust center cannot resolve
func (h *Handler) subscriberTrustCenterLinks(ctx context.Context, sub *ent.Subscriber, token string) (verifyURL, unsubscribeURL string, branding email.TrustCenterBranding) {
	tc, customDomain, ok := h.subscriberTrustCenterDomain(ctx, sub)
	if !ok {
		return "", "", email.TrustCenterBranding{}
	}

	return trustcenterurl.SubscribeVerifyURLWithToken(customDomain, tc.Slug, token),
		trustcenterurl.UnsubscribeURLWithToken(customDomain, tc.Slug, token),
		email.TrustCenterBrandingFromSetting(tc.Edges.Setting)
}

// createEvent creates a new event in the database but requires mapped input
func (h *Handler) createEvent(ctx context.Context, input ent.CreateEventInput) (*ent.Event, error) {
	event, err := transaction.FromContext(ctx).Event.Create().SetInput(input).Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error creating event")

		return nil, err
	}

	return event, nil
}

// checkForEventID checks if the event ID exists in the database
func (h *Handler) checkForEventID(ctx context.Context, id string) (bool, error) {
	exists, err := transaction.FromContext(ctx).Event.Query().Where(event.EventID(id)).Exist(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error checking for event ID")
		return false, err
	}

	return exists, nil
}

// getOrganizationSettingByOrgID returns the organization setting for a given organization
func (h *Handler) getOrganizationSettingByOrgID(ctx context.Context, orgID string) (*ent.OrganizationSetting, error) {
	setting, err := transaction.FromContext(ctx).OrganizationSetting.Query().
		Where(organizationsetting.OrganizationID(orgID)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error fetching organization settings")

		return nil, err
	}

	return setting, nil
}

// jitProvisionMembership adds an organization membership for a user who successfully authenticated against
// the organization's configured identity provider, when SSO login is enforced and just-in-time provisioning
// is enabled for the organization; existing members are left unchanged
func (h *Handler) jitProvisionMembership(ctx context.Context, orgID string, user *ent.User) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	setting, err := h.getOrganizationSettingByOrgID(allowCtx, orgID)
	if err != nil {
		return err
	}

	if !setting.IdentityProviderLoginEnforced || !setting.IdentityProviderJitProvisioning {
		return nil
	}

	// when an allowlist is configured, only provision users whose authenticated email domain is in it;
	// an empty list provisions any user who authenticates against the identity provider
	if domains := setting.JitAllowedEmailDomains; len(domains) > 0 {
		userDomain := sso.EmailDomain(user.Email)
		if !lo.ContainsBy(domains, func(d string) bool {
			return strings.EqualFold(strings.TrimSpace(d), userDomain)
		}) {
			return nil
		}
	}

	exists, err := transaction.FromContext(ctx).OrgMembership.Query().
		Where(
			orgmembership.UserID(user.ID),
			orgmembership.OrganizationID(orgID),
		).
		Exist(allowCtx)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	// the membership create hooks resolve the organization from the caller, so scope the context to the
	// target org before creating the membership
	memberCtx := auth.WithCaller(allowCtx, &auth.Caller{
		SubjectID:       user.ID,
		SubjectEmail:    user.Email,
		OrganizationID:  orgID,
		OrganizationIDs: []string{orgID},
	})

	role := enums.RoleMember

	return transaction.FromContext(ctx).OrgMembership.Create().
		SetInput(ent.CreateOrgMembershipInput{
			OrganizationID: orgID,
			UserID:         user.ID,
			Role:           &role,
		}).
		Exec(memberCtx)
}

// getUserDefaultOrgID returns the default organization ID for a user
func (h *Handler) getUserDefaultOrgID(ctx context.Context, userID string) (string, error) {
	us, err := transaction.FromContext(ctx).UserSetting.Query().Where(usersetting.UserID(userID)).WithDefaultOrg().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error fetching user settings")

		return "", err
	}

	if us.Edges.DefaultOrg == nil {
		return "", fmt.Errorf("%w: default org", ErrMissingField)
	}

	return us.Edges.DefaultOrg.ID, nil
}
