package main

import (
	"context"
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/riverboat/test/common"
)

// the main function here will insert a job into the river to export contents
// this will be picked up by the river server and processed
func main() {
	client := common.NewInsertOnlyRiverClient()

	contentID := flag.String("id", "", "ID of the content to export")

	flag.Parse()

	_, err := client.Insert(context.Background(), corejobs.ExportContentArgs{
		ExportID: *contentID,
	}, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("error inserting the job to create export")
	}

	log.Info().Msg("create export job successfully inserted")
}
