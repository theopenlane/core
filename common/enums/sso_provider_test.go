package enums_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
)

func TestToSSOProvider(t *testing.T) {
	tests := []struct {
		input    string
		expected enums.SSOProvider
	}{
		{"okta", enums.SSOProviderOkta},
		{"ONE_LOGIN", enums.SSOProviderOneLogin},
		{"GOOGLE_WORKSPACE", enums.SSOProviderGoogleWorkspace},
		{"slack", enums.SSOProviderSlack},
		{"GITHUB", enums.SSOProviderGithub},
		{"unknown", enums.SSOProviderInvalid},
		{"NONE", enums.SSOProviderNone},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("to %s", tc.input), func(t *testing.T) {
			res := enums.ToSSOProvider(tc.input)
			assert.Check(t, is.Equal(tc.expected, *res))
		})
	}
}
