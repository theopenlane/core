package hooks

import (
	"context"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	entgen "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/logx"
)

// init registers the subscriber link listeners so gala setup picks them up automatically
func init() { registerListeners(SubscriberLinkListeners) }

// SubscriberLinkListeners links a newly created subscriber to an existing contact and/or user with a matching email
func SubscriberLinkListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaSubscriber,
			Operations: []string{entityops.OpCreate},
			Caller: func(*auth.Caller, entityops.MutationPayload) *auth.Caller {
				return &auth.Caller{
					Capabilities: auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
				}
			},
			Handle: handleSubscriberCreatedLink,
		},
	}
}

// handleSubscriberCreatedLink links a subscriber to the owning organization's contact
// and/or user matching its email
func handleSubscriberCreatedLink(inv entityops.Invocation, _ entityops.MutationPayload) error {
	sub, ok, err := entityops.LoadEntity(inv.Context, inv.EntityID, inv.Client.Subscriber.Get)
	if err != nil || !ok {
		return err
	}

	if sub.Email == "" {
		return nil
	}

	update := inv.Client.Subscriber.UpdateOneID(sub.ID)
	changed := false

	if sub.ContactID == "" {
		contactID, err := matchSubscriberContactID(inv.Context, inv.Client, sub.OwnerID, sub.Email)
		if err != nil {
			return err
		}

		if contactID != "" {
			update.SetContactID(contactID)

			changed = true
		}
	}

	if sub.UserID == "" {
		userID, err := orgUserIDByEmail(inv.Context, inv.Client, sub.OwnerID, sub.Email)
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

	if err := update.Exec(inv.Context); err != nil {
		logx.FromContext(inv.Context).Error().Err(err).Msg("failed linking subscriber to contact/user")

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

// orgUserIDByEmail returns the id of a user who is a member of the organization and whose
// email matches, or empty when no user matches; duplicates resolve best-effort to the first id
func orgUserIDByEmail(ctx context.Context, client *entgen.Client, orgID, email string) (string, error) {
	userID, err := client.User.Query().
		Where(
			user.EmailEqualFold(email),
			user.HasOrgMembershipsWith(orgmembership.OrganizationID(orgID)),
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
