package entitlements

import (
	"context"

	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/shared/models"
)

const (
	subjectType     = "organization"
	TupleObjectType = "feature"
	TupleRelation   = "enabled"
)

var baseTupleRequest = fgax.TupleRequest{
	SubjectType: subjectType,
	ObjectType:  TupleObjectType,
	Relation:    TupleRelation,
}

// DeleteModuleTuple removes the enabled feature from the organization in the authorization service
func DeleteModuleTuple(ctx context.Context, authz *fgax.Client, orgID, moduleName string) error {
	deleteTuple := getFeatureTupleKey(orgID, moduleName)

	_, err := authz.WriteTupleKeys(ctx, nil, []fgax.TupleKey{deleteTuple})

	return err
}

// CreateFeatureTuples writes default feature tuples to FGA and inserts them into
// the feature cache if available.
func CreateFeatureTuples(ctx context.Context, authz *fgax.Client, orgID string, feats []models.OrgModule) error {
	tuples := make([]fgax.TupleKey, 0, len(feats))

	for _, f := range feats {
		tuples = append(tuples, getFeatureTupleKey(orgID, f.String()))
	}

	if _, err := authz.WriteTupleKeys(ctx, tuples, nil); err != nil {
		return err
	}

	return nil
}

func getFeatureTupleKey(orgID, module string) fgax.TupleKey {
	tuple := baseTupleRequest

	tuple.SubjectID = orgID
	tuple.ObjectID = module

	return fgax.GetTupleKey(tuple)
}

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
