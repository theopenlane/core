package validator_test

import (
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/validator"
)

func TestValidateDomains(t *testing.T) {
	validDomains := []string{
		"example.com",
		"sub.example.com",
		"sub.sub.example.com",
		"https://example.com",
		"http://example.com",
		"http://example.com/hello",
	}

	for _, d := range validDomains {
		t.Run(fmt.Sprintf("valid domain: %s", d), func(t *testing.T) {
			funcCheck := validator.ValidateDomains()
			err := funcCheck([]string{d})
			assert.Empty(t, err)
		})
	}

	longDomain := gofakeit.LetterN(256) + ".com"

	invalidDomains := []string{
		"example",
		"example.",
		longDomain,
	}

	for _, d := range invalidDomains {
		t.Run(fmt.Sprintf("invalid domain: %s", d), func(t *testing.T) {
			funcCheck := validator.ValidateDomains()
			err := funcCheck([]string{d})
			assert.NotEmpty(t, err)
		})
	}
}
