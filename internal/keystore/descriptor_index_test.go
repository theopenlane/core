package keystore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/integrations/types"
)

var (
	providerA = types.ProviderTypeFromString("github")
	providerB = types.ProviderTypeFromString("slack")
)

func TestGroupDescriptors_Empty(t *testing.T) {
	data := map[string]string{}
	result := groupDescriptors(data, func(_ string) types.ProviderType { return providerA })
	assert.Empty(t, result)
}

func TestGroupDescriptors_SingleEntry(t *testing.T) {
	data := map[string]string{"key1": "val1"}
	result := groupDescriptors(data, func(_ string) types.ProviderType { return providerA })

	assert.Len(t, result, 1)
	assert.Equal(t, []string{"val1"}, result[providerA])
}

func TestGroupDescriptors_MultipleProviders(t *testing.T) {
	data := map[string]string{
		"gh:op1":    "desc-gh-1",
		"gh:op2":    "desc-gh-2",
		"slack:op1": "desc-slack-1",
	}
	providerOf := func(k string) types.ProviderType {
		switch k {
		case "gh:op1", "gh:op2":
			return providerA
		default:
			return providerB
		}
	}

	result := groupDescriptors(data, providerOf)

	assert.Len(t, result, 2)
	assert.Len(t, result[providerA], 2)
	assert.Len(t, result[providerB], 1)
	assert.Equal(t, []string{"desc-slack-1"}, result[providerB])
}

func TestGroupDescriptors_SameProvider(t *testing.T) {
	data := map[string]string{
		"op1": "desc1",
		"op2": "desc2",
		"op3": "desc3",
	}
	result := groupDescriptors(data, func(_ string) types.ProviderType { return providerB })

	assert.Len(t, result, 1)
	assert.Len(t, result[providerB], 3)
}

func TestFlattenDescriptors_Empty(t *testing.T) {
	result := flattenDescriptors(map[types.ProviderType][]string{})
	assert.Nil(t, result)
}

func TestFlattenDescriptors_SingleProvider(t *testing.T) {
	grouped := map[types.ProviderType][]string{
		providerA: {"a", "b", "c"},
	}
	result := flattenDescriptors(grouped)
	assert.ElementsMatch(t, []string{"a", "b", "c"}, result)
}

func TestFlattenDescriptors_MultipleProviders(t *testing.T) {
	grouped := map[types.ProviderType][]string{
		providerA: {"a", "b"},
		providerB: {"c", "d"},
	}
	result := flattenDescriptors(grouped)
	assert.ElementsMatch(t, []string{"a", "b", "c", "d"}, result)
}
