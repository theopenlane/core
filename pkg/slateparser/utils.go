package slateparser

import (
	"encoding/json"
	"fmt"
	"maps"
)

// ContainsCommentsInTextJSON checks if the provided slate JSON elements contain any comments
func ContainsCommentsInTextJSON(elements []any) bool {
	children := getChildrenFromSlateTextJSON(elements)

	for _, child := range children {
		if childMap, ok := child.(map[string]any); ok {
			if _, hasComment := childMap["comment"]; hasComment {
				return true
			}
		}
	}

	return false
}

func getChildrenFromSlateTextJSON(elements []any) []any {
	children := make([]any, 0)
	for _, elem := range elements {
		var m map[string]any
		switch v := elem.(type) {
		case string:
			if err := json.Unmarshal([]byte(v), &m); err != nil {
				continue
			}
		case map[string]any:
			m = v
		default:
			continue
		}

		if c, ok := m["children"].([]any); ok {
			children = append(children, c...)
		}
	}

	return children
}

// OnlyCommentsAdded checks if the only changes between the old and new slate JSON elements are the addition of comments
func OnlyCommentsAdded(oldText []any, newText []any) bool {
	// for each I want to see if the only change is the addition of a comment, if so return true, otherwise false
	oldChildren := getChildrenFromSlateTextJSON(oldText)
	newChildren := getChildrenFromSlateTextJSON(newText)

	if len(oldChildren) != len(newChildren) {
		fmt.Printf("old and new children have different lengths, old: %d, new: %d\n", len(oldChildren), len(newChildren))
		return false
	}

	// compare all children, and check the text is the same, comments are OK
	for i, oldChild := range oldChildren {
		// get the map
		oldChildMap, oldOK := oldChild.(map[string]any)
		newChildMap, newOK := newChildren[i].(map[string]any)

		// if its not a map, we can't compare, so we assume it's not just a comment change and return false
		if !oldOK || !newOK {
			return false
		}

		// if they are equal, continue to the next one
		if maps.Equal(oldChildMap, newChildMap) {
			continue
		}

		allowedKeys := map[string]bool{
			"text":    true,
			"comment": true,
		}

		// if there are other keys besides text and comment, return false
		for key := range newChildMap {
			if !allowedKeys[key] {
				return false
			}
		}

		// if they are not equal, check the text is the same
		oldText, oldOK := oldChildMap["text"]
		newText, newOK := newChildMap["text"]

		if oldOK && newOK {
			oldTextStr, oldOK := oldText.(string)
			newTextStr, newOK := newText.(string)

			// if they are strings and they are not equal, then it's not just a comment change, return false
			if oldOK && newOK && oldTextStr != newTextStr {
				return false
			}
		}

	}

	return true
}
