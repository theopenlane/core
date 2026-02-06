package hooks

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"entgo.io/ent"
	"github.com/golang-jwt/jwt/v5"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
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
				return handleExistingNDARequest(ctx, queryCtx, existingRequest)
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
				if err := createNDARequestNotification(ctx, request, tc.OwnerID); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
				}
				return v, nil
			}

			if err := sendTrustCenterNDARequestEmail(ctx, ndaAuthEmailData{
				requestID:     request.ID,
				email:         request.Email,
				trustCenterID: request.TrustCenterID,
			}); err != nil {
				return nil, err
			}

			return v, nil
		})
	}, ent.OpCreate)
}

func handleExistingNDARequest(ctx, queryCtx context.Context, existing *generated.TrustCenterNDARequest) (*generated.TrustCenterNDARequest, error) {
	switch existing.Status {
	case enums.TrustCenterNDARequestStatusSigned:
		// if already signed, resend auth email
		if err := sendTrustCenterAuthEmail(ctx, ndaAuthEmailData{
			requestID:     existing.ID,
			email:         existing.Email,
			trustCenterID: existing.TrustCenterID,
		}); err != nil {
			return nil, err
		}

		return existing, nil
	case enums.TrustCenterNDARequestStatusApproved, enums.TrustCenterNDARequestStatusRequested:
		// if its approved, or requested (no authorization required), resend NDA email
		if err := sendTrustCenterNDARequestEmail(ctx, ndaAuthEmailData{
			requestID:     existing.ID,
			email:         existing.Email,
			trustCenterID: existing.TrustCenterID,
		}); err != nil {
			return nil, err
		}

		return existing, nil
	case enums.TrustCenterNDARequestStatusNeedsApproval:
		// if needs approval, recreate notification
		tc, err := getTrustCenter(ctx, existing.TrustCenterID)
		if err != nil {
			return nil, err
		}

		if err := createNDARequestNotification(ctx, existing, tc.OwnerID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
		}

		return existing, nil
	case enums.TrustCenterNDARequestStatusDeclined:
		// if previously declined, set to needs approval again to restart the process
		if err := transactionFromContext(ctx).TrustCenterNDARequest.UpdateOne(existing).SetStatus(enums.TrustCenterNDARequestStatusNeedsApproval).
			Exec(queryCtx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("email", existing.Email).Msg("failed to update NDA request status to needs approval")
			return nil, err
		}

		tc, err := getTrustCenter(ctx, existing.TrustCenterID)
		if err != nil {
			return nil, err
		}

		if err := createNDARequestNotification(ctx, existing, tc.OwnerID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
		}
	}

	// otherwise do nothing invalid
	return existing, nil
}

// getTrustCenter is a helper to get the trust center for an NDA request by ID
func getTrustCenter(ctx context.Context, trustCenterID string) (*generated.TrustCenter, error) {
	tc, err := transactionFromContext(ctx).TrustCenter.Query().
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
			if isDeleteOp(ctx, m) {
				if err := handleNDARequestDelete(ctx, m); err != nil {
					return nil, err
				}

				return next.Mutate(ctx, m)
			}

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
				requestID := ""

				// if this is a trust center nda request, we can get the email and trust center ID from the request
				request, ok := retVal.(*generated.TrustCenterNDARequest)
				if ok {
					email = request.Email
					trustCenterID = request.TrustCenterID
					requestID = request.ID
				} else {
					// otherwise its a document data mutation, get from the anonymous trust center user context
					anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx)
					if ok {
						email = anon.SubjectEmail
						trustCenterID = anon.TrustCenterID

						// lookup the request ID for the JWT
						req, err := transactionFromContext(ctx).TrustCenterNDARequest.Query().
							Where(
								trustcenterndarequest.EmailEQ(email),
								trustcenterndarequest.TrustCenterIDEQ(trustCenterID),
							).
							Only(ctx)
						if err != nil {
							logx.FromContext(ctx).Error().Err(err).Msg("failed to lookup NDA request for anonymous user upon signing, unable to set sub for JWT and send auth email")

							return retVal, err
						}

						requestID = req.ID
					}
				}

				// send auth email upon signing
				if err := sendTrustCenterAuthEmail(ctx, ndaAuthEmailData{
					requestID:     requestID,
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
				requestID:     request.ID,
				email:         request.Email,
				trustCenterID: request.TrustCenterID,
			}); err != nil {
				return nil, err
			}

			return v, nil
		})
	}, ent.OpUpdateOne|ent.OpUpdate|ent.OpDeleteOne)
}

func handleNDARequestDelete(ctx context.Context, m *generated.TrustCenterNDARequestMutation) error {
	id, ok := m.ID()
	if !ok {
		logx.FromContext(ctx).Error().Msg("missing ID for deleted NDA request, unable to cleanup tuples")

		return nil
	}

	tcID, ok := m.TrustCenterID()
	if !ok {
		oldID, err := m.OldTrustCenterID(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("missing trust center ID for deleted NDA request, unable to cleanup tuples")

			return nil
		}
		tcID = oldID
	}

	// delete any tuples associated with the nda request
	deleteTuple := fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:   fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, id),
		SubjectType: "user",
		ObjectID:    tcID,
		ObjectType:  "trust_center",
		Relation:    "nda_signed",
	})

	if _, err := m.Authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{deleteTuple}); err != nil {
		return err
	}

	return nil
}

func createNDARequestNotification(ctx context.Context, ndaRequest *generated.TrustCenterNDARequest, ownerID string) error {
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
			"url":             "trust-center/NDAs",
		},
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	_, err := transactionFromContext(ctx).Notification.Create().SetInput(input).Save(allowCtx)

	return err
}

type ndaAuthEmailData struct {
	requestID     string
	email         string
	trustCenterID string
}

func sendTrustCenterNDARequestEmail(ctx context.Context, ndaRequest ndaAuthEmailData) error {
	if ndaRequest.trustCenterID == "" || ndaRequest.email == "" {
		logx.FromContext(ctx).Info().Msg("missing trust center ID or email for auth email")
		return nil
	}

	if ndaRequest.requestID == "" {
		logx.FromContext(ctx).Error().Msg("created NDA request has empty ID, unable to set sub for JWT and send email")
		return ErrMissingIDForTrustCenterNDARequest
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

	accessToken, err := generateTrustCenterJWT(ctx, tc, ndaRequest)
	if err != nil {
		return err
	}

	fullURL, err := addTokenToURLAndShorten(ctx, buildNDATrustCenterURL(tc), accessToken)
	if err != nil {
		return err
	}

	emailMsg, err := transactionFromContext(ctx).Emailer.NewTrustCenterNDARequestEmail(emailtemplates.Recipient{
		Email: ndaRequest.email,
	}, "", emailtemplates.TrustCenterNDARequestData{
		OrganizationName:      tc.Edges.Setting.CompanyName,
		TrustCenterNDAFullURL: fullURL,
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
		logx.FromContext(ctx).Error().Msg("missing trust center ID or email for auth email")

		return nil
	}

	if ndaRequest.requestID == "" {
		logx.FromContext(ctx).Error().Msg("created NDA request has empty ID, unable to set sub for JWT and send email")

		return ErrMissingIDForTrustCenterNDARequest
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

	accessToken, err := generateTrustCenterJWT(ctx, tc, ndaRequest)
	if err != nil {
		return err
	}

	fullURL, err := addTokenToURLAndShorten(ctx, getTrustCenterBaseURL(tc), accessToken)
	if err != nil {
		return err
	}

	// pass the full url and leave the token empty
	emailMsg, err := transactionFromContext(ctx).Emailer.NewTrustCenterAuthEmail(emailtemplates.Recipient{
		Email: ndaRequest.email,
	}, "", emailtemplates.TrustCenterAuthData{
		OrganizationName:       tc.Edges.Setting.CompanyName,
		TrustCenterAuthFullURL: fullURL,
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

// addTokenToURLAndShorten adds a token to the URL as a query parameter and returns a shortened version of the URL if the shortlinks client is available, otherwise returns the original URL with the token added
func addTokenToURLAndShorten(ctx context.Context, baseURL url.URL, token string) (string, error) {
	url := baseURL.ResolveReference(&url.URL{RawQuery: url.Values{"token": []string{token}}.Encode()})

	regularLink := url.String()

	// if no shortlinks client, return the regular link with the token
	if transactionFromContext(ctx).Shortlinks == nil {
		return regularLink, nil
	}

	shortenedURL, shortenErr := transactionFromContext(ctx).Shortlinks.Create(ctx, regularLink, "")
	if shortenErr != nil {
		// don't log the full link as it contains a confidential token, just log the base URL
		logx.FromContext(ctx).Error().Str("baseURL", baseURL.String()).Err(shortenErr).Msg("failed to shorten URL, using original")

		return regularLink, nil
	}

	return shortenedURL, nil
}

func generateTrustCenterJWT(ctx context.Context, tc *generated.TrustCenter, ndaRequest ndaAuthEmailData) (string, error) {
	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, ndaRequest.requestID)

	duration := tc.TokenManager.Config().TrustCenterNDARequestAccessDuration

	accessToken, _, err := transactionFromContext(ctx).TokenManager.CreateTokenPair(&tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:        anonUserID,
		OrgID:         tc.OwnerID,
		TrustCenterID: tc.ID,
		Email:         ndaRequest.email,
	}, tokens.WithAccessDuration(duration))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create token for auth email")

		return "", err
	}

	return accessToken, nil
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
