package models

// SearchSnippet represents a piece of matched content with surrounding context
type SearchSnippet struct {
	Field string `json:"field"`
	Text  string `json:"text"`
}

// SearchContext provides information about why a particular entity matched the search query
type SearchContext struct {
	EntityID      string           `json:"entityID"`
	EntityType    string           `json:"entityType"`
	MatchedFields []string         `json:"matchedFields"`
	Snippets      []*SearchSnippet `json:"snippets,omitempty"`
}
