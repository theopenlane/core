package model

import "github.com/theopenlane/core/internal/integrations/types"

type IntegrationRecommendation struct {
	ID                    string                       `json:"id"`
	Family                string                       `json:"family,omitempty"`
	DisplayName           string                       `json:"displayName"`
	Description           string                       `json:"description,omitempty"`
	Category              string                       `json:"category,omitempty"`
	DocsURL               string                       `json:"docsURL,omitempty"`
	LogoURL               string                       `json:"logoURL,omitempty"`
	Tags                  []string                     `json:"tags"`
	Active                bool                         `json:"active"`
	Visible               bool                         `json:"visible"`
	Score                 int                          `json:"score"`
	Label                 string                       `json:"label"`
	RecommendationSignals []types.RecommendationSignal `json:"-"`
}
