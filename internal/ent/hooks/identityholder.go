package hooks

import (
	"context"
	"strings"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/pkg/logx"
	pkgobjects "github.com/theopenlane/core/v2/pkg/objects"
)

// skipOpenlaneUserAssignmentKey is used to denote a marker to skip processing in the
// hook and just return immediately. this is only used because we update the identity holder
// while inside the hook still
type skipOpenlaneUserAssignmentKey struct{}

// HookAssignOpenlaneUser automatically (un)sets the is_openlane_user based off
// the existence of the user in the org.
// this hook only really exists as a form of support for existing org members so that when new identity holders
// are created/updated for existing org members, we add them in correctly
func HookAssignOpenlaneUser() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.IdentityHolderFunc(func(ctx context.Context, m *generated.IdentityHolderMutation) (generated.Value, error) {

			if shouldSkip, _ := ctx.Value(skipOpenlaneUserAssignmentKey{}).(bool); shouldSkip {
				return next.Mutate(ctx, m)
			}

			var emails []string
			var ids []string

			email, _ := m.Email()

			switch m.Op() {
			case ent.OpCreate:
				emails = append(emails, email)

			case ent.OpUpdateOne:

				if strings.TrimSpace(email) == "" {

					var err error
					email, err = m.OldEmail(ctx)
					if err != nil {
						return nil, err
					}
				}

				emails = append(emails, email)

				id, _ := m.ID()
				ids = append(ids, id)

			case ent.OpUpdate:

				var err error

				ids, err = m.IDs(ctx)
				if err != nil {
					return nil, err
				}

				holders, err := m.Client().IdentityHolder.Query().
					Where(identityholder.IDIn(ids...)).
					Select(identityholder.FieldEmail).
					All(ctx)
				if err != nil {
					return nil, err
				}

				emails = lo.Map(holders, func(u *generated.IdentityHolder, _ int) string {
					return u.Email
				})
			}

			v, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			if m.Op().Is(ent.OpCreate) {
				ids = []string{v.(*generated.IdentityHolder).ID}
			}

			if len(ids) == 0 {
				return v, nil
			}

			ctx = context.WithValue(ctx, skipOpenlaneUserAssignmentKey{}, true)

			userMatchingEmails, err := m.Client().OrgMembership.Query().
				Where(
					orgmembership.HasUserWith(user.EmailIn(emails...))).
				QueryUser().
				Select(user.FieldEmail).
				Strings(ctx)
			if err != nil {
				return nil, err
			}

			if len(userMatchingEmails) > 0 {

				err := m.Client().IdentityHolder.Update().
					Where(identityholder.IDIn(ids...),
						identityholder.EmailIn(userMatchingEmails...)).
					SetIsOpenlaneUser(true).
					Exec(ctx)

				if err != nil {
					return nil, err
				}
			}

			err = m.Client().IdentityHolder.Update().
				Where(identityholder.IDIn(ids...),
					identityholder.EmailNotIn(userMatchingEmails...)).
				SetIsOpenlaneUser(false).
				Exec(ctx)
			if err != nil {
				return nil, err
			}

			if val, ok := v.(*generated.IdentityHolder); ok {
				_, exists := lo.Find(userMatchingEmails, func(e string) bool {
					return e == val.Email
				})

				val.IsOpenlaneUser = exists
			}

			return v, nil
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookIdentityHolderFiles runs on identity holder mutations to check for uploaded files
func HookIdentityHolderFiles() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.IdentityHolderFunc(func(ctx context.Context, m *generated.IdentityHolderMutation) (generated.Value, error) {
			fileIDs := pkgobjects.GetFileIDsFromContext(ctx)
			if len(fileIDs) > 0 {
				var err error

				ctx, err = pkgobjects.ProcessFilesForMutation(ctx, m, "identityHolderFiles")
				if err != nil {
					return nil, err
				}

				m.AddFileIDs(fileIDs...)
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookIdentityHolderSoftDelete clears identity_holder_id on linked directory accounts
// when an identity holder is soft-deleted, preventing stale foreign key references
func HookIdentityHolderSoftDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.IdentityHolderFunc(func(ctx context.Context, m *generated.IdentityHolderMutation) (generated.Value, error) {
			if !entx.CheckIsSoftDeleteType(ctx, m.Type()) {
				return next.Mutate(ctx, m)
			}

			holderID, ok := m.ID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			if clearErr := m.Client().DirectoryAccount.Update().
				Where(directoryaccount.IdentityHolderID(holderID)).
				ClearIdentityHolderID().
				Exec(ctx); clearErr != nil {
				logx.FromContext(ctx).Error().Err(clearErr).Str("identity_holder_id", holderID).Msg("failed to clear identity_holder_id on directory accounts after identity holder soft-delete")

				return retVal, clearErr
			}

			return retVal, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne)
}
