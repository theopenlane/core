package placeholder

import (
	"encoding/json"
	"maps"
	"regexp"
	"strconv"
	"time"
)

// variablePattern matches a {{name}} placeholder, optionally offset by days as {{name+14}} or {{name-3}}, and optionally
// carrying a default as {{name|default}}, the default kept with its leading bar
var variablePattern = regexp.MustCompile(`\{\{([A-Za-z][A-Za-z0-9_.]*)([+-]\d+)?(\|[^{}]*)?\}\}`)

// Variables are the values that fill {{name}}, {{name+N}} and {{name|default}} placeholders
type Variables map[string]string

// With returns the variables with overrides layered on top, leaving both inputs unchanged
func (v Variables) With(overrides Variables) Variables {
	merged := maps.Clone(v)
	if merged == nil {
		merged = Variables{}
	}

	maps.Copy(merged, overrides)

	return merged
}

// Fill replaces every placeholder in doc with the variable's value escaped for a JSON string, moved by the placeholder's
// day offset when it has one, or its default when unset, leaving an unset placeholder without a default unchanged
func (v Variables) Fill(doc string) string {
	return variablePattern.ReplaceAllStringFunc(doc, v.fillValue)
}

// fillValue returns the value one placeholder resolves to
func (v Variables) fillValue(placeholder string) string {
	match := variablePattern.FindStringSubmatch(placeholder)

	if value, ok := v[match[1]]; ok {
		return offsetValue(placeholder, value, match[2])
	}

	if match[3] == "" {
		return placeholder
	}

	return match[3][1:]
}

// offsetValue returns value escaped for a JSON string, moved by offset days when there is one, leaving the placeholder
// unchanged when value is not a date or the offset does not parse
func offsetValue(placeholder, value, offset string) string {
	if offset == "" {
		return JSONStringContent(value)
	}

	day, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return placeholder
	}

	days, err := strconv.Atoi(offset)
	if err != nil {
		return placeholder
	}

	return day.AddDate(0, 0, days).Format(time.DateOnly)
}

// JSONStringContent escapes value for use inside a JSON string, so a filled value cannot break the document it lands in
func JSONStringContent(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	return string(encoded[1 : len(encoded)-1])
}
