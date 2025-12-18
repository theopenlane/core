//go:build gencmd

package prompts

import (
	"regexp"

	"github.com/manifoldco/promptui"
)

// GenerateHistory prompts the user if a history schema is found to
// generate the history command
func GenerateHistory() bool {
	prompt := promptui.Prompt{
		Label:     "Generate history command as well? (y/n):",
		IsConfirm: true,
		Templates: templates,
	}

	resp, _ := prompt.Run()

	return isConfirm(resp)
}

// acceptable answers: y,yes,Y,YES,1
var confirmExp = regexp.MustCompile("(?i)y(?:es)?|1")

func isConfirm(confirm string) bool {
	return confirmExp.MatchString(confirm)
}
