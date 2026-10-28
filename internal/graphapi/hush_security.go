package graphapi

import (
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func validateCreateHushInput(input generated.CreateHushInput) error {
	if input.SecretValue != nil {
		return rout.InvalidField("secretValue")
	}

	return nil
}
