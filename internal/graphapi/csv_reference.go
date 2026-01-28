package graphapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"sort"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/graphapi/common"
)

// CSV reference sentinel errors for field access.
var (
	// errCSVInvalidStringSliceField indicates a non-addressable or unsupported slice field.
	errCSVInvalidStringSliceField = errors.New("invalid field for string slice")
	// errCSVFieldNotStringSlice indicates the field is not a []string.
	errCSVFieldNotStringSlice = errors.New("field is not a []string")
	// errCSVFieldNotString indicates the field is not a string.
	errCSVFieldNotString = errors.New("field is not a string")
)

// csvReferenceRule describes how a CSV field maps to target IDs.
type csvReferenceRule struct {
	SourceField string
	TargetField string
	Lookup      func(ctx context.Context, values []string) (map[string]string, error)
	Create      func(ctx context.Context, values []string) (map[string]string, error)
	Normalize   func(string) string
}

// csvFieldKind describes supported csv target field shapes.
type csvFieldKind int

const (
	csvFieldKindUnknown csvFieldKind = iota
	csvFieldKindString
	csvFieldKindStringSlice
)

// csvRuleMeta captures json field names and target kind for a rule.
type csvRuleMeta struct {
	SourceJSONKey string
	TargetJSONKey string
	TargetInInput bool
	TargetKind    csvFieldKind
}

// csvInputFieldName is the wrapper field name used in CSV input structs.
const csvInputFieldName = "Input"

// resolveCSVReferencesForSchema resolves CSV references using the generated registry for a schema.
func resolveCSVReferencesForSchema(ctx context.Context, schemaName string, inputs any) error {
	rules, err := BuildCSVReferenceRulesFromRegistry(schemaName)
	if err != nil {
		return err
	}

	if len(rules) == 0 {
		return nil
	}

	return resolveCSVReferenceRules(ctx, inputs, rules...)
}

// csvRuleLookupCache caches resolved values keyed by lookup function pointers.
type csvRuleLookupCache struct {
	cache map[csvLookupCacheKey]map[string]string
}

// csvLookupCacheKey identifies a unique lookup+normalize function pair.
type csvLookupCacheKey struct {
	lookupPtr    uintptr
	normalizePtr uintptr
}

// newCSVRuleLookupCache creates an empty lookup cache.
func newCSVRuleLookupCache() *csvRuleLookupCache {
	return &csvRuleLookupCache{cache: make(map[csvLookupCacheKey]map[string]string)}
}

// get retrieves cached resolved values for a lookup key.
func (c *csvRuleLookupCache) get(key csvLookupCacheKey) map[string]string {
	if cached, ok := c.cache[key]; ok {
		return lo.Assign(map[string]string{}, cached)
	}
	return map[string]string{}
}

// set stores resolved values in the cache.
func (c *csvRuleLookupCache) set(key csvLookupCacheKey, resolved map[string]string) {
	if len(resolved) > 0 {
		c.cache[key] = resolved
	}
}

// resolveCSVReferenceRules resolves lookup rules and writes IDs into target fields on the inputs.
func resolveCSVReferenceRules(ctx context.Context, inputs any, rules ...csvReferenceRule) error {
	if len(rules) == 0 {
		return nil
	}

	slice := reflect.ValueOf(inputs)
	if slice.Kind() != reflect.Slice {
		return common.NewValidationError("csv inputs must be a slice")
	}

	elemType := slice.Type().Elem()
	for elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}

	rowStates, err := csvRowStatesFromSlice(slice)
	if err != nil {
		return err
	}

	lookupCache := newCSVRuleLookupCache()

	for _, rule := range rules {
		if err := resolveCSVReferenceRule(ctx, elemType, rowStates, rule, lookupCache); err != nil {
			return err
		}
	}

	return csvRowStatesToStructs(rowStates)
}

// resolveCSVReferenceRule processes a single rule: collects values, resolves them, and updates rows.
func resolveCSVReferenceRule(ctx context.Context, elemType reflect.Type, rowStates []*csvRowState, rule csvReferenceRule, cache *csvRuleLookupCache) error {
	meta, ok := csvRuleMetaForType(elemType, rule)
	if !ok {
		return nil
	}

	normalize := rule.Normalize
	if normalize == nil {
		normalize = normalizeCSVReferenceKey
	}

	values := collectCSVRuleValues(rowStates, meta.SourceJSONKey)
	unique := normalizeUniqueValues(values, normalize)
	if len(unique) == 0 {
		return nil
	}

	cacheKey := csvLookupCacheKey{
		lookupPtr:    reflect.ValueOf(rule.Lookup).Pointer(),
		normalizePtr: reflect.ValueOf(normalize).Pointer(),
	}

	resolved, err := resolveCSVRuleValues(ctx, rule, unique, cache.get(cacheKey))
	if err != nil {
		return err
	}

	cache.set(cacheKey, resolved)

	missing := missingCSVValues(unique, resolved)
	if len(missing) > 0 {
		fieldName := lowerCamelField(rule.SourceField)
		sort.Strings(missing)

		return common.NewValidationErrorWithFields(fmt.Sprintf("unable to resolve %s: %s", fieldName, strings.Join(missing, ", ")), fieldName)
	}

	return applyCSVRuleToRows(rowStates, meta, rule.TargetField, resolved, normalize)
}

// collectCSVRuleValues extracts all source values from row states for a given JSON key.
func collectCSVRuleValues(rowStates []*csvRowState, sourceKey string) []string {
	values := make([]string, 0, len(rowStates))
	for _, row := range rowStates {
		if row == nil {
			continue
		}
		values = append(values, csvMapFieldStrings(row.data, sourceKey)...)
	}
	return values
}

// resolveCSVRuleValues performs lookup and optional create to resolve all values to IDs.
func resolveCSVRuleValues(ctx context.Context, rule csvReferenceRule, unique map[string]string, resolved map[string]string) (map[string]string, error) {
	missing := missingCSVValues(unique, resolved)
	if len(missing) > 0 {
		lookedUp, err := rule.Lookup(ctx, missing)
		if err != nil {
			return nil, err
		}

		maps.Copy(resolved, lookedUp)
	}

	missing = missingCSVValues(unique, resolved)
	if len(missing) > 0 && rule.Create != nil {
		created, err := rule.Create(ctx, missing)
		if err != nil {
			return nil, err
		}

		maps.Copy(resolved, created)
	}

	return resolved, nil
}

// applyCSVRuleToRows writes resolved IDs into target fields on each row.
func applyCSVRuleToRows(rowStates []*csvRowState, meta csvRuleMeta, targetField string, resolved map[string]string, normalize func(string) string) error {
	targetPath := meta.targetPath()
	fieldName := lowerCamelField(targetField)

	for _, row := range rowStates {
		if row == nil {
			continue
		}

		ids := lo.FilterMap(csvMapFieldStrings(row.data, meta.SourceJSONKey), func(value string, _ int) (string, bool) {
			key := normalize(value)
			id, ok := resolved[key]
			return id, ok
		})

		if len(ids) == 0 {
			continue
		}

		if err := setCSVRuleTargetField(row.data, targetPath, meta.TargetKind, ids, fieldName); err != nil {
			return err
		}
	}

	return nil
}

// setCSVRuleTargetField sets the target field value based on its kind (string or slice).
func setCSVRuleTargetField(data map[string]any, targetPath []string, kind csvFieldKind, ids []string, fieldName string) error {
	switch kind {
	case csvFieldKindString:
		if len(ids) > 1 {
			return common.NewValidationErrorWithFields(fmt.Sprintf("multiple values provided for %s", fieldName), fieldName)
		}
		return setCSVMapStringField(data, targetPath, ids[0], fieldName)
	case csvFieldKindStringSlice:
		existing, err := getCSVMapStringSlice(data, targetPath)
		if err != nil {
			return err
		}
		merged := lo.Uniq(lo.Compact(append(existing, ids...)))
		setCSVMapValue(data, targetPath, merged)
		return nil
	default:
		return errCSVInvalidStringSliceField
	}
}

// csvRowStatesToStructs converts all row states back to their target structs.
func csvRowStatesToStructs(rowStates []*csvRowState) error {
	for _, row := range rowStates {
		if row == nil {
			continue
		}

		if err := csvMapToStruct(row.data, row.target); err != nil {
			return err
		}
	}

	return nil
}

// csvRuleMetaForType maps rule field names to JSON keys and target kind for a given element type.
func csvRuleMetaForType(elemType reflect.Type, rule csvReferenceRule) (csvRuleMeta, bool) {
	if rule.SourceField == "" || rule.TargetField == "" || rule.Lookup == nil {
		return csvRuleMeta{}, false
	}

	if elemType.Kind() != reflect.Struct {
		return csvRuleMeta{}, false
	}

	sourceMeta, ok := csvFieldMetaForType(elemType, rule.SourceField)
	if !ok {
		return csvRuleMeta{}, false
	}

	if meta, ok := csvFieldMetaForType(elemType, rule.TargetField); ok {
		if !csvReferenceTargetAllowed(elemType, rule.TargetField) {
			return csvRuleMeta{}, false
		}
		return csvRuleMeta{
			SourceJSONKey: sourceMeta.jsonKey,
			TargetJSONKey: meta.jsonKey,
			TargetInInput: false,
			TargetKind:    meta.kind,
		}, true
	}

	if inputField, ok := elemType.FieldByName(csvInputFieldName); ok {
		inputType := inputField.Type
		for inputType.Kind() == reflect.Pointer {
			inputType = inputType.Elem()
		}
		if meta, ok := csvFieldMetaForType(inputType, rule.TargetField); ok {
			if !csvReferenceTargetAllowed(elemType, rule.TargetField) {
				return csvRuleMeta{}, false
			}
			return csvRuleMeta{
				SourceJSONKey: sourceMeta.jsonKey,
				TargetJSONKey: meta.jsonKey,
				TargetInInput: true,
				TargetKind:    meta.kind,
			}, true
		}
	}

	return csvRuleMeta{}, false
}

// targetPath returns the JSON map path for the rule target.
func (meta csvRuleMeta) targetPath() []string {
	if meta.TargetInInput {
		return []string{csvInputFieldName, meta.TargetJSONKey}
	}

	return []string{meta.TargetJSONKey}
}

type csvFieldMeta struct {
	jsonKey string
	kind    csvFieldKind
}

// csvFieldMetaForType locates a field by name and returns its JSON key and kind.
func csvFieldMetaForType(t reflect.Type, fieldName string) (csvFieldMeta, bool) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return csvFieldMeta{}, false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		if field.Anonymous {
			if meta, ok := csvFieldMetaForType(field.Type, fieldName); ok {
				return meta, true
			}
			continue
		}

		if field.Name != fieldName {
			continue
		}

		jsonKey := csvJSONFieldName(field)
		if jsonKey == "" {
			return csvFieldMeta{}, false
		}

		return csvFieldMeta{
			jsonKey: jsonKey,
			kind:    csvFieldKindFromType(field.Type),
		}, true
	}

	return csvFieldMeta{}, false
}

// csvJSONFieldName returns the JSON field name for a struct field.
func csvJSONFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}

	name := strings.Split(tag, ",")[0]
	if name == "" {
		return field.Name
	}

	return name
}

// csvFieldKindFromType returns the supported csv field kind for a type.
func csvFieldKindFromType(t reflect.Type) csvFieldKind {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.String:
		return csvFieldKindString
	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			return csvFieldKindStringSlice
		}
	}

	return csvFieldKindUnknown
}

type csvRowState struct {
	target any
	data   map[string]any
}

// csvRowStatesFromSlice builds JSON maps for each row in the slice.
func csvRowStatesFromSlice(slice reflect.Value) ([]*csvRowState, error) {
	if slice.Kind() != reflect.Slice {
		return nil, common.NewValidationError("csv inputs must be a slice")
	}

	states := make([]*csvRowState, 0, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		state, err := csvRowStateFromValue(slice.Index(i))
		if err != nil {
			return nil, err
		}

		if state != nil {
			states = append(states, state)
		}
	}

	return states, nil
}

// csvRowStateFromValue marshals a row into a JSON map and retains its target pointer.
func csvRowStateFromValue(value reflect.Value) (*csvRowState, error) {
	if !value.IsValid() {
		return nil, nil
	}

	target := value
	for target.Kind() == reflect.Pointer {
		if target.IsNil() {
			return nil, nil
		}
		target = target.Elem()
	}
	if target.Kind() != reflect.Struct {
		return nil, nil
	}

	var ptr any
	if value.Kind() == reflect.Pointer {
		ptr = value.Interface()
	} else if value.CanAddr() {
		ptr = value.Addr().Interface()
	} else {
		return nil, nil
	}

	data, err := csvStructToMap(ptr)
	if err != nil {
		return nil, err
	}

	return &csvRowState{
		target: ptr,
		data:   data,
	}, nil
}

// csvStructToMap marshals a struct to a JSON map.
func csvStructToMap(input any) (map[string]any, error) {
	payload, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// csvMapToStruct unmarshals a JSON map back into a struct pointer.
func csvMapToStruct(data map[string]any, target any) error {
	if target == nil {
		return nil
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(payload, target)
}

// csvMapFieldStrings extracts string values for a single top-level JSON key.
func csvMapFieldStrings(data map[string]any, key string) []string {
	value, ok := data[key]
	if !ok {
		return nil
	}

	return csvAnyToStrings(value)
}

// csvAnyToStrings converts JSON values into a trimmed list of strings.
func csvAnyToStrings(value any) []string {
	switch v := value.(type) {
	case string:
		item := strings.TrimSpace(v)
		if item == "" {
			return nil
		}
		return []string{item}
	case []string:
		return lo.FilterMap(v, func(item string, _ int) (string, bool) {
			item = strings.TrimSpace(item)
			return item, item != ""
		})
	case []any:
		return lo.FilterMap(v, func(item any, _ int) (string, bool) {
			s, ok := item.(string)
			if !ok {
				return "", false
			}
			s = strings.TrimSpace(s)
			return s, s != ""
		})
	default:
		return nil
	}
}

// getCSVMapValue retrieves a nested JSON value by path.
func getCSVMapValue(data map[string]any, path []string) (any, bool) {
	current := any(data)
	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}

		value, ok := m[key]
		if !ok {
			return nil, false
		}

		current = value
	}

	return current, true
}

// setCSVMapValue writes a nested JSON value by path, creating maps as needed.
func setCSVMapValue(data map[string]any, path []string, value any) {
	if len(path) == 1 {
		data[path[0]] = value
		return
	}

	current := data
	for _, key := range path[:len(path)-1] {
		next, ok := current[key].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[key] = next
		}

		current = next
	}

	current[path[len(path)-1]] = value
}

// getCSVMapStringSlice reads a string slice from a nested JSON path.
func getCSVMapStringSlice(data map[string]any, path []string) ([]string, error) {
	value, ok := getCSVMapValue(data, path)
	if !ok || value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []string:
		return v, nil
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, errCSVFieldNotStringSlice
			}
			out = append(out, s)
		}

		return out, nil
	default:
		return nil, errCSVFieldNotStringSlice
	}
}

// setCSVMapStringField writes a single string value, enforcing conflict checks.
func setCSVMapStringField(data map[string]any, path []string, value string, fieldName string) error {
	existing, ok := getCSVMapValue(data, path)
	if ok && existing != nil {
		str, ok := existing.(string)
		if !ok {
			return errCSVFieldNotString
		}

		if str != "" {
			if str == value {
				return nil
			}

			return common.NewValidationErrorWithFields(fmt.Sprintf("conflicting values for %s", fieldName), fieldName)
		}
	}

	setCSVMapValue(data, path, value)
	return nil
}

// csvReferenceTargetExclusions blocks mapping to specified targets for input types.
var csvReferenceTargetExclusions = map[string]map[string]struct{}{
	"CreateDirectoryAccountInput": {
		"GroupIDs": {},
	},
}

// csvReferenceTargetAllowed returns false for excluded targets on specific input types.
func csvReferenceTargetAllowed(elemType reflect.Type, target string) bool {
	typeName := csvReferenceInputTypeName(elemType)
	if typeName == "" {
		return true
	}

	excluded, ok := csvReferenceTargetExclusions[typeName]
	if !ok {
		return true
	}

	_, blocked := excluded[target]

	return !blocked
}

// csvReferenceInputTypeName determines the canonical input type name for exclusion matching.
func csvReferenceInputTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}

	if inputField, ok := t.FieldByName(csvInputFieldName); ok {
		ft := inputField.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			name := ft.Name()
			if strings.HasPrefix(name, "Create") && strings.HasSuffix(name, "Input") {
				return name
			}
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.Anonymous {
			continue
		}

		ft := field.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}

		name := ft.Name()
		if strings.HasPrefix(name, "Create") && strings.HasSuffix(name, "Input") {
			return name
		}
	}

	return t.Name()
}

// lowerCamelField converts a field name to lower camel case for validation errors.
func lowerCamelField(value string) string {
	if value == "" {
		return value
	}

	return strings.ToLower(value[:1]) + value[1:]
}

// missingCSVValues lists unresolved original values based on normalized keys.
func missingCSVValues(unique map[string]string, resolved map[string]string) []string {
	return lo.FilterMap(lo.Entries(unique), func(entry lo.Entry[string, string], _ int) (string, bool) {
		_, ok := resolved[entry.Key]
		return entry.Value, !ok
	})
}
