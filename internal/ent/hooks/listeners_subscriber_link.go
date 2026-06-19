package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/contact"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaSubscriberLinkListeners registers a listener that links a newly created
// subscriber to an existing contact and/or user with a matching email asynchronously
// after the subscriber mutation commits
func RegisterGalaSubscriberLinkListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeSubscriber),
			Name:       "subscriber.link_identity",
			Operations: []string{ent.OpCreate.String()},
			Handle:     handleSubscriberCreatedLink,
		},
	)
}

// handleSubscriberCreatedLink matches a newly created subscriber to an existing contact
// and/or user by email within the subscriber's owning organization and records the
// association via the contact_id and user_id edges. Matching runs under an internal
// caller scoped to the subscriber's owner because subscribers can be created anonymously
// through the trust center, in which case the originating caller cannot read contacts
func handleSubscriberCreatedLink(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	subscriberID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || subscriberID == "" {
		return nil
	}

	allowCtx := auth.WithCaller(privacy.DecisionContext(ctx.Context, privacy.Allow), &auth.Caller{
		Capabilities: auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
	})

	sub, err := client.Subscriber.Get(allowCtx, subscriberID)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil
		}

		return err
	}

	if sub.Email == "" {
		return nil
	}

	update := client.Subscriber.UpdateOneID(sub.ID)
	changed := false

	if sub.ContactID == "" {
		contactID, err := matchSubscriberContactID(allowCtx, client, sub.OwnerID, sub.Email)
		if err != nil {
			return err
		}

		if contactID != "" {
			update.SetContactID(contactID)

			changed = true
		}
	}

	if sub.UserID == "" {
		userID, err := matchSubscriberUserID(allowCtx, client, sub.OwnerID, sub.Email)
		if err != nil {
			return err
		}

		if userID != "" {
			update.SetUserID(userID)

			changed = true
		}
	}

	if !changed {
		return nil
	}

	if err := update.Exec(allowCtx); err != nil {
		logx.FromContext(allowCtx).Error().Err(err).Str("subscriber_id", sub.ID).Msg("failed linking subscriber to contact/user")

		return err
	}

	return nil
}

// matchSubscriberContactID returns the id of a contact in the organization whose email
// matches the subscriber email, or empty when no contact matches
func matchSubscriberContactID(ctx context.Context, client *entgen.Client, ownerID, email string) (string, error) {
	contactID, err := client.Contact.Query().
		Where(
			contact.OwnerID(ownerID),
			contact.EmailEqualFold(email),
		).
		FirstID(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return "", nil
		}

		return "", err
	}

	return contactID, nil
}

// matchSubscriberUserID returns the id of a user who is a member of the organization and
// whose email matches the subscriber email, or empty when no user matches
func matchSubscriberUserID(ctx context.Context, client *entgen.Client, ownerID, email string) (string, error) {
	userID, err := client.User.Query().
		Where(
			user.EmailEqualFold(email),
			user.HasOrgMembershipsWith(orgmembership.OrganizationID(ownerID)),
		).
		FirstID(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return "", nil
		}

		return "", err
	}

	return userID, nil
}
