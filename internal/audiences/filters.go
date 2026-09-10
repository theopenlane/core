package audiences

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
	"github.com/theopenlane/core/v2/internal/ent/generated/group"
	"github.com/theopenlane/core/v2/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
)

const (
	// metadata key for the subscriber's unsubscribe token.
	MetadataUnsubscribeTokenKey = "unsubscribeToken"
	resolveBatchSize            = 100
)

var (
	errAudienceFiltersRequired       = errors.New("audience filters must include at least one selector")
	errManualAudienceFilters         = errors.New("manual audiences cannot define filters")
	errDynamicAudienceFiltersMissing = errors.New("dynamic audiences require filters")
	errUnsupportedAudienceType       = errors.New("unsupported audience type")
	errSelectorSchemaRequired        = errors.New("selector schema is required")
	errSchemaNotRegistered           = errors.New("schema is not registered")
	errUnsupportedRecipientSource    = errors.New("schema cannot be used as an audience recipient source")
	errKeyMatchUnsupported           = errors.New("key_match is not supported for audience selectors yet")
	errSourceSelectorsUnsupported    = errors.New("source selectors are not supported for audience selectors yet")
)

type filterSet struct {
	// exported so json marshalling can work when taking out the stored filters
	// from the db object
	Selectors []entityops.TargetSelector `json:"selectors,omitempty"`
}

// ResolvedRecipient is a recipient that was matched by an audience filter
type ResolvedRecipient struct {
	entityops.AudienceMemberProjection
	// schema where the recipient is from
	Source string
	// ID identifier of the object
	SourceObjectID string
}

type recipientSelectorOptions[T any] struct {
	targetType reflect.Type
	fetchFn    func(lastKnownID string) ([]T, error)
	id         func(T) string
	recipient  func(T) ResolvedRecipient
}

func parseSelectors(filters map[string]any) ([]entityops.TargetSelector, error) {
	if len(filters) == 0 {
		return nil, errAudienceFiltersRequired
	}

	buf, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("marshal audience filters: %w", err)
	}

	if _, ok := filters["selectors"]; ok {
		var set filterSet
		if err := json.Unmarshal(buf, &set); err != nil {
			return nil, fmt.Errorf("decode audience selectors: %w", err)
		}

		return set.Selectors, nil
	}

	var selector entityops.TargetSelector
	if err := json.Unmarshal(buf, &selector); err != nil {
		return nil, fmt.Errorf("decode audience selector: %w", err)
	}

	return []entityops.TargetSelector{selector}, nil
}

// ValidateFilters validates the filters and selectors provided. and also sets some
// basic rules like ensure manual audiences cannot have filters.
func ValidateFilters(audienceType enums.AudienceType, filters map[string]any) error {
	switch audienceType {
	case enums.AudienceTypeManual:
		if len(filters) > 0 {
			return errManualAudienceFilters
		}

		return nil
	case enums.AudienceTypeDynamic:
		if len(filters) == 0 {
			return errDynamicAudienceFiltersMissing
		}

		selectors, err := parseSelectors(filters)
		if err != nil {
			return err
		}

		if len(selectors) == 0 {
			return errAudienceFiltersRequired
		}

		for i, selector := range selectors {
			if err := validateSelector(selector); err != nil {
				return fmt.Errorf("selector %d: %w", i, err)
			}
		}

		return nil
	default:
		return fmt.Errorf("%w: %q", errUnsupportedAudienceType, audienceType)
	}
}

// ResolveRecipients finds recipients that match the provided cel filters and then processes them in batches
func ResolveRecipients(ctx context.Context, db *generated.Client, filters map[string]any, handlerFn func([]ResolvedRecipient) error) error {
	selectors, err := parseSelectors(filters)
	if err != nil {
		return err
	}

	for _, selector := range selectors {
		if err := validateSelector(selector); err != nil {
			return err
		}

		if err := resolveSelectors(ctx, db, selector, handlerFn); err != nil {
			return err
		}
	}

	return nil
}

func validateSelector(selector entityops.TargetSelector) error {
	if selector.Schema.IsZero() {
		return errSelectorSchemaRequired
	}

	schema, ok := entityops.LookupSchema(selector.Schema.Name)
	if !ok {
		return fmt.Errorf("%w: %q", errSchemaNotRegistered, selector.Schema.Name)
	}

	switch schema.Snake {
	case entityops.SchemaContact.Snake, entityops.SchemaIdentityHolder.Snake,
		entityops.SchemaSubscriber.Snake, entityops.SchemaUser.Snake, entityops.SchemaGroup.Snake:
	default:
		return fmt.Errorf("%w: %q", errUnsupportedRecipientSource, schema.Snake)
	}

	if selector.KeyMatch != nil {
		return errKeyMatchUnsupported
	}

	if len(selector.SourceContext) > 0 || !selector.SourceSchema.IsZero() {
		return errSourceSelectorsUnsupported
	}

	return nil
}

func resolveSelectors(ctx context.Context, db *generated.Client, selector entityops.TargetSelector, handlerFn func([]ResolvedRecipient) error) error {
	schema, _ := entityops.LookupSchema(selector.Schema.Name)

	switch schema.Snake {
	case entityops.SchemaSubscriber.Snake:
		return resolveSelector(ctx, selector, recipientSelectorOptions[*generated.Subscriber]{
			targetType: reflect.TypeFor[entityops.SubscriberProjection](),

			fetchFn: func(lastKnownID string) ([]*generated.Subscriber, error) {

				query := db.Subscriber.Query().
					Order(subscriber.ByID()).
					Limit(resolveBatchSize).
					Where(subscriber.Active(true),
						subscriber.VerifiedEmail(true),
						subscriber.Unsubscribed(false),
						subscriber.EmailNEQ(""))

				if lastKnownID != "" {
					query.Where(subscriber.IDGT(lastKnownID))
				}

				return query.All(ctx)
			},

			id: func(s *generated.Subscriber) string { return s.ID },

			recipient: func(s *generated.Subscriber) ResolvedRecipient {
				return ResolvedRecipient{
					AudienceMemberProjection: entityops.AudienceMemberProjection{
						Email: s.Email, SubscriberID: s.ID,
						Metadata: map[string]any{
							MetadataUnsubscribeTokenKey: s.Token,
						},
					},
					Source: entityops.SchemaSubscriber.Snake, SourceObjectID: s.ID,
				}
			},
		}, handlerFn)

	case entityops.SchemaUser.Snake:

		return resolveUserRecipients(ctx, db, selector, "", handlerFn)

	case entityops.SchemaGroup.Snake:

		return resolveSelector(ctx, selector, recipientSelectorOptions[*generated.Group]{
			targetType: reflect.TypeFor[entityops.GroupProjection](),
			fetchFn: func(lastKnownID string) ([]*generated.Group, error) {

				query := db.Group.Query().
					Order(group.ByID()).
					Limit(resolveBatchSize)

				if lastKnownID != "" {
					query.Where(group.IDGT(lastKnownID))
				}

				return query.All(ctx)
			},

			id: func(g *generated.Group) string { return g.ID },

			recipient: func(g *generated.Group) ResolvedRecipient {
				return ResolvedRecipient{
					AudienceMemberProjection: entityops.AudienceMemberProjection{
						GroupID: g.ID,
					},
				}
			},
		}, func(groups []ResolvedRecipient) error {
			for _, g := range groups {
				if err := resolveUserRecipients(ctx, db, entityops.TargetSelector{}, g.GroupID, handlerFn); err != nil {
					return err
				}
			}
			return nil
		})

	case entityops.SchemaContact.Snake:
		return resolveSelector(ctx, selector, recipientSelectorOptions[*generated.Contact]{
			targetType: reflect.TypeFor[entityops.ContactProjection](),
			fetchFn: func(lastKnownID string) ([]*generated.Contact, error) {
				query := db.Contact.Query().
					Where(
						contact.EmailNotNil(),
						contact.EmailNEQ(""),
					).
					Order(contact.ByID()).
					Limit(resolveBatchSize)

				if lastKnownID != "" {
					query.Where(contact.IDGT(lastKnownID))
				}

				return query.All(ctx)
			},

			id: func(c *generated.Contact) string { return c.ID },

			recipient: func(c *generated.Contact) ResolvedRecipient {
				return ResolvedRecipient{
					AudienceMemberProjection: entityops.AudienceMemberProjection{
						Email:     c.Email,
						FullName:  c.FullName,
						ContactID: c.ID,
					},
					Source:         entityops.SchemaContact.Snake,
					SourceObjectID: c.ID,
				}
			},
		}, handlerFn)

	case entityops.SchemaIdentityHolder.Snake:
		return resolveSelector(ctx, selector, recipientSelectorOptions[*generated.IdentityHolder]{
			targetType: reflect.TypeFor[entityops.IdentityHolderProjection](),
			fetchFn: func(lastKnownID string) ([]*generated.IdentityHolder, error) {
				query := db.IdentityHolder.Query().
					Where(
						identityholder.EmailNEQ(""),
					).
					Order(identityholder.ByID()).
					Limit(resolveBatchSize)

				if lastKnownID != "" {
					query.Where(identityholder.IDGT(lastKnownID))
				}

				return query.All(ctx)
			},
			id: func(holder *generated.IdentityHolder) string {
				return holder.ID
			},
			recipient: func(holder *generated.IdentityHolder) ResolvedRecipient {
				return ResolvedRecipient{
					AudienceMemberProjection: entityops.AudienceMemberProjection{
						Email:    holder.Email,
						FullName: holder.FullName,
						UserID:   holder.UserID,
					},
					Source:         entityops.SchemaIdentityHolder.Snake,
					SourceObjectID: holder.ID,
				}
			},
		}, handlerFn)
	default:
		return fmt.Errorf("%w: %q", errUnsupportedRecipientSource, schema.Snake)
	}
}

func resolveUserRecipients(ctx context.Context, db *generated.Client, selector entityops.TargetSelector, groupID string, handlerFn func([]ResolvedRecipient) error) error {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || len(caller.OrgIDs()) == 0 {
		return auth.ErrNoAuthUser
	}

	return resolveSelector(ctx, selector, recipientSelectorOptions[*generated.User]{
		targetType: reflect.TypeFor[entityops.UserProjection](),
		fetchFn: func(lastKnownID string) ([]*generated.User, error) {

			query := db.User.Query().
				Limit(resolveBatchSize).
				Order(user.ByID()).
				Where(
					user.EmailNEQ(""),
					user.HasOrgMembershipsWith(orgmembership.OrganizationIDIn(caller.OrgIDs()...)),
				)

			if groupID != "" {
				query.Where(user.HasGroupMembershipsWith(groupmembership.GroupID(groupID)))
			}

			if lastKnownID != "" {
				query.Where(user.IDGT(lastKnownID))
			}

			return query.All(ctx)
		},

		id: func(u *generated.User) string { return u.ID },

		recipient: func(u *generated.User) ResolvedRecipient {

			recipient := ResolvedRecipient{
				AudienceMemberProjection: entityops.AudienceMemberProjection{
					Email:    u.Email,
					FullName: strings.TrimSpace(u.FirstName + " " + u.LastName),
					UserID:   u.ID,
					GroupID:  groupID,
				},
				Source: entityops.SchemaUser.Snake, SourceObjectID: u.ID,
			}

			if groupID != "" {
				recipient.Source = entityops.SchemaGroup.Snake
				recipient.SourceObjectID = groupID
			}

			return recipient
		},
	}, handlerFn)
}

func resolveSelector[T any](ctx context.Context, selector entityops.TargetSelector, opts recipientSelectorOptions[T], handlerFn func([]ResolvedRecipient) error) error {
	evaluator, err := entityops.NewEvaluator(opts.targetType, nil)
	if err != nil {
		return err
	}

	var lastID string

	for {
		items, err := opts.fetchFn(lastID)
		if err != nil {
			return err
		}

		recipients := make([]ResolvedRecipient, 0, len(items))

		for _, item := range items {
			lastID = opts.id(item)

			match, err := entityops.MatchSelector(ctx, evaluator, selector.Expression, item)
			if err != nil {
				return err
			}

			if !match {
				continue
			}

			recipients = append(recipients, opts.recipient(item))
		}

		if len(recipients) > 0 {
			if err := handlerFn(recipients); err != nil {
				return err
			}
		}

		if len(items) < resolveBatchSize {
			break
		}
	}

	return nil
}
