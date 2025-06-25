package entitlements

import (
	"context"
	"fmt"

	"github.com/theopenlane/iam/fgax"
)

// TupleStruct represents a tuple of user, relation, and object
type TupleStruct struct {
	User     string `yaml:"user"`
	Relation string `yaml:"relation"`
	Object   string `yaml:"object"`
}

// CreateTupleStruct creates a tuple struct for each feature of each product
func (sc *StripeClient) CreateTupleStruct(ctx context.Context) []TupleStruct {
	products := sc.GetAllProductPricesMapped(ctx)

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

// SyncTuples updates openFGA tuples for a given subject and object type/relation.
// It adds tuples for items in 'newItems' not in 'oldItems', and deletes tuples for items in 'oldItems' not in 'newItems'.
// This function is generic and can be reused for different subject/object types and relations.
func SyncTuples(
	ctx context.Context,
	client *fgax.Client,
	subjectID, subjectType, objectType, relation string,
	oldItems, newItems []string,
) error {
	if client == nil {
		return nil
	}

	addMap := make(map[string]struct{}, len(newItems))
	for _, item := range newItems {
		addMap[item] = struct{}{}
	}
	for _, item := range oldItems {
		delete(addMap, item)
	}

	delMap := make(map[string]struct{}, len(oldItems))
	for _, item := range oldItems {
		delMap[item] = struct{}{}
	}
	for _, item := range newItems {
		delete(delMap, item)
	}

	adds := make([]fgax.TupleKey, 0, len(addMap))
	for item := range addMap {
		adds = append(adds, fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   subjectID,
			SubjectType: subjectType,
			ObjectID:    item,
			ObjectType:  objectType,
			Relation:    relation,
		}))
	}

	dels := make([]fgax.TupleKey, 0, len(delMap))
	for item := range delMap {
		dels = append(dels, fgax.GetTupleKey(fgax.TupleRequest{
			SubjectID:   subjectID,
			SubjectType: subjectType,
			ObjectID:    item,
			ObjectType:  objectType,
			Relation:    relation,
		}))
	}

	if len(adds) == 0 && len(dels) == 0 {
		return nil
	}

	_, err := client.WriteTupleKeys(ctx, adds, dels)
	return err
}
