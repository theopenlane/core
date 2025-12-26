package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/common/models"
)

func TestSort(t *testing.T) {
	tests := []struct {
		name     string
		items    []models.Sortable
		expected []models.Sortable
	}{
		{
			name: "Sort ImplementationGuidance",
			items: []models.Sortable{
				&models.ImplementationGuidance{ReferenceID: "CC1.1"},
				&models.ImplementationGuidance{ReferenceID: "CC2.1"},
				&models.ImplementationGuidance{ReferenceID: "CC2.2"},
			},
			expected: []models.Sortable{
				&models.ImplementationGuidance{ReferenceID: "CC1.1"},
				&models.ImplementationGuidance{ReferenceID: "CC2.1"},
				&models.ImplementationGuidance{ReferenceID: "CC2.2"},
			},
		},
		{
			name: "Sort AssessmentMethod",
			items: []models.Sortable{
				&models.AssessmentMethod{ID: "id2"},
				&models.AssessmentMethod{ID: "id1"},
				&models.AssessmentMethod{ID: "id3"},
			},
			expected: []models.Sortable{
				&models.AssessmentMethod{ID: "id1"},
				&models.AssessmentMethod{ID: "id2"},
				&models.AssessmentMethod{ID: "id3"},
			},
		},
		{
			name: "Sort ExampleEvidence",
			items: []models.Sortable{
				&models.ExampleEvidence{DocumentationType: "Policy", Description: "description of the example evidence"},
				&models.ExampleEvidence{DocumentationType: "A Policy", Description: "description of the example evidence"},
				&models.ExampleEvidence{DocumentationType: "Policy", Description: "another description of the example evidence"},
			},
			expected: []models.Sortable{
				&models.ExampleEvidence{DocumentationType: "A Policy", Description: "description of the example evidence"},
				&models.ExampleEvidence{DocumentationType: "Policy", Description: "another description of the example evidence"},
				&models.ExampleEvidence{DocumentationType: "Policy", Description: "description of the example evidence"},
			},
		},
		{
			name: "Sort AssessmentObjective",
			items: []models.Sortable{
				&models.AssessmentObjective{ID: "id-1"},
				&models.AssessmentObjective{ID: "id-3"},
				&models.AssessmentObjective{ID: "id-2"},
			},
			expected: []models.Sortable{
				&models.AssessmentObjective{ID: "id-1"},
				&models.AssessmentObjective{ID: "id-2"},
				&models.AssessmentObjective{ID: "id-3"},
			},
		},
		{
			name:     "Sort empty slice",
			items:    []models.Sortable{},
			expected: []models.Sortable{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.Sort(tt.items)
			assert.Equal(t, tt.expected, result)
		})
	}
}
