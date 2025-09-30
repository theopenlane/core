package main

import (
	"context"
	"fmt"
	"os"

	"github.com/theopenlane/core/config"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/riverboat/pkg/riverqueue"
	"github.com/urfave/cli/v3"
)

func main() {
	if err := app().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func app() *cli.Command {
	return &cli.Command{
		Name:  "managedgroups",
		Usage: "Ensure each org member has a managed group named after them in the org",
		Description: `

Examples:
  # Dry-run by default (preview changes without making them)
  managedgroups

  # Actually make changes
  managedgroups --dry-run=false

  # Use custom config
  managedgroups --config /path/to/config.yaml orgs`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Value:   "./config/.config.yaml",
				Usage:   "config file path",
				Sources: cli.EnvVars("CORE_CONFIG"),
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "preview changes without making them (global default)",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "enable debug logging",
			},
		},
		Action: reconcileManagedGroups,
	}
}

func createDB(c *cli.Command) (*generated.Client, error) {
	cfgLoc := c.Root().String("config")
	cfg, err := config.Load(&cfgLoc)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if cfg.JobQueue.ConnectionURI == "" {
		return nil, fmt.Errorf("missing required job queue connection URI in config")
	}

	jobOpts := []riverqueue.Option{
		riverqueue.WithConnectionURI(cfg.JobQueue.ConnectionURI),
	}

	ctx := context.Background()

	fgaClient, err := fgax.CreateFGAClientWithStore(ctx, cfg.Authz)
	if err != nil {
		return nil, fmt.Errorf("failed to create FGA client: %w", err)
	}

	entOpts := []generated.Option{
		generated.Authz(*fgaClient),
	}

	dbClient, err := entdb.New(ctx, cfg.DB, jobOpts, entOpts...)
	if err != nil {
		return nil, fmt.Errorf("database client: %w", err)
	}
	return dbClient, nil
}

func reconcileManagedGroups(ctx context.Context, c *cli.Command) error {
	// allow this internal tool to bypass privacy checks
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	db, err := createDB(c)
	if err != nil {
		return err
	}

	orgs, err := db.Organization.Query().
		Where(
			organization.DeletedAtIsNil(),
			// only process non-personal organizations
			organization.PersonalOrg(false),
		).
		All(ctx)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		fmt.Println("No non-personal organizations found")
		return nil
	}

	fmt.Printf("Found %d non-personal organizations to process\n", len(orgs))

	dryRun := c.Root().Bool("dry-run")
	createdGroups := 0
	addedMemberships := 0

	for _, org := range orgs {
		fmt.Printf("Processing organization: %s (%s)\n", org.Name, org.ID)

		memberships, err := db.OrgMembership.Query().
			Where(orgmembership.OrganizationID(org.ID)).
			All(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("Found %d memberships in org %s\n", len(memberships), org.Name)

		for _, m := range memberships {
			fmt.Printf("Processing user membership: %s in org %s\n", m.UserID, m.OrganizationID)

			newCtx := auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{
				SubjectID:       m.UserID,
				OrganizationID:  m.OrganizationID,
				OrganizationIDs: []string{m.OrganizationID},
			})

			user, err := db.User.Get(newCtx, m.UserID)
			if err != nil {
				return err
			}

			fmt.Printf("User: %s (display name: %s)\n", user.ID, user.DisplayName)

			existing, err := db.Group.Query().
				Where(
					group.OwnerID(org.ID),
					group.IsManaged(true),
					group.Name(user.DisplayName),
				).
				Only(newCtx)
			if err != nil && !generated.IsNotFound(err) {
				return err
			}

			if existing == nil {
				fmt.Printf("No existing managed group found for user '%s' in org '%s'\n", user.DisplayName, org.ID)
				if dryRun {
					fmt.Printf("[DRY-RUN] create managed group '%s' in org '%s' for user '%s'\n", user.DisplayName, org.ID, user.ID)
				} else {
					desc := user.DisplayName
					g, err := db.Group.Create().
						SetInput(generated.CreateGroupInput{
							Name:        user.DisplayName,
							Description: &desc,
							Tags:        []string{"managed", user.DisplayName},
						}).
						SetIsManaged(true).
						SetOwnerID(org.ID).
						Save(newCtx)
					if err != nil {
						return err
					}
					existing = g
					createdGroups++
				}
			} else {
				fmt.Printf("Found existing managed group '%s' for user '%s' in org '%s'\n", existing.Name, user.DisplayName, org.ID)
			}

			if existing != nil {
				// ensure membership exists
				cnt, err := db.GroupMembership.Query().
					Where(
						groupmembership.GroupID(existing.ID),
						groupmembership.UserID(user.ID),
					).
					Count(newCtx)
				if err != nil {
					return err
				}

				if cnt == 0 {
					if dryRun {
						fmt.Printf("[DRY-RUN] add user '%s' to managed group '%s' in org '%s'\n", user.ID, existing.Name, org.ID)
					} else {
						role := enums.RoleMember
						if err := db.GroupMembership.Create().
							SetInput(generated.CreateGroupMembershipInput{
								Role:    &role,
								UserID:  user.ID,
								GroupID: existing.ID,
							}).
							Exec(newCtx); err != nil {
							return err
						}
						addedMemberships++
					}
				}
			}
		}
	}

	fmt.Printf("Done. Created %d groups and added %d memberships\n", createdGroups, addedMemberships)
	if dryRun {
		fmt.Println("(dry-run) No changes were persisted.")
	}

	return nil
}
