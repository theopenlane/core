package entitlements

import (
	"context"

	"github.com/theopenlane/iam/fgax"
)

// SyncTuples updates openFGA tuples for a given subject and object type/relation.
// It adds tuples for items in 'newItems' not in 'oldItems', and deletes tuples for items in 'oldItems' not in 'newItems'.
// This function is generic and can be reused for different subject/object types and relations.
func SyncTuples(ctx context.Context, client *fgax.Client, subjectID, subjectType, objectType, relation string, oldItems, newItems []string,
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
