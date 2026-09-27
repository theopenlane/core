package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// ValidateUserInput reports whether a payload satisfies the definition's user input schema, so a
// caller can reject an install before authorizing rather than failing on every later sync
func (r *Runtime) ValidateUserInput(ctx context.Context, def types.Definition, userInput json.RawMessage) error {
	if def.UserInput == nil {
		return nil
	}

	return validatePayload(ctx, def.UserInput.Schema, userInput, ErrUserInputInvalid)
}

// validatePayload validates data against a JSON schema, returning the sentinel error when validation fails
func validatePayload(ctx context.Context, schema, data json.RawMessage, sentinel error) error {
	if len(schema) == 0 {
		return nil
	}

	// an absent payload still has to be checked, otherwise required fields never fire
	if jsonx.IsEmptyRawMessage(data) {
		data = json.RawMessage("{}")
	}

	result, err := jsonx.ValidateSchema(schema, data)
	if err != nil {
		return err
	}

	if !result.Valid() {
		issues := jsonx.ValidationErrorStrings(result)

		logx.FromContext(ctx).Info().Strs("issues", issues).Msg("schema validation failed")

		return fmt.Errorf("%w: %s", sentinel, strings.Join(issues, "; "))
	}

	return nil
}
