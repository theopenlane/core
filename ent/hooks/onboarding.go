package hooks

import (
	"context"
	"regexp"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/utils/keygen"

	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/ent/generated/organization"
	"github.com/theopenlane/ent/generated/privacy"
)

// HookOnboarding runs on onboarding mutations to create the organization and settings
func HookOnboarding() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.OnboardingFunc(func(ctx context.Context, m *generated.OnboardingMutation) (generated.Value, error) {
			// create organization with company name and domains in settings
			companyName, _ := m.CompanyName()
			domains, domainsOK := m.Domains()

			input := generated.CreateOrganizationInput{
				Name:        companyName,
				DisplayName: &companyName,
			}

			org, err := createOrgUniqueName(ctx, m, input, 0)
			if err != nil {
				return nil, err
			}

			// set the organization ID on the mutation
			m.SetOrganizationID(org.ID)

			// update the organization settings with the domains, if provided
			if domainsOK && len(domains) > 0 {
				setting, err := org.Setting(ctx)
				if err != nil {
					return nil, err
				}

				settingInput := generated.UpdateOrganizationSettingInput{
					Domains: domains,
				}

				err = m.Client().OrganizationSetting.UpdateOneID(setting.ID).SetInput(settingInput).Exec(ctx)
				if err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// createOrgUniqueName creates an organization with a unique name by appending a random string to the provided name
// if the organization already exists, it will retry up to 10 times
func createOrgUniqueName(ctx context.Context, m *generated.OnboardingMutation, input generated.CreateOrganizationInput, attempt int) (*generated.Organization, error) {
	const maxRetries = 10

	if attempt > maxRetries {
		return nil, ErrMaxAttemptsOrganization
	}

	// check for the existence of the organization with the given name
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	exists, err := m.Client().Organization.Query().Where(
		organization.Name(input.Name),
		organization.DeletedAtIsNil(),
	).Exist(allowCtx)
	if err != nil {
		return nil, err
	}

	if exists {
		input.Name = uniqueOrganizationName(input.Name)
		attempt++

		return createOrgUniqueName(ctx, m, input, attempt)
	}

	org, err := m.Client().Organization.Create().SetInput(input).Save(ctx)
	if err != nil {
		return nil, err
	}

	return org, nil
}

// alphaNumericDashRegex is regex to remove all non-alphanumeric characters except for dashes
var alphaNumericDashRegex = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

// uniqueOrganizationName generates a unique organization name by appending a random string to the provided name
// if we need a new name, we are going to covert it to a lowercase alphanumeric string
// e.g. Acme Corp, Inc. -> acme-corp-inc-abc123
func uniqueOrganizationName(name string) string {
	// trim trailing whitespace from the name
	name = strings.TrimSpace(name)
	// replace spaces with dashes
	name = strings.ReplaceAll(name, " ", "-")

	return strings.ToLower(alphaNumericDashRegex.ReplaceAllString(name, "")) + "-" + keygen.AlphaNumeric(6) //nolint:mnd
}
