package model

import (
	_ "embed"
	"encoding/json"
	"log"
	"maps"
	"sort"
	"strings"
	"sync"

	openfga "github.com/openfga/go-sdk"
	language "github.com/openfga/language/pkg/go/transformer"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// relationPartsCount is the expected number of parts when splitting a relation like "can_view_object"
	relationPartsCount = 3
	// scopePartsCount is the expected number of parts when splitting a scope like "write:control"
	scopePartsCount = 2
)

//go:embed generated/crud.fga
var embeddedCrudModel []byte

//go:embed roles/roles.fga
var embeddedRolesModel []byte

var (
	// CanView allows read-only access to an object
	CanView = "can_view"
	// CanEdit allows read and write access to an object
	CanEdit = "can_edit"
	// CanDelete allows deletion of an object
	CanDelete = "can_delete"
)

var (
	// Read is an alias for can_view
	Read = "read"
	// Write is an alias for can_edit
	Write = "write"
	// Delete is an alias for can_delete
	Delete = "delete"
)

var (
	aliasToRelation = map[string]string{
		"read":   CanView,
		"write":  CanEdit,
		"delete": CanDelete,
	}

	crudOnce  sync.Once
	crudModel *openfga.AuthorizationModel
	crudErr   error

	rolesOnce  sync.Once
	rolesModel *openfga.AuthorizationModel
	rolesErr   error
)

func parseAuthorizationModel(embeddedModel []byte) (*openfga.AuthorizationModel, error) {
	protoModel, err := language.TransformDSLToProto(string(embeddedModel))
	if err != nil {
		return nil, errors.Wrap(err, "parse fga model dsl")
	}
	rawJSON, err := protojson.Marshal(protoModel)
	if err != nil {
		return nil, errors.Wrap(err, "marshal fga model")
	}
	var model openfga.AuthorizationModel
	if err := json.Unmarshal(rawJSON, &model); err != nil {
		return nil, errors.Wrap(err, "decode fga model json")
	}
	return &model, nil
}

// GetAuthorizationModel returns the parsed embedded authorization model
func GetCrudAuthorizationModel() (*openfga.AuthorizationModel, error) {
	crudOnce.Do(func() {
		crudModel, crudErr = parseAuthorizationModel(embeddedCrudModel)
	})
	return crudModel, crudErr
}

func GetRolesAuthorizationModel() (*openfga.AuthorizationModel, error) {
	rolesOnce.Do(func() {
		rolesModel, rolesErr = parseAuthorizationModel(embeddedRolesModel)
	})
	return rolesModel, rolesErr
}

// RelationsForService returns relations shaped like can_<verb>_<object> that directly accept service subjects.
func RelationsForService() ([]string, error) {
	model, err := GetCrudAuthorizationModel()
	if err != nil {
		return nil, err
	}

	var relations []string

	for _, td := range model.GetTypeDefinitions() {
		if td.Metadata == nil || td.Metadata.Relations == nil {
			continue
		}

		for rel, meta := range *td.Metadata.Relations {
			parts := strings.SplitN(rel, "_", relationPartsCount)
			if len(parts) != relationPartsCount || parts[0] != "can" {
				continue
			}

			for _, ref := range meta.GetDirectlyRelatedUserTypes() {
				if ref.Type == "service" {
					relations = append(relations, rel)
					break
				}
			}
		}
	}

	sort.Strings(relations)

	return relations, nil
}

// getRelations is a helper that returns relations for a given verb (e.g., "manage" or "create") shaped like can_<verb>_<object>
func getRelations(embeddedModel []byte, relationType string, isCrud bool) ([]string, error) {
	var model *openfga.AuthorizationModel
	var err error

	if isCrud {
		model, err = GetCrudAuthorizationModel()
	} else {
		model, err = GetRolesAuthorizationModel()
	}

	if err != nil {
		return nil, err
	}

	var relations []string

	for _, td := range model.GetTypeDefinitions() {
		if td.Metadata == nil || td.Metadata.Relations == nil {
			continue
		}

		for rel := range *td.Metadata.Relations {
			if relationType == "manage" {
				log.Printf("rel: %s", rel)
			}

			parts := strings.SplitN(rel, "_", relationPartsCount)
			if len(parts) == relationPartsCount && parts[0] == "can" && parts[1] == relationType {
				relations = append(relations, rel)
			}
		}
	}

	sort.Strings(relations)

	return relations, nil
}

// roleRelations returns relations shaped like can_manage_<role> that indicates role management
func roleRelations() ([]string, error) {
	log.Printf("Parsing roles model for roleRelations")
	return getRelations(embeddedRolesModel, "manage", false)
}

// createRelations returns relations shaped like can_create_<object> that are used for group-based creation access
func createRelations() ([]string, error) {
	log.Printf("Parsing crud model for createRelations")
	return getRelations(embeddedCrudModel, "create", true)
}

// DefaultServiceScopeSet returns the default service scopes as a set
func DefaultServiceScopeSet() (map[string]struct{}, error) {
	scopes, err := RelationsForService()
	if err != nil {
		return nil, err
	}

	set := make(map[string]struct{}, len(scopes))
	for _, s := range scopes {
		set[s] = struct{}{}
	}

	return set, nil
}

// NormalizeScope returns the relation name for a provided scope, handling common aliases
// Accepts verb:object (e.g., write:control) and simple verbs (read/write/delete)
func NormalizeScope(scope string) string {
	raw := strings.TrimSpace(scope)
	if raw == "" {
		return ""
	}

	normalized := strings.ToLower(raw)

	mapVerb := func(verb string) string {
		if rel, ok := aliasToRelation[verb]; ok {
			return rel
		}

		return verb
	}

	if parts := strings.SplitN(normalized, ":", scopePartsCount); len(parts) == scopePartsCount && parts[1] != "" {
		return mapVerb(parts[0]) + "_" + parts[1]
	}

	if rel := mapVerb(normalized); rel != "" {
		return rel
	}

	return normalized
}

// ScopeAliases returns a copy of the supported alias mapping
func ScopeAliases() map[string]string {
	aliases := make(map[string]string, len(aliasToRelation))
	maps.Copy(aliases, aliasToRelation)

	return aliases
}

// ScopeOptions groups available scopes by object (verb mapped back via alias map)
func ScopeOptions() (map[string][]string, error) {
	rels, err := RelationsForService()
	if err != nil {
		return nil, err
	}

	relToVerb := map[string]string{}
	for verb, rel := range aliasToRelation {
		relToVerb[rel] = verb
	}

	opts := make(map[string][]string)

	for _, rel := range rels {
		parts := strings.SplitN(rel, "_", relationPartsCount)
		if len(parts) != relationPartsCount || parts[0] != "can" {
			continue
		}

		verb, ok := relToVerb[strings.Join(parts[0:2], "_")]
		if !ok {
			continue
		}

		obj := parts[2]
		if obj == "" {
			continue
		}

		opts[obj] = append(opts[obj], verb)
	}

	for obj := range opts {
		sort.Strings(opts[obj])
	}

	return opts, nil
}

func getRelationsOptionsForObject(rels []string) ([]string, error) {
	objs := make([]string, 0, len(rels))
	for _, rel := range rels {
		parts := strings.SplitN(rel, "_", relationPartsCount)

		obj := parts[2]
		if obj == "" {
			continue
		}

		objs = append(objs, obj)
	}

	sort.Strings(objs)

	return objs, nil
}

// CreateOptions returns objects with verbs that support creation
func CreateOptions() ([]string, error) {
	rels, err := createRelations()
	if err != nil {
		return nil, err
	}

	return getRelationsOptionsForObject(rels)
}

// RoleOptions returns objects with verbs that support roles
func RoleOptions() ([]string, error) {
	rels, err := roleRelations()
	if err != nil {
		return nil, err
	}

	if len(rels) == 0 {
		log.Fatal("no relations found for role management - ensure the embedded model contains relations shaped like can_manage_<role>")
	}

	return getRelationsOptionsForObject(rels)
}
