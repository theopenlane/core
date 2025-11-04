package schema

import "github.com/brianvoe/gofakeit/v7"

// defaultTagColors is a list of default colors to use for tags
// if no color is specified, a random color from this list will be chosen
var defaultTagColors = []string{
	"#2CCBAB", // teal
	"#0EA5E9", // blue
	"#10B981", // green
	"#F59E0B", // amber
	"#EF4444", // red
	"#8B5CF6", // violet
	"#EC4899", // pink
	"#14B8A6", // cyan
	"#6366F1", // indigo
	"#84CC16", // lime
	"#F97316", // orange
	"#64748B", // slate
}

// defaultRandomColor returns a random color from the defaultTagColors list
func defaultRandomColor() string {
	randomIndex := gofakeit.Number(0, len(defaultTagColors)-1)
	return defaultTagColors[randomIndex]
}
