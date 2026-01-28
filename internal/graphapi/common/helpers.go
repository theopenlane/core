package common //nolint:revive

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/objects/store"
	"github.com/theopenlane/core/internal/objects/upload"
	"github.com/theopenlane/core/pkg/logx"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

// CSV and GraphQL defaults used across helper functions.
const (
	// DefaultMaxMemoryMB is the default max memory for multipart forms (32MB)
	DefaultMaxMemoryMB = 32
	// GraphPath is the api graph path for the graphql server
	GraphPath = "query"
	// Default Complexity limit for graphql queries set by default, this can be overridden in the config
	DefaultComplexityLimit = 100
	// IntrospectionComplexity is the complexity limit for introspection queries
	IntrospectionComplexity = 200
	// csvInputWrapperFieldName is the wrapper field used for bulk CSV input structs.
	csvInputWrapperFieldName = "Input"
)

// injectFileUploader adds the file uploader as middleware to the graphql operation
// this is used to handle file uploads to a storage backend, add the file to the file schema
// and add the uploaded files to the echo context
func injectFileUploader(u *objects.Service) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (any, error) {
		rctx := graphql.GetFieldContext(ctx)

		// if the field context is nil or its not a resolver, return the next handler
		if rctx == nil || !rctx.IsResolver {
			return next(ctx)
		}

		// if the field context is a resolver, handle the file uploads
		op := graphql.GetOperationContext(ctx)

		// only handle mutations because the file uploads are only in mutations
		if op.Operation.Operation != "mutation" {
			return next(ctx)
		}

		// Use consolidated parsing logic for GraphQL variables
		inputKey := graphutils.GetInputFieldVariableName(ctx)

		// Parse files from GraphQL variables using the consolidated parser
		filesMap, err := pkgobjects.ParseFilesFromSource(op.Variables)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to parse files from graphql variables")

			return nil, err
		}

		// Convert to flat list, filtering out input key and adding object details
		uploads := []pkgobjects.File{}
		for k, files := range filesMap {
			// skip the input key
			if k == inputKey {
				continue
			}

			for _, fileUpload := range files {
				// Buffer the file in memory if small enough, otherwise leave as-is
				if fileUpload.RawFile != nil {
					buffered, err := pkgobjects.NewBufferedReaderFromReader(fileUpload.RawFile)
					if err == nil {
						fileUpload.RawFile = buffered
					}
				}

				// Add object details using existing logic
				enhanced, err := retrieveObjectDetails(rctx, op.Variables, inputKey, k, &fileUpload)
				if err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to retrieve object details for upload")

					return nil, err
				}

				uploads = append(uploads, *enhanced)
			}
		}

		// return the next handler if there are no uploads
		if len(uploads) == 0 {
			return next(ctx)
		}

		if err := setOrganizationForUploads(ctx, op.Variables, inputKey); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context for uploads")

			return nil, err
		}

		// Clean up any temporary files created by multipart form parser
		ec, err := echocontext.EchoContextFromContext(ctx)
		if err == nil && ec.Request().MultipartForm != nil {
			multipartForm := ec.Request().MultipartForm

			defer func() {
				if removeErr := multipartForm.RemoveAll(); removeErr != nil {
					logx.FromContext(ctx).Warn().Err(removeErr).Msg("failed to clean multipart form")
				}
			}()
		}

		ctx, _, err = upload.HandleUploads(ctx, u, uploads)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to handle file uploads")

			return nil, err
		}

		// add the uploaded files to the echo context if there are any
		// this is useful for using other middleware that depends on the echo context
		// and the uploaded files (e.g. body dump middleware)
		if ec != nil {
			ec.SetRequest(ec.Request().WithContext(ctx))
		}

		// process the rest of the resolver
		field, err := next(ctx)
		if err != nil {
			// rollback the uploaded files in case of an error
			upload.HandleRollback(ctx, u, uploads)
			logx.FromContext(ctx).Error().Err(err).Msg("failed to process resolver after file upload, rolling back uploads")

			return nil, err
		}

		// add the file permissions before returning the field
		if ctx, err = store.AddFilePermissions(ctx); err != nil {
			// rollback the uploaded files in case of an error
			upload.HandleRollback(ctx, u, uploads)

			logx.FromContext(ctx).Error().Err(err).Msg("failed to add file permissions to uploaded files")

			return nil, err
		}

		return field, nil
	}
}

// UnmarshalBulkData unmarshals the input bulk data into a slice of the given type
func UnmarshalBulkData[T any](input graphql.Upload) ([]*T, error) {
	// read the csv file
	var data []*T

	stream, readErr := io.ReadAll(input.File)
	if readErr != nil {
		return nil, readErr
	}
	// Configure the CSV reader gocsv will use
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.LazyQuotes = true    // tolerate odd quotes
		r.FieldsPerRecord = -1 // allow variable field counts
		r.TrimLeadingSpace = true
		return r
	})

	processed := preprocessCSVListCells[T](stream)
	if err := gocsv.UnmarshalBytes(processed, &data); err != nil {
		return nil, wrapCSVUnmarshalError(err, processed)
	}
	normalizeCSVEnumInputs(data)

	return data, nil
}

// wrapCSVUnmarshalError adds row, column, and header context for CSV parse errors
func wrapCSVUnmarshalError(err error, data []byte) error {
	var parseErr *csv.ParseError
	if !errors.As(err, &parseErr) {
		return err
	}

	headerName := stripCSVInputPrefix(csvHeaderForColumn(data, parseErr.Column))
	message := fmt.Sprintf("csv parse error on line %d, column %d", parseErr.Line, parseErr.Column)
	if headerName != "" {
		message = fmt.Sprintf("%s (header %q)", message, headerName)
	}
	if parseErr.Err != nil {
		message = fmt.Sprintf("%s: %s", message, parseErr.Err.Error())
	}

	if hint := jsonColumnHint(parseErr.Err); hint != "" {
		message = fmt.Sprintf("%s %s", message, hint)
	}

	if headerName != "" {
		return NewValidationErrorWithFields(message, headerName)
	}

	return NewValidationError(message)
}

// csvHeaderForColumn returns the header name for a 1-based column index
func csvHeaderForColumn(data []byte, column int) string {
	if column <= 0 {
		return ""
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return ""
	}

	index := column - 1
	if index < 0 || index >= len(headers) {
		return ""
	}

	return strings.TrimSpace(headers[index])
}

// jsonColumnHint adds guidance when a CSV value fails JSON parsing
func jsonColumnHint(err error) string {
	if err == nil {
		return ""
	}

	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &syntaxErr) && !errors.As(err, &typeErr) {
		return ""
	}

	return "list or object values must be valid JSON and the cell should be quoted, e.g. \"[\\\"security\\\",\\\"compliance\\\"]\""
}

// enumInfo caches enum metadata for CSV normalization.
type enumInfo struct {
	defaultValue string
	normalized   map[string]string
	ok           bool
}

// dateTimeType caches the DateTime reflect type for pointer checks
var dateTimeType = reflect.TypeFor[models.DateTime]()

// normalizeCSVEnumInputs walks decoded CSV rows and normalizes enum values
func normalizeCSVEnumInputs(data any) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		// Only CSV row slices are expected here; skip anything else to avoid panics
		return
	}

	cache := map[reflect.Type]enumInfo{}
	for i := 0; i < v.Len(); i++ {
		// Normalize each row independently while sharing enum metadata cache
		normalizeCSVEnumValue(v.Index(i), cache)
	}
}

// csvInputWrapper identifies wrapper types needing CSV header prefixing.
type csvInputWrapper interface {
	CSVInputWrapper()
}

// preprocessCSVListCells converts list-like CSV cell values into JSON arrays for slice fields.
// Valid JSON arrays are preserved, and scalar JSON values are wrapped into arrays.
func preprocessCSVListCells[T any](data []byte) []byte {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		return data
	}

	headers := records[0]
	changed := prefixCSVInputHeaders[T](headers)

	fieldKinds := csvListFieldKinds[T]()
	normalizedHeaders := make([]string, len(headers))
	for i, header := range headers {
		normalizedHeaders[i] = normalizeCSVHeader(header)
	}

	if len(fieldKinds) > 0 {
		for rowIdx := 1; rowIdx < len(records); rowIdx++ {
			row := records[rowIdx]
			for colIdx, headerKey := range normalizedHeaders {
				if colIdx >= len(row) {
					continue
				}
				kind, ok := fieldKinds[headerKey]
				if !ok {
					continue
				}

				cell := strings.TrimSpace(row[colIdx])
				if cell == "" {
					continue
				}

				if kind == reflect.Map {
					if json.Valid([]byte(cell)) {
						continue
					}
					continue
				}

				if kind == reflect.Slice {
					normalized, updated := normalizeCSVListCell(cell)
					if updated {
						row[colIdx] = normalized
						changed = true
					}
				}
			}
			records[rowIdx] = row
		}
	}

	if !changed {
		return data
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.WriteAll(records); err != nil {
		return data
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return data
	}

	return buf.Bytes()
}

// csvListFieldKinds collects CSV list-capable field kinds for a generic input type.
func csvListFieldKinds[T any]() map[string]reflect.Kind {
	var value T
	t := reflect.TypeOf(value)
	if t == nil {
		t = reflect.TypeOf((*T)(nil)).Elem()
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	fieldKinds := map[string]reflect.Kind{}
	collectCSVListFieldKinds(t, fieldKinds, []string{""})
	if len(fieldKinds) == 0 {
		return nil
	}

	return fieldKinds
}

// collectCSVListFieldKinds walks struct fields to capture slice/map CSV columns.
func collectCSVListFieldKinds(t reflect.Type, fieldKinds map[string]reflect.Kind, prefixes []string) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	if len(prefixes) == 0 {
		prefixes = []string{""}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		fieldType := field.Type
		if field.Anonymous {
			collectCSVListFieldKinds(fieldType, fieldKinds, prefixes)
			continue
		}

		names := csvFieldNames(field)
		if len(names) == 0 {
			continue
		}

		combined := combineCSVPrefixes(prefixes, names)

		kind := fieldType.Kind()
		if kind == reflect.Pointer {
			kind = fieldType.Elem().Kind()
		}

		if kind == reflect.Struct {
			collectCSVListFieldKinds(fieldType, fieldKinds, combined)
			continue
		}

		if kind != reflect.Slice && kind != reflect.Map {
			continue
		}

		for _, name := range combined {
			fieldKinds[normalizeCSVHeader(name)] = kind
		}
	}
}

// prefixCSVInputHeaders adds Input:: prefixes when a csvInput wrapper is used.
func prefixCSVInputHeaders[T any](headers []string) bool {
	if !isCSVInputWrapper[T]() {
		return false
	}

	var value T
	t := reflect.TypeOf(value)
	if t == nil {
		t = reflect.TypeOf((*T)(nil)).Elem()
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	if _, ok := t.FieldByName(csvInputWrapperFieldName); !ok {
		return false
	}

	embeddedHeaders := csvEmbeddedHeaderSet(t)
	changed := false
	for i, header := range headers {
		trimmed := strings.TrimSpace(header)
		if trimmed == "" {
			continue
		}
		if hasCSVInputPrefix(trimmed) {
			continue
		}
		if _, ok := embeddedHeaders[normalizeCSVHeader(trimmed)]; ok {
			continue
		}
		headers[i] = csvInputWrapperFieldName + gocsv.FieldsCombiner + trimmed
		changed = true
	}

	return changed
}

// isCSVInputWrapper reports whether the type implements csvInputWrapper.
func isCSVInputWrapper[T any]() bool {
	var value T
	if _, ok := any(value).(csvInputWrapper); ok {
		return true
	}
	if _, ok := any(&value).(csvInputWrapper); ok {
		return true
	}
	return false
}

// csvEmbeddedHeaderSet returns header names for embedded fields to avoid double prefixing.
func csvEmbeddedHeaderSet(t reflect.Type) map[string]struct{} {
	set := map[string]struct{}{}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return set
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.Anonymous {
			continue
		}
		collectCSVHeaderNames(field.Type, set, []string{""})
	}

	return set
}

// collectCSVHeaderNames walks fields to gather CSV header names for embedded structs.
func collectCSVHeaderNames(t reflect.Type, headers map[string]struct{}, prefixes []string) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}
	if len(prefixes) == 0 {
		prefixes = []string{""}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		fieldType := field.Type
		if field.Anonymous {
			collectCSVHeaderNames(fieldType, headers, prefixes)
			continue
		}

		names := csvFieldNames(field)
		if len(names) == 0 {
			continue
		}

		combined := combineCSVPrefixes(prefixes, names)

		kind := fieldType.Kind()
		if kind == reflect.Pointer {
			kind = fieldType.Elem().Kind()
		}
		if kind == reflect.Struct {
			collectCSVHeaderNames(fieldType, headers, combined)
			continue
		}

		for _, name := range combined {
			headers[normalizeCSVHeader(name)] = struct{}{}
		}
	}
}

// csvFieldNames extracts CSV header names from struct tags or field names.
func csvFieldNames(field reflect.StructField) []string {
	tag := strings.TrimSpace(field.Tag.Get("csv"))
	if tag != "" {
		parts := strings.Split(tag, ",")
		names := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if part == "omitempty" || part == "partial" || strings.HasPrefix(part, "default=") {
				continue
			}
			names = append(names, part)
		}
		if len(names) == 1 && names[0] == "-" {
			return nil
		}
		if len(names) > 0 {
			return names
		}
	}

	return []string{field.Name}
}

// combineCSVPrefixes prefixes field names for nested CSV headers.
func combineCSVPrefixes(prefixes []string, names []string) []string {
	if len(prefixes) == 0 {
		prefixes = []string{""}
	}
	out := make([]string, 0, len(prefixes)*len(names))
	for _, prefix := range prefixes {
		for _, name := range names {
			if prefix == "" {
				out = append(out, name)
			} else {
				out = append(out, prefix+gocsv.FieldsCombiner+name)
			}
		}
	}
	return out
}

// hasCSVInputPrefix reports whether a header already has the Input:: prefix.
func hasCSVInputPrefix(value string) bool {
	parts := strings.SplitN(value, gocsv.FieldsCombiner, 2)
	if len(parts) < 2 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(parts[0]), csvInputWrapperFieldName)
}

// stripCSVInputPrefix removes the Input:: prefix when present.
func stripCSVInputPrefix(header string) string {
	parts := strings.SplitN(header, gocsv.FieldsCombiner, 2)
	if len(parts) < 2 {
		return header
	}
	if strings.EqualFold(strings.TrimSpace(parts[0]), csvInputWrapperFieldName) {
		return parts[1]
	}
	return header
}

// normalizeCSVHeader lowercases and strips separators for header matching.
func normalizeCSVHeader(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(" ", "", "_", "", "-", "", ".", "")
	return replacer.Replace(value)
}

// normalizeCSVListCell ensures list-like values are encoded as JSON arrays.
func normalizeCSVListCell(cell string) (string, bool) {
	if json.Valid([]byte(cell)) {
		var decoded any
		if err := json.Unmarshal([]byte(cell), &decoded); err == nil {
			switch decoded.(type) {
			case []any:
				return cell, false
			case map[string]any:
				return cell, false
			case nil:
				return cell, false
			default:
				items := []string{fmt.Sprint(decoded)}
				payload, err := json.Marshal(items)
				if err != nil {
					return cell, false
				}
				return string(payload), true
			}
		}
		return cell, false
	}

	items := splitCSVList(cell)
	if len(items) == 0 {
		return cell, false
	}

	payload, err := json.Marshal(items)
	if err != nil {
		return cell, false
	}

	return string(payload), true
}

// splitCSVList splits list-like CSV values on common delimiters.
func splitCSVList(value string) []string {
	var delimiter string
	switch {
	case strings.Contains(value, ";"):
		delimiter = ";"
	case strings.Contains(value, "|"):
		delimiter = "|"
	case strings.Contains(value, ","):
		delimiter = ","
	default:
		item := strings.TrimSpace(value)
		if item == "" {
			return nil
		}
		return []string{item}
	}

	parts := strings.Split(value, delimiter)
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		items = append(items, item)
	}

	if len(items) == 0 {
		return nil
	}

	return items
}

// normalizeCSVEnumValue recursively normalizes enum fields on a struct value
func normalizeCSVEnumValue(v reflect.Value, cache map[reflect.Type]enumInfo) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			// Nil rows are allowed (e.g., failed CSV parse); nothing to normalize
			return
		}
		// Work with the concrete struct value for field traversal
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		// Only structs contain named fields to normalize
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			// Skip unexported or non-settable fields to avoid reflection panics
			continue
		}

		switch field.Kind() {
		case reflect.Pointer:
			// Pointer enums are treated differently to allow nil/default handling
			normalizeEnumPointerField(field, cache)
			// DateTime pointers from CSV can be zero when the cell is blank, so clear them
			normalizeDateTimePointerField(field)
		case reflect.String:
			// Value enums are normalized in place
			normalizeEnumStringField(field, cache)
		case reflect.Slice:
			// Normalize slices of enum values (e.g., []enums.Channel)
			normalizeEnumSliceField(field, cache)
		case reflect.Struct:
			// Recurse into nested structs to normalize embedded inputs
			normalizeCSVEnumValue(field, cache)
		}
	}
}

// normalizeEnumPointerField handles pointer enum fields, clearing empty inputs to allow defaults
func normalizeEnumPointerField(field reflect.Value, cache map[reflect.Type]enumInfo) {
	if field.IsNil() {
		// No value provided, so nothing to normalize
		return
	}

	info := enumInfoForType(field.Type().Elem(), cache)
	if !info.ok {
		// Not an enum type (no Values method), skip normalization
		return
	}

	// Enum values are string-backed; normalize the raw string representation
	raw := field.Elem().String()
	normalized, needsClear := normalizeEnumValue(raw, info)
	if needsClear {
		// Treat empty CSV cells as "unset" so ent defaults can apply
		field.Set(reflect.Zero(field.Type()))

		return
	}

	if normalized == raw {
		// Already canonical
		return
	}

	// Assign the normalized enum value back into the pointer
	next := reflect.New(field.Type().Elem()).Elem()
	next.SetString(normalized)
	field.Set(next.Addr())
}

// normalizeDateTimePointerField clears zero DateTime pointers so empty CSV cells stay unset
func normalizeDateTimePointerField(field reflect.Value) {
	if field.IsNil() {
		// No value provided, so nothing to normalize
		return
	}

	if field.Type().Elem() != dateTimeType {
		// Only DateTime pointers need this normalization
		return
	}

	dt := field.Elem().Interface().(models.DateTime)
	if dt.IsZero() {
		// gocsv allocates DateTime pointers even for empty cells, so clear them for defaults
		field.Set(reflect.Zero(field.Type()))
	}
}

// normalizeEnumStringField handles value enum fields and applies defaults for empty inputs
func normalizeEnumStringField(field reflect.Value, cache map[reflect.Type]enumInfo) {
	info := enumInfoForType(field.Type(), cache)
	if !info.ok {
		// Not an enum type (no Values method), skip normalization
		return
	}

	raw := field.String()
	normalized, needsClear := normalizeEnumValue(raw, info)
	if needsClear {
		// Value enums cannot be nil; use the enum's default if available
		if info.defaultValue != "" {
			field.SetString(info.defaultValue)
		} else {
			// Fall back to empty string if no default exists
			field.SetString("")
		}
		return
	}

	if normalized != raw {
		// Normalize user input (spacing/case) to the canonical enum string
		field.SetString(normalized)
	}
}

// normalizeEnumSliceField normalizes slices of enum values while preserving positions
func normalizeEnumSliceField(field reflect.Value, cache map[reflect.Type]enumInfo) {
	info := enumInfoForType(field.Type().Elem(), cache)
	if !info.ok {
		// Not a slice of enums; skip
		return
	}

	for i := 0; i < field.Len(); i++ {
		elem := field.Index(i)
		if !elem.CanSet() {
			// Skip non-settable elements to avoid reflection panics
			continue
		}
		raw := elem.String()
		normalized, needsClear := normalizeEnumValue(raw, info)
		if needsClear {
			// Preserve empty entries rather than removing them
			elem.SetString("")
			continue
		}
		if normalized != raw {
			// Normalize each element independently
			elem.SetString(normalized)
		}
	}
}

// enumInfoForType extracts enum metadata using the Values method and caches it by type
func enumInfoForType(t reflect.Type, cache map[reflect.Type]enumInfo) enumInfo {
	if info, ok := cache[t]; ok {
		return info
	}

	info := enumInfo{}
	if t.Kind() != reflect.String {
		// Enums are string-backed; skip non-string types
		cache[t] = info
		return info
	}

	method, ok := t.MethodByName("Values")
	var receiver reflect.Value
	if ok {
		// Value receiver provides Values()
		receiver = reflect.Zero(t)
	} else {
		// Pointer receiver provides Values()
		ptrType := reflect.PointerTo(t)
		method, ok = ptrType.MethodByName("Values")
		if !ok {
			// No Values method means this is not an enum type
			cache[t] = info
			return info
		}
		receiver = reflect.New(t)
	}

	if method.Type.NumIn() != 1 || method.Type.NumOut() != 1 {
		// Values must be func() []string; otherwise skip
		cache[t] = info
		return info
	}
	out := method.Type.Out(0)
	if out.Kind() != reflect.Slice || out.Elem().Kind() != reflect.String {
		// Values must return []string to be considered an enum
		cache[t] = info
		return info
	}

	results := method.Func.Call([]reflect.Value{receiver})
	values, ok := results[0].Interface().([]string)
	if !ok {
		// Defensive: unexpected Values return type
		cache[t] = info
		return info
	}

	// Build a lookup of normalized token -> canonical enum value
	info.normalized = lo.Associate(values, func(value string) (string, string) {
		return normalizeEnumToken(value), value
	})
	delete(info.normalized, "")
	if len(values) > 0 {
		// Default to the first enum entry when value enums are left blank
		info.defaultValue = values[0]
	}
	info.ok = true
	cache[t] = info

	return info
}

// normalizeEnumValue maps raw CSV input to a canonical enum value and signals when it is empty
func normalizeEnumValue(raw string, info enumInfo) (string, bool) {
	// Normalize once so blanks and human-friendly input have a consistent shape
	normalized := normalizeEnumToken(raw)
	if normalized == "" {
		// Empty cells should not fail validation; they either clear pointers or use defaults
		return "", true
	}

	if len(info.normalized) > 0 {
		// Prefer exact known values to avoid over-normalizing unusual enum strings
		if canonical, ok := info.normalized[normalized]; ok {
			return canonical, false
		}
	}

	// Fall back to the normalized token for user-friendly inputs like "In Review"
	return normalized, false
}

// normalizeEnumToken trims, snake-cases, and uppercases a value to align user input with enum constants
func normalizeEnumToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		// Skip empty values to avoid bogus keys
		return ""
	}
	// SnakeCase unifies "In Review" and "IN_REVIEW", then uppercase matches enum constants
	return strings.ToUpper(lo.SnakeCase(value))
}

// inputWithOwnerID is a struct that contains the owner id
// this is used to unmarshal the owner id from the input
type inputWithOwnerID struct {
	OwnerID *string `json:"ownerID"`
}

// inputWithTemplateKind is a struct that contains the template kind
// this is used to unmarshal the template kind from the input
type inputWithTemplateKind struct {
	Kind *enums.TemplateKind `json:"kind"`
}

// GetOrgOwnerFromInput retrieves the owner id from the input
// input can be of any type, but must contain an owner id field
// if the owner id is not found, it returns nil
func GetOrgOwnerFromInput[T any](input *T) (*string, error) {
	if input == nil {
		return nil, nil
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var ownerInput inputWithOwnerID
	if err := json.Unmarshal(inputBytes, &ownerInput); err != nil {
		return nil, err
	}

	if ownerInput.OwnerID != nil {
		return ownerInput.OwnerID, nil
	}

	var wrappedOwner struct {
		Input inputWithOwnerID `json:"Input"`
	}
	if err := json.Unmarshal(inputBytes, &wrappedOwner); err == nil && wrappedOwner.Input.OwnerID != nil {
		return wrappedOwner.Input.OwnerID, nil
	}

	var wrappedOwnerLower struct {
		Input inputWithOwnerID `json:"input"`
	}
	if err := json.Unmarshal(inputBytes, &wrappedOwnerLower); err == nil && wrappedOwnerLower.Input.OwnerID != nil {
		return wrappedOwnerLower.Input.OwnerID, nil
	}

	return nil, nil
}

// templateKindFromVariables extracts the TemplateKind from the input variables
func templateKindFromVariables(variables map[string]any, inputKey string) *enums.TemplateKind {
	if inputKey == "" {
		return nil
	}

	rawInput, ok := variables[inputKey]
	if !ok || rawInput == nil {
		return nil
	}

	inputBytes, err := json.Marshal(rawInput)
	if err != nil {
		return nil
	}

	var input inputWithTemplateKind
	if err := json.Unmarshal(inputBytes, &input); err != nil {
		return nil
	}

	if input.Kind == nil || *input.Kind == enums.TemplateKindInvalid {
		return nil
	}

	return input.Kind
}

// GetBulkUploadOwnerInput retrieves the owner id from the bulk upload input
// if there are multiple owner ids, it returns an error
// this is used to ensure that the owner id is consistent across all inputs
func GetBulkUploadOwnerInput[T any](input []*T) (*string, error) {
	var ownerID *string

	for _, i := range input {
		ownerInputID, err := GetOrgOwnerFromInput(i)
		if err != nil {
			return nil, err
		}

		if ownerInputID == nil {
			log.Error().Msg("owner id not found in bulk upload input")

			return nil, gqlerrors.NewCustomError(
				gqlerrors.BadRequestErrorCode,
				"unable to determine the organization owner id from the input, no owner id found",
				ErrNoOrganizationID,
			)
		}

		// if the owner doesn't match the previous owner, return an error
		if ownerID != nil && *ownerInputID != *ownerID {
			log.Error().Msg("multiple owner ids found in bulk upload input")

			return nil, gqlerrors.NewCustomError(
				gqlerrors.BadRequestErrorCode,
				"unable to determine the organization owner id from the input, multiple owner ids found",
				ErrNoOrganizationID,
			)
		}

		ownerID = ownerInputID
	}

	return ownerID, nil
}

// SetOrganizationInAuthContext sets the organization in the auth context based on the input if it is not already set
// in most cases this is a no-op because the organization id is set in the auth middleware
// only when multiple organizations are authorized (e.g. with a PAT) is this necessary
func SetOrganizationInAuthContext(ctx context.Context, inputOrgID *string) error {
	// if org is in context or the user is a system admin, return
	if ok, err := checkOrgInContext(ctx); ok && err == nil {
		return nil
	}

	// If no input provided, fallback to a single authorized org (e.g., API token with one org)
	if inputOrgID == nil {
		if au, err := auth.GetAuthenticatedUserFromContext(ctx); err == nil {
			if len(au.OrganizationIDs) == 1 && au.OrganizationIDs[0] != "" {
				return auth.SetOrganizationIDInAuthContext(ctx, au.OrganizationIDs[0])
			}
		}
	}

	return setOrgFromInputInContext(ctx, inputOrgID)
}

// SetOrganizationInAuthContextBulkRequest sets the organization in the auth context based on the input if it is not already set
// in most cases this is a no-op because the organization id is set in the auth middleware
// in the case of personal access tokens, this is necessary to ensure the organization id is set
// the organization must be the same across all inputs in the bulk request
func SetOrganizationInAuthContextBulkRequest[T any](ctx context.Context, input []*T) error {
	// if org is in context or the user is a system admin, return
	if ok, err := checkOrgInContext(ctx); ok && err == nil {
		return nil
	}

	ownerID, err := GetBulkUploadOwnerInput(input)
	if err != nil {
		return err
	}

	return setOrgFromInputInContext(ctx, ownerID)
}

// checkOrgInContext checks if the organization is already set in the context
// if the organization is set, it returns true
// if the user is a system admin, it also returns true
func checkOrgInContext(ctx context.Context) (bool, error) {
	// allow system admins to bypass the organization check
	isAdmin, err := rule.CheckIsSystemAdminWithContext(ctx)
	if err == nil && isAdmin {
		log.Debug().Bool("isAdmin", isAdmin).Msg("user is system admin, bypassing setting organization in auth context")

		return true, nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err == nil && orgID != "" {
		return true, nil
	}

	return false, nil
}

// setOrgFromInputInContext sets the organization in the auth context based on the input org ID, ensuring
// the org is authenticated and exists in the context
func setOrgFromInputInContext(ctx context.Context, inputOrgID *string) error {
	if inputOrgID == nil {
		// this would happen on a PAT authenticated request because the org id is not set
		return ErrNoOrganizationID
	}

	// ensure this org is authenticated
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return err
	}

	if !lo.Contains(orgIDs, *inputOrgID) {
		return fmt.Errorf("%w: organization id %s not found in the authenticated organizations", rout.ErrBadRequest, *inputOrgID)
	}

	err = auth.SetOrganizationIDInAuthContext(ctx, *inputOrgID)
	if err != nil {
		return err
	}

	return nil
}

// CheckAllowedAuthType checks how the user is authenticated and returns an error
// if the user is authenticated with an API token for a user owned setting
func CheckAllowedAuthType(ctx context.Context) error {
	ac, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return err
	}

	if ac.AuthenticationType == auth.APITokenAuthentication {
		return fmt.Errorf("%w: unable to use API token to update user settings", rout.ErrBadRequest)
	}

	return nil
}

// retrieveObjectDetails retrieves the object details from the field context
func retrieveObjectDetails(rctx *graphql.FieldContext, variables map[string]any, inputKey, key string, upload *pkgobjects.File) (*pkgobjects.File, error) {
	// loop through the arguments in the request
	for _, arg := range rctx.Field.Arguments {
		// check if the argument is an upload
		if argIsUpload(arg) {
			// check if the argument name matches the key
			if arg.Name == key {
				objectID, ok := rctx.Args["id"].(string)
				if ok {
					upload.CorrelatedObjectID = objectID
					upload.Parent.ID = objectID
				}

				objectType := stripOperation(rctx.Field.Name)
				upload.CorrelatedObjectType = objectType
				// Parent.Type must use snake_case to align with authz tuple expectations (e.g., trust_center_doc)
				upload.Parent.Type = lo.SnakeCase(objectType)
				upload.FieldName = arg.Name
				upload.Key = arg.Name // Also set Key in FileMetadata for backwards compatibility

				if key == "templateFiles" {
					if kind := templateKindFromVariables(variables, inputKey); kind != nil {
						objects.SetTemplateKindHint(upload, *kind)
					} else if rctx.Field.Name == "createTrustCenterNDA" || rctx.Field.Name == "updateTrustCenterNDA" {
						objects.SetTemplateKindHint(upload, enums.TemplateKindTrustCenterNda)
					}
				}

				return upload, nil
			}
		}
	}

	log.Debug().Str("key", key).Msg("unable to determine object type - no matching upload argument found")

	return upload, ErrUnableToDetermineObjectType
}

// argIsUpload checks if the argument is an upload
func argIsUpload(arg *ast.Argument) bool {
	if arg == nil || arg.Value == nil || arg.Value.ExpectedType == nil {
		return false
	}

	if arg.Value.ExpectedType.NamedType == "Upload" {
		return true
	}

	if arg.Value.ExpectedType.Elem != nil && arg.Value.ExpectedType.Elem.NamedType == "Upload" {
		return true
	}

	return false
}

// stripOperation strips the operation prefix from the field name and returns the remainder unchanged
// e.g. updateUser becomes User, createTrustCenterDoc becomes TrustCenterDoc
func stripOperation(field string) string {
	operations := []string{"create", "update", "delete", "get"}

	result := field
	for _, op := range operations {
		if strings.HasPrefix(result, op) {
			result = strings.ReplaceAll(result, op, "")

			break
		}
	}

	return strings.TrimPrefix(result, "Upload")
}

// IsEmpty checks if the given interface is empty
func IsEmpty(x any) bool {
	if x == nil {
		return true
	}

	switch v := x.(type) {
	case string:
		return v == lo.Empty[string]()
	case []int, []string, []float64, []any:
		return isEmptySlice(v)
	case map[string]any:
		return len(v) == 0
	case int, int8, int16, int32, int64:
		return v == lo.Empty[int]()
	case uint, uint8, uint16, uint32, uint64:
		return v == lo.Empty[uint]()
	case float32, float64:
		return v == lo.Empty[float64]()
	case bool:
		return v == lo.Empty[bool]()
	default:
		// fallback to this helper, which expects a comparable type
		return lo.IsNil(v)
	}

}

// isEmptySlice checks if the given interface is an empty slice
func isEmptySlice(x any) bool {
	switch v := x.(type) {
	case []int:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case []float64:
		return len(v) == 0
	case []any:
		return len(v) == 0
	default:
		return false
	}
}

// ConvertToObject converts an object to a specific type
func ConvertToObject[J any](obj any) (*J, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result J
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// setOrganizationForUploads ensures an organization is present in the auth context
// we want this for token-authenticated requests where the active org is not pre-selected (e.g., PATs)
func setOrganizationForUploads(ctx context.Context, variables map[string]any, inputKey string) error {
	if orgID, err := auth.GetOrganizationIDFromContext(ctx); err == nil && orgID != "" {
		return nil
	}

	ownerID, err := getOwnerIDFromVariables(variables, inputKey)
	if err != nil {
		return err
	}

	return SetOrganizationInAuthContext(ctx, ownerID)
}

// getOwnerIDFromVariables attempts to extract an owner/organization ID from the GraphQL variables map
func getOwnerIDFromVariables(variables map[string]any, inputKey string) (*string, error) {
	// Prefer the primary input payload (e.g., "input") if available
	if inputKey != "" {
		if rawInput, ok := variables[inputKey]; ok && rawInput != nil {
			inputBytes, err := json.Marshal(rawInput)
			if err != nil {
				return nil, err
			}

			var owner inputWithOwnerID
			if err := json.Unmarshal(inputBytes, &owner); err != nil {
				return nil, err
			}

			if owner.OwnerID != nil && *owner.OwnerID != "" {
				return owner.OwnerID, nil
			}
		}
	}

	// Also handle cases where ownerID is passed as a top-level variable
	if rawOwner, ok := variables["ownerID"]; ok {
		if ownerStr, ok := rawOwner.(string); ok && ownerStr != "" {
			return &ownerStr, nil
		}
	}

	return nil, nil
}
