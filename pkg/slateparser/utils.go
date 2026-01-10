package slateparser

import "encoding/json"

// ContainsCommentsInTextJSON checks if the provided slate JSON elements contain any comments
func ContainsCommentsInTextJSON(elements []interface{}) bool {
	for _, elem := range elements {
		var m map[string]interface{}
		switch v := elem.(type) {
		case string:
			if err := json.Unmarshal([]byte(v), &m); err != nil {
				continue
			}
		case map[string]interface{}:
			m = v
		default:
			continue
		}

		if children, ok := m["children"].([]interface{}); ok {
			for _, child := range children {
				if childMap, ok := child.(map[string]interface{}); ok {
					if _, hasComment := childMap["comment"]; hasComment {
						return true
					}
				}
			}
		}
	}

	return false
}
