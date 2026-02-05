package hooks

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"entgo.io/ent"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/jobs"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/pkg/domain"
	"github.com/theopenlane/core/pkg/logx"
)

// HookTrustCenterNDARequestCreate handles new NDA request creation
func HookTrustCenterNDARequestCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterNDARequestFunc(func(ctx context.Context, m *generated.TrustCenterNDARequestMutation) (generated.Value, error) {
			trustCenterID, ok := m.TrustCenterID()
			if !ok || trustCenterID == "" {
				logx.FromContext(ctx).Error().Msg("trust center ID is required for NDA request")

				return nil, ErrTrustCenterIDRequired
			}

			n, err := m.Client().Template.Query().
				Select(template.FieldID).
				Where(template.KindEQ(enums.TemplateKindTrustCenterNda)).
				Count(ctx)
			if err != nil {
				return nil, err
			}

			if n == 0 {
				return nil, ErrNDATemplateRequired
			}

			email, _ := m.Email()

			queryCtx := ctx
			if _, isAnon := auth.AnonymousTrustCenterUserFromContext(ctx); isAnon {
				queryCtx = privacy.DecisionContext(ctx, privacy.Allow)
			}

			existingRequest, err := m.Client().TrustCenterNDARequest.Query().
				Where(
					trustcenterndarequest.TrustCenterIDEQ(trustCenterID),
					trustcenterndarequest.EmailEQ(email),
				).
				Only(queryCtx)
			if err != nil && !generated.IsNotFound(err) {
				return nil, err
			}

			if existingRequest != nil {
				return handleExistingNDARequest(ctx, queryCtx, m, existingRequest)
			}

			tc, err := m.Client().TrustCenter.Query().
				Where(trustcenter.IDEQ(trustCenterID)).
				WithSetting().
				Only(queryCtx)
			if err != nil {
				return nil, err
			}

			requiresApproval := tc.Edges.Setting != nil && tc.Edges.Setting.NdaApprovalRequired
			if requiresApproval {
				m.SetStatus(enums.TrustCenterNDARequestStatusNeedsApproval)
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			request, ok := v.(*generated.TrustCenterNDARequest)
			if !ok {
				return v, nil
			}

			if requiresApproval {
				if err := createNDARequestNotification(ctx, m, request, tc.OwnerID); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
				}
				return v, nil
			}

			if err := sendTrustCenterNDARequestEmail(ctx, ndaAuthEmailData{
				email:         request.Email,
				trustCenterID: request.TrustCenterID,
			}); err != nil {
				return nil, err
			}

			return v, nil
		})
	}, ent.OpCreate)
}

func handleExistingNDARequest(ctx, queryCtx context.Context, m *generated.TrustCenterNDARequestMutation, existing *generated.TrustCenterNDARequest) (*generated.TrustCenterNDARequest, error) {
	switch existing.Status {
	case enums.TrustCenterNDARequestStatusSigned:
		// if already signed, resend auth email
		if err := sendTrustCenterAuthEmail(ctx, ndaAuthEmailData{
			email:         existing.Email,
			trustCenterID: existing.TrustCenterID,
		}); err != nil {
			return nil, err
		}

		return existing, nil
	case enums.TrustCenterNDARequestStatusApproved, enums.TrustCenterNDARequestStatusRequested:
		// if its approved, or requested (no authorization required), resend NDA email
		if err := sendTrustCenterNDARequestEmail(ctx, ndaAuthEmailData{
			email:         existing.Email,
			trustCenterID: existing.TrustCenterID,
		}); err != nil {
			return nil, err
		}

		return existing, nil
	case enums.TrustCenterNDARequestStatusNeedsApproval:
		// if needs approval, recreate notification
		tc, err := getTrustCenter(ctx, m.Client(), existing.TrustCenterID)
		if err != nil {
			return nil, err
		}

		if err := createNDARequestNotification(ctx, m, existing, tc.OwnerID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
		}

		return existing, nil
	case enums.TrustCenterNDARequestStatusDeclined:
		// if previously declined, set to needs approval again to restart the process
		if err := m.Client().TrustCenterNDARequest.UpdateOne(existing).SetStatus(enums.TrustCenterNDARequestStatusNeedsApproval).
			Exec(queryCtx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("email", existing.Email).Msg("failed to update NDA request status to needs approval")
			return nil, err
		}

		tc, err := getTrustCenter(ctx, m.Client(), existing.TrustCenterID)
		if err != nil {
			return nil, err
		}

		if err := createNDARequestNotification(ctx, m, existing, tc.OwnerID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
		}
	}

	// otherwise do nothing invalid
	return existing, nil
}

// getTrustCenter is a helper to get the trust center for an NDA request by ID
func getTrustCenter(ctx context.Context, client *generated.Client, trustCenterID string) (*generated.TrustCenter, error) {
	tc, err := client.TrustCenter.Query().
		Where(trustcenter.IDEQ(trustCenterID)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to get trust center for existing NDA request")
		return nil, err
	}

	return tc, nil
}

// HookTrustCenterNDARequestUpdate handles NDA request status updates - sends email when approved
func HookTrustCenterNDARequestUpdate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterNDARequestFunc(func(ctx context.Context, m *generated.TrustCenterNDARequestMutation) (generated.Value, error) {
			status, ok := m.Status()

			// on update one, check if status is set, if not get old status
			if m.Op().Is(ent.OpUpdateOne) && (!ok || status == "") {
				oldStatus, err := m.OldStatus(ctx)

				// if status isn't set on mutation, set to the old status
				if err == nil && status == "" {
					status = oldStatus
				}
			}

			if !ok || (status != enums.TrustCenterNDARequestStatusApproved && status != enums.TrustCenterNDARequestStatusSigned) {
				return next.Mutate(ctx, m)
			}

			// if approved or signed, set the timestamp in the ISO8601 format
			now, err := models.ToDateTime(time.Now().UTC().Format(time.RFC3339))
			if err != nil {
				return nil, err
			}

			if status == enums.TrustCenterNDARequestStatusSigned {
				m.SetSignedAt(*now)

				retVal, err := next.Mutate(ctx, m)
				if err != nil {
					return nil, err
				}

				email := ""
				trustCenterID := ""

				// if this is a trust center nda request, we can get the email and trust center ID from the request
				request, ok := retVal.(*generated.TrustCenterNDARequest)
				if ok {
					email = request.Email
					trustCenterID = request.TrustCenterID
				} else {
					// otherwise its a document data mutation, get from the anonymous trust center user context
					anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx)
					if ok {
						email = anon.SubjectEmail
						trustCenterID = anon.TrustCenterID
					}
				}

				if email == "" || trustCenterID == "" {
					logx.FromContext(ctx).Error().Msg("missing email or trust center ID to send auth email upon NDA signing")
					return retVal, nil
				}

				// send auth email upon signing
				if err := sendTrustCenterAuthEmail(ctx, ndaAuthEmailData{
					email:         email,
					trustCenterID: trustCenterID,
				}); err != nil {
					return nil, err
				}

				return retVal, nil
			}

			m.SetApprovedAt(*now)

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			request, ok := v.(*generated.TrustCenterNDARequest)
			if !ok {
				return v, nil
			}

			if err := sendTrustCenterNDARequestEmail(ctx, ndaAuthEmailData{
				email:         request.Email,
				trustCenterID: request.TrustCenterID,
			}); err != nil {
				return nil, err
			}

			return v, nil
		})
	}, ent.OpUpdateOne|ent.OpUpdate)
}

func createNDARequestNotification(ctx context.Context, m *generated.TrustCenterNDARequestMutation, ndaRequest *generated.TrustCenterNDARequest, ownerID string) error {
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
		Data: map[string]interface{}{
			"nda_request_id":  ndaRequest.ID,
			"trust_center_id": ndaRequest.TrustCenterID,
			"email":           ndaRequest.Email,
		},
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	_, err := m.Client().Notification.Create().SetInput(input).Save(allowCtx)

	return err
}

type ndaAuthEmailData struct {
	email         string
	trustCenterID string
}

func sendTrustCenterNDARequestEmail(ctx context.Context, ndaRequest ndaAuthEmailData) error {
	if ndaRequest.trustCenterID == "" || ndaRequest.email == "" {
		logx.FromContext(ctx).Info().Msg("missing trust center ID or email for auth email")
		return nil
	}

	tc, err := transactionFromContext(ctx).TrustCenter.Query().
		Where(trustcenter.IDEQ(ndaRequest.trustCenterID)).
		WithCustomDomain().
		WithSetting().
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get trust center for NDA email")

		return err
	}

	accessToken, err := generateTrustCenterJWT(ctx, tc, ndaRequest.email)
	if err != nil {
		return err
	}

	trustCenterURL := getTrustCenterEmailURL(ctx, buildNDATrustCenterURL(tc))

	emailMsg, err := transactionFromContext(ctx).Emailer.NewTrustCenterNDARequestEmail(emailtemplates.Recipient{
		Email: ndaRequest.email,
	}, accessToken, emailtemplates.TrustCenterNDARequestData{
		OrganizationName: tc.Edges.Setting.CompanyName,
		TrustCenterURL:   trustCenterURL,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA email")

		return err
	}

	if _, err = transactionFromContext(ctx).Job.Insert(ctx, jobs.EmailArgs{Message: *emailMsg}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to queue NDA email")

		return err
	}

	return nil
}

func sendTrustCenterAuthEmail(ctx context.Context, ndaRequest ndaAuthEmailData) error {
	if ndaRequest.trustCenterID == "" || ndaRequest.email == "" {
		logx.FromContext(ctx).Info().Msg("missing trust center ID or email for auth email")

		return nil
	}

	tc, err := transactionFromContext(ctx).TrustCenter.Query().
		Where(trustcenter.IDEQ(ndaRequest.trustCenterID)).
		WithCustomDomain().
		WithSetting().
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get trust center for auth email")
		return err
	}

	accessToken, err := generateTrustCenterJWT(ctx, tc, ndaRequest.email)
	if err != nil {
		return err
	}

	trustCenterURL := getTrustCenterEmailURL(ctx, getTrustCenterBaseURL(tc))

	emailMsg, err := transactionFromContext(ctx).Emailer.NewTrustCenterAuthEmail(emailtemplates.Recipient{
		Email: ndaRequest.email,
	}, accessToken, emailtemplates.TrustCenterAuthData{
		OrganizationName: tc.Edges.Setting.CompanyName,
		TrustCenterURL:   trustCenterURL,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create auth email")
		return err
	}

	if _, err = transactionFromContext(ctx).Job.Insert(ctx, jobs.EmailArgs{Message: *emailMsg}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to queue auth email")
		return err
	}

	return nil
}

func generateTrustCenterJWT(ctx context.Context, tc *generated.TrustCenter, email string) (string, error) {
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, uuid.New().String())

	duration := tc.TokenManager.Config().TrustCenterNDARequestAccessDuration

	accessToken, _, err := transactionFromContext(ctx).TokenManager.CreateTokenPair(&tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:        anonUserID,
		OrgID:         tc.OwnerID,
		TrustCenterID: tc.ID,
		Email:         email,
	}, tokens.WithAccessDuration(duration))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create token for auth email")

		return "", err
	}

	return accessToken, nil
}

func getTrustCenterEmailURL(ctx context.Context, u url.URL) string {
	trustCenterURLStr := u.String()

	if transactionFromContext(ctx).Shortlinks == nil {
		return trustCenterURLStr
	}

	shortenedURL, shortenErr := transactionFromContext(ctx).Shortlinks.Create(ctx, trustCenterURLStr, "")
	if shortenErr != nil {
		logx.FromContext(ctx).Error().Err(shortenErr).Msg("failed to shorten trust center URL, using original")

		return trustCenterURLStr
	}

	return shortenedURL
}

func buildNDATrustCenterURL(tc *generated.TrustCenter) url.URL {
	const ndaPath = "/access/sign-nda"

	trustCenterURL := getTrustCenterBaseURL(tc)
	trustCenterURL.Path = "/" + tc.Slug + ndaPath

	return trustCenterURL
}

// getTrustCenterBaseURL builds the base URL for a trust center
func getTrustCenterBaseURL(tc *generated.TrustCenter) url.URL {
	trustCenterURL := url.URL{Scheme: "https"}

	if tc.Edges.CustomDomain != nil {
		customHost := tc.Edges.CustomDomain.CnameRecord
		if normalized, err := domain.NormalizeHostname(customHost); err == nil {
			customHost = normalized
		}
		trustCenterURL.Host = customHost

		return trustCenterURL
	}

	defaultHost := trustCenterConfig.DefaultTrustCenterDomain
	if normalized, err := domain.NormalizeHostname(defaultHost); err == nil {
		defaultHost = normalized
	}
	trustCenterURL.Host = defaultHost

	return trustCenterURL
}
