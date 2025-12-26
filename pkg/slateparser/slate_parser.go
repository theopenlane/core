package slateparser

import (
	"regexp"
	"strings"
)

// Mention represents a user mention extracted from Slate formatted text
type Mention struct {
	UserID          string
	UserDisplayName string
	ObjectType      string
	ObjectID        string
	ObjectName      string
}

// FieldsToCheck defines the database fields that should be checked for mentions
var FieldsToCheck = []string{
	"comment.text",
	"internalPolicy.Details",
	"procedure.Details",
	"risk.Details",
	"task.Details",
}

// checkForMentions parses the Slate formatted text and extracts all user mentions
// Returns a map where the key is the data-slate-id and the value is the Mention struct
func CheckForMentions(text string, objectType string, objectID string, objectName string) map[string]Mention {
	mentions := make(map[string]Mention)

	// Regular expression to match Slate mention elements
	// Looks for: <div data-slate-node="element" data-slate-inline="true" data-slate-void="true" data-slate-key="..." data-slate-id="..." data-slate-value="...">
	mentionRegex := regexp.MustCompile(`<div[^>]*data-slate-node="element"[^>]*data-slate-inline="true"[^>]*data-slate-void="true"[^>]*data-slate-key="([^"]*)"[^>]*data-slate-id="([^"]*)"[^>]*data-slate-value="([^"]*)"[^>]*>`)

	// Also handle alternative attribute order
	altMentionRegex := regexp.MustCompile(`<div[^>]*data-slate-key="([^"]*)"[^>]*data-slate-id="([^"]*)"[^>]*data-slate-value="([^"]*)"[^>]*data-slate-node="element"[^>]*>`)

	// Find all matches
	matches := mentionRegex.FindAllStringSubmatch(text, -1)
	altMatches := altMentionRegex.FindAllStringSubmatch(text, -1)

	// Process matches from first regex
	for _, match := range matches {
		if len(match) >= 4 {
			userID := match[1]            // data-slate-key (user.ID)
			slateID := match[2]           // data-slate-id (unique identifier)
			displayName := match[3]       // data-slate-value (user.DisplayName)

			mentions[slateID] = Mention{
				UserID:          userID,
				UserDisplayName: displayName,
				ObjectType:      objectType,
				ObjectID:        objectID,
				ObjectName:      objectName,
			}
		}
	}

	// Process matches from alternative regex
	for _, match := range altMatches {
		if len(match) >= 4 {
			userID := match[1]
			slateID := match[2]
			displayName := match[3]

			// Only add if not already present
			if _, exists := mentions[slateID]; !exists {
				mentions[slateID] = Mention{
					UserID:          userID,
					UserDisplayName: displayName,
					ObjectType:      objectType,
					ObjectID:        objectID,
					ObjectName:      objectName,
				}
			}
		}
	}

	return mentions
}

// CheckForNewMentions compares old and new text to identify newly added mentions
// Takes old text, new text, and returns only the slate IDs that are new
func CheckForNewMentions(oldText string, newText string, objectType string, objectID string, objectName string) []string {
	// Get mentions from both texts
	oldMentions := CheckForMentions(oldText, objectType, objectID, objectName)
	newMentions := CheckForMentions(newText, objectType, objectID, objectName)

	// Find mentions that exist in new but not in old
	newSlateIDs := make([]string, 0)
	for slateID := range newMentions {
		if _, existedBefore := oldMentions[slateID]; !existedBefore {
			newSlateIDs = append(newSlateIDs, slateID)
		}
	}

	return newSlateIDs
}

// GetNewMentions is a helper that returns the full Mention objects for new mentions
func GetNewMentions(oldText string, newText string, objectType string, objectID string, objectName string) map[string]Mention {
	newMentions := CheckForMentions(newText, objectType, objectID, objectName)
	newSlateIDs := CheckForNewMentions(oldText, newText, objectType, objectID, objectName)

	result := make(map[string]Mention)
	for _, slateID := range newSlateIDs {
		if mention, exists := newMentions[slateID]; exists {
			result[slateID] = mention
		}
	}

	return result
}

// ExtractMentionedUserIDs extracts just the user IDs from a mention map
func ExtractMentionedUserIDs(mentions map[string]Mention) []string {
	userIDs := make([]string, 0, len(mentions))
	seen := make(map[string]bool)

	for _, mention := range mentions {
		// Deduplicate in case the same user is mentioned multiple times
		if !seen[mention.UserID] {
			userIDs = append(userIDs, mention.UserID)
			seen[mention.UserID] = true
		}
	}

	return userIDs
}

// IsValidSlateText checks if the text contains valid Slate formatted content
func IsValidSlateText(text string) bool {
	// Check for basic Slate structure indicators
	return strings.Contains(text, "data-slate-node") ||
		strings.Contains(text, "data-slate-key") ||
		strings.Contains(text, "data-slate-id")
}
