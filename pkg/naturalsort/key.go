package naturalsort

import (
	"fmt"
	"strings"
)

// maxSignificantDigits is the longest digit run whose length fits the two-digit prefix; longer runs share its prefix
const maxSignificantDigits = 99

// Key returns a key whose bytewise order is the natural order of s: ASCII letters compare case-insensitively and every
// run of ASCII digits compares by numeric value, so "AC-2" < "AC-2(1)" < "AC-10"; keys must be compared bytewise (C collation)
func Key(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); {
		c := s[i]
		if !isDigit(c) {
			b.WriteByte(toLowerASCII(c))
			i++

			continue
		}

		start := i
		for i < len(s) && isDigit(s[i]) {
			i++
		}

		writeNumber(&b, s[start:i])
	}

	return b.String()
}

// writeNumber writes a digit run as its two-digit significant length followed by its significant digits
func writeNumber(b *strings.Builder, digits string) {
	significant := strings.TrimLeft(digits, "0")
	if significant == "" {
		significant = "0"
	}

	fmt.Fprintf(b, "%02d%s", min(len(significant), maxSignificantDigits), significant)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func toLowerASCII(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}

	return c
}
