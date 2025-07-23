package main

import (
	"context"
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/riverboat/test/common"
)

// the main function here will insert a delete_export_content job into the river
// this will be picked up by the river server and processed
func main() {
	client := common.NewInsertOnlyRiverClient()

	flag.Parse()

	_, err := client.Insert(context.Background(), corejobs.DeleteExportContentArgs{}, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("error inserting the job to delete all exports")
	}

	log.Info().Msg("delete export content job successfully inserted")
}
