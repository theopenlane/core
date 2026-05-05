package slateparser

import (
	"encoding/json"
	"strings"
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

// getChildrenFromSlateTextJSON recursively collects all leaf nodes (nodes with a "text" key)
// from the slate element tree, handling both JSON string and map[string]any inputs
func getChildrenFromSlateTextJSON(elements []any) []any {
	var leaves []any
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
		collectLeafNodes(m, &leaves)
	}
	return leaves
}

// collectLeafNodes walks a slate node tree and appends leaf nodes (those with a "text" key) to leaves
func collectLeafNodes(m map[string]any, leaves *[]any) {
	if _, hasText := m["text"]; hasText {
		*leaves = append(*leaves, m)
		return
	}

	if children, ok := m["children"].([]any); ok {
		for _, child := range children {
			if childMap, ok := child.(map[string]any); ok {
				collectLeafNodes(childMap, leaves)
			}
		}
	}
}

// valEqualBestEffort compares two any values for scalar JSON types without panicking on non-comparable types
func valEqualBestEffort(a, b any) bool {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case nil:
		return b == nil
	default:
		// non-comparable type (slice, nested map, etc.) — conservatively treat as not equal
		return false
	}
}

func isCommentKey(key string) bool {
	return key == "comment" || strings.HasPrefix(key, "comment_")
}

// OnlyCommentsAdded checks if the only changes between the old and new slate JSON elements are the addition of comments
func OnlyCommentsAdded(oldText []any, newText []any) bool {
	oldLeaves := getChildrenFromSlateTextJSON(oldText)
	newLeaves := getChildrenFromSlateTextJSON(newText)

	if len(oldLeaves) != len(newLeaves) || len(newLeaves) == 0 {
		return false
	}

	for i, oldChild := range oldLeaves {
		oldLeaf, oldOK := oldChild.(map[string]any)
		newLeaf, newOK := newLeaves[i].(map[string]any)

		if !oldOK || !newOK || oldLeaf == nil || newLeaf == nil {
			return false
		}

		// text must be unchanged
		oldTextStr, _ := oldLeaf["text"].(string)
		newTextStr, _ := newLeaf["text"].(string)
		if oldTextStr != newTextStr {
			return false
		}

		// new leaf may only add comment-related keys; all other keys must exist in old with equal values
		for key, newVal := range newLeaf {
			if isCommentKey(key) {
				continue
			}

			oldVal, exists := oldLeaf[key]
			if !exists || !valEqualBestEffort(oldVal, newVal) {
				return false
			}
		}

		// no non-comment keys should be removed from old
		for key := range oldLeaf {
			if isCommentKey(key) {
				continue
			}

			if _, exists := newLeaf[key]; !exists {
				return false
			}
		}
	}

	return true
}
