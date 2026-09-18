package slateparser

import (
	"bytes"
	"encoding/json"
	"html"
	"reflect"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/samber/lo"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
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

func isCommentKey(key string) bool {
	return key == "comment" || strings.HasPrefix(key, "comment_")
}

// NoDetailsChanged checks if the only changes between the old and new slate JSON elements are the addition of comments (or no changes at all)
func NoDetailsChanged(oldText []any, newText []any) bool {
	oldLeaves, oldOK := getNodeLeaves(oldText)
	newLeaves, newOK := getNodeLeaves(newText)
	if !oldOK || !newOK {
		return false
	}

	oldIndex, newIndex := 0, 0
	oldOffset, newOffset := 0, 0
	// Adding a comment can split one Slate leaf into several leaves, so compare matching text spans.
	for oldIndex < len(oldLeaves) && newIndex < len(newLeaves) {
		oldLeaf := oldLeaves[oldIndex]
		newLeaf := newLeaves[newIndex]
		oldValue := oldLeaf.node["text"].(string)
		newValue := newLeaf.node["text"].(string)

		if oldOffset == len(oldValue) {
			oldIndex++
			oldOffset = 0
			continue
		}
		if newOffset == len(newValue) {
			newIndex++
			newOffset = 0
			continue
		}

		// Only comment-related attributes may differ between matching spans.
		if !sameNonCommentAttributes(oldLeaf.node, newLeaf.node) {
			return false
		}

		length := min(len(oldValue)-oldOffset, len(newValue)-newOffset)
		if oldValue[oldOffset:oldOffset+length] != newValue[newOffset:newOffset+length] {
			return false
		}

		oldOffset += length
		newOffset += length
	}

	for oldIndex < len(oldLeaves) && oldOffset == len(oldLeaves[oldIndex].node["text"].(string)) {
		oldIndex++
		oldOffset = 0
	}
	for newIndex < len(newLeaves) && newOffset == len(newLeaves[newIndex].node["text"].(string)) {
		newIndex++
		newOffset = 0
	}

	return oldIndex == len(oldLeaves) && newIndex == len(newLeaves)
}

func sameNonCommentAttributes(oldLeaf, newLeaf map[string]any) bool {
	for key, oldValue := range oldLeaf {
		if key == "text" || isCommentKey(key) {
			continue
		}

		newValue, exists := newLeaf[key]
		if !exists || !reflect.DeepEqual(oldValue, newValue) {
			return false
		}
	}

	for key := range newLeaf {
		if key == "text" || isCommentKey(key) {
			continue
		}
		if _, exists := oldLeaf[key]; !exists {
			return false
		}
	}

	return true
}

// MergeComments preserves existing comment markers when a new marker is added from a stale document snapshot.
func MergeComments(oldComment []any, newComment []any) ([]any, bool) {
	decodeSlateElements(newComment)
	oldLeaves, oldOK := getNodeLeaves(oldComment)
	newLeaves, newOK := getNodeLeaves(newComment)
	if !oldOK || !newOK || oldLeaves[len(oldLeaves)-1].end != newLeaves[len(newLeaves)-1].end {
		return newComment, false
	}

	oldMarkers := make(map[string]struct{})
	for _, leaf := range oldLeaves {
		for key := range leaf.node {
			if strings.HasPrefix(key, "comment_") {
				oldMarkers[key] = struct{}{}
			}
		}
	}

	hasAddedMarker := false
	for _, leaf := range newLeaves {
		for key := range leaf.node {
			if strings.HasPrefix(key, "comment_") {
				if _, exists := oldMarkers[key]; !exists {
					hasAddedMarker = true
					break
				}
			}
		}
	}

	if !hasAddedMarker {
		return newComment, false
	}

	merged := false
	for oldIndex, newIndex := 0, 0; oldIndex < len(oldLeaves) && newIndex < len(newLeaves); {
		oldLeaf := oldLeaves[oldIndex]
		newLeaf := newLeaves[newIndex]
		if oldLeaf.start < newLeaf.end && newLeaf.start < oldLeaf.end {
			for key, value := range oldLeaf.node {
				if isCommentKey(key) {
					if _, exists := newLeaf.node[key]; !exists {
						newLeaf.node[key] = value
						merged = true
					}
				}
			}
		}

		if oldLeaf.end <= newLeaf.end {
			oldIndex++
		} else {
			newIndex++
		}
	}

	return newComment, merged
}

type nodeLeaf struct {
	node       map[string]any
	start, end int
}

func getNodeLeaves(elements []any) ([]nodeLeaf, bool) {
	leaves := getChildrenFromSlateTextJSON(elements)
	if len(leaves) == 0 {
		return nil, false
	}

	offset := 0
	valid := true

	parsedLeaves := lo.Map(leaves, func(child any, _ int) nodeLeaf {
		currentLeaf, ok := child.(map[string]any)
		if !ok || currentLeaf == nil {
			valid = false
			return nodeLeaf{}
		}

		text, ok := currentLeaf["text"].(string)
		if !ok {
			valid = false
			return nodeLeaf{}
		}

		parsedLeaf := nodeLeaf{node: currentLeaf, start: offset, end: offset + len(text)}
		offset += len(text)

		return parsedLeaf
	})
	if !valid {
		return nil, false
	}

	return parsedLeaves, true
}

func decodeSlateElements(elements []any) {
	for i, element := range elements {
		encoded, ok := element.(string)
		if !ok {
			continue
		}

		var decoded map[string]any
		if err := json.Unmarshal([]byte(encoded), &decoded); err == nil {
			elements[i] = decoded
		}
	}
}

// DoesMarkdownMatchSlate verifies the slate value preserves the existing Markdown text.
func DoesMarkdownMatchSlate(markdown string, newText []any) bool {
	if markdown == "" || !ContainsCommentsInTextJSON(newText) {
		return false
	}

	var rendered bytes.Buffer
	converter := goldmark.New(goldmark.WithExtensions(extension.Table))
	if err := converter.Convert([]byte(markdown), &rendered); err != nil {
		return false
	}

	txt := strings.Join(strings.Fields(html.UnescapeString(bluemonday.StrictPolicy().Sanitize(rendered.String()))), "")

	var slateText strings.Builder
	if !isSlateTextValid(newText, &slateText) {
		return false
	}

	return txt != "" && txt == strings.Join(strings.Fields(slateText.String()), "")
}

func isSlateTextValid(elements []any, text *strings.Builder) bool {
	for _, element := range elements {
		var node map[string]any
		switch value := element.(type) {
		case string:
			if err := json.Unmarshal([]byte(value), &node); err != nil {
				return false
			}
		case map[string]any:
			node = value
		default:
			return false
		}

		if nodeType, _ := node["type"].(string); nodeType == "mention" {
			value, ok := node["value"].(string)
			if !ok {
				return false
			}

			text.WriteString(value)
			continue
		}

		if value, exists := node["text"]; exists {
			value, ok := value.(string)
			if !ok {
				return false
			}

			text.WriteString(value)
			continue
		}

		children, ok := node["children"].([]any)
		if !ok || !isSlateTextValid(children, text) {
			return false
		}
	}

	return true
}
