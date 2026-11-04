package placeholder

import (
	"encoding/json"
	"maps"
	"regexp"
	"strconv"
	"time"
)

// variablePattern matches a {{name}} or {{name|default}} placeholder, the default kept with its leading bar
var variablePattern = regexp.MustCompile(`\{\{([A-Za-z][A-Za-z0-9_.]*)(\|[^{}]*)?\}\}`)

// todayPattern matches a {{TODAY}} placeholder, optionally offset by days as {{TODAY+14}} or {{TODAY-3}}
var todayPattern = regexp.MustCompile(`\{\{TODAY([+-]\d+)?\}\}`)

// Variables are the values that fill {{name}} and {{name|default}} placeholders
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

// Fill replaces every variable placeholder in doc with the variable's value escaped for a JSON string, or its default
// when unset, leaving an unset placeholder without a default unchanged
func (v Variables) Fill(doc string) string {
	return variablePattern.ReplaceAllStringFunc(doc, v.fillValue)
}

// fillValue returns the value one variable placeholder resolves to
func (v Variables) fillValue(placeholder string) string {
	match := variablePattern.FindStringSubmatch(placeholder)

	if value, ok := v[match[1]]; ok {
		return JSONStringContent(value)
	}

	if match[2] == "" {
		return placeholder
	}

	return match[2][1:]
}

// Today is the day {{TODAY}} placeholders are filled relative to
type Today time.Time

// Fill replaces every {{TODAY}} placeholder in doc with the day moved by its offset
func (d Today) Fill(doc string) string {
	return todayPattern.ReplaceAllStringFunc(doc, d.fillValue)
}

// fillValue returns the date one placeholder names, leaving a placeholder whose offset does not parse unchanged
func (d Today) fillValue(placeholder string) string {
	days := 0

	if offset := todayPattern.FindStringSubmatch(placeholder)[1]; offset != "" {
		parsed, err := strconv.Atoi(offset)
		if err != nil {
			return placeholder
		}

		days = parsed
	}

	return time.Time(d).UTC().AddDate(0, 0, days).Format(time.DateOnly)
}

// JSONStringContent escapes value for use inside a JSON string, so a filled value cannot break the document it lands in
func JSONStringContent(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	return string(encoded[1 : len(encoded)-1])
}
