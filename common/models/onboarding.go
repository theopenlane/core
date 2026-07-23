package models

// InputType identifies how a question is rendered and answered
type InputType string

const (
	// InputTypeString is a free text input
	InputTypeString InputType = "string"
	// InputTypeBoolean is a true or false input
	InputTypeBoolean = "boolean"
	// InputTypeCheckbox is a single checkbox input
	InputTypeCheckbox = "checkbox"
	// InputTypeSelect is a single choice from a list
	InputTypeSelect = "select"
	// InputTypeMultiselect is multiple choices from a list
	InputTypeMultiselect = "multiselect"
	// InputTypeMultiInput is multiple free text values
	InputTypeMultiInput = "multi-input"
)

// Questionnaire is a versioned set of onboarding steps
type Questionnaire struct {
	// Version is the schema version of the questionnaire
	Version string `json:"version" yaml:"version"`
	// Steps are the ordered steps a user completes
	Steps []Step `json:"steps" yaml:"steps"`
}

// Step is a single page of questions in the questionnaire
type Step struct {
	// Key uniquely identifies the step
	Key string `json:"key" yaml:"key"`
	// Title is the display title of the step
	Title string `json:"title" yaml:"title"`
	// Description is optional helper text for the step
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Order is the position of the step in the flow
	Order int `json:"order" yaml:"order"`
	// Hidden hides the step from the user
	Hidden bool `json:"hidden" yaml:"hidden"`
	// Questions are the questions shown on the step
	Questions []Question `json:"questions" yaml:"questions"`
	// Cards are the informational cards shown on the step
	Cards []Card `json:"cards,omitempty" yaml:"cards,omitempty"`
	// DynamicCards indicates the cards are populated at runtime
	DynamicCards bool `json:"-" yaml:"dynamicCards,omitempty"`
}

// Question is a single prompt the user answers
type Question struct {
	// Key uniquely identifies the question
	Key string `json:"key" yaml:"key"`
	// Label is the display label of the question
	Label string `json:"label" yaml:"label"`
	// Description is optional helper text for the question
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// InputType is how the question is rendered and answered
	InputType InputType `json:"inputType" yaml:"inputType"`
	// Format is an optional format hint for the input
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
	// CheckboxLabel is the label shown beside a checkbox input
	CheckboxLabel string `json:"checkboxLabel,omitempty" yaml:"checkboxLabel,omitempty"`
	// Required marks the question as mandatory
	Required bool `json:"required" yaml:"required"`
	// Hidden hides the question from the user
	Hidden bool `json:"hidden" yaml:"hidden"`
	// DependsOn conditionally shows the question based on another answer
	DependsOn *QuestionDependency `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	// DynamicOptions indicates the options are populated at runtime
	DynamicOptions bool `json:"-" yaml:"dynamicOptions,omitempty"`
	// Options are the selectable choices for the question
	Options []QuestionOption `json:"options,omitempty" yaml:"options,omitempty"`
}

// QuestionDependency gates a question on another question's answer
type QuestionDependency struct {
	// Key is the question this dependency refers to
	Key string `json:"key" yaml:"key"`
	// Equals is the value the referenced answer must match
	Equals any `json:"equals" yaml:"equals"`
}

// Card is an informational card shown on a step
type Card struct {
	// Key uniquely identifies the card
	Key string `json:"key" yaml:"key"`
	// Title is the display title of the card
	Title string `json:"title" yaml:"title"`
	// Description is the body text of the card
	Description string `json:"description" yaml:"description"`
}

// QuestionOption is a selectable choice for a question
type QuestionOption struct {
	// Value is the stored value of the option
	Value string `json:"value" yaml:"value"`
	// Label is the display label of the option
	Label string `json:"label" yaml:"label"`
	// Description is optional helper text for the option
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// LogoURL is an optional logo shown with the option
	LogoURL string `json:"logoUrl,omitempty" yaml:"logoUrl,omitempty"`
	// Priority orders the option relative to others
	Priority int `json:"priority,omitempty" yaml:"priority,omitempty"`
	// Hidden hides the option from the user
	Hidden bool `json:"hidden" yaml:"hidden"`
}
