package validator_test

import (
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/ent/validator"
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

func TestValidateURL(t *testing.T) {
	validURLs := []string{
		"example.com",
		"sub.example.com",
		"sub.sub.example.com",
		"https://example.com",
		"http://example.com",
		"http://example.com/hello",
		"https://meow.s3.us-easwt-2.amazonaws.com",
		"https://meow.s3.us-east-2.amazonaws.com/organization-111111111-foobar.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ABCDEFEACP3%2F20241001%2Fus-east-2%2Fs3%2Faws4_request&X-Amz-Date=20021001T18112AC&X-Amz-Expires=900&X-Amz-SignedHeaders=host&response-content-disposition=attachment&x-id=GetObject&X-Amz-Signature=56796c66aaf1f895b593eefd1e935e3cd3f7gg2f6azza72284c10ca5854510e2",
	}

	for _, d := range validURLs {
		t.Run(fmt.Sprintf("valid url: %s", d), func(t *testing.T) {
			funcCheck := validator.ValidateURL()
			err := funcCheck(d)
			assert.Empty(t, err)
		})
	}

	longDomain := gofakeit.LetterN(2048) + ".com"

	invalidDomains := []string{
		"example",
		"example.",
		longDomain,
	}

	for _, d := range invalidDomains {
		t.Run(fmt.Sprintf("invalid url: %s", d), func(t *testing.T) {
			funcCheck := validator.ValidateURL()
			err := funcCheck(d)
			assert.NotEmpty(t, err)
		})
	}
}
