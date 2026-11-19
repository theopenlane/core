package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/logx"

	models "github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
	"github.com/theopenlane/core/v2/internal/ent/generated/sharelink"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/email"
	authmw "github.com/theopenlane/core/v2/pkg/middleware/auth"
	"github.com/theopenlane/core/v2/pkg/middleware/ratelimit"
	"github.com/theopenlane/core/v2/pkg/middleware/transaction"
	"github.com/theopenlane/core/v2/pkg/objects/storage/proxy"
)

const (
	// shareLinkCodeSpace is the number of distinct six digit verification codes
	shareLinkCodeSpace = 1000000
	// shareLinkCodeDigits is the number of digits in a verification code
	shareLinkCodeDigits = 6
	// shareLinkCodeSeparator separates the recipient hash from the code in the digest stored at rest
	shareLinkCodeSeparator = ":"
	// shareLinkCodeDuration is how long a verification code stays valid
	shareLinkCodeDuration = 10 * time.Minute
	// shareLinkCodeMaxAttempts is the number of failed verifications after which the code is invalidated
	shareLinkCodeMaxAttempts = 5
)

// forwarderSecretPresented reports whether the request carries the configured forwarder secret; when it does not the
// response has already been written: 404 when no secret is configured so the endpoint is effectively disabled, and
// 403 when the presented secret does not match
func (h *Handler) forwarderSecretPresented(ctx echo.Context) (bool, error) {
	if h.ForwarderSecret == "" {
		return false, h.NotFound(ctx, ErrNotFound)
	}

	presented := ctx.Request().Header.Get(ratelimit.ForwarderSecretHeader)
	if subtle.ConstantTimeCompare([]byte(presented), []byte(h.ForwarderSecret)) != 1 {
		logx.FromContext(ctx.Request().Context()).Warn().Msg("forwarder secret mismatch")

		return false, h.Forbidden(ctx, ErrUnauthorized)
	}

	return true, nil
}

// shareLinkOrgContext returns the context a located share link is read and written under: an internal caller scoped
// to the organization owning the share, so org-scoped mutation filters and the file organization filter apply to it
func shareLinkOrgContext(ctx context.Context, link *ent.ShareLink) context.Context {
	return auth.WithOrgInternalCaller(ctx, link.OwnerID)
}

// shareLinkLive reports whether a share link is neither revoked nor expired at the given time
func shareLinkLive(link *ent.ShareLink, now time.Time) bool {
	return link.RevokedAt == nil && now.Before(link.ExpiresAt)
}

// verifiedShareLink loads the share link within the organization named by the presented share token and reports
// whether the token is the one minted for it; an unknown share and a token that does not verify are both reported as
// not found
func verifiedShareLink(ctx context.Context, links *ent.ShareLinkClient, id, presented string) (*ent.ShareLink, bool, error) {
	orgID, signature, ok := proxy.SplitShareToken(presented)
	if !ok {
		return nil, false, nil
	}

	tokenOrgID, err := ulid.Parse(orgID)
	if err != nil {
		return nil, false, nil
	}

	link, err := links.Get(auth.WithOrgInternalCaller(ctx, orgID), id)

	switch {
	case ent.IsNotFound(err):
		return nil, false, nil
	case err != nil:
		return nil, false, err
	}

	token := tokens.ShareToken{ShareID: link.ID, OrgID: tokenOrgID, SigningInfo: tokens.SigningInfo{ExpiresAt: link.ExpiresAt}}

	err = token.Verify(signature, link.Secret)

	switch {
	case errors.Is(err, tokens.ErrTokenExpired):
		// an expired token is refused before its signature is checked, so it is matched against the stored token
		return link, subtle.ConstantTimeCompare([]byte(presented), []byte(link.Token)) == 1, nil
	default:
		return link, err == nil, nil
	}
}

// recipientMatch returns 1 when the email names one of the recipients the share link is bound to and 0 otherwise
func recipientMatch(link *ent.ShareLink, presented string) int {
	return recipientHashMatch(link, proxy.RecipientHash(presented))
}

// recipientHashMatch returns 1 when the hash is that of one of the recipients the share link is bound to and 0
// otherwise; every recipient is compared so the time taken does not reveal which one matched
func recipientHashMatch(link *ent.ShareLink, presentedHash string) int {
	return lo.Reduce(link.RecipientEmails, func(match int, recipient string, _ int) int {
		return match | subtle.ConstantTimeCompare([]byte(presentedHash), []byte(proxy.RecipientHash(recipient)))
	}, 0)
}

// shareLinkCodeHash returns the digest stored at rest for a verification code; it covers the recipient the code was
// issued to, so a code only verifies together with that address
func shareLinkCodeHash(code, recipientHash string) []byte {
	sum := sha256.Sum256([]byte(recipientHash + shareLinkCodeSeparator + code))

	return sum[:]
}

// newShareLinkCode generates a six digit verification code
func newShareLinkCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(shareLinkCodeSpace))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", shareLinkCodeDigits, n.Int64()), nil
}

// shareLinkCodePending reports whether the share link holds a verification code that has not yet expired
func shareLinkCodePending(link *ent.ShareLink, now time.Time) bool {
	return len(link.CodeHash) > 0 && link.CodeExpiresAt != nil && now.Before(*link.CodeExpiresAt)
}

// shareLinkSharingEnabled reports whether the organization owning the share link still allows external sharing
func shareLinkSharingEnabled(ctx context.Context, settings *ent.OrganizationSettingClient, link *ent.ShareLink) (bool, error) {
	setting, err := settings.Query().
		Where(organizationsetting.OrganizationID(link.OwnerID)).
		Select(organizationsetting.FieldExternalSharingEnabled).
		Only(ctx)
	if err != nil {
		return false, err
	}

	return setting.ExternalSharingEnabled, nil
}

// ShareLinkResolveHandler describes a share link to the file worker once the share token verifies: 404 for an unknown
// share or token, 410 with the reason when the share is revoked, expired, or exhausted; it never counts a use, and lists
// the files of a recipient-bound share only when the request carries its share-verified token
func (h *Handler) ShareLinkResolveHandler(ctx echo.Context) error {
	if ok, err := h.forwarderSecretPresented(ctx); !ok {
		return err
	}

	in, err := BindAndValidate[models.ShareLinkResolveRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := logx.WithField(ctx.Request().Context(), "share_link_id", in.ID)
	client := transaction.FromContext(reqCtx)

	link, found, err := verifiedShareLink(reqCtx, client.ShareLink, in.ID, in.Token)

	switch {
	case err != nil:
		logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	case !found:
		return h.NotFound(ctx, ErrNotFound)
	}

	orgCtx := shareLinkOrgContext(reqCtx, link)

	sharingEnabled, err := shareLinkSharingEnabled(orgCtx, client.OrganizationSetting, link)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link organization setting")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if reason := shareLinkGoneReason(link, sharingEnabled, time.Now()); reason != "" {
		return ctx.JSON(http.StatusGone, models.ShareLinkGoneResponse{Reason: reason, LandingPage: link.LandingPage})
	}

	recipientBound := len(link.RecipientEmails) > 0

	var files []*ent.File

	// the files of a recipient-bound share are listed only once the recipient has verified
	if !recipientBound || h.shareLinkRecipientVerified(link, in.VerifiedToken) {
		files, err = link.QueryFiles().All(orgCtx)
		if err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link files")

			return h.InternalServerError(ctx, ErrProcessingRequest)
		}
	}

	return h.Success(ctx, models.ShareLinkResolveResponse{
		Name:           link.Name,
		Organization:   shareLinkOrgName(orgCtx, link),
		ExpiresAt:      link.ExpiresAt.UTC(),
		LandingPage:    link.LandingPage,
		RecipientBound: recipientBound,
		Files: lo.Map(files, func(f *ent.File, _ int) models.ShareLinkResolveFile {
			return models.ShareLinkResolveFile{ID: f.ID, Name: f.ProvidedFileName, ContentType: f.DetectedContentType}
		}),
	})
}

// shareLinkGoneReason returns why a share link can no longer be used, or an empty string while it is usable; a link
// whose organization has turned external sharing off is reported as revoked
func shareLinkGoneReason(link *ent.ShareLink, sharingEnabled bool, now time.Time) string {
	switch {
	case link.RevokedAt != nil || !sharingEnabled:
		return models.ShareLinkConsumeReasonRevoked
	case !now.Before(link.ExpiresAt):
		return models.ShareLinkConsumeReasonExpired
	case link.MaxUses != proxy.UnlimitedShareUses && link.UseCount >= link.MaxUses:
		return models.ShareLinkConsumeReasonExhausted
	default:
		return ""
	}
}

// ShareLinkCodeHandler issues a verification code to the recipient of a share link; once the request is validated it
// always responds 204 so a caller cannot learn whether the share exists or whom it is bound to, and only emails the code
// when the share is live, its organization still allows external sharing, no earlier code is still valid, and the email
// names its recipient
func (h *Handler) ShareLinkCodeHandler(ctx echo.Context) error {
	if ok, err := h.forwarderSecretPresented(ctx); !ok {
		return err
	}

	in, err := BindAndValidate[models.ShareLinkCodeRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := logx.WithField(ctx.Request().Context(), "share_link_id", in.ID)
	client := transaction.FromContext(reqCtx)

	link, found, err := verifiedShareLink(reqCtx, client.ShareLink, in.ID, in.Token)

	switch {
	case err != nil:
		logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	case !found:
		return ctx.NoContent(http.StatusNoContent)
	}

	now := time.Now()

	// one code is issued per validity window, so reissuing cannot reset the attempt cap or flood the recipient
	if !shareLinkLive(link, now) || shareLinkCodePending(link, now) || recipientMatch(link, in.Email) != 1 {
		return ctx.NoContent(http.StatusNoContent)
	}

	orgCtx := shareLinkOrgContext(reqCtx, link)

	sharingEnabled, err := shareLinkSharingEnabled(orgCtx, client.OrganizationSetting, link)

	switch {
	case err != nil:
		logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link organization setting")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	case !sharingEnabled:
		return ctx.NoContent(http.StatusNoContent)
	}

	code, err := newShareLinkCode()
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error generating share link code")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if err := client.ShareLink.UpdateOneID(link.ID).
		SetCodeHash(shareLinkCodeHash(code, proxy.RecipientHash(in.Email))).
		SetCodeExpiresAt(now.Add(shareLinkCodeDuration)).
		SetCodeAttempts(0).
		Exec(orgCtx); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error storing share link code")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if err := h.sendEmail(orgCtx, email.ShareLinkCodeOp.Name(), email.ShareLinkCodeEmail{
		Email:            proxy.NormalizeRecipientEmail(in.Email),
		OrgName:          shareLinkOrgName(orgCtx, link),
		ShareName:        link.Name,
		Code:             code,
		ExpiresInMinutes: int(shareLinkCodeDuration.Minutes()),
	}); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error sending share link code email")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// shareLinkOrgName returns the display name of the organization owning the share link, or an empty string when it
// cannot be resolved so the email falls back to a generic sender
func shareLinkOrgName(ctx context.Context, link *ent.ShareLink) string {
	org, err := link.QueryOwner().Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("error resolving share link organization")

		return ""
	}

	if org.DisplayName != "" {
		return org.DisplayName
	}

	return org.Name
}

// ShareLinkVerifyHandler checks a verification code against the share link's current code and returns a
// share-verified token on a match; every failure counts against the code and after the fifth it can no longer verify
func (h *Handler) ShareLinkVerifyHandler(ctx echo.Context) error {
	if ok, err := h.forwarderSecretPresented(ctx); !ok {
		return err
	}

	in, err := BindAndValidate[models.ShareLinkVerifyRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := logx.WithField(ctx.Request().Context(), "share_link_id", in.ID)
	client := transaction.FromContext(reqCtx)

	link, found, err := verifiedShareLink(reqCtx, client.ShareLink, in.ID, in.Token)

	switch {
	case err != nil:
		logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	case !found:
		return h.Forbidden(ctx, ErrShareLinkVerificationFailed)
	}

	now := time.Now()

	if !shareLinkCodeVerifiable(link, now) {
		return h.Forbidden(ctx, ErrShareLinkVerificationFailed)
	}

	orgCtx := shareLinkOrgContext(reqCtx, link)

	sharingEnabled, err := shareLinkSharingEnabled(orgCtx, client.OrganizationSetting, link)

	switch {
	case err != nil:
		logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link organization setting")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	case !sharingEnabled:
		return h.Forbidden(ctx, ErrShareLinkVerificationFailed)
	}

	claimed, err := claimShareLinkCodeAttempt(orgCtx, client.ShareLink, link.ID)

	switch {
	case err != nil:
		logx.FromContext(reqCtx).Error().Err(err).Msg("error claiming share link code attempt")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	case !claimed:
		return h.Forbidden(ctx, ErrShareLinkVerificationFailed)
	}

	emailMatch := recipientMatch(link, in.Email)
	codeMatch := subtle.ConstantTimeCompare(shareLinkCodeHash(in.Code, proxy.RecipientHash(in.Email)), link.CodeHash)

	// a code that has used up its attempts stays in place until it expires, so no new code can be issued to reset the cap
	if emailMatch&codeMatch != 1 {
		return h.Forbidden(ctx, ErrShareLinkVerificationFailed)
	}

	if err := client.ShareLink.UpdateOneID(link.ID).
		ClearCodeHash().
		ClearCodeExpiresAt().
		SetCodeAttempts(0).
		Exec(orgCtx); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error clearing share link code")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	token, err := proxy.NewShareVerifiedToken(h.TokenManager, h.TokenManager.Config().Issuer, link.ID, proxy.RecipientHash(in.Email))
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error minting share verified token")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.Success(ctx, models.ShareLinkVerifyResponse{Token: token})
}

// shareLinkCodeVerifiable reports whether the share link is live, bound to a recipient, and holds an unexpired code
// that has not exhausted its attempts
func shareLinkCodeVerifiable(link *ent.ShareLink, now time.Time) bool {
	return shareLinkLive(link, now) &&
		len(link.RecipientEmails) > 0 &&
		shareLinkCodePending(link, now) &&
		link.CodeAttempts < shareLinkCodeMaxAttempts
}

// claimShareLinkCodeAttempt counts one verification attempt against the current code in a single conditional update,
// so concurrent guesses cannot exceed the attempt cap, and reports whether an attempt was still available
func claimShareLinkCodeAttempt(ctx context.Context, links *ent.ShareLinkClient, id string) (bool, error) {
	claimed, err := links.Update().
		Where(sharelink.ID(id), sharelink.CodeAttemptsLT(shareLinkCodeMaxAttempts)).
		AddCodeAttempts(1).
		Save(ctx)

	return claimed > 0, err
}

// useCountBelowMaxUses matches share links that still have downloads left
func useCountBelowMaxUses() predicate.ShareLink {
	return func(s *sql.Selector) {
		s.Where(sql.ColumnsLT(s.C(sharelink.FieldUseCount), s.C(sharelink.FieldMaxUses)))
	}
}

// ShareLinkConsumeHandler records a download through a share link; recipient verification and the session binding are
// checked first, then the check and the use count increment are a single conditional update so concurrent downloads
// cannot both pass the use cap, and one access event row is written once the download is allowed
func (h *Handler) ShareLinkConsumeHandler(ctx echo.Context) error {
	if ok, err := h.forwarderSecretPresented(ctx); !ok {
		return err
	}

	in, err := BindAndValidate[models.ShareLinkConsumeRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := logx.WithField(ctx.Request().Context(), "share_link_id", in.ID)
	client := transaction.FromContext(reqCtx)

	link, found, err := verifiedShareLink(reqCtx, client.ShareLink, in.ID, in.Token)

	switch {
	case err != nil:
		logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	case !found:
		return h.NotFound(ctx, ErrNotFound)
	}

	if len(link.RecipientEmails) > 0 && !h.shareLinkRecipientVerified(link, in.VerifiedToken) {
		return h.Success(ctx, models.ShareLinkConsumeResponse{Allowed: false, Reason: models.ShareLinkConsumeReasonVerificationRequired})
	}

	if link.RequireSession {
		reason, err := h.shareLinkSessionDenyReason(ctx, client.File, link, in.FileID)

		switch {
		case err != nil:
			logx.FromContext(reqCtx).Error().Err(err).Msg("error authenticating share link session")

			return h.InternalServerError(ctx, ErrProcessingRequest)
		case reason != "":
			return h.Success(ctx, models.ShareLinkConsumeResponse{Allowed: false, Reason: reason})
		}
	}

	orgCtx := shareLinkOrgContext(reqCtx, link)
	now := time.Now()

	sharingEnabled, err := shareLinkSharingEnabled(orgCtx, client.OrganizationSetting, link)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link organization setting")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	// the file is loaded before a use is counted, so a file that is deleted or outside the share never costs one
	served, err := link.QueryFiles().Where(file.ID(in.FileID)).Only(orgCtx)

	switch {
	case ent.IsNotFound(err):
		return h.Success(ctx, models.ShareLinkConsumeResponse{Allowed: false, Reason: shareLinkDenyReason(link, sharingEnabled, false, now)})
	case err != nil:
		logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link file")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if !sharingEnabled {
		return h.Success(ctx, models.ShareLinkConsumeResponse{Allowed: false, Reason: models.ShareLinkConsumeReasonRevoked})
	}

	consumed, err := client.ShareLink.Update().
		Where(
			sharelink.ID(link.ID),
			sharelink.RevokedAtIsNil(),
			sharelink.ExpiresAtGT(now),
			sharelink.Or(sharelink.MaxUses(proxy.UnlimitedShareUses), sharelink.MaxUsesIsNil(), useCountBelowMaxUses()),
		).
		AddUseCount(1).
		Save(orgCtx)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error consuming share link")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	if consumed == 0 {
		link, err = client.ShareLink.Get(orgCtx, link.ID)
		if err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("error loading share link")

			return h.InternalServerError(ctx, ErrProcessingRequest)
		}

		return h.Success(ctx, models.ShareLinkConsumeResponse{Allowed: false, Reason: shareLinkDenyReason(link, sharingEnabled, true, now)})
	}

	// the error is returned so the transaction rolls back and the use is not counted without its access event
	if err := recordShareLinkAccessEvent(orgCtx, client.ShareLinkAccessEvent, link, in); err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Msg("error recording share link access event")

		return err
	}

	return h.Success(ctx, models.ShareLinkConsumeResponse{
		Allowed: true,
		File:    &models.ShareLinkConsumeFile{Key: served.StoragePath, ContentType: served.DetectedContentType, FileName: served.ProvidedFileName},
	})
}

// shareLinkRecipientVerified reports whether the share-verified token was minted for this share link and one of its
// current recipients
func (h *Handler) shareLinkRecipientVerified(link *ent.ShareLink, verifiedToken string) bool {
	shareID, recipientHash, err := proxy.VerifyShareVerifiedToken(h.TokenManager, h.TokenManager.Config().Issuer, verifiedToken)
	if err != nil {
		return false
	}

	return shareID == link.ID && recipientHashMatch(link, recipientHash) == 1
}

// shareLinkSessionDenyReason authenticates the credential forwarded with the request and returns why it may not
// download the file, or an empty string when it is the share link creator and can still read the file
func (h *Handler) shareLinkSessionDenyReason(ctx echo.Context, files *ent.FileClient, link *ent.ShareLink, fileID string) (string, error) {
	validator, err := h.AuthOptions.Validator()
	if err != nil {
		return "", err
	}

	if err := authmw.AuthenticateRequest(ctx, h.AuthOptions, validator); err != nil {
		logx.FromContext(ctx.Request().Context()).Debug().Err(err).Str("share_link_id", link.ID).Msg("share link session credential rejected")

		return models.ShareLinkConsumeReasonSessionRequired, nil
	}

	// the file read runs as the authenticated caller so its normal privacy rules decide
	callerCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(callerCtx)
	if !ok || caller.SubjectID != link.CreatedBy {
		return models.ShareLinkConsumeReasonForbidden, nil
	}

	// the file is loaded, not just checked for existence, so the caller's access to it is evaluated
	if _, err := files.Query().Where(file.ID(fileID)).Only(callerCtx); err != nil {
		logx.FromContext(callerCtx).Debug().Err(err).Str("share_link_id", link.ID).Msg("share link creator file read refused")

		return models.ShareLinkConsumeReasonForbidden, nil
	}

	return "", nil
}

// recordShareLinkAccessEvent writes the access event row for the download that was just allowed on the share link
func recordShareLinkAccessEvent(ctx context.Context, events *ent.ShareLinkAccessEventClient, link *ent.ShareLink, in *models.ShareLinkConsumeRequest) error {
	return events.Create().
		SetShareLinkID(link.ID).
		SetOwnerID(link.OwnerID).
		SetFileID(in.FileID).
		SetIPAddress(in.IP).
		SetUserAgent(in.UserAgent).
		Exec(ctx)
}

// shareLinkDenyReason explains why a download through the share link is refused, checked in the order the worker
// reports them: revoked, expired, a file outside the share, then no downloads left
func shareLinkDenyReason(link *ent.ShareLink, sharingEnabled, shared bool, now time.Time) string {
	switch {
	case link.RevokedAt != nil || !sharingEnabled:
		return models.ShareLinkConsumeReasonRevoked
	case !now.Before(link.ExpiresAt):
		return models.ShareLinkConsumeReasonExpired
	case !shared:
		return models.ShareLinkConsumeReasonUnknownFile
	default:
		return models.ShareLinkConsumeReasonExhausted
	}
}
