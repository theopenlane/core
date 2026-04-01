package scim

import "github.com/samber/lo"

// ExtractMemberIDsFromValue extracts and deduplicates member IDs from a SCIM members value
func ExtractMemberIDsFromValue(value any) []string {
	members, ok := value.([]any)
	if !ok {
		return nil
	}

	memberIDs := make([]string, 0, len(members))

	for _, m := range members {
		memberMap, ok := m.(map[string]any)
		if !ok {
			continue
		}

		memberID, ok := memberMap["value"].(string)
		if !ok || memberID == "" {
			continue
		}

		memberIDs = append(memberIDs, memberID)
	}

	return lo.Uniq(memberIDs)
}
