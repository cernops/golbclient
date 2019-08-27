// +build !go1.10

package native

import (
	"strings"
)

func RemoveAllWhitespaces(old string) string {
	return strings.Replace(old, " ", "", -1)
}

