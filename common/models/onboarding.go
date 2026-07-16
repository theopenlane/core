package models

type InputType string

const (
	InputTypeString      InputType = "string"
	InputTypeBoolean              = "boolean"
	InputTypeCheckbox             = "checkbox"
	InputTypeSelect               = "select"
	InputTypeMultiselect          = "multiselect"
	InputTypeMultiInput           = "multi-input"
)

type Questionnaire struct {
	Version string `json:"version" yaml:"version"`
	Steps   []Step `json:"steps" yaml:"steps"`
}

type Step struct {
	Key            string               `json:"key" yaml:"key"`
	Title          string               `json:"title" yaml:"title"`
	Description    string               `json:"description,omitempty" yaml:"description,omitempty"`
	Order          int                  `json:"order" yaml:"order"`
	Hidden         bool                 `json:"hidden" yaml:"hidden"`
	Questions      []Question           `json:"questions" yaml:"questions"`
	Modules        []Module             `json:"modules,omitempty" yaml:"modules,omitempty"`
	DynamicModules bool                 `json:"-" yaml:"dynamicModules,omitempty"`
	Tasks          []TaskRule `json:"-" yaml:"tasks,omitempty"`
}

type Question struct {
	Key            string                        `json:"key" yaml:"key"`
	Label          string                        `json:"label" yaml:"label"`
	Description    string                        `json:"description,omitempty" yaml:"description,omitempty"`
	InputType      InputType                     `json:"inputType" yaml:"inputType"`
	Format         string                        `json:"format,omitempty" yaml:"format,omitempty"`
	CheckboxLabel  string                        `json:"checkboxLabel,omitempty" yaml:"checkboxLabel,omitempty"`
	Required       bool                          `json:"required" yaml:"required"`
	Hidden         bool                          `json:"hidden" yaml:"hidden"`
	DependsOn      *QuestionDependency           `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	DynamicOptions bool                          `json:"-" yaml:"dynamicOptions,omitempty"`
	Options        []QuestionOption              `json:"options,omitempty" yaml:"options,omitempty"`
	Tasks          map[string]TaskRule `json:"-" yaml:"tasks,omitempty"`
}

type QuestionDependency struct {
	Key    string `json:"key" yaml:"key"`
	Equals any    `json:"equals" yaml:"equals"`
}

type Module struct {
	Key         string `json:"key" yaml:"key"`
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
}

type QuestionOption struct {
	Value       string `json:"value" yaml:"value"`
	Label       string `json:"label" yaml:"label"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	LogoURL     string `json:"logoUrl,omitempty" yaml:"logoUrl,omitempty"`
	Priority    int    `json:"priority,omitempty" yaml:"priority,omitempty"`
	Hidden      bool   `json:"hidden" yaml:"hidden"`
}

type TaskRule struct {
	Key                string         `json:"-" yaml:"key"`
	Title              string         `json:"-" yaml:"title"`
	Details            string         `json:"-" yaml:"details"`
	Priority           int            `json:"-" yaml:"priority"`
	AvailableAfterDays int            `json:"-" yaml:"availableAfterDays,omitempty"`
	Metadata           map[string]any `json:"-" yaml:"metadata,omitempty"`
}
