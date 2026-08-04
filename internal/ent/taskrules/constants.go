package taskrules

// Source categorizes where a suggested task came from, defaults to recommendation
type Source string

const (
	// SourceRecommendations marks a task as a general product recommendation; the default
	SourceRecommendations Source = "openlane_recommendations"
	// SourceOnboarding marks a task as part of the guided onboarding flow
	SourceOnboarding Source = "openlane_onboarding"
)
