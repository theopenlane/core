package entitlements

import (
	"fmt"
)

type TupleStruct struct {
	User     string `yaml:"user"`
	Relation string `yaml:"relation"`
	Object   string `yaml:"object"`
}

func (sc *StripeClient) CreateTupleStruct() []TupleStruct {
	products := sc.GetAllProductPricesMapped()

	tuples := []TupleStruct{}
	// Iterate over products and features to create the tuple struct
	for _, product := range products {
		for _, feature := range product.Features {
			// Create a tuple for each feature
			tuple := TupleStruct{
				User:     fmt.Sprintf("plan:%s", product.Name),
				Relation: "associated_plan",
				Object:   fmt.Sprintf("feature:%s", feature.Lookupkey),
			}

			// Print the tuple
			fmt.Printf("User: %s, Relation: %s, Object: %s\n", tuple.User, tuple.Relation, tuple.Object)
			// Append the tuple to the list
			tuples = append(tuples, tuple)
		}
	}

	return tuples
}
