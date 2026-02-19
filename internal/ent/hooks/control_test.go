//go:build test

package hooks_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/hooks"
)

func TestControlVisibilityTupleAction(t *testing.T) {
	tests := []struct {
		name                 string
		isTrustCenterControl bool
		newVisibility        enums.TrustCenterDocumentVisibility
		oldVisibility        enums.TrustCenterDocumentVisibility
		visibilityChanged    bool
		expectedWrite        bool
		expectedDelete       bool
	}{
		{
			name:                 "not a trust center control, no action",
			isTrustCenterControl: false,
			newVisibility:        enums.TrustCenterDocumentVisibilityPubliclyVisible,
			oldVisibility:        enums.TrustCenterDocumentVisibilityNotVisible,
			visibilityChanged:    true,
			expectedWrite:        false,
			expectedDelete:       false,
		},
		{
			name:                 "visibility not changed, no action",
			isTrustCenterControl: true,
			newVisibility:        enums.TrustCenterDocumentVisibilityPubliclyVisible,
			oldVisibility:        enums.TrustCenterDocumentVisibilityPubliclyVisible,
			visibilityChanged:    false,
			expectedWrite:        false,
			expectedDelete:       false,
		},
		{
			name:                 "same visibility values, no action",
			isTrustCenterControl: true,
			newVisibility:        enums.TrustCenterDocumentVisibilityNotVisible,
			oldVisibility:        enums.TrustCenterDocumentVisibilityNotVisible,
			visibilityChanged:    true,
			expectedWrite:        false,
			expectedDelete:       false,
		},
		{
			name:                 "not visible to publicly visible, should write",
			isTrustCenterControl: true,
			newVisibility:        enums.TrustCenterDocumentVisibilityPubliclyVisible,
			oldVisibility:        enums.TrustCenterDocumentVisibilityNotVisible,
			visibilityChanged:    true,
			expectedWrite:        true,
			expectedDelete:       false,
		},
		{
			name:                 "protected to publicly visible, should write",
			isTrustCenterControl: true,
			newVisibility:        enums.TrustCenterDocumentVisibilityPubliclyVisible,
			oldVisibility:        enums.TrustCenterDocumentVisibilityProtected,
			visibilityChanged:    true,
			expectedWrite:        true,
			expectedDelete:       false,
		},
		{
			name:                 "publicly visible to not visible, should delete",
			isTrustCenterControl: true,
			newVisibility:        enums.TrustCenterDocumentVisibilityNotVisible,
			oldVisibility:        enums.TrustCenterDocumentVisibilityPubliclyVisible,
			visibilityChanged:    true,
			expectedWrite:        false,
			expectedDelete:       true,
		},
		{
			name:                 "publicly visible to protected, should delete",
			isTrustCenterControl: true,
			newVisibility:        enums.TrustCenterDocumentVisibilityProtected,
			oldVisibility:        enums.TrustCenterDocumentVisibilityPubliclyVisible,
			visibilityChanged:    true,
			expectedWrite:        false,
			expectedDelete:       true,
		},
		{
			name:                 "not visible to protected, no action",
			isTrustCenterControl: true,
			newVisibility:        enums.TrustCenterDocumentVisibilityProtected,
			oldVisibility:        enums.TrustCenterDocumentVisibilityNotVisible,
			visibilityChanged:    true,
			expectedWrite:        false,
			expectedDelete:       false,
		},
		{
			name:                 "protected to not visible, no action",
			isTrustCenterControl: true,
			newVisibility:        enums.TrustCenterDocumentVisibilityNotVisible,
			oldVisibility:        enums.TrustCenterDocumentVisibilityProtected,
			visibilityChanged:    true,
			expectedWrite:        false,
			expectedDelete:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldWrite, shouldDelete := hooks.ControlVisibilityTupleAction(
				tt.isTrustCenterControl,
				tt.newVisibility,
				tt.oldVisibility,
				tt.visibilityChanged,
			)

			assert.Equal(t, tt.expectedWrite, shouldWrite, "shouldWrite mismatch")
			assert.Equal(t, tt.expectedDelete, shouldDelete, "shouldDelete mismatch")
		})
	}
}
