//go:build ignore

package main

import (
	"context"
	"sort"

	"github.com/rs/zerolog/log"

	"github.com/theopenlane/shared/entitlements"
)

// main is a test script to fetch plans from Stripe and write them to a YAML file
func main() {
	ctx := context.Background()

	// Get the list of plans
	client := initStripeClient()

	plans := client.GetAllProductPricesMapped(ctx)

	sortPlansByFeatureCount(plans)

	// Write the plans to a YAML file
	err := entitlements.WritePlansToYAML(plans, "pkg/entitlements/test/plans.yaml")
	if err != nil {
		log.Fatal().Msgf("failed to write plans to YAML file: %v", err)
	}

	log.Info().Msgf("Plans written to plans.yaml successfully")
}

func initStripeClient() entitlements.StripeClient {
	client, err := entitlements.NewStripeClient(entitlements.WithAPIKey(""))
	if err != nil {
		log.Fatal().Msgf("failed to initialize Stripe client: %v", err)
	}

	return *client
}

// sortPlansByFeatureCount sorts the plans by the count of features in each plan
func sortPlansByFeatureCount(plans []entitlements.Product) {
	sort.Slice(plans, func(i, j int) bool {
		return len(plans[i].Features) < len(plans[j].Features)
	})
}
