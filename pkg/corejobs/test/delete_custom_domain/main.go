//go:build ignore

package main

import (
	"context"
	"flag"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/riverboat/test/common"
)

// the main function here will insert a delete_custom_domain job into the river
// this will be picked up by the river server and processed
func main() {
	client := common.NewInsertOnlyRiverClient()

	// Parse command line arguments
	customDomainID := flag.String("custom-domain-id", "", "ID of the custom domain")
	dnsVerificationID := flag.String("dns-verification-id", "", "ID of the DNS verification")
	cfCustomHostnameID := flag.String("cf-hostname-id", "", "ID of the Cloudflare custom hostname")
	cfZoneID := flag.String("cf-zone-id", "", "ID of the Cloudflare zone")
	flag.Parse()

	_, err := client.Insert(context.Background(), corejobs.DeleteCustomDomainArgs{
		CustomDomainID:             *customDomainID,
		DNSVerificationID:          *dnsVerificationID,
		CloudflareCustomHostnameID: *cfCustomHostnameID,
		CloudflareZoneID:           *cfZoneID,
	}, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("error inserting delete_custom_domain job")
	}

	log.Info().Msg("delete_custom_domain job successfully inserted")
}
