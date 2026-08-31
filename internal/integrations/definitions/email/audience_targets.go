package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/audiences"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/audiencemember"
	"github.com/theopenlane/core/v2/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/subscriber"
)

const (
	audienceTargetSourceKey       = "audience_source"
	audienceTargetAudienceIDKey   = "audience_id"
	audienceTargetSourceObjectKey = "source_object_id"
	audienceTargetBatchSize       = 100
	audienceTargetMetadataFields  = 3
)

var errUnsupportedAudienceType = errors.New("unsupported audience type")

type audienceRecipient struct {
	email          string
	fullName       string
	contactID      string
	userID         string
	groupID        string
	subscriberID   string
	source         string
	audienceID     string
	sourceObjectID string
	metadata       map[string]any
}

type audienceRecipientResolveOptions struct {
	audienceID   string
	audienceType enums.AudienceType
	audience     *generated.Audience
	db           *generated.Client
	orgID        string
	filters      map[string]any
}

type campaignRecipientHandlerFunc func([]audienceRecipient) error

func snapshotCampaignAudiences(ctx context.Context, db *generated.Client, camp *generated.Campaign) error {
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	records, err := camp.QueryAudiences().All(allowCtx)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	return snapshotCampaignRecipients(allowCtx, db, camp, func(handle campaignRecipientHandlerFunc) error {
		for _, aud := range records {
			opts := audienceRecipientResolveOptions{
				audienceID:   aud.ID,
				audienceType: aud.AudienceType,
				audience:     aud,
				db:           db,
				orgID:        camp.OwnerID,
				filters:      aud.Filters,
			}

			if err := resolveAudienceRecipients(allowCtx, opts, handle); err != nil {
				return err
			}
		}

		return nil
	})
}

func snapshotCampaignRecipients(ctx context.Context, db *generated.Client, camp *generated.Campaign, resolveFn func(campaignRecipientHandlerFunc) error) error {
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

	err := resolveFn(func(page []audienceRecipient) error {
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
			key := normalizeAudienceEmail(target.Email)
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

func (s *set) add(recipient audienceRecipient) bool {
	key := normalizeAudienceEmail(recipient.email)
	if key == "" {
		return false
	}

	if _, ok := s.seen[key]; ok {
		return false
	}

	s.seen[key] = struct{}{}

	return true
}

func resolveAudienceRecipients(ctx context.Context, opts audienceRecipientResolveOptions, handle campaignRecipientHandlerFunc) error {
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

			recipients := make([]audienceRecipient, 0, len(members))
			for _, member := range members {
				lastID = member.ID
				recipients = append(recipients, audienceRecipient{
					email:          member.Email,
					fullName:       member.FullName,
					contactID:      member.ContactID,
					userID:         member.UserID,
					groupID:        member.GroupID,
					subscriberID:   member.SubscriberID,
					source:         audiencemember.Label,
					audienceID:     opts.audienceID,
					sourceObjectID: member.ID,
					metadata:       member.Metadata,
				})
			}

			if len(recipients) > 0 {
				if err := handle(recipients); err != nil {
					return err
				}
			}

			if len(members) < audienceTargetBatchSize {
				break
			}
		}

		return nil
	case enums.AudienceTypeDynamic:
		return audiences.ResolveRecipients(ctx, opts.db, opts.orgID, opts.filters, func(page []audiences.Recipient) error {
			recipients := make([]audienceRecipient, 0, len(page))
			for _, recipient := range page {
				recipients = append(recipients, audienceRecipient{
					email:          recipient.Email,
					fullName:       recipient.FullName,
					contactID:      recipient.ContactID,
					userID:         recipient.UserID,
					groupID:        recipient.GroupID,
					subscriberID:   recipient.SubscriberID,
					source:         recipient.Source,
					audienceID:     opts.audienceID,
					sourceObjectID: recipient.SourceObjectID,
					metadata:       recipient.Metadata,
				})
			}

			return handle(recipients)
		})
	default:
		return fmt.Errorf("%w: %q", errUnsupportedAudienceType, opts.audienceType)
	}
}

func resolveTrustCenterSubscriberRecipients(ctx context.Context, db *generated.Client, camp *generated.Campaign, handle campaignRecipientHandlerFunc) error {
	var lastID string
	for {
		query := db.Subscriber.Query().
			Where(
				subscriber.TrustCenterID(camp.TrustCenterID),
				subscriber.Active(true),
				subscriber.VerifiedEmail(true),
				subscriber.Unsubscribed(false),
			).
			Order(subscriber.ByID()).
			Limit(audienceTargetBatchSize)

		if lastID != "" {
			query.Where(subscriber.IDGT(lastID))
		}

		subscribers, err := query.All(ctx)
		if err != nil {
			return err
		}

		recipients := make([]audienceRecipient, 0, len(subscribers))
		for _, sub := range subscribers {
			lastID = sub.ID
			recipients = append(recipients, audienceRecipient{
				email:          sub.Email,
				subscriberID:   sub.ID,
				source:         subscriber.Label,
				sourceObjectID: sub.ID,
				metadata: map[string]any{
					MetadataUnsubscribeTokenKey: sub.Token,
				},
			})
		}

		if len(recipients) > 0 {
			if err := handle(recipients); err != nil {
				return err
			}
		}

		if len(subscribers) < audienceTargetBatchSize {
			break
		}
	}

	return nil
}

func buildAudienceCampaignTarget(db *generated.Client, camp *generated.Campaign, recipient audienceRecipient) *generated.CampaignTargetCreate {
	create := db.CampaignTarget.Create().
		SetCampaignID(camp.ID).
		SetOwnerID(camp.OwnerID).
		SetEmail(recipient.email).
		SetNillableContactID(lo.EmptyableToPtr(recipient.contactID)).
		SetNillableUserID(lo.EmptyableToPtr(recipient.userID)).
		SetNillableGroupID(lo.EmptyableToPtr(recipient.groupID)).
		SetNillableSubscriberID(lo.EmptyableToPtr(recipient.subscriberID))

	if strings.TrimSpace(recipient.fullName) != "" {
		create.SetFullName(recipient.fullName)
	}

	metadata := audienceTargetMetadata(recipient)
	if len(metadata) > 0 {
		create.SetMetadata(metadata)
	}

	return create
}

func audienceTargetMetadata(recipient audienceRecipient) map[string]any {
	metadata := make(map[string]any, len(recipient.metadata)+audienceTargetMetadataFields)
	for key, value := range recipient.metadata {
		metadata[key] = value
	}

	if recipient.source != "" {
		metadata[audienceTargetSourceKey] = recipient.source
	}

	if recipient.audienceID != "" {
		metadata[audienceTargetAudienceIDKey] = recipient.audienceID
	}

	if recipient.sourceObjectID != "" {
		metadata[audienceTargetSourceObjectKey] = recipient.sourceObjectID
	}

	return metadata
}

func normalizeAudienceEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
