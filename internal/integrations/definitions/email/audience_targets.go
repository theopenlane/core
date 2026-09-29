package email

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"

	"github.com/theopenlane/core/v2/internal/audiences"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/audiencemember"
)

const (
	audienceTargetSourceKey       = "audience_source"
	audienceTargetAudienceIDKey   = "audience_id"
	audienceTargetSourceObjectKey = "source_object_id"
	audienceTargetBatchSize       = 100
	audienceTargetMetadataFields  = 3
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
	// fetch existing audiences for the campaign so we can precompute the needed
	// users
	audiences, err := camp.QueryAudiences().All(ctx)
	if err != nil {
		return err
	}

	if len(audiences) == 0 {
		return nil
	}

	return snapshotCampaignRecipients(ctx, db, camp, func(handle recipientHandlerFunc) error {
		for _, aud := range audiences {
			opts := recipientResolveOptions{
				audienceID:   aud.ID,
				audienceType: aud.AudienceType,
				audience:     aud,
				client:       db,
				filters:      aud.Filters,
			}

			if err := resolveRecipientsForAudience(ctx, opts, handle); err != nil {
				return err
			}
		}

		return nil
	})
}

func snapshotCampaignRecipients(ctx context.Context, db *generated.Client, camp *generated.Campaign, resolveFn func(recipientHandlerFunc) error) error {
	// use a set to prevent a recipient appearing multiple times.
	// If for some weird reason, the campaign run fails and is retried, we do not want to create the same target again
	// so fetch the existing ones, then dedupe
	//
	// Or even if the target is created outside the run via api
	targets := &campaignTargetSet{
		seen: map[string]struct{}{},
	}

	if err := targets.load(ctx, db, camp.ID); err != nil {
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

	handlerFn := func(recipients []audiences.ResolvedRecipient) error {
		for _, recipient := range recipients {
			if !targets.add(recipient) {
				continue
			}

			builders = append(builders, buildAudienceCampaignTarget(db, camp, recipient))

			// if we are at bulk max size, flush and reset
			if len(builders) == audienceTargetBatchSize {
				if err := createTargetsFn(); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := resolveFn(handlerFn); err != nil {
		return err
	}

	// make sure to flush out the remaining builders
	return createTargetsFn()
}

func resolveRecipientsForAudience(ctx context.Context, opts recipientResolveOptions, handlerFn recipientHandlerFunc) error {

	if opts.audienceType == enums.AudienceTypeDynamic {
		return audiences.ResolveRecipients(ctx, opts.client, opts.filters, func(recipients []audiences.ResolvedRecipient) error {
			for i := range recipients {
				recipients[i].AudienceID = opts.audienceID
			}

			return handlerFn(recipients)
		})
	}

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
				Source:         audiencemember.Label,
				SourceObjectID: member.ID,
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
			})
		}

		if len(recipients) > 0 {
			if err := handlerFn(recipients); err != nil {
				return err
			}
		}

		// no more data to process so return
		if len(members) < audienceTargetBatchSize {
			break
		}
	}

	return nil

}

func resolveTrustCenterSubscriberRecipients(ctx context.Context, db *generated.Client, camp *generated.Campaign, handle recipientHandlerFunc) error {
	filters := map[string]any{
		"schema": entityops.SchemaSubscriber.Snake,
		"expression": fmt.Sprintf(
			"target.trust_center_id == %q && target.active && target.verified_email && !target.unsubscribed",
			*camp.TrustCenterID,
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
