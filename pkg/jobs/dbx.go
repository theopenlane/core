package jobs

import (
	"context"
	"fmt"

	dbx "github.com/theopenlane/dbx/pkg/dbxclient"
	dbxenums "github.com/theopenlane/dbx/pkg/enums"

	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
)

type DBxArgs struct {
	OrganizationID string `json:"organization_id"`
	GeoLocation    string `json:"geo_location"`
}

func (DBxArgs) Kind() string { return "dbx" }

type DBxWorker struct {
	river.WorkerDefaults[DBxArgs]

	Config dbx.Config
}

func validateInput(job *river.Job[DBxArgs]) error {
	if job.Args.OrganizationID == "" {
		return fmt.Errorf("missing organization id") // nolint:goerr113
	}

	if job.Args.GeoLocation == "" {
		return fmt.Errorf("missing geo location") // nolint:goerr113
	}

	return nil
}

func (w *DBxWorker) Work(ctx context.Context, job *river.Job[DBxArgs]) error {
	// if its not enabled, return early
	if !w.Config.Enabled {
		return nil
	}

	log.Info().Msg("creating dbx database")

	input := dbx.CreateDatabaseInput{
		OrganizationID: job.Args.OrganizationID,
		Geo:            &job.Args.GeoLocation,
		Provider:       &dbxenums.Local, // todo: change this to the actual provider later
	}

	log.Debug().
		Str("org", input.OrganizationID).
		Str("geo", *input.Geo).
		Str("provider", input.Provider.String()).
		Msg("creating database")

	client := w.Config.NewDefaultClient()

	if _, err := client.CreateDatabase(ctx, input); err != nil {
		log.Error().Err(err).Msg("failed to create database")

		return err
	}

	return nil
}
