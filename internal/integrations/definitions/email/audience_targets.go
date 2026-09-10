package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/audiences"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/audiencemember"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
)

const (
	audienceTargetSourceKey       = "audience_source"
	audienceTargetAudienceIDKey   = "audience_id"
	audienceTargetSourceObjectKey = "source_object_id"
	audienceTargetBatchSize       = 100
	audienceTargetMetadataFields  = 3
)

var (
	errUnsupportedAudienceType = errors.New("unsupported audience type")
)

type recipientResolveOptions struct {
	audienceID   string
	audienceType enums.AudienceType
	audience     *generated.Audience
	client       *generated.Client
	filters      map[string]any
}

type recipientHandlerFunc func([]audiences.ResolvedRecipient) error

func snapshotCampaignAudiences(ctx context.Context, db *generated.Client, camp *generated.Campaign) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	records, err := camp.QueryAudiences().All(allowCtx)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	return snapshotCampaignRecipients(allowCtx, db, camp, func(handle recipientHandlerFunc) error {
		for _, aud := range records {
			opts := recipientResolveOptions{
				audienceID:   aud.ID,
				audienceType: aud.AudienceType,
				audience:     aud,
				client:       db,
				filters:      aud.Filters,
			}

			if err := resolveAudienceRecipients(allowCtx, opts, handle); err != nil {
				return err
			}
		}

		return nil
	})
}

func snapshotCampaignRecipients(ctx context.Context, db *generated.Client, camp *generated.Campaign, resolveFn func(recipientHandlerFunc) error) error {
	recipients := &set{
		seen: map[string]struct{}{},
	}
	if err := recipients.loadExistingCampaignTargets(ctx, db, camp.ID); err != nil {
		return err
	}

	builders := make([]*generated.CampaignTargetCreate, 0, audienceTargetBatchSize)
	createTargetsFn := func() error {
		if len(builders) == 0 {
			return nil
		}

		if err := db.CampaignTarget.CreateBulk(builders...).Exec(ctx); err != nil {
			return err
		}

		builders = builders[:0]

		return nil
	}

	err := resolveFn(func(page []audiences.ResolvedRecipient) error {
		for _, recipient := range page {
			if !recipients.add(recipient) {
				continue
			}

			builders = append(builders, buildAudienceCampaignTarget(db, camp, recipient))
			if len(builders) == audienceTargetBatchSize {
				if err := createTargetsFn(); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return createTargetsFn()
}

type set struct {
	seen map[string]struct{}
}

func (s *set) loadExistingCampaignTargets(ctx context.Context, db *generated.Client, campaignID string) error {
	var lastID string
	for {
		query := db.CampaignTarget.Query().
			Where(campaigntarget.CampaignIDEQ(campaignID)).
			Select(campaigntarget.FieldID, campaigntarget.FieldEmail).
			Order(campaigntarget.ByID()).
			Limit(audienceTargetBatchSize)

		if lastID != "" {
			query.Where(campaigntarget.IDGT(lastID))
		}

		targets, err := query.All(ctx)
		if err != nil {
			return err
		}

		for _, target := range targets {
			lastID = target.ID
			key := canonicalizeEmail(target.Email)
			if key != "" {
				s.seen[key] = struct{}{}
			}
		}

		if len(targets) < audienceTargetBatchSize {
			break
		}
	}

	return nil
}

func (s *set) add(recipient audiences.ResolvedRecipient) bool {
	key := canonicalizeEmail(recipient.Email)
	if key == "" {
		return false
	}

	if _, ok := s.seen[key]; ok {
		return false
	}

	s.seen[key] = struct{}{}
	return true
}

func resolveAudienceRecipients(ctx context.Context, opts recipientResolveOptions, handlerFn recipientHandlerFunc) error {
	switch opts.audienceType {
	case enums.AudienceTypeManual:
		var lastID string

		for {
			query := opts.audience.QueryAudienceMembers().
				Order(audiencemember.ByID()).
				Limit(audienceTargetBatchSize)

			if lastID != "" {
				query.Where(audiencemember.IDGT(lastID))
			}

			members, err := query.All(ctx)
			if err != nil {
				return err
			}

			recipients := make([]audiences.ResolvedRecipient, 0, len(members))
			for _, member := range members {
				lastID = member.ID
				recipients = append(recipients, audiences.ResolvedRecipient{
					AudienceMemberProjection: entityops.AudienceMemberProjection{
						Email:        member.Email,
						FullName:     member.FullName,
						ContactID:    member.ContactID,
						UserID:       member.UserID,
						GroupID:      member.GroupID,
						SubscriberID: member.SubscriberID,
						AudienceID:   opts.audienceID,
						Metadata:     member.Metadata,
					},
					Source:         audiencemember.Label,
					SourceObjectID: member.ID,
				})
			}

			if len(recipients) > 0 {
				if err := handlerFn(recipients); err != nil {
					return err
				}
			}

			if len(members) < audienceTargetBatchSize {
				break
			}
		}

		return nil
	case enums.AudienceTypeDynamic:

		return audiences.ResolveRecipients(ctx, opts.client, opts.filters, func(page []audiences.ResolvedRecipient) error {
			for i := range page {
				page[i].AudienceID = opts.audienceID
			}

			return handlerFn(page)

		})
	default:
		return fmt.Errorf("%w: %q", errUnsupportedAudienceType, opts.audienceType)
	}
}

func resolveTrustCenterSubscriberRecipients(ctx context.Context, db *generated.Client, camp *generated.Campaign, handle recipientHandlerFunc) error {
	filters := map[string]any{
		"schema": entityops.SchemaSubscriber.Snake,
		"expression": fmt.Sprintf(
			"target.trust_center_id == %q && target.active && target.verified_email && !target.unsubscribed",
			camp.TrustCenterID,
		),
	}

	return audiences.ResolveRecipients(ctx, db, filters, handle)
}

func buildAudienceCampaignTarget(db *generated.Client, camp *generated.Campaign, recipient audiences.ResolvedRecipient) *generated.CampaignTargetCreate {
	create := db.CampaignTarget.Create().
		SetCampaignID(camp.ID).
		SetOwnerID(camp.OwnerID).
		SetEmail(recipient.Email).
		SetNillableContactID(lo.EmptyableToPtr(recipient.ContactID)).
		SetNillableUserID(lo.EmptyableToPtr(recipient.UserID)).
		SetNillableGroupID(lo.EmptyableToPtr(recipient.GroupID)).
		SetNillableSubscriberID(lo.EmptyableToPtr(recipient.SubscriberID))

	if strings.TrimSpace(recipient.FullName) != "" {
		create.SetFullName(recipient.FullName)
	}

	metadata := getAudienceMetadata(recipient)
	if len(metadata) > 0 {
		create.SetMetadata(metadata)
	}

	return create
}

func getAudienceMetadata(recipient audiences.ResolvedRecipient) map[string]any {
	metadata := make(map[string]any, len(recipient.Metadata)+audienceTargetMetadataFields)
	for key, value := range recipient.Metadata {
		metadata[key] = value
	}

	if recipient.Source != "" {
		metadata[audienceTargetSourceKey] = recipient.Source
	}

	if recipient.AudienceID != "" {
		metadata[audienceTargetAudienceIDKey] = recipient.AudienceID
	}

	if recipient.SourceObjectID != "" {
		metadata[audienceTargetSourceObjectKey] = recipient.SourceObjectID
	}

	return metadata
}

func canonicalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }
