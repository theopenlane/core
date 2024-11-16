package soiree

import (
	"strings"
)

var (
	singleWildcard = "*" // Now only one wildcard variable for single node
	multiWildcard  = "**"
)

// matchTopicPattern checks if the given subject matches the pattern with wildcards
func matchTopicPattern(pattern, subject string) bool {
	// Special case: single wildcard matches an empty string
	if pattern == singleWildcard && subject == "" {
		return true
	}

	patternParts := strings.Split(pattern, ".")
	subjectParts := strings.Split(subject, ".")

	// Handle the case where pattern ends with ".**", it should not match just "event"
	if len(patternParts) > 1 && patternParts[len(patternParts)-1] == multiWildcard && len(subjectParts) == 1 && subjectParts[0] == patternParts[0] {
		return false
	}

	return matchParts(patternParts, subjectParts)
}

// matchParts is an iterative function to match pattern parts with subject parts
func matchParts(patternParts, subjectParts []string) bool {
	p, s := 0, 0

	for p < len(patternParts) && s < len(subjectParts) {
		switch patternParts[p] {
		case singleWildcard:
			// The single wildcard should match exactly one non-empty subject part
			s++
		case multiWildcard:
			// '**' matches any number of subject parts
			if p == len(patternParts)-1 {
				// If '**' is the last part in the pattern, it matches the rest of the subject
				return true
			}
			// Try to match the rest of the pattern with the remaining subject parts
			for i := s; i <= len(subjectParts); i++ {
				if matchParts(patternParts[p+1:], subjectParts[i:]) {
					return true
				}
			}
			return false
		default:
			// Exact match required for non-wildcard parts
			if patternParts[p] != subjectParts[s] {
				return false
			}
			s++
		}
		p++
	}

	// If we've reached the end of pattern parts and subject parts simultaneously, it's a match
	if p == len(patternParts) && s == len(subjectParts) {
		return true
	}

	// If we've reached the end of the subject but the pattern has remaining parts (other than '**'), it's not a match
	if s == len(subjectParts) {
		for i := p; i < len(patternParts); i++ {
			if patternParts[i] != multiWildcard {
				return false
			}
		}
		return true
	}

	// If we've reached the end of the pattern but not the subject, it's not a match
	return false
}

// isValidTopicName checks if the topic name is valid, obviously
func isValidTopicName(topicName string) bool {
	return !strings.ContainsAny(topicName, "?[")
}
