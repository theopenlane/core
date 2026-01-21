package slateparser

import (
	"regexp"

	"github.com/theopenlane/utils/ulids"
)

// Mention represents a user mention extracted from Slate formatted text
type Mention struct {
	// UserID is the unique identifier of the mentioned user
	UserID string
	// UserDisplayName is the display name shown for the mentioned user
	UserDisplayName string
	// ObjectType is the type of object where the mention was found (e.g., "comment", "task")
	ObjectType string
	// ObjectID is the unique identifier of the object containing the mention
	ObjectID string
	// ObjectName is the human-readable name of the object containing the mention
	ObjectName string
}

var (
	// divTagRegex matches div tags with their attributes
	divTagRegex = regexp.MustCompile(`<div\b([^>]*)>`)
	// Individual attribute regexes for extracting data-slate-* values
	slateKeyRegex   = regexp.MustCompile(`\bdata-slate-key="([^"]*)"`)
	slateIDRegex    = regexp.MustCompile(`\bdata-slate-id="([^"]*)"`)
	slateValueRegex = regexp.MustCompile(`\bdata-slate-value="([^"]*)"`)
	// slateAttrRegex matches data-slate-* attributes within HTML tags to verify valid Slate content
	slateAttrRegex = regexp.MustCompile(`<[^>]+data-slate-(node|key|id)="[^"]*"[^>]*>`)
)

// CheckForMentions parses the Slate formatted text and extracts all user mentions
// Returns a map where the key is the data-slate-id and the value is the Mention struct
func CheckForMentions(text string, objectType string, objectID string, objectName string) map[string]Mention {
	mentions := make(map[string]Mention)

	// Find all div tags
	matches := divTagRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		attrs := match[1] // The attributes portion of the div tag

		// Extract all three required attributes
		keyMatch := slateKeyRegex.FindStringSubmatch(attrs)
		idMatch := slateIDRegex.FindStringSubmatch(attrs)
		valueMatch := slateValueRegex.FindStringSubmatch(attrs)

		// All three attributes must be present
		if keyMatch == nil || idMatch == nil || valueMatch == nil {
			continue
		}

		userID := keyMatch[1]        // data-slate-key (user.ID)
		slateID := idMatch[1]        // data-slate-id (unique identifier)
		displayName := valueMatch[1] // data-slate-value (user.DisplayName)

		mentions[slateID] = Mention{
			UserID:          userID,
			UserDisplayName: displayName,
			ObjectType:      objectType,
			ObjectID:        objectID,
			ObjectName:      objectName,
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

// GetNewMentions is a helper that returns the full Mention objects for new mentions.
// It parses both texts once and returns only mentions that exist in newText but not in oldText.
func GetNewMentions(oldText string, newText string, objectType string, objectID string, objectName string) map[string]Mention {
	oldMentions := CheckForMentions(oldText, objectType, objectID, objectName)
	newMentions := CheckForMentions(newText, objectType, objectID, objectName)

	result := make(map[string]Mention)
	for slateID, mention := range newMentions {
		if _, existedBefore := oldMentions[slateID]; !existedBefore {
			result[slateID] = mention
		}
	}

	return result
}

// ExtractMentionedUserIDs extracts just the user IDs from a mention map.
// It validates that each UserID is a valid ULID and skips invalid ones.
func ExtractMentionedUserIDs(mentions map[string]Mention) []string {
	userIDs := make([]string, 0, len(mentions))
	seen := make(map[string]bool)

	for _, mention := range mentions {
		// Validate that the UserID is a valid ULID
		if _, err := ulids.Parse(mention.UserID); err != nil {
			continue
		}

		// Deduplicate in case the same user is mentioned multiple times
		if !seen[mention.UserID] {
			userIDs = append(userIDs, mention.UserID)
			seen[mention.UserID] = true
		}
	}

	return userIDs
}

// IsValidSlateText checks if the text contains valid Slate formatted content
// by verifying data-slate-* attributes exist within HTML tags
func IsValidSlateText(text string) bool {
	return slateAttrRegex.MatchString(text)
}
