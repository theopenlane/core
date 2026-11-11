package urlx

import (
	"errors"
	"regexp"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain words", in: "Acme Corp", want: "acme-corp"},
		{name: "camel case split", in: "AcmeCorp", want: "acme-corp"},
		{name: "trailing period", in: "Acme Inc.", want: "acme-inc"},
		{name: "comma and ampersand", in: "Smith, Jones & Co", want: "smith-jones-co"},
		{name: "underscores and repeated spaces", in: "acme__corp   llc", want: "acme-corp-llc"},
		{name: "surrounding whitespace and hyphens", in: "  -Acme-  ", want: "acme"},
		{name: "non-ascii collapses", in: "Café Düsseldorf", want: "caf-d-sseldorf"},
		{name: "digits kept", in: "Acme 2 Go", want: "acme-2-go"},
		{name: "already normalized is stable", in: "acme-inc", want: "acme-inc"},
		{name: "nothing usable", in: "...", want: ""},
		{name: "empty", in: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Check(t, is.Equal(tt.want, Slug(tt.in)))
		})
	}
}

func TestUniqueSlug(t *testing.T) {
	t.Parallel()

	suffixed := regexp.MustCompile(`^acme-corp-[a-z0-9]{4}$`)

	t.Run("free base is returned bare", func(t *testing.T) {
		t.Parallel()

		got, err := UniqueSlug("acme-corp", func(string) (bool, error) { return false, nil })
		assert.NilError(t, err)
		assert.Check(t, is.Equal("acme-corp", got))
	})

	t.Run("taken base gets a short suffix", func(t *testing.T) {
		t.Parallel()

		got, err := UniqueSlug("acme-corp", func(candidate string) (bool, error) {
			return candidate == "acme-corp", nil
		})
		assert.NilError(t, err)
		assert.Check(t, suffixed.MatchString(got), got)
	})

	t.Run("empty base is never offered bare", func(t *testing.T) {
		t.Parallel()

		var offered []string

		got, err := UniqueSlug("", func(candidate string) (bool, error) {
			offered = append(offered, candidate)

			return false, nil
		})
		assert.NilError(t, err)
		assert.Check(t, is.Len(offered, 1))
		assert.Check(t, is.Regexp(`^[a-z0-9]{4}$`, got))
	})

	t.Run("exhausted after bounded attempts", func(t *testing.T) {
		t.Parallel()

		attempts := 0

		_, err := UniqueSlug("acme-corp", func(string) (bool, error) {
			attempts++

			return true, nil
		})
		assert.Check(t, is.ErrorIs(err, ErrSlugExhausted))
		assert.Check(t, is.Equal(slugAttempts, attempts))
	})

	t.Run("lookup error is returned", func(t *testing.T) {
		t.Parallel()

		boom := errors.New("boom")

		_, err := UniqueSlug("acme-corp", func(string) (bool, error) { return false, boom })
		assert.Check(t, is.ErrorIs(err, boom))
	})
}
