//go:build test

package testharness

import (
	"testing"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
)

// MustVendorQuestion returns the default vendor scoring question for key
func MustVendorQuestion(t *testing.T, key string) models.VendorScoringQuestionDef {
	t.Helper()

	question, found := lo.Find(models.DefaultVendorScoringQuestions, func(question models.VendorScoringQuestionDef) bool {
		return question.Key == key
	})
	if !found {
		t.Fatalf("vendor scoring question %q not found", key)
	}

	return question
}
