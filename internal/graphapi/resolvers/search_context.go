package resolvers


import (
	"context"
	"reflect"
	"strings"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/models"
)

const (
	maxSnippetLength = 100
)

type searchCtxTracker struct {
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
	if _, ok := t.contexts[entityID]; !ok {
		t.contexts[entityID] = &models.SearchContext{
			EntityID:      entityID,
			EntityType:    entityType,
			MatchedFields: fieldMatches,
			Snippets:      make([]*models.SearchSnippet, 0),
		}
	}

	t.contexts[entityID].Snippets = append(t.contexts[entityID].Snippets, t.extractSnippets(entity, fieldMatches)...)
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

		var text string

		switch field.Kind() {

		case reflect.String:
			text = field.String()

		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.String {
				// handle string slices (tags - a good example here)
				strs := make([]string, 0, field.Len())
				for i := 0; i < field.Len(); i++ {
					strs = append(strs, field.Index(i).String())
				}
				text = strings.Join(strs, ", ")
			}

		case reflect.Pointer:
			if !field.IsNil() && field.Elem().Kind() == reflect.String {
				text = field.Elem().String()
			}
		}

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
		switch field.Kind() {
		case reflect.String:
			matched = strings.Contains(strings.ToLower(field.String()), queryLower)
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.String {
				for i := 0; i < field.Len(); i++ {
					if strings.Contains(strings.ToLower(field.Index(i).String()), queryLower) {
						matched = true
						break
					}
				}
			}
		case reflect.Pointer:
			if !field.IsNil() && field.Elem().Kind() == reflect.String {
				matched = strings.Contains(strings.ToLower(field.Elem().String()), queryLower)
			}
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
//
//nolint:gocyclo
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
	case *generated.JobRunnerConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"DisplayID", "Name", "Tags"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "JobRunner", matchedFields, node)
			}
		}
	case *generated.JobTemplateConnection:
		for _, edge := range conn.Edges {
			if edge == nil || edge.Node == nil {
				continue
			}
			node := edge.Node
			matchedFields := checker.check(node, []string{"Description", "DisplayID", "Tags", "Title"})
			if len(matchedFields) > 0 {
				tracker.addMatch(node.ID, "JobTemplate", matchedFields, node)
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
