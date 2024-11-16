package main

import (
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/pkg/entitlements"
)

func main() {
	// Get the list of plans
	client := initStripeClient()

	plans := client.GetProducts()

	// Write the plans to a YAML file
	err := entitlements.WritePlansToYAML(plans, "plans.yaml")
	if err != nil {
		log.Fatal().Msgf("failed to write plans to YAML file: %v", err)
	}

	log.Info().Msgf("Plans written to plans.yaml successfully")
}

func initStripeClient() entitlements.StripeClient {
	client := entitlements.NewStripeClient(entitlements.WithAPIKey(""))
	return *client
}
