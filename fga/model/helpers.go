package fgamodel

import (
	_ "embed"
	"encoding/json"
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

//go:embed model.fga
var embeddedModel []byte

var (
	aliasToRelation = map[string]string{
		"read":   "can_view",
		"write":  "can_edit",
		"delete": "can_delete",
	}

	parseOnce sync.Once
	parseErr  error
	parsed    *openfga.AuthorizationModel
)

// GetAuthorizationModel returns the parsed embedded authorization model
func GetAuthorizationModel() (*openfga.AuthorizationModel, error) {
	parseOnce.Do(func() {
		protoModel, err := language.TransformDSLToProto(string(embeddedModel))
		if err != nil {
			parseErr = errors.Wrap(err, "parse fga model dsl")
			return
		}

		rawJSON, err := protojson.Marshal(protoModel)
		if err != nil {
			parseErr = errors.Wrap(err, "marshal fga model")
			return
		}

		var model openfga.AuthorizationModel
		if err := json.Unmarshal(rawJSON, &model); err != nil {
			parseErr = errors.Wrap(err, "decode fga model json")
			return
		}

		parsed = &model
	})

	return parsed, parseErr
}

// RelationsForService returns relations shaped like can_<verb>_<object> that directly accept service subjects.
func RelationsForService() ([]string, error) {
	model, err := GetAuthorizationModel()
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
