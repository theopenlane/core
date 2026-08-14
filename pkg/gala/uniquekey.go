package gala

import "strings"

// uniqueKeySeparator delimits unique key segments
const uniqueKeySeparator = ":"

// UniqueKeyNamespace mints colon-delimited insert-time dedup keys under one declared prefix.
// River scopes uniqueness by job kind, so a prefix must be unique within its kind family
type UniqueKeyNamespace struct {
	prefix string
}

// NewUniqueKeyNamespace declares a unique key namespace with the given prefix
func NewUniqueKeyNamespace(prefix string) UniqueKeyNamespace {
	return UniqueKeyNamespace{prefix: prefix}
}

// Key returns the namespaced key for the given segments
func (n UniqueKeyNamespace) Key(segments ...string) string {
	return n.prefix + uniqueKeySeparator + strings.Join(segments, uniqueKeySeparator)
}

// AppendKeySegments extends an existing unique key with additional segments
func AppendKeySegments(key string, segments ...string) string {
	return key + uniqueKeySeparator + strings.Join(segments, uniqueKeySeparator)
}
