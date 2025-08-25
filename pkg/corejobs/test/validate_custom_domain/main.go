//go:build ignore

package main

import (
	"context"
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/riverboat/test/common"
)

// the main function here will insert a validate_custom_domain job into the river
// this will be picked up by the river server and processed
func main() {
	client := common.NewInsertOnlyRiverClient()

	customDomainID := flag.String("custom-domain-id", "", "ID of the custom domain")

	flag.Parse()

	_, err := client.Insert(context.Background(), corejobs.ValidateCustomDomainArgs{CustomDomainID: *customDomainID}, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("error inserting validate_custom_domain job")
	}

	log.Info().Msg("validate_custom_domain job successfully inserted")
}
