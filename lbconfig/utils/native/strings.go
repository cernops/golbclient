package native

import (
	"strings"
	"unicode"
)

// Optimized version to minimize object allocations (less ~6KB)
func RemoveAllWhitespaces(old string) string {
	var b strings.Builder
	b.Grow(len(old))
	for _, ch := range old {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}