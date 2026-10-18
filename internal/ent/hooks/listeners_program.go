package hooks

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// init registers the program listeners
func init() { registerListeners(ProgramListeners) }

// ProgramListeners is a list a mutation handlers that concerns a program
func ProgramListeners() []gala.Registration {
	return []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaProgram,
			Operations: []string{entityops.OpUpdate, entityops.OpUpdateOne},
			Fields:     []string{program.FieldAuditorReady},
			RowMatch: []entityops.FieldMatch{{
				Field: program.FieldAuditorReady,
				In:    []string{"true"},
			}},
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation)
			},
			Handle: handleProgramMutation,
		},
	}
}

func handleProgramMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	program, ok, err := entityops.LoadEntity(inv.Context, payload.EntityID, inv.Client.Program.Get)
	switch {
	case err != nil:
		return fmt.Errorf("failed to query program: %w", err)
	case !ok:
		return nil
	}

	email := strings.TrimSpace(program.AuditorEmail)
	if email == "" {
		return nil
	}

	exists, err := inv.Client.OrgMembership.Query().
		Where(
			orgmembership.OrganizationID(program.OwnerID),
			orgmembership.HasUserWith(user.EmailEqualFold(email)),
		).
		Exist(inv.Context)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	input := generated.CreateInviteInput{
		Recipient: email,
		Role:      lo.ToPtr(enums.RoleAuditor),
		OwnerID:   lo.ToPtr(program.OwnerID),
	}

	_, err = inv.Client.Invite.Create().SetInput(input).Save(inv.Context)
	return err
}
