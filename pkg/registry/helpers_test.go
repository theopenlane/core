package registry_test

import (
	"reflect"
	"testing"

	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/registry"
)

type Framework struct {
	Name      string    `json:"name" jsonschema:"minLength=1,description=name of the framework,example=SOC 2,example=NIST CSF"`
	Framework string    `json:"framework" jsonschema:"minLength=1,description=shortname (slug) of the framework,example=soc2,example=nist_csf"`
	Version   string    `json:"version" jsonschema:"minLength=1,description=version of the framework,example=2017,example=1.1"`
	WebLink   string    `json:"web_link,omitempty" jsonschema:"type=string,description=link to the documentation for the controls"`
	Controls  []Control `json:"controls" jsonschema:"description=a list of controls for the framework"`
}

type Control struct {
	Name        string `json:"name,omitempty" jsonschema:"description=name of the control"`
	Description string `json:"description,omitempty" jsonschema:"description=short description of the control"`
	RefCode     string `json:"ref_code" jsonschema:"minLength=1,description=unique identifier for the control in the framework; sometimes referred to as a control number,example=CC1.1,example=AC-2"`
	Category    string `json:"category" jsonschema:"minLength=1,description=category of the control,example=Security,example=Availability"`
	Subcategory string `json:"subcategory,omitempty" jsonschema:"description=subcategory of the control,example=System Operation,example=Control Environment"`
	DTI         string `json:"dti,omitempty" jsonschema:"enum=easy,enum=medium,enum=difficult,description=difficulty to implement rating of the control"`
	DTC         string `json:"dtc,omitempty" jsonschema:"enum=easy,enum=medium,enum=difficult,description=difficulty to collect evidence rating of the control"`
	Guidance    string `json:"guidance,omitempty" jsonschema:"description=guidance or suggested steps for implementing the control"`
}

func TestGetSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    reflect.Type
		expected *jsonschema.Schema
		wantErr  bool
	}{
		{
			name:  "Valid struct",
			input: reflect.TypeOf(Framework{}),
			expected: &jsonschema.Schema{
				Type: "object",
			},
			wantErr: false,
		},
		{
			name:     "Nil type",
			input:    nil,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := registry.GetSchema(tt.input)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Type, schema.Type)
			assert.NotNil(t, schema.Properties)
		})
	}
}
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "Valid struct",
			input: Framework{
				Name:      "SOC 2",
				Framework: "soc2",
				Version:   "2017",
				WebLink:   "https://example.com",
				Controls: []Control{
					{
						Name:        "Control 1",
						Description: "Description 1",
						RefCode:     "CC1.1",
						Category:    "Security",
						Subcategory: "System Operation",
						DTI:         "easy",
						DTC:         "medium",
						Guidance:    "Guidance 1",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "Nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "Invalid struct",
			input: Framework{
				Name:      "",
				Framework: "",
				Version:   "",
				WebLink:   "",
				Controls: []Control{
					{
						Name:        "",
						Description: "",
						RefCode:     "",
						Category:    "",
						Subcategory: "",
						DTI:         "",
						DTC:         "",
						Guidance:    "",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := registry.Validate(tt.input)
			if tt.wantErr {
				assert.NotEmpty(t, vr.Error())

				return
			}

			require.Empty(t, vr)
			assert.Equal(t, "{}", vr.Error())
		})
	}
}
