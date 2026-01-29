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
	"github.com/theopenlane/core/internal/ent/generated/organization"
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
				return next.Mutate(ctx, m)
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

			sendTrustCenterNDARequestEmail(ctx, m, request)

			return v, nil
		})
	}, ent.OpCreate)
}

func handleExistingNDARequest(ctx, queryCtx context.Context, m *generated.TrustCenterNDARequestMutation, existing *generated.TrustCenterNDARequest) (*generated.TrustCenterNDARequest, error) {
	if existing.Status == enums.TrustCenterNDARequestStatusApproved {
		sendTrustCenterAuthEmail(ctx, m, existing)
		return existing, nil
	}

	if existing.Status == enums.TrustCenterNDARequestStatusNeedsApproval {
		tc, err := m.Client().TrustCenter.Query().
			Where(trustcenter.IDEQ(existing.TrustCenterID)).
			Only(queryCtx)
		if err != nil {
			return nil, err
		}

		if err := createNDARequestNotification(ctx, m, existing, tc.OwnerID); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA request notification")
		}
		return existing, nil
	}

	sendTrustCenterNDARequestEmail(ctx, m, existing)
	return existing, nil
}

// HookTrustCenterNDARequestUpdate handles NDA request status updates - sends email when approved
func HookTrustCenterNDARequestUpdate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterNDARequestFunc(func(ctx context.Context, m *generated.TrustCenterNDARequestMutation) (generated.Value, error) {
			status, ok := m.Status()
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

				return next.Mutate(ctx, m)
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

			sendTrustCenterNDARequestEmail(ctx, m, request)

			return v, nil
		})
	}, ent.OpUpdateOne)
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

func sendTrustCenterNDARequestEmail(ctx context.Context, m *generated.TrustCenterNDARequestMutation, ndaRequest *generated.TrustCenterNDARequest) {
	if ndaRequest.TrustCenterID == "" || ndaRequest.Email == "" {
		return
	}

	tc, err := m.Client().TrustCenter.Query().
		Where(trustcenter.IDEQ(ndaRequest.TrustCenterID)).
		WithCustomDomain().
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get trust center for NDA email")
		return
	}

	org, err := m.Client().Organization.Query().
		Where(organization.ID(tc.OwnerID)).
		Select(organization.FieldDisplayName).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get organization for NDA email")
		return
	}

	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, uuid.New().String())

	accessToken, _, err := m.TokenManager.CreateTokenPair(&tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:        anonUserID,
		OrgID:         tc.OwnerID,
		TrustCenterID: ndaRequest.TrustCenterID,
		Email:         ndaRequest.Email,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create token for NDA email")
		return
	}

	trustCenterURL := buildNDATrustCenterURL(tc)

	emailMsg, err := m.Emailer.NewTrustCenterNDARequestEmail(emailtemplates.Recipient{
		Email: ndaRequest.Email,
	}, accessToken, emailtemplates.TrustCenterNDARequestData{
		OrganizationName: org.DisplayName,
		TrustCenterURL:   trustCenterURL,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create NDA email")
		return
	}

	if _, err = m.Job.Insert(ctx, jobs.EmailArgs{Message: *emailMsg}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to queue NDA email")
	}
}

func buildNDATrustCenterURL(tc *generated.TrustCenter) string {
	const ndaPath = "/access/sign-nda"

	trustCenterURL := url.URL{Scheme: "https"}

	if tc.Edges.CustomDomain != nil {
		customHost := tc.Edges.CustomDomain.CnameRecord
		if normalized, err := domain.NormalizeHostname(customHost); err == nil {
			customHost = normalized
		}
		trustCenterURL.Host = customHost
		trustCenterURL.Path = ndaPath
		return trustCenterURL.String()
	}

	defaultHost := trustCenterConfig.DefaultTrustCenterDomain
	if normalized, err := domain.NormalizeHostname(defaultHost); err == nil {
		defaultHost = normalized
	}
	trustCenterURL.Host = defaultHost
	trustCenterURL.Path = "/" + tc.Slug + ndaPath

	return trustCenterURL.String()
}

func sendTrustCenterAuthEmail(ctx context.Context, m *generated.TrustCenterNDARequestMutation, ndaRequest *generated.TrustCenterNDARequest) {
	if ndaRequest.TrustCenterID == "" || ndaRequest.Email == "" {
		return
	}

	tc, err := m.Client().TrustCenter.Query().
		Where(trustcenter.IDEQ(ndaRequest.TrustCenterID)).
		WithCustomDomain().
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get trust center for auth email")
		return
	}

	org, err := m.Client().Organization.Query().
		Where(organization.ID(tc.OwnerID)).
		Select(organization.FieldDisplayName).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get organization for auth email")
		return
	}

	anonUserID := fmt.Sprintf("%s%s", authmanager.AnonTrustCenterJWTPrefix, uuid.New().String())

	accessToken, _, err := m.TokenManager.CreateTokenPair(&tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonUserID,
		},
		UserID:        anonUserID,
		OrgID:         tc.OwnerID,
		TrustCenterID: ndaRequest.TrustCenterID,
		Email:         ndaRequest.Email,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create token for auth email")
		return
	}

	trustCenterURL := buildNDATrustCenterURL(tc)

	emailMsg, err := m.Emailer.NewTrustCenterAuthEmail(emailtemplates.Recipient{
		Email: ndaRequest.Email,
	}, accessToken, emailtemplates.TrustCenterAuthData{
		OrganizationName: org.DisplayName,
		TrustCenterURL:   trustCenterURL,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create auth email")
		return
	}

	if _, err = m.Job.Insert(ctx, jobs.EmailArgs{Message: *emailMsg}, nil); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to queue auth email")
	}
}
