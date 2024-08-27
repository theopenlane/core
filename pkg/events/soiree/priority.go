package soiree

// Priority type for listener priority levels
type Priority int

const (
	Lowest Priority = iota + 1 // Lowest priority
	Low
	Normal
	High
	Highest
)
