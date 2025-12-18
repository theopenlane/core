//go:build gencmd

package prompts

import (
	"errors"

	"github.com/manifoldco/promptui"
)

// Name prompts the user for the name of the command
func Name(defaultName string) (string, error) {
	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New("name is required") //nolint:err113
		}

		if input[len(input)-1:] == "s" {
			return errors.New("name should be singular") //nolint:err113
		}

		return nil
	}

	prompt := promptui.Prompt{
		Label:     "Name of the command (should be the singular version of the object):",
		Default:   defaultName,
		Templates: templates,
		Validate:  validate,
	}

	return prompt.Run()
}
