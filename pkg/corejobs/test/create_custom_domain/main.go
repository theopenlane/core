package main

import (
	"context"
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/riverboat/test/common"
)

// the main function here will insert a create_custom_domain job into the river
// this will be picked up by the river server and processed
func main() {
	client := common.NewInsertOnlyRiverClient()

	// Parse command line arguments
	customDomainID := flag.String("custom-domain-id", "", "ID of the custom domain")
	flag.Parse()

	_, err := client.Insert(
		context.Background(),
		corejobs.CreateCustomDomainArgs{
			CustomDomainID: *customDomainID,
		},
		nil,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("error inserting create_custom_domain job")
	}

	log.Info().Msg("create_custom_domain job successfully inserted")
}
