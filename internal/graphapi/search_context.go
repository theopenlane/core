package graphapi

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkparser "github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/pkg/mapx"
)

const (
	maxSnippetLength = 100
)

var snippetMarkdown = goldmark.New(
	goldmark.WithExtensions(extension.Table),
	goldmark.WithParserOptions(goldmarkparser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
)

var searchSanitizer = bluemonday.StrictPolicy()

type searchCtxTracker struct {
	mu       sync.Mutex
	contexts map[string]*models.SearchContext
	query    string
}

func newContextTracker(query string) *searchCtxTracker {
	return &searchCtxTracker{
		contexts: make(map[string]*models.SearchContext),
		query:    query,
	}
}

func (t *searchCtxTracker) addMatch(entityID, entityType string, fieldMatches []string, entity any) {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry := mapx.GetOrInit(t.contexts, entityID, func() *models.SearchContext {
		return &models.SearchContext{
			EntityID:      entityID,
			EntityType:    entityType,
			MatchedFields: fieldMatches,
		}
	})
	entry.Snippets = append(entry.Snippets, t.extractSnippets(entity, fieldMatches)...)
}

func (t *searchCtxTracker) extractSnippets(entity any, matchedFields []string) []*models.SearchSnippet {
	snippets := make([]*models.SearchSnippet, 0)

	if entity == nil {
		return snippets
	}

	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return snippets
		}

		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return snippets
	}

	queryLower := strings.ToLower(t.query)

	for _, fieldName := range matchedFields {

		field := val.FieldByNameFunc(func(n string) bool {
			return strings.EqualFold(n, fieldName)
		})

		if !field.IsValid() {
			continue
		}

		text := sanitizeContent(retrieveFieldValue(field))
		if text == "" {
			continue
		}

		snippet := t.createSnippet(fieldName, text, queryLower)
		if snippet == nil {
			continue
		}

		snippets = append(snippets, snippet)
	}

	return snippets
}

func retrieveFieldValue(field reflect.Value) string {
	switch field.Kind() {
	case reflect.String:
		return field.String()

	case reflect.Slice:
		// tags, mapped categories and aliases are []string
		if field.Type().Elem().Kind() != reflect.String {
			return ""
		}

		val := make([]string, 0, field.Len())
		for i := 0; i < field.Len(); i++ {
			val = append(val, field.Index(i).String())
		}

		return strings.Join(val, ", ")

	case reflect.Pointer:
		// handle nullable string values
		if !field.IsNil() && field.Elem().Kind() == reflect.String {
			return field.Elem().String()
		}
	}

	return ""
}

func sanitizeContent(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	var b bytes.Buffer
	if err := snippetMarkdown.Convert([]byte(text), &b); err != nil {
		return text
	}

	sanitized := searchSanitizer.Sanitize(b.String())

	return strings.Join(strings.Fields(sanitized), " ")
}

// createSnippet creates a snippet with highlighted match to improve the surrounding context
// needed
func (t *searchCtxTracker) createSnippet(fieldName, text, queryLower string) *models.SearchSnippet {
	textLower := strings.ToLower(text)
	idx := strings.Index(textLower, queryLower)

	if idx == -1 {
		if len(text) > maxSnippetLength {
			return &models.SearchSnippet{
				Field: fieldName,
				Text:  text[:maxSnippetLength] + "...", // if too long, add ...
			}
		}

		return &models.SearchSnippet{
			Field: fieldName,
			Text:  text,
		}
	}

	contextSize := 50
	start := max(0, idx-contextSize)

	end := min(len(text), idx+len(t.query)+contextSize)

	snippet := text[start:end]

	if start > 0 {
		snippet = "..." + snippet
	}

	if end < len(text) {
		snippet += "..."
	}

	return &models.SearchSnippet{
		Field: fieldName,
		Text:  snippet,
	}
}

func (t *searchCtxTracker) getContexts() []*models.SearchContext {
	t.mu.Lock()
	defer t.mu.Unlock()

	contexts := make([]*models.SearchContext, 0, len(t.contexts))
	for _, ctx := range t.contexts {
		contexts = append(contexts, ctx)
	}
	return contexts
}

type fieldMatchChecker struct {
	query string
}

// check checks which of the given fields match the query for the entity
func (c *fieldMatchChecker) check(entity any, fieldNames []string) []string {
	matches := make([]string, 0)

	if entity == nil {
		return matches
	}

	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return matches
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return matches
	}

	queryLower := strings.ToLower(c.query)

	for _, fieldName := range fieldNames {
		field := val.FieldByNameFunc(func(n string) bool {
			return strings.EqualFold(n, fieldName)
		})

		if !field.IsValid() {
			continue
		}

		matched := false
		text := sanitizeContent(retrieveFieldValue(field))
		if text != "" {
			matched = strings.Contains(strings.ToLower(text), queryLower)
		}

		if matched {
			matches = append(matches, fieldName)
		}
	}

	// If the ID matches exactly, add it
	idField := val.FieldByName("ID")
	if idField.IsValid() && idField.Kind() == reflect.String {
		if idField.String() == c.query {
			matches = append(matches, "ID")
		}
	}

	return matches
}

// highlightSearchContext processes search results using type switches for better type safety
// This is a non-magical alternative to processSearchResults that explicitly handles each type
func highlightSearchContext(_ context.Context, query string, results any, tracker *searchCtxTracker) {
	if results == nil || tracker == nil {
		return
	}

	checker := fieldMatchChecker{query: query}

	switch conn := results.(type) {
	case *generated.ActionPlanConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Details", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "ActionPlan", matchedFields, node)
			}
		}
	case *generated.AssetConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Asset", matchedFields, node)
			}
		}
	case *generated.ContactConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Email", "FullName", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Contact", matchedFields, node)
			}
		}
	case *generated.ControlConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Aliases", "Category", "Description", "DisplayID", "MappedCategories", "RefCode", "Subcategory", "Tags", "Title"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Control", matchedFields, node)
			}
		}
	case *generated.EntityConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Description", "DisplayName", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Entity", matchedFields, node)
			}
		}
	case *generated.EvidenceConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"DisplayID", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Evidence", matchedFields, node)
			}
		}
	case *generated.GroupConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"DisplayID", "DisplayName", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Group", matchedFields, node)
			}
		}
	case *generated.InternalPolicyConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Details", "DisplayID", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "InternalPolicy", matchedFields, node)
			}
		}
	case *generated.InviteConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Recipient"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Invite", matchedFields, node)
			}
		}
	case *generated.NarrativeConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Description", "DisplayID", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Narrative", matchedFields, node)
			}
		}
	case *generated.OrganizationConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"DisplayName", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Organization", matchedFields, node)
			}
		}
	case *generated.ProcedureConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Details", "DisplayID", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Procedure", matchedFields, node)
			}
		}
	case *generated.ProgramConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Description", "DisplayID", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Program", matchedFields, node)
			}
		}
	case *generated.RiskConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"DisplayID", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Risk", matchedFields, node)
			}
		}
	case *generated.ScanConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Tags", "Target"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Scan", matchedFields, node)
			}
		}
	case *generated.StandardConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Domains", "Framework", "GoverningBody", "Name", "ShortName", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Standard", matchedFields, node)
			}
		}
	case *generated.SubcontrolConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Aliases", "Category", "Description", "DisplayID", "MappedCategories", "RefCode", "Subcategory", "Tags", "Title"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Subcontrol", matchedFields, node)
			}
		}
	case *generated.SubprocessorConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Subprocessor", matchedFields, node)
			}
		}
	case *generated.SubscriberConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Email", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Subscriber", matchedFields, node)
			}
		}
	case *generated.TaskConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"DisplayID", "Tags", "Title"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Task", matchedFields, node)
			}
		}
	case *generated.TemplateConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "Template", matchedFields, node)
			}
		}
	case *generated.UserConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"DisplayID", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "User", matchedFields, node)
			}
		}
	}
}
