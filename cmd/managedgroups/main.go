//go:build climanagedgroups

package main

import (
	"context"
	"fmt"
	"os"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/config"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	_ "github.com/theopenlane/core/internal/ent/generated/runtime"
	"github.com/theopenlane/core/internal/entdb"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/core/pkg/logx"
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
			&cli.StringSliceFlag{
				Name:  "orgs",
				Usage: "specific organization IDs to process (if not set, all non-personal orgs are processed)",
			},
		},
		Action: reconcileManagedGroups,
	}
}

func createDB(c *cli.Command) (*generated.Client, error) {
	cfgLoc := c.Root().String("config")

	cfg, err := config.Load(&cfgLoc)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err) //nolint:err113
	}

	if cfg.JobQueue.ConnectionURI == "" {
		return nil, fmt.Errorf("missing required job queue connection URI in config") //nolint:err113
	}

	jobOpts := []riverqueue.Option{
		riverqueue.WithConnectionURI(cfg.JobQueue.ConnectionURI),
	}

	ctx := context.Background()

	fgaClient, err := fgax.CreateFGAClientWithStore(ctx, cfg.Authz)
	if err != nil {
		return nil, fmt.Errorf("failed to create FGA client: %w", err) //nolint:err113
	}

	entOpts := []generated.Option{
		generated.Authz(*fgaClient),
	}

	dbClient, err := entdb.New(ctx, cfg.DB, jobOpts, entOpts...)
	if err != nil {
		return nil, fmt.Errorf("database client: %w", err) //nolint:err113
	}

	return dbClient, nil
}

func reconcileManagedGroups(ctx context.Context, c *cli.Command) error {
	// setup logging
	setupLogging(c.Bool("debug"))

	// allow this internal tool to bypass privacy checks
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	db, err := createDB(c)
	if err != nil {
		return err
	}

	orgIDs := c.StringSlice("orgs")
	where := []predicate.Organization{
		organization.DeletedAtIsNil(),
		// only process non-personal organizations
		organization.PersonalOrg(false),
	}

	if len(orgIDs) > 0 {
		where = append(where, organization.IDIn(orgIDs...))
	}

	orgs, err := db.Organization.Query().
		Where(
			where...,
		).
		All(ctx)
	if err != nil {
		return err
	}

	if len(orgs) == 0 {
		log.Info().Msg("no non-personal organizations found, nothing to do")
		return nil
	}

	dryRun := c.Root().Bool("dry-run")
	createdGroups := 0
	addedMemberships := 0

	msg := "starting managed groups reconciliation"
	if dryRun {
		msg = "[DRY-RUN] " + msg
	}
	log.Info().Int("org_count", len(orgs)).Msg(msg)

	for _, org := range orgs {
		log.Info().Str("org_id", org.ID).Str("org_name", org.Name).Msg("processing organization")

		memberships, err := db.OrgMembership.Query().
			Where(orgmembership.OrganizationID(org.ID)).
			All(ctx)
		if err != nil {
			return err
		}

		log.Debug().Str("org", org.Name).Int("memberships", len(memberships)).Msg("fetched organization memberships")

		for _, m := range memberships {
			log.Debug().Str("user_id", m.UserID).Str("org_id", m.OrganizationID).Msg("processing org membership")

			newCtx := auth.WithAuthenticatedUser(ctx, &auth.AuthenticatedUser{
				SubjectID:       m.UserID,
				OrganizationID:  m.OrganizationID,
				OrganizationIDs: []string{m.OrganizationID},
			})

			user, err := db.User.Get(newCtx, m.UserID)
			if err != nil {
				return err
			}

			log.Debug().Str("user_id", user.ID).Str("display_name", user.DisplayName).Msg("fetching user for org membership")

			existing, err := db.Group.Query().
				Where(
					group.OwnerID(org.ID),
					group.IsManaged(true),
					group.DisplayNameEqualFold(user.DisplayName),
					func(s *sql.Selector) {
						s.Where(
							sqljson.ValueContains(group.FieldTags, user.ID),
						)
					},
				).
				Only(newCtx)
			if err != nil && !generated.IsNotFound(err) {
				return err
			}

			if existing == nil {
				log.Info().Str("user_id", user.ID).Str("display_name", user.DisplayName).Str("org_id", org.ID).Msg("no existing managed group found for user, creating")

				if dryRun {
					log.Info().Str("user_id", user.ID).Str("display_name", user.DisplayName).Str("org_id", org.ID).Msg("[DRY-RUN] create managed group for user")
				} else {
					groupName := fmt.Sprintf("%s - %s", user.DisplayName, user.ID)

					desc := fmt.Sprintf("Group for %s", user.DisplayName)

					g, err := db.Group.Create().
						SetInput(generated.CreateGroupInput{
							Name:        groupName,
							DisplayName: &user.DisplayName,
							Description: &desc,
							Tags:        []string{"managed", user.DisplayName, user.ID},
						}).
						SetIsManaged(true).
						SetOwnerID(org.ID).
						Save(newCtx)
					if err != nil {
						return err
					}

					existing = g
					createdGroups++

					log.Info().Str("group_id", g.ID).Str("user_id", user.ID).Str("org_id", org.ID).Msg("created managed group for user")
				}
			} else {
				log.Debug().Str("group_name", existing.Name).Str("user_id", user.DisplayName).Str("org_id", org.ID).Msg("found existing managed group for user")
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
						log.Info().Str("group_id", existing.ID).Str("group_name", existing.Name).Str("user_id", user.ID).Str("org_id", org.ID).Msg("[DRY-RUN] add user to managed group")
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
						log.Info().Str("group_id", existing.ID).Str("group_name", existing.Name).Str("user_id", user.ID).Str("org_id", org.ID).Msg("added user to managed group")
					}
				}
			}
		}
	}

	log.Info().Int("created_groups", createdGroups).Int("added_memberships", addedMemberships).Msg("managed groups reconciliation complete")

	if dryRun {
		log.Info().Msg("[DRY-RUN] No changes were persisted.")
	}

	return nil
}

func setupLogging(debug bool) {
	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}

	logx.Configure(logx.LoggerConfig{
		Level:     level,
		Pretty:    true,
		Writer:    os.Stderr,
		SetGlobal: true,
	})
}
